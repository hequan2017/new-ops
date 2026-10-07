package service

import (
	"encoding/json"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/model/common/request"
	"github.com/hequan2017/new-ops/server/plugin/asset/model"
	"gorm.io/gorm"
)

// recordHostHistory 在事务内记录主机资产变更历史（失败仅记日志不阻断主流程）
func recordHostHistory(tx *gorm.DB, hostID uint, action string, snapshot any, operator string) {
	b, err := json.Marshal(snapshot)
	if err != nil {
		global.GVA_LOG.Warn("asset 历史快照序列化失败: " + err.Error())
		return
	}
	if err := tx.Create(&model.AssetHostHistory{
		HostID:   hostID,
		Action:   action,
		Snapshot: string(b),
		Operator: operator,
	}).Error; err != nil {
		global.GVA_LOG.Warn("asset 历史记录写入失败: " + err.Error())
	}
}

// GetAssetHostHistoryList 分页查询主机变更历史（新→旧）
func (s *AssetHostService) GetAssetHostHistoryList(hostID uint, info request.PageInfo) (list []*model.AssetHostHistory, total int64, err error) {
	db := global.GVA_DB.Model(&model.AssetHostHistory{}).Where("host_id = ?", hostID)
	if err = db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err = db.Scopes(info.Paginate()).Order("id DESC").Find(&list).Error
	return list, total, err
}
