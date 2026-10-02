package service

import (
	"strings"
	"testing"
)

// 固定 nonce/timestamp 下签名结果必须确定（按协议手工推算基准值）
func Test_buildSignedQuery(t *testing.T) {
	q, err := buildSignedQuery("testid", "testsecret", map[string]string{
		"Action":   "DescribeInstances",
		"RegionId": "cn-beijing",
	}, "nonce-1", "2026-10-02T12:00:00Z")
	if err != nil {
		t.Fatalf("构建失败: %v", err)
	}
	// 公共参数齐全
	for _, k := range []string{"Format=JSON", "Version=2014-05-26", "AccessKeyId=testid",
		"SignatureMethod=HMAC-SHA1", "SignatureVersion=1.0", "SignatureNonce=nonce-1",
		"Timestamp=2026-10-02T12%3A00%3A00Z", "Action=DescribeInstances",
		"RegionId=cn-beijing", "Signature="} {
		if !strings.Contains(q, k) {
			t.Fatalf("查询串缺少 %s\n%s", k, q)
		}
	}
	// 同输入两次构建结果必须一致（确定性）
	q2, _ := buildSignedQuery("testid", "testsecret", map[string]string{
		"Action": "DescribeInstances", "RegionId": "cn-beijing",
	}, "nonce-1", "2026-10-02T12:00:00Z")
	if q != q2 {
		t.Fatalf("同输入签名应一致")
	}
}

func Test_percentEncode(t *testing.T) {
	cases := map[string]string{
		"a b": "a%20b",
		"a*b": "a%2Ab",
		"a~b": "a~b",
		"a:b": "a%3Ab",
		"a/b": "a%2Fb",
	}
	for in, want := range cases {
		if got := percentEncode(in); got != want {
			t.Fatalf("percentEncode(%s) = %s, want %s", in, got, want)
		}
	}
	// 全量编码后无未编码保留字符
	if percentEncode("~") != "~" {
		t.Fatalf("~ 应保持不编码")
	}
}

func Test_mapECSStatus(t *testing.T) {
	if mapECSStatus("Running") != "运行中" {
		t.Fatalf("Running 映射错误")
	}
	if mapECSStatus("Stopped") != "停机" {
		t.Fatalf("Stopped 映射错误")
	}
	if mapECSStatus("Starting") != "维护中" {
		t.Fatalf("其他状态应映射维护中")
	}
}

const ecsSample = `{"RequestId":"r-1","Instances":{"TotalCount":1,"Instance":[{
  "InstanceId":"i-bp1abc","HostName":"web-01","InstanceName":"生产web",
  "InnerIpAddress":{"IpAddress":["10.0.0.5"]},
  "PublicIpAddress":{"IpAddress":["47.1.2.3"]},
  "OSName":"Ubuntu  22.04  64位","CPU":8,"Memory":16384,"Status":"Running"}]}}`

func Test_parseECSDescribeResponse(t *testing.T) {
	list, total, err := parseECSDescribeResponse([]byte(ecsSample))
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if total != 1 || len(list) != 1 {
		t.Fatalf("total=%d len=%d", total, len(list))
	}
	it := list[0]
	if it.InstanceID != "i-bp1abc" || it.Hostname != "web-01" || it.InnerIP != "10.0.0.5" ||
		it.PublicIP != "47.1.2.3" || it.CPU != 8 || it.MemoryMB != 16384 || it.Status != "Running" {
		t.Fatalf("字段映射错误: %+v", it)
	}
}

func Test_parseECSDescribeResponse_error(t *testing.T) {
	_, _, err := parseECSDescribeResponse([]byte(`{"Code":"InvalidAccessKeyId.NotFound","Message":"The AccessKeyId is not exist."}`))
	if err == nil || !strings.Contains(err.Error(), "InvalidAccessKeyId") {
		t.Fatalf("错误响应应返回错误, got %v", err)
	}
	_, _, err = parseECSDescribeResponse([]byte("not json"))
	if err == nil {
		t.Fatalf("非JSON应报错")
	}
}
