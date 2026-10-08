// Package model 白泽容器管理：应用模板与模板部署实例（M8 C3）
package model

import (
	"time"

	"github.com/hequan2017/new-ops/server/global"
)

// DockerAppTemplate 应用模板（compose + {{参数}} 渲染）
type DockerAppTemplate struct {
	global.GVA_MODEL
	Name        string `json:"name" gorm:"comment:模板名;unique" binding:"required"`
	Category    string `json:"category" gorm:"comment:分类(如 web/db/监控);index"`
	Description string `json:"description" gorm:"type:text;comment:描述"`
	ComposeTpl  string `json:"composeTpl" gorm:"type:longtext;comment:compose 模板内容(支持 {{参数}} 占位)"`
	ParamsSchema string `json:"paramsSchema" gorm:"type:text;comment:参数定义(JSON 数组[{key,label,default,required}])"`
	Notes       string `json:"notes" gorm:"type:text;comment:备注"`
}

func (DockerAppTemplate) TableName() string { return "container_app_template" }

// DockerAppInstance 模板部署实例（追溯：哪个项目出自哪个模板与参数）
type DockerAppInstance struct {
	global.GVA_MODEL
	ProjectName  string     `json:"projectName" gorm:"comment:compose 项目名;unique"`
	TemplateID   uint       `json:"templateId" gorm:"comment:模板ID;index"`
	TemplateName string     `json:"templateName" gorm:"comment:模板名快照"`
	EndpointID   uint       `json:"endpointId" gorm:"comment:Docker接入点ID;index"`
	Params       string     `json:"params" gorm:"type:text;comment:渲染参数(JSON)"`
	CreatedBy    string     `json:"createdBy" gorm:"comment:部署人"`
	DeployedAt   *time.Time `json:"deployedAt" gorm:"comment:最近部署时间"`
}

func (DockerAppInstance) TableName() string { return "container_app_instance" }
