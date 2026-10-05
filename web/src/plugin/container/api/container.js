import service from '@/utils/request'

// 接入点容器列表（实时查询 Docker API）
export const getContainerList = (params) => service({ url: '/container/container/list', method: 'get', params })

// 容器生命周期动作（start/stop/restart/remove）
export const containerAction = (params) => service({ url: '/container/container/action', method: 'post', params })

// 创建容器（端口/挂载/环境变量/资源限制）
export const createContainer = (data) => service({ url: '/container/container', method: 'post', data, timeout: 90000 })
