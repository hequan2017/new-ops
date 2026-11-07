import service from '@/utils/request'

// 技能
export const getSkills = (params) => service({ url: '/llmops/skill/list', method: 'get', params })
export const saveSkill = (data) => service({ url: '/llmops/skill', method: 'post', data })
export const deleteSkill = (params) => service({ url: '/llmops/skill', method: 'delete', params })
export const generateSkill = (data) => service({ url: '/llmops/skill/generate', method: 'post', data, timeout: 140000 })
export const exportSkills = (data) => service({ url: '/llmops/skill/export', method: 'post', data })
export const importSkills = (data) => service({ url: '/llmops/skill/import', method: 'post', data })

// MCP 工具
export const getMcpTools = () => service({ url: '/llmops/mcp/list', method: 'get' })
export const saveMcpTool = (data) => service({ url: '/llmops/mcp', method: 'post', data })
export const deleteMcpTool = (params) => service({ url: '/llmops/mcp', method: 'delete', params })

// 连接器
export const getConnectors = () => service({ url: '/llmops/connector/list', method: 'get' })
export const saveConnector = (data) => service({ url: '/llmops/connector', method: 'post', data })
export const testConnector = (id) => service({ url: '/llmops/connector/test', method: 'post', data: { id } })
export const deleteConnector = (params) => service({ url: '/llmops/connector', method: 'delete', params })

// 令牌与审计
export const getSkillTokens = () => service({ url: '/llmops/skill/token/list', method: 'get' })
export const createSkillToken = (data) => service({ url: '/llmops/skill/token', method: 'post', data })
export const revokeSkillToken = (id) => service({ url: '/llmops/skill/token/revoke', method: 'post', data: { id } })
export const getSkillCallLogs = (params) => service({ url: '/llmops/skill/callLogs', method: 'get', params })
