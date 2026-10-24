// Package initialize 启动期安全守卫
package initialize

import "strings"

// JWTKeyGuard 生产环境弱签名键守卫（纯逻辑，可单测）。
// qmPlus 为 gin-vue-admin 全网皆知的默认签名键：public 环境沿用即任何人可伪造
// 任意 userId/authorityId 的 JWT（完全绕过认证）。命中返回 (true, 原因)，调用方拒绝启动。
func JWTKeyGuard(env, key string) (bool, string) {
	if env != "public" {
		return false, ""
	}
	k := strings.TrimSpace(key)
	if k == "" {
		return true, "生产环境（env=public）JWT signing-key 为空，拒绝启动——请配置强随机签名键（如 openssl rand -hex 32）"
	}
	if k == "qmPlus" {
		return true, "生产环境（env=public）JWT signing-key 为 GVA 著名默认值 qmPlus（全网皆知，可伪造任意身份 JWT），拒绝启动——请更换为强随机签名键（如 openssl rand -hex 32）"
	}
	return false, ""
}
