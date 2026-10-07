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
        <el-table-column label="操作" width="340" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" icon="link" :loading="testingId === row.ID" @click="onTest(row)">连接测试</el-button>
            <el-button link type="success" icon="grid" @click="openBrowser(row)">资源浏览</el-button>
            <el-button link type="primary" icon="data-line" @click="openOverview(row)">总览</el-button>
            <el-button link type="warning" icon="key" @click="openGrant(row)">命名空间授权</el-button>
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

    <ResourceBrowser
      v-model:visible="browserVisible"
      :cluster-id="browserClusterId"
      :cluster-name="browserName"
      @scale="onScale"
      @restart="onRestart"
    />

    <OverviewDrawer v-model:visible="overviewVisible" :cluster-id="overviewClusterId" :cluster-name="overviewName" />

    <el-drawer v-model="grantVisible" :title="`命名空间授权 · ${grantClusterName}`" size="55%" append-to-body>
      <div class="ops-btn-list" style="margin-bottom: 8px">
        <el-select v-model="grantForm.namespace" size="small" style="width: 180px" placeholder="命名空间">
          <el-option v-for="n in grantNsOptions" :key="n.name" :label="n.name" :value="n.name" />
        </el-select>
        <el-select v-model="grantForm.userId" size="small" style="width: 160px" filterable placeholder="用户">
          <el-option v-for="u in grantUsers" :key="u.ID" :label="u.nickName || u.userName" :value="u.ID" />
        </el-select>
        <el-button type="primary" size="small" icon="plus" @click="submitGrant">授权</el-button>
        <el-button size="small" icon="refresh" @click="loadGrants">刷新</el-button>
      </div>
      <el-table :data="grants" size="small" border>
        <el-table-column prop="namespace" label="命名空间" min-width="140" />
        <el-table-column prop="username" label="用户名" min-width="120" />
        <el-table-column prop="nickName" label="昵称" min-width="120" />
        <el-table-column label="操作" width="80">
          <template #default="{ row }">
            <el-button link type="danger" icon="delete" @click="onDeleteGrant(row)">收回</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-drawer>

    <ScaleDialog ref="scaleDialogRef" @done="onScaleDone" />
  </div>
</template>

<script setup>
  import {
    createK8sCluster,
    deleteK8sCluster,
    getK8sClusterList,
    testK8sCluster
  } from '@/plugin/k8s/api/k8sCluster'
  import { getNsVisibility, getNsGrantList, createNsGrant, deleteNsGrant } from '@/plugin/k8s/api/k8sResource'
  import { getUserList } from '@/api/user'
  import ResourceBrowser from '@/plugin/k8s/components/ResourceBrowser.vue'
  import OverviewDrawer from '@/plugin/k8s/components/OverviewDrawer.vue'
  import ScaleDialog from '@/plugin/k8s/components/ScaleDialog.vue'
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

  // 资源浏览抽屉
  const browserVisible = ref(false)
  const browserClusterId = ref(0)
  const browserName = ref('')
  const scaleDialogRef = ref(null)
  const openBrowser = (row) => {
    browserClusterId.value = row.ID
    browserName.value = row.name
  }
  const onScale = (row) => {
    scaleDialogRef.value?.open(row, browserClusterId.value)
  }
  const onRestart = (row) => {
    scaleDialogRef.value?.confirmRestart(row, browserClusterId.value)
  }

  // 集群总览抽屉
  const overviewVisible = ref(false)
  const overviewClusterId = ref(0)
  const overviewName = ref('')
  const openOverview = (row) => {
    overviewClusterId.value = row.ID
    overviewName.value = row.name
    overviewVisible.value = true
  }

  // 命名空间授权（三级 RBAC）
  const grantVisible = ref(false)
  const grantClusterId = ref(0)
  const grantClusterName = ref('')
  const grants = ref([])
  const grantNsOptions = ref([])
  const grantUsers = ref([])
  const grantForm = ref({ namespace: '', userId: null })

  const loadGrants = async () => {
    const res = await getNsGrantList({ clusterId: grantClusterId.value })
    if (res.code === 0) grants.value = res.data || []
  }

  const openGrant = async (row) => {
    grantClusterId.value = row.ID
    grantClusterName.value = row.name
    grantVisible.value = true
    loadGrants()
    getNsVisibility({ clusterId: row.ID }).then((res) => {
      if (res.code === 0) grantNsOptions.value = res.data || []
    })
    getUserList({ page: 1, pageSize: 100 }).then((res) => {
      if (res.code === 0) grantUsers.value = res.data.list || []
    })
  }

  const submitGrant = async () => {
    if (!grantForm.value.namespace || !grantForm.value.userId) {
      ElMessage.warning('命名空间与用户必填')
      return
    }
    const res = await createNsGrant({
      clusterId: grantClusterId.value,
      namespace: grantForm.value.namespace,
      userId: grantForm.value.userId
    })
    if (res.code === 0) {
      ElMessage.success('授权成功')
      grantForm.value = { namespace: '', userId: null }
      loadGrants()
    }
  }

  const onDeleteGrant = (row) => {
    ElMessageBox.confirm(`确定收回 ${row.username} 在「${row.namespace}」的访问权限吗？`, '提示', {
      confirmButtonText: '收回', cancelButtonText: '取消', type: 'warning'
    }).then(async () => {
      const res = await deleteNsGrant(row.ID)
      if (res.code === 0) {
        ElMessage.success('已收回')
        loadGrants()
      }
    })
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
