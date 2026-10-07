<template>
  <div class="h-full pod-shell">
    <div class="pod-shell-bar">
      <span class="pod-shell-title">{{ podName }}</span>
      <el-select
        v-model="container"
        size="small"
        style="width: 200px"
        placeholder="选择容器"
        @change="reconnect"
      >
        <el-option v-for="c in containers" :key="c" :label="c" :value="c" />
      </el-select>
    </div>
    <PodShellTerm :key="container" :cluster-id="clusterId" :namespace="namespace" :pod="podName" :container="container" />
  </div>
</template>

<script setup>
  import { ref, onMounted } from 'vue'
  import { ElMessage } from 'element-plus'
  import PodShellTerm from '@/plugin/k8s/components/PodShellTerm.vue'
  import { getK8sPodDetail } from '@/plugin/k8s/api/k8sResource'

  const props = defineProps({
    clusterId: { type: Number, required: true },
    namespace: { type: String, required: true },
    pod: { type: String, required: true }
  })

  const podName = ref(props.pod)
  const containers = ref([])
  const container = ref('')
  const reconnect = () => {}

  onMounted(async () => {
    const res = await getK8sPodDetail({
      clusterId: props.clusterId, namespace: props.namespace, name: props.pod
    })
    if (res.code === 0 && res.data?.containers?.length) {
      containers.value = res.data.containers.map((c) => c.name)
      container.value = containers.value[0]
    } else {
      // 详情失败时不阻塞：后端空 container 会取首个容器
      container.value = ''
      ElMessage.warning(res.msg || '容器列表获取失败，使用默认容器')
    }
  })
</script>

<style scoped>
.pod-shell-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 8px;
}
.pod-shell-title {
  font-size: 13px;
  color: #606266;
}
</style>
