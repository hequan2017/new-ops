import service from '@/utils/request'

// 创建 Compose 项目（校验后立即 up）
export const createComposeProject = (data) => service({ url: '/container/compose', method: 'post', data })
// 更新 compose 内容
export const updateComposeProject = (data) => service({ url: '/container/compose', method: 'put', data })
// 删除项目（先 down）
export const deleteComposeProject = (params) => service({ url: '/container/compose', method: 'delete', params })
// 项目列表
export const getComposeList = (params) => service({ url: '/container/compose/list', method: 'get', params })
// 项目详情
export const getComposeDetail = (params) => service({ url: '/container/compose/detail', method: 'get', params })
// 实时 ps
export const composePs = (params) => service({ url: '/container/compose/ps', method: 'get', params })
// up/down/restart
export const composeAction = (data) => service({ url: '/container/compose/action', method: 'post', data })
