// Package initialize asset 插件表迁移（幂等）
package initialize

import (
	"context"
	"fmt"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/asset/model"
	"github.com/pkg/errors"
)

// Gorm 注册资产中心数据表
func Gorm(ctx context.Context) {
	err := global.GVA_DB.WithContext(ctx).AutoMigrate(
		new(model.AssetHost),
		new(model.AssetRoom),
		new(model.AssetRack),
		new(model.AssetProductLine),
		new(model.AssetHostHistory),
		new(model.AssetGroup),
		new(model.AssetGroupHost),
		new(model.AssetGroupUser),
		new(model.CredCredential),
	)
	if err != nil {
		err = errors.Wrap(err, "asset 注册表失败!")
		global.GVA_LOG.Error(fmt.Sprintf("%+v", err))
	}
}
