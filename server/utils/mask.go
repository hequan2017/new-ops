package utils

import (
	"net"
	"strconv"
)

// MaskString 通用字符串脱敏（展示口径）：≥8 字符保留首尾各 2 位；4-7 保留首尾各 1 位；≤3 全遮蔽
func MaskString(s string) string {
	r := []rune(s)
	switch {
	case len(r) >= 8:
		return string(r[:2]) + "****" + string(r[len(r)-2:])
	case len(r) >= 4:
		return string(r[:1]) + "****" + string(r[len(r)-1:])
	default:
		return "****"
	}
}

// MaskIP IP 脱敏（展示口径）：IPv4 保留前两段（192.168.*.*）；非 IPv4 回退通用字符串脱敏
func MaskIP(ip string) string {
	if p := net.ParseIP(ip); p != nil {
		if v4 := p.To4(); v4 != nil {
			return strconv.Itoa(int(v4[0])) + "." + strconv.Itoa(int(v4[1])) + ".*.*"
		}
	}
	return MaskString(ip)
}

// MaskPhone 手机号/长号脱敏（展示口径）：≥7 位保留前 3 后 4（138****1234）；更短全遮蔽
func MaskPhone(phone string) string {
	r := []rune(phone)
	if len(r) < 7 {
		return "****"
	}
	return string(r[:3]) + "****" + string(r[len(r)-4:])
}

// MaskTail 保留末 keep 位、不足 keep 位原样返回（用于"仅留存末 N 位"的存储口径，不带遮蔽符）
func MaskTail(s string, keep int) string {
	r := []rune(s)
	if keep <= 0 || len(r) <= keep {
		return string(r)
	}
	return string(r[len(r)-keep:])
}
