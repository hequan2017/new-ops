import service from '@/utils/request'

// 应用模板（compose + 参数渲染）
export const getAppTemplateList = () => service({ url: '/container/appTemplate/list', method: 'get' })
export const saveAppTemplate = (data) => service({ url: '/container/appTemplate', method: 'post', data })
export const deleteAppTemplate = (params) => service({ url: '/container/appTemplate', method: 'delete', params })
export const deployAppTemplate = (data) => service({ url: '/container/appTemplate/deploy', method: 'post', data, timeout: 200000 })
export const getAppInstances = () => service({ url: '/container/appTemplate/instances', method: 'get' })
