package initialize

import (
	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/task"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

func Timer() {
	go func() {
		var option []cron.Option
		option = append(option, cron.WithSeconds())
		// Reload 会再次调用 Timer：同 taskName 已存在时先摘除，避免每次 reload 叠加一个 entry 重复执行
		global.GVA_Timer.RemoveTaskByName("ClearDB", "定时清理数据库【日志，黑名单】内容")
		// 清理DB定时任务
		_, err := global.GVA_Timer.AddTaskByFunc("ClearDB", "@daily", func() {
			err := task.ClearTable(global.GVA_DB) // 定时任务方法定在task文件包中
			if err != nil {
				global.GVA_LOG.Error("定时清理数据库失败", zap.Error(err))
			}
		}, "定时清理数据库【日志，黑名单】内容", option...)
		if err != nil {
			global.GVA_LOG.Error("注册定时清理任务失败", zap.Error(err))
		}

		// 其他定时任务定在这里 参考上方使用方法

		//_, err := global.GVA_Timer.AddTaskByFunc("定时任务标识", "corn表达式", func() {
		//	具体执行内容...
		//  ......
		//}, option...)
		//if err != nil {
		//	fmt.Println("add timer error:", err)
		//}
	}()
}
