// Package initialize term 插件表迁移（幂等）
package initialize

import (
	"context"
	"fmt"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/term/model"
	"github.com/pkg/errors"
)

// Gorm 注册终端会话审计数据表
func Gorm(ctx context.Context) {
	err := global.GVA_DB.WithContext(ctx).AutoMigrate(
		new(model.TermSession),
		new(model.TermSessionStream),
		new(model.TermSessionCommand),
	)
	if err != nil {
		err = errors.Wrap(err, "term 注册表失败!")
		global.GVA_LOG.Error(fmt.Sprintf("%+v", err))
	}
}
