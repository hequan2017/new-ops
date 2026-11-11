import service from '@/utils/request'

// 制度库
export const getPolicyDocs = (params) => service({ url: '/policyreg/doc/list', method: 'get', params })
export const getPolicyDoc = (params) => service({ url: '/policyreg/doc', method: 'get', params })
export const importPolicyDoc = (data) => service({ url: '/policyreg/doc', method: 'post', data })
export const cleanPolicyDoc = (params) => service({ url: '/policyreg/clean', method: 'post', params })

// 标注
export const extractEntities = (params) => service({ url: '/policyreg/entity/extract', method: 'post', params })
export const getEntities = (params) => service({ url: '/policyreg/entity/list', method: 'get', params })
export const preLabel = (params) => service({ url: '/policyreg/label/pre', method: 'post', params })
export const getTokenLabels = (params) => service({ url: '/policyreg/label/list', method: 'get', params })
export const updateTokenLabel = (data) => service({ url: '/policyreg/label/token', method: 'post', data })

// 鉴权
export const getAclGrants = (params) => service({ url: '/policyreg/acl/list', method: 'get', params })
export const searchByRole = (params) => service({ url: '/policyreg/acl/search', method: 'get', params })

// 问答台
export const getPolicyPrompts = () => service({ url: '/policyreg/prompt/list', method: 'get' })
export const renderPolicyPrompt = (data) => service({ url: '/policyreg/prompt/render', method: 'post', data })
export const policyQa = (data) => service({ url: '/policyreg/qa', method: 'post', data, timeout: 60000 })
export const getPolicyQaLogs = (params) => service({ url: '/policyreg/qa/logs', method: 'get', params })
