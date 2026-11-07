// Package service oplib：管道执行入参
package service

// PipeRunIn 管道执行请求
type PipeRunIn struct {
	ID    uint   `json:"id" binding:"required"`
	Input string `json:"input"`
}
