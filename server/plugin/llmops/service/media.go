// Package service llmops：音视频资产（元数据编目 + ASR 转写适配器）与时间范围解析
// 口径：媒体文件仅编目元数据（大小/MIME/类别），转写走 AsrAdapter 可替换点——
// 默认 NoopAsrAdapter 返回未配置错误；转写成功后正文落 rawText 并记 parse 事件。
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
)

// 错误码（llmops 段续用）
const (
	ErrCodeAsrNotConfigured = 2223
	ErrCodeMediaExtBad      = 2224
	ErrCodeMediaNotTranscribable = 2225
	ErrCodeTimeRange        = 2226
)

// mediaKindExts 扩展名→媒体类别
var mediaKindExts = map[string]string{
	".png": "image", ".jpg": "image", ".jpeg": "image", ".gif": "image", ".webp": "image", ".bmp": "image", ".svg": "image",
	".mp3": "audio", ".wav": "audio", ".flac": "audio", ".m4a": "audio", ".aac": "audio", ".ogg": "audio",
	".mp4": "video", ".mov": "video", ".avi": "video", ".mkv": "video", ".webm": "video", ".flv": "video",
}

// MediaKindOf 扩展名→媒体类别（image/audio/video，非媒体返回空串）
func MediaKindOf(ext string) string {
	return mediaKindExts[strings.ToLower(ext)]
}

// AsrAdapter ASR 转写适配器（可替换点：后续接入模型服务或外部转写）
type AsrAdapter interface {
	Transcribe(ctx context.Context, content []byte, filename string) (string, error)
}

// NoopAsrAdapter 默认适配器：未配置转写服务
type NoopAsrAdapter struct{}

func (*NoopAsrAdapter) Transcribe(context.Context, []byte, string) (string, error) {
	return "", fmt.Errorf("[%d]ASR 转写服务未配置", ErrCodeAsrNotConfigured)
}

// Asr 当前 ASR 适配器（替换点）
var Asr AsrAdapter = &NoopAsrAdapter{}

// mediaMeta 媒体编目元数据
type mediaMeta struct {
	Filename  string `json:"filename"`
	Ext       string `json:"ext"`
	MIME      string `json:"mime"`
	Kind      string `json:"kind"`
	ParseMode string `json:"parseMode"`
	Note      string `json:"note"`
}

// ImportMedia 音视频/图片上传：仅元数据编目（status=draft，转写/OCR 待适配器）
func (s *LlmopsServiceStruct) ImportMedia(fh *multipart.FileHeader, providedName, tags string) (*model.LlmopsDataAsset, error) {
	if fh == nil || fh.Size == 0 {
		return nil, fmt.Errorf("[%d]上传文件为空", ErrCodeImportNoFile)
	}
	if fh.Size > ImportMaxBytes {
		return nil, fmt.Errorf("[%d]文件超过 %dMB 上限", ErrCodeImportTooBig, ImportMaxBytes>>20)
	}
	ext := strings.ToLower(filepath.Ext(fh.Filename))
	kind := MediaKindOf(ext)
	if kind == "" {
		return nil, fmt.Errorf("[%d]非媒体文件类型: %s", ErrCodeMediaExtBad, ext)
	}
	// 不读取内容，仅取元数据（大文件友好）
	meta, _ := json.Marshal(mediaMeta{
		Filename: fh.Filename, Ext: ext, MIME: mimeExts[ext], Kind: kind,
		ParseMode: "metadata-only",
		Note:      "媒体元数据编目，转写/OCR 待适配器扩展",
	})
	asset := &model.LlmopsDataAsset{
		Name:     BuildDocName(providedName, fh.Filename),
		DataType: kind,
		Source:   "upload",
		Status:   "draft",
		MetaJSON: string(meta),
		Tags:     strings.TrimSpace(tags),
		Size:     fh.Size,
	}
	if err := global.GVA_DB.Create(asset).Error; err != nil {
		return nil, fmt.Errorf("[%d]编目入库失败: %w", ErrCodeImportFailed, err)
	}
	recordAssetEvent(global.GVA_DB, asset.ID, "collect", "media:"+fh.Filename)
	return asset, nil
}

// TranscribeMedia 音频/视频转写：调 ASR 适配器，成功后正文落 rawText(status=parsed)
func (s *LlmopsServiceStruct) TranscribeMedia(assetId uint) (*model.LlmopsDataAsset, error) {
	var asset model.LlmopsDataAsset
	if err := global.GVA_DB.First(&asset, assetId).Error; err != nil {
		return nil, fmt.Errorf("[%d]数据资产不存在", ErrCodeAssetNotFound)
	}
	if asset.DataType != "audio" && asset.DataType != "video" {
		return nil, fmt.Errorf("[%d]仅音频/视频资产支持转写（当前 %s）", ErrCodeMediaNotTranscribable, asset.DataType)
	}
	text, err := Asr.Transcribe(context.Background(), nil, asset.Name)
	if err != nil {
		return nil, err
	}
	if err := global.GVA_DB.Model(&asset).Updates(map[string]any{
		"raw_text": text, "status": "parsed",
	}).Error; err != nil {
		return nil, err
	}
	recordAssetEvent(global.GVA_DB, assetId, "parse", "ASR 转写完成")
	asset.RawText = text
	asset.Status = "parsed"
	return &asset, nil
}

// ParseTimeRange 时间范围解析（支持 RFC3339 / 2006-01-02 15:04:05 / 2006-01-02，begin>end 报错）
func ParseTimeRange(begin, end string) (time.Time, time.Time, error) {
	var zero time.Time
	if begin == "" && end == "" {
		return zero, zero, nil
	}
	parse := func(s, field string) (time.Time, error) {
		for _, layout := range []string{time.RFC3339, "2006-01-02 15:04:05", "2006-01-02"} {
			if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
				return t, nil
			}
		}
		return zero, fmt.Errorf("[%d]%s 时间格式非法: %s", ErrCodeTimeRange, field, s)
	}
	bt, et := zero, zero
	var err error
	if begin != "" {
		if bt, err = parse(begin, "begin"); err != nil {
			return zero, zero, err
		}
	}
	if end != "" {
		if et, err = parse(end, "end"); err != nil {
			return zero, zero, err
		}
	}
	if !bt.IsZero() && !et.IsZero() && bt.After(et) {
		return zero, zero, fmt.Errorf("[%d]begin 不得晚于 end", ErrCodeTimeRange)
	}
	return bt, et, nil
}
