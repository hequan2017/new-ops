// Package service 白泽监控：SSH 性能采集器（M6）
// 单命令组合采样（1 秒内完成）：loadavg / cpu 两次 /proc/stat delta / meminfo / df root；
// 解析为纯函数（可单测），采集复用 asset SSH 通道（指纹校验/凭据保险库）。
package service

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/hequan2017/new-ops/server/global"
	assetModel "github.com/hequan2017/new-ops/server/plugin/asset/model"
	assetSvc "github.com/hequan2017/new-ops/server/plugin/asset/service"
	"github.com/hequan2017/new-ops/server/plugin/monitor/model"
)

// perfCmd 性能采样命令（cpu/net 两次采样间隔 1s；NET 标记行分块供解析）
const perfCmd = `cat /proc/loadavg; grep 'cpu ' /proc/stat; echo ---NET1---; grep -v ':' /proc/net/dev | grep -v lo; sleep 1; grep 'cpu ' /proc/stat; echo ---NET2---; grep -v ':' /proc/net/dev | grep -v lo; grep -E 'MemTotal|MemAvailable' /proc/meminfo; df -kP / | tail -1`

// PerfSample 一次采样的六指标（net 为 1s 窗口速率 KB/s，全部非 lo 网卡之和）
type PerfSample struct {
	CPUPercent  float64
	MemPercent  float64
	DiskPercent float64
	Load1       float64
	NetRxKBs    float64
	NetTxKBs    float64
}

// cpuStatLine 解析 /proc/stat 的 cpu 聚合行 → (user,nice,system,idle,iowait,irq,softirq,steal 总和与 idle 和)
func cpuStatLine(line string) (total, idle float64) {
	fields := strings.Fields(line)
	// cpu user nice system idle iowait irq softirq steal ...
	if len(fields) < 5 {
		return 0, 0
	}
	vals := make([]float64, 0, len(fields)-1)
	for _, f := range fields[1:] {
		v, err := strconv.ParseFloat(f, 64)
		if err != nil {
			return 0, 0
		}
		vals = append(vals, v)
	}
	idle = vals[3]
	if len(vals) > 4 {
		idle += vals[4] // iowait 计入空闲
	}
	for _, v := range vals {
		total += v
	}
	return total, idle
}

// netDevLine 解析 /proc/net/dev 数据行 → (iface, rxBytes, txBytes, ok)
func netDevLine(line string) (string, float64, float64, bool) {
	// 形如 "  eth0: 12345 100 ... 67890 200 ..."
	i := strings.Index(line, ":")
	if i < 0 {
		return "", 0, 0, false
	}
	iface := strings.TrimSpace(line[:i])
	if iface == "" || iface == "lo" {
		return "", 0, 0, false
	}
	f := strings.Fields(line[i+1:])
	if len(f) < 10 {
		return "", 0, 0, false
	}
	rx, err1 := strconv.ParseFloat(f[0], 64)
	tx, err2 := strconv.ParseFloat(f[8], 64)
	if err1 != nil || err2 != nil {
		return "", 0, 0, false
	}
	return iface, rx, tx, true
}

// netBlockRate 计算两次 net/dev 采样块的速率（KB/s，全网卡求和；纯函数，可单测）
func netBlockRate(block1, block2 []string, window time.Duration) (rxKBs, txKBs float64) {
	if window <= 0 || len(block1) == 0 || len(block2) == 0 {
		return 0, 0
	}
	sum := func(block []string) (rx, tx float64) {
		for _, ln := range block {
			if _, r, t, ok := netDevLine(ln); ok {
				rx += r
				tx += t
			}
		}
		return rx, tx
	}
	rx1, tx1 := sum(block1)
	rx2, tx2 := sum(block2)
	secs := window.Seconds()
	if secs <= 0 || rx2 < rx1 || tx2 < tx1 {
		return 0, 0 // 计数器回绕/重启异常时按 0 处理
	}
	return (rx2 - rx1) / 1024 / secs, (tx2 - tx1) / 1024 / secs
}

// ParsePerfOutput 解析采样输出（纯函数，可单测）
// 结构：loadavg / cpu 行×2 / MemTotal / MemAvailable / df 尾行，
// 其后可选 NET 分块（---NET1--- 与 ---NET2--- 标记，各含若干 net/dev 行）——向后兼容旧格式。
func ParsePerfOutput(out string) (*PerfSample, error) {
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) < 6 {
		return nil, fmt.Errorf("采样输出不完整（%d 行）", len(lines))
	}
	s := &PerfSample{}
	// load1
	if f := strings.Fields(lines[0]); len(f) > 0 {
		s.Load1, _ = strconv.ParseFloat(f[0], 64)
	}
	// cpu delta（两次采样）
	t1, i1 := cpuStatLine(lines[1])
	t2, i2 := cpuStatLine(lines[2])
	if t2 > t1 {
		s.CPUPercent = (1 - (i2-i1)/(t2-t1)) * 100
		if s.CPUPercent < 0 {
			s.CPUPercent = 0
		}
	}
	// mem
	var totalKB, availKB float64
	for _, ln := range lines[3:5] {
		f := strings.Fields(ln)
		if len(f) < 2 {
			continue
		}
		v, _ := strconv.ParseFloat(f[1], 64)
		switch f[0] {
		case "MemTotal:":
			totalKB = v
		case "MemAvailable:":
			availKB = v
		}
	}
	if totalKB > 0 {
		s.MemPercent = (totalKB - availKB) / totalKB * 100
	}
	// df 尾行：Filesystem 1024-blocks Used Available Capacity Mounted
	if f := strings.Fields(lines[5]); len(f) >= 5 {
		s.DiskPercent, _ = strconv.ParseFloat(strings.TrimSuffix(f[4], "%"), 64)
	}
	// NET 分块（可选）：按标记切两次采样块求速率
	var b1, b2 []string
	cur := ""
	for _, ln := range lines[6:] {
		switch strings.TrimSpace(ln) {
		case "---NET1---":
			cur = "1"
			continue
		case "---NET2---":
			cur = "2"
			continue
		}
		switch cur {
		case "1":
			b1 = append(b1, ln)
		case "2":
			b2 = append(b2, ln)
		}
	}
	if len(b1) > 0 && len(b2) > 0 {
		s.NetRxKBs, s.NetTxKBs = netBlockRate(b1, b2, time.Second)
	}
	return s, nil
}

// CollectPerf 对单主机执行性能采样（SSH 通道，超时 15s）
func (s *MonitorService) CollectPerf(host *assetModel.AssetHost) (*PerfSample, error) {
	if host.IP == "" || host.CredentialID == nil {
		return nil, fmt.Errorf("主机 %s 缺少 IP 或凭据", host.Hostname)
	}
	secret, credType, username, err := assetSvc.Service.CredCredentialService.GetPlaintext(*host.CredentialID)
	if err != nil {
		return nil, err
	}
	if credType != assetModel.CredTypeSSHPassword && credType != assetModel.CredTypeSSHKey {
		return nil, fmt.Errorf("凭据 id=%d 非 SSH 类型", *host.CredentialID)
	}
	client, _, err := assetSvc.DialSSH(host.IP, assetSvc.SSHAuth{
		Username: username, Password: secret, PrivateKey: secret,
		ExpectedFingerprint: host.SSHFP,
	})
	if err != nil {
		return nil, err
	}
	defer client.Close()
	session, err := client.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()
	out, err := session.CombinedOutput(perfCmd)
	if err != nil {
		return nil, fmt.Errorf("采样命令失败: %w", err)
	}
	return ParsePerfOutput(string(out))
}

// CollectAll 全量采集：对所有绑定 SSH 凭据且非报废的主机并发采样（上限 5）并落库
func (s *MonitorService) CollectAll() {
	var hosts []*assetModel.AssetHost
	global.GVA_DB.Where("credential_id IS NOT NULL AND status != ?", assetModel.AssetStatusRetired).
		Find(&hosts)
	sem := make(chan struct{}, 5)
	var wg sync.WaitGroup
	for _, h := range hosts {
		wg.Add(1)
		go func(h *assetModel.AssetHost) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			sample, err := s.CollectPerf(h)
			if err != nil {
				return // 单机失败不影响整体（下一轮重试）
			}
			now := time.Now()
			rows := []model.MonitorMetric{
				{AssetID: h.ID, Name: model.MetricCPUPercent, Value: round2(sample.CPUPercent), TS: now},
				{AssetID: h.ID, Name: model.MetricMemPercent, Value: round2(sample.MemPercent), TS: now},
				{AssetID: h.ID, Name: model.MetricDiskPercent, Value: round2(sample.DiskPercent), TS: now},
				{AssetID: h.ID, Name: model.MetricLoad1, Value: round2(sample.Load1), TS: now},
				{AssetID: h.ID, Name: model.MetricNetRxKBs, Value: round2(sample.NetRxKBs), TS: now},
				{AssetID: h.ID, Name: model.MetricNetTxKBs, Value: round2(sample.NetTxKBs), TS: now},
			}
			global.GVA_DB.Create(&rows)
		}(h)
	}
	wg.Wait()
	s.PruneOld()
	s.EvalAllRules() // 采集完成后统一评估告警规则（指标+端口探活）
}

// round2 保留两位小数
func round2(v float64) float64 {
	return float64(int64(v*100+0.5)) / 100
}

// GetMetrics 查询主机指标序列（时间升序，上限 2000 点）
func (s *MonitorService) GetMetrics(assetID uint, names []string, since time.Time) ([]model.MonitorMetric, error) {
	db := global.GVA_DB.Model(&model.MonitorMetric{}).
		Where("asset_id = ? AND ts >= ?", assetID, since)
	if len(names) > 0 {
		db = db.Where("name IN ?", names)
	}
	var list []model.MonitorMetric
	err := db.Order("ts ASC").Limit(2000).Find(&list).Error
	return list, err
}

// PruneOld 清理过期指标（保留 30 天）
func (s *MonitorService) PruneOld() {
	global.GVA_DB.Where("ts < ?", time.Now().Add(-30*24*time.Hour)).
		Delete(&model.MonitorMetric{})
}
