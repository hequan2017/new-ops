// Package service crossauth：SM4 分组密码纯 Go 实现（需求 3.7-10）
// 口径：GB/T 32907-2016（128 位块、32 轮、S 盒 τ + L 变换；解密=轮密钥逆序）。
// 标准向量：key=plaintext=0123456789abcdeffedcba9876543210 → 681edf34d206965e86b3e94f536e4246。
package service

import (
	"fmt"

	"encoding/binary"
)

// 错误码（crossauth 段 2900-2999）
const (
	ErrCodeCaSm4Param    = 2915
	ErrCodeCaSm4NotFound = 2916
)

// sm4Sbox GB/T 32907 标准 S 盒
var sm4Sbox = [256]byte{
	0xd6, 0x90, 0xe9, 0xfe, 0xcc, 0xe1, 0x3d, 0xb7, 0x16, 0xb6, 0x14, 0xc2, 0x28, 0xfb, 0x2c, 0x05,
	0x2b, 0x67, 0x9a, 0x76, 0x2a, 0xbe, 0x04, 0xc3, 0xaa, 0x44, 0x13, 0x26, 0x49, 0x86, 0x06, 0x99,
	0x9c, 0x42, 0x50, 0xf4, 0x91, 0xef, 0x98, 0x7a, 0x33, 0x54, 0x0b, 0x43, 0xed, 0xcf, 0xac, 0x62,
	0xe4, 0xb3, 0x1c, 0xa9, 0xc9, 0x08, 0xe8, 0x95, 0x80, 0xdf, 0x94, 0xfa, 0x75, 0x8f, 0x3f, 0xa6,
	0x47, 0x07, 0xa7, 0xfc, 0xf3, 0x73, 0x17, 0xba, 0x83, 0x59, 0x3c, 0x19, 0xe6, 0x85, 0x4f, 0xa8,
	0x68, 0x6b, 0x81, 0xb2, 0x71, 0x64, 0xda, 0x8b, 0xf8, 0xeb, 0x0f, 0x4b, 0x70, 0x56, 0x9d, 0x35,
	0x1e, 0x24, 0x0e, 0x5e, 0x63, 0x58, 0xd1, 0xa2, 0x25, 0x22, 0x7c, 0x3b, 0x01, 0x21, 0x78, 0x87,
	0xd4, 0x00, 0x46, 0x57, 0x9f, 0xd3, 0x27, 0x52, 0x4c, 0x36, 0x02, 0xe7, 0xa0, 0xc4, 0xc8, 0x9e,
	0xea, 0xbf, 0x8a, 0xd2, 0x40, 0xc7, 0x38, 0xb5, 0xa3, 0xf7, 0xf2, 0xce, 0xf9, 0x61, 0x15, 0xa1,
	0xe0, 0xae, 0x5d, 0xa4, 0x9b, 0x34, 0x1a, 0x55, 0xad, 0x93, 0x32, 0x30, 0xf5, 0x8c, 0xb1, 0xe3,
	0x1d, 0xf6, 0xe2, 0x2e, 0x82, 0x66, 0xca, 0x60, 0xc0, 0x29, 0x23, 0xab, 0x0d, 0x53, 0x4e, 0x6f,
	0xd5, 0xdb, 0x37, 0x45, 0xde, 0xfd, 0x8e, 0x2f, 0x03, 0xff, 0x6a, 0x72, 0x6d, 0x6c, 0x5b, 0x51,
	0x8d, 0x1b, 0xaf, 0x92, 0xbb, 0xdd, 0xbc, 0x7f, 0x11, 0xd9, 0x5c, 0x41, 0x1f, 0x10, 0x5a, 0xd8,
	0x0a, 0xc1, 0x31, 0x88, 0xa5, 0xcd, 0x7b, 0xbd, 0x2d, 0x74, 0xd0, 0x12, 0xb8, 0xe5, 0xb4, 0xb0,
	0x89, 0x69, 0x97, 0x4a, 0x0c, 0x96, 0x77, 0x7e, 0x65, 0xb9, 0xf1, 0x09, 0xc5, 0x6e, 0xc6, 0x84,
	0x18, 0xf0, 0x7d, 0xec, 0x3a, 0xdc, 0x4d, 0x20, 0x79, 0xee, 0x5f, 0x3e, 0xd7, 0xcb, 0x39, 0x48,
}

// sm4FK 系统参数
var sm4FK = [4]uint32{0xa3b1bac6, 0x56aa3350, 0x677d9197, 0xb27022dc}

// sm4CK 固定参数（按 GB/T 32907 公式生成：ck=(4i+j)*7 mod 256）
var sm4CK [32]uint32

func init() {
	for i := 0; i < 32; i++ {
		sm4CK[i] = uint32((4*i+0)*7%256)<<24 | uint32((4*i+1)*7%256)<<16 |
			uint32((4*i+2)*7%256)<<8 | uint32((4*i+3)*7%256)
	}
}

// rotl 32 位循环左移
func rotl(x uint32, n uint) uint32 {
	return x<<n | x>>(32-n)
}

// sm4Tau S 盒逐字节代换
func sm4Tau(a uint32) uint32 {
	return uint32(sm4Sbox[(a>>24)&0xff])<<24 | uint32(sm4Sbox[(a>>16)&0xff])<<16 |
		uint32(sm4Sbox[(a>>8)&0xff])<<8 | uint32(sm4Sbox[a&0xff])
}

// sm4L 加密轮线性变换 L(B)=B^(<<2)^(<<10)^(<<18)^(<<24)
func sm4L(b uint32) uint32 {
	return b ^ rotl(b, 2) ^ rotl(b, 10) ^ rotl(b, 18) ^ rotl(b, 24)
}

// sm4LKey 密钥扩展线性变换 L'(B)=B^(<<13)^(<<23)
func sm4LKey(b uint32) uint32 {
	return b ^ rotl(b, 13) ^ rotl(b, 23)
}

// sm4ExpandKeys 轮密钥扩展
func sm4ExpandKeys(key []byte) [32]uint32 {
	var mk [4]uint32
	for i := 0; i < 4; i++ {
		mk[i] = binary.BigEndian.Uint32(key[i*4 : i*4+4])
	}
	var k [4]uint32
	for i := 0; i < 4; i++ {
		k[i] = mk[i] ^ sm4FK[i]
	}
	var rk [32]uint32
	for i := 0; i < 32; i++ {
		tmp := k[1] ^ k[2] ^ k[3] ^ sm4CK[i]
		rk[i] = k[0] ^ sm4LKey(sm4Tau(tmp))
		k[0], k[1], k[2], k[3] = k[1], k[2], k[3], rk[i]
	}
	return rk
}

// sm4EncryptBlock 单块加密（16 字节）
func sm4EncryptBlock(rk [32]uint32, block []byte) []byte {
	var x [4]uint32
	for i := 0; i < 4; i++ {
		x[i] = binary.BigEndian.Uint32(block[i*4 : i*4+4])
	}
	for i := 0; i < 32; i++ {
		tmp := x[1] ^ x[2] ^ x[3] ^ rk[i]
		x[0], x[1], x[2], x[3] = x[1], x[2], x[3], x[0]^sm4L(sm4Tau(tmp))
	}
	out := make([]byte, 16)
	binary.BigEndian.PutUint32(out[0:4], x[3])
	binary.BigEndian.PutUint32(out[4:8], x[2])
	binary.BigEndian.PutUint32(out[8:12], x[1])
	binary.BigEndian.PutUint32(out[12:16], x[0])
	return out
}

// SM4EncryptECB ECB 模式加密（PKCS#7 填充，返回 hex）
func SM4EncryptECB(key []byte, plaintext []byte) (string, error) {
	if len(key) != 16 {
		return "", fmt.Errorf("[%d]SM4 密钥须为 16 字节", ErrCodeCaSm4Param)
	}
	pad := 16 - len(plaintext)%16
	padded := make([]byte, len(plaintext)+pad)
	copy(padded, plaintext)
	for i := len(plaintext); i < len(padded); i++ {
		padded[i] = byte(pad)
	}
	rk := sm4ExpandKeys(key)
	hexOut := make([]byte, 0, len(padded)*2)
	for off := 0; off < len(padded); off += 16 {
		enc := sm4EncryptBlock(rk, padded[off:off+16])
		hexOut = append(hexOut, []byte(hexEncode(enc))...)
	}
	return string(hexOut), nil
}

// SM4DecryptECB ECB 模式解密（轮密钥逆序 + 去 PKCS#7）
func SM4DecryptECB(key []byte, hexText string) ([]byte, error) {
	if len(key) != 16 {
		return nil, fmt.Errorf("[%d]SM4 密钥须为 16 字节", ErrCodeCaSm4Param)
	}
	data, err := hexDecode(hexText)
	if err != nil || len(data) == 0 || len(data)%16 != 0 {
		return nil, fmt.Errorf("[%d]密文须为 16 字节倍数的 hex", ErrCodeCaSm4Param)
	}
	rk := sm4ExpandKeys(key)
	plain := make([]byte, 0, len(data))
	for off := 0; off < len(data); off += 16 {
		var x [4]uint32
		for i := 0; i < 4; i++ {
			x[i] = uint32(data[off+i*4])<<24 | uint32(data[off+i*4+1])<<16 | uint32(data[off+i*4+2])<<8 | uint32(data[off+i*4+3])
		}
		for i := 0; i < 32; i++ {
			tmp := x[1] ^ x[2] ^ x[3] ^ rk[31-i]
			x[0], x[1], x[2], x[3] = x[1], x[2], x[3], x[0]^sm4L(sm4Tau(tmp))
		}
		blk := make([]byte, 16)
		binary.BigEndian.PutUint32(blk[0:4], x[3])
		binary.BigEndian.PutUint32(blk[4:8], x[2])
		binary.BigEndian.PutUint32(blk[8:12], x[1])
		binary.BigEndian.PutUint32(blk[12:16], x[0])
		plain = append(plain, blk...)
	}
	if len(plain) == 0 {
		return nil, fmt.Errorf("[%d]密文为空", ErrCodeCaSm4Param)
	}
	pad := int(plain[len(plain)-1])
	if pad < 1 || pad > 16 || pad > len(plain) {
		return nil, fmt.Errorf("[%d]填充非法", ErrCodeCaSm4Param)
	}
	return plain[:len(plain)-pad], nil
}

// hexEncode/hexDecode 小工具（避免与 asset/crypto 依赖）
func hexEncode(b []byte) string {
	const digits = "0123456789abcdef"
	out := make([]byte, len(b)*2)
	for i, v := range b {
		out[i*2] = digits[v>>4]
		out[i*2+1] = digits[v&0x0f]
	}
	return string(out)
}

func hexDecode(s string) ([]byte, error) {
	if len(s)%2 != 0 {
		return nil, fmt.Errorf("hex 长度须为偶数")
	}
	out := make([]byte, len(s)/2)
	for i := 0; i < len(out); i++ {
		hi, ok1 := hexNibble(s[i*2])
		lo, ok2 := hexNibble(s[i*2+1])
		if !ok1 || !ok2 {
			return nil, fmt.Errorf("非法 hex 字符")
		}
		out[i] = hi<<4 | lo
	}
	return out, nil
}

func hexNibble(c byte) (byte, bool) {
	switch {
	case c >= '0' && c <= '9':
		return c - '0', true
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10, true
	case c >= 'A' && c <= 'F':
		return c - 'A' + 10, true
	}
	return 0, false
}
