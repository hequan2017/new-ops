import service from '@/utils/request'

// 数据接入
export const getEduRecords = (params) => service({ url: '/eduadmin/record/list', method: 'get', params })
export const importEduRecords = (data) => service({ url: '/eduadmin/record/import', method: 'post', data, timeout: 120000 })
export const cleanEduRecords = (data) => service({ url: '/eduadmin/clean', method: 'post', data })
export const maskPayload = (data) => service({ url: '/eduadmin/mask', method: 'post', data })

// 知识库
export const searchEduPolicies = (params) => service({ url: '/eduadmin/policy/search', method: 'get', params })
export const searchEduAudit = (params) => service({ url: '/eduadmin/audit/search', method: 'get', params })

// 工具台
export const screenCandidates = (params) => service({ url: '/eduadmin/screen', method: 'get', params })
export const suggestBudget = (params) => service({ url: '/eduadmin/budget/suggest', method: 'get', params })
export const getEduInvoices = (params) => service({ url: '/eduadmin/expense/list', method: 'get', params })
export const getRiskAlerts = (params) => service({ url: '/eduadmin/risk/list', method: 'get', params })
export const scanEduRisk = (data) => service({ url: '/eduadmin/risk/scan', method: 'post', data })
export const getEduDashboard = (params) => service({ url: '/eduadmin/dashboard', method: 'get', params })

// AI-12345 工单
export const getTickets = (params) => service({ url: '/eduadmin/ticket/list', method: 'get', params })
export const createTicket = (data) => service({ url: '/eduadmin/ticket', method: 'post', data })
export const getTicket = (params) => service({ url: '/eduadmin/ticket', method: 'get', params })
export const advanceTicket = (params) => service({ url: '/eduadmin/ticket/advance', method: 'post', params })
export const autoDispatch = (params) => service({ url: '/eduadmin/dispatch/auto', method: 'post', params })
export const getTimelines = (params) => service({ url: '/eduadmin/progress/list', method: 'get', params })
export const getTicketStats = () => service({ url: '/eduadmin/ticketStats', method: 'get' })
export const exportTicketStats = () => service({ url: '/eduadmin/ticketStats/export', method: 'get' })
