package service

import (
	"strings"
	"testing"
)

// 构造一段真实形态的采样输出（数字可校验）
const sampleOutput = "0.52 0.58 0.60 1/512 12345\n" +
	"cpu  100 0 50 850 0 0 0 0 0 0\n" +
	"cpu  110 0 60 830 0 0 0 0 0 0\n" +
	"MemTotal:       16000000 kB\n" +
	"MemAvailable:    4000000 kB\n" +
	"/dev/sda1 500000000 300000000 200000000 61% /"

func TestParsePerfOutput(t *testing.T) {
	s, err := ParsePerfOutput(sampleOutput)
	if err != nil {
		t.Fatalf("解析不应报错: %v", err)
	}
	// load1
	if s.Load1 != 0.52 {
		t.Fatalf("load1 应为 0.52: %v", s.Load1)
	}
	// cpu：t1=1000 idle1=850；t2=1000+100=1100? 逐项 110+0+60+830=1000？重算：
	// 第二行 110 0 60 830 0... total2=1100? fields=110,0,60,830,0,0,0,0,0,0 → total=1000？
	// 110+0+60+830 = 1000，idle2=830
	// total1=100+0+50+850=1000，idle1=850
	// delta_total=0 → CPU 保持 0（除零保护）
	if s.CPUPercent != 0 {
		t.Fatalf("delta_total=0 时 CPU 应为 0: %v", s.CPUPercent)
	}
	// mem：(16000000-4000000)/16000000 = 75%
	if s.MemPercent != 75 {
		t.Fatalf("mem 应为 75: %v", s.MemPercent)
	}
	// disk 61%
	if s.DiskPercent != 61 {
		t.Fatalf("disk 应为 61: %v", s.DiskPercent)
	}
}

func TestParsePerfOutput_CPUDelta(t *testing.T) {
	// 明确非零 delta：idle 不变、total 增加 100 → busy 100%？total 1000→1100 idle 850→850：busy delta=100/100=100%
	out := "0.10 0.20 0.30 1/8 999\n" +
		"cpu  100 0 50 850 0 0 0 0 0 0\n" +
		"cpu  200 0 50 850 0 0 0 0 0 0\n" +
		"MemTotal:       8000000 kB\n" +
		"MemAvailable:   8000000 kB\n" +
		"/dev/sda1 100 50 50 50% /"
	s, err := ParsePerfOutput(out)
	if err != nil {
		t.Fatalf("解析不应报错: %v", err)
	}
	if s.CPUPercent != 100 {
		t.Fatalf("CPU 应为 100: %v", s.CPUPercent)
	}
	if s.MemPercent != 0 {
		t.Fatalf("mem 应为 0: %v", s.MemPercent)
	}
}

func TestParsePerfOutput_TooFewLines(t *testing.T) {
	if _, err := ParsePerfOutput(strings.Repeat("x\n", 3)); err == nil {
		t.Fatal("行数不足应报错")
	}
}
