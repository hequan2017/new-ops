// Package service llmops 网页采集单测（正文抽取/协议白名单/httptest 拉取全链）
package service

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const sampleHTML = `<!doctype html>
<html><head><title>白泽 &amp; 平台介绍</title>
<style>.a{color:red}</style><script>var x="ignored";</script></head>
<body><!-- 注释不出现 -->
<h1>标题一</h1>
<p>第一段&nbsp;&nbsp;正文 <b>加粗</b></p>
<div>第二段<br>换行内容</div>
<script type="text/javascript">if(1&lt;2){alert("no")}</script>
<ul><li>条目甲</li><li>条目乙</li></ul>
</body></html>`

// TestExtractHTMLText 抽取契约：标题反解实体、script/style/注释剔除、块级换行、行内空白归一
func TestExtractHTMLText(t *testing.T) {
	title, text := ExtractHTMLText(sampleHTML)
	if title != "白泽 & 平台介绍" {
		t.Fatalf("title=%q", title)
	}
	for _, want := range []string{"标题一", "第一段 正文 加粗", "第二段", "换行内容", "条目甲", "条目乙"} {
		if !strings.Contains(text, want) {
			t.Fatalf("正文缺 %q\n%s", want, text)
		}
	}
	for _, banned := range []string{".a{color:red}", `var x=`, "ignored", "alert", "注释不出现", "<", "&nbsp;"} {
		if strings.Contains(text, banned) {
			t.Fatalf("正文不应含 %q\n%s", banned, text)
		}
	}
	// 块级元素应各自成行（li 换行断言结构而非拼接）
	if !strings.Contains(text, "\n条目乙") {
		t.Fatalf("块级标签应换行\n%s", text)
	}
}

// TestValidateWebURL 协议白名单（SSRF 面收缩：非 http(s) 一律拒绝）
func TestValidateWebURL(t *testing.T) {
	for _, ok := range []string{"http://a.com/x", "https://a.com/x?y=1", " http://a.com "} {
		if err := ValidateWebURL(ok); err != nil {
			t.Fatalf("%q 应放行: %v", ok, err)
		}
	}
	for _, bad := range []string{"", "notaurl", "ftp://a.com", "file:///etc/passwd", "javascript:alert(1)", "data:text/html,x", "http://"} {
		if err := ValidateWebURL(bad); err == nil {
			t.Fatalf("%q 应拒绝", bad)
		}
	}
}

// TestFetchWebPage_HttpTest httptest 全链：抽取正文 + 非 200 报错 + 空正文报错
func TestFetchWebPage_HttpTest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.UserAgent() != webFetchUserAgent {
			t.Errorf("UA 未携带: %q", r.UserAgent())
		}
		_, _ = w.Write([]byte(sampleHTML))
	}))
	defer srv.Close()

	title, text, err := LlmopsService.FetchWebPage(srv.URL)
	if err != nil {
		t.Fatalf("拉取失败: %v", err)
	}
	if title != "白泽 & 平台介绍" || !strings.Contains(text, "第一段") {
		t.Fatalf("抽取异常 title=%q text=%q", title, text)
	}

	errSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer errSrv.Close()
	if _, _, err := LlmopsService.FetchWebPage(errSrv.URL); err == nil || !strings.Contains(err.Error(), "403") {
		t.Fatalf("非 200 应报错: %v", err)
	}

	emptySrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("<html><body>  </body></html>"))
	}))
	defer emptySrv.Close()
	if _, _, err := LlmopsService.FetchWebPage(emptySrv.URL); err == nil || !strings.Contains(err.Error(), "正文") {
		t.Fatalf("空正文应报错: %v", err)
	}

	// 非 html 响应类型拒绝
	binSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		_, _ = w.Write([]byte("bin"))
	}))
	defer binSrv.Close()
	if _, _, err := LlmopsService.FetchWebPage(binSrv.URL); err == nil || !strings.Contains(err.Error(), "响应类型") {
		t.Fatalf("非 html 类型应报错: %v", err)
	}
}
