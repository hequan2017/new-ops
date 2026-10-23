import service from '@/utils/request'

// Agent 实例列表（状态按心跳时效计算）
export const getAgentInstanceList = (params) => service({ url: '/agent/instance/list', method: 'get', params })
// 签发/重置接入令牌（明文仅本次返回）
export const issueAgentToken = (data) => service({ url: '/agent/instance/token', method: 'post', data })
// 吊销 Agent 注册
export const revokeAgentInstance = (params) => service({ url: '/agent/instance', method: 'delete', params })
// 文件分发（server 本地文件 → agent 目标机绝对路径，SHA256 校验落盘）
export const pushAgentFile = (data) => service({ url: '/agent/file/push', method: 'post', data })
// Agent 任务留痕分页
export const getAgentTaskList = (params) => service({ url: '/agent/task/list', method: 'get', params })
