import service from '@/utils/request'

// 模型注册表
export const getModels = () => service({ url: '/llmops/model/list', method: 'get' })
export const saveModel = (data) => service({ url: '/llmops/model', method: 'post', data })
export const deleteModel = (params) => service({ url: '/llmops/model', method: 'delete', params })

// 在线测试（chat/embed/rerank 三协议）
export const testModel = (data) => service({ url: '/llmops/model/test', method: 'post', data, timeout: 140000 })

// 调用日志 / 能力调度
export const getModelCallLogs = (params) => service({ url: '/llmops/model/callLogs', method: 'get', params })
export const selectModel = (params) => service({ url: '/llmops/model/select', method: 'get', params })

// 模型部署
export const getDeploys = () => service({ url: '/llmops/deploy/list', method: 'get' })
export const saveDeployPlan = (data) => service({ url: '/llmops/deploy', method: 'post', data })
export const executeDeploy = (id) => service({ url: '/llmops/deploy/execute', method: 'post', data: { id } })
export const destroyDeploy = (id) => service({ url: '/llmops/deploy/destroy', method: 'post', data: { id } })
export const deleteDeploy = (params) => service({ url: '/llmops/deploy', method: 'delete', params })
