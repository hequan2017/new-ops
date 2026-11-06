import service from '@/utils/request'

// 知识库
export const getKbList = () => service({ url: '/llmops/kb/list', method: 'get' })
export const saveKb = (data) => service({ url: '/llmops/kb', method: 'post', data })
export const deleteKb = (params) => service({ url: '/llmops/kb', method: 'delete', params })

// 文档
export const ingestDoc = (data) => service({ url: '/llmops/kb/ingest', method: 'post', data })
export const getKbDocs = (params) => service({ url: '/llmops/kb/docs', method: 'get', params })
export const deleteKbDoc = (params) => service({ url: '/llmops/kb/doc', method: 'delete', params })

// 索引与检索
export const indexVector = (data) => service({ url: '/llmops/kb/vector/index', method: 'post', data, timeout: 140000 })
export const indexFulltext = (data) => service({ url: '/llmops/kb/fulltext/index', method: 'post', data })
export const unifiedSearch = (data) => service({ url: '/llmops/knowledge/search', method: 'post', data, timeout: 140000 })
export const getSearchLogs = (params) => service({ url: '/llmops/knowledge/searchLogs', method: 'get', params })

// 摘要
export const generateSummary = (data) => service({ url: '/llmops/kb/summary/generate', method: 'post', data, timeout: 140000 })
export const getDocSummary = (params) => service({ url: '/llmops/kb/summary', method: 'get', params })

// 资产列表（入库选择来源）
export const getAssetList = (params) => service({ url: '/llmops/asset/list', method: 'get', params })
