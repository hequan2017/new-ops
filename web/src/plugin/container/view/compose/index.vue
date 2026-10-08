<template>
  <div>
    <div class="ops-card">
      <el-form inline>
        <el-form-item label="关键字">
          <el-input v-model="keyword" placeholder="项目名/备注" clearable style="width: 180px" @keyup.enter="getList" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="getList">查 询</el-button>
          <el-button @click="openCreate">新建项目</el-button>
        </el-form-item>
      </el-form>

      <el-table :data="tableData" v-loading="loading" stripe>
        <el-table-column prop="ID" label="ID" width="60" />
        <el-table-column prop="name" label="项目名" min-width="140" show-overflow-tooltip />
        <el-table-column prop="endpointName" label="接入点" min-width="130" show-overflow-tooltip />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="{ 运行中: 'success', 部分运行: 'warning', 已停止: 'info' }[row.status] || 'info'" size="small">
              {{ row.status }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="最近操作" width="160">
          <template #default="{ row }">
            {{ (row.lastOpAt || row.updatedAt || '—').replace('T', ' ').slice(0, 19) }}
          </template>
        </el-table-column>
        <el-table-column prop="createdBy" label="创建人" width="100" />
        <el-table-column prop="notes" label="备注" min-width="120" show-overflow-tooltip />
        <el-table-column label="操作" width="330" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openDetail(row)">详情</el-button>
            <el-button link type="success" :loading="psLoadingId === row.ID" @click="onPs(row)">ps</el-button>
            <el-button link type="success" @click="onAction(row, 'up')">up</el-button>
            <el-button link type="warning" @click="onAction(row, 'restart')">restart</el-button>
            <el-button link type="warning" @click="onAction(row, 'down')">down</el-button>
            <el-button link type="danger" @click="onDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <el-dialog v-model="createVisible" title="新建 Compose 项目" width="720px">
      <el-form label-width="100px">
        <el-form-item label="项目名" required>
          <el-input v-model="createForm.name" placeholder="小写字母/数字开头，仅含小写字母数字-_，如 baize-demo" />
        </el-form-item>
        <el-form-item label="接入点" required>
          <el-select v-model="createForm.endpointId" style="width: 100%" placeholder="选择 Docker 接入点">
            <el-option v-for="e in endpoints" :key="e.ID" :label="`${e.name}（${e.status}）`" :value="e.ID" />
          </el-select>
        </el-form-item>
        <el-form-item label="compose 内容" required>
          <el-input v-model="createForm.content" type="textarea" :rows="14" spellcheck="false" class="compose-editor" placeholder="services:&#10;  web:&#10;    image: nginx:alpine&#10;    ports:&#10;      - &quot;8080:80&quot;" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="createForm.notes" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createVisible = false">取 消</el-button>
        <el-button type="primary" :loading="creating" @click="submitCreate">校验并部署</el-button>
      </template>
    </el-dialog>

    <el-drawer v-model="detailVisible" :title="`Compose · ${detail?.name || ''}`" size="65%">
      <div v-loading="detailLoading">
        <div class="ops-btn-list" style="margin-bottom: 8px">
          <el-tag :type="{ 运行中: 'success', 部分运行: 'warning', 已停止: 'info' }[detail?.status] || 'info'" size="small">
            {{ detail?.status }}
          </el-tag>
          <el-button size="small" type="primary" @click="onEditContent">编辑内容</el-button>
          <el-button size="small" :loading="psLoadingId === detail?.ID" @click="onPsDetail">ps</el-button>
        </div>
        <h4>服务（最近 ps）</h4>
        <el-table :data="detail?.services || []" size="small" border>
          <el-table-column prop="name" label="容器" min-width="180" show-overflow-tooltip />
          <el-table-column prop="service" label="服务" min-width="120" />
          <el-table-column label="状态" width="110">
            <template #default="{ row }">
              <el-tag :type="row.state === 'running' ? 'success' : 'info'" size="small">{{ row.state }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="ports" label="端口" min-width="150" show-overflow-tooltip />
        </el-table>
        <h4>compose 文件</h4>
        <pre class="compose-pre">{{ detail?.content || '（空）' }}</pre>
      </div>
    </el-drawer>

    <el-dialog v-model="editVisible" title="编辑 compose 内容" width="720px">
      <el-input v-model="editContent" type="textarea" :rows="16" spellcheck="false" class="compose-editor" />
      <template #footer>
        <el-button @click="editVisible = false">取 消</el-button>
        <el-button type="primary" :loading="saving" @click="submitEdit">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
  import { ref, onMounted } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import {
    getComposeList, getComposeDetail, createComposeProject, updateComposeProject,
    deleteComposeProject, composePs, composeAction
  } from '@/plugin/container/api/compose'
  import { getEndpointList } from '@/plugin/container/api/dockerEndpoint'

  const keyword = ref('')
  const loading = ref(false)
  const tableData = ref([])
  const endpoints = ref([])

  const createVisible = ref(false)
  const creating = ref(false)
  const createForm = ref({ name: '', endpointId: null, content: '', notes: '' })

  const detailVisible = ref(false)
  const detailLoading = ref(false)
  const detail = ref(null)
  const psLoadingId = ref(0)

  const editVisible = ref(false)
  const editContent = ref('')
  const saving = ref(false)

  const getList = async () => {
    loading.value = true
    try {
      const res = await getComposeList({ keyword: keyword.value })
      if (res.code === 0) tableData.value = res.data || []
    } finally {
      loading.value = false
    }
  }

  const loadEndpoints = async () => {
    const res = await getEndpointList({})
    if (res.code === 0) endpoints.value = res.data || []
  }

  const openCreate = () => {
    createForm.value = { name: '', endpointId: null, content: '', notes: '' }
    loadEndpoints()
    createVisible.value = true
  }

  const submitCreate = async () => {
    const f = createForm.value
    if (!f.name || !f.endpointId || !f.content.trim()) {
      ElMessage.warning('项目名/接入点/compose 内容必填')
      return
    }
    creating.value = true
    try {
      const res = await createComposeProject(f)
      if (res.code === 0) {
        ElMessage.success('创建成功，已提交部署')
        createVisible.value = false
        getList()
      }
    } finally {
      creating.value = false
    }
  }

  const openDetail = async (row) => {
    detailVisible.value = true
    detailLoading.value = true
    detail.value = null
    try {
      const res = await getComposeDetail({ id: row.ID })
      if (res.code === 0) detail.value = res.data
      else ElMessage.error(res.msg || '详情获取失败')
    } finally {
      detailLoading.value = false
    }
  }

  const onPsDetail = async () => {
    if (!detail.value) return
    psLoadingId.value = detail.value.ID
    try {
      const res = await composePs({ id: detail.value.ID })
      if (res.code === 0) {
        detail.value.services = res.data || []
        ElMessage.success(`ps 完成：${(res.data || []).length} 个服务`)
      } else {
        ElMessage.error(res.msg || 'ps 失败')
      }
    } finally {
      psLoadingId.value = 0
    }
  }

  const onPs = async (row) => {
    psLoadingId.value = row.ID
    try {
      const res = await composePs({ id: row.ID })
      if (res.code === 0) {
        ElMessage.success(`ps 完成：${(res.data || []).length} 个服务`)
        getList()
      } else {
        ElMessage.error(res.msg || 'ps 失败')
      }
    } finally {
      psLoadingId.value = 0
    }
  }

  const onAction = (row, action) => {
    const tip = action === 'down' ? `确定 down 项目「${row.name}」吗？其容器与默认网络将被移除。` : `确定对项目「${row.name}」执行 ${action} 吗？`
    ElMessageBox.confirm(tip, '提示', {
      confirmButtonText: '执行', cancelButtonText: '取消', type: action === 'down' ? 'warning' : 'info'
    }).then(async () => {
      const res = await composeAction({ projectId: row.ID, action })
      if (res.code === 0) {
        ElMessage.success(`${action} 执行成功`)
        getList()
      } else {
        ElMessage.error(res.msg || `${action} 失败`)
      }
    })
  }

  const onEditContent = () => {
    editContent.value = detail.value?.content || ''
    editVisible.value = true
  }

  const submitEdit = async () => {
    if (!editContent.value.trim()) {
      ElMessage.warning('内容不能为空')
      return
    }
    saving.value = true
    try {
      const res = await updateComposeProject({ projectId: detail.value.ID, content: editContent.value })
      if (res.code === 0) {
        ElMessage.success('已保存（配合 up 增量应用变更）')
        editVisible.value = false
        detail.value.content = editContent.value
      }
    } finally {
      saving.value = false
    }
  }

  const onDelete = (row) => {
    ElMessageBox.confirm(
      `确定删除项目「${row.name}」吗？将先 down 移除其容器与默认网络，再删除登记。`,
      '危险操作',
      { confirmButtonText: '删除', cancelButtonText: '取消', type: 'error' }
    ).then(async () => {
      const res = await deleteComposeProject({ id: row.ID })
      if (res.code === 0) {
        ElMessage.success('删除成功')
        getList()
      }
    })
  }

  onMounted(getList)
</script>

<style scoped>
.compose-editor :deep(textarea) {
  font-family: Consolas, Monaco, 'Courier New', monospace;
  font-size: 12px;
  line-height: 1.5;
}
.compose-pre {
  white-space: pre-wrap;
  word-break: break-all;
  font-size: 12px;
  max-height: 45vh;
  overflow: auto;
  background: #f5f7fa;
  border-radius: 4px;
  padding: 10px;
}
</style>
