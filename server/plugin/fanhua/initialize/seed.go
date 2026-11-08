// Package initialize fanhua：内置种子（幂等，N03 起逐条追加）
package initialize

import (
	"context"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/fanhua/model"
)

// parallelSeed 沪-普-英三元组种子（20 条，来源：沪普对照示例）
var parallelSeed = []model.FanhuaParallel{
	{Sh: "侬好", Pu: "你好", En: "Hello", Source: "沪普对照示例"},
	{Sh: "饭吃过了伐", Pu: "吃饭了吗", En: "Have you eaten?", Source: "沪普对照示例"},
	{Sh: "今朝天气交关好", Pu: "今天天气非常好", En: "The weather is very nice today.", Source: "沪普对照示例"},
	{Sh: "明朝见", Pu: "明天见", En: "See you tomorrow.", Source: "沪普对照示例"},
	{Sh: "阿拉一道去白相", Pu: "我们一起去玩", En: "Let's hang out together.", Source: "沪普对照示例"},
	{Sh: "小囡困觉了", Pu: "小孩睡觉了", En: "The kid is asleep.", Source: "沪普对照示例"},
	{Sh: "迭桩事体交关要紧", Pu: "这件事情非常重要", En: "This matter is very important.", Source: "沪普对照示例"},
	{Sh: "侬晓得伐", Pu: "你知道吗", En: "Do you know?", Source: "沪普对照示例"},
	{Sh: "呒没解决问题", Pu: "没有解决问题", En: "The problem wasn't solved.", Source: "沪普对照示例"},
	{Sh: "迭样式老灵额", Pu: "这样子很棒", En: "This looks great.", Source: "沪普对照示例"},
	{Sh: "谢谢侬", Pu: "谢谢你", En: "Thank you.", Source: "沪普对照示例"},
	{Sh: "再会", Pu: "再见", En: "Goodbye.", Source: "沪普对照示例"},
	{Sh: "覅急", Pu: "不要着急", En: "Don't rush.", Source: "沪普对照示例"},
	{Sh: "身体适意伐", Pu: "身体舒服吗", En: "Are you feeling well?", Source: "沪普对照示例"},
	{Sh: "汏手间勒辣啥地方", Pu: "洗手间在哪里", En: "Where is the restroom?", Source: "沪普对照示例"},
	{Sh: "价钿忒贵了", Pu: "价格太贵了", En: "The price is too high.", Source: "沪普对照示例"},
	{Sh: "慢慢交来", Pu: "慢慢来", En: "Take your time.", Source: "沪普对照示例"},
	{Sh: "一心一意帮侬讲", Pu: "一心一意跟你说", En: "I'm telling you wholeheartedly.", Source: "沪普对照示例"},
	{Sh: "弄堂里向老闹猛", Pu: "弄堂里面很热闹", En: "The alley is bustling.", Source: "沪普对照示例"},
	{Sh: "夜到吃啥物事", Pu: "晚上吃什么", En: "What's for dinner tonight?", Source: "沪普对照示例"},
}

// SeedParallels 沪-普-英平行语料种子（按沪语句幂等）
func SeedParallels(ctx context.Context) {
	for i := range parallelSeed {
		var exist model.FanhuaParallel
		err := global.GVA_DB.WithContext(ctx).Where("sh = ?", parallelSeed[i].Sh).First(&exist).Error
		if err != nil {
			if err := global.GVA_DB.WithContext(ctx).Create(&parallelSeed[i]).Error; err != nil {
				global.GVA_LOG.Error("fanhua 平行语料种子失败: " + err.Error())
			}
		}
	}
}
