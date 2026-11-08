// Package initialize fanhua：海派文化实体种子（需求 3.4-10，50 词条，按词条幂等）
package initialize

import (
	"context"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/fanhua/model"
)

// entitySeed 50 词条（地域 15 / 文化符号 15 / 地标 20）
var entitySeed = []model.FanhuaEntity{
	// 地域 15
	{Category: "地域", Term: "黄浦区"}, {Category: "地域", Term: "静安区"},
	{Category: "地域", Term: "徐汇区"}, {Category: "地域", Term: "长宁区"},
	{Category: "地域", Term: "普陀区"}, {Category: "地域", Term: "虹口区"},
	{Category: "地域", Term: "杨浦区"}, {Category: "地域", Term: "闵行区"},
	{Category: "地域", Term: "宝山区"}, {Category: "地域", Term: "嘉定区"},
	{Category: "地域", Term: "浦东新区", AliasJSON: `["浦东"]`},
	{Category: "地域", Term: "松江区", AliasJSON: `["松江"]`},
	{Category: "地域", Term: "青浦区", AliasJSON: `["青浦"]`},
	{Category: "地域", Term: "奉贤区", AliasJSON: `["奉贤"]`},
	{Category: "地域", Term: "崇明区", AliasJSON: `["崇明"]`},
	// 文化符号 15
	{Category: "文化符号", Term: "滑稽戏", Desc: "上海地方喜剧戏曲形式"},
	{Category: "文化符号", Term: "沪剧", Desc: "上海地方戏曲剧种"},
	{Category: "文化符号", Term: "评弹", Desc: "苏州评话与弹词的总称，沪上书场常见"},
	{Category: "文化符号", Term: "弄堂", AliasJSON: `["里弄"]`, Desc: "上海特色居民区巷道"},
	{Category: "文化符号", Term: "石库门", AliasJSON: `["石库门房子"]`, Desc: "上海近代特色民居"},
	{Category: "文化符号", Term: "亭子间", Desc: "石库门顶层小间"},
	{Category: "文化符号", Term: "老虎窗", Desc: "屋顶凸出天窗"},
	{Category: "文化符号", Term: "旗袍", Desc: "近代上海女性代表性服饰"},
	{Category: "文化符号", Term: "本帮菜", Desc: "上海本地风味菜系"},
	{Category: "文化符号", Term: "生煎", AliasJSON: `["生煎馒头"]`, Desc: "上海经典点心"},
	{Category: "文化符号", Term: "小笼包", AliasJSON: `["小笼"]`, Desc: "南翔小笼为代表"},
	{Category: "文化符号", Term: "大饼油条", Desc: "传统早点组合"},
	{Category: "文化符号", Term: "老虎灶", Desc: "旧时熟水店"},
	{Category: "文化符号", Term: "月份牌", Desc: "老上海商业广告画"},
	{Category: "文化符号", Term: "沪语", AliasJSON: `["上海闲话", "上海话"]`, Desc: "上海方言"},
	// 地标 20
	{Category: "地标", Term: "外滩"}, {Category: "地标", Term: "豫园"},
	{Category: "地标", Term: "城隍庙"}, {Category: "地标", Term: "南京路", AliasJSON: `["南京东路", "南京西路"]`},
	{Category: "地标", Term: "淮海路", AliasJSON: `["淮海中路"]`}, {Category: "地标", Term: "徐家汇"},
	{Category: "地标", Term: "陆家嘴"}, {Category: "地标", Term: "朱家角"},
	{Category: "地标", Term: "七宝老街", AliasJSON: `["七宝"]`}, {Category: "地标", Term: "新天地"},
	{Category: "地标", Term: "田子坊"}, {Category: "地标", Term: "武康路"},
	{Category: "地标", Term: "外白渡桥"}, {Category: "地标", Term: "苏州河"},
	{Category: "地标", Term: "黄浦江"}, {Category: "地标", Term: "世纪公园"},
	{Category: "地标", Term: "静安寺"}, {Category: "地标", Term: "龙华寺"},
	{Category: "地标", Term: "玉佛寺"}, {Category: "地标", Term: "大世界"},
}

// SeedEntities 海派文化实体种子（按词条幂等）
func SeedEntities(ctx context.Context) {
	for i := range entitySeed {
		var exist model.FanhuaEntity
		err := global.GVA_DB.WithContext(ctx).Where("term = ?", entitySeed[i].Term).First(&exist).Error
		if err != nil {
			if err := global.GVA_DB.WithContext(ctx).Create(&entitySeed[i]).Error; err != nil {
				global.GVA_LOG.Error("fanhua 实体种子失败: " + err.Error())
			}
		}
	}
}
