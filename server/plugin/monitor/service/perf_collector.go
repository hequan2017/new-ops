// Package service 白泽监控：SSH 性能采集器（M6）
// 单命令组合采样（1 秒内完成）：loadavg / cpu 两次 /proc/stat delta / meminfo / df root；
// 解析为纯函数（可单测），采集复用 asset SSH 通道（指纹校验/凭据保险库）。
package service

import (
	"fmt"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hequan2017/new-ops/server/global"
	assetModel "github.com/hequan2017/new-ops/server/plugin/asset/model"
	assetSvc "github.com/hequan2017/new-ops/server/plugin/asset/service"
	"github.com/hequan2017/new-ops/server/plugin/monitor/model"
)

// perfCmd 性能采样命令（cpu/net 两次采样间隔 1s；NET 标记行分块供解析）
// net/dev 用 tail 跳过两行表头（数据行含冒号，不能用 grep -v ':' 过滤；lo 网卡排除）
const perfCmd = `cat /proc/loadavg; grep 'cpu ' /proc/stat; echo ---NET1---; tail -n +3 /proc/net/dev | grep -v 'lo:'; sleep 1; grep 'cpu ' /proc/stat; echo ---NET2---; tail -n +3 /proc/net/dev | grep -v 'lo:'; grep -E 'MemTotal|MemAvailable' /proc/meminfo; df -kP / | tail -1`

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
// 特征扫描（不依赖行序，NET 块可插在 cpu 两次采样之间）：loadavg 首行 /
// "cpu " 行两次（delta）/ MemTotal+MemAvailable / df 行（含 %）/ NET 标记分块求速率。
func ParsePerfOutput(out string) (*PerfSample, error) {
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) < 6 {
		return nil, fmt.Errorf("采样输出不完整（%d 行）", len(lines))
	}
	s := &PerfSample{}
	cpuRows := make([]string, 0, 2)
	var b1, b2 []string
	netMark := ""
	memTotal, memAvail := 0.0, 0.0

	for idx, ln := range lines {
		trimmed := strings.TrimSpace(ln)
		switch {
		case strings.HasPrefix(trimmed, "cpu "):
			cpuRows = append(cpuRows, trimmed)
		case trimmed == "---NET1---":
			netMark = "1"
			continue
		case trimmed == "---NET2---":
			netMark = "2"
			continue
		case strings.HasPrefix(trimmed, "MemTotal:"):
			if f := strings.Fields(trimmed); len(f) >= 2 {
				memTotal, _ = strconv.ParseFloat(f[1], 64)
			}
			continue
		case strings.HasPrefix(trimmed, "MemAvailable:"):
			if f := strings.Fields(trimmed); len(f) >= 2 {
				memAvail, _ = strconv.ParseFloat(f[1], 64)
			}
			continue
		case strings.Contains(trimmed, "%") && strings.Contains(trimmed, "/"):
			// df 尾行：Filesystem 1024-blocks Used Available Capacity Mounted
			if f := strings.Fields(trimmed); len(f) >= 5 {
				s.DiskPercent, _ = strconv.ParseFloat(strings.TrimSuffix(f[4], "%"), 64)
			}
			continue
		}
		// load1：首个字段为浮点且是第一行
		if idx == 0 {
			if f := strings.Fields(trimmed); len(f) > 0 {
				s.Load1, _ = strconv.ParseFloat(f[0], 64)
			}
			continue
		}
		// 其余行按当前 NET 块归属收集
		switch netMark {
		case "1":
			b1 = append(b1, trimmed)
		case "2":
			b2 = append(b2, trimmed)
		}
	}

	// cpu delta（两次采样）
	if len(cpuRows) >= 2 {
		t1, i1 := cpuStatLine(cpuRows[0])
		t2, i2 := cpuStatLine(cpuRows[1])
		if t2 > t1 {
			s.CPUPercent = (1 - (i2-i1)/(t2-t1)) * 100
			if s.CPUPercent < 0 {
				s.CPUPercent = 0
			}
		}
	}
	// mem
	if memTotal > 0 {
		s.MemPercent = (memTotal - memAvail) / memTotal * 100
	}
	// net 速率（1s 窗口）
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
	// 采样命令限时 15s（原 CombinedOutput 无超时：TCP 通而网络黑洞时 goroutine 永久阻塞，
	// 拖死整轮 wg.Wait，且定时任务每 5 分钟再叠加一轮全量拨号）
	type cmdOut struct {
		b   []byte
		err error
	}
	ch := make(chan cmdOut, 1)
	go func() {
		b, e := session.CombinedOutput(perfCmd)
		ch <- cmdOut{b, e}
	}()
	select {
	case r := <-ch:
		if r.err != nil {
			return nil, fmt.Errorf("采样命令失败: %w", r.err)
		}
		return ParsePerfOutput(string(r.b))
	case <-time.After(15 * time.Second):
		_ = session.Close() // 中断阻塞的执行
		return nil, fmt.Errorf("采样命令超时(15s)")
	}
}

// collectMu 防重入锁：上一轮未结束（存在挂死主机）时跳过本轮，避免 goroutine/连接叠加
var collectMu sync.Mutex

// collectLastDone 上轮完成时刻（unix 秒）；TryLock 只防并发，冷却窗口防背靠背连点
var collectLastDone atomic.Int64

// collectCoolingDown 冷却窗口判定：上轮 10s 内完成过则本轮跳过（last<=0 表示从未采集过不冷却）
func collectCoolingDown(last, now int64) bool {
	return last > 0 && now-last < 10
}

// CollectAll 全量采集：对所有绑定 SSH 凭据且非报废的主机并发采样（上限 5）并落库
func (s *MonitorService) CollectAll() {
	if !collectMu.TryLock() {
		return
	}
	defer collectMu.Unlock()
	// 手动连点冷却：上轮 10s 内刚完成则跳过本轮（与 TryLock 互补——一个防并发，一个防背靠背）
	if collectCoolingDown(collectLastDone.Load(), time.Now().Unix()) {
		return
	}
	defer func() { collectLastDone.Store(time.Now().Unix()) }()
	// 跳过已装在线 agent 的资产：SSH 采样（1s 窗口口径）与 agent 30s 上报双通道并行，
	// 同一指标序列混入两种口径的点（图表毛刺）且白白消耗 SSH 并发与凭据解密。
	// 直查表名避免 monitor→agent import 环（agent 已依赖 monitor 指标模型）；90s 为 agent 在线 TTL
	var agentAssetIDs []uint
	global.GVA_DB.Table("agent_instances").
		Where("last_heartbeat > ?", time.Now().Add(-90*time.Second)).
		Pluck("asset_id", &agentAssetIDs)
	db := global.GVA_DB.Where("credential_id IS NOT NULL AND status != ?", assetModel.AssetStatusRetired)
	if len(agentAssetIDs) > 0 {
		db = db.Where("id NOT IN ?", agentAssetIDs)
	}
	var hosts []*assetModel.AssetHost
	db.Find(&hosts)
	sem := make(chan struct{}, 5)
	var wg sync.WaitGroup
	for _, h := range hosts {
		wg.Add(1)
		go func(h *assetModel.AssetHost) {
			defer wg.Done()
			// cron 级 Recover 不覆盖 job 内再起的 goroutine：远端输出解析 panic 会打死
			// 整个 server——单机失败记日志跳过，不影响本轮其余主机
			defer func() {
				if r := recover(); r != nil {
					global.GVA_LOG.Error(fmt.Sprintf("monitor 采集主机 %s(id=%d) panic: %v\n%s", h.Hostname, h.ID, r, debug.Stack()))
				}
			}()
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

// lastPruneUnix 清理降频水位（每小时最多一次；采集轮 5 分钟一次无需每轮清）
var lastPruneUnix int64

// PruneOld 清理过期指标（保留 30 天）。agent 接入后写入量日增百万行级，原实现两个问题：
// ①ts 在复合索引中非前导列，WHERE ts < ? 走全表扫描；②整表 DELETE 为大事务。
// 现走 ts 单列索引（模型已补）+ 按主键分批删（每批 5000），并按小时降频
func (s *MonitorService) PruneOld() {
	now := time.Now().Unix()
	last := atomic.LoadInt64(&lastPruneUnix)
	if now-last < 3600 {
		return
	}
	if !atomic.CompareAndSwapInt64(&lastPruneUnix, last, now) {
		return // 并发调用里已有赢家在清
	}
	cutoff := time.Now().Add(-30 * 24 * time.Hour)
	for {
		res := global.GVA_DB.Where("id IN (SELECT id FROM monitor_metrics WHERE ts < ? LIMIT 5000)", cutoff).
			Delete(&model.MonitorMetric{})
		if res.Error != nil {
			global.GVA_LOG.Error("monitor 指标清理分批删除失败: " + res.Error.Error())
			return
		}
		if res.RowsAffected == 0 {
			return
		}
	}
}
