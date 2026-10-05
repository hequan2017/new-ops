import service from '@/utils/request'

// 集群管理 API 实现在 k8sCluster.js，此处转出口供资源视图统一引用
export {
  getK8sClusterList,
  createK8sCluster,
  deleteK8sCluster,
  testK8sCluster
} from './k8sCluster'

// K8s 资源查询
export const getK8sClusterPodList = (params) => {
  return service({ url: '/k8s/pod/list', method: 'get', params })
}
export const getK8sClusterDeploymentList = (params) => {
  return service({ url: '/k8s/deployment/list', method: 'get', params })
}
export const getK8sClusterNodeList = (params) => {
  return service({ url: '/k8s/node/list', method: 'get', params })
}
export const getK8sClusterPodLogs = (params) => {
  return service({ url: '/k8s/pod/logs', method: 'get', params })
}
