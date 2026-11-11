// Package initialize crossauth：应急预案种子（需求 3.7-16，10 类安全事件按名幂等）
package initialize

import (
	"context"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/crossauth/model"
)

// playbookSeed 10 类常见安全事件预案
var playbookSeed = []model.CaPlaybook{
	{EventName: "DDoS 攻击", ResponseFlow: "监测确认→流量清洗→源头封禁→恢复验证→复盘", Duties: "安全团队牵头，运维配合清洗，网络服务商协同", Measures: "启用高防 IP、限速、黑洞路由"},
	{EventName: "数据泄露", ResponseFlow: "隔离泄露源→影响评估→上报监管→通知受影响人→取证", Duties: "安全团队+法务+归属校管理员", Measures: "吊销凭据、密钥轮换、审计溯源"},
	{EventName: "勒索病毒", ResponseFlow: "断网隔离→样本分析→备份恢复→全网查杀", Duties: "运维牵头处置，安全团队研判", Measures: "切断横向传播、冷备恢复、补丁加固"},
	{EventName: "账号盗用", ResponseFlow: "冻结账号→强制下线→MFA 重置→溯源", Duties: "安全团队处置，身份系统管理员配合", Measures: "临时封禁、强制改密、启用 TOTP"},
	{EventName: "钓鱼攻击", ResponseFlow: "预警通报→邮件网关拦截→用户教育→标记溯源", Duties: "安全团队+各校管理员转发预警", Measures: "黑名单拦截、反诈提示、仿真演练"},
	{EventName: "网页篡改", ResponseFlow: "下线受影响页面→恢复备份→漏洞修复→监测复查", Duties: "运维恢复，开发排查注入点", Measures: "WAF 加固、静态化发布、文件完整性校验"},
	{EventName: "系统瘫痪", ResponseFlow: "快速定级→主备切换→限流降级→根因排查", Duties: "运维牵头，归属校管理员通报用户", Measures: "熔断降级、灰度恢复、容量扩容"},
	{EventName: "越权访问", ResponseFlow: "阻断会话→权限矩阵复核→影响面审计→加固", Duties: "安全团队+平台管理员", Measures: "回收授权、收紧通配规则、三级审计核查"},
	{EventName: "供应链攻击", ResponseFlow: "冻结受影响组件→版本回退→依赖审计→重建信任链", Duties: "运维+开发联合，安全团队评估", Measures: "依赖锁定、镜像签名校验、制品库扫描"},
	{EventName: "物理入侵", ResponseFlow: "现场控制→门禁核查→资产盘点→上报保卫部门", Duties: "机房管理员+保卫处", Measures: "门禁权限复核、监控取证、锁具更换"},
}

// SeedCaPlaybooks 预案种子（按事件名幂等）
func SeedCaPlaybooks(ctx context.Context) {
	for i := range playbookSeed {
		var exist model.CaPlaybook
		err := global.GVA_DB.WithContext(ctx).Where("event_name = ?", playbookSeed[i].EventName).First(&exist).Error
		if err != nil {
			if err := global.GVA_DB.WithContext(ctx).Create(&playbookSeed[i]).Error; err != nil {
				global.GVA_LOG.Error("crossauth 预案种子失败: " + err.Error())
			}
		}
	}
}
