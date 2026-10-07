// Package initialize 白泽 Kubernetes 插件初始化
package initialize

import (
	"context"
	"fmt"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/k8s/model"
	"github.com/pkg/errors"
)

// Gorm 注册 k8s 数据表（幂等）
func Gorm(ctx context.Context) {
	err := global.GVA_DB.WithContext(ctx).AutoMigrate(
		new(model.K8sCluster),
		new(model.K8sHelmRepo),
	)
	if err != nil {
		err = errors.Wrap(err, "k8s 注册表失败!")
		global.GVA_LOG.Error(fmt.Sprintf("%+v", err))
	}
}
