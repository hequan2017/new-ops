<template>
  <div>
    <div class="ops-card">
      <el-form inline>
        <el-form-item>
          <el-button type="primary" @click="openEdit()">注册模型</el-button>
          <el-button @click="getModels().then(loadModels)">刷 新</el-button>
          <el-button @click="logsVisible = true; loadLogs()">调用日志</el-button>
        </el-form-item>
      </el-form>
      <el-table :data="models" v-loading="loading" stripe>
        <el-table-column prop="name" label="注册名" width="170" show-overflow-tooltip />
        <el-table-column prop="provider" label="供应商" width="120" show-overflow-tooltip />
        <el-table-column prop="baseUrl" label="API 地址" min-width="200" show-overflow-tooltip />
        <el-table-column label="能力" width="180">
          <template #default="{ row }">
            <el-tag v-for="c in (row.capability || '').split(',')" :key="c" size="small" style="margin-right: 4px">{{ capLabel(c) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="默认调度" width="90">
          <template #default="{ row }">
            <el-tag v-if="row.defaultFor" type="success" size="small">{{ capLabel(row.defaultFor) }}</el-tag>
            <span v-else>—</span>
          </template>
        </el-table-column>
        <el-table-column prop="ctxLen" label="上下文" width="90" />
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag size="small" :type="row.status === 'enabled' ? 'success' : 'info'">{{ row.status === 'enabled' ? '启用' : '停用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="170" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
            <el-button link type="warning" @click="openTest(row)">测试</el-button>
            <el-button link type="danger" @click="onDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <div class="ops-card" style="margin-top: 12px">
      <h4 style="margin: 0 0 10px">模型服务一键部署（经容器接入点）</h4>
      <el-table :data="deploys" v-loading="deploysLoading" stripe size="small">
        <el-table-column prop="name" label="部署名" width="180" show-overflow-tooltip />
        <el-table-column prop="image" label="镜像" min-width="200" show-overflow-tooltip />
        <el-table-column prop="ports" label="端口" width="140" show-overflow-tooltip />
        <el-table-column prop="baseUrl" label="服务地址" min-width="160" show-overflow-tooltip />
        <el-table-column label="状态" width="110">
          <template #default="{ row }">
            <el-tag size="small" :type="deployTag(row.status)">{{ deployLabel(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="errMsg" label="最近信息" min-width="140" show-overflow-tooltip />
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openDeployEdit(row)">编辑</el-button>
            <el-button link type="success" :loading="row._execing" @click="doExecute(row)">部署</el-button>
            <el-button link type="warning" @click="doDestroy(row)">销毁</el-button>
            <el-button link type="danger" @click="onDeleteDeploy(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 注册/编辑 -->
    <el-dialog v-model="editVisible" :title="form.id ? '编辑模型' : '注册模型'" width="600px">
      <el-form label-width="110px">
        <el-form-item label="注册名">
          <el-input v-model="form.name" placeholder="如 deepseek-chat" />
        </el-form-item>
        <el-form-item label="供应商">
          <el-input v-model="form.provider" placeholder="如 deepseek / 本地 vLLM" />
        </el-form-item>
        <el-form-item label="API 地址">
          <el-input v-model="form.baseUrl" placeholder="如 http://127.0.0.1:8000/v1（含 /v1）" />
        </el-form-item>
        <el-form-item label="API Key">
          <el-input v-model="form.apiKey" type="password" show-password placeholder="留空表示不修改" />
        </el-form-item>
        <el-form-item label="能力标签">
          <el-checkbox-group v-model="form.caps">
            <el-checkbox value="chat">对话</el-checkbox>
            <el-checkbox value="embed">向量</el-checkbox>
            <el-checkbox value="rerank">重排</el-checkbox>
          </el-checkbox-group>
        </el-form-item>
        <el-form-item label="默认调度">
          <el-select v-model="form.defaultFor" clearable placeholder="不担任默认" style="width: 200px">
            <el-option v-for="c in form.caps" :key="c" :label="capLabel(c)" :value="c" />
          </el-select>
        </el-form-item>
        <el-form-item label="上下文长度">
          <el-input-number v-model="form.ctxLen" :min="512" :step="1024" />
        </el-form-item>
        <el-form-item label="启用">
          <el-switch v-model="form.enabled" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.notes" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editVisible = false">取 消</el-button>
        <el-button type="primary" :loading="saving" @click="doSave">保 存</el-button>
      </template>
    </el-dialog>

    <!-- 测试台 -->
    <el-dialog v-model="testVisible" :title="`在线测试 · ${testRow?.name || ''}`" width="640px">
      <el-radio-group v-model="testProtocol" style="margin-bottom: 10px">
        <el-radio-button value="chat">对话</el-radio-button>
        <el-radio-button value="embed">向量</el-radio-button>
        <el-radio-button value="rerank">重排</el-radio-button>
      </el-radio-group>
      <el-input v-if="testProtocol === 'chat'" v-model="testPrompt" type="textarea" :rows="3" placeholder="输入测试问题" />
      <template v-else-if="testProtocol === 'embed'">
        <el-input v-model="testInput" type="textarea" :rows="3" placeholder="每行一条输入文本" />
      </template>
      <template v-else>
        <el-input v-model="testQuery" placeholder="查询语句" style="margin-bottom: 8px" />
        <el-input v-model="testDocs" type="textarea" :rows="4" placeholder="每行一条候选文档" />
      </template>
      <div style="margin: 10px 0">
        <el-button type="primary" :loading="testLoading" @click="doTest">运 行 测 试</el-button>
      </div>
      <pre v-if="testResult" class="ops-pre">{{ testResult }}</pre>
    </el-dialog>

    <!-- 部署编辑 -->
    <el-dialog v-model="deployVisible" :title="deployForm.id ? '编辑部署' : '新建部署'" width="560px">
      <el-form label-width="110px">
        <el-form-item label="部署名">
          <el-input v-model="deployForm.name" placeholder="如 qwen-chat" />
        </el-form-item>
        <el-form-item label="接入点ID">
          <el-input-number v-model="deployForm.endpointId" :min="0" placeholder="容器接入点" />
        </el-form-item>
        <el-form-item label="镜像">
          <el-input v-model="deployForm.image" placeholder="如 vllm/vllm-openai:latest" />
        </el-form-item>
        <el-form-item label="端口映射">
          <el-input v-model="deployForm.ports" placeholder="8000:8000, 多个逗号分隔" />
        </el-form-item>
        <el-form-item label="环境变量">
          <el-input v-model="deployForm.envs" placeholder="K=V, 多个逗号分隔" />
        </el-form-item>
        <el-form-item label="服务地址">
          <el-input v-model="deployForm.baseUrl" placeholder="http://宿主IP:端口/v1（健康探测）" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="deployForm.notes" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="deployVisible = false">取 消</el-button>
        <el-button type="primary" :loading="deploySaving" @click="doSaveDeploy">保 存</el-button>
      </template>
    </el-dialog>

    <!-- 调用日志 -->
    <el-dialog v-model="logsVisible" title="模型调用日志" width="900px">
      <el-table :data="logs" v-loading="logsLoading" stripe size="small">
        <el-table-column prop="modelName" label="模型" width="140" show-overflow-tooltip />
        <el-table-column prop="protocol" label="协议" width="80" />
        <el-table-column label="时延" width="90">
          <template #default="{ row }">{{ row.latencyMs }} ms</template>
        </el-table-column>
        <el-table-column prop="totalTokens" label="tokens" width="80" />
        <el-table-column label="结果" width="70">
          <template #default="{ row }">
            <el-tag size="small" :type="row.ok ? 'success' : 'danger'">{{ row.ok ? '成功' : '失败' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="error" label="错误" min-width="180" show-overflow-tooltip />
        <el-table-column label="时间" width="170">
          <template #default="{ row }">{{ fmtTime(row.CreatedAt) }}</template>
        </el-table-column>
      </el-table>
    </el-dialog>
  </div>
</template>

<script setup>
  import { ref, reactive, onMounted } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import {
    getModels, saveModel, deleteModel, testModel, getModelCallLogs,
    getDeploys, saveDeployPlan, executeDeploy, destroyDeploy, deleteDeploy,
  } from '@/plugin/llmops/api/model'

  const capNames = { chat: '对话', embed: '向量', rerank: '重排' }
  const capLabel = (c) => capNames[c] || c
  const deployNames = { pending: '待部署', deploying: '部署中', running: '运行中', stopped: '已停止', failed: '失败' }
  const deployLabel = (s) => deployNames[s] || s
  const deployTag = (s) => ({ pending: 'info', deploying: 'warning', running: 'success', stopped: 'info', failed: 'danger' }[s] || 'info')
  const fmtTime = (t) => (t ? String(t).replace('T', ' ').slice(0, 19) : '—')

  const models = ref([])
  const loading = ref(false)
  const loadModels = async (res) => {
    loading.value = true
    try {
      const r = res || (await getModels())
      models.value = r.data || []
    } finally {
      loading.value = false
    }
  }

  const editVisible = ref(false)
  const saving = ref(false)
  const form = reactive({ id: 0, name: '', provider: '', baseUrl: '', apiKey: '', caps: [], defaultFor: '', ctxLen: 8192, enabled: true, notes: '' })
  const openEdit = (row) => {
    Object.assign(form, {
      id: row?.ID || 0, name: row?.name || '', provider: row?.provider || '', baseUrl: row?.baseUrl || '',
      apiKey: '', caps: (row?.capability || '').split(',').filter(Boolean), defaultFor: row?.defaultFor || '',
      ctxLen: row?.ctxLen || 8192, enabled: row?.status !== 'disabled', notes: row?.notes || '',
    })
    editVisible.value = true
  }
  const doSave = async () => {
    saving.value = true
    try {
      await saveModel({
        id: form.id, name: form.name, provider: form.provider, baseUrl: form.baseUrl, apiKey: form.apiKey,
        capability: form.caps.join(','), ctxLen: form.ctxLen, defaultFor: form.defaultFor,
        status: form.enabled ? 'enabled' : 'disabled', notes: form.notes,
      })
      ElMessage.success('保存成功')
      editVisible.value = false
      loadModels()
    } finally {
      saving.value = false
    }
  }
  const onDelete = (row) => {
    ElMessageBox.confirm(`确认删除模型「${row.name}」？`, '提示', { type: 'warning' })
      .then(async () => {
        await deleteModel({ id: row.ID })
        ElMessage.success('删除成功')
        loadModels()
      })
      .catch(() => {})
  }

  // 测试台
  const testVisible = ref(false)
  const testRow = ref(null)
  const testProtocol = ref('chat')
  const testPrompt = ref('你好，请介绍一下你自己')
  const testInput = ref('白泽统一运维开发平台')
  const testQuery = ref('什么是大模型微调')
  const testDocs = ref('SFT 是有监督微调\nRLHF 是人类反馈强化学习\n向量检索用于语义匹配')
  const testLoading = ref(false)
  const testResult = ref('')
  const openTest = (row) => {
    testRow.value = row
    testResult.value = ''
    if ((row.capability || '').includes('embed') && !(row.capability || '').includes('chat')) testProtocol.value = 'embed'
    else if ((row.capability || '').includes('rerank') && !(row.capability || '').includes('chat')) testProtocol.value = 'rerank'
    else testProtocol.value = 'chat'
    testVisible.value = true
  }
  const doTest = async () => {
    const payload = { modelId: testRow.value.ID, protocol: testProtocol.value }
    if (testProtocol.value === 'chat') payload.messages = [{ role: 'user', content: testPrompt.value }]
    else if (testProtocol.value === 'embed') payload.input = testInput.value.split('\n').filter(Boolean)
    else {
      payload.query = testQuery.value
      payload.documents = testDocs.value.split('\n').filter(Boolean)
    }
    testLoading.value = true
    try {
      const res = await testModel(payload)
      testResult.value = JSON.stringify(res.data, null, 2)
    } finally {
      testLoading.value = false
    }
  }

  // 调用日志
  const logsVisible = ref(false)
  const logs = ref([])
  const logsLoading = ref(false)
  const loadLogs = async () => {
    logsLoading.value = true
    try {
      const res = await getModelCallLogs({ page: 1, pageSize: 50 })
      logs.value = res.data?.list || []
    } finally {
      logsLoading.value = false
    }
  }

  // 部署
  const deploys = ref([])
  const deploysLoading = ref(false)
  const deployVisible = ref(false)
  const deploySaving = ref(false)
  const deployForm = reactive({ id: 0, name: '', endpointId: 0, image: '', ports: '', envs: '', baseUrl: '', notes: '' })
  const loadDeploys = async () => {
    deploysLoading.value = true
    try {
      const res = await getDeploys()
      deploys.value = (res.data || []).map((d) => ({ ...d, _execing: false }))
    } finally {
      deploysLoading.value = false
    }
  }
  const openDeployEdit = (row) => {
    Object.assign(deployForm, {
      id: row?.ID || 0, name: row?.name || '', endpointId: row?.endpointId || 0, image: row?.image || '',
      ports: row?.ports || '', envs: row?.envs || '', baseUrl: row?.baseUrl || '', notes: row?.notes || '',
    })
    deployVisible.value = true
  }
  const doSaveDeploy = async () => {
    deploySaving.value = true
    try {
      await saveDeployPlan(deployForm)
      ElMessage.success('保存成功')
      deployVisible.value = false
      loadDeploys()
    } finally {
      deploySaving.value = false
    }
  }
  const doExecute = async (row) => {
    row._execing = true
    try {
      const res = await executeDeploy(row.ID)
      ElMessage.success(res.data?.status === 'running' ? '部署完成，服务健康' : '容器已创建，服务探测未过（保持部署中）')
      loadDeploys()
    } finally {
      row._execing = false
    }
  }
  const doDestroy = (row) => {
    ElMessageBox.confirm(`确认销毁部署「${row.name}」的容器？`, '提示', { type: 'warning' })
      .then(async () => {
        await destroyDeploy(row.ID)
        ElMessage.success('已销毁')
        loadDeploys()
      })
      .catch(() => {})
  }
  const onDeleteDeploy = (row) => {
    ElMessageBox.confirm(`确认删除部署计划「${row.name}」？`, '提示', { type: 'warning' })
      .then(async () => {
        await deleteDeploy({ id: row.ID })
        ElMessage.success('删除成功')
        loadDeploys()
      })
      .catch(() => {})
  }

  onMounted(() => {
    loadModels()
    loadDeploys()
  })
</script>
