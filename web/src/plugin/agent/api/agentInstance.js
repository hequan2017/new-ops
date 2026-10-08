import service from '@/utils/request'

// Agent 实例列表（状态按心跳时效计算）
export const getAgentInstanceList = (params) => service({ url: '/agent/instance/list', method: 'get', params })
// 签发/重置接入令牌（明文仅本次返回）
export const issueAgentToken = (data) => service({ url: '/agent/instance/token', method: 'post', data })
// 吊销 Agent 注册
export const revokeAgentInstance = (params) => service({ url: '/agent/instance', method: 'delete', params })
