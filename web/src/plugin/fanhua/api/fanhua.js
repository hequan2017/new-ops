import service from '@/utils/request'

// 语料库
export const getCorpora = (params) => service({ url: '/fanhua/corpus/list', method: 'get', params })
export const saveCorpus = (data) => service({ url: '/fanhua/corpus', method: 'post', data })
export const importCorpora = (data) => service({ url: '/fanhua/corpus/import', method: 'post', data, timeout: 120000 })
export const exportCorpora = (params) => service({ url: '/fanhua/corpus/export', method: 'get', params })
export const deleteCorpus = (params) => service({ url: '/fanhua/corpus', method: 'delete', params })

// 平行语料
export const getParallels = (params) => service({ url: '/fanhua/parallel/list', method: 'get', params })
export const searchParallel = (params) => service({ url: '/fanhua/parallel/search', method: 'get', params })
export const exportParallels = (params) => service({ url: '/fanhua/parallel/export', method: 'get', params })

// 标注
export const buildLabels = (data) => service({ url: '/fanhua/label/build', method: 'post', data })
export const getTextLabels = (params) => service({ url: '/fanhua/label/list', method: 'get', params })
export const getLabelStats = () => service({ url: '/fanhua/label/stats', method: 'get' })

// 知识库
export const getVoiceEntries = (params) => service({ url: '/fanhua/voice/list', method: 'get', params })
export const searchCulture = (data) => service({ url: '/fanhua/voice/search', method: 'post', data, timeout: 60000 })

// 审查测试台
export const getValuesHealth = (data) => service({ url: '/fanhua/values/health', method: 'post', data })
export const auditValues = (data) => service({ url: '/fanhua/valuesAudit', method: 'post', data })
export const getAuditLogs = (params) => service({ url: '/fanhua/valuesAudit/logs', method: 'get', params })
export const getPrompts = (params) => service({ url: '/fanhua/prompt/list', method: 'get', params })
export const renderPrompt = (data) => service({ url: '/fanhua/prompt/render', method: 'post', data })

// 城市科学
export const govFusion = (params) => service({ url: '/fanhua/govFusion', method: 'get', params })

// 方言问答
export const dialectQa = (data) => service({ url: '/fanhua/dialectQa', method: 'post', data, timeout: 180000 })
