// Package initialize pipeline 插件表迁移（幂等）
package initialize

import (
	"context"
	"fmt"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/pipeline/model"
	"github.com/pkg/errors"
)

// Gorm 注册流水线数据表
func Gorm(ctx context.Context) {
	err := global.GVA_DB.WithContext(ctx).AutoMigrate(
		new(model.Pipeline),
		new(model.PipelineStage),
		new(model.PipelineStep),
		new(model.PipelineBuild),
		new(model.BuildLog),
	)
	if err != nil {
		err = errors.Wrap(err, "pipeline 注册表失败!")
		global.GVA_LOG.Error(fmt.Sprintf("%+v", err))
	}
}
