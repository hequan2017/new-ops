// Package service ziqiang：区块链存证（需求 3.5-8）
// 口径：存证表 dataHash/prevHash/height/timestamp；SHA256 链式追加
// （BlockHash=f(height|ts|dataHash|prevHash)，创世 prevHash 全 0）；完整性校验
// API 逐块重算三重断言（块哈希/前链指针/原文哈希）定位首个篡改块，纯函数可测。
package service

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/errx"
	"github.com/hequan2017/new-ops/server/plugin/ziqiang/model"
)

// 错误码（ziqiang 段 2700-2799）
const (
	ErrCodeEvidenceParam    = 2714
	ErrCodeEvidenceNotFound = 2715
)

// GenesisPrevHash 创世块前链指针（64 个 0）
var GenesisPrevHash = strings.Repeat("0", 64)

// DataFingerprint 原文 SHA256（hex）
func DataFingerprint(payload string) string {
	sum := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(sum[:])
}

// EvidenceFingerprint 块哈希=f(height|ts|dataHash|prevHash)（纯函数，测试锚点）
func EvidenceFingerprint(dataHash, prevHash string, height, ts int64) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%d|%d|%s|%s", height, ts, dataHash, prevHash)))
	return hex.EncodeToString(sum[:])
}

// chainAppend 链式追加（纯函数，测试锚点）：height=prev+1，创世块 prevHash 全 0
func chainAppend(prev *model.ZqEvidence, payload string, ts int64) model.ZqEvidence {
	dataHash := DataFingerprint(payload)
	var height int64
	prevHash := GenesisPrevHash
	if prev != nil {
		height = prev.Height + 1
		prevHash = prev.BlockHash
	}
	return model.ZqEvidence{
		Payload: payload, DataHash: dataHash, PrevHash: prevHash,
		BlockHash: EvidenceFingerprint(dataHash, prevHash, height, ts),
		Height:    height, Timestamp: ts,
	}
}

// chainIssue 完整性校验问题项
type chainIssue struct {
	Height int64  `json:"height"`
	ID     uint   `json:"id"`
	Reason string `json:"reason"`
}

// verifyChain 链完整性校验（纯函数，测试锚点）：按块高顺序逐块重算——
// ①块哈希可复算 ②前链指针衔接（创世全 0）③dataHash 与原文一致。
func verifyChain(blocks []model.ZqEvidence) []chainIssue {
	issues := make([]chainIssue, 0)
	for i := range blocks {
		b := blocks[i]
		if b.PrevHash != GenesisPrevHash && (i == 0 || b.PrevHash != blocks[i-1].BlockHash) {
			issues = append(issues, chainIssue{Height: b.Height, ID: b.ID, Reason: "前链指针断裂"})
		}
		if b.PrevHash == GenesisPrevHash && i != 0 {
			issues = append(issues, chainIssue{Height: b.Height, ID: b.ID, Reason: "非创世块出现全 0 前链指针"})
		}
		if want := EvidenceFingerprint(b.DataHash, b.PrevHash, b.Height, b.Timestamp); b.BlockHash != want {
			issues = append(issues, chainIssue{Height: b.Height, ID: b.ID, Reason: "块哈希重算不符"})
		}
		if want := DataFingerprint(b.Payload); b.DataHash != want {
			issues = append(issues, chainIssue{Height: b.Height, ID: b.ID, Reason: "原文哈希不符（原文被篡改）"})
		}
	}
	return issues
}

// AppendEvidence 追加存证块（链式：取当前最高块，height+1，prevHash 接其 BlockHash）
func (s *ZiqiangService) AppendEvidence(payload, bizKey string) (*model.ZqEvidence, error) {
	if strings.TrimSpace(payload) == "" {
		return nil, zqErr(ErrCodeEvidenceParam, "存证原文必填")
	}
	var prev model.ZqEvidence
	_ = s.repo.LatestEvidence(s.dbh(), &prev) // 空链时 prev 为零值→创世块
	block := chainAppend(&prev, payload, time.Now().Unix())
	block.BizKey = bizKey
	if err := s.repo.CreateEvidence(s.dbh(), &block); err != nil {
		return nil, errx.Wrap(ErrCodeZqStore, "落库存证块失败", err)
	}
	return &block, nil
}

// ChainReport 完整性校验报告
type ChainReport struct {
	OK     bool         `json:"ok"`
	Blocks int          `json:"blocks"`
	Issues []chainIssue `json:"issues"`
}

// VerifyChain 链完整性校验（篡改检测报告）
func (s *ZiqiangService) VerifyChain() (*ChainReport, error) {
	blocks, err := s.repo.EvidenceChainScan(s.dbh(), "")
	if err != nil {
		return nil, errx.Wrap(ErrCodeZqStore, "扫描存证链失败", err)
	}
	issues := verifyChain(blocks)
	return &ChainReport{OK: len(issues) == 0, Blocks: len(blocks), Issues: issues}, nil
}

// ChainHead 链头摘要
type ChainHead struct {
	Height    int64  `json:"height"`    // 最新块高（空链 -1）
	BlockHash string `json:"blockHash"` // 最新块哈希（空链为创世前链指针）
	Total     int64  `json:"total"`
	LastAt    int64  `json:"lastAt"` // 最新块时间戳
}

// ChainHead 链头摘要（最新块高/块哈希/总量）
func (s *ZiqiangService) ChainHead() (*ChainHead, error) {
	total, err := s.repo.CountEvidence(s.dbh())
	if err != nil {
		return nil, errx.Wrap(ErrCodeZqStore, "统计存证块失败", err)
	}
	head := &ChainHead{Height: -1, BlockHash: GenesisPrevHash, Total: total}
	var last model.ZqEvidence
	if err := s.repo.LatestEvidence(s.dbh(), &last); err == nil {
		head.Height = last.Height
		head.BlockHash = last.BlockHash
		head.LastAt = last.Timestamp
	}
	return head, nil
}

// VerifyBizChain 业务号局部核验（该业务全部块按块高重算，口径同全链校验）
func (s *ZiqiangService) VerifyBizChain(bizKey string) (*ChainReport, error) {
	if strings.TrimSpace(bizKey) == "" {
		return nil, zqErr(ErrCodeEvidenceParam, "业务号必填")
	}
	blocks, err := s.repo.EvidenceChainScan(s.dbh(), bizKey)
	if err != nil {
		return nil, errx.Wrap(ErrCodeZqStore, "扫描存证链失败", err)
	}
	issues := verifyChain(blocks)
	return &ChainReport{OK: len(issues) == 0, Blocks: len(blocks), Issues: issues}, nil
}

// GetEvidenceList 存证列表（业务号过滤，块高升序分页）
func (s *ZiqiangService) GetEvidenceList(bizKey string, page, pageSize int) ([]model.ZqEvidence, int64, error) {
	list, total, err := s.repo.EvidenceList(s.dbh(), bizKey, page, pageSize)
	if err != nil {
		return nil, 0, errx.Wrap(ErrCodeZqStore, "查询存证列表失败", err)
	}
	return list, total, nil
}

// GetEvidence 存证详情
func (s *ZiqiangService) GetEvidence(id uint) (*model.ZqEvidence, error) {
	b, err := s.repo.FindEvidence(s.dbh(), id)
	if err != nil {
		return nil, zqErr(ErrCodeEvidenceNotFound, "存证不存在")
	}
	return b, nil
}
