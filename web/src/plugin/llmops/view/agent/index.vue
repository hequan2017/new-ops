<template>
  <div style="display: flex; gap: 12px; align-items: flex-start">
    <!-- 左：智能体列表 -->
    <div class="ops-card" style="width: 380px; flex-shrink: 0">
      <el-form inline>
        <el-form-item>
          <el-button type="primary" size="small" @click="openEdit()">新建智能体</el-button>
          <el-button size="small" @click="loadAgents">刷 新</el-button>
        </el-form-item>
      </el-form>
      <el-table :data="agents" v-loading="loading" stripe size="small" highlight-current-row @current-change="selectAgent">
        <el-table-column prop="name" label="智能体" min-width="110" show-overflow-tooltip />
        <el-table-column label="模式" width="80">
          <template #default="{ row }">
            <el-tag size="small">{{ modeLabel(row.mode) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="阶段" width="76">
          <template #default="{ row }">
            <el-tag size="small" :type="stageTag(row.status)">{{ stageLabel(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="170" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click.stop="openEdit(row)">编辑</el-button>
            <el-button link type="success" size="small" @click="doPublish(row)">发布</el-button>
            <el-button link type="warning" size="small" @click.stop="openVersions(row)">版本</el-button>
            <el-button link type="danger" size="small" @click.stop="onDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 监控区 -->
      <template v-if="monitor">
        <h4 style="margin: 12px 0 6px">监控（近 {{ monitor.hours }}h）</h4>
        <el-descriptions :column="2" size="small" border>
          <el-descriptions-item label="调用">{{ monitor.stat.total }}</el-descriptions-item>
          <el-descriptions-item label="成功/失败">{{ monitor.stat.okCount }} / {{ monitor.stat.failCount }}</el-descriptions-item>
          <el-descriptions-item label="平均时延">{{ monitor.stat.avgLatencyMs }} ms</el-descriptions-item>
          <el-descriptions-item label="P95">{{ monitor.stat.p95LatencyMs }} ms</el-descriptions-item>
        </el-descriptions>
        <div style="margin-top: 6px; font-size: 12px; color: #888">
          时延分布 &lt;0.5s: {{ monitor.buckets[0] }} · 0.5-2s: {{ monitor.buckets[1] }} · 2-10s: {{ monitor.buckets[2] }} · &gt;10s: {{ monitor.buckets[3] }}
        </div>
        <el-button link type="danger" size="small" style="margin-top: 4px" @click="openAbnormal">异常会话({{ monitor.abnormalCount }})</el-button>
      </template>
    </div>

    <!-- 右：调试会话窗 -->
    <div class="ops-card" style="flex: 1; min-width: 0">
      <el-form inline>
        <el-form-item>
          <el-select v-model="sessionId" placeholder="选择会话" style="width: 220px" @change="loadMessages">
            <el-option v-for="s in sessions" :key="s.ID" :label="s.title" :value="s.ID" />
          </el-select>
          <el-button size="small" style="margin-left: 8px" @click="newSession">新建会话</el-button>
          <el-button size="small" @click="openMultiRun" :disabled="!currentAgent || currentAgent.mode !== 'multi'">多智能体运行</el-button>
        </el-form-item>
      </el-form>
      <div ref="chatBox" style="height: 380px; overflow-y: auto; border: 1px solid #eee; border-radius: 6px; padding: 10px">
        <div v-for="m in messages" :key="m.ID" :style="{ textAlign: m.role === 'user' ? 'right' : 'left', margin: '6px 0' }">
          <el-tag size="small" :type="m.role === 'user' ? 'primary' : 'success'" style="margin-right: 6px">{{ m.role === 'user' ? '我' : '智能体' }}</el-tag>
          <span style="white-space: pre-wrap">{{ m.content }}</span>
          <div v-if="m.planJSON" style="font-size: 12px; color: #e6a23c; margin-top: 2px">
            意图：{{ parseJSON(m.planJSON)?.intent }} · 步骤：{{ (parseJSON(m.planJSON)?.steps || []).length }}
          </div>
          <div v-if="m.traceJSON && (parseJSON(m.traceJSON) || []).length" style="font-size: 12px; color: #909399; margin-top: 2px">
            工具轨迹：
            <div v-for="(t, i) in parseJSON(m.traceJSON)" :key="i">
              #{{ t.turn }} {{ t.tool }} {{ t.err ? '✗ ' + t.err : '✓' }}
            </div>
          </div>
        </div>
        <div v-if="!messages.length" style="color: #999; text-align: center; padding-top: 40px">发送消息开始调测</div>
      </div>
      <div style="margin-top: 10px; display: flex; gap: 8px">
        <el-input v-model="draft" placeholder="输入消息（Enter 发送）" @keyup.enter="send" :disabled="!sessionId" />
        <el-button type="primary" :loading="sending" @click="send" :disabled="!sessionId">发 送</el-button>
      </div>
    </div>

    <!-- 编辑 -->
    <el-dialog v-model="editVisible" :title="form.id ? '编辑智能体' : '新建智能体'" width="640px">
      <el-form label-width="110px">
        <el-form-item label="名称"><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="构建模式">
          <el-select v-model="form.mode" style="width: 240px">
            <el-option label="问答助手" value="qa" />
            <el-option label="交互应用" value="interactive" />
            <el-option label="工作流" value="workflow" />
            <el-option label="多智能体" value="multi" />
            <el-option label="Agentic" value="agentic" />
          </el-select>
        </el-form-item>
        <el-form-item label="系统提示词"><el-input v-model="form.sysPrompt" type="textarea" :rows="3" /></el-form-item>
        <el-form-item label="主模型ID"><el-input-number v-model="form.modelId" :min="1" /></el-form-item>
        <el-form-item label="知识库ID"><el-input v-model="form.knowledgeIds" placeholder="逗号分隔，可空" /></el-form-item>
        <el-form-item label="轮次上限"><el-input-number v-model="form.maxTurns" :min="1" :max="20" /></el-form-item>
        <el-form-item label="配置 JSON">
          <el-input v-model="form.configJSON" type="textarea" :rows="5"
            :placeholder="form.mode === 'multi' ? '{&quot;pattern&quot;:&quot;converge&quot;,&quot;subs&quot;:[{&quot;agentId&quot;:2,&quot;role&quot;:&quot;分析&quot;,&quot;prompt&quot;:&quot;分析 {input}&quot;}]}' : form.mode === 'workflow' ? '{&quot;nodes&quot;:[{&quot;id&quot;:&quot;a&quot;,&quot;type&quot;:&quot;llm&quot;,&quot;prompt&quot;:&quot;处理 {input}&quot;}]}' : '{}'" />
        </el-form-item>
        <el-form-item label="描述"><el-input v-model="form.desc" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editVisible = false">取 消</el-button>
        <el-button type="primary" :loading="saving" @click="doSave">保 存</el-button>
      </template>
    </el-dialog>

    <!-- 版本 -->
    <el-dialog v-model="verVisible" :title="`版本快照 · ${verRow?.name || ''}`" width="640px">
      <el-table :data="versions" stripe size="small">
        <el-table-column prop="version" label="版本" width="70" />
        <el-table-column prop="remark" label="说明" min-width="180" show-overflow-tooltip />
        <el-table-column label="操作" width="90">
          <template #default="{ row }">
            <el-button link type="warning" size="small" @click="doRollback(row)">回滚</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>

    <!-- 多智能体运行 -->
    <el-dialog v-model="multiVisible" title="多智能体运行" width="620px">
      <el-input v-model="multiInput" type="textarea" :rows="3" placeholder="输入任务" />
      <pre v-if="multiResult" class="ops-pre" style="margin-top: 10px">{{ multiResult }}</pre>
      <template #footer>
        <el-button @click="multiVisible = false">关 闭</el-button>
        <el-button type="primary" :loading="multiLoading" @click="doMultiRun">运 行</el-button>
      </template>
    </el-dialog>

    <!-- 异常会话 -->
    <el-dialog v-model="abnormalVisible" title="异常会话" width="720px">
      <el-table :data="abnormal" stripe size="small">
        <el-table-column prop="sessionId" label="会话ID" width="80" />
        <el-table-column prop="title" label="标题" min-width="160" show-overflow-tooltip />
        <el-table-column prop="failCount" label="失败次数" width="90" />
        <el-table-column prop="lastErr" label="最近错误" min-width="220" show-overflow-tooltip />
      </el-table>
    </el-dialog>
  </div>
</template>

<script setup>
  import { ref, reactive, onMounted, nextTick } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import {
    getAgents, saveAgent, deleteAgent, setAgentStage, publishAgent, getAgentVersions, rollbackAgent,
    createSession, getSessions, getMessages, debugChat,
    multiRun, getAgentMonitor, getAbnormalSessions,
  } from '@/plugin/llmops/api/agent'

  const modeNames = { qa: '问答', interactive: '交互', workflow: '工作流', multi: '多智能体', agentic: 'Agentic' }
  const modeLabel = (m) => modeNames[m] || m
  const stageNames = { draft: '创建', debug: '调测', eval: '评测', published: '发布', operating: '运营', monitor: '监控' }
  const stageLabel = (s) => stageNames[s] || s
  const stageTag = (s) => ({ draft: 'info', debug: 'warning', eval: '', published: 'success', operating: 'success', monitor: 'primary' }[s] || 'info')
  const parseJSON = (s) => {
    try { return JSON.parse(s) } catch (e) { return null }
  }

  const agents = ref([])
  const loading = ref(false)
  const currentAgent = ref(null)
  const loadAgents = async () => {
    loading.value = true
    try {
      const res = await getAgents()
      agents.value = res.data || []
      loadMonitor()
    } finally {
      loading.value = false
    }
  }
  const selectAgent = async (row) => {
    currentAgent.value = row
    sessions.value = []
    sessionId.value = null
    messages.value = []
    if (row) await loadSessions()
  }

  // 编辑
  const editVisible = ref(false)
  const saving = ref(false)
  const form = reactive({ id: 0, name: '', mode: 'qa', sysPrompt: '', modelId: 1, knowledgeIds: '', maxTurns: 5, configJSON: '', desc: '' })
  const openEdit = (row) => {
    Object.assign(form, {
      id: row?.ID || 0, name: row?.name || '', mode: row?.mode || 'qa', sysPrompt: row?.sysPrompt || '',
      modelId: row?.modelId || 1, knowledgeIds: row?.knowledgeIds || '', maxTurns: row?.maxTurns || 5,
      configJSON: row?.configJSON || '', desc: row?.desc || '',
    })
    editVisible.value = true
  }
  const doSave = async () => {
    saving.value = true
    try {
      await saveAgent(form)
      ElMessage.success('保存成功')
      editVisible.value = false
      loadAgents()
    } finally {
      saving.value = false
    }
  }
  const onDelete = (row) => {
    ElMessageBox.confirm(`确认删除智能体「${row.name}」？`, '提示', { type: 'warning' })
      .then(async () => {
        await deleteAgent({ id: row.ID })
        ElMessage.success('删除成功')
        loadAgents()
      })
      .catch(() => {})
  }
  const doPublish = (row) => {
    ElMessageBox.prompt('发布说明', `发布 ${row.name}`, { inputValue: '首次发布' })
      .then(async ({ value }) => {
        const res = await publishAgent({ id: row.ID, remark: value || '' })
        ElMessage.success(`已发布 v${res.data?.version}`)
        loadAgents()
      })
      .catch(() => {})
  }
  const verVisible = ref(false)
  const verRow = ref(null)
  const versions = ref([])
  const openVersions = async (row) => {
    verRow.value = row
    const res = await getAgentVersions({ agentId: row.ID })
    versions.value = res.data || []
    verVisible.value = true
  }
  const doRollback = (row) => {
    ElMessageBox.confirm(`确认回滚到 v${row.version}？`, '提示', { type: 'warning' })
      .then(async () => {
        await rollbackAgent({ agentId: verRow.value.ID, version: row.version })
        ElMessage.success('已回滚')
        verVisible.value = false
        loadAgents()
      })
      .catch(() => {})
  }

  // 会话与调试
  const sessions = ref([])
  const sessionId = ref(null)
  const messages = ref([])
  const loadSessions = async () => {
    const res = await getSessions({ agentId: currentAgent.value.ID, page: 1, pageSize: 50 })
    sessions.value = res.data?.list || []
  }
  const newSession = async () => {
    if (!currentAgent.value) {
      ElMessage.warning('请先选择智能体')
      return
    }
    const res = await createSession({ agentId: currentAgent.value.ID })
    sessionId.value = res.data?.ID
    messages.value = []
    loadSessions()
  }
  const loadMessages = async () => {
    if (!sessionId.value) return
    const res = await getMessages({ sessionId: sessionId.value })
    messages.value = res.data || []
    scrollBottom()
  }
  const draft = ref('')
  const sending = ref(false)
  const chatBox = ref(null)
  const scrollBottom = () => nextTick(() => {
    if (chatBox.value) chatBox.value.scrollTop = chatBox.value.scrollHeight
  })
  const send = async () => {
    if (!draft.value || !sessionId.value) return
    const msg = draft.value
    draft.value = ''
    messages.value.push({ role: 'user', content: msg })
    scrollBottom()
    sending.value = true
    try {
      const res = await debugChat({ sessionId: sessionId.value, message: msg })
      messages.value.push({
        role: 'assistant', content: res.data?.reply,
        planJSON: res.data?.plan ? JSON.stringify(res.data.plan) : '',
        traceJSON: res.data?.traces?.length ? JSON.stringify(res.data.traces) : '',
      })
      scrollBottom()
    } finally {
      sending.value = false
    }
  }

  // 多智能体运行
  const multiVisible = ref(false)
  const multiLoading = ref(false)
  const multiInput = ref('')
  const multiResult = ref('')
  const openMultiRun = () => {
    multiResult.value = ''
    multiVisible.value = true
  }
  const doMultiRun = async () => {
    if (!multiInput.value || !currentAgent.value) return
    multiLoading.value = true
    try {
      const res = await multiRun({ agentId: currentAgent.value.ID, input: multiInput.value })
      multiResult.value = JSON.stringify(res.data, null, 2)
    } finally {
      multiLoading.value = false
    }
  }

  // 监控
  const monitor = ref(null)
  const loadMonitor = async () => {
    if (!agents.value.length) {
      monitor.value = null
      return
    }
    try {
      const res = await getAgentMonitor({ agentId: currentAgent.value?.ID || agents.value[0].ID, hours: 24 })
      monitor.value = res.data
    } catch (e) {
      monitor.value = null
    }
  }
  const abnormalVisible = ref(false)
  const abnormal = ref([])
  const openAbnormal = async () => {
    if (!currentAgent.value) return
    const res = await getAbnormalSessions({ agentId: currentAgent.value.ID, hours: 24 })
    abnormal.value = res.data?.list || []
    abnormalVisible.value = true
  }

  onMounted(loadAgents)
  void setAgentStage
</script>
