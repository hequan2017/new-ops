// Package service llmops：编目治理（分类目录树、资产挂分类/打标签、生命周期事件流水）
package service

import (
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
)

// 错误码（llmops 段续用）
const (
	ErrCodeCatNotFound  = 2218
	ErrCodeCatHasChild  = 2219
	ErrCodeCatCycle     = 2220
	ErrCodeCatInUse     = 2221
	ErrCodeAssetEvent   = 2222
)

// AssetEventActions 生命周期事件动作白名单
var AssetEventActions = []string{"collect", "parse", "catalog", "tag", "update", "offline", "restore", "delete"}

// ValidAssetAction 事件动作白名单校验
func ValidAssetAction(a string) bool {
	for _, v := range AssetEventActions {
		if v == a {
			return true
		}
	}
	return false
}

// recordAssetEvent 写事件流水（失败仅记日志不阻断主流程）
func recordAssetEvent(db *gorm.DB, assetId uint, action, detail string) {
	if db == nil {
		db = global.GVA_DB
	}
	if err := db.Create(&model.LlmopsAssetEvent{AssetId: assetId, Action: action, Detail: detail}).Error; err != nil {
		global.GVA_LOG.Error(fmt.Sprintf("llmops 资产事件写入失败 asset=%d action=%s: %v", assetId, action, err))
	}
}

// CategoryNode 分类树节点
type CategoryNode struct {
	model.LlmopsCategory
	Children []CategoryNode `json:"children"`
}

// BuildCategoryTree 平表→树（纯函数，测试锚点）：根节点正常建树；
// 孤儿（父缺失）与环引用成员挂根降级为平铺，数据不丢；depth 上限防超深递归。
func BuildCategoryTree(cats []model.LlmopsCategory) []CategoryNode {
	byId := make(map[uint]*model.LlmopsCategory, len(cats))
	for i := range cats {
		byId[cats[i].ID] = &cats[i]
	}
	childrenOf := make(map[uint][]uint)
	for _, c := range cats {
		childrenOf[c.ParentId] = append(childrenOf[c.ParentId], c.ID)
	}
	visited := make(map[uint]bool, len(cats))
	var build func(id uint, depth int) *CategoryNode
	build = func(id uint, depth int) *CategoryNode {
		if visited[id] || depth > 32 { // 环引用/超深防护
			return nil
		}
		visited[id] = true
		c := byId[id]
		node := &CategoryNode{LlmopsCategory: *c, Children: []CategoryNode{}}
		for _, cid := range childrenOf[id] {
			if child := build(cid, depth+1); child != nil {
				node.Children = append(node.Children, *child)
			}
		}
		return node
	}
	var roots []CategoryNode
	for _, c := range cats { // 第一遍：正常根
		if c.ParentId == 0 {
			if n := build(c.ID, 0); n != nil {
				roots = append(roots, *n)
			}
		}
	}
	for _, c := range cats { // 第二遍：孤儿与环引用成员挂根，不丢数据
		if !visited[c.ID] {
			if n := build(c.ID, 0); n != nil {
				roots = append(roots, *n)
			}
		}
	}
	return roots
}

// SaveCategory 新建/更新分类（名称必填；父分类存在性 + 环检测）
func (s *LlmopsServiceStruct) SaveCategory(cat *model.LlmopsCategory) error {
	if cat == nil || strings.TrimSpace(cat.Name) == "" {
		return fmt.Errorf("[%d]分类名不能为空", ErrCodeCatNotFound)
	}
	if cat.ParentId != 0 {
		var parent model.LlmopsCategory
		if err := global.GVA_DB.First(&parent, cat.ParentId).Error; err != nil {
			return fmt.Errorf("[%d]父分类不存在", ErrCodeCatNotFound)
		}
	}
	if cat.ID > 0 {
		// 环检测：沿父链上溯不得回到自身
		pid := cat.ParentId
		for i := 0; pid != 0 && i < 64; i++ {
			if pid == cat.ID {
				return fmt.Errorf("[%d]父分类不可形成环", ErrCodeCatCycle)
			}
			var p model.LlmopsCategory
			if err := global.GVA_DB.First(&p, pid).Error; err != nil {
				break
			}
			pid = p.ParentId
		}
		return global.GVA_DB.Model(cat).Updates(map[string]any{
			"name": cat.Name, "parent_id": cat.ParentId, "sort": cat.Sort, "remark": cat.Remark,
		}).Error
	}
	return global.GVA_DB.Create(cat).Error
}

// GetCategoryTree 分类树
func (s *LlmopsServiceStruct) GetCategoryTree() ([]CategoryNode, error) {
	var cats []model.LlmopsCategory
	if err := global.GVA_DB.Order("sort asc, id asc").Find(&cats).Error; err != nil {
		return nil, err
	}
	return BuildCategoryTree(cats), nil
}

// DeleteCategory 删除分类（有子分类或有资产引用则拒绝）
func (s *LlmopsServiceStruct) DeleteCategory(id uint) error {
	var childCount int64
	global.GVA_DB.Model(&model.LlmopsCategory{}).Where("parent_id = ?", id).Count(&childCount)
	if childCount > 0 {
		return fmt.Errorf("[%d]存在子分类，不可删除", ErrCodeCatHasChild)
	}
	var assetCount int64
	global.GVA_DB.Model(&model.LlmopsDataAsset{}).Where("category_id = ?", id).Count(&assetCount)
	if assetCount > 0 {
		return fmt.Errorf("[%d]分类下仍有 %d 个资产，请先移出", ErrCodeCatInUse, assetCount)
	}
	res := global.GVA_DB.Delete(&model.LlmopsCategory{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("[%d]分类不存在", ErrCodeCatNotFound)
	}
	return nil
}

// AssignAssetCategory 资产挂分类/打标签（同时记 tag/catalog 事件流水）
func (s *LlmopsServiceStruct) AssignAssetCategory(assetId, categoryId uint, tags string) error {
	var asset model.LlmopsDataAsset
	if err := global.GVA_DB.First(&asset, assetId).Error; err != nil {
		return fmt.Errorf("[%d]数据资产不存在", ErrCodeAssetNotFound)
	}
	if categoryId != 0 {
		var cat model.LlmopsCategory
		if err := global.GVA_DB.First(&cat, categoryId).Error; err != nil {
			return fmt.Errorf("[%d]分类不存在", ErrCodeCatNotFound)
		}
	}
	updates := map[string]any{}
	if categoryId > 0 || categoryId == 0 {
		updates["category_id"] = categoryId
	}
	if tags != "" || asset.Tags != "" {
		updates["tags"] = strings.TrimSpace(tags)
	}
	if err := global.GVA_DB.Model(&asset).Updates(updates).Error; err != nil {
		return err
	}
	action := "tag"
	if categoryId > 0 {
		action = "catalog"
	}
	recordAssetEvent(global.GVA_DB, assetId, action, fmt.Sprintf("categoryId=%d tags=%s", categoryId, updates["tags"]))
	return nil
}

// GetAssetEvents 资产事件流水（倒序分页）
func (s *LlmopsServiceStruct) GetAssetEvents(assetId uint, page, pageSize int) ([]model.LlmopsAssetEvent, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	db := global.GVA_DB.Model(&model.LlmopsAssetEvent{}).Where("asset_id = ?", assetId)
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.LlmopsAssetEvent
	err := db.Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}
