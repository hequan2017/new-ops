// Package model eduadmin 教育管理一模通管：自动报销（需求 3.6-16）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// EduInvoice 报销发票（状态机：提交→审核中→支付/驳回）
type EduInvoice struct {
	global.GVA_MODEL
	InvoiceNo    string  `json:"invoiceNo" gorm:"type:varchar(100);uniqueIndex;comment:发票号"`
	Dept         string  `json:"dept" gorm:"type:varchar(100);index;comment:报销部门"`
	Applicant    string  `json:"applicant" gorm:"type:varchar(100);index;comment:申请人"`
	Category     string  `json:"category" gorm:"type:varchar(32);comment:差旅/办公/设备/招待"`
	Amount       float64 `json:"amount" gorm:"comment:金额(元)"`
	InvoiceDate  string  `json:"invoiceDate" gorm:"type:varchar(16);comment:发票日期"`
	ImageMeta    string  `json:"imageMeta" gorm:"type:varchar(500);comment:发票图像元数据(路径/哈希)"`
	Status       string  `json:"status" gorm:"type:varchar(16);index;comment:提交/审核中/支付/驳回"`
	RejectReason string  `json:"rejectReason" gorm:"type:varchar(500);comment:驳回原因"`
}

func (EduInvoice) TableName() string { return "edu_invoice" }
