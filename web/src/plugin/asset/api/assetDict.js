import service from '@/utils/request'

// 机房
export const createAssetRoom = (data) => {
  return service({ url: '/asset/room/create', method: 'post', data })
}
export const deleteAssetRoom = (data) => {
  return service({ url: '/asset/room/delete', method: 'delete', data })
}
export const updateAssetRoom = (data) => {
  return service({ url: '/asset/room/update', method: 'put', data })
}
export const getAssetRoomList = (params) => {
  return service({ url: '/asset/room/list', method: 'get', params })
}

// 机柜
export const createAssetRack = (data) => {
  return service({ url: '/asset/rack/create', method: 'post', data })
}
export const deleteAssetRack = (data) => {
  return service({ url: '/asset/rack/delete', method: 'delete', data })
}
export const updateAssetRack = (data) => {
  return service({ url: '/asset/rack/update', method: 'put', data })
}
export const getAssetRackList = (params) => {
  return service({ url: '/asset/rack/list', method: 'get', params })
}

// 产品线
export const createAssetProductLine = (data) => {
  return service({ url: '/asset/productLine/create', method: 'post', data })
}
export const deleteAssetProductLine = (data) => {
  return service({ url: '/asset/productLine/delete', method: 'delete', data })
}
export const updateAssetProductLine = (data) => {
  return service({ url: '/asset/productLine/update', method: 'put', data })
}
export const getAssetProductLineList = (params) => {
  return service({ url: '/asset/productLine/list', method: 'get', params })
}
