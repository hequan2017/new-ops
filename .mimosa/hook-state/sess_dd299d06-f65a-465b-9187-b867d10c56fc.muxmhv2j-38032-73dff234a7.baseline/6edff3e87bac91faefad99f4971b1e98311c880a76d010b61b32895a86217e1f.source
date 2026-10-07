import service from '@/utils/request'

// 创建接入点
export const createEndpoint = (data) => service({ url: '/container/endpoint', method: 'post', data })
// 更新接入点
export const updateEndpoint = (data) => service({ url: '/container/endpoint', method: 'put', data })
// 删除接入点
export const deleteEndpoint = (params) => service({ url: '/container/endpoint', method: 'delete', params })
// 接入点列表
export const getEndpointList = (params) => service({ url: '/container/endpoint/list', method: 'get', params })
// 手动巡检
export const checkEndpoint = (params) => service({ url: '/container/endpoint/check', method: 'post', params })
