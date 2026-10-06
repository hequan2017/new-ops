<template>
  <el-drawer v-model="visible" :title="`资源浏览 · ${clusterName}`" size="75%" destroy-on-close>
    <el-tabs v-model="activeTab">
      <el-tab-pane label="Pods" name="pods">
        <el-table :data="pods" size="small" v-loading="loading">
          <el-table-column prop="name" label="名称" min-width="200" />
          <el-table-column prop="namespace" label="命名空间" min-width="110" />
          <el-table-column prop="status" label="状态" width="110">
            <template #default="{ row }">
              <el-tag :type="row.status === 'Running' ? 'success' : 'warning'" size="small">{{ row.status }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="restarts" label="重启" width="70" />
          <el-table-column prop="node" label="节点" min-width="130" />
          <el-table-column prop="age" label="年龄" width="80" />
          <el-table-column label="操作" width="90" fixed="right">
            <template #default="{ row }">
              <el-button link type="primary" icon="document" @click="showLogs(row)">日志</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>
      <el-tab-pane label="Deployments" name="deployments">
        <el-table :data="deployments" size="small" v-loading="loading">
          <el-table-column prop="name" label="名称" min-width="180" />
          <el-table-column prop="namespace" label="命名空间" min-width="110" />
          <el-table-column label="副本" width="120">
            <template #default="{ row }">{{ row.ready }}/{{ row.replicas }}</template>
          </el-table-column>
          <el-table-column prop="age" label="年龄" width="90" />
        </el-table>
      </el-tab-pane>
      <el-tab-pane label="Nodes" name="nodes">
        <el-table :data="nodes" size="small" v-loading="loading">
          <el-table-column prop="name" label="名称" min-width="180" />
          <el-table-column prop="status" label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="row.status === 'Ready' ? 'success' : 'danger'" size="small">{{ row.status }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="version" label="版本" min-width="150" />
          <el-table-column prop="internal" label="InternalIP" min-width="130" />
          <el-table-column prop="age" label="年龄" width="90" />
        </el-table>
      </el-tab-pane>
      <el-tab-pane label="Services" name="services">
        <el-table :data="services" size="small" v-loading="loading">
          <el-table-column prop="name" label="名称" min-width="160" />
          <el-table-column prop="namespace" label="命名空间" min-width="110" />
          <el-table-column prop="type" label="类型" width="100" />
          <el-table-column prop="clusterIp" label="ClusterIP" min-width="130" />
          <el-table-column prop="ports" label="端口" min-width="140" />
          <el-table-column prop="age" label="年龄" width="90" />
        </el-table>
      </el-tab-pane>
      <el-tab-pane label="ConfigMaps" name="configmaps">
        <el-table :data="configMaps" size="small" v-loading="loading">
          <el-table-column prop="name" label="名称" min-width="160" />
          <el-table-column prop="namespace" label="命名空间" min-width="110" />
          <el-table-column prop="dataKeys" label="数据键" min-width="180" />
          <el-table-column prop="age" label="年龄" width="90" />
        </el-table>
      </el-tab-pane>
      <el-tab-pane label="Secrets" name="secrets">
        <el-table :data="secrets" size="small" v-loading="loading">
          <el-table-column prop="name" label="名称" min-width="160" />
          <el-table-column prop="namespace" label="命名空间" min-width="110" />
          <el-table-column prop="type" label="类型" min-width="140" />
          <el-table-column prop="dataKeys" label="数据键" min-width="180" />
          <el-table-column prop="age" label="年龄" width="90" />
        </el-table>
      </el-tab-pane>
    </el-tabs>

    <el-drawer v-model="logsVisible" :title="`日志 · ${logsPod}`" size="60%" append-to-body>
      <pre class="logs-pre">{{ logsText || '（无日志）' }}</pre>
    </el-drawer>
  </el-drawer>
</template>

<script setup>
  import {
    getK8sClusterList, getK8sClusterPodList, getK8sClusterPodLogs,
    getK8sClusterDeploymentList, getK8sClusterNodeList,
    getK8sServiceList, getK8sConfigMapList, getK8sSecretList
  } from '@/plugin/k8s/api/k8sResource'
  import { ref, watch } from 'vue'

  const visible = defineModel('visible', { type: Boolean })
  const props = defineProps({
    clusterId: { type: Number, required: true },
    clusterName: { type: String, default: '' }
  })

  const activeTab = ref('pods')
  const loading = ref(false)
  const pods = ref([])
  const deployments = ref([])
  const nodes = ref([])
  const services = ref([])
  const configMaps = ref([])
  const secrets = ref([])
  const logsVisible = ref(false)
  const logsPod = ref('')
  const logsText = ref('')

  const loadAll = async () => {
    loading.value = true
    try {
      const [p, d, n, sv, cm, sec] = await Promise.all([
        getK8sClusterPodList({ clusterId: props.clusterId }),
        getK8sClusterDeploymentList({ clusterId: props.clusterId }),
        getK8sClusterNodeList({ clusterId: props.clusterId }),
        getK8sServiceList({ clusterId: props.clusterId }),
        getK8sConfigMapList({ clusterId: props.clusterId }),
        getK8sSecretList({ clusterId: props.clusterId })
      ])
      pods.value = p.code === 0 ? p.data : []
      deployments.value = d.code === 0 ? d.data : []
      nodes.value = n.code === 0 ? n.data : []
      services.value = sv.code === 0 ? sv.data : []
      configMaps.value = cm.code === 0 ? cm.data : []
      secrets.value = sec.code === 0 ? sec.data : []
    } finally {
      loading.value = false
    }
  }

  const showLogs = async (pod) => {
    logsPod.value = pod.name
    const res = await getK8sClusterPodLogs({
      clusterId: props.clusterId, namespace: pod.namespace,
      pod: pod.name, tailLines: 500
    })
    logsText.value = res.code === 0 ? res.data : `获取失败：${res.msg}`
    logsVisible.value = true
  }

  watch(visible, (v) => {
    if (v) loadAll()
  })
</script>

<style scoped>
.logs-pre {
  white-space: pre-wrap;
  word-break: break-all;
  font-size: 12px;
  max-height: 65vh;
  overflow: auto;
}
</style>
