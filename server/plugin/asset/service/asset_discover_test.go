package service

import "testing"

func TestExpandCIDR(t *testing.T) {
	// /30：网络地址+广播地址跳过 → 剩 2 个
	ips, err := ExpandCIDR("192.168.1.4/30", 0)
	if err != nil {
		t.Fatalf("/30 展开不应报错: %v", err)
	}
	if len(ips) != 2 || ips[0] != "192.168.1.5" || ips[1] != "192.168.1.6" {
		t.Fatalf("/30 展开结果错误: %v", ips)
	}
	// /32：单主机（前后缀 >=31 不跳过）
	ips, err = ExpandCIDR("10.0.0.7/32", 0)
	if err != nil || len(ips) != 1 || ips[0] != "10.0.0.7" {
		t.Fatalf("/32 展开错误: %v err=%v", ips, err)
	}
	// 非法格式
	if _, err := ExpandCIDR("not-a-cidr", 0); err == nil {
		t.Fatal("非法 CIDR 应报错")
	}
	// IPv6 拒绝
	if _, err := ExpandCIDR("2001:db8::/64", 0); err == nil {
		t.Fatal("IPv6 应报错")
	}
	// 超上限
	if _, err := ExpandCIDR("10.0.0.0/8", 4096); err == nil {
		t.Fatal("超上限应报错")
	}
	// 上限内最大段 /20 = 4096
	ips, err = ExpandCIDR("172.16.0.0/20", 0)
	if err != nil || len(ips) != 4094 {
		t.Fatalf("/20 应展开 4094 个（去网络/广播）: %d err=%v", len(ips), err)
	}
}

func TestIsSSHBanner(t *testing.T) {
	cases := map[string]bool{
		"SSH-2.0-OpenSSH_8.9p1 Ubuntu-3ubuntu0.4": true,
		"SSH-1.99-OpenSSH_5.3":                    true,
		"HTTP/1.1 400 Bad Request":                false,
		"":                                        false,
	}
	for in, want := range cases {
		if got := isSSHBanner(in); got != want {
			t.Fatalf("isSSHBanner(%q)=%v want %v", in, got, want)
		}
	}
}

func TestSortDiscoveredByIP(t *testing.T) {
	in := []DiscoveredHost{{IP: "10.0.0.10"}, {IP: "10.0.0.2"}, {IP: "192.168.1.5"}, {IP: "9.9.9.9"}}
	sortDiscoveredByIP(in)
	want := []string{"9.9.9.9", "10.0.0.2", "10.0.0.10", "192.168.1.5"}
	for i, w := range want {
		if in[i].IP != w {
			t.Fatalf("IPv4 数值序 want[%d]=%s got %s", i, w, in[i].IP)
		}
	}

	// 空表与非 IPv4 输入不 panic（回退字符串序）
	empty := []DiscoveredHost{}
	sortDiscoveredByIP(empty)
	bad := []DiscoveredHost{{IP: "b"}, {IP: "a"}}
	sortDiscoveredByIP(bad)
	if bad[0].IP != "a" {
		t.Fatalf("非法 IP 应回退字符串序, got %s", bad[0].IP)
	}
}
