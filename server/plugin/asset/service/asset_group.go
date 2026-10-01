package service

import (
	"fmt"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/asset/model"
	"gorm.io/gorm"
)

// asset 插件错误码补充（资产组）
const (
	ErrCodeGroupNameRequired = 1012
	ErrCodeGroupDuplicate    = 1013
)

// AssetGroupService 资产组服务（数据权限单元）
type AssetGroupService struct{}

// validateGroup 资产组校验（纯逻辑）
func validateGroup(g *model.AssetGroup) error {
	if g == nil || g.Name == "" {
		return newServiceErr(ErrCodeGroupNameRequired, "资产组名称不能为空")
	}
	return nil
}

// CreateAssetGroup 创建资产组（名称唯一，可同时绑定主机与用户）
func (s *AssetGroupService) CreateAssetGroup(g *model.AssetGroup, hostIDs, userIDs []uint) error {
	if err := validateGroup(g); err != nil {
		return err
	}
	var count int64
	global.GVA_DB.Model(&model.AssetGroup{}).Where("name = ?", g.Name).Count(&count)
	if count > 0 {
		return newServiceErr(ErrCodeGroupDuplicate, fmt.Sprintf("资产组已存在: %s", g.Name))
	}
	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(g).Error; err != nil {
			return err
		}
		return s.replaceGroupRelations(tx, g.ID, hostIDs, userIDs)
	})
}

// DeleteAssetGroup 删除资产组（联动清理关联）
func (s *AssetGroupService) DeleteAssetGroup(id uint) error {
	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("group_id = ?", id).Delete(&model.AssetGroupHost{}).Error; err != nil {
			return err
		}
		if err := tx.Where("group_id = ?", id).Delete(&model.AssetGroupUser{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.AssetGroup{}, id).Error
	})
}

// UpdateAssetGroup 更新资产组（名称唯一检查 + 关联重写）
func (s *AssetGroupService) UpdateAssetGroup(g *model.AssetGroup, hostIDs, userIDs []uint) error {
	if err := validateGroup(g); err != nil {
		return err
	}
	var count int64
	global.GVA_DB.Model(&model.AssetGroup{}).Where("name = ? AND id <> ?", g.Name, g.ID).Count(&count)
	if count > 0 {
		return newServiceErr(ErrCodeGroupDuplicate, fmt.Sprintf("资产组已存在: %s", g.Name))
	}
	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.AssetGroup{}).Where("id = ?", g.ID).Omit("ID").Updates(map[string]any{
			"name":  g.Name,
			"notes": g.Notes,
		}).Error; err != nil {
			return err
		}
		return s.replaceGroupRelations(tx, g.ID, hostIDs, userIDs)
	})
}

// replaceGroupRelations 重写资产组的主机与用户关联
func (s *AssetGroupService) replaceGroupRelations(tx *gorm.DB, groupID uint, hostIDs, userIDs []uint) error {
	if err := tx.Where("group_id = ?", groupID).Delete(&model.AssetGroupHost{}).Error; err != nil {
		return err
	}
	if err := tx.Where("group_id = ?", groupID).Delete(&model.AssetGroupUser{}).Error; err != nil {
		return err
	}
	for _, hid := range hostIDs {
		if err := tx.Create(&model.AssetGroupHost{GroupID: groupID, HostID: hid}).Error; err != nil {
			return err
		}
	}
	for _, uid := range userIDs {
		if err := tx.Create(&model.AssetGroupUser{GroupID: groupID, UserID: uid}).Error; err != nil {
			return err
		}
	}
	return nil
}

// AssetGroupWithRelations 资产组列表项（含关联 ID）
type AssetGroupWithRelations struct {
	model.AssetGroup
	HostIDs  []uint `json:"hostIds"`
	UserIDs  []uint `json:"userIds"`
	HostNum  int64  `json:"hostNum"`
	UserNum  int64  `json:"userNum"`
}

// GetAssetGroupList 资产组全量列表（含关联 ID 与计数）
func (s *AssetGroupService) GetAssetGroupList() (list []*AssetGroupWithRelations, err error) {
	var groups []model.AssetGroup
	if err = global.GVA_DB.Order("id DESC").Find(&groups).Error; err != nil {
		return nil, err
	}
	for i := range groups {
		item := &AssetGroupWithRelations{AssetGroup: groups[i]}
		global.GVA_DB.Model(&model.AssetGroupHost{}).Where("group_id = ?", groups[i].ID).
			Pluck("host_id", &item.HostIDs)
		global.GVA_DB.Model(&model.AssetGroupUser{}).Where("group_id = ?", groups[i].ID).
			Pluck("user_id", &item.UserIDs)
		item.HostNum = int64(len(item.HostIDs))
		item.UserNum = int64(len(item.UserIDs))
		list = append(list, item)
	}
	return list, nil
}

// GetAssetGroupUserIDs 查询用户所属资产组绑定的全部主机 ID（数据权限过滤用）
func (s *AssetGroupService) GetUserAuthorizedHostIDs(userID uint) ([]uint, error) {
	ids := []uint{}
	err := global.GVA_DB.
		Table("asset_group_hosts").
		Joins("JOIN asset_group_users ON asset_group_users.group_id = asset_group_hosts.group_id").
		Where("asset_group_users.user_id = ?", userID).
		Distinct().Pluck("asset_group_hosts.host_id", &ids).Error
	return ids, err
}
