<template>
  <div>
    <div class="ops-card">
      <el-form inline>
        <el-form-item label="关键字">
          <el-input v-model="keyword" placeholder="名称/备注" clearable style="width: 180px" @keyup.enter="getList" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="getList">查 询</el-button>
          <el-button @click="openDialog()">新建变量组</el-button>
        </el-form-item>
      </el-form>

      <el-table :data="tableData" v-loading="loading" stripe>
        <el-table-column prop="ID" label="ID" width="60" />
        <el-table-column prop="name" label="名称" min-width="140" show-overflow-tooltip />
        <el-table-column label="变量" min-width="240">
          <template #default="{ row }">
            <template v-if="parseVars(row.variables).length">
              <el-tag
                v-for="kv in parseVars(row.variables)"
                :key="kv.key"
                size="small"
                style="margin-right: 6px"
              >
                {{ kv.key }}={{ kv.value }}
              </el-tag>
            </template>
            <span v-else class="ops-text-muted">（空）</span>
          </template>
        </el-table-column>
        <el-table-column prop="assetGroupId" label="关联资产组" width="100">
          <template #default="{ row }">{{ row.assetGroupId || '全局' }}</template>
        </el-table-column>
        <el-table-column prop="notes" label="备注" min-width="120" show-overflow-tooltip />
        <el-table-column label="操作" width="140" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openDialog(row)">编辑</el-button>
            <el-button link type="danger" @click="onDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <el-dialog v-model="dialogVisible" :title="form.ID ? '编辑变量组' : '新建变量组'" width="620px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="90px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="form.name" placeholder="如 web-params" />
        </el-form-item>
        <el-form-item label="变量">
          <el-table :data="form.vars" size="small" border>
            <el-table-column label="key" min-width="120">
              <template #default="{ row }"><el-input v-model="row.key" placeholder="port" /></template>
            </el-table-column>
            <el-table-column label="value" min-width="140">
              <template #default="{ row }"><el-input v-model="row.value" placeholder="8080" /></template>
            </el-table-column>
            <el-table-column width="60">
              <template #default="{ $index }">
                <el-button link type="danger" @click="form.vars.splice($index, 1)">删</el-button>
              </template>
            </el-table-column>
          </el-table>
          <el-button style="margin-top: 6px" @click="form.vars.push({ key: '', value: '' })">加一行</el-button>
        </el-form-item>
        <el-form-item label="关联资产组">
          <el-input-number v-model="form.assetGroupId" :min="0" placeholder="0 为全局" style="width: 160px" />
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
  import { createVarGroup, updateVarGroup, deleteVarGroup, getVarGroupList } from '@/plugin/job/api/jobScript'

  defineOptions({ name: 'jobVarGroup' })

  const keyword = ref('')
  const tableData = ref([])
  const loading = ref(false)

  const parseVars = (s) => {
    try {
      return JSON.parse(s || '[]')
    } catch {
      return []
    }
  }

  const getList = async () => {
    loading.value = true
    try {
      const res = await getVarGroupList({ keyword: keyword.value || undefined })
      if (res.code === 0) tableData.value = res.data || []
    } finally {
      loading.value = false
    }
  }

  const dialogVisible = ref(false)
  const formRef = ref(null)
  const emptyForm = () => ({ ID: 0, name: '', vars: [{ key: '', value: '' }], assetGroupId: 0, notes: '' })
  const form = reactive(emptyForm())
  const rules = { name: [{ required: true, message: '请输入名称', trigger: 'blur' }] }

  const openDialog = (row) => {
    Object.assign(form, emptyForm())
    if (row) {
      Object.assign(form, row, { vars: parseVars(row.variables).length ? parseVars(row.variables) : [{ key: '', value: '' }] })
    }
    dialogVisible.value = true
  }

  const submitForm = () => {
    formRef.value.validate(async (valid) => {
      if (!valid) return
      const vars = form.vars.filter((v) => v.key)
      const payload = {
        ID: form.ID,
        name: form.name,
        variables: JSON.stringify(vars),
        assetGroupId: form.assetGroupId || null,
        notes: form.notes
      }
      const res = form.ID ? await updateVarGroup(payload) : await createVarGroup(payload)
      if (res.code === 0) {
        ElMessage.success(res.msg)
        dialogVisible.value = false
        getList()
      }
    })
  }

  const onDelete = (row) => {
    ElMessageBox.confirm(`确定删除变量组「${row.name}」吗？`, '提示', {
      confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning'
    }).then(async () => {
      const res = await deleteVarGroup({ id: row.ID })
      if (res.code === 0) {
        ElMessage.success('删除成功')
        getList()
      }
    })
  }

  onMounted(getList)
</script>
