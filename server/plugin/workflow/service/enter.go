// Package service 白泽工单服务聚合
package service

// Service 工单服务单例（生产会话在 dbh() 惰性解析，避免构造期捕获未初始化的 GVA_DB）
var Service = NewWorkflowService(nil, nil)
