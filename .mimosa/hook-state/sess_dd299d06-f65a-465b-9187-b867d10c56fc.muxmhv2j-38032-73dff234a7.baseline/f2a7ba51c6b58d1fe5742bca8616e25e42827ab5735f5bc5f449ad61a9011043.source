// Package service 白泽资产中心：CIDR 网段发现（go-webssh 路线）
// 流程：展开 CIDR（上限保护）→ 并发 TCP 探测 → 抓取 SSH banner →
// 前端勾选后按 IP 导入为资产（已存在跳过）。
package service

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/asset/model"
)

// discover 插件内错误码（沿用 asset 1000-1099 段）
const ErrCodeCIDRInvalid = 1006

// maxDiscoverHosts 单次发现展开的地址数上限（防 /0 类超大段拖垮服务）
const maxDiscoverHosts = 4096

// DiscoveredHost 探测结果（仅收录端口开放的主机）
type DiscoveredHost struct {
	IP     string `json:"ip"`
	Banner string `json:"banner"` // SSH 服务横幅（如 SSH-2.0-OpenSSH_8.9p1 Ubuntu-...）
}

// ExpandCIDR 展开 CIDR 为全部 IPv4 地址（纯逻辑，可单测；超过 max 上限报错）
func ExpandCIDR(cidr string, max int) ([]string, error) {
	if max <= 0 {
		max = maxDiscoverHosts
	}
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, newServiceErr(ErrCodeCIDRInvalid, fmt.Sprintf("CIDR 格式不合法: %s", cidr))
	}
	if ip.To4() == nil {
		return nil, newServiceErr(ErrCodeCIDRInvalid, "仅支持 IPv4 网段")
	}
	// 从网络地址起遍历，跳过网络地址与广播地址（前缀 <31 时）
	base := ip.Mask(ipnet.Mask).To4()
	ones, bits := ipnet.Mask.Size()
	total := 1 << (bits - ones)
	if total > max {
		return nil, newServiceErr(ErrCodeCIDRInvalid,
			fmt.Sprintf("网段包含 %d 个地址，超过单次上限 %d（请缩小范围）", total, max))
	}
	out := make([]string, 0, total)
	for i := 0; i < total; i++ {
		cur := net.IPv4(base[0], base[1], base[2], base[3]).To4()
		inc := uint32(cur[0])<<24 | uint32(cur[1])<<16 | uint32(cur[2])<<8 | uint32(cur[3])
		inc += uint32(i)
		addr := net.IPv4(byte(inc>>24), byte(inc>>16), byte(inc>>8), byte(inc)).To4().String()
		if ones < 31 && (i == 0 || i == total-1) {
			continue // 跳过网络地址与广播地址
		}
		out = append(out, addr)
	}
	return out, nil
}

// isSSHBanner 判定抓到的横幅是否 SSH 服务（纯逻辑）
func isSSHBanner(b string) bool {
	return strings.HasPrefix(b, "SSH-")
}

// probeSSH TCP 拨号并抓取 SSH 横幅（超时整体控制；非 SSH 服务返回 ok=false）
func probeSSH(ctx context.Context, ip string, port int, timeout time.Duration) (DiscoveredHost, bool) {
	addr := net.JoinHostPort(ip, fmt.Sprintf("%d", port))
	d := net.Dialer{Timeout: timeout}
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return DiscoveredHost{}, false
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(timeout))
	buf := make([]byte, 128)
	n, err := conn.Read(buf)
	if err != nil && n == 0 {
		return DiscoveredHost{}, false
	}
	banner := strings.TrimSpace(string(buf[:n]))
	if !isSSHBanner(banner) {
		return DiscoveredHost{}, false
	}
	return DiscoveredHost{IP: ip, Banner: banner}, true
}

// DiscoverHosts 并发探测网段内 SSH 服务（concurrency<=0 取 50，上限 200）
func (s *AssetHostService) DiscoverHosts(ctx context.Context, cidr string, port, concurrency, timeoutMs int) ([]DiscoveredHost, error) {
	if port <= 0 || port > 65535 {
		port = 22
	}
	if concurrency <= 0 {
		concurrency = 50
	}
	if concurrency > 200 {
		concurrency = 200
	}
	if timeoutMs <= 0 {
		timeoutMs = 1500
	}
	if timeoutMs > 10000 {
		timeoutMs = 10000
	}
	addrs, err := ExpandCIDR(cidr, maxDiscoverHosts)
	if err != nil {
		return nil, err
	}
	timeout := time.Duration(timeoutMs) * time.Millisecond
	sem := make(chan struct{}, concurrency)
	var mu sync.Mutex
	var out []DiscoveredHost
	var wg sync.WaitGroup
	for _, ip := range addrs {
		wg.Add(1)
		go func(a string) {
			defer wg.Done()
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				return
			}
			if h, ok := probeSSH(ctx, a, port, timeout); ok {
				mu.Lock()
				out = append(out, h)
				mu.Unlock()
			}
		}(ip)
	}
	wg.Wait()
	return out, nil
}

// ImportDiscoveredHosts 将发现的主机导入为资产（已存在跳过；主机名先取 IP，后续可人工改）
// 返回新建数与跳过数
func (s *AssetHostService) ImportDiscoveredHosts(hosts []DiscoveredHost, operator string) (created, skipped int, err error) {
	for _, h := range hosts {
		if net.ParseIP(h.IP) == nil {
			skipped++
			continue
		}
		var exist model.AssetHost
		dbErr := global.GVA_DB.Where("ip = ?", h.IP).First(&exist).Error
		if dbErr == nil {
			skipped++ // 已存在（含未软删）
			continue
		}
		note := "网段发现导入"
		if h.Banner != "" {
			note += "，SSH横幅: " + h.Banner
		}
		host := &model.AssetHost{
			Hostname: h.IP,
			IP:       h.IP,
			Status:   model.AssetStatusRunning,
			Notes:    note,
		}
		if err := s.CreateAssetHost(host, operator); err != nil {
			// 并发重复等场景按跳过计，不中断整体导入
			if se, ok := err.(*ServiceError); ok && se.Code == ErrCodeIPDuplicate {
				skipped++
				continue
			}
			return created, skipped, err
		}
		created++
	}
	return created, skipped, nil
}
