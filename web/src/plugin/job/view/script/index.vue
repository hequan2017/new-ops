<template>
  <div>
    <div class="ops-card">
      <el-form inline>
        <el-form-item label="关键字">
          <el-input v-model="keyword" placeholder="名称/备注" clearable style="width: 180px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="语言">
          <el-select v-model="language" clearable style="width: 120px" @change="onSearch">
            <el-option v-for="l in ['shell', 'python', 'yaml']" :key="l" :label="l" :value="l" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="onSearch">查 询</el-button>
          <el-button @click="openDialog()">新建脚本</el-button>
        </el-form-item>
      </el-form>

      <el-table :data="tableData" v-loading="loading" stripe>
        <el-table-column prop="ID" label="ID" width="60" />
        <el-table-column prop="name" label="名称" min-width="140" show-overflow-tooltip />
        <el-table-column prop="language" label="语言" width="90">
          <template #default="{ row }">
            <el-tag size="small" :type="{ shell: 'success', python: 'warning', yaml: 'info' }[row.language] || 'info'">
              {{ row.language }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="version" label="版本" width="70" />
        <el-table-column prop="notes" label="备注" min-width="140" show-overflow-tooltip />
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openDialog(row)">编辑</el-button>
            <el-button link @click="openVersions(row)">版本历史</el-button>
            <el-button link type="danger" @click="onDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <el-dialog v-model="dialogVisible" :title="form.ID ? '编辑脚本' : '新建脚本'" width="680px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="80px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="form.name" placeholder="如 check-disk" />
        </el-form-item>
        <el-form-item label="语言" prop="language">
          <el-select v-model="form.language" style="width: 160px">
            <el-option v-for="l in ['shell', 'python', 'yaml']" :key="l" :label="l" :value="l" />
          </el-select>
        </el-form-item>
        <el-form-item label="内容">
          <el-input v-model="form.content" type="textarea" :rows="10" placeholder="shell 命令/脚本内容；可用 {{变量名}} 占位，执行时经变量组渲染" />
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

    <el-drawer v-model="versionsVisible" :title="`版本历史 · ${versionsScript}`" size="50%">
      <el-table :data="versions" stripe>
        <el-table-column prop="version" label="版本" width="70" />
        <el-table-column prop="operator" label="操作人" width="100" />
        <el-table-column label="更新时间" width="160">
          <template #default="{ row }">
            {{ (row.UpdatedAt || '').replace('T', ' ').slice(0, 19) }}
          </template>
        </el-table-column>
        <el-table-column label="内容">
          <template #default="{ row }">
            <el-collapse>
              <el-collapse-item :title="`v${row.version} 内容`" :name="row.version">
                <pre class="ops-pre">{{ row.content }}</pre>
              </el-collapse-item>
            </el-collapse>
          </template>
        </el-table-column>
      </el-table>
    </el-drawer>
  </div>
</template>

<script setup>
  import { ref, reactive, onMounted } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import {
    createScript, updateScript, deleteScript, getScriptList, getScriptVersions
  } from '@/plugin/job/api/jobScript'

  defineOptions({ name: 'jobScript' })

  const keyword = ref('')
  const language = ref('')
  const tableData = ref([])
  const loading = ref(false)

  const getList = async () => {
    loading.value = true
    try {
      const res = await getScriptList({ keyword: keyword.value || undefined, language: language.value || undefined })
      if (res.code === 0) tableData.value = res.data || []
    } finally {
      loading.value = false
    }
  }
  const onSearch = getList

  const dialogVisible = ref(false)
  const formRef = ref(null)
  const emptyForm = () => ({ ID: 0, name: '', language: 'shell', content: '', notes: '' })
  const form = reactive(emptyForm())
  const rules = {
    name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
    language: [{ required: true, message: '请选择语言', trigger: 'change' }]
  }

  const openDialog = (row) => {
    Object.assign(form, emptyForm())
    if (row) Object.assign(form, row)
    dialogVisible.value = true
  }

  const submitForm = () => {
    formRef.value.validate(async (valid) => {
      if (!valid) return
      const res = form.ID ? await updateScript(form) : await createScript(form)
      if (res.code === 0) {
        ElMessage.success(res.msg)
        dialogVisible.value = false
        getList()
      }
    })
  }

  const onDelete = (row) => {
    ElMessageBox.confirm(`确定删除脚本「${row.name}」吗？历史版本将一并删除。`, '提示', {
      confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning'
    }).then(async () => {
      const res = await deleteScript({ id: row.ID })
      if (res.code === 0) {
        ElMessage.success('删除成功')
        getList()
      }
    })
  }

  const versionsVisible = ref(false)
  const versions = ref([])
  const versionsScript = ref('')
  const openVersions = async (row) => {
    versionsScript.value = row.name
    const res = await getScriptVersions({ id: row.ID })
    if (res.code === 0) {
      versions.value = res.data || []
      versionsVisible.value = true
    }
  }

  onMounted(getList)
</script>

<style scoped>
  .ops-pre {
    margin: 0;
    max-height: 300px;
    overflow: auto;
    white-space: pre-wrap;
    word-break: break-all;
    font-size: 12px;
    background: var(--el-fill-color-light);
    padding: 8px;
    border-radius: 4px;
  }
</style>
