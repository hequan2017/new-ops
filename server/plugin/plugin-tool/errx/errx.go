// Package errx 白泽插件统一业务错误契约：分段错误码 + AppError + gin 响应映射。
//
// 各插件沿用 README「接口与错误码规范」的分段错误码（asset=1000-1099 … policyreg=3000-3099）。
// 业务层返回 *AppError（替代此前 typed struct / 错误码拼字符串 / 丢弃错误 三种并存的形态），
// api 层统一 Respond 出响应，使分段错误码真正到达前端；非业务错误退化为通用失败码 7。
package errx

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/hequan2017/new-ops/server/model/common/response"
)

// AppError 业务错误：分段错误码 + 用户可读消息（+可选底层原因）。
type AppError struct {
	Code  int
	Msg   string
	cause error
}

func (e *AppError) Error() string { return e.Msg }

// Unwrap 保留错误链，供 errors.Is/As 追溯底层原因。
func (e *AppError) Unwrap() error { return e.cause }

// New 构造业务错误。
func New(code int, msg string) *AppError { return &AppError{Code: code, Msg: msg} }

// Wrap 构造带因错误：消息追加「: 原因」，Unwrap 链保留底层错误。
func Wrap(code int, msg string, cause error) *AppError {
	if cause == nil {
		return &AppError{Code: code, Msg: msg}
	}
	return &AppError{Code: code, Msg: msg + ": " + cause.Error(), cause: cause}
}

// CodeOf 提取分段错误码：链上存在 *AppError 返回其码，否则 0。
func CodeOf(err error) int {
	var ae *AppError
	if errors.As(err, &ae) {
		return ae.Code
	}
	return 0
}

// Respond 统一错误→响应映射（沿用 {code,data,msg} 响应结构，HTTP 200）：
// *AppError 以其分段码与消息响应；其余错误以通用失败码 7 返回 fallback 文案。
// 成功路径不经过本函数。
func Respond(c *gin.Context, err error, fallback string) {
	var ae *AppError
	if errors.As(err, &ae) {
		response.Result(ae.Code, map[string]interface{}{}, ae.Msg, c)
		return
	}
	response.FailWithMessage(fallback, c)
}
