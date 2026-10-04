<template>
  <div>
    <div class="ops-card">
      <el-form inline>
        <el-form-item label="关键字">
          <el-input v-model="keyword" placeholder="名称/描述" clearable style="width: 180px" @keyup.enter="getList" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="getList">查 询</el-button>
          <el-button @click="openDialog()">新建流水线</el-button>
        </el-form-item>
      </el-form>

      <el-table :data="tableData" v-loading="loading" stripe>
        <el-table-column prop="ID" label="ID" width="60" />
        <el-table-column prop="name" label="名称" min-width="150" show-overflow-tooltip />
        <el-table-column prop="description" label="描述" min-width="160" show-overflow-tooltip />
        <el-table-column label="阶段/步骤" width="110">
          <template #default="{ row }">
            {{ (row.stages || []).length }} 段 / {{ (row.stages || []).reduce((n, s) => n + (s.steps || []).length, 0) }} 步
          </template>
        </el-table-column>
        <el-table-column label="审批 gate" width="100">
          <template #default="{ row }">
            <el-tag v-if="(row.stages || []).some((s) => s.approval)" type="warning" size="small">有</el-tag>
            <span v-else>—</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.enabled ? 'success' : 'info'" size="small">{{ row.enabled ? '启用' : '停用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="140" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openDialog(row)">编排</el-button>
            <el-button link type="danger" @click="onDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <el-dialog v-model="dialogVisible" :title="form.ID ? '编排流水线' : '新建流水线'" width="860px" top="4vh">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="80px">
        <el-row :gutter="12">
          <el-col :span="8">
            <el-form-item label="名称" prop="name">
              <el-input v-model="form.name" placeholder="如 deploy-web" />
            </el-form-item>
          </el-col>
          <el-col :span="10">
            <el-form-item label="描述">
              <el-input v-model="form.description" />
            </el-form-item>
          </el-col>
          <el-col :span="6">
            <el-form-item label="启用">
              <el-switch v-model="form.enabled" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-card v-for="(st, si) in form.stages" :key="si" shadow="never" style="margin-bottom: 10px">
          <template #header>
            <div style="display: flex; align-items: center; gap: 10px">
              <el-tag>阶段 {{ si + 1 }}</el-tag>
              <el-input v-model="st.name" placeholder="阶段名称" style="width: 200px" />
              <el-checkbox v-model="st.approval">人工审批</el-checkbox>
              <el-checkbox v-model="st.continueOnError">失败继续</el-checkbox>
              <el-button link type="danger" style="margin-left: auto" @click="form.stages.splice(si, 1)">删阶段</el-button>
            </div>
          </template>
          <el-table :data="st.steps" size="small">
            <el-table-column label="步骤名" min-width="120">
              <template #default="{ row }"><el-input v-model="row.name" placeholder="如 编译" /></template>
            </el-table-column>
            <el-table-column label="类型" width="100">
              <template #default="{ row }">
                <el-select v-model="row.type" style="width: 84px">
                  <el-option label="shell" value="shell" />
                  <el-option label="http" value="http" />
                </el-select>
              </template>
            </el-table-column>
            <el-table-column label="内容 / 地址" min-width="220">
              <template #default="{ row }">
                <el-input v-if="row.type === 'shell'" v-model="row.shellContent" type="textarea" :rows="2" placeholder="shell 命令" />
                <el-input v-else v-model="row.httpUrl" placeholder="http(s)://…" />
              </template>
            </el-table-column>
            <el-table-column label="超时s" width="90">
              <template #default="{ row }"><el-input-number v-model="row.timeoutSec" :min="0" :max="3600" style="width: 76px" /></template>
            </el-table-column>
            <el-table-column width="50">
              <template #default="{ $index }">
                <el-button link type="danger" @click="st.steps.splice($index, 1)">删</el-button>
              </template>
            </el-table-column>
          </el-table>
          <el-button style="margin-top: 6px" size="small" @click="st.steps.push(emptyStep())">加步骤</el-button>
        </el-card>

        <el-button @click="form.stages.push(emptyStage())">加阶段</el-button>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取 消</el-button>
        <el-button type="primary" :loading="saving" @click="submitForm">保 存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
  import { ref, reactive, onMounted } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import {
    createPipeline, updatePipeline, deletePipeline, getPipelineList
  } from '@/plugin/pipeline/api/pipeline'

  defineOptions({ name: 'pipelineList' })

  const keyword = ref('')
  const tableData = ref([])
  const loading = ref(false)

  const getList = async () => {
    loading.value = true
    try {
      const res = await getPipelineList({ keyword: keyword.value || undefined })
      if (res.code === 0) tableData.value = res.data || []
    } finally {
      loading.value = false
    }
  }

  const emptyStep = () => ({ name: '', type: 'shell', shellContent: '', httpUrl: '', timeoutSec: 0 })
  const emptyStage = () => ({ name: '', approval: false, continueOnError: false, steps: [emptyStep()] })

  const dialogVisible = ref(false)
  const saving = ref(false)
  const formRef = ref(null)
  const emptyForm = () => ({ ID: 0, name: '', description: '', enabled: true, stages: [emptyStage()] })
  const form = reactive(emptyForm())
  const rules = { name: [{ required: true, message: '请输入名称', trigger: 'blur' }] }

  const openDialog = (row) => {
    Object.assign(form, emptyForm())
    if (row) {
      const stages = (row.stages || []).map((s) => ({
        name: s.name,
        approval: !!s.approval,
        continueOnError: !!s.continueOnError,
        steps: (s.steps || []).map((t) => ({ ...t }))
      }))
      Object.assign(form, { ...row, stages: stages.length ? stages : [emptyStage()] })
    }
    dialogVisible.value = true
  }

  const submitForm = () => {
    formRef.value.validate(async (valid) => {
      if (!valid) return
      saving.value = true
      try {
        const payload = { ...form, stages: form.stages }
        const res = form.ID ? await updatePipeline(payload) : await createPipeline(payload)
        if (res.code === 0) {
          ElMessage.success(res.msg)
          dialogVisible.value = false
          getList()
        }
      } finally {
        saving.value = false
      }
    })
  }

  const onDelete = (row) => {
    ElMessageBox.confirm(`确定删除流水线「${row.name}」吗？阶段与步骤将一并删除。`, '提示', {
      confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning'
    }).then(async () => {
      const res = await deletePipeline({ id: row.ID })
      if (res.code === 0) {
        ElMessage.success('删除成功')
        getList()
      }
    })
  }

  onMounted(getList)
</script>
