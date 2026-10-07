import service from '@/utils/request'

// 实例 CRUD 与检测
export const createInstance = (data) => service({ url: '/dbops/instance', method: 'post', data })
export const updateInstance = (data) => service({ url: '/dbops/instance', method: 'put', data })
export const deleteInstance = (params) => service({ url: '/dbops/instance', method: 'delete', params })
export const getInstanceList = (params) => service({ url: '/dbops/instance/list', method: 'get', params })
export const testInstance = (params) => service({ url: '/dbops/instance/test', method: 'post', params })

// SQL 工单
export const createOrder = (data) => service({ url: '/dbops/order', method: 'post', data })
export const auditOrder = (params) => service({ url: '/dbops/order/audit', method: 'post', params })
export const cancelOrder = (params) => service({ url: '/dbops/order/cancel', method: 'post', params })
export const getOrderList = (data) => service({ url: '/dbops/order/list', method: 'post', data })
