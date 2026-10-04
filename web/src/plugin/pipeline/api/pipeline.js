import service from '@/utils/request'

// 创建流水线（含阶段与步骤）
export const createPipeline = (data) => service({ url: '/pipeline', method: 'post', data })

// 更新流水线（阶段/步骤整体替换）
export const updatePipeline = (data) => service({ url: '/pipeline', method: 'put', data })

// 删除流水线
export const deletePipeline = (params) => service({ url: '/pipeline', method: 'delete', params })

// 流水线列表（含嵌套定义）
export const getPipelineList = (params) => service({ url: '/pipeline/list', method: 'get', params })

// 流水线详情
export const getPipelineDetail = (params) => service({ url: '/pipeline/find', method: 'get', params })
