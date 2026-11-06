import service from '@/utils/request'

// 数据资产
export const getAssetList = (params) => service({ url: '/llmops/asset/list', method: 'get', params })
export const getAsset = (params) => service({ url: '/llmops/asset', method: 'get', params })
export const deleteAsset = (params) => service({ url: '/llmops/asset', method: 'delete', params })

// 采集导入（multipart）
export const importDoc = (file, name, tags) => {
  const data = new FormData()
  data.append('file', file)
  if (name) data.append('name', name)
  if (tags) data.append('tags', tags)
  return service({ url: '/llmops/asset/importDoc', method: 'post', data })
}
export const importQa = (file, name, tags) => {
  const data = new FormData()
  data.append('file', file)
  if (name) data.append('name', name)
  if (tags) data.append('tags', tags)
  return service({ url: '/llmops/asset/importQa', method: 'post', data })
}

// 网页采集
export const importWeb = (data) => service({ url: '/llmops/asset/importWeb', method: 'post', data })

// 分类
export const getCategoryTree = () => service({ url: '/llmops/category/tree', method: 'get' })

// 事件流水
export const getAssetEvents = (params) => service({ url: '/llmops/asset/events', method: 'get', params })
