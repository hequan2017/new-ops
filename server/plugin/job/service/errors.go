// Package service job 插件业务逻辑与错误码
package service

// job 插件错误码段：1200-1299（DEV_PLAN 3.6）
const (
	ErrCodeParamInvalid     = 1203
	ErrCodeHostNotAuthorized = 1201
	ErrCodeNoCredential     = 1202
	ErrCodeBatchNotRunning  = 1204
)

// JobError 带业务错误码的作业错误
type JobError struct {
	Code int
	Msg  string
}

func (e *JobError) Error() string { return e.Msg }

func newJobErr(code int, msg string) *JobError {
	return &JobError{Code: code, Msg: msg}
}
