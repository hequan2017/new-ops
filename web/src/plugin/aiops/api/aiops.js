import service from '@/utils/request'

// LLM 供应商
export const getProviderList = () => service({ url: '/aiops/provider/list', method: 'get' })
export const saveProvider = (data) => service({ url: '/aiops/provider', method: 'post', data })
export const deleteProvider = (params) => service({ url: '/aiops/provider', method: 'delete', params })

// LLM 网关对话
export const llmChat = (data) => service({ url: '/aiops/llm/chat', method: 'post', data })

// Pod 诊断
export const diagnosePod = (data) => service({ url: '/aiops/diagnose', method: 'post', data, timeout: 120000 })
export const getDiagnosisList = (data) => service({ url: '/aiops/diagnosis/list', method: 'post', data })
export const getDiagnosis = (params) => service({ url: '/aiops/diagnosis', method: 'get', params })
