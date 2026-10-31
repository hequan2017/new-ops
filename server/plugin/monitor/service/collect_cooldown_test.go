package service

import "testing"

func TestCollectCoolingDown(t *testing.T) {
	cases := []struct {
		name string
		last int64
		now  int64
		want bool
	}{
		{"从未采集不冷却", 0, 1000, false},
		{"10s 内连点冷却", 1000, 1009, true},
		{"恰好 10s 不冷却", 1000, 1010, false},
		{"超过 10s 不冷却", 1000, 1011, false},
	}
	for _, c := range cases {
		if got := collectCoolingDown(c.last, c.now); got != c.want {
			t.Fatalf("%s: want %v got %v", c.name, c.want, got)
		}
	}
}
