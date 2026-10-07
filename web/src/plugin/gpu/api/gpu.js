import service from '@/utils/request'

// 节点
export const createGpuNode = (data) => service({ url: '/gpu/node', method: 'post', data })
export const updateGpuNode = (data) => service({ url: '/gpu/node', method: 'put', data })
export const deleteGpuNode = (params) => service({ url: '/gpu/node', method: 'delete', params })
export const getGpuNodeList = () => service({ url: '/gpu/node/list', method: 'get' })

// 规格
export const createGpuSpec = (data) => service({ url: '/gpu/spec', method: 'post', data })
export const deleteGpuSpec = (params) => service({ url: '/gpu/spec', method: 'delete', params })
export const getGpuSpecList = () => service({ url: '/gpu/spec/list', method: 'get' })

// 实例
export const startGpuInstance = (data) => service({ url: '/gpu/instance', method: 'post', data })
export const releaseGpuInstance = (params) => service({ url: '/gpu/instance/release', method: 'post', params })
export const getGpuInstanceList = (params) => service({ url: '/gpu/instance/list', method: 'get', params })
