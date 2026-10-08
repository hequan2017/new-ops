<template>
  <div>
    <div class="ops-card">
      <el-form inline>
        <el-form-item>
          <el-button type="primary" @click="openEdit()">新增供应商</el-button>
          <el-button @click="getList">刷 新</el-button>
        </el-form-item>
      </el-form>
      <el-table :data="tableData" v-loading="loading" stripe>
        <el-table-column prop="name" label="名称" width="160" />
        <el-table-column prop="baseUrl" label="API 地址" min-width="220" show-overflow-tooltip />
        <el-table-column prop="model" label="模型" width="180" show-overflow-tooltip />
        <el-table-column label="默认" width="80">
          <template #default="{ row }">
            <el-tag v-if="row.isDefault" type="success" size="small">默认</el-tag>
            <span v-else>—</span>
          </template>
        </el-table-column>
        <el-table-column prop="timeoutSec" label="超时(s)" width="80" />
        <el-table-column prop="notes" label="备注" min-width="140" show-overflow-tooltip />
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
            <el-button link type="danger" @click="onDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <div class="ops-card" style="margin-top: 12px">
      <h4 style="margin: 0 0 10px">网关对话测试（OpenAI 兼容 chat/completions）</h4>
      <el-input v-model="chatPrompt" type="textarea" :rows="3" placeholder="输入测试问题，经默认供应商调用" />
      <div style="margin: 10px 0">
        <el-button type="primary" :loading="chatLoading" @click="doChat">发 送</el-button>
      </div>
      <pre v-if="chatReply" class="ops-pre">{{ chatReply }}</pre>
    </div>

    <el-dialog v-model="editVisible" :title="form.ID ? '编辑供应商' : '新增供应商'" width="560px">
      <el-form label-width="110px">
        <el-form-item label="名称">
          <el-input v-model="form.name" placeholder="如 mock-openai / deepseek" />
        </el-form-item>
        <el-form-item label="API 地址">
          <el-input v-model="form.baseUrl" placeholder="如 http://127.0.0.1:8890/v1（含 /v1）" />
        </el-form-item>
        <el-form-item label="API Key">
          <el-input v-model="form.apiKey" type="password" show-password placeholder="留空表示不修改" />
        </el-form-item>
        <el-form-item label="模型名">
          <el-input v-model="form.model" placeholder="如 gpt-4o-mini / deepseek-chat" />
        </el-form-item>
        <el-form-item label="默认供应商">
          <el-switch v-model="form.isDefault" />
        </el-form-item>
        <el-form-item label="超时(秒)">
          <el-input-number v-model="form.timeoutSec" :min="5" :max="600" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.notes" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editVisible = false">取 消</el-button>
        <el-button type="primary" @click="submitEdit">保 存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
  import { ref, reactive, onMounted } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { getProviderList, saveProvider, deleteProvider, llmChat } from '@/plugin/aiops/api/aiops'

  defineOptions({ name: 'aiopsLlm' })

  const tableData = ref([])
  const loading = ref(false)

  const getList = async () => {
    loading.value = true
    try {
      const res = await getProviderList()
      if (res.code === 0) tableData.value = res.data || []
    } finally {
      loading.value = false
    }
  }

  const editVisible = ref(false)
  const form = reactive({ ID: 0, name: '', baseUrl: '', apiKey: '', model: '', isDefault: false, timeoutSec: 60, notes: '' })

  const openEdit = (row) => {
    Object.assign(form, {
      ID: row?.ID || 0, name: row?.name || '', baseUrl: row?.baseUrl || '',
      apiKey: '', model: row?.model || '', isDefault: !!row?.isDefault,
      timeoutSec: row?.timeoutSec || 60, notes: row?.notes || ''
    })
    editVisible.value = true
  }

  const submitEdit = async () => {
    if (!form.name.trim() || !form.baseUrl.trim() || !form.model.trim()) {
      ElMessage.warning('名称/API 地址/模型名必填')
      return
    }
    const res = await saveProvider({ ...form })
    if (res.code === 0) {
      ElMessage.success('已保存')
      editVisible.value = false
      getList()
    }
  }

  const onDelete = (row) => {
    ElMessageBox.confirm(`确定删除供应商「${row.name}」吗？`, '提示', {
      confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning'
    }).then(async () => {
      const res = await deleteProvider({ id: row.ID })
      if (res.code === 0) {
        ElMessage.success('已删除')
        getList()
      }
    })
  }

  const chatPrompt = ref('')
  const chatReply = ref('')
  const chatLoading = ref(false)
  const doChat = async () => {
    if (!chatPrompt.value.trim()) {
      ElMessage.warning('请输入测试问题')
      return
    }
    chatLoading.value = true
    chatReply.value = ''
    try {
      const res = await llmChat({ prompt: chatPrompt.value })
      if (res.code === 0) {
        chatReply.value = `${res.data.content}\n\n—— ${res.data.provider} / ${res.data.model}`
      } else {
        chatReply.value = `调用失败：${res.msg}`
      }
    } finally {
      chatLoading.value = false
    }
  }

  onMounted(getList)
</script>

<style scoped>
  .ops-pre {
    margin: 0;
    max-height: 360px;
    overflow: auto;
    white-space: pre-wrap;
    word-break: break-word;
    font-size: 12px;
    background: var(--el-fill-color-light);
    padding: 8px;
    border-radius: 4px;
  }
</style>
