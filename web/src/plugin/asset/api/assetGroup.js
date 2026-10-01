import service from '@/utils/request'

export const createAssetGroup = (data) => {
  return service({ url: '/asset/group/create', method: 'post', data })
}
export const deleteAssetGroup = (data) => {
  return service({ url: '/asset/group/delete', method: 'delete', data })
}
export const updateAssetGroup = (data) => {
  return service({ url: '/asset/group/update', method: 'put', data })
}
export const getAssetGroupList = () => {
  return service({ url: '/asset/group/list', method: 'get' })
}
