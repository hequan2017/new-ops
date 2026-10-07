import service from '@/utils/request'

// ---------- 定义 ----------
export const createDefinition = (data) => service({ url: '/workflow/definition', method: 'post', data })
export const updateDefinition = (data) => service({ url: '/workflow/definition', method: 'put', data })
export const deleteDefinition = (params) => service({ url: '/workflow/definition', method: 'delete', params })
export const getDefinitions = (params) => service({ url: '/workflow/definition/list', method: 'get', params })

// ---------- 实例 ----------
export const startInstance = (data) => service({ url: '/workflow/instance', method: 'post', data })
export const submitAction = (params) => service({ url: '/workflow/instance/action', method: 'post', params })
export const getInstances = (data, params) => service({ url: '/workflow/instance/list', method: 'post', data, params })
export const getInstanceLogs = (params) => service({ url: '/workflow/instance/logs', method: 'get', params })
