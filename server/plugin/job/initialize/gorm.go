// Package initialize job 插件表迁移（幂等）
package initialize

import (
	"context"
	"fmt"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/job/model"
	"github.com/pkg/errors"
)

// Gorm 注册批量作业数据表
func Gorm(ctx context.Context) {
	err := global.GVA_DB.WithContext(ctx).AutoMigrate(
		new(model.JobExecRecord),
		new(model.JobExecResult),
	)
	if err != nil {
		err = errors.Wrap(err, "job 注册表失败!")
		global.GVA_LOG.Error(fmt.Sprintf("%+v", err))
	}
}
