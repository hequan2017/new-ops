// Package service artsci 域服务（文理学科 AI 工具集：求解器/批改/出题组卷/文生图场景）
package service

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/artsci/repo"
	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/errx"
)

// ErrCodeArtsciStore 落库类错误统一码（artsci 段 2500-2599 内首个未用码：
// 2500-2518 为既有业务码，2519 保留给裸 gorm 错误的 Wrap 归口）
const ErrCodeArtsciStore = 2519

// artErr 段内业务错误构造：消息保留 "[code]" 前缀（既有单测断言口径），
// 码同时可经 errx.CodeOf 提取，api 层 errx.Respond 以分段码响应。
func artErr(code int, format string, args ...any) *errx.AppError {
	return errx.New(code, fmt.Sprintf("[%d]", code)+fmt.Sprintf(format, args...))
}

// artWrapErr 段内业务错误带因构造：消息 "[code]msg: 原因"，Unwrap 链保留底层错误。
func artWrapErr(code int, msg string, cause error) *errx.AppError {
	return errx.Wrap(code, fmt.Sprintf("[%d]", code)+msg, cause)
}

// ArtsciService artsci 服务（依赖注入：db 为 nil 时惰性回退底座全局库——
// 包级单例构造早于底座建库，禁止在构造期捕获 GVA_DB）
type ArtsciService struct {
	db   *gorm.DB
	repo repo.Repo
}

// NewArtsciService 构造：显式传 db 供单测内存库注入，生产传 nil
func NewArtsciService(db *gorm.DB) *ArtsciService {
	return &ArtsciService{db: db}
}

// dbh 当前会话：显式注入优先，否则底座全局库
func (s *ArtsciService) dbh() *gorm.DB {
	if s.db != nil {
		return s.db
	}
	return global.GVA_DB
}

// ArtsciApp 服务单例出口（工具域服务聚合）
var ArtsciApp = NewArtsciService(nil)
