// Package service llmops：文档采集（上传→解析正文→编目落库）
// 解析口径：文本类扩展名直接读取为 UTF-8 正文（status=parsed）；pdf/docx 等二进制类型
// 暂不解析正文，仅编目元数据并保持 draft（正文抽取待后续解析适配器扩展）。
package service

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"gorm.io/gorm"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
)

// 错误码（llmops 段续用）
const (
	ErrCodeImportNoFile  = 2202
	ErrCodeImportTooBig  = 2203
	ErrCodeImportEmpty   = 2204
	ErrCodeImportFailed  = 2205
	ImportMaxBytes int64 = 20 << 20 // 20MB 上传上限
)

// textExts 文本类扩展名：正文直接落 rawText
var textExts = map[string]bool{
	".txt": true, ".md": true, ".csv": true, ".json": true,
	".yaml": true, ".yml": true, ".xml": true, ".log": true,
	".html": true, ".htm": true, ".sql": true, ".go": true, ".py": true,
}

// mimeExts 扩展名→MIME 编目映射（编目元数据用，未列出按 application/octet-stream）
var mimeExts = map[string]string{
	".txt": "text/plain", ".md": "text/markdown", ".csv": "text/csv",
	".json": "application/json", ".yaml": "application/yaml", ".yml": "application/yaml",
	".xml": "application/xml", ".html": "text/html", ".htm": "text/html",
	".pdf": "application/pdf", ".png": "image/png", ".jpg": "image/jpeg",
	".jpeg": "image/jpeg", ".gif": "image/gif", ".mp3": "audio/mpeg",
	".wav": "audio/wav", ".mp4": "video/mp4", ".zip": "application/zip",
}

// ExtParseMode 扩展名→解析模式：text（正文直读）/ binary（仅编目）
func ExtParseMode(ext string) string {
	if textExts[strings.ToLower(ext)] {
		return "text"
	}
	return "binary"
}

// BuildDocName 归一化资产名：优先用户命名，否则取文件名去扩展；去空白、限长 128
func BuildDocName(provided, filename string) string {
	name := strings.TrimSpace(provided)
	if name == "" {
		name = strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
	}
	name = strings.TrimSpace(name)
	r := []rune(name)
	if len(r) > 128 {
		name = string(r[:128])
	}
	if name == "" {
		name = "未命名文档"
	}
	return name
}

// docMeta 编目元数据
type docMeta struct {
	Filename  string `json:"filename"`
	Ext       string `json:"ext"`
	MIME      string `json:"mime"`
	ParseMode string `json:"parseMode"`
	Note      string `json:"note,omitempty"`
}

// ImportDoc 上传采集：读取文件→文本类落 rawText(status=parsed)，二进制仅编目(status=draft)
func (s *LlmopsServiceStruct) ImportDoc(fh *multipart.FileHeader, providedName, tags string) (*model.LlmopsDataAsset, error) {
	if fh == nil || fh.Size == 0 {
		return nil, fmt.Errorf("[%d]上传文件为空", ErrCodeImportNoFile)
	}
	if fh.Size > ImportMaxBytes {
		return nil, fmt.Errorf("[%d]文件超过 %dMB 上限", ErrCodeImportTooBig, ImportMaxBytes>>20)
	}
	f, err := fh.Open()
	if err != nil {
		return nil, fmt.Errorf("[%d]读取上传文件失败: %w", ErrCodeImportFailed, err)
	}
	defer f.Close()
	content, err := io.ReadAll(io.LimitReader(f, ImportMaxBytes+1))
	if err != nil {
		return nil, fmt.Errorf("[%d]读取上传文件失败: %w", ErrCodeImportFailed, err)
	}
	if len(content) == 0 {
		return nil, fmt.Errorf("[%d]上传文件内容为空", ErrCodeImportEmpty)
	}

	ext := strings.ToLower(filepath.Ext(fh.Filename))
	mode := ExtParseMode(ext)
	meta := docMeta{Filename: fh.Filename, Ext: ext, MIME: mimeExts[ext], ParseMode: mode}
	raw := ""
	status := "draft"
	if mode == "text" {
		if !utf8.Valid(content) {
			meta.Note = "非 UTF-8 编码，仅编目不解析正文"
		} else {
			raw = string(content)
			status = "parsed"
		}
	} else {
		meta.Note = "二进制类型暂不解析正文，仅编目元数据（正文抽取待解析适配器扩展）"
	}
	metaJSON, _ := json.Marshal(meta)

	asset := &model.LlmopsDataAsset{
		Name:     BuildDocName(providedName, fh.Filename),
		DataType: "doc",
		Source:   "upload",
		Status:   status,
		RawText:  raw,
		MetaJSON: string(metaJSON),
		Tags:     strings.TrimSpace(tags),
		Size:     fh.Size,
	}
	if err := global.GVA_DB.Create(asset).Error; err != nil {
		return nil, fmt.Errorf("[%d]采集入库失败: %w", ErrCodeImportFailed, err)
	}
	return asset, nil
}

// GetAssetDocContent 提取文档正文（供后续切块/向量化复用；未解析返回错误）
func (s *LlmopsServiceStruct) GetAssetDocContent(id uint) (string, error) {
	var asset model.LlmopsDataAsset
	err := global.GVA_DB.First(&asset, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", fmt.Errorf("[%d]数据资产不存在", ErrCodeAssetNotFound)
		}
		return "", err
	}
	if asset.RawText == "" {
		return "", fmt.Errorf("[%d]资产无已解析正文（状态 %s）", ErrCodeImportEmpty, asset.Status)
	}
	return asset.RawText, nil
}
