import service from '@/utils/request'

// ---------- 脚本库 ----------
export const createScript = (data) => service({ url: '/job/script', method: 'post', data })
export const updateScript = (data) => service({ url: '/job/script', method: 'put', data })
export const deleteScript = (params) => service({ url: '/job/script', method: 'delete', params })
export const getScriptList = (params) => service({ url: '/job/script/list', method: 'get', params })
export const getScriptVersions = (params) => service({ url: '/job/script/versions', method: 'get', params })

// ---------- 变量组 ----------
export const createVarGroup = (data) => service({ url: '/job/vargroup', method: 'post', data })
export const updateVarGroup = (data) => service({ url: '/job/vargroup', method: 'put', data })
export const deleteVarGroup = (params) => service({ url: '/job/vargroup', method: 'delete', params })
export const getVarGroupList = (params) => service({ url: '/job/vargroup/list', method: 'get', params })
