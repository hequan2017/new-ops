// Package service policyreg 域服务（高校制度大模型：导入/清洗/分词/实体/标注/鉴权/问答）
package service

import (
	"gorm.io/gorm"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/policyreg/repo"
)

// PolicyService policyreg 服务（依赖注入：db 为 nil 时惰性回退底座全局库——
// 包级单例构造早于底座建库，禁止在构造期捕获 GVA_DB）
type PolicyService struct {
	db   *gorm.DB
	repo repo.Repo
}

// NewPolicyService 构造：显式传 db 供单测内存库注入，生产传 nil
func NewPolicyService(db *gorm.DB) *PolicyService {
	return &PolicyService{db: db}
}

// dbh 当前会话：显式注入优先，否则底座全局库
func (s *PolicyService) dbh() *gorm.DB {
	if s.db != nil {
		return s.db
	}
	return global.GVA_DB
}

// PolicyApp 服务单例出口（policyreg 域服务聚合）
var PolicyApp = NewPolicyService(nil)
