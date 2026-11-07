import service from '@/utils/request'

// 智能体管理
export const getAgents = () => service({ url: '/llmops/agent/list', method: 'get' })
export const saveAgent = (data) => service({ url: '/llmops/agent', method: 'post', data })
export const deleteAgent = (params) => service({ url: '/llmops/agent', method: 'delete', params })
export const setAgentStage = (data) => service({ url: '/llmops/agent/stage', method: 'post', data })
export const publishAgent = (data) => service({ url: '/llmops/agent/publish', method: 'post', data })
export const getAgentVersions = (params) => service({ url: '/llmops/agent/versions', method: 'get', params })
export const rollbackAgent = (data) => service({ url: '/llmops/agent/rollback', method: 'post', data })

// 会话与调试
export const createSession = (data) => service({ url: '/llmops/agent/session', method: 'post', data })
export const getSessions = (params) => service({ url: '/llmops/agent/sessions', method: 'get', params })
export const getMessages = (params) => service({ url: '/llmops/agent/messages', method: 'get', params })
export const debugChat = (data) => service({ url: '/llmops/agent/debugChat', method: 'post', data, timeout: 140000 })
export const agentChat = (data) => service({ url: '/llmops/agent/chat', method: 'post', data, timeout: 140000 })

// 评测 / 记忆 / 多智能体 / 监控
export const getAgentEvals = () => service({ url: '/llmops/agent/eval/list', method: 'get' })
export const saveAgentEval = (data) => service({ url: '/llmops/agent/eval', method: 'post', data })
export const runAgentEval = (id) => service({ url: '/llmops/agent/eval/run', method: 'post', data: { id }, timeout: 180000 })
export const getAgentEvalResults = (params) => service({ url: '/llmops/agent/eval/results', method: 'get', params })
export const saveMemory = (data) => service({ url: '/llmops/agent/memory', method: 'post', data })
export const listMemories = (params) => service({ url: '/llmops/agent/memory/list', method: 'get', params })
export const deleteMemory = (params) => service({ url: '/llmops/agent/memory', method: 'delete', params })
export const multiRun = (data) => service({ url: '/llmops/agent/multiRun', method: 'post', data, timeout: 180000 })
export const getAgentMonitor = (params) => service({ url: '/llmops/agent/monitor', method: 'get', params })
export const getAbnormalSessions = (params) => service({ url: '/llmops/agent/abnormalSessions', method: 'get', params })
