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

// 触发构建（异步）
export const startBuild = (data) => service({ url: '/pipeline/build/start', method: 'post', data })

// 取消构建
export const cancelBuild = (params) => service({ url: '/pipeline/build/cancel', method: 'post', params })

// 审批放行
export const approveBuild = (params) => service({ url: '/pipeline/build/approve', method: 'post', params })

// 构建分页列表
export const getBuildList = (data) => service({ url: '/pipeline/build/list', method: 'post', data })

// 构建日志
export const getBuildLogs = (params) => service({ url: '/pipeline/build/logs', method: 'get', params })
