// Package service oplib：执行入参
package service

// RunIn 算子执行请求
type RunIn struct {
	ID   uint   `json:"id" binding:"required"`
	Args string `json:"args"`
}
