// Package service ziqiang：图像鉴权语料与 aHash（需求 3.5-2）
// 口径：aHash 8×8 均值阈值感知哈希（image/jpeg、image/png 标准库解码，
// 最近邻降采样 + ITU-R BT.601 灰度）；比对=汉明距离（bits.OnesCount64）+ 相似度 1-d/64。
package service

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/binary"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"math/bits"
	"strings"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/ziqiang/model"
)

// 错误码（ziqiang 段 2700-2799）
const (
	ErrCodeZqImgParam   = 2704
	ErrCodeZqImgNotFound = 2705
	ErrCodeZqImgDecode  = 2706
)

// AHash 8×8 均值阈值感知哈希（纯函数，测试锚点）
func AHash(data []byte) (uint64, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return 0, fmt.Errorf("[%d]图像解码失败: %w", ErrCodeZqImgDecode, err)
	}
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	if w < 8 || h < 8 {
		return 0, fmt.Errorf("[%d]图像尺寸过小（≥8×8）", ErrCodeZqImgDecode)
	}
	gray := make([]float64, 64)
	for gy := 0; gy < 8; gy++ {
		sy := bounds.Min.Y + (gy+1)*h/8 - h/16
		for gx := 0; gx < 8; gx++ {
			sx := bounds.Min.X + (gx+1)*w/8 - w/16
			r, g, b, _ := img.At(sx, sy).RGBA()
			gray[gy*8+gx] = 0.299*float64(r>>8) + 0.587*float64(g>>8) + 0.114*float64(b>>8)
		}
	}
	mean := 0.0
	for _, v := range gray {
		mean += v
	}
	mean /= 64
	var hash uint64
	for i, v := range gray {
		if v > mean {
			hash |= 1 << uint(63-i)
		}
	}
	return hash, nil
}

// AHashHex 哈希十六进制（纯函数）
func AHashHex(h uint64) string {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, h)
	return hex.EncodeToString(b)
}

// HammingDistance 汉明距离（纯函数，测试锚点）
func HammingDistance(a, b uint64) int {
	return bits.OnesCount64(a ^ b)
}

// HashSimilarity 相似度 = 1 - d/64（纯函数）
func HashSimilarity(a, b uint64) float64 {
	return float64(int((1-float64(HammingDistance(a, b))/64)*10000+0.5)) / 10000
}

// parseHashHex 十六进制→uint64
func parseHashHex(s string) (uint64, error) {
	raw, err := hex.DecodeString(strings.TrimSpace(s))
	if err != nil || len(raw) != 8 {
		return 0, fmt.Errorf("[%d]哈希须为 16 位十六进制", ErrCodeZqImgParam)
	}
	return binary.BigEndian.Uint64(raw), nil
}

// SaveImageCorpus 登记图像鉴权语料（imageData base64 可选，传入即算 aHash）
func (s *ZiqiangService) SaveImageCorpus(in *model.ZqImageCorpus, imageDataB64 string) (*model.ZqImageCorpus, error) {
	if strings.TrimSpace(in.Name) == "" || strings.TrimSpace(in.Path) == "" {
		return nil, fmt.Errorf("[%d]名称与路径必填", ErrCodeZqImgParam)
	}
	if strings.TrimSpace(imageDataB64) != "" {
		raw, err := base64.StdEncoding.DecodeString(imageDataB64)
		if err != nil {
			return nil, fmt.Errorf("[%d]imageData 须为 base64: %w", ErrCodeZqImgParam, err)
		}
		h, err := AHash(raw)
		if err != nil {
			return nil, err
		}
		in.HashHex = AHashHex(h)
	}
	if err := global.GVA_DB.Create(in).Error; err != nil {
		return nil, err
	}
	return in, nil
}

// GetImageCorpora 图像语料列表（图库/关键字过滤，倒序分页）
func (s *ZiqiangService) GetImageCorpora(gallery, keyword string, page, pageSize int) ([]model.ZqImageCorpus, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}
	db := global.GVA_DB.Model(&model.ZqImageCorpus{})
	if gallery != "" {
		db = db.Where("gallery = ?", gallery)
	}
	if kw := strings.TrimSpace(keyword); kw != "" {
		db = db.Where("name LIKE ?", "%"+kw+"%")
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.ZqImageCorpus
	err := db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// CompareResult 比对结果
type CompareResult struct {
	Distance   int     `json:"distance"`
	Similarity float64 `json:"similarity"`
}

// CompareHashes 十六进制哈希两两比对（纯编排）
func (s *ZiqiangService) CompareHashes(hashA, hashB string) (*CompareResult, error) {
	a, err := parseHashHex(hashA)
	if err != nil {
		return nil, err
	}
	b, err := parseHashHex(hashB)
	if err != nil {
		return nil, err
	}
	return &CompareResult{Distance: HammingDistance(a, b), Similarity: HashSimilarity(a, b)}, nil
}

// CompareWithCorpus 与库内语料哈希比对
func (s *ZiqiangService) CompareWithCorpus(corpusId uint, hashHex string) (*CompareResult, error) {
	var corpus model.ZqImageCorpus
	if err := global.GVA_DB.First(&corpus, corpusId).Error; err != nil {
		return nil, fmt.Errorf("[%d]图像语料不存在", ErrCodeZqImgNotFound)
	}
	if corpus.HashHex == "" {
		return nil, fmt.Errorf("[%d]该语料未计算 aHash", ErrCodeZqImgParam)
	}
	return s.CompareHashes(hashHex, corpus.HashHex)
}
