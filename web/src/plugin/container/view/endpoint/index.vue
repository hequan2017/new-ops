<template>
  <div>
    <div class="ops-card">
      <el-form inline>
        <el-form-item label="关键字">
          <el-input v-model="keyword" placeholder="名称/地址/备注" clearable style="width: 180px" @keyup.enter="getList" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="status" clearable style="width: 110px" @change="getList">
            <el-option v-for="s in ['在线', '离线', '未知']" :key="s" :label="s" :value="s" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="getList">查 询</el-button>
          <el-button @click="openDialog()">新建接入点</el-button>
        </el-form-item>
      </el-form>

      <el-table :data="tableData" v-loading="loading" stripe>
        <el-table-column prop="ID" label="ID" width="60" />
        <el-table-column prop="name" label="名称" min-width="130" show-overflow-tooltip />
        <el-table-column prop="addr" label="地址" min-width="220" show-overflow-tooltip />
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="{ 在线: 'success', 离线: 'danger' }[row.status] || 'info'" size="small">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="dockerVersion" label="API版本" width="100" />
        <el-table-column label="最近巡检" width="160">
          <template #default="{ row }">
            {{ (row.lastCheckAt || '—').replace('T', ' ').slice(0, 19) }}
          </template>
        </el-table-column>
        <el-table-column prop="notes" label="备注" min-width="120" show-overflow-tooltip />
        <el-table-column label="操作" width="190" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openDialog(row)">编辑</el-button>
            <el-button link type="success" :loading="checkingId === row.ID" @click="onCheck(row)">巡检</el-button>
            <el-button link type="danger" @click="onDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <el-dialog v-model="dialogVisible" :title="form.ID ? '编辑接入点' : '新建接入点'" width="620px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="form.name" placeholder="如 本机Docker" />
        </el-form-item>
        <el-form-item label="地址" prop="addr">
          <el-input v-model="form.addr" placeholder="unix:///var/run/docker.sock 或 tcp://host:2376" />
        </el-form-item>
        <el-form-item label="TLS 凭据">
          <el-select v-model="form.tlsCredentialId" clearable placeholder="tcp+tls 时选择 docker_tls 凭据" style="width: 100%">
            <el-option
              v-for="c in tlsCredOptions"
              :key="c.ID"
              :label="`${c.name}（TLS）`"
              :value="c.ID"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.notes" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取 消</el-button>
        <el-button type="primary" @click="submitForm">确 定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
  import { ref, reactive, onMounted, onUnmounted } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import {
    createEndpoint, updateEndpoint, deleteEndpoint, getEndpointList, checkEndpoint
  } from '@/plugin/container/api/dockerEndpoint'
  import { getCredentialList } from '@/plugin/asset/api/credential'

  defineOptions({ name: 'containerEndpoint' })

  const keyword = ref('')
  const status = ref('')
  const tableData = ref([])
  const loading = ref(false)
  const tlsCredOptions = ref([])
  const checkingId = ref(0)
  let pollTimer = null

  const loadCreds = async () => {
    const res = await getCredentialList()
    if (res.code === 0) {
      tlsCredOptions.value = (res.data || []).filter((x) => x.type === 'docker_tls')
    }
  }

  const getList = async () => {
    loading.value = true
    try {
      const res = await getEndpointList({ keyword: keyword.value || undefined, status: status.value || undefined })
      if (res.code === 0) tableData.value = res.data || []
    } finally {
      loading.value = false
    }
  }

  const onCheck = async (row) => {
    checkingId.value = row.ID
    try {
      const res = await checkEndpoint({ id: row.ID })
      if (res.code === 0) {
        ElMessage.success(`巡检通过：API ${res.data.version}`)
      }
      getList()
    } finally {
      checkingId.value = 0
    }
  }

  const dialogVisible = ref(false)
  const formRef = ref(null)
  const emptyForm = () => ({ ID: 0, name: '', addr: '', tlsCredentialId: null, notes: '' })
  const form = reactive(emptyForm())
  const rules = {
    name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
    addr: [{ required: true, message: '请输入地址', trigger: 'blur' }]
  }

  const openDialog = (row) => {
    Object.assign(form, emptyForm())
    if (row) Object.assign(form, row)
    dialogVisible.value = true
  }

  const submitForm = () => {
    formRef.value.validate(async (valid) => {
      if (!valid) return
      const res = form.ID ? await updateEndpoint(form) : await createEndpoint(form)
      if (res.code === 0) {
        ElMessage.success(res.msg)
        dialogVisible.value = false
        getList()
      }
    })
  }

  const onDelete = (row) => {
    ElMessageBox.confirm(`确定删除接入点「${row.name}」吗？`, '提示', {
      confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning'
    }).then(async () => {
      const res = await deleteEndpoint({ id: row.ID })
      if (res.code === 0) {
        ElMessage.success('删除成功')
        getList()
      }
    })
  }

  onMounted(() => {
    loadCreds()
    getList()
    pollTimer = setInterval(getList, 35000) // 跟随 30s 巡检周期轻刷新
  })
  onUnmounted(() => {
    if (pollTimer) clearInterval(pollTimer)
  })
</script>
