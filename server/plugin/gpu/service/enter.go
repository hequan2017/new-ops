// Package service 白泽 GPU 算力服务聚合
package service

// Service GPU 服务单例（生产会话在 dbh() 惰性解析，避免构造期捕获未初始化的 GVA_DB）
var Service = NewGpuService(nil)
