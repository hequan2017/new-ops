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
export const getK8sServiceList = (params) => {
  return service({ url: '/k8s/service/list', method: 'get', params })
}
export const getK8sConfigMapList = (params) => {
  return service({ url: '/k8s/configmap/list', method: 'get', params })
}
export const getK8sSecretList = (params) => {
  return service({ url: '/k8s/secret/list', method: 'get', params })
}

// K8s 集群总览与 Node 管理
export const getK8sClusterOverview = (params) => {
  return service({ url: '/k8s/cluster/overview', method: 'get', params })
}
export const getK8sNodeDetail = (params) => {
  return service({ url: '/k8s/node/detail', method: 'get', params })
}
export const cordonK8sNode = (clusterId, name, cordon) => {
  const params = new URLSearchParams({ clusterId })
  const form = new FormData()
  form.append('name', name)
  form.append('cordon', cordon ? 'true' : 'false')
  return service({
    url: `/k8s/node/cordon?${params}`,
    method: 'post',
    data: form,
    headers: { 'Content-Type': 'multipart/form-data' }
  })
}
export const drainK8sNode = (clusterId, name, gracePeriod = -1) => {
  const params = new URLSearchParams({ clusterId })
  const form = new FormData()
  form.append('name', name)
  form.append('gracePeriod', String(gracePeriod))
  return service({
    url: `/k8s/node/drain?${params}`,
    method: 'post',
    data: form,
    headers: { 'Content-Type': 'multipart/form-data' }
  })
}

// K2 补齐：工作负载/Pod 详情/PVC/Ingress/Event
export const getK8sStatefulSetList = (params) => {
  return service({ url: '/k8s/statefulset/list', method: 'get', params })
}
export const getK8sDaemonSetList = (params) => {
  return service({ url: '/k8s/daemonset/list', method: 'get', params })
}
export const getK8sWorkloadYaml = (params) => {
  return service({ url: '/k8s/workload/yaml', method: 'get', params })
}
export const previewK8sWorkloadYaml = (clusterId, data) => {
  return service({ url: '/k8s/workload/diff', method: 'post', params: { clusterId }, data })
}
export const applyK8sWorkloadYaml = (clusterId, data) => {
  return service({ url: '/k8s/workload/apply', method: 'post', params: { clusterId }, data })
}
export const getK8sPodDetail = (params) => {
  return service({ url: '/k8s/pod/detail', method: 'get', params })
}
export const deleteK8sPod = (clusterId, namespace, name) => {
  const params = new URLSearchParams({ clusterId })
  const form = new FormData()
  form.append('namespace', namespace)
  form.append('name', name)
  return service({
    url: `/k8s/pod/delete?${params}`,
    method: 'post',
    data: form,
    headers: { 'Content-Type': 'multipart/form-data' }
  })
}
export const getK8sPvcList = (params) => {
  return service({ url: '/k8s/pvc/list', method: 'get', params })
}
export const getK8sIngressList = (params) => {
  return service({ url: '/k8s/ingress/list', method: 'get', params })
}
export const getK8sEventList = (params) => {
  return service({ url: '/k8s/event/list', method: 'get', params })
}

// Helm release 管理（M8 K3）
export const getHelmList = (params) => {
  return service({ url: '/k8s/helm/list', method: 'get', params })
}
export const getHelmHistory = (params) => {
  return service({ url: '/k8s/helm/history', method: 'get', params })
}
export const installHelmRelease = (clusterId, form) => {
  const params = new URLSearchParams({ clusterId })
  return service({
    url: `/k8s/helm/install?${params}`,
    method: 'post',
    data: form,
    headers: { 'Content-Type': 'multipart/form-data' }
  })
}
export const uninstallHelmRelease = (clusterId, namespace, name) => {
  const params = new URLSearchParams({ clusterId })
  const form = new FormData()
  form.append('namespace', namespace)
  form.append('name', name)
  return service({
    url: `/k8s/helm/uninstall?${params}`,
    method: 'post',
    data: form,
    headers: { 'Content-Type': 'multipart/form-data' }
  })
}
export const rollbackHelmRelease = (clusterId, namespace, name, revision = 0) => {
  const params = new URLSearchParams({ clusterId })
  const form = new FormData()
  form.append('namespace', namespace)
  form.append('name', name)
  form.append('revision', String(revision))
  return service({
    url: `/k8s/helm/rollback?${params}`,
    method: 'post',
    data: form,
    headers: { 'Content-Type': 'multipart/form-data' }
  })
}
export const installHelmFromRepo = (clusterId, data) => {
  const params = new URLSearchParams({ clusterId })
  const form = new FormData()
  Object.keys(data).forEach((k) => form.append(k, data[k]))
  return service({
    url: `/k8s/helm/install-repo?${params}`,
    method: 'post',
    data: form,
    headers: { 'Content-Type': 'multipart/form-data' }
  })
}
export const getHelmReleaseDetail = (params) => {
  return service({ url: '/k8s/helm/detail', method: 'get', params })
}
export const getHelmRepoList = () => {
  return service({ url: '/k8s/helm/repo/list', method: 'get' })
}
export const createHelmRepo = (data) => {
  return service({ url: '/k8s/helm/repo', method: 'post', data })
}
export const deleteHelmRepo = (id) => {
  return service({ url: '/k8s/helm/repo', method: 'delete', params: { id } })
}

// chart 元数据（values 表单模式数据源）
export const getHelmChartMeta = (params) => {
  return service({ url: '/k8s/helm/chart/meta', method: 'get', params })
}
export const inspectHelmChart = (file) => {
  const form = new FormData()
  form.append('chart', file)
  return service({
    url: '/k8s/helm/chart/inspect',
    method: 'post',
    data: form,
    headers: { 'Content-Type': 'multipart/form-data' }
  })
}

// 命名空间授权（三级 RBAC）
export const getNsVisibility = (params) => {
  return service({ url: '/k8s/cluster/ns-visibility', method: 'get', params })
}
export const getNsGrantList = (params) => {
  return service({ url: '/k8s/grant/list', method: 'get', params })
}
export const createNsGrant = (data) => {
  return service({ url: '/k8s/grant', method: 'post', data })
}
export const deleteNsGrant = (id) => {
  return service({ url: '/k8s/grant', method: 'delete', params: { id } })
}
