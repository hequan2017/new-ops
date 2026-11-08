// Package model fanhua 繁花-文科大模型：沪普英问答日志（需求 3.4-20）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// FanhuaDialectQaLog 方言问答日志（响应时延/阈值/助老兼容留痕）
type FanhuaDialectQaLog struct {
	global.GVA_MODEL
	Question     string `json:"question" gorm:"type:varchar(500);comment:提问"`
	Answer       string `json:"answer" gorm:"type:mediumtext;comment:回答"`
	Engine       string `json:"engine" gorm:"type:varchar(16);default:rule;comment:llm/rule"`
	ModelId      uint   `json:"modelId" gorm:"default:0;comment:模型ID(0=规则口径)"`
	LatencyMs    int64  `json:"latencyMs" gorm:"comment:时延(毫秒)"`
	OverThreshold bool  `json:"overThreshold" gorm:"comment:超 1.2s 阈值"`
	WakeCompat   bool   `json:"wakeCompat" gorm:"comment:助老语音唤醒兼容口径"`
	CorpusHits   int    `json:"corpusHits" gorm:"comment:注入平行语料条数"`
	Ok           bool   `json:"ok" gorm:"index;comment:是否成功"`
	ErrMsg       string `json:"errMsg" gorm:"type:varchar(500);comment:错误"`
}

func (FanhuaDialectQaLog) TableName() string { return "fanhua_dialect_qa_log" }
