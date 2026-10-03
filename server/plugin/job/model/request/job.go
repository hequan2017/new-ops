// Package request job 插件请求结构
package request

// BatchExecReq 发起批量命令执行
type BatchExecReq struct {
	HostIDs      []uint `json:"hostIds" binding:"required"` // 目标主机（受数据权限约束）
	Command      string `json:"command" binding:"required"` // 命令
	Concurrency  int    `json:"concurrency"`                // 并发上限 1-50，默认 10
	TimeoutSec   int    `json:"timeoutSec"`                 // 单机超时 5-600 秒，默认 30
	CredentialID *uint  `json:"credentialId"`               // 统一凭据（空则用各主机绑定凭据）
}
