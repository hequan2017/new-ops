import service from '@/utils/request'

// 模板
export const getPrompts = () => service({ url: '/llmops/prompt/list', method: 'get' })
export const savePrompt = (data) => service({ url: '/llmops/prompt', method: 'post', data })
export const deletePrompt = (params) => service({ url: '/llmops/prompt', method: 'delete', params })
export const renderPrompt = (data) => service({ url: '/llmops/prompt/render', method: 'post', data, timeout: 140000 })
export const optimizePrompt = (data) => service({ url: '/llmops/prompt/optimize', method: 'post', data, timeout: 140000 })
export const getPromptVersions = (params) => service({ url: '/llmops/prompt/versions', method: 'get', params })

// 配套资产
export const getPromptAssets = (params) => service({ url: '/llmops/prompt/asset/list', method: 'get', params })
export const savePromptAsset = (data) => service({ url: '/llmops/prompt/asset', method: 'post', data })
export const deletePromptAsset = (params) => service({ url: '/llmops/prompt/asset', method: 'delete', params })
