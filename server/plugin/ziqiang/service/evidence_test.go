// Package service ziqiang：区块链存证单测（纯函数，无 DB）
package service

import (
	"strings"
	"testing"

	"github.com/hequan2017/new-ops/server/plugin/ziqiang/model"
)

// buildChain 纯函数造链（模拟 AppendEvidence 的链式口径）
func buildChain(payloads []string) []model.ZqEvidence {
	blocks := make([]model.ZqEvidence, 0, len(payloads))
	var prev *model.ZqEvidence
	for i, p := range payloads {
		b := chainAppend(prev, p, int64(1700000000+i))
		b.ID = uint(i + 1)
		blocks = append(blocks, b)
		prev = &blocks[len(blocks)-1]
	}
	return blocks
}

// TestChainAppendLinkage 链式追加：创世全 0、指针衔接、块高递增、哈希 64 位 hex
func TestChainAppendLinkage(t *testing.T) {
	blocks := buildChain([]string{"专利文本A", "开源协议文本B", "论文文本C"})
	if len(blocks) != 3 {
		t.Fatalf("应有 3 块: %d", len(blocks))
	}
	if blocks[0].PrevHash != GenesisPrevHash {
		t.Fatalf("创世块前链指针应为全 0: %s", blocks[0].PrevHash)
	}
	for i := 1; i < len(blocks); i++ {
		if blocks[i].PrevHash != blocks[i-1].BlockHash {
			t.Fatalf("块 %d 前链指针应衔接上一块: %s != %s", i, blocks[i].PrevHash, blocks[i-1].BlockHash)
		}
		if blocks[i].Height != blocks[i-1].Height+1 {
			t.Fatalf("块高应递增: %d", blocks[i].Height)
		}
	}
	if want := DataFingerprint("专利文本A"); blocks[0].DataHash != want {
		t.Fatalf("dataHash 应为原文 SHA256: %s", blocks[0].DataHash)
	}
	for _, b := range blocks {
		if len(b.BlockHash) != 64 {
			t.Fatalf("块哈希应为 64 位 hex: %s", b.BlockHash)
		}
	}
}

// TestEvidenceFingerprintDeterminism 同参同哈希、异参异哈希
func TestEvidenceFingerprintDeterminism(t *testing.T) {
	h1 := EvidenceFingerprint("aa", GenesisPrevHash, 0, 1700000000)
	if h1 != EvidenceFingerprint("aa", GenesisPrevHash, 0, 1700000000) {
		t.Fatal("同参数应得同哈希")
	}
	for _, mut := range []struct {
		dataHash, prevHash string
		height, ts         int64
	}{{"ab", GenesisPrevHash, 0, 1700000000}, {"aa", "11", 0, 1700000000}, {"aa", GenesisPrevHash, 1, 1700000000}, {"aa", GenesisPrevHash, 0, 1700000001}} {
		if EvidenceFingerprint(mut.dataHash, mut.prevHash, mut.height, mut.ts) == h1 {
			t.Fatalf("参数变动 %+v 哈希应不同", mut)
		}
	}
}

// TestVerifyChainIntact 完整链零问题
func TestVerifyChainIntact(t *testing.T) {
	if issues := verifyChain(buildChain([]string{"A", "B", "C", "D"})); len(issues) != 0 {
		t.Fatalf("完整链应零问题: %+v", issues)
	}
	if issues := verifyChain(buildChain([]string{"只此一块"})); len(issues) != 0 {
		t.Fatalf("单块链应零问题: %+v", issues)
	}
}

// TestVerifyChainTamperDetection 三类篡改定位：原文改写/块哈希改写/指针断裂
func TestVerifyChainTamperDetection(t *testing.T) {
	blocks := buildChain([]string{"人创文本", "AIGC 文本", "存证三"})

	// ①原文被篡改 → dataHash 重算不符
	tampered := append(blocks[:0:0], blocks...)
	tampered[1].Payload = "被替换的文本"
	issues := verifyChain(tampered)
	if len(issues) != 1 || issues[0].Height != 1 || !strings.Contains(issues[0].Reason, "原文哈希不符") {
		t.Fatalf("应定位块 1 原文篡改: %+v", issues)
	}

	// ②块哈希被篡改 → 块哈希重算不符（并引发后块指针断裂）
	tampered = append(blocks[:0:0], blocks...)
	tampered[1].BlockHash = strings.Repeat("f", 64)
	issues = verifyChain(tampered)
	if len(issues) == 0 || issues[0].Height != 1 || !strings.Contains(issues[0].Reason, "块哈希重算不符") {
		t.Fatalf("应定位块 1 哈希篡改: %+v", issues)
	}
	if len(issues) < 2 || issues[1].Height != 2 || !strings.Contains(issues[1].Reason, "前链指针断裂") {
		t.Fatalf("后块应连锁报指针断裂: %+v", issues)
	}

	// ③非创世块出现全 0 前链指针
	tampered = append(blocks[:0:0], blocks...)
	tampered[1].PrevHash = GenesisPrevHash
	tampered[1].BlockHash = EvidenceFingerprint(tampered[1].DataHash, GenesisPrevHash, tampered[1].Height, tampered[1].Timestamp)
	issues = verifyChain(tampered)
	if len(issues) == 0 || !strings.Contains(issues[0].Reason, "全 0 前链指针") {
		t.Fatalf("非创世全 0 指针应报问题: %+v", issues)
	}
}
