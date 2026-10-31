package service

import "testing"

// Test_shortenID 容器 ID 截断（12 位短 ID 口径，恰 12 位不截）
func Test_shortenID(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"64 位完整 ID 截 12", "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef", "0123456789ab"},
		{"恰 12 位原样", "0123456789ab", "0123456789ab"},
		{"短于 12 原样", "abc", "abc"},
		{"空串", "", ""},
	}
	for _, c := range cases {
		if got := shortenID(c.in); got != c.want {
			t.Fatalf("%s: want %q got %q", c.name, c.want, got)
		}
	}
}
