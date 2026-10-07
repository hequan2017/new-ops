<template>
  <el-dialog v-model="scaleVisible" title="扩缩容" width="440px">
    <el-form label-width="90px">
      <el-form-item label="Deployment">
        <el-input :model-value="scaleTarget" disabled />
      </el-form-item>
      <el-form-item label="目标副本">
        <el-input-number v-model="scaleReplicas" :min="0" :max="500" style="width: 100%" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="scaleVisible = false">取 消</el-button>
      <el-button type="primary" :loading="scaling" @click="submitScale">确 定</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
  import { ref } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import service from '@/utils/request'

  const scaleVisible = ref(false)
  const scaling = ref(false)
  const scaleTarget = ref('')
  const scaleReplicas = ref(1)
  const scaleTargetNs = ref('')
  const scaleTargetName = ref('')
  const scaleClusterId = ref(0)

  const emit = defineEmits(['done'])

  const open = (row, clusterId) => {
    scaleTarget.value = `${row.namespace}/${row.name}`
    scaleTargetNs.value = row.namespace
    scaleTargetName.value = row.name
    scaleClusterId.value = clusterId
    scaleReplicas.value = row.replicas || 1
    scaleVisible.value = true
  }

  const submitScale = async () => {
    scaling.value = true
    try {
      const params = new URLSearchParams({ clusterId: scaleClusterId.value })
      const form = new FormData()
      form.append('namespace', scaleTargetNs.value)
      form.append('name', scaleTargetName.value)
      form.append('replicas', scaleReplicas.value)
      const res = await service({
        url: `/k8s/deployment/scale?${params}`,
        method: 'post',
        data: form,
        headers: { 'Content-Type': 'multipart/form-data' }
      })
      if (res.code === 0) {
        ElMessage.success('扩缩容已提交')
        scaleVisible.value = false
        emit('done')
      }
    } finally {
      scaling.value = false
    }
  }

  const confirmRestart = (row, clusterId) => {
    ElMessageBox.confirm(
      `确定滚动重启 ${row.namespace}/${row.name} 吗？将触发滚动更新。`,
      '提示',
      { confirmButtonText: '重启', cancelButtonText: '取消', type: 'warning' }
    ).then(async () => {
      const params = new URLSearchParams({ clusterId })
      const form = new FormData()
      form.append('namespace', row.namespace)
      form.append('name', row.name)
      await service({
        url: `/k8s/deployment/restart?${params}`,
        method: 'post',
        data: form,
        headers: { 'Content-Type': 'multipart/form-data' }
      })
      ElMessage.success('滚动重启已提交')
      emit('done')
    })
  }

  defineExpose({ open, confirmRestart })
</script>
