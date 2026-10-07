// Package service job 插件业务逻辑与错误码
// 错误码分段：1201-1207 脚本库/变量组（job_script.go）、1211-1214 批量执行
package service

// 批量执行错误码
const (
	ErrCodeParamInvalid      = 1211
	ErrCodeHostNotAuthorized = 1212
	ErrCodeNoCredential      = 1213
	ErrCodeBatchNotRunning   = 1214
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
