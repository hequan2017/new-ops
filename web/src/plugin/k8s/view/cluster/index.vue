<template>
  <div>
    <el-card shadow="never">
      <warning-bar title="kubeconfig 在服务端以 AES-256-GCM 加密存储，任何接口均不回显其内容；注册时会立即做一次连接测试。" />
      <div class="ops-btn-list">
        <el-input
          v-model="keyword"
          placeholder="集群名称 / 备注"
          clearable
          style="width: 220px"
          @keyup.enter="getList"
          @clear="getList"
        >
          <template #prefix><el-icon><search /></el-icon></template>
        </el-input>
        <el-button type="primary" icon="plus" @click="openDialog()">注册集群</el-button>
      </div>

      <el-table :data="list" style="width: 100%">
        <el-table-column prop="name" label="集群名称" min-width="150" />
        <el-table-column prop="server" label="API Server" min-width="200">
          <template #default="{ row }">{{ row.server || '-' }}</template>
        </el-table-column>
        <el-table-column prop="version" label="版本" width="180">
          <template #default="{ row }">{{ row.version || '-' }}</template>
        </el-table-column>
        <el-table-column prop="nodeCount" label="节点数" width="90">
          <template #default="{ row }">{{ row.nodeCount || '-' }}</template>
        </el-table-column>
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="row.status === '在线' ? 'success' : row.status === '离线' ? 'danger' : 'info'">
              {{ row.status || '-' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="remark" label="备注" min-width="140">
          <template #default="{ row }">{{ row.remark || '-' }}</template>
        </el-table-column>
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" icon="link" :loading="testingId === row.ID" @click="onTest(row)">连接测试</el-button>
            <el-button link type="danger" icon="delete" @click="onDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" title="注册集群" width="560px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="90px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="form.name" placeholder="如 生产集群" />
        </el-form-item>
        <el-form-item label="kubeconfig" prop="kubeconfig">
          <el-input
            v-model="form.kubeconfig"
            type="textarea"
            :rows="10"
            placeholder="粘贴 kubeconfig YAML 内容（服务端加密存储，接口不回显）"
          />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.remark" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取 消</el-button>
        <el-button type="primary" @click="submitForm">注册并测试</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
  import {
    createK8sCluster,
    deleteK8sCluster,
    getK8sClusterList,
    testK8sCluster
  } from '@/plugin/k8s/api/k8sCluster'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { onMounted, reactive, ref } from 'vue'

  defineOptions({ name: 'K8sCluster' })

  const keyword = ref('')
  const list = ref([])
  const testingId = ref(0)
  const dialogVisible = ref(false)
  const formRef = ref(null)
  const emptyForm = () => ({ ID: 0, name: '', kubeconfig: '', remark: '' })
  const form = reactive(emptyForm())
  const rules = {
    name: [{ required: true, message: '请输入集群名称', trigger: 'blur' }],
    kubeconfig: [{ required: true, message: '请粘贴 kubeconfig 内容', trigger: 'blur' }]
  }

  const getList = async () => {
    const res = await getK8sClusterList(keyword.value || undefined)
    if (res.code === 0) list.value = res.data || []
  }

  const openDialog = () => {
    Object.assign(form, emptyForm())
    dialogVisible.value = true
  }

  const submitForm = () => {
    formRef.value.validate(async (valid) => {
      if (!valid) return
      const res = await createK8sCluster(form)
      if (res.code === 0) {
        ElMessage.success('注册成功（连接测试通过）')
        dialogVisible.value = false
        getList()
      }
    })
  }

  const onTest = async (row) => {
    testingId.value = row.ID
    try {
      const res = await testK8sCluster({ id: row.ID })
      if (res.code === 0) ElMessage.success(`连接成功：${res.data}`)
    } finally {
      testingId.value = 0
    }
  }

  const onDelete = (row) => {
    ElMessageBox.confirm(`确定删除集群「${row.name}」吗？`, '提示', {
      confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning'
    }).then(async () => {
      const res = await deleteK8sCluster({ id: row.ID })
      if (res.code === 0) {
        ElMessage.success('删除成功')
        getList()
      }
    })
  }

  onMounted(getList)
</script>
