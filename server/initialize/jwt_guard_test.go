package initialize

import "testing"

func TestJWTKeyGuard(t *testing.T) {
	cases := []struct {
		name   string
		env    string
		key    string
		refuse bool
	}{
		{"local 环境弱键放行", "local", "qmPlus", false},
		{"local 环境空键放行", "local", "", false},
		{"public 强键放行", "public", "6f26a9d2a1caa84fbd05bb12e9d9c7870f1c5cffa3347848e667a0dc7c7290ff", false},
		{"public 空键拒绝", "public", "", true},
		{"public 空白键拒绝", "public", "   ", true},
		{"public qmPlus 拒绝", "public", "qmPlus", true},
	}
	for _, c := range cases {
		refuse, reason := JWTKeyGuard(c.env, c.key)
		if refuse != c.refuse {
			t.Fatalf("%s: refuse=%v want %v（reason=%q）", c.name, refuse, c.refuse, reason)
		}
		if c.refuse && reason == "" {
			t.Fatalf("%s: 拒绝时应给出原因", c.name)
		}
	}
}
