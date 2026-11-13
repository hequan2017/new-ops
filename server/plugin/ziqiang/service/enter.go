// Package service ziqiang 域服务（自强-理工科大模型：文本/图像鉴权语料、术语校验与
// 差分脱敏、法律分词与抄袭识别、实体图谱、知识库检索、区块链存证、提示词库、知识产权问答）
package service

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/errx"
	"github.com/hequan2017/new-ops/server/plugin/ziqiang/repo"
)

// ErrCodeZqStore 落库类错误统一码（ziqiang 段 2700-2799 内首个未用码：
// 2700-2718 为既有业务码，2719 保留给裸 gorm 错误的 Wrap 归口）
const ErrCodeZqStore = 2719

// zqErr 段内业务错误构造：消息保留 "[code]" 前缀（既有单测断言口径），
// 码同时可经 errx.CodeOf 提取，api 层 errx.Respond 以分段码响应。
func zqErr(code int, format string, args ...any) *errx.AppError {
	return errx.New(code, fmt.Sprintf("[%d]", code)+fmt.Sprintf(format, args...))
}

// zqWrapErr 段内业务错误带因构造：消息 "[code]msg: 原因"，Unwrap 链保留底层错误。
func zqWrapErr(code int, msg string, cause error) *errx.AppError {
	return errx.Wrap(code, fmt.Sprintf("[%d]", code)+msg, cause)
}

// ZiqiangService ziqiang 服务（依赖注入：db 为 nil 时惰性回退底座全局库——
// 包级单例构造早于底座建库，禁止在构造期捕获 GVA_DB）
type ZiqiangService struct {
	db   *gorm.DB
	repo repo.Repo
}

// NewZiqiangService 构造：显式传 db 供单测内存库注入，生产传 nil
func NewZiqiangService(db *gorm.DB) *ZiqiangService {
	return &ZiqiangService{db: db}
}

// dbh 当前会话：显式注入优先，否则底座全局库
func (s *ZiqiangService) dbh() *gorm.DB {
	if s.db != nil {
		return s.db
	}
	return global.GVA_DB
}

// ZiqiangApp 服务单例出口（自强域服务聚合）
var ZiqiangApp = NewZiqiangService(nil)
