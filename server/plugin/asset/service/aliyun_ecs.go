// Package service 阿里云 ECS 实例同步（M1 云同步通道）
// 依赖策略：不引入官方 SDK（传递依赖过重），自实现 OpenAPI RPC V1 签名（纯标准库），
// 仅使用 DescribeInstances 单接口；后续接入更多云产品时再评估 SDK。
package service

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/hequan2017/new-ops/server/plugin/asset/model"
)

const (
	ecsEndpoint    = "https://ecs.aliyuncs.com"
	ecsAPIVersion  = "2014-05-26"
	ecsHTTPTimeout = 15 * time.Second
)

// percentEncode 阿里云 RPC 签名的 RFC3986 编码
func percentEncode(s string) string {
	e := url.QueryEscape(s)
	e = strings.ReplaceAll(e, "+", "%20")
	e = strings.ReplaceAll(e, "*", "%2A")
	e = strings.ReplaceAll(e, "%7E", "~")
	return e
}

// buildSignedQuery 组装带签名的公共参数查询串（ nonce/timestamp 由调用方注入以便单测）
func buildSignedQuery(accessKeyID, accessKeySecret string, extra map[string]string, nonce, timestamp string) (string, error) {
	if accessKeyID == "" || accessKeySecret == "" {
		return "", errors.New("AccessKeyID/Secret 不能为空")
	}
	params := map[string]string{
		"Format":           "JSON",
		"Version":          ecsAPIVersion,
		"AccessKeyId":      accessKeyID,
		"SignatureMethod":  "HMAC-SHA1",
		"SignatureVersion": "1.0",
		"SignatureNonce":   nonce,
		"Timestamp":        timestamp,
	}
	for k, v := range extra {
		params[k] = v
	}
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	pairs := make([]string, 0, len(keys))
	for _, k := range keys {
		pairs = append(pairs, percentEncode(k)+"="+percentEncode(params[k]))
	}
	canonicalQuery := strings.Join(pairs, "&")
	stringToSign := "GET&" + percentEncode("/") + "&" + percentEncode(canonicalQuery)
	// 阿里云 RPC V1 签名协议固定要求 HMAC-SHA1（协议约束，非自选算法）
	mac := hmac.New(sha1.New, []byte(accessKeySecret+"&"))
	mac.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return canonicalQuery + "&Signature=" + percentEncode(signature), nil
}

// ecsInstance 阿里云实例标准化结构
type ecsInstance struct {
	InstanceID string
	Hostname   string
	InnerIP    string
	PublicIP   string
	OSName     string
	CPU        int
	MemoryMB   int
	Status     string
}

// mapECSStatus 阿里云实例状态 → 白泽资产状态
func mapECSStatus(aliStatus string) model.AssetStatus {
	switch strings.ToLower(aliStatus) {
	case "running":
		return model.AssetStatusRunning
	case "stopped":
		return model.AssetStatusStopped
	default:
		return model.AssetStatusMaintain // 启动中/停止中/其他 → 维护中
	}
}

// parseECSDescribeResponse 解析 DescribeInstances 响应
func parseECSDescribeResponse(body []byte) (instances []ecsInstance, total int, err error) {
	var resp struct {
		Instances struct {
			Instance []struct {
				InstanceId   string `json:"InstanceId"`
				HostName     string `json:"HostName"`
				InstanceName string `json:"InstanceName"`
				InnerIP      struct {
					IpAddress []string `json:"IpAddress"`
				} `json:"InnerIpAddress"`
				PublicIP struct {
					IpAddress []string `json:"IpAddress"`
				} `json:"PublicIpAddress"`
				OSName    string `json:"OSName"`
				CPU       int    `json:"CPU"`
				Memory    int    `json:"Memory"`
				Status    string `json:"Status"`
			} `json:"Instance"`
			TotalCount int `json:"TotalCount"`
		} `json:"Instances"`
		Code      string `json:"Code"`
		Message   string `json:"Message"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, 0, fmt.Errorf("响应解析失败: %w", err)
	}
	if resp.Code != "" {
		return nil, 0, fmt.Errorf("阿里云错误 %s: %s", resp.Code, resp.Message)
	}
	for _, it := range resp.Instances.Instance {
		inner := ""
		if len(it.InnerIP.IpAddress) > 0 {
			inner = it.InnerIP.IpAddress[0]
		}
		pub := ""
		if len(it.PublicIP.IpAddress) > 0 {
			pub = it.PublicIP.IpAddress[0]
		}
		instances = append(instances, ecsInstance{
			InstanceID: it.InstanceId,
			Hostname:   it.HostName,
			InnerIP:    inner,
			PublicIP:   pub,
			OSName:     it.OSName,
			CPU:        it.CPU,
			MemoryMB:   it.Memory,
			Status:     it.Status,
		})
	}
	return instances, resp.Instances.TotalCount, nil
}

// fetchECSInstances 调用 DescribeInstances 拉取全量实例（自动分页）
func fetchECSInstances(accessKeyID, accessKeySecret, region string, pageSize int) ([]ecsInstance, error) {
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 100
	}
	var all []ecsInstance
	for page := 1; ; page++ {
		nonce := fmt.Sprintf("%d-%d", time.Now().UnixNano(), page)
		timestamp := time.Now().UTC().Format("2006-01-02T15:04:05Z")
		query, err := buildSignedQuery(accessKeyID, accessKeySecret, map[string]string{
			"Action":     "DescribeInstances",
			"RegionId":   region,
			"PageSize":   fmt.Sprint(pageSize),
			"PageNumber": fmt.Sprint(page),
		}, nonce, timestamp)
		if err != nil {
			return nil, err
		}
		client := &http.Client{Timeout: ecsHTTPTimeout}
		respHTTP, err := client.Get(ecsEndpoint + "/?" + query)
		if err != nil {
			return nil, fmt.Errorf("请求阿里云失败: %w", err)
		}
		body, _ := io.ReadAll(respHTTP.Body)
		respHTTP.Body.Close()
		if respHTTP.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("阿里云 HTTP %d: %s", respHTTP.StatusCode, truncateStr(string(body), 200))
		}
		instances, total, err := parseECSDescribeResponse(body)
		if err != nil {
			return nil, err
		}
		all = append(all, instances...)
		if len(all) >= total || len(instances) == 0 {
			return all, nil
		}
	}
}

func truncateStr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
