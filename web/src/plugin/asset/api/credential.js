import service from '@/utils/request'

export const createCredential = (data) => {
  return service({ url: '/asset/credential/create', method: 'post', data })
}
export const deleteCredential = (data) => {
  return service({ url: '/asset/credential/delete', method: 'delete', data })
}
export const updateCredential = (data) => {
  return service({ url: '/asset/credential/update', method: 'put', data })
}
export const getCredentialList = (params) => {
  return service({ url: '/asset/credential/list', method: 'get', params })
}
