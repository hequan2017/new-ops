// Package service ziqiang：aHash 单测（标准库编解码合成图，无 DB）
package service

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
)

// encodeTestPNG 合成 16×16 图像：上半白下半黑（可整除 8 便于采样确定性）
func encodeTestPNG(t *testing.T, invert bool) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 16, 16))
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			white := y < 8
			if invert {
				white = !white
			}
			if white {
				img.Set(x, y, color.White)
			} else {
				img.Set(x, y, color.Black)
			}
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("合成图编码失败: %v", err)
	}
	return buf.Bytes()
}

// TestAHashDeterministic 同图同哈希、反色不同哈希、汉明距离语义
func TestAHashDeterministic(t *testing.T) {
	data := encodeTestPNG(t, false)
	h1, err := AHash(data)
	if err != nil {
		t.Fatalf("解码应成功: %v", err)
	}
	h2, _ := AHash(data)
	if h1 != h2 {
		t.Fatal("同图哈希应一致")
	}
	if hex := AHashHex(h1); len(hex) != 16 {
		t.Fatalf("哈希十六进制应 16 位: %s", hex)
	}
	hInv, _ := AHash(encodeTestPNG(t, true))
	d := HammingDistance(h1, hInv)
	if d <= 0 || d > 64 {
		t.Fatalf("反色距离应在 (0,64]: %d", d)
	}
	if HammingDistance(h1, h1) != 0 {
		t.Fatal("自身距离应为 0")
	}
	if s := HashSimilarity(h1, h1); s != 1 {
		t.Fatalf("自身相似度应为 1: %v", s)
	}
}

// TestAHashErrors 坏数据与过小图
func TestAHashErrors(t *testing.T) {
	if _, err := AHash([]byte("not an image")); err == nil {
		t.Fatal("非图像应报解码失败")
	}
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	if _, err := AHash(buf.Bytes()); err == nil {
		t.Fatal("过小图应报错")
	}
}

// TestParseHashHex 十六进制解析
func TestParseHashHex(t *testing.T) {
	h, err := parseHashHex("ffffffffffffffff")
	if err != nil || h != ^uint64(0) {
		t.Fatalf("解析不符: %v %v", h, err)
	}
	if _, err := parseHashHex("zz"); err == nil {
		t.Fatal("非法十六进制应报错")
	}
	if _, err := parseHashHex("abcd"); err == nil {
		t.Fatal("长度不足应报错")
	}
}
