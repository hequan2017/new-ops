package utils

import "testing"

func TestFirstUpperLower(t *testing.T) {
	cases := []struct {
		in        string
		wantUpper string
		wantLower string
	}{
		{"hostName", "HostName", "hostName"},
		{"", "", ""},
		{"a", "A", "a"},
	}
	for _, c := range cases {
		if got := FirstUpper(c.in); got != c.wantUpper {
			t.Fatalf("FirstUpper(%q): want %q got %q", c.in, c.wantUpper, got)
		}
		if got := FirstLower(c.wantUpper); got != c.wantLower {
			t.Fatalf("FirstLower(%q): want %q got %q", c.wantUpper, c.wantLower, got)
		}
	}
}

func TestHumpToUnderscore(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"HostName", "host_name"},
		{"hostName", "host_name"},
		{"IP", "i_p"},
		{"", ""},
	}
	for _, c := range cases {
		if got := HumpToUnderscore(c.in); got != c.want {
			t.Fatalf("HumpToUnderscore(%q): want %q got %q", c.in, c.want, got)
		}
	}
}

func TestPointer(t *testing.T) {
	v := 42
	p := Pointer(v)
	if p == nil || *p != 42 {
		t.Fatalf("Pointer 应返回入值指针, got %v", p)
	}
}

func TestRandomString(t *testing.T) {
	s := RandomString(16)
	if len(s) != 16 {
		t.Fatalf("RandomString(16) 长度应为 16, got %d", len(s))
	}
	if RandomString(0) != "" {
		t.Fatal("RandomString(0) 应为空串")
	}
}
