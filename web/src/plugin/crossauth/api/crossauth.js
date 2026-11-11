import service from '@/utils/request'

// 身份对接
export const getFederations = () => service({ url: '/crossauth/federation/list', method: 'get' })
export const getSyncUsers = (params) => service({ url: '/crossauth/syncUser/list', method: 'get', params })
export const runCaSync = (params) => service({ url: '/crossauth/federation/sync', method: 'post', params })
export const getCaSyncTasks = (params) => service({ url: '/crossauth/federation/tasks', method: 'get', params })

// 权限管理
export const getTrustRules = (params) => service({ url: '/crossauth/trust/list', method: 'get', params })
export const checkTrust = (data) => service({ url: '/crossauth/trust/check', method: 'post', data })
export const checkMatrix = (data) => service({ url: '/crossauth/matrix/check', method: 'post', data })
export const getRoleTemplates = (params) => service({ url: '/crossauth/roleTemplate/list', method: 'get', params })

// 分级扫描
export const scanClassify = (data) => service({ url: '/crossauth/classify/scan', method: 'post', data })
export const getClassifyResults = (params) => service({ url: '/crossauth/classify/list', method: 'get', params })

// 监控与告警
export const getFiveDim = () => service({ url: '/crossauth/monitor/fiveDim', method: 'get' })
export const getAccessAlerts = (params) => service({ url: '/crossauth/access/alerts', method: 'get', params })
export const getAccessLogs = (params) => service({ url: '/crossauth/access/logs', method: 'get', params })

// 运维工单
export const getOpsTickets = (params) => service({ url: '/crossauth/opsTicket/list', method: 'get', params })
export const createOpsTicket = (data) => service({ url: '/crossauth/opsTicket', method: 'post', data })
export const getOpsMessages = (params) => service({ url: '/crossauth/opsTicket/messages', method: 'get', params })
