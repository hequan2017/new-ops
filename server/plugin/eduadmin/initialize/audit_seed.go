// Package initialize eduadmin：审计知识语料种子（需求 3.6-37，六域 20 条按标题幂等）
package initialize

import (
	"context"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/eduadmin/model"
)

// auditSeed 审计条目 20 条（财务 5/采购 4/人事 3/资产 3/工程 3/数据安全 2）
var auditSeed = []model.EduAuditItem{
	{Domain: "财务", Title: "预算执行审计要点", RiskPoints: "预算调整未审批、执行率异常偏低、科目间串用", ControlAdvice: "调整留痕、按月执行率看板、科目锁定"},
	{Domain: "财务", Title: "费用报销审计要点", RiskPoints: "超标准报销、连号发票、重复报销、审批链缺失", ControlAdvice: "类别限额校验、发票号唯一性检查、双人复核"},
	{Domain: "财务", Title: "科研经费专项审计", RiskPoints: "横向经费私账、劳务费虚领、设备费混用", ControlAdvice: "专账核算、劳务名单比对、设备购置单独审批"},
	{Domain: "财务", Title: "收费合规审计", RiskPoints: "超标准收费、白条收费、收入不入账", ControlAdvice: "收费项目公示、统一收缴系统、日清月结"},
	{Domain: "财务", Title: "往来款项审计", RiskPoints: "长期挂账、坏账未核销、保证金未退", ControlAdvice: "账龄分析、年度清理、核销审批链"},
	{Domain: "采购", Title: "采购程序合规审计", RiskPoints: "化整为零规避招标、评委组成不合规、中标价异常", ControlAdvice: "年度采购计划归集、评委随机抽取、价格库比对"},
	{Domain: "采购", Title: "单一来源采购审计", RiskPoints: "理由不充分、供应商关联关系未披露", ControlAdvice: "专家论证留痕、关联方申报、公示期管理"},
	{Domain: "采购", Title: "合同履行审计", RiskPoints: "未按合同验收、擅自变更标的、尾款超付", ControlAdvice: "验收单与合同逐项比对、变更审批、付款进度控制"},
	{Domain: "采购", Title: "供应商管理审计", RiskPoints: "围标串标迹象、供应商库未年审", ControlAdvice: "投标文件雷同性检测、年度资格复核"},
	{Domain: "人事", Title: "招聘合规审计", RiskPoints: "未按编制招聘、程序缺失、亲属回避未执行", ControlAdvice: "编制前置校验、五步流程留痕、回避申报"},
	{Domain: "人事", Title: "绩效薪酬审计", RiskPoints: "绩效分无依据、津贴超范围发放", ControlAdvice: "绩效模板归档、发放清单与在册名单比对"},
	{Domain: "人事", Title: "社保公积金审计", RiskPoints: "基数不实、漏缴多缴、离职未停缴", ControlAdvice: "基数年度核定、月度对账、离职联动停缴"},
	{Domain: "资产", Title: "固定资产账实相符审计", RiskPoints: "有账无物、有物无账、私自处置", ControlAdvice: "年度盘点、条码管理、处置审批链"},
	{Domain: "资产", Title: "设备出租出借审计", RiskPoints: "无偿外借、租金未收、到期未收回", ControlAdvice: "出借审批、租金台账、到期提醒"},
	{Domain: "资产", Title: "无形资产审计", RiskPoints: "专利转化未备案、软件license超授权数", ControlAdvice: "转化登记、授权数量对账"},
	{Domain: "工程", Title: "基建工程结算审计", RiskPoints: "高估冒算、隐蔽工程签证缺失、材料价差无依据", ControlAdvice: "结算书逐项核对、签证影像留痕、信息价比对"},
	{Domain: "工程", Title: "工程变更审计", RiskPoints: "先施工后补签证、变更累计超合同额未重新招标", ControlAdvice: "变更前置审批、累计变更阈值预警"},
	{Domain: "工程", Title: "维修工程审计", RiskPoints: "化整为零拆分立项、虚假维修", ControlAdvice: "立项归集、现场影像比对"},
	{Domain: "数据安全", Title: "个人信息保护审计", RiskPoints: "超范围收集、脱敏缺失、导出无审批", ControlAdvice: "最小必要采集、脱敏策略、导出审批留痕"},
	{Domain: "数据安全", Title: "账号权限审计", RiskPoints: "离职未回收、高权限长期未复核", ControlAdvice: "离职联动回收、季度权限复核"},
}

// SeedEduAuditItems 审计知识语料种子（按标题幂等）
func SeedEduAuditItems(ctx context.Context) {
	for i := range auditSeed {
		var exist model.EduAuditItem
		err := global.GVA_DB.WithContext(ctx).Where("title = ?", auditSeed[i].Title).First(&exist).Error
		if err != nil {
			if err := global.GVA_DB.WithContext(ctx).Create(&auditSeed[i]).Error; err != nil {
				global.GVA_LOG.Error("eduadmin 审计语料种子失败: " + err.Error())
			}
		}
	}
}
