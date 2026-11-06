// Package service llmops：结构化数据源接入（MySQL/PostgreSQL 连接测试、库表清单、采样预览）
// 连接口径：database/sql 直连（go-sql-driver/mysql 直接依赖，lib/pq 已在 require 列表）；
// 密码即用即弃——AES-GCM 解密后仅拼 DSN，不落日志；表名白名单 ^[A-Za-z0-9_]+$ 防注入。
package service

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"

	"gorm.io/gorm"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/asset/crypto"
	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
)

// 错误码（llmops 段续用）
const (
	ErrCodeDsNotFound  = 2211
	ErrCodeDsTypeBad   = 2212
	ErrCodeDsConnFail  = 2213
	ErrCodeDsNameDup   = 2214
	ErrCodeDsTableBad  = 2215
	dsConnTimeout            = 5 * time.Second
	dsPreviewMaxRows         = 100
)

var reTableName = regexp.MustCompile(`^[A-Za-z0-9_]{1,64}$`)

// DataSourceIn 数据源保存/直连测试入参（password 明文仅入参态，不入模型不回显）
type DataSourceIn struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
	User     string `json:"user"`
	Password string `json:"password"`
	Notes    string `json:"notes"`
}

// ValidSourceType 数据源类型白名单
func ValidSourceType(t string) bool {
	return t == "mysql" || t == "postgres"
}

// ValidTableName 表名白名单（防 SQL 注入，预览/清单共用）
func ValidTableName(name string) bool {
	return reTableName.MatchString(name)
}

// BuildDSN 拼接连接串（纯函数，测试锚点；密码明文仅存在于返回值）
func BuildDSN(in *DataSourceIn) string {
	if in.Type == "postgres" {
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable connect_timeout=5",
			in.Host, in.Port, in.User, in.Password, in.Database)
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true&loc=Local&timeout=5s",
		in.User, in.Password, in.Host, in.Port, in.Database)
}

// openDS 打开数据源连接（调用方负责 Close）
func openDS(in *DataSourceIn) (*sql.DB, error) {
	if !ValidSourceType(in.Type) {
		return nil, fmt.Errorf("[%d]不支持的数据源类型: %s", ErrCodeDsTypeBad, in.Type)
	}
	db, err := sql.Open(driverName(in.Type), BuildDSN(in))
	if err != nil {
		return nil, fmt.Errorf("[%d]连接失败: %w", ErrCodeDsConnFail, err)
	}
	return db, nil
}

func driverName(t string) string {
	if t == "postgres" {
		return "postgres"
	}
	return "mysql"
}

// SaveDataSource 新建/更新数据源（密码留空表示不改；名称唯一校验）
func (s *LlmopsServiceStruct) SaveDataSource(in *DataSourceIn) (*model.LlmopsDataSource, error) {
	if in == nil || strings.TrimSpace(in.Name) == "" || !ValidSourceType(in.Type) {
		return nil, fmt.Errorf("[%d]名称必填且类型须为 mysql/postgres", ErrCodeDsTypeBad)
	}
	if in.Host == "" || in.Port <= 0 || in.Port > 65535 {
		return nil, fmt.Errorf("[%d]主机/端口非法", ErrCodeDsTypeBad)
	}
	var count int64
	q := global.GVA_DB.Model(&model.LlmopsDataSource{}).Where("name = ?", in.Name)
	if in.ID > 0 {
		q = q.Where("id <> ?", in.ID)
	}
	q.Count(&count)
	if count > 0 {
		return nil, fmt.Errorf("[%d]数据源名称已存在", ErrCodeDsNameDup)
	}

	if in.ID > 0 {
		var ds model.LlmopsDataSource
		if err := global.GVA_DB.First(&ds, in.ID).Error; err != nil {
			return nil, fmt.Errorf("[%d]数据源不存在", ErrCodeDsNotFound)
		}
		updates := map[string]any{
			"name": in.Name, "type": in.Type, "host": in.Host, "port": in.Port,
			"database": in.Database, "user": in.User, "notes": in.Notes,
		}
		if in.Password != "" {
			enc, err := crypto.Encrypt(in.Password)
			if err != nil {
				return nil, fmt.Errorf("密码加密失败: %w", err)
			}
			updates["password_enc"] = enc
		}
		if err := global.GVA_DB.Model(&ds).Updates(updates).Error; err != nil {
			return nil, err
		}
		return &ds, nil
	}
	if in.Password == "" {
		return nil, fmt.Errorf("[%d]新建数据源必须提供密码", ErrCodeDsTypeBad)
	}
	enc, err := crypto.Encrypt(in.Password)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}
	ds := &model.LlmopsDataSource{
		Name: in.Name, Type: in.Type, Host: in.Host, Port: in.Port,
		Database: in.Database, User: in.User, PasswordEnc: enc, Notes: in.Notes,
	}
	if err := global.GVA_DB.Create(ds).Error; err != nil {
		return nil, err
	}
	return ds, nil
}

// GetDataSources 数据源列表（密码不回显）
func (s *LlmopsServiceStruct) GetDataSources() ([]model.LlmopsDataSource, error) {
	var list []model.LlmopsDataSource
	err := global.GVA_DB.Order("id").Find(&list).Error
	return list, err
}

// DeleteDataSource 数据源删除（软删除）
func (s *LlmopsServiceStruct) DeleteDataSource(id uint) error {
	res := global.GVA_DB.Delete(&model.LlmopsDataSource{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("[%d]数据源不存在", ErrCodeDsNotFound)
	}
	return nil
}

// loadSourceWithPassword 取数据源并解密密码（即用即弃）
func (s *LlmopsServiceStruct) loadSourceWithPassword(id uint) (*DataSourceIn, error) {
	var ds model.LlmopsDataSource
	if err := global.GVA_DB.First(&ds, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("[%d]数据源不存在", ErrCodeDsNotFound)
		}
		return nil, err
	}
	pwd, err := crypto.Decrypt(ds.PasswordEnc)
	if err != nil {
		return nil, fmt.Errorf("密码解密失败: %w", err)
	}
	return &DataSourceIn{
		ID: ds.ID, Name: ds.Name, Type: ds.Type, Host: ds.Host, Port: ds.Port,
		Database: ds.Database, User: ds.User, Password: pwd,
	}, nil
}

// pingDB 连接测试（5s 超时）
func pingDB(in *DataSourceIn) error {
	db, err := openDS(in)
	if err != nil {
		return err
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), dsConnTimeout)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("[%d]连接失败: %w", ErrCodeDsConnFail, err)
	}
	return nil
}

// TestDataSource 测试已保存数据源连通性
func (s *LlmopsServiceStruct) TestDataSource(id uint) error {
	in, err := s.loadSourceWithPassword(id)
	if err != nil {
		return err
	}
	return pingDB(in)
}

// TestDataSourceDirect 直连测试（未保存的表单参数）
func (s *LlmopsServiceStruct) TestDataSourceDirect(in *DataSourceIn) error {
	return pingDB(in)
}

// ListDataSourceTables 库表清单（mysql: SHOW TABLES；postgres: public schema）
func (s *LlmopsServiceStruct) ListDataSourceTables(id uint) ([]string, error) {
	in, err := s.loadSourceWithPassword(id)
	if err != nil {
		return nil, err
	}
	db, err := openDS(in)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), dsConnTimeout)
	defer cancel()

	query := "SHOW TABLES"
	if in.Type == "postgres" {
		query = "SELECT tablename FROM pg_tables WHERE schemaname = 'public' ORDER BY tablename"
	}
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("[%d]库表清单查询失败: %w", ErrCodeDsConnFail, err)
	}
	defer rows.Close()
	tables := make([]string, 0, 64)
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}
		tables = append(tables, t)
	}
	return tables, rows.Err()
}

// PreviewDataSourceTable 采样预览（默认 20 行，上限 100；列名+行数据）
func (s *LlmopsServiceStruct) PreviewDataSourceTable(id uint, table string, limit int) (columns []string, rowsOut [][]string, err error) {
	if !ValidTableName(table) {
		return nil, nil, fmt.Errorf("[%d]表名非法（仅允许字母/数字/下划线）", ErrCodeDsTableBad)
	}
	if limit <= 0 || limit > dsPreviewMaxRows {
		limit = 20
	}
	in, err := s.loadSourceWithPassword(id)
	if err != nil {
		return nil, nil, err
	}
	db, err := openDS(in)
	if err != nil {
		return nil, nil, err
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), dsConnTimeout)
	defer cancel()

	// 表名已过 ^[A-Za-z0-9_]+$ 白名单，按方言加引号双保险
	quoted := "`" + table + "`"
	if in.Type == "postgres" {
		quoted = `"` + table + `"`
	}
	rows, err := db.QueryContext(ctx, fmt.Sprintf("SELECT * FROM %s LIMIT %d", quoted, limit))
	if err != nil {
		return nil, nil, fmt.Errorf("[%d]采样查询失败: %w", ErrCodeDsConnFail, err)
	}
	defer rows.Close()
	columns, err = rows.Columns()
	if err != nil {
		return nil, nil, err
	}
	vals := make([]any, len(columns))
	ptrs := make([]any, len(columns))
	for i := range vals {
		ptrs[i] = &vals[i]
	}
	for rows.Next() {
		if err := rows.Scan(ptrs...); err != nil {
			return nil, nil, err
		}
		row := make([]string, len(columns))
		for i, v := range vals {
			if v == nil {
				row[i] = "NULL"
				continue
			}
			switch b := v.(type) {
			case []byte:
				row[i] = string(b)
			default:
				row[i] = fmt.Sprintf("%v", b)
			}
		}
		rowsOut = append(rowsOut, row)
	}
	return columns, rowsOut, rows.Err()
}
