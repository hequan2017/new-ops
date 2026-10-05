import service from '@/utils/request'

// 接入点容器列表（实时查询 Docker API）
export const getContainerList = (params) => service({ url: '/container/container/list', method: 'get', params })

// 容器生命周期动作（start/stop/restart/remove）
export const containerAction = (params) => service({ url: '/container/container/action', method: 'post', params })
