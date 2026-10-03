// Package service 终端会话审计服务
package service

import (
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/model/common/request"
	"github.com/hequan2017/new-ops/server/plugin/term/model"
)

// TermAuditService 会话审计服务
type TermAuditService struct{}

// StartSession 创建会话记录
func (s *TermAuditService) StartSession(sess *model.TermSession) error {
	return global.GVA_DB.Create(sess).Error
}

// EndSession 结束会话
func (s *TermAuditService) EndSession(id uint) {
	now := time.Now()
	global.GVA_DB.Model(&model.TermSession{}).Where("id = ?", id).
		Updates(map[string]any{"status": model.TermSessionFinished, "ended_at": &now})
}

// AppendStream 落一条流镜像（桥接侧直接调用，量可控）
func (s *TermAuditService) AppendStream(sessionID uint, seq int64, direction uint8, payload string) error {
	return global.GVA_DB.Create(&model.TermSessionStream{
		SessionID: sessionID, Seq: seq, Direction: direction, Payload: payload,
	}).Error
}

// AppendCommand 落一条命令抽取
func (s *TermAuditService) AppendCommand(sessionID uint, seq int64, command string) error {
	return global.GVA_DB.Create(&model.TermSessionCommand{
		SessionID: sessionID, Seq: seq, Command: command,
	}).Error
}

// GetSessionList 会话分页列表
func (s *TermAuditService) GetSessionList(info request.PageInfo, username, status string) (list []*model.TermSession, total int64, err error) {
	db := global.GVA_DB.Model(&model.TermSession{})
	if username != "" {
		db = db.Where("username LIKE ?", "%"+username+"%")
	}
	if status != "" {
		db = db.Where("status = ?", status)
	}
	if err = db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err = db.Scopes(info.Paginate()).Order("id DESC").Find(&list).Error
	return list, total, err
}

// GetSessionStreams 会话流镜像（回放：seq 升序，limit 上限 20000）
func (s *TermAuditService) GetSessionStreams(sessionID uint) (list []*model.TermSessionStream, err error) {
	err = global.GVA_DB.Where("session_id = ?", sessionID).Order("seq ASC").Limit(20000).Find(&list).Error
	return list, err
}

// GetSessionCommands 会话命令列表
func (s *TermAuditService) GetSessionCommands(sessionID uint) (list []*model.TermSessionCommand, err error) {
	err = global.GVA_DB.Where("session_id = ?", sessionID).Order("seq ASC").Limit(5000).Find(&list).Error
	return list, err
}

// ---------- 命令抽取纯逻辑（可单测） ----------

var ansiRe = regexp.MustCompile("\x1b\\[[0-9;?]*[a-zA-Z]|\x1b][^\x07]*\x07")

// inputAccumulator 上行输入累积器：剥离 ANSI 序列与退格，按 \r 切命令
type inputAccumulator struct {
	mu  sync.Mutex
	buf strings.Builder
}

func newInputAccumulator() *inputAccumulator { return &inputAccumulator{} }

// feed 喂入一段上行输入，返回本次抽出的命令（按出现顺序）
func (a *inputAccumulator) feed(s string) []string {
	a.mu.Lock()
	defer a.mu.Unlock()
	var commands []string
	for _, r := range s {
		switch {
		case r == '\r' || r == '\n':
			cmd := strings.TrimSpace(stripControl(a.buf.String()))
			a.buf.Reset()
			if cmd != "" {
				commands = append(commands, cmd)
			}
		case r == '\x7f' || r == '\b':
			b := []rune(a.buf.String())
			if len(b) > 0 {
				a.buf.Reset()
				a.buf.WriteString(string(b[:len(b)-1]))
			}
		default:
			a.buf.WriteRune(r)
		}
	}
	return commands
}

// flush 结束会话时取出残留未回车的缓冲
func (a *inputAccumulator) flush() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	cmd := strings.TrimSpace(stripControl(a.buf.String()))
	a.buf.Reset()
	return cmd
}

// stripControl 去除 ANSI 转义序列与其余控制字符（保留可打印字符与空格）
func stripControl(s string) string {
	s = ansiRe.ReplaceAllString(s, "")
	var b strings.Builder
	for _, r := range s {
		if r >= 32 && r != 127 {
			b.WriteRune(r)
		}
	}
	return b.String()
}
