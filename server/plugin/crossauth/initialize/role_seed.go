// Package initialize crossauth：角色模板种子（需求 3.7-5，8 条按角色名幂等）
package initialize

import (
	"context"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/crossauth/model"
)

// roleTemplateSeed 角色模板 8 条
var roleTemplateSeed = []model.CaRoleTemplate{
	{RoleName: "跨校教师(教学)", Category: "教学", Description: "跨校授课教师：课程资源与教学数据", PermJSON: `{"grants":[{"resource":"教学资源库","actions":["read","download"]},{"resource":"课程管理","actions":["read","update"]}]}`},
	{RoleName: "跨校科研人员", Category: "科研", Description: "跨校科研人员：算力与实验平台", PermJSON: `{"grants":[{"resource":"算力平台","actions":["read","apply","use"]},{"resource":"实验平台","actions":["read","reserve"]}]}`},
	{RoleName: "跨校学生(实训)", Category: "实训", Description: "跨校实训学生：实训课程与工具集", PermJSON: `{"grants":[{"resource":"实训环境","actions":["read","use"]},{"resource":"学科工具集","actions":["read","execute"]}]}`},
	{RoleName: "跨校管理员(校级)", Category: "管理", Description: "校级管理员：全校配置与审计", PermJSON: `{"grants":[{"resource":"平台管理","actions":["read","update","admin"]},{"resource":"审计日志","actions":["read"]}]}`},
	{RoleName: "跨校管理员(院系级)", Category: "管理", Description: "院系级管理员：本院用户与资源", PermJSON: `{"grants":[{"resource":"院系管理","actions":["read","update"]},{"resource":"用户管理","actions":["read"]}]}`},
	{RoleName: "企业用户", Category: "企业", Description: "产学合作企业：公开课与合作项目", PermJSON: `{"grants":[{"resource":"公开课","actions":["read"]},{"resource":"合作项目","actions":["read","update"]}]}`},
	{RoleName: "审计员", Category: "审计", Description: "审计员：三级审计记录只读", PermJSON: `{"grants":[{"resource":"审计日志","actions":["read","export"]}]}`},
	{RoleName: "运维人员", Category: "运维", Description: "运维人员：集群与安全事件处置", PermJSON: `{"grants":[{"resource":"服务器群","actions":["read","operate"]},{"resource":"安全事件","actions":["read","handle"]}]}`},
}

// SeedCaRoleTemplates 角色模板种子（按角色名幂等）
func SeedCaRoleTemplates(ctx context.Context) {
	for i := range roleTemplateSeed {
		var exist model.CaRoleTemplate
		err := global.GVA_DB.WithContext(ctx).Where("role_name = ?", roleTemplateSeed[i].RoleName).First(&exist).Error
		if err != nil {
			if err := global.GVA_DB.WithContext(ctx).Create(&roleTemplateSeed[i]).Error; err != nil {
				global.GVA_LOG.Error("crossauth 角色模板种子失败: " + err.Error())
			}
		}
	}
}
