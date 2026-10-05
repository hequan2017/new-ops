import service from '@/utils/request'

export const createK8sCluster = (data) => {
  return service({ url: '/k8s/cluster/create', method: 'post', data })
}
export const deleteK8sCluster = (data) => {
  return service({ url: '/k8s/cluster/delete', method: 'delete', data })
}
export const getK8sClusterList = (params) => {
  return service({ url: '/k8s/cluster/list', method: 'get', params })
}
export const testK8sCluster = (params) => {
  return service({ url: '/k8s/cluster/test', method: 'get', params })
}
