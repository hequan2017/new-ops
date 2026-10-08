<template>
  <div>
    <div class="ops-card">
      <el-form inline>
        <el-form-item label="集群">
          <el-select v-model="diagForm.clusterId" style="width: 180px" @change="onClusterChange">
            <el-option v-for="cl in clusters" :key="cl.ID" :label="cl.name" :value="cl.ID" />
          </el-select>
        </el-form-item>
        <el-form-item label="命名空间">
          <el-select v-model="diagForm.namespace" filterable allow-create style="width: 170px" @change="loadPods">
            <el-option v-for="ns in namespaces" :key="ns" :label="ns" :value="ns" />
          </el-select>
        </el-form-item>
        <el-form-item label="Pod">
          <el-select v-model="diagForm.pod" filterable style="width: 240px">
            <el-option v-for="p in pods" :key="p.name" :label="`${p.name}（${p.status}）`" :value="p.name" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="diagLoading" @click="doDiagnose">开始诊断</el-button>
          <el-button @click="loadHistory">诊断历史</el-button>
        </el-form-item>
      </el-form>
      <el-alert type="info" :closable="false"
        title="采集 Pod 状态/容器信息/事件/日志尾部（50 行），经 LLM 网关生成根因分析与修复建议，结论留档。" />
    </div>

    <div v-if="result" class="ops-card" style="margin-top: 12px">
      <h4 style="margin: 0 0 10px">诊断结论 · {{ result.pod }}（{{ result.provider }} / {{ result.model }}）</h4>
      <pre class="ops-pre">{{ result.analysis }}</pre>
      <el-collapse style="margin-top: 10px">
        <el-collapse-item title="采集快照（送审 LLM 的原文）">
          <pre class="ops-pre">{{ result.snapshot }}</pre>
        </el-collapse-item>
      </el-collapse>
    </div>

    <div class="ops-card" style="margin-top: 12px">
      <h4 style="margin: 0 0 10px">诊断历史</h4>
      <el-table :data="history" v-loading="histLoading" stripe size="small">
        <el-table-column prop="ID" label="ID" width="60" />
        <el-table-column prop="cluster" label="集群" width="120" />
        <el-table-column label="Pod" min-width="220">
          <template #default="{ row }">{{ row.namespace }}/{{ row.pod }}</template>
        </el-table-column>
        <el-table-column prop="provider" label="供应商" width="120" />
        <el-table-column prop="model" label="模型" width="140" show-overflow-tooltip />
        <el-table-column label="时间" width="160">
          <template #default="{ row }">{{ (row.createdAt || '').replace('T', ' ').slice(0, 19) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="90" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openHistory(row)">详情</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <el-drawer v-model="detailVisible" :title="`诊断详情 #${detail?.ID || ''}`" size="60%">
      <h4 style="margin: 0 0 10px">结论</h4>
      <pre class="ops-pre">{{ detail?.analysis }}</pre>
      <h4 style="margin: 12px 0 10px">采集快照</h4>
      <pre class="ops-pre">{{ detail?.snapshot }}</pre>
    </el-drawer>
  </div>
</template>

<script setup>
  import { ref, reactive, onMounted } from 'vue'
  import { ElMessage } from 'element-plus'
  import { diagnosePod, getDiagnosisList, getDiagnosis } from '@/plugin/aiops/api/aiops'
  import { getK8sClusterList } from '@/plugin/k8s/api/k8sCluster'
  import { getK8sClusterPodList, getNsVisibility } from '@/plugin/k8s/api/k8sResource'

  defineOptions({ name: 'aiopsDiagnose' })

  const clusters = ref([])
  const namespaces = ref([])
  const pods = ref([])
  const diagForm = reactive({ clusterId: null, namespace: '', pod: '' })
  const diagLoading = ref(false)
  const result = ref(null)

  const onClusterChange = async () => {
    pods.value = []
    diagForm.pod = ''
    namespaces.value = []
    diagForm.namespace = ''
    if (!diagForm.clusterId) return
    // ns 下拉与资源浏览同口径：ns-visibility 取当前用户可见（granted）命名空间
    const res = await getNsVisibility({ clusterId: diagForm.clusterId })
    if (res.code === 0) namespaces.value = (res.data || []).filter((n) => n.granted).map((n) => n.name)
  }

  const loadPods = async () => {
    if (!diagForm.clusterId) return
    const res = await getK8sClusterPodList({ clusterId: diagForm.clusterId, namespace: diagForm.namespace || undefined })
    if (res.code === 0) {
      pods.value = res.data || []
      if (diagForm.namespace && diagForm.pod) {
        const hit = pods.value.some((p) => p.name === diagForm.pod)
        if (!hit) diagForm.pod = ''
      }
    }
  }

  const doDiagnose = async () => {
    if (!diagForm.clusterId || !diagForm.namespace || !diagForm.pod) {
      ElMessage.warning('请选择集群、命名空间与 Pod')
      return
    }
    diagLoading.value = true
    result.value = null
    try {
      const cl = clusters.value.find((x) => x.ID === diagForm.clusterId)
      const res = await diagnosePod({
        ...diagForm, cluster: cl?.name || ''
      })
      if (res.code === 0) {
        result.value = res.data
        ElMessage.success('诊断完成')
        loadHistory()
      }
    } finally {
      diagLoading.value = false
    }
  }

  const history = ref([])
  const histLoading = ref(false)
  const loadHistory = async () => {
    histLoading.value = true
    try {
      const res = await getDiagnosisList({ page: 1, pageSize: 20 })
      if (res.code === 0) history.value = res.data.list || []
    } finally {
      histLoading.value = false
    }
  }

  const detailVisible = ref(false)
  const detail = ref(null)
  const openHistory = async (row) => {
    const res = await getDiagnosis({ id: row.ID })
    if (res.code === 0) {
      detail.value = res.data
      detailVisible.value = true
    }
  }

  onMounted(async () => {
    loadHistory()
    const res = await getK8sClusterList()
    if (res.code === 0) {
      clusters.value = res.data || []
      if (clusters.value.length) {
        diagForm.clusterId = clusters.value[0].ID
        onClusterChange()
      }
    }
  })
</script>

<style scoped>
  .ops-pre {
    margin: 0;
    max-height: 420px;
    overflow: auto;
    white-space: pre-wrap;
    word-break: break-word;
    font-size: 12px;
    background: var(--el-fill-color-light);
    padding: 8px;
    border-radius: 4px;
  }
</style>
