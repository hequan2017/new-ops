import service from '@/utils/request'

// 数据集
export const getDatasets = (params) => service({ url: '/llmops/dataset/list', method: 'get', params })
export const saveDataset = (data) => service({ url: '/llmops/dataset', method: 'post', data })
export const deleteDataset = (params) => service({ url: '/llmops/dataset', method: 'delete', params })
export const importDataset = (data) => service({ url: '/llmops/dataset/import', method: 'post', data })
export const exportDataset = (params) => service({ url: '/llmops/dataset/export', method: 'get', params })
export const getDatasetRefs = (params) => service({ url: '/llmops/dataset/refs', method: 'get', params })
