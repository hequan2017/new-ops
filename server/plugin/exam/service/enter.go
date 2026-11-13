// Package service exam：考试域服务（题库/组卷/场次/考生/答题/判分/成绩/调度/防作弊/实操/续考）
package service

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/exam/repo"
	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/errx"
)

// ErrCodeExamStore 落库类错误统一码（exam 段 2400-2499 内首个未用码：
// 2400-2420 为既有业务码，2421 保留给裸 gorm 错误的 Wrap 归口）
const ErrCodeExamStore = 2421

// examErr 段内业务错误构造：消息保留 "[code]" 前缀（既有单测断言 2413/2417 的口径），
// 码同时可经 errx.CodeOf 提取，api 层 errx.Respond 以分段码响应。
func examErr(code int, format string, args ...any) *errx.AppError {
	return errx.New(code, fmt.Sprintf("[%d]", code)+fmt.Sprintf(format, args...))
}

// examWrapErr 段内业务错误带因构造：消息 "[code]msg: 原因"，Unwrap 链保留底层错误。
func examWrapErr(code int, msg string, cause error) *errx.AppError {
	return errx.Wrap(code, fmt.Sprintf("[%d]", code)+msg, cause)
}

// ExamService exam 服务（依赖注入：db 为 nil 时惰性回退底座全局库——
// 包级单例构造早于底座建库，禁止在构造期捕获 GVA_DB）
type ExamService struct {
	db   *gorm.DB
	repo repo.Repo
}

// NewExamService 构造：显式传 db 供单测内存库注入，生产传 nil
func NewExamService(db *gorm.DB) *ExamService {
	return &ExamService{db: db}
}

// dbh 当前会话：显式注入优先，否则底座全局库
func (s *ExamService) dbh() *gorm.DB {
	if s.db != nil {
		return s.db
	}
	return global.GVA_DB
}

// ExamApp 服务单例出口（api 层引用标识不变）
var ExamApp = NewExamService(nil)
