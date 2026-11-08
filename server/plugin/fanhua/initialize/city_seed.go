// Package initialize fanhua：城市治理指标树种子（需求 3.4-17，含视觉评价维度）
package initialize

import (
	"context"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/fanhua/model"
)

// govRootSeed 根指标（含子叶）
var govRootSeed = []struct {
	Name    string
	Weight  int
	Visual  bool
	Desc    string
	Leaves  []model.FanhuaGovIndicator
}{
	{
		Name: "环境品质", Weight: 30, Desc: "空气/噪声等物理环境",
		Leaves: []model.FanhuaGovIndicator{
			{Name: "空气质量", Weight: 50, Desc: "AQI 反向归一化得分"},
			{Name: "噪声水平", Weight: 50, Desc: "噪声监测反向归一化得分"},
		},
	},
	{
		Name: "公共服务", Weight: 40, Desc: "设施可达与服务满意度",
		Leaves: []model.FanhuaGovIndicator{
			{Name: "设施可达性", Weight: 60, Desc: "POI 15 分钟生活圈覆盖率"},
			{Name: "服务满意度", Weight: 40, Desc: "社区问卷满意度"},
		},
	},
	{
		Name: "视觉评价", Weight: 30, Visual: true, Desc: "街景与建筑风貌视觉维度",
		Leaves: []model.FanhuaGovIndicator{
			{Name: "街景美观度", Weight: 40, Visual: true, Desc: "街景图像视觉评分"},
			{Name: "建筑协调性", Weight: 30, Visual: true, Desc: "风貌协调视觉评分"},
			{Name: "绿化视觉", Weight: 30, Visual: true, Desc: "绿化覆盖视觉评分"},
		},
	},
}

// SeedGovIndicators 治理指标树种子（根先插、子挂父 ID，按名幂等）
func SeedGovIndicators(ctx context.Context) {
	for _, root := range govRootSeed {
		var rootRow model.FanhuaGovIndicator
		err := global.GVA_DB.WithContext(ctx).Where("name = ? AND parent_id = 0", root.Name).First(&rootRow).Error
		if err != nil {
			rootRow = model.FanhuaGovIndicator{Name: root.Name, ParentId: 0, Weight: root.Weight, Visual: root.Visual, Desc: root.Desc}
			if err := global.GVA_DB.WithContext(ctx).Create(&rootRow).Error; err != nil {
				global.GVA_LOG.Error("fanhua 治理指标根种子失败: " + err.Error())
				continue
			}
		}
		for _, leaf := range root.Leaves {
			var exist model.FanhuaGovIndicator
			err := global.GVA_DB.WithContext(ctx).Where("name = ? AND parent_id = ?", leaf.Name, rootRow.ID).First(&exist).Error
			if err != nil {
				leaf.ParentId = rootRow.ID
				if err := global.GVA_DB.WithContext(ctx).Create(&leaf).Error; err != nil {
					global.GVA_LOG.Error("fanhua 治理指标叶种子失败: " + err.Error())
				}
			}
		}
	}
}
