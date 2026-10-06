<template>
  <div>
    <div class="ops-card">
      <el-form inline>
        <el-form-item label="关键字">
          <el-input v-model="keyword" placeholder="名称/主机/备注" clearable style="width: 180px" @keyup.enter="getList" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="status" clearable style="width: 110px" @change="getList">
            <el-option v-for="s in ['在线', '离线', '未知']" :key="s" :label="s" :value="s" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="getList">查 询</el-button>
          <el-button @click="openDialog()">新建实例</el-button>
        </el-form-item>
      </el-form>

      <el-table :data="tableData" v-loading="loading" stripe>
        <el-table-column prop="ID" label="ID" width="60" />
        <el-table-column prop="name" label="名称" min-width="130" show-overflow-tooltip />
        <el-table-column label="地址" min-width="160">
          <template #default="{ row }">{{ row.host }}:{{ row.port }}</template>
        </el-table-column>
        <el-table-column prop="username" label="账号" width="100" />
        <el-table-column prop="database" label="默认库" width="110" show-overflow-tooltip />
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="{ 在线: 'success', 离线: 'danger' }[row.status] || 'info'" size="small">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="最近检测" width="160">
          <template #default="{ row }">{{ (row.lastCheckAt || '—').replace('T', ' ').slice(0, 19) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openDialog(row)">编辑</el-button>
            <el-button link type="success" :loading="testingId === row.ID" @click="onTest(row)">检测</el-button>
            <el-button link type="danger" @click="onDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <el-dialog v-model="dialogVisible" :title="form.ID ? '编辑实例' : '新建实例'" width="560px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="90px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="form.name" placeholder="如 生产主库" />
        </el-form-item>
        <el-form-item label="主机" prop="host">
          <el-input v-model="form.host" placeholder="IP 或域名" />
        </el-form-item>
        <el-form-item label="端口">
          <el-input-number v-model="form.port" :min="1" :max="65535" style="width: 160px" />
        </el-form-item>
        <el-form-item label="账号">
          <el-input v-model="form.username" />
        </el-form-item>
        <el-form-item label="密码">
          <el-input v-model="form.password" type="password" show-password
            :placeholder="form.ID ? '留空不修改' : '将 AES-GCM 加密落库'" />
        </el-form-item>
        <el-form-item label="默认库">
          <el-input v-model="form.database" />
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
  import { ref, reactive, onMounted } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import {
    createInstance, updateInstance, deleteInstance, getInstanceList, testInstance
  } from '@/plugin/dbops/api/dbops'

  defineOptions({ name: 'dbopsInstance' })

  const keyword = ref('')
  const status = ref('')
  const tableData = ref([])
  const loading = ref(false)
  const testingId = ref(0)

  const getList = async () => {
    loading.value = true
    try {
      const res = await getInstanceList({ keyword: keyword.value || undefined, status: status.value || undefined })
      if (res.code === 0) tableData.value = res.data || []
    } finally {
      loading.value = false
    }
  }

  const onTest = async (row) => {
    testingId.value = row.ID
    try {
      const res = await testInstance({ id: row.ID })
      if (res.code === 0) {
        ElMessage.success(`检测通过：${res.data.detail}`)
      }
      getList()
    } finally {
      testingId.value = 0
    }
  }

  const dialogVisible = ref(false)
  const formRef = ref(null)
  const emptyForm = () => ({ ID: 0, name: '', host: '', port: 3306, username: '', password: '', database: '', notes: '' })
  const form = reactive(emptyForm())
  const rules = {
    name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
    host: [{ required: true, message: '请输入主机', trigger: 'blur' }]
  }

  const openDialog = (row) => {
    Object.assign(form, emptyForm())
    if (row) {
      Object.assign(form, row, { password: '' })
    }
    dialogVisible.value = true
  }

  const submitForm = () => {
    formRef.value.validate(async (valid) => {
      if (!valid) return
      const payload = { ...form, passwordEnc: form.password || '' }
      delete payload.password
      const res = form.ID ? await updateInstance(payload) : await createInstance(payload)
      if (res.code === 0) {
        ElMessage.success(res.msg)
        dialogVisible.value = false
        getList()
      }
    })
  }

  const onDelete = (row) => {
    ElMessageBox.confirm(`确定删除实例「${row.name}」吗？`, '提示', {
      confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning'
    }).then(async () => {
      const res = await deleteInstance({ id: row.ID })
      if (res.code === 0) {
        ElMessage.success('已删除')
        getList()
      }
    })
  }

  onMounted(getList)
</script>
