package service

import (
	"errors"
	"fmt"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/model/common/request"
	"github.com/hequan2017/new-ops/server/plugin/asset/model"
	"gorm.io/gorm"
)

// AssetHostService 主机资产服务
type AssetHostService struct{}

// CreateAssetHost 创建主机资产（校验 + IP 唯一性 + 产品线关联 + 变更历史）
func (s *AssetHostService) CreateAssetHost(h *model.AssetHost, operator string) error {
	if err := validateHost(h); err != nil {
		return err
	}
	var count int64
	global.GVA_DB.Model(&model.AssetHost{}).Where("ip = ?", h.IP).Count(&count)
	if count > 0 {
		return newServiceErr(ErrCodeIPDuplicate, fmt.Sprintf("内网IP已存在: %s", h.IP))
	}
	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(h).Error; err != nil {
			return err
		}
		if err := s.replaceProductLines(tx, h); err != nil {
			return err
		}
		recordHostHistory(tx, h.ID, model.HostHistoryCreate, h, operator)
		return nil
	})
}

// DeleteAssetHost 删除主机资产（软删除，联动清理产品线关联 + 变更历史）
func (s *AssetHostService) DeleteAssetHost(id uint, operator string) error {
	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		var host model.AssetHost
		if err := tx.First(&host, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return err
		}
		if err := tx.Model(&host).Association("ProductLines").Clear(); err != nil {
			return err
		}
		if err := tx.Delete(&host).Error; err != nil {
			return err
		}
		recordHostHistory(tx, host.ID, model.HostHistoryDelete, host, operator)
		return nil
	})
}

// DeleteAssetHostByIds 批量删除主机资产
func (s *AssetHostService) DeleteAssetHostByIds(ids []uint, operator string) error {
	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		for _, id := range ids {
			var host model.AssetHost
			if err := tx.First(&host, id).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					continue
				}
				return err
			}
			if err := tx.Model(&host).Association("ProductLines").Clear(); err != nil {
				return err
			}
			if err := tx.Delete(&host).Error; err != nil {
				return err
			}
			recordHostHistory(tx, host.ID, model.HostHistoryDelete, host, operator)
		}
		return nil
	})
}

// UpdateAssetHost 更新主机资产（校验 + IP 占用检查 + 关联重写 + 变更历史）
func (s *AssetHostService) UpdateAssetHost(h *model.AssetHost, operator string) error {
	if err := validateHost(h); err != nil {
		return err
	}
	var count int64
	global.GVA_DB.Model(&model.AssetHost{}).Where("ip = ? AND id <> ?", h.IP, h.ID).Count(&count)
	if count > 0 {
		return newServiceErr(ErrCodeIPDuplicate, fmt.Sprintf("内网IP已被其他资产占用: %s", h.IP))
	}
	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.AssetHost{}).Where("id = ?", h.ID).Omit("ProductLines").Updates(h).Error; err != nil {
			return err
		}
		if err := s.replaceProductLines(tx, h); err != nil {
			return err
		}
		recordHostHistory(tx, h.ID, model.HostHistoryUpdate, h, operator)
		return nil
	})
}

// GetAssetHost 根据 ID 获取主机资产（含机房/机柜/产品线）
func (s *AssetHostService) GetAssetHost(id uint) (h *model.AssetHost, err error) {
	h = &model.AssetHost{}
	err = global.GVA_DB.Preload("Room").Preload("Rack").Preload("ProductLines").First(h, id).Error
	return h, err
}

// GetAssetHostList 分页查询主机资产
// 过滤：keyword（主机名/IP/SN/负责人模糊）、status、roomId、productLineId
func (s *AssetHostService) GetAssetHostList(info request.PageInfo, status string, roomID, productLineID *uint) (list []*model.AssetHost, total int64, err error) {
	db := global.GVA_DB.Model(&model.AssetHost{})
	if info.Keyword != "" {
		kw := "%" + info.Keyword + "%"
		db = db.Where("hostname LIKE ? OR ip LIKE ? OR sn LIKE ? OR owner LIKE ?", kw, kw, kw, kw)
	}
	if status != "" {
		db = db.Where("status = ?", status)
	}
	if roomID != nil && *roomID > 0 {
		db = db.Where("room_id = ?", *roomID)
	}
	if productLineID != nil && *productLineID > 0 {
		db = db.Joins("JOIN asset_host_product_lines ON asset_host_product_lines.asset_host_id = asset_hosts.id").
			Where("asset_host_product_lines.asset_product_line_id = ?", *productLineID)
	}
	if err = db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err = db.Scopes(info.Paginate()).Preload("Room").Preload("Rack").Preload("ProductLines").
		Order("id DESC").Find(&list).Error
	return list, total, err
}

// replaceProductLines 重写主机-产品线关联（按传入 ProductLines 的 ID）
func (s *AssetHostService) replaceProductLines(tx *gorm.DB, h *model.AssetHost) error {
	if len(h.ProductLines) == 0 {
		if err := tx.Model(h).Association("ProductLines").Clear(); err != nil {
			return err
		}
		return nil
	}
	ids := make([]uint, 0, len(h.ProductLines))
	for _, pl := range h.ProductLines {
		ids = append(ids, pl.ID)
	}
	var lines []model.AssetProductLine
	if err := tx.Find(&lines, ids).Error; err != nil {
		return err
	}
	return tx.Model(h).Association("ProductLines").Replace(&lines)
}
