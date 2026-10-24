package initialize

import (
	"errors"

	"github.com/hequan2017/new-ops/server/global"
	"go.uber.org/zap"
)

// Reload 优雅地重新加载系统配置
func Reload() error {
	global.GVA_LOG.Info("正在重新加载系统配置...")

	// 重新加载配置文件
	if err := global.GVA_VP.ReadInConfig(); err != nil {
		global.GVA_LOG.Error("重新读取配置文件失败!", zap.Error(err))
		return err
	}

	// 重新建立数据库连接：先建新连接成功后再关旧库（原顺序先关旧库再建新库，
	// 新库建失败时返回 nil，旧库已关——全站所有 DB 请求 500 直到重启）
	newDB := Gorm()
	if newDB == nil {
		global.GVA_LOG.Error("重建数据库连接失败，保留原连接继续服务")
		return errors.New("重建数据库连接失败，已保留原连接")
	}
	oldDB := global.GVA_DB
	global.GVA_DB = newDB
	if oldDB != nil {
		if db, err := oldDB.DB(); err == nil {
			if err := db.Close(); err != nil {
				global.GVA_LOG.Error("关闭原数据库连接失败!", zap.Error(err))
			}
		}
	}

	// 重新初始化其他配置
	OtherInit()
	DBList()

	if global.GVA_DB != nil {
		// 确保数据库表结构是最新的
		RegisterTables()
	}

	// 重新初始化定时任务
	Timer()

	global.GVA_LOG.Info("系统配置重新加载完成")
	return nil
}
