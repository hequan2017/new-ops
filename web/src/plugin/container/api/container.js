import service from '@/utils/request'

// 接入点容器列表（实时查询 Docker API）
export const getContainerList = (params) => service({ url: '/container/container/list', method: 'get', params })

// 容器生命周期动作（start/stop/restart/remove）
export const containerAction = (params) => service({ url: '/container/container/action', method: 'post', params })

// 创建容器（端口/挂载/环境变量/资源限制）
export const createContainer = (data) => service({ url: '/container/container', method: 'post', data, timeout: 90000 })

// 镜像列表
export const getImageList = (params) => service({ url: '/container/image/list', method: 'get', params })
// 拉取镜像（异步）
export const pullImage = (params) => service({ url: '/container/image/pull', method: 'post', params })
// 拉取状态
export const pullStatus = (params) => service({ url: '/container/image/pull-status', method: 'get', params })
// 删除镜像
export const removeImage = (params) => service({ url: '/container/image', method: 'delete', params })

// 网络列表/创建/删除
export const getNetworkList = (params) => service({ url: '/container/network/list', method: 'get', params })
export const createNetwork = (data) => service({ url: '/container/network', method: 'post', data })
export const removeNetwork = (params) => service({ url: '/container/network', method: 'delete', params })

// 卷列表/删除
export const getVolumeList = (params) => service({ url: '/container/volume/list', method: 'get', params })
export const removeVolume = (params) => service({ url: '/container/volume', method: 'delete', params })

// 容器即时统计
export const getContainerStats = (params) => service({ url: '/container/container/stats', method: 'get', params })
