// Package initialize fanhua：海派文化语音库种子（需求 3.4-15，20 条，按标题幂等）
package initialize

import (
	"context"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/fanhua/model"
)

// voiceSeed 20 条语音库条目
var voiceSeed = []model.FanhuaVoiceEntry{
	{Title: "燕燕做媒·选段", Category: "沪剧台词", Content: "沪剧《罗汉钱》燕燕做媒选段：媒人芝麻官，两边去说合。", SpeakerType: "女声", DurationSec: 96, TextSource: "沪剧《罗汉钱》"},
	{Title: "燕燕做媒·对唱", Category: "沪剧台词", Content: "燕燕做媒对唱选段：你做媒人来我搭桥，成全好事乐陶陶。", SpeakerType: "男声", DurationSec: 88, TextSource: "沪剧《罗汉钱》"},
	{Title: "雷雨·选段", Category: "沪剧台词", Content: "沪剧《雷雨》选段：家依照旧，人已两样。", SpeakerType: "男声", DurationSec: 120, TextSource: "沪剧《雷雨》"},
	{Title: "芦荡火种·选段", Category: "沪剧台词", Content: "沪剧《芦荡火种》选段：芦苇荡里摆战场，军民团结心一条。", SpeakerType: "男声", DurationSec: 110, TextSource: "沪剧《芦荡火种》"},
	{Title: "碧落黄泉·志超读信", Category: "沪剧台词", Content: "沪剧《碧落黄泉》志超读信选段：一字一句读分明，字里行间见真情。", SpeakerType: "男声", DurationSec: 132, TextSource: "沪剧《碧落黄泉》"},
	{Title: "妓女泪·杨金妹", Category: "沪剧台词", Content: "沪剧《妓女泪》选段：杨金妹哭诉身世，一字泪一行。", SpeakerType: "女声", DurationSec: 105, TextSource: "沪剧《妓女泪》"},
	{Title: "星星之火·选段", Category: "沪剧台词", Content: "沪剧《星星之火》选段：拿摩温凶得来，做工妹苦处诉勿完。", SpeakerType: "女声", DurationSec: 118, TextSource: "沪剧《星星之火》"},
	{Title: "陆雅臣·卖娘子", Category: "沪剧台词", Content: "沪剧《陆雅臣》选段：陆雅臣卖娘子，悔恨交加。", SpeakerType: "男声", DurationSec: 98, TextSource: "沪剧《陆雅臣》"},
	{Title: "笃笃笃卖糖粥", Category: "弄堂俚语", Content: "笃笃笃，卖糖粥，三斤胡桃四斤壳，吃侬肉，还侬壳。", SpeakerType: "童声", DurationSec: 30, TextSource: "上海传统童谣"},
	{Title: "落雨喽打烊喽", Category: "弄堂俚语", Content: "落雨喽，打烊喽，小八腊子开会喽。", SpeakerType: "童声", DurationSec: 22, TextSource: "上海传统童谣"},
	{Title: "一歇哭一歇笑", Category: "弄堂俚语", Content: "一歇哭，一歇笑，两只眼睛开大炮。", SpeakerType: "童声", DurationSec: 18, TextSource: "上海传统童谣"},
	{Title: "从前有个老老头", Category: "弄堂俚语", Content: "从前头，有个老老头，坐勒门口数楼楼。", SpeakerType: "童声", DurationSec: 35, TextSource: "上海传统童谣"},
	{Title: "山上有只老虎", Category: "弄堂俚语", Content: "山浪有只老虎，老虎要吃人，拿伊关勒笼子里。", SpeakerType: "童声", DurationSec: 28, TextSource: "上海传统童谣"},
	{Title: "炒黄豆炒绿豆", Category: "弄堂俚语", Content: "炒炒炒黄豆，炒好黄豆翻跟斗。", SpeakerType: "童声", DurationSec: 24, TextSource: "上海传统童谣"},
	{Title: "弄堂口老虎灶", Category: "弄堂俚语", Content: "弄堂口老虎灶，天没亮就冒热气，泡开水一角钿。", SpeakerType: "男声", DurationSec: 42, TextSource: "弄堂口述"},
	{Title: "再别康桥·节选", Category: "经典朗读", Content: "轻轻的我走了，正如我轻轻的来；我轻轻的招手，作别西天的云彩。", SpeakerType: "男声", DurationSec: 150, TextSource: "徐志摩《再别康桥》"},
	{Title: "再别康桥·尾声", Category: "经典朗读", Content: "悄悄的我走了，正如我悄悄的来；我挥一挥衣袖，不带走一片云彩。", SpeakerType: "女声", DurationSec: 140, TextSource: "徐志摩《再别康桥》"},
	{Title: "匆匆·节选", Category: "经典朗读", Content: "燕子去了，有再来的时候；杨柳枯了，有再青的时候。", SpeakerType: "女声", DurationSec: 160, TextSource: "朱自清《匆匆》"},
	{Title: "背影·节选", Category: "经典朗读", Content: "我看见他戴着黑布小帽，穿着黑布大马褂，蹒跚地走到铁道边。", SpeakerType: "男声", DurationSec: 175, TextSource: "朱自清《背影》"},
	{Title: "弄堂一日·晨景", Category: "经典朗读", Content: "弄堂的早晨从倒马桶声音里醒来，生煎馒头香气飘过亭子间窗口。", SpeakerType: "童声", DurationSec: 88, TextSource: "海派散文示例"},
}

// SeedVoiceEntries 海派文化语音库种子（按标题幂等）
func SeedVoiceEntries(ctx context.Context) {
	for i := range voiceSeed {
		var exist model.FanhuaVoiceEntry
		err := global.GVA_DB.WithContext(ctx).Where("title = ?", voiceSeed[i].Title).First(&exist).Error
		if err != nil {
			if err := global.GVA_DB.WithContext(ctx).Create(&voiceSeed[i]).Error; err != nil {
				global.GVA_LOG.Error("fanhua 语音库种子失败: " + err.Error())
			}
		}
	}
}
