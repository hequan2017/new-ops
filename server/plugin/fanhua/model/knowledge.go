// Package model fanhua 繁花-文科大模型：海派文化知识库（需求 3.4-15 语音库/三语平行）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// FanhuaVoiceEntry 海派文化语音库条目
type FanhuaVoiceEntry struct {
	global.GVA_MODEL
	Title       string `json:"title" gorm:"type:varchar(200);comment:条目标题"`
	Category    string `json:"category" gorm:"type:varchar(16);index;comment:沪剧台词/弄堂俚语/经典朗读"`
	Content     string `json:"content" gorm:"type:text;comment:文本内容"`
	SpeakerType string `json:"speakerType" gorm:"type:varchar(8);index;comment:男声/女声/童声"`
	DurationSec int    `json:"durationSec" gorm:"comment:音频时长(秒)"`
	MediaPath   string `json:"mediaPath" gorm:"type:varchar(500);comment:音频文件路径(占位)"`
	TextSource  string `json:"textSource" gorm:"type:varchar(200);comment:文本出处"`
}

func (FanhuaVoiceEntry) TableName() string { return "fanhua_voice_entry" }
