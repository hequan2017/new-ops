package service

import "testing"

func TestNormalizeExecBackend(t *testing.T) {
	cases := []struct {
		in   string
		want string
		fail bool
	}{
		{"", "ssh", false},
		{"ssh", "ssh", false},
		{"agent", "agent", false},
		{"exec", "", true},
		{"SSH", "", true}, // 大小写敏感，前端约定小写
	}
	for _, c := range cases {
		got, err := normalizeExecBackend(c.in)
		if c.fail {
			if err == nil {
				t.Fatalf("normalizeExecBackend(%q) 应报错", c.in)
			}
			continue
		}
		if err != nil || got != c.want {
			t.Fatalf("normalizeExecBackend(%q)=%q,%v want %q", c.in, got, err, c.want)
		}
	}
}
