package utils

import "testing"

func TestMaskString(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"长串保留首尾各2", "abcdef123456", "ab****56"},
		{"中串保留首尾各1", "abcd123", "a****3"},
		{"短串全遮蔽", "abc", "****"},
		{"空串全遮蔽", "", "****"},
		{"中文按rune计", "白泽运维平台啊", "白****啊"},
	}
	for _, c := range cases {
		if got := MaskString(c.in); got != c.want {
			t.Fatalf("%s: want %q got %q", c.name, c.want, got)
		}
	}
}

func TestMaskIP(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"IPv4保留前两段", "192.168.1.100", "192.168.*.*"},
		{"非法长串回退通用脱敏", "not-an-ip", "no****ip"},
		{"非法短串回退全遮蔽", "x", "****"},
		{"IPv6回退通用脱敏", "fe80:dead:beef::1", "fe****:1"},
	}
	for _, c := range cases {
		if got := MaskIP(c.in); got != c.want {
			t.Fatalf("%s: want %q got %q", c.name, c.want, got)
		}
	}
}

func TestMaskPhone(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"11位手机号", "13812341234", "138****1234"},
		{"7位保留前3后4", "1234567", "123****4567"},
		{"过短全遮蔽", "123456", "****"},
		{"空串全遮蔽", "", "****"},
	}
	for _, c := range cases {
		if got := MaskPhone(c.in); got != c.want {
			t.Fatalf("%s: want %q got %q", c.name, c.want, got)
		}
	}
}

func TestMaskTail(t *testing.T) {
	cases := []struct {
		name string
		in   string
		keep int
		want string
	}{
		{"长于keep只留尾部", "abcdefg", 4, "defg"},
		{"不超keep原样返回", "abcd", 4, "abcd"},
		{"短于keep原样返回", "ab", 4, "ab"},
		{"keep为0返回原文", "abcd", 0, "abcd"},
		{"keep为负返回原文", "abcd", -1, "abcd"},
		{"空串返回空", "", 4, ""},
	}
	for _, c := range cases {
		if got := MaskTail(c.in, c.keep); got != c.want {
			t.Fatalf("%s: want %q got %q", c.name, c.want, got)
		}
	}
}
