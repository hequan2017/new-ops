// Package initialize asset 插件授权种子（幂等）
// 为超级管理员角色（888）补齐 API casbin 策略与菜单-角色绑定
package initialize

import (
	"context"
	"fmt"

	"github.com/hequan2017/new-ops/server/global"
	model "github.com/hequan2017/new-ops/server/model/system"
	"github.com/hequan2017/new-ops/server/utils"
	"go.uber.org/zap"
)

const superAdminAuthorityId = "888"

// assetApis 与 api.go 保持一致（路径+方法）
var assetApis = []struct {
	Path   string
	Method string
}{
	{"/asset/host/create", "POST"},
	{"/asset/host/delete", "DELETE"},
	{"/asset/host/deleteByIds", "DELETE"},
	{"/asset/host/update", "PUT"},
	{"/asset/host/find", "GET"},
	{"/asset/host/list", "POST"},
}

// assetMenus 与 menu.go 保持一致（菜单 name）
var assetMenus = []string{"asset", "assetHost"}

// Casbin 为超管角色补齐策略与菜单绑定（幂等）
func Casbin(ctx context.Context) {
	e := utils.GetCasbin()
	if e == nil {
		zap.L().Warn("asset 插件：casbin 未初始化，跳过策略注册")
		return
	}
	for _, api := range assetApis {
		has, err := e.HasPolicy(superAdminAuthorityId, api.Path, api.Method)
		if err != nil {
			zap.L().Error("asset 插件：查询 casbin 策略失败", zap.Error(err), zap.String("path", api.Path))
			continue
		}
		if has {
			continue
		}
		if _, err := e.AddPolicy(superAdminAuthorityId, api.Path, api.Method); err != nil {
			zap.L().Error("asset 插件：添加 casbin 策略失败", zap.Error(err), zap.String("path", api.Path))
		}
	}

	// 菜单绑定到 888
	bindMenusToSuperAdmin(ctx)
}

// bindMenusToSuperAdmin 将资产中心菜单绑定到超管角色（幂等）
func bindMenusToSuperAdmin(ctx context.Context) {
	for _, name := range assetMenus {
		var menu model.SysBaseMenu
		if err := global.GVA_DB.WithContext(ctx).Where("name = ?", name).First(&menu).Error; err != nil {
			zap.L().Warn(fmt.Sprintf("asset 插件：菜单 %s 未找到，跳过角色绑定", name))
			continue
		}
		var count int64
		global.GVA_DB.Model(&model.SysAuthorityMenu{}).
			Where("sys_base_menu_id = ? AND sys_authority_authority_id = ?", menu.ID, superAdminAuthorityId).
			Count(&count)
		if count > 0 {
			continue
		}
		binding := model.SysAuthorityMenu{
			MenuId:      fmt.Sprint(menu.ID),
			AuthorityId: superAdminAuthorityId,
		}
		if err := global.GVA_DB.Create(&binding).Error; err != nil {
			zap.L().Error("asset 插件：菜单绑定角色失败", zap.Error(err), zap.String("menu", name))
		}
	}
}
