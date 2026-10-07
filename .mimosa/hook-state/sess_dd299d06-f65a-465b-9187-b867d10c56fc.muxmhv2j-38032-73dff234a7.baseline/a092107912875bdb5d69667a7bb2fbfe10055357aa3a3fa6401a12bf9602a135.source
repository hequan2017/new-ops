package service

import (
	"errors"
	"fmt"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/asset/model"
	"gorm.io/gorm"
)

// asset 插件错误码补充（机房/机柜/产品线）
const (
	ErrCodeRoomNameRequired  = 1005
	ErrCodeRoomDuplicate     = 1006
	ErrCodeRackRoomNotFound  = 1007
	ErrCodePLNameRequired    = 1008
	ErrCodePLDuplicate       = 1009
	ErrCodeRoomHasRacks      = 1010
	ErrCodeRackNameRequired  = 1011
)

// AssetRoomService 机房服务
type AssetRoomService struct{}

// validateRoom 机房校验（纯逻辑）
func validateRoom(r *model.AssetRoom) error {
	if r == nil || r.Name == "" {
		return newServiceErr(ErrCodeRoomNameRequired, "机房名称不能为空")
	}
	return nil
}

// CreateAssetRoom 创建机房（名称唯一）
func (s *AssetRoomService) CreateAssetRoom(r *model.AssetRoom) error {
	if err := validateRoom(r); err != nil {
		return err
	}
	var count int64
	global.GVA_DB.Model(&model.AssetRoom{}).Where("name = ?", r.Name).Count(&count)
	if count > 0 {
		return newServiceErr(ErrCodeRoomDuplicate, fmt.Sprintf("机房已存在: %s", r.Name))
	}
	return global.GVA_DB.Create(r).Error
}

// DeleteAssetRoom 删除机房（存在机柜时禁止删除）
func (s *AssetRoomService) DeleteAssetRoom(id uint) error {
	var count int64
	global.GVA_DB.Model(&model.AssetRack{}).Where("room_id = ?", id).Count(&count)
	if count > 0 {
		return newServiceErr(ErrCodeRoomHasRacks, "该机房下存在机柜，请先删除机柜")
	}
	return global.GVA_DB.Delete(&model.AssetRoom{}, id).Error
}

// UpdateAssetRoom 更新机房
func (s *AssetRoomService) UpdateAssetRoom(r *model.AssetRoom) error {
	if err := validateRoom(r); err != nil {
		return err
	}
	var count int64
	global.GVA_DB.Model(&model.AssetRoom{}).Where("name = ? AND id <> ?", r.Name, r.ID).Count(&count)
	if count > 0 {
		return newServiceErr(ErrCodeRoomDuplicate, fmt.Sprintf("机房已存在: %s", r.Name))
	}
	return global.GVA_DB.Model(&model.AssetRoom{}).Where("id = ?", r.ID).Omit("ID").Updates(map[string]any{
		"name":    r.Name,
		"region":  r.Region,
		"address": r.Address,
		"notes":   r.Notes,
	}).Error
}

// GetAssetRoomList 机房全量列表（机房数量有限，不分页，供下拉与页面使用）
func (s *AssetRoomService) GetAssetRoomList(keyword string) (list []*model.AssetRoom, err error) {
	db := global.GVA_DB.Model(&model.AssetRoom{})
	if keyword != "" {
		kw := "%" + keyword + "%"
		db = db.Where("name LIKE ? OR region LIKE ?", kw, kw)
	}
	err = db.Order("id DESC").Find(&list).Error
	return list, err
}

// AssetRackService 机柜服务
type AssetRackService struct{}

// validateRack 机柜校验（纯逻辑）
func validateRack(r *model.AssetRack) error {
	if r == nil || r.Name == "" {
		return newServiceErr(ErrCodeRackNameRequired, "机柜名称不能为空")
	}
	if r.RoomID == 0 {
		return newServiceErr(ErrCodeRackRoomNotFound, "请选择所属机房")
	}
	if r.TotalU < 0 {
		return newServiceErr(ErrCodeRoomNameRequired, "总U数不能为负")
	}
	return nil
}

// CreateAssetRack 创建机柜（校验机房存在）
func (s *AssetRackService) CreateAssetRack(r *model.AssetRack) error {
	if err := validateRack(r); err != nil {
		return err
	}
	var room model.AssetRoom
	if err := global.GVA_DB.First(&room, r.RoomID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return newServiceErr(ErrCodeRackRoomNotFound, "所属机房不存在")
		}
		return err
	}
	return global.GVA_DB.Create(r).Error
}

// DeleteAssetRack 删除机柜
func (s *AssetRackService) DeleteAssetRack(id uint) error {
	return global.GVA_DB.Delete(&model.AssetRack{}, id).Error
}

// UpdateAssetRack 更新机柜
func (s *AssetRackService) UpdateAssetRack(r *model.AssetRack) error {
	if err := validateRack(r); err != nil {
		return err
	}
	return global.GVA_DB.Model(&model.AssetRack{}).Where("id = ?", r.ID).Omit("ID").Updates(map[string]any{
		"room_id": r.RoomID,
		"name":    r.Name,
		"total_u": r.TotalU,
		"notes":   r.Notes,
	}).Error
}

// GetAssetRackList 机柜列表（可按机房过滤）
func (s *AssetRackService) GetAssetRackList(roomID *uint) (list []*model.AssetRack, err error) {
	db := global.GVA_DB.Model(&model.AssetRack{})
	if roomID != nil && *roomID > 0 {
		db = db.Where("room_id = ?", *roomID)
	}
	err = db.Preload("Room").Order("id DESC").Find(&list).Error
	return list, err
}

// AssetProductLineService 产品线服务
type AssetProductLineService struct{}

// validateProductLine 产品线校验（纯逻辑）
func validateProductLine(p *model.AssetProductLine) error {
	if p == nil || p.Name == "" {
		return newServiceErr(ErrCodePLNameRequired, "产品线名称不能为空")
	}
	return nil
}

// CreateAssetProductLine 创建产品线（名称唯一）
func (s *AssetProductLineService) CreateAssetProductLine(p *model.AssetProductLine) error {
	if err := validateProductLine(p); err != nil {
		return err
	}
	var count int64
	global.GVA_DB.Model(&model.AssetProductLine{}).Where("name = ?", p.Name).Count(&count)
	if count > 0 {
		return newServiceErr(ErrCodePLDuplicate, fmt.Sprintf("产品线已存在: %s", p.Name))
	}
	return global.GVA_DB.Create(p).Error
}

// DeleteAssetProductLine 删除产品线（先清空主机关联）
func (s *AssetProductLineService) DeleteAssetProductLine(id uint) error {
	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		var pl model.AssetProductLine
		if err := tx.First(&pl, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return err
		}
		if err := tx.Model(&pl).Association("Hosts").Clear(); err != nil {
			return err
		}
		return tx.Delete(&pl).Error
	})
}

// UpdateAssetProductLine 更新产品线
func (s *AssetProductLineService) UpdateAssetProductLine(p *model.AssetProductLine) error {
	if err := validateProductLine(p); err != nil {
		return err
	}
	var count int64
	global.GVA_DB.Model(&model.AssetProductLine{}).Where("name = ? AND id <> ?", p.Name, p.ID).Count(&count)
	if count > 0 {
		return newServiceErr(ErrCodePLDuplicate, fmt.Sprintf("产品线已存在: %s", p.Name))
	}
	return global.GVA_DB.Model(&model.AssetProductLine{}).Where("id = ?", p.ID).Omit("ID").Updates(map[string]any{
		"name":  p.Name,
		"owner": p.Owner,
		"level": p.Level,
		"notes": p.Notes,
	}).Error
}

// GetAssetProductLineList 产品线全量列表
func (s *AssetProductLineService) GetAssetProductLineList(keyword string) (list []*model.AssetProductLine, err error) {
	db := global.GVA_DB.Model(&model.AssetProductLine{})
	if keyword != "" {
		kw := "%" + keyword + "%"
		db = db.Where("name LIKE ? OR owner LIKE ?", kw, kw)
	}
	err = db.Order("id DESC").Find(&list).Error
	return list, err
}
