// Package initialize artsci：内置工具种子（幂等，M03 起逐条追加）
package initialize

import (
	"context"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/artsci/model"
)

// SeedTools 内置工具种子（按名幂等：存在即跳过）
func SeedTools(ctx context.Context) {
	presets := []model.ArtsciTool{
		{
			Name: "Allen-Cahn 相场演化", Subject: "数学", Module: "allen-cahn", Status: "enabled",
			ParamsSchema: `{"epsilon":{"type":"number"},"gridN":{"type":"number"},"dt":{"type":"number"},"steps":{"type":"number"},"snapshotEvery":{"type":"number"}}`,
			Desc:         "一维 Allen-Cahn 方程显式有限差分求解（周期边界），返回演化帧序列与能量衰减（3.3-1）",
		},
	}
	for i := range presets {
		var exist model.ArtsciTool
		err := global.GVA_DB.WithContext(ctx).Where("name = ?", presets[i].Name).First(&exist).Error
		if err != nil {
			if err := global.GVA_DB.WithContext(ctx).Create(&presets[i]).Error; err != nil {
				global.GVA_LOG.Error("artsci 工具种子失败: " + err.Error())
			}
		}
	}
}
