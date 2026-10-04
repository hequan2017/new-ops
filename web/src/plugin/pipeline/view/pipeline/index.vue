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
        <el-table-column label="操作" width="190" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openDialog(row)">编排</el-button>
            <el-button link type="success" @click="openBuild(row)">构建</el-button>
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
          <el-col :span="12">
            <el-form-item label="Webhook">
              <div style="display: flex; align-items: center; gap: 8px; width: 100%">
                <el-switch v-model="form.webhookEnabled" />
                <el-input
                  v-if="form.webhookEnabled && form.webhookToken"
                  :model-value="form.webhookToken"
                  readonly
                  size="small"
                >
                  <template #append>
                    <el-button @click="copyWebhook(form.webhookToken)">复制地址</el-button>
                  </template>
                </el-input>
                <span v-else-if="form.webhookEnabled" class="ops-text-muted">保存后生成令牌</span>
              </div>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="定时触发">
              <div style="display: flex; align-items: center; gap: 8px; width: 100%">
                <el-switch v-model="form.cronEnabled" />
                <el-input
                  v-if="form.cronEnabled"
                  v-model="form.cronSpec"
                  placeholder="cron 表达式，如 0 9 * * *"
                  style="flex: 1"
                />
              </div>
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

    <el-drawer v-model="buildVisible" :title="`构建 · ${buildPipeline?.name || ''}`" size="60%">
      <el-form inline>
        <el-form-item label="参数">
          <div style="width: 420px">
            <div v-for="(kv, i) in buildParams" :key="i" style="display: flex; gap: 6px; margin-bottom: 6px">
              <el-input v-model="kv.key" placeholder="key" style="width: 140px" />
              <el-input v-model="kv.value" placeholder="value" style="flex: 1" />
              <el-button link type="danger" @click="buildParams.splice(i, 1)">删</el-button>
            </div>
            <el-button size="small" @click="buildParams.push({ key: '', value: '' })">加参数</el-button>
          </div>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="triggering" @click="submitStartBuild">触发构建</el-button>
        </el-form-item>
      </el-form>

      <el-table :data="buildList" v-loading="buildLoading" stripe size="small">
        <el-table-column prop="buildNo" label="#" width="55" />
        <el-table-column label="状态" width="95">
          <template #default="{ row }">
            <el-tag :type="buildTagType(row.status)" size="small">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="operator" label="触发人" width="90" />
        <el-table-column label="开始时间" width="150">
          <template #default="{ row }">{{ (row.startedAt || '—').replace('T', ' ').slice(0, 19) }}</template>
        </el-table-column>
        <el-table-column label="结束时间" width="150">
          <template #default="{ row }">{{ (row.finishedAt || '—').replace('T', ' ').slice(0, 19) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openLogs(row)">日志</el-button>
            <el-button v-if="row.status === '等待审批'" link type="warning" @click="doApprove(row)">放行</el-button>
            <el-button v-if="['等待中', '等待审批', '执行中'].includes(row.status)" link type="danger" @click="doCancel(row)">取消</el-button>
            <el-button v-if="['成功', '失败', '已取消'].includes(row.status)" link type="success" @click="doRestart(row)">重跑</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-drawer v-model="logsVisible" :title="`日志 · 构建 #${logsBuild?.buildNo || ''}`" size="55%" append-to-body>
        <div ref="logBox" class="build-log">
          <div v-for="l in logsList" :key="l.ID" :class="['log-line', 'lv-' + l.level]">
            <span class="log-prefix">[{{ l.stageName || '系统' }}{{ l.stepName ? '/' + l.stepName : '' }}]</span>
            {{ l.content }}
          </div>
          <el-empty v-if="!logsList.length" description="暂无日志" />
        </div>
      </el-drawer>
    </el-drawer>
  </div>
</template>

<script setup>
  import { ref, reactive, onMounted, onUnmounted, nextTick } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { useUserStore } from '@/pinia/modules/user'
  import {
    createPipeline, updatePipeline, deletePipeline, getPipelineList,
    startBuild, cancelBuild, approveBuild, getBuildList, getBuildLogs, restartBuild
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
  const emptyForm = () => ({ ID: 0, name: '', description: '', enabled: true, webhookEnabled: false, webhookToken: '', cronEnabled: false, cronSpec: '', stages: [emptyStage()] })
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

  // ---------- 构建抽屉 ----------
  const buildVisible = ref(false)
  const buildPipeline = ref(null)
  const buildParams = ref([])
  const triggering = ref(false)
  const buildList = ref([])
  const buildLoading = ref(false)
  const logsVisible = ref(false)
  const logsBuild = ref(null)
  const logsList = ref([])
  const logBox = ref(null)
  let buildTimer = null

  const openBuild = (row) => {
    buildPipeline.value = row
    buildParams.value = [{ key: '', value: '' }]
    buildVisible.value = true
    loadBuilds()
    if (!buildTimer) buildTimer = setInterval(pollBuilds, 3000)
  }

  const loadBuilds = async () => {
    if (!buildPipeline.value) return
    buildLoading.value = true
    try {
      const res = await getBuildList({ page: 1, pageSize: 20, pipelineId: buildPipeline.value.ID })
      if (res.code === 0) buildList.value = res.data.list || []
    } finally {
      buildLoading.value = false
    }
  }

  const pollBuilds = () => {
    if (!buildVisible.value || !buildPipeline.value) return
    getBuildList({ page: 1, pageSize: 20, pipelineId: buildPipeline.value.ID }).then((res) => {
      if (res.code === 0) {
        buildList.value = res.data.list || []
        // 日志抽屉打开且对应构建进行中时静默刷新日志
        if (logsVisible.value && logsBuild.value) {
          const cur = buildList.value.find((b) => b.ID === logsBuild.value.ID)
          if (cur && ['等待中', '等待审批', '执行中'].includes(cur.status)) refreshLogs()
        }
      }
    })
  }

  const submitStartBuild = async () => {
    const params = {}
    buildParams.value.forEach((kv) => {
      if (kv.key) params[kv.key] = kv.value
    })
    triggering.value = true
    try {
      const res = await startBuild({ pipelineId: buildPipeline.value.ID, params })
      if (res.code === 0) {
        ElMessage.success(`构建 #${res.data.buildNo} 已触发`)
        loadBuilds()
      }
    } finally {
      triggering.value = false
    }
  }

  const doApprove = async (row) => {
    const res = await approveBuild({ id: row.ID })
    if (res.code === 0) {
      ElMessage.success('已放行')
      loadBuilds()
    }
  }

  const doCancel = (row) => {
    ElMessageBox.confirm(`确定取消构建 #${row.buildNo} 吗？`, '提示', {
      confirmButtonText: '取消构建', cancelButtonText: '返回', type: 'warning'
    }).then(async () => {
      const res = await cancelBuild({ id: row.ID })
      if (res.code === 0) loadBuilds()
    })
  }

  const openLogs = async (row) => {
    logsBuild.value = row
    logsList.value = []
    logsVisible.value = true
    closeStream()
    if (['等待中', '等待审批', '执行中'].includes(row.status)) {
      startStream(row.ID) // 进行中构建走 SSE 实时流
    } else {
      await refreshLogs()
    }
  }

  // SSE 实时日志流（断线/异常降级为一次性拉取）
  let evtSource = null
  const closeStream = () => {
    if (evtSource) {
      evtSource.close()
      evtSource = null
    }
  }
  const startStream = (buildId) => {
    const proto = location.protocol === 'https:' ? 'https' : 'http'
    const base = import.meta.env.VITE_BASE_API || '/api'
    evtSource = new EventSource(
      `${proto}://${location.host}${base}/sse/pipeline/build/logs?token=${useUserStore().token}&id=${buildId}`
    )
    evtSource.addEventListener('log', (e) => {
      try {
        logsList.value.push(JSON.parse(e.data))
        nextTick(scrollLogBox)
      } catch (err) {
        /* 忽略畸形帧 */
      }
    })
    evtSource.addEventListener('status', (e) => {
      try {
        const d = JSON.parse(e.data)
        if (logsBuild.value) logsBuild.value.status = d.status
      } catch (err) {
        /* ignore */
      }
    })
    evtSource.addEventListener('done', () => {
      closeStream()
      loadBuilds()
    })
    evtSource.onerror = () => {
      closeStream()
      refreshLogs()
    }
  }

  const refreshLogs = async () => {
    const res = await getBuildLogs({ id: logsBuild.value.ID, page: 1, pageSize: 200 })
    if (res.code === 0) {
      logsList.value = res.data.list || []
      nextTick(scrollLogBox)
    }
  }

  const scrollLogBox = () => {
    if (logBox.value) logBox.value.scrollTop = logBox.value.scrollHeight
  }

  const buildTagType = (s) =>
    ({ 成功: 'success', 失败: 'danger', 已取消: 'info', 执行中: 'primary', 等待审批: 'warning' }[s] || 'info')

  const doRestart = async (row) => {
    const res = await restartBuild({ id: row.ID })
    if (res.code === 0) {
      ElMessage.success(`构建 #${res.data.buildNo} 已重跑`)
      loadBuilds()
    }
  }

  const copyWebhook = (token) => {
    const base = import.meta.env.VITE_BASE_API || '/api'
    const url = `${location.origin}${base}/pipeline/webhook/${token}`
    navigator.clipboard?.writeText(url)
    ElMessage.success('webhook 地址已复制')
  }

  onMounted(getList)
  onUnmounted(() => {
    if (buildTimer) clearInterval(buildTimer)
    closeStream()
  })
</script>

<style scoped>
  .build-log {
    height: calc(100vh - 160px);
    overflow: auto;
    background: #0b1021;
    color: #e2e8f0;
    border-radius: 6px;
    padding: 10px;
    font-family: 'JetBrains Mono', Consolas, monospace;
    font-size: 12px;
  }
  .log-line {
    white-space: pre-wrap;
    word-break: break-all;
    margin-bottom: 2px;
  }
  .log-prefix {
    color: #7dd3fc;
    margin-right: 6px;
  }
  .lv-stderr {
    color: #fca5a5;
  }
  .lv-system {
    color: #fbbf24;
  }
</style>
