import service from '@/utils/request'

export const sftpList = (params) => {
  return service({ url: '/term/sftp/list', method: 'get', params })
}
export const sftpMkdir = (data) => {
  return service({ url: '/term/sftp/mkdir', method: 'post', data })
}
export const sftpDelete = (data) => {
  return service({ url: '/term/sftp/delete', method: 'post', data })
}
export const sftpRename = (data) => {
  return service({ url: '/term/sftp/rename', method: 'post', data })
}
export const sftpUpload = (data) => {
  return service({
    url: '/term/sftp/upload',
    method: 'post',
    data,
    headers: { 'Content-Type': 'multipart/form-data' }
  })
}
export const sftpDownload = (params) => {
  return service({
    url: '/term/sftp/download',
    method: 'get',
    params,
    responseType: 'blob'
  })
}
