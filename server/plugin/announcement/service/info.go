package service

import (
	"gorm.io/gorm"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/announcement/model"
	"github.com/hequan2017/new-ops/server/plugin/announcement/model/request"
	"github.com/hequan2017/new-ops/server/plugin/announcement/repo"
)

// InfoService 公告服务（依赖注入：db 为 nil 时惰性回退底座全局库——
// 包级单例构造早于底座建库，禁止在构造期捕获 GVA_DB）
type InfoService struct {
	db   *gorm.DB
	repo repo.Repo
}

// NewInfoService 构造：显式传 db 供单测内存库注入，生产传 nil
func NewInfoService(db *gorm.DB) *InfoService {
	return &InfoService{db: db}
}

// dbh 当前会话：显式注入优先，否则底座全局库
func (s *InfoService) dbh() *gorm.DB {
	if s.db != nil {
		return s.db
	}
	return global.GVA_DB
}

// CreateInfo 创建公告记录
func (s *InfoService) CreateInfo(info *model.Info) error {
	return s.repo.Create(s.dbh(), info)
}

// DeleteInfo 删除公告记录
func (s *InfoService) DeleteInfo(ID string) error {
	return s.repo.Delete(s.dbh(), ID)
}

// DeleteInfoByIds 批量删除公告记录
func (s *InfoService) DeleteInfoByIds(IDs []string) error {
	return s.repo.DeleteByIds(s.dbh(), IDs)
}

// UpdateInfo 更新公告记录
func (s *InfoService) UpdateInfo(info model.Info) error {
	return s.repo.Update(s.dbh(), info)
}

// GetInfo 根据ID获取公告记录
func (s *InfoService) GetInfo(ID string) (model.Info, error) {
	row, err := s.repo.FindByID(s.dbh(), ID)
	if err != nil {
		return model.Info{}, err
	}
	return *row, nil
}

// GetInfoInfoList 分页获取公告记录
func (s *InfoService) GetInfoInfoList(info request.InfoSearch) (list []model.Info, total int64, err error) {
	return s.repo.List(s.dbh(), info)
}

// GetInfoDataSource 获取数据源（发布者选项）
func (s *InfoService) GetInfoDataSource() (map[string][]map[string]any, error) {
	users, err := s.repo.UserOptions(s.dbh())
	if err != nil {
		return nil, err
	}
	return map[string][]map[string]any{"userID": users}, nil
}
