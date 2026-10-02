// Package initialize asset 插件授权种子（幂等）
// 888（超管）全部接口；9528（普通用户）资产只读 + 数据按资产组过滤
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
const normalAuthorityId = "9528"

// assetApis 与 api.go 保持一致（路径+方法）：888 全量
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
	{"/asset/host/history", "GET"},
	{"/asset/host/export", "GET"},
	{"/asset/host/import", "POST"},
	{"/asset/host/collect", "POST"},
	{"/asset/room/create", "POST"},
	{"/asset/room/delete", "DELETE"},
	{"/asset/room/update", "PUT"},
	{"/asset/room/list", "GET"},
	{"/asset/rack/create", "POST"},
	{"/asset/rack/delete", "DELETE"},
	{"/asset/rack/update", "PUT"},
	{"/asset/rack/list", "GET"},
	{"/asset/productLine/create", "POST"},
	{"/asset/productLine/delete", "DELETE"},
	{"/asset/productLine/update", "PUT"},
	{"/asset/productLine/list", "GET"},
	{"/asset/group/create", "POST"},
	{"/asset/group/delete", "DELETE"},
	{"/asset/group/update", "PUT"},
	{"/asset/group/list", "GET"},
	{"/asset/credential/create", "POST"},
	{"/asset/credential/delete", "DELETE"},
	{"/asset/credential/update", "PUT"},
	{"/asset/credential/list", "GET"},
}

// assetReadOnlyApis 普通用户（9528）只读策略（不含凭据：凭据仅超管）
var assetReadOnlyApis = []struct {
	Path   string
	Method string
}{
	{"/asset/host/list", "POST"},
	{"/asset/host/find", "GET"},
	{"/asset/host/history", "GET"},
	{"/asset/host/export", "GET"},
	{"/asset/room/list", "GET"},
	{"/asset/rack/list", "GET"},
	{"/asset/productLine/list", "GET"},
	{"/asset/group/list", "GET"},
}

// assetMenus 与 menu.go 保持一致（菜单 name）：888/9528 双角色绑定
var assetMenus = []string{"asset", "assetHost", "assetRoom", "assetProductLine", "assetGroup", "assetCredential"}

// Casbin 注册角色策略与菜单绑定（幂等）
func Casbin(ctx context.Context) {
	e := utils.GetCasbin()
	if e == nil {
		zap.L().Warn("asset 插件：casbin 未初始化，跳过策略注册")
		return
	}
	ensurePolicies(e, superAdminAuthorityId, assetApis)
	ensurePolicies(e, normalAuthorityId, assetReadOnlyApis)
	bindMenusToAuthorities(ctx)
}

// ensurePolicies 幂等补齐角色-接口策略
func ensurePolicies(e interface {
	HasPolicy(...interface{}) (bool, error)
	AddPolicy(...interface{}) (bool, error)
}, authorityId string, apis []struct {
	Path   string
	Method string
}) {
	for _, api := range apis {
		has, err := e.HasPolicy(authorityId, api.Path, api.Method)
		if err != nil {
			zap.L().Error("asset 插件：查询 casbin 策略失败", zap.Error(err), zap.String("path", api.Path))
			continue
		}
		if has {
			continue
		}
		if _, err := e.AddPolicy(authorityId, api.Path, api.Method); err != nil {
			zap.L().Error("asset 插件：添加 casbin 策略失败", zap.Error(err), zap.String("path", api.Path), zap.String("role", authorityId))
		}
	}
}

// bindMenusToAuthorities 将资产中心菜单绑定到 888 与 9528（幂等）
func bindMenusToAuthorities(ctx context.Context) {
	for _, name := range assetMenus {
		authorityIds := []string{superAdminAuthorityId, normalAuthorityId}
		if name == "assetCredential" {
			authorityIds = []string{superAdminAuthorityId} // 凭据菜单仅超管可见
		}
		var menu model.SysBaseMenu
		if err := global.GVA_DB.WithContext(ctx).Where("name = ?", name).First(&menu).Error; err != nil {
			zap.L().Warn(fmt.Sprintf("asset 插件：菜单 %s 未找到，跳过角色绑定", name))
			continue
		}
		for _, authorityId := range authorityIds {
			var count int64
			global.GVA_DB.Model(&model.SysAuthorityMenu{}).
				Where("sys_base_menu_id = ? AND sys_authority_authority_id = ?", menu.ID, authorityId).
				Count(&count)
			if count > 0 {
				continue
			}
			binding := model.SysAuthorityMenu{
				MenuId:      fmt.Sprint(menu.ID),
				AuthorityId: authorityId,
			}
			if err := global.GVA_DB.Create(&binding).Error; err != nil {
				zap.L().Error("asset 插件：菜单绑定角色失败", zap.Error(err), zap.String("menu", name))
			}
		}
	}
}
