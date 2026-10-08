// Package service 白泽数据库工单：goInception 审核/执行/备份（M6 收官）
// 协议（真机实测，138 dbops-inception v1.3.0）：以 MySQL 协议直连引擎（无需用户名）；目标库与动作
// 经 magic 块首行注释选项指定——审核 /*--user=..;--password=..;--host=..;--port=..;--check=1;*/，
// 执行换 --execute=1（--backup=true 生成回滚，--ignore-warnings=1 放行警告）；
// 定界为裸语句 inception_magic_start/commit；整块经 multiStatements=true 单包提交；
// 结果集 errlevel 0 通过/1 警告/2 错误。注意 mysql CLI 会按分号拆包不适用，须 go 驱动/pymysql 单包。
// 执行后回滚语句：备份库 `<目标host非字母数字转下划线>_<端口>_<库名>`（以结果集 backup_dbname 为准），
// 信息表 $_$Inception_backup_information$_$ 按 opid_time 查得表名，再到同名回滚表取 rollback_statement。
// 块提交统一走 queryEngineBlock（inception_engine.go，载荷即送审对象非注入面）。
package service

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	gormMysql "gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/asset/crypto"
	"github.com/hequan2017/new-ops/server/plugin/dbops/model"
)

// 错误码接续 1701-1708
const (
	ErrCodeInceptionConnFailed = 1709
	ErrCodeExecuteFailed       = 1710
)

// magic 定界裸语句（goInception 协议常量；不得加注释包裹）
const (
	magicStmtStart = "inception_magic_start;"
	magicStmtEnd   = "inception_magic_commit;"
)

// sqlIdentifierReg 标识符白名单：库名仅字母数字下划线（进 use/备份库拼接前强校验）
var sqlIdentifierReg = regexp.MustCompile(`^[A-Za-z0-9_]{1,64}$`)

// auditRow 审核单行结果（对齐 goInception 结果集列）
type auditRow struct {
	OrderID      string `json:"orderid,omitempty"`
	Stage        string `json:"stage,omitempty"`
	ErrLevel     int    `json:"errlevel"`
	ErrMsg       string `json:"errormessage,omitempty"`
	SQL          string `json:"sqlstatement,omitempty"`
	AffectedRows string `json:"affected_rows,omitempty"`
	Sequence     string `json:"sequence,omitempty"`
	BackupDbname string `json:"backup_dbname,omitempty"`
	ExecuteTime  string `json:"execute_time,omitempty"`
	SqlSha1      string `json:"sqlsha1,omitempty"`
}

// GetInceptionConfig 读审核引擎配置（密码不回显）
func (s *DbopsService) GetInceptionConfig() (*model.DbopsInceptionConfig, error) {
	var cfg model.DbopsInceptionConfig
	err := global.GVA_DB.Where("name = ?", "default").First(&cfg).Error
	if err == gorm.ErrRecordNotFound {
		return &model.DbopsInceptionConfig{Name: "default", Port: 4000, BackupPort: 3306}, nil
	}
	if err != nil {
		return nil, err
	}
	cfg.BackupPassEnc = ""
	return &cfg, nil
}

// SaveInceptionConfig 保存审核引擎配置（密码留空不改）
func (s *DbopsService) SaveInceptionConfig(cfg *model.DbopsInceptionConfig) error {
	if cfg == nil {
		return newDbErr(ErrCodeOrderInvalid, "配置不能为空")
	}
	cfg.Name = "default"
	if cfg.Port <= 0 {
		cfg.Port = 4000
	}
	if cfg.BackupPort <= 0 {
		cfg.BackupPort = 3306
	}
	var count int64
	global.GVA_DB.Model(&model.DbopsInceptionConfig{}).Where("name = ?", "default").Count(&count)
	if count == 0 {
		if cfg.BackupPassEnc != "" {
			enc, err := crypto.Encrypt(cfg.BackupPassEnc)
			if err != nil {
				return fmt.Errorf("密码加密失败: %w", err)
			}
			cfg.BackupPassEnc = enc
		}
		return global.GVA_DB.Create(cfg).Error
	}
	updates := map[string]any{
		"host": cfg.Host, "port": cfg.Port,
		"backup_host": cfg.BackupHost, "backup_port": cfg.BackupPort, "backup_user": cfg.BackupUser,
		"notes": cfg.Notes,
	}
	if cfg.BackupPassEnc != "" {
		enc, err := crypto.Encrypt(cfg.BackupPassEnc)
		if err != nil {
			return fmt.Errorf("密码加密失败: %w", err)
		}
		updates["backup_pass_enc"] = enc
	}
	return global.GVA_DB.Model(&model.DbopsInceptionConfig{}).Where("name = ?", "default").Updates(updates).Error
}

// inceptionTarget 目标库连接信息（进 magic 首行注释选项）
type inceptionTarget struct {
	InstanceHost string
	InstancePort int
	Username     string
	Password     string
	Database     string
}

// engineDSN 组引擎 DSN：goInception 默认无鉴权，整块 multiStatements 单包提交
func engineDSN(cfg *model.DbopsInceptionConfig) string {
	return fmt.Sprintf("tcp(%s:%d)/?timeout=10s&readTimeout=120s&writeTimeout=10s&multiStatements=true&charset=utf8mb4",
		cfg.Host, cfg.Port)
}

// magicOptionSafe 校验凭据可安全嵌入注释选项（; */ 换行会截断/逃逸选项注释）
func magicOptionSafe(s string) bool {
	return !strings.ContainsAny(s, ";*/\r\n")
}

// wrapMagicBlock 按 goInception 协议组装 magic 块：首行注释选项指定目标库与动作（check=1 审核 /
// execute=1 执行+备份），定界为裸语句；工单未自带 use 时前置 use 目标库（库名走白名单校验）。
func wrapMagicBlock(t inceptionTarget, body string, execute bool) (string, error) {
	if !magicOptionSafe(t.Username) || !magicOptionSafe(t.Password) || !magicOptionSafe(t.InstanceHost) {
		return "", newDbErr(ErrCodeOrderInvalid, "目标库凭据含注释选项非法字符（; */ 换行）")
	}
	if !sqlIdentifierReg.MatchString(t.Database) {
		return "", newDbErr(ErrCodeOrderInvalid, "实例默认库名非法（仅字母数字下划线）: "+t.Database)
	}
	var buf bytes.Buffer
	buf.WriteString("/*--user=")
	buf.WriteString(t.Username)
	buf.WriteString(";--password=")
	buf.WriteString(t.Password)
	buf.WriteString(";--host=")
	buf.WriteString(t.InstanceHost)
	buf.WriteString(";--port=")
	buf.WriteString(strconv.Itoa(t.InstancePort))
	if execute {
		// ignore-warnings：警告已在审核阶段人工过目（errlevel<2 才会到执行），
		// 引擎执行态默认遇警告即停（如实测 'name' 关键字警告拦住整个块），故放行
		buf.WriteString(";--execute=1;--backup=true;--ignore-warnings=1")
	} else {
		buf.WriteString(";--check=1")
	}
	buf.WriteString(";*/\n")
	buf.WriteString(magicStmtStart)
	buf.WriteByte('\n')
	if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(body)), "use ") {
		buf.WriteString("use ")
		buf.WriteString(t.Database)
		buf.WriteString(";\n")
	}
	buf.WriteString(strings.TrimSpace(body))
	buf.WriteByte('\n')
	buf.WriteString(magicStmtEnd)
	return buf.String(), nil
}

// readAllResultsets 读取全部结果集（goInception 结果为单个汇总结果集，兼容多结果集读取）
func readAllResultsets(rows *sql.Rows) ([]auditRow, error) {
	out := []auditRow{}
	for {
		cols, err := rows.Columns()
		if err != nil {
			return nil, err
		}
		idx := map[string]int{}
		for i, c := range cols {
			idx[strings.ToLower(c)] = i
		}
		get := func(rec []sql.RawBytes, names ...string) string {
			for _, n := range names {
				if i, ok := idx[n]; ok {
					return string(rec[i])
				}
			}
			return ""
		}
		rec := make([]sql.RawBytes, len(cols))
		scan := make([]any, len(cols))
		for i := range rec {
			scan[i] = &rec[i]
		}
		for rows.Next() {
			if err := rows.Scan(scan...); err != nil {
				return nil, err
			}
			// goInception 实测列名 order_id/error_level/error_message/sql；
			// 兼容旧 Inception 口径 orderid/errlevel/errormessage/sqlstatement
			row := auditRow{
				OrderID: get(rec, "order_id", "orderid"), Stage: get(rec, "stage"),
				SQL: get(rec, "sql", "sqlstatement"), AffectedRows: get(rec, "affected_rows"),
				Sequence: get(rec, "sequence"), BackupDbname: get(rec, "backup_dbname"),
				ExecuteTime: get(rec, "execute_time"), SqlSha1: get(rec, "sqlsha1"),
			}
			row.ErrLevel, _ = strconv.Atoi(get(rec, "error_level", "errlevel"))
			row.ErrMsg = get(rec, "error_message", "errormessage")
			out = append(out, row)
		}
		if !rows.NextResultSet() {
			break
		}
	}
	return out, nil
}

// runInceptionBlock 连引擎提交一个 magic 块（execute=true 时执行并生成备份）
func (s *DbopsService) runInceptionBlock(order *model.DbopsOrder, execute bool) ([]auditRow, error) {
	var cfg model.DbopsInceptionConfig
	if err := global.GVA_DB.Where("name = ?", "default").First(&cfg).Error; err != nil {
		return nil, newDbErr(ErrCodeAuditNotConfigured, "SQL 审核引擎（goInception）未配置")
	}
	host, port, user, pass, err := s.GetInstanceCredential(order.InstanceID)
	if err != nil {
		return nil, err
	}
	var inst model.DbopsInstance
	global.GVA_DB.First(&inst, order.InstanceID)
	block, berr := wrapMagicBlock(inceptionTarget{
		InstanceHost: host, InstancePort: port, Username: user, Password: pass, Database: inst.Database,
	}, order.AuditPayload, execute)
	if berr != nil {
		return nil, berr
	}
	db, err := sql.Open("mysql", engineDSN(&cfg))
	if err != nil {
		return nil, newDbErr(ErrCodeInceptionConnFailed, "审核引擎 DSN 无效: "+err.Error())
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		return nil, newDbErr(ErrCodeInceptionConnFailed, "审核引擎连接失败: "+err.Error())
	}
	rows, qerr := queryEngineBlock(db, block)
	if qerr != nil {
		// 引擎对整体失败以 MySQL 错误回传（如连接目标库失败/语法 fatal）
		return []auditRow{{ErrLevel: 2, ErrMsg: qerr.Error(), Stage: "CHECKED"}}, nil
	}
	defer rows.Close()
	return readAllResultsets(rows)
}

// AuditOrder 审核（errlevel=2 → 审核不通过保持待审核；全部通过 → 审核通过）
func (s *DbopsService) AuditOrder(orderID uint, operator string) (*model.DbopsOrder, error) {
	var order model.DbopsOrder
	if err := global.GVA_DB.First(&order, orderID).Error; err != nil {
		return nil, newDbErr(ErrCodeOrderNotFound, "工单不存在")
	}
	if order.Status != model.OrderPending {
		return nil, newDbErr(ErrCodeOrderInvalid, "工单不在待审核状态")
	}
	rows, err := s.runInceptionBlock(&order, false)
	if err != nil {
		return nil, err
	}
	resultBytes := marshalRows(rows)
	hasError := false
	for _, r := range rows {
		if r.ErrLevel >= 2 {
			hasError = true
			break
		}
	}
	status := model.OrderApproved
	if hasError {
		status = model.OrderPending
	}
	if err := global.GVA_DB.Model(&model.DbopsOrder{}).Where("id = ?", orderID).Updates(map[string]any{
		"status": status, "audit_result": resultBytes,
	}).Error; err != nil {
		return nil, err
	}
	global.GVA_DB.First(&order, orderID)
	return &order, nil
}

// ExecuteOrder 执行已审核工单（结果含备份回滚语句查询）
func (s *DbopsService) ExecuteOrder(orderID uint, operator string) (*model.DbopsOrder, error) {
	var order model.DbopsOrder
	if err := global.GVA_DB.First(&order, orderID).Error; err != nil {
		return nil, newDbErr(ErrCodeOrderNotFound, "工单不存在")
	}
	if order.Status != model.OrderApproved {
		return nil, newDbErr(ErrCodeOrderInvalid, "工单不在审核通过状态，无法执行")
	}
	global.GVA_DB.Model(&order).Update("status", model.OrderExecuting)
	rows, err := s.runInceptionBlock(&order, true)
	if err != nil {
		global.GVA_DB.Model(&order).Update("status", model.OrderApproved)
		return nil, err
	}
	hasError := false
	for _, r := range rows {
		if r.ErrLevel >= 2 {
			hasError = true
			break
		}
	}
	rollbacks := s.fetchRollbacks(&order, rows)
	status := model.OrderSuccess
	if hasError {
		status = model.OrderFailed
	}
	now := time.Now()
	execPayload := map[string]any{
		"rows": rows, "operator": operator, "finishedAt": now.Format(time.RFC3339),
		"rollbacks": rollbacks,
	}
	global.GVA_DB.Model(&model.DbopsOrder{}).Where("id = ?", orderID).Updates(map[string]any{
		"status": status, "exec_result": marshalRows(execPayload), "finished_at": &now,
	})
	global.GVA_DB.First(&order, orderID)
	return &order, nil
}

// fetchRollbacks 从备份库按 opid_time 取回滚语句（GORM builder；备份未配置/查询失败不阻断执行结果）。
// 库名优先取结果集 backup_dbname（引擎权威口径），兜底按 <host点转下划线>_<端口>_<库名> 推导；
// 信息表 $_$Inception_backup_information$_$（opid_time→tablename），回滚语句在同名回滚表。
func (s *DbopsService) fetchRollbacks(order *model.DbopsOrder, rows []auditRow) []map[string]string {
	var cfg model.DbopsInceptionConfig
	if err := global.GVA_DB.Where("name = ?", "default").First(&cfg).Error; err != nil || cfg.BackupHost == "" {
		return nil
	}
	pass := ""
	if cfg.BackupPassEnc != "" {
		p, err := crypto.Decrypt(cfg.BackupPassEnc)
		if err != nil {
			return nil
		}
		pass = p
	}
	var inst model.DbopsInstance
	global.GVA_DB.First(&inst, order.InstanceID)
	dbName := inst.Database
	if !sqlIdentifierReg.MatchString(dbName) {
		return nil
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/?timeout=5s&readTimeout=10s", cfg.BackupUser, pass, cfg.BackupHost, cfg.BackupPort)
	backup, err := gorm.Open(gormMysql.Open(dsn))
	if err != nil {
		return nil
	}
	sqlDB, gerr := backup.DB()
	if gerr == nil {
		sqlDB.SetConnMaxLifetime(time.Minute)
	}
	out := []map[string]string{}
	type rollbackRec struct {
		RollbackStatement string
	}
	for _, r := range rows {
		if r.Sequence == "" {
			continue
		}
		schema := r.BackupDbname
		if schema == "" {
			// 引擎命名口径：host 中非字母数字字符（.、- 等）一律转下划线，拼 _端口_库名
			schema = strings.Map(func(r rune) rune {
				if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_' {
					return r
				}
				return '_'
			}, inst.Host) + "_" + strconv.Itoa(inst.Port) + "_" + dbName
		}
		var tablename string
		if err := backup.Table(schema + ".$__$Inception_backup_information$_$").Select("tablename").
			Where("opid_time = ?", r.Sequence).Scan(&tablename).Error; err != nil || tablename == "" {
			continue
		}
		var rec rollbackRec
		if err := backup.Table(schema+"."+tablename).Select("rollback_statement").
			Where("opid_time = ?", r.Sequence).Scan(&rec).Error; err == nil && rec.RollbackStatement != "" {
			out = append(out, map[string]string{"sequence": r.Sequence, "rollback": rec.RollbackStatement})
		}
	}
	return out
}

// TestInception 测试审核引擎连通（TCP 探活；真实审核靠目标库凭据，逐工单验证）
func (s *DbopsService) TestInception() (bool, string, error) {
	var cfg model.DbopsInceptionConfig
	if err := global.GVA_DB.Where("name = ?", "default").First(&cfg).Error; err != nil {
		return false, "", newDbErr(ErrCodeAuditNotConfigured, "SQL 审核引擎（goInception）未配置")
	}
	if cfg.Host == "" {
		return false, "", newDbErr(ErrCodeInceptionConnFailed, "引擎主机未配置")
	}
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port)), 3*time.Second)
	if err != nil {
		return false, "", newDbErr(ErrCodeInceptionConnFailed, "引擎不可达: "+err.Error())
	}
	_ = conn.Close()
	return true, "goInception 端口可达（真实审核需目标库凭据可用）", nil
}

func marshalRows(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "[]"
	}
	return string(b)
}
