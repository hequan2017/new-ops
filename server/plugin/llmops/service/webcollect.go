// Package service llmops：网页采集（URL 拉取→HTML 正文抽取→编目落库）
// 抽取口径：纯标准库——去 script/style/noscript/注释，块级标签换行，去标签后 HTML 反转义；
// 仅放行 http/https，15s 超时 + 2MB 大小上限（内网段黑名单等 SSRF 增强留备用池）。
package service

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
)

// 错误码（llmops 段续用）
const (
	ErrCodeWebURLBad   = 2206
	ErrCodeWebFetch    = 2207
	webFetchTimeout    = 15 * time.Second
	webFetchMaxBytes   int64 = 2 << 20
	webFetchUserAgent        = "BaiZe-llmops/0.3 (+webpage-collector)"
)

var webHTTPClient = &http.Client{Timeout: webFetchTimeout}

// reStripBlocks 需整体剔除的元素（Go RE2 不支持反向引用，逐标签编译）
var reStripBlocks = []*regexp.Regexp{
	regexp.MustCompile(`(?is)<script[^>]*>.*?</script>`),
	regexp.MustCompile(`(?is)<style[^>]*>.*?</style>`),
	regexp.MustCompile(`(?is)<noscript[^>]*>.*?</noscript>`),
	regexp.MustCompile(`(?is)<svg[^>]*>.*?</svg>`),
}

var (
	reCommentStrip = regexp.MustCompile(`(?s)<!--.*?-->`)
	reTitleTag     = regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)
	reBlockBreak   = regexp.MustCompile(`(?i)<br\s*/?>|</(p|div|li|tr|h[1-6]|section|article|blockquote|pre|table|thead|tbody|ul|ol|dl|dt|dd|form)\s*>`)
	reAnyTag       = regexp.MustCompile(`(?s)<[^>]*>`)
	reSpaces       = regexp.MustCompile(`[ \t\r\f\x{00a0}]+`) // 含 &nbsp; 反解出的 U+00A0
)

// ValidateWebURL 采集地址协议白名单（拒绝 file/ftp/javascript/data 等非 http(s) scheme）
func ValidateWebURL(raw string) error {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Host == "" {
		return fmt.Errorf("[%d]URL 无法解析: %s", ErrCodeWebURLBad, raw)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("[%d]仅支持 http/https，拒绝 %q", ErrCodeWebURLBad, u.Scheme)
	}
	return nil
}

// ExtractHTMLText 抽取网页标题与正文纯文本（纯标准库，测试锚点）
func ExtractHTMLText(raw string) (title, text string) {
	s := raw
	for _, re := range reStripBlocks {
		s = re.ReplaceAllString(s, " ")
	}
	s = reCommentStrip.ReplaceAllString(s, " ")
	if m := reTitleTag.FindStringSubmatch(s); len(m) == 2 {
		title = strings.TrimSpace(html.UnescapeString(m[1]))
	}
	s = reBlockBreak.ReplaceAllString(s, "\n")
	s = reAnyTag.ReplaceAllString(s, " ")
	s = html.UnescapeString(s)
	var lines []string
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(reSpaces.ReplaceAllString(line, " "))
		if line != "" {
			lines = append(lines, line)
		}
	}
	return title, strings.Join(lines, "\n")
}

// webMeta 网页资产编目元数据
type webMeta struct {
	URL       string `json:"url"`
	Title     string `json:"title"`
	ParseMode string `json:"parseMode"`
	FetchedAt string `json:"fetchedAt"`
}

// FetchWebPage 拉取网页并抽取正文（200 校验 + 2MB 截断上限）
func (s *LlmopsServiceStruct) FetchWebPage(rawURL string) (title, text string, err error) {
	if err = ValidateWebURL(rawURL); err != nil {
		return "", "", err
	}
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return "", "", fmt.Errorf("[%d]URL 非法: %w", ErrCodeWebURLBad, err)
	}
	req.Header.Set("User-Agent", webFetchUserAgent)
	resp, err := webHTTPClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("[%d]网页拉取失败: %w", ErrCodeWebFetch, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("[%d]网页拉取失败: HTTP %d", ErrCodeWebFetch, resp.StatusCode)
	}
	ct := resp.Header.Get("Content-Type")
	if ct != "" && !strings.Contains(strings.ToLower(ct), "html") && !strings.Contains(strings.ToLower(ct), "text/plain") {
		return "", "", fmt.Errorf("[%d]不支持的响应类型: %s", ErrCodeWebFetch, ct)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, webFetchMaxBytes+1))
	if err != nil {
		return "", "", fmt.Errorf("[%d]网页读取失败: %w", ErrCodeWebFetch, err)
	}
	title, text = ExtractHTMLText(string(body))
	if strings.TrimSpace(text) == "" {
		return title, text, fmt.Errorf("[%d]网页未抽取到正文", ErrCodeWebFetch)
	}
	return title, text, nil
}

// ImportWebPage 网页采集落库：正文落 rawText(status=parsed)，metaJSON 编目 URL/标题
func (s *LlmopsServiceStruct) ImportWebPage(rawURL, providedName, tags string) (*model.LlmopsDataAsset, error) {
	title, text, err := s.FetchWebPage(rawURL)
	if err != nil {
		return nil, err
	}
	name := strings.TrimSpace(providedName)
	if name == "" {
		if title != "" {
			name = title
		} else {
			if u, err := url.Parse(rawURL); err == nil {
				name = u.Host
			} else {
				name = "未命名网页"
			}
		}
	}
	meta, _ := json.Marshal(webMeta{
		URL: rawURL, Title: title, ParseMode: "html-text",
		FetchedAt: time.Now().Format(time.RFC3339),
	})
	asset := &model.LlmopsDataAsset{
		Name:     BuildDocName(name, "page.html"),
		DataType: "webpage",
		Source:   "url",
		Status:   "parsed",
		RawText:  text,
		MetaJSON: string(meta),
		Tags:     strings.TrimSpace(tags),
		Size:     int64(len(text)),
	}
	if err := global.GVA_DB.Create(asset).Error; err != nil {
		return nil, fmt.Errorf("[%d]采集入库失败: %w", ErrCodeImportFailed, err)
	}
	return asset, nil
}
