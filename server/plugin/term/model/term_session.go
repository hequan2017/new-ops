// Package model 终端会话审计数据模型
package model

import (
	"time"

	"github.com/hequan2017/new-ops/server/global"
)

// 会话状态
const (
	TermSessionActive   = "进行中"
	TermSessionFinished = "已结束"
)

// 流方向
const (
	StreamDirUp   = 0 // 上行（用户输入）
	StreamDirDown = 1 // 下行（远端输出）
)

// TermSession 终端会话（元数据）
type TermSession struct {
	global.GVA_MODEL
	HostID       uint       `json:"hostId" gorm:"comment:主机ID;index"`
	Hostname     string     `json:"hostname" gorm:"comment:主机名快照"`
	IP           string     `json:"ip" gorm:"comment:目标IP"`
	UserID       uint       `json:"userId" gorm:"comment:操作用户ID;index"`
	Username     string     `json:"username" gorm:"comment:操作用户"`
	CredentialID uint       `json:"credentialId" gorm:"comment:使用凭据ID"`
	ClientIP     string     `json:"clientIp" gorm:"comment:客户端来源IP"`
	Cols         int        `json:"cols"`
	Rows         int        `json:"rows"`
	Status       string     `json:"status" gorm:"comment:状态;default:进行中"`
	StartedAt    time.Time  `json:"startedAt"`
	EndedAt      *time.Time `json:"endedAt"`
	Fingerprint  string     `json:"fingerprint" gorm:"comment:主机公钥指纹(SHA256,TOFU)"`
}

// TermSessionStream 终端流镜像（全量录像数据，按 seq 回放）
type TermSessionStream struct {
	global.GVA_MODEL
	SessionID uint   `json:"sessionId" gorm:"comment:会话ID;index:idx_term_stream_sid_seq,priority:1"`
	Seq       int64  `json:"seq" gorm:"comment:会话内序号;index:idx_term_stream_sid_seq,priority:2"`
	Direction uint8  `json:"direction" gorm:"comment:0上行/1下行"`
	Payload   string `json:"payload" gorm:"type:longtext;comment:数据内容"`
}

// TermSessionCommand 命令抽取结果（上行输入按行归一化）
type TermSessionCommand struct {
	global.GVA_MODEL
	SessionID uint   `json:"sessionId" gorm:"comment:会话ID;index"`
	Seq       int64  `json:"seq" gorm:"comment:触发序号"`
	Command   string `json:"command" gorm:"type:text;comment:命令"`
}
