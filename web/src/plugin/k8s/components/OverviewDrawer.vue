<template>
  <el-drawer v-model="visible" :title="`集群总览 · ${clusterName}`" size="55%" destroy-on-close>
    <div v-loading="loading">
      <template v-if="overview">
        <el-descriptions :column="2" border size="small">
          <el-descriptions-item label="Kubernetes 版本">{{ overview.version || '-' }}</el-descriptions-item>
          <el-descriptions-item label="API Server">{{ overview.apiServer || '-' }}</el-descriptions-item>
          <el-descriptions-item label="节点">
            共 {{ overview.nodes?.total ?? 0 }} · 就绪 {{ overview.nodes?.ready ?? 0 }} · 隔离 {{ overview.nodes?.cordoned ?? 0 }}
          </el-descriptions-item>
          <el-descriptions-item label="命名空间">{{ overview.namespaces ?? 0 }}</el-descriptions-item>
          <el-descriptions-item label="Pod">
            共 {{ overview.pods?.total ?? 0 }} · Running {{ overview.pods?.running ?? 0 }}
          </el-descriptions-item>
          <el-descriptions-item label="metrics-server">
            <el-tag :type="overview.usage?.metricsAvailable ? 'success' : 'info'" size="small">
              {{ overview.usage?.metricsAvailable ? '可用' : '不可用（用量降级隐藏）' }}
            </el-tag>
          </el-descriptions-item>
        </el-descriptions>

        <template v-if="overview.usage?.metricsAvailable">
          <h4>资源用量（metrics-server 汇总 / 可分配）</h4>
          <div class="usage-block">
            <div class="usage-row">
              <span class="usage-label">CPU</span>
              <el-progress :percentage="clampPercent(overview.usage.cpuPercent)" :stroke-width="16" text-inside />
              <span class="usage-text">{{ overview.usage.cpuUsedCores?.toFixed(2) }} / {{ overview.usage.cpuAllocCores?.toFixed(2) }} 核</span>
            </div>
            <div class="usage-row">
              <span class="usage-label">内存</span>
              <el-progress :percentage="clampPercent(overview.usage.memPercent)" :stroke-width="16" text-inside status="warning" />
              <span class="usage-text">{{ overview.usage.memUsedGB?.toFixed(2) }} / {{ overview.usage.memAllocGB?.toFixed(2) }} GiB</span>
            </div>
          </div>
        </template>
      </template>
    </div>
  </el-drawer>
</template>

<script setup>
  import { getK8sClusterOverview } from '@/plugin/k8s/api/k8sResource'
  import { ref, watch } from 'vue'

  const visible = defineModel('visible', { type: Boolean })
  const props = defineProps({
    clusterId: { type: Number, required: true },
    clusterName: { type: String, default: '' }
  })

  const loading = ref(false)
  const overview = ref(null)

  const clampPercent = (v) => Math.max(0, Math.min(100, Math.round(v || 0)))

  const load = async () => {
    loading.value = true
    try {
      const res = await getK8sClusterOverview({ clusterId: props.clusterId })
      if (res.code === 0) overview.value = res.data
    } finally {
      loading.value = false
    }
  }

  watch(visible, (v) => {
    if (v) load()
  })
</script>

<style scoped>
.usage-block {
  display: flex;
  flex-direction: column;
  gap: 14px;
  margin-top: 10px;
}
.usage-row {
  display: flex;
  align-items: center;
  gap: 12px;
}
.usage-label {
  width: 42px;
  color: #606266;
  font-size: 13px;
}
.usage-row :deep(.el-progress) {
  flex: 1;
}
.usage-text {
  width: 170px;
  font-size: 12px;
  color: #909399;
  text-align: right;
  white-space: nowrap;
}
</style>
