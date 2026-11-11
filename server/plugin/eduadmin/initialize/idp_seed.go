// Package initialize eduadmin：跨校身份认证种子（需求 3.6-34，12 所共建高校按名幂等）
package initialize

import (
	"context"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/eduadmin/model"
)

// idpSeed 12 所共建高校适配器（SAML2.0 8 所 / OAuth2.0 4 所，互认范围口径）
var idpSeed = []model.EduIdpAdapter{
	{School: "上海大学", Protocol: "SAML2.0", MetadataEndpoint: "https://idp.shu.edu/saml/metadata", MutualScope: "教学资源,图书馆", Status: "启用"},
	{School: "复旦大学", Protocol: "SAML2.0", MetadataEndpoint: "https://idp.fudan.edu.cn/saml/metadata", MutualScope: "教学资源,电子书库", Status: "启用"},
	{School: "上海交通大学", Protocol: "SAML2.0", MetadataEndpoint: "https://idp.sjtu.edu.cn/saml/metadata", MutualScope: "算力,实验平台", Status: "启用"},
	{School: "同济大学", Protocol: "SAML2.0", MetadataEndpoint: "https://idp.tongji.edu.cn/saml/metadata", MutualScope: "教学资源,实验平台", Status: "启用"},
	{School: "华东师范大学", Protocol: "SAML2.0", MetadataEndpoint: "https://idp.ecnu.edu.cn/saml/metadata", MutualScope: "教学资源,电子书库", Status: "启用"},
	{School: "华东理工大学", Protocol: "OAuth2.0", MetadataEndpoint: "https://oauth.ecust.edu.cn/authorize", ClientId: "eduadmin-gateway", MutualScope: "教学资源,算力", Status: "启用"},
	{School: "东华大学", Protocol: "OAuth2.0", MetadataEndpoint: "https://oauth.dhu.edu.cn/authorize", ClientId: "eduadmin-gateway", MutualScope: "图书馆,电子书库", Status: "启用"},
	{School: "上海外国语大学", Protocol: "SAML2.0", MetadataEndpoint: "https://idp.shisu.edu.cn/saml/metadata", MutualScope: "图书馆", Status: "启用"},
	{School: "上海财经大学", Protocol: "SAML2.0", MetadataEndpoint: "https://idp.sufe.edu.cn/saml/metadata", MutualScope: "教学资源", Status: "启用"},
	{School: "上海理工大学", Protocol: "OAuth2.0", MetadataEndpoint: "https://oauth.usst.edu.cn/authorize", ClientId: "eduadmin-gateway", MutualScope: "实验平台,算力", Status: "启用"},
	{School: "上海师范大学", Protocol: "SAML2.0", MetadataEndpoint: "https://idp.shnu.edu.cn/saml/metadata", MutualScope: "教学资源,电子书库", Status: "启用"},
	{School: "上海海洋大学", Protocol: "OAuth2.0", MetadataEndpoint: "https://oauth.shou.edu.cn/authorize", ClientId: "eduadmin-gateway", MutualScope: "实验平台", Status: "启用"},
}

// SeedEduIdpAdapters 12 所共建高校种子（按校名幂等）
func SeedEduIdpAdapters(ctx context.Context) {
	for i := range idpSeed {
		var exist model.EduIdpAdapter
		err := global.GVA_DB.WithContext(ctx).Where("school = ?", idpSeed[i].School).First(&exist).Error
		if err != nil {
			if err := global.GVA_DB.WithContext(ctx).Create(&idpSeed[i]).Error; err != nil {
				global.GVA_LOG.Error("eduadmin 身份适配器种子失败: " + err.Error())
			}
		}
	}
}
