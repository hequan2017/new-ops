import service from '@/utils/request'

// 算子
export const getOperators = (params) => service({ url: '/oplib/operator/list', method: 'get', params })
export const saveOperator = (data) => service({ url: '/oplib/operator', method: 'post', data })
export const runOperator = (data) => service({ url: '/oplib/operator/run', method: 'post', data, timeout: 60000 })
export const getRunLogs = (params) => service({ url: '/oplib/operator/runLogs', method: 'get', params })
export const deleteOperator = (params) => service({ url: '/oplib/operator', method: 'delete', params })

// 管道
export const getPipelines = () => service({ url: '/oplib/pipeline/list', method: 'get' })
export const savePipeline = (data) => service({ url: '/oplib/pipeline', method: 'post', data })
export const runPipeline = (data) => service({ url: '/oplib/pipeline/run', method: 'post', data, timeout: 120000 })
export const deletePipeline = (params) => service({ url: '/oplib/pipeline', method: 'delete', params })
