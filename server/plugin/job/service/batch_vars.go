package service

import (
	"github.com/hequan2017/new-ops/server/plugin/job/model"
)

// variableGroupsByAssetGroupID 把变量组按关联资产组建索引（同组多个时取 ID 最大即最新）
func variableGroupsByAssetGroupID(vgs []model.JobVariableGroup) map[uint]model.JobVariableGroup {
	m := make(map[uint]model.JobVariableGroup)
	for _, vg := range vgs {
		if vg.AssetGroupID == nil || *vg.AssetGroupID == 0 {
			continue
		}
		if old, ok := m[*vg.AssetGroupID]; !ok || vg.ID > old.ID {
			m[*vg.AssetGroupID] = vg
		}
	}
	return m
}
