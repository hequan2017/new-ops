<template>
  <div>
    <!-- 训练任务 -->
    <div class="ops-card">
      <el-form inline>
        <el-form-item>
          <el-button type="primary" @click="openTrainEdit()">新建训练任务</el-button>
          <el-button @click="loadJobs">刷 新</el-button>
        </el-form-item>
        <el-form-item label="类型">
          <el-select v-model="jobQuery.type" clearable placeholder="全部" style="width: 130px" @change="loadJobs">
            <el-option v-for="t in trainTypes" :key="t" :label="t" :value="t" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="jobQuery.status" clearable placeholder="全部" style="width: 120px" @change="loadJobs">
            <el-option v-for="s in ['draft','queued','running','succeeded','failed','stopped']" :key="s" :label="s" :value="s" />
          </el-select>
        </el-form-item>
      </el-form>
      <el-table :data="jobs" v-loading="jobLoading" stripe size="small">
        <el-table-column prop="name" label="任务名" width="170" show-overflow-tooltip />
        <el-table-column prop="type" label="类型" width="100">
          <template #default="{ row }"><el-tag size="small">{{ row.type }}</el-tag></template>
        </el-table-column>
        <el-table-column prop="baseModel" label="基座" min-width="140" show-overflow-tooltip />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag size="small" :type="statusTag(row.status)">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="errMsg" label="信息" min-width="140" show-overflow-tooltip />
        <el-table-column label="操作" width="290" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="openTrainEdit(row)">编辑</el-button>
            <el-button link type="info" size="small" @click="doDryRun(row)">dry-run</el-button>
            <el-button link type="success" size="small" @click="doStart(row)">启动</el-button>
            <el-button link type="warning" size="small" @click="doStop(row)">停止</el-button>
            <el-button link type="primary" size="small" @click="openLog(row)">日志</el-button>
            <el-button link type="danger" size="small" @click="onDeleteJob(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 模型评估 -->
    <div class="ops-card" style="margin-top: 12px">
      <el-form inline>
        <el-form-item><h4 style="margin: 0">模型评估</h4></el-form-item>
        <el-form-item>
          <el-button type="primary" size="small" @click="evalVisible = true">新建评估</el-button>
          <el-button size="small" @click="loadEvals">刷 新</el-button>
        </el-form-item>
      </el-form>
      <el-table :data="evals" v-loading="evalLoading" stripe size="small">
        <el-table-column prop="name" label="评估名" width="170" show-overflow-tooltip />
        <el-table-column prop="modelId" label="被测模型ID" width="110" />
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag size="small" :type="statusTag(row.status)">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="通过/样本" width="100">
          <template #default="{ row }">{{ row.passCount }} / {{ row.sampleCount }}</template>
        </el-table-column>
        <el-table-column prop="errMsg" label="信息" min-width="140" show-overflow-tooltip />
        <el-table-column label="操作" width="130" fixed="right">
          <template #default="{ row }">
            <el-button link type="success" size="small" :loading="row._running" @click="doRunEval(row)">执行</el-button>
            <el-button link type="primary" size="small" @click="openEvalResults(row)">明细</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 产物 -->
    <div class="ops-card" style="margin-top: 12px">
      <el-form inline>
        <el-form-item><h4 style="margin: 0">训练/压缩产物</h4></el-form-item>
        <el-form-item>
          <el-button type="primary" size="small" @click="artVisible = true">登记产物</el-button>
          <el-button size="small" @click="loadArtifacts">刷 新</el-button>
        </el-form-item>
      </el-form>
      <el-table :data="artifacts" v-loading="artLoading" stripe size="small">
        <el-table-column prop="name" label="产物名" width="170" show-overflow-tooltip />
        <el-table-column prop="jobType" label="来源" width="100" />
        <el-table-column prop="method" label="方法" width="90" />
        <el-table-column prop="path" label="路径" min-width="220" show-overflow-tooltip />
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button link type="success" size="small" @click="openRegisterModel(row)">注册服务</el-button>
            <el-button link type="danger" size="small" @click="onDeleteArtifact(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 任务编辑 -->
    <el-dialog v-model="trainVisible" :title="trainForm.id ? '编辑任务' : '新建训练任务'" width="620px">
      <el-form label-width="110px">
        <el-form-item label="任务名"><el-input v-model="trainForm.name" /></el-form-item>
        <el-form-item label="类型">
          <el-select v-model="trainForm.type" style="width: 220px">
            <el-option v-for="t in trainTypes" :key="t" :label="typeLabel(t)" :value="t" />
          </el-select>
        </el-form-item>
        <el-form-item label="基座模型"><el-input v-model="trainForm.baseModel" placeholder="如 /models/qwen-7b" /></el-form-item>
        <el-form-item label="数据集ID"><el-input-number v-model="trainForm.datasetId" :min="0" /></el-form-item>
        <el-form-item v-if="trainForm.type === 'rl'" label="前置任务ID"><el-input-number v-model="trainForm.preJobId" :min="0" /></el-form-item>
        <el-form-item label="超参 JSON">
          <el-input v-model="trainForm.hyperJSON" type="textarea" :rows="5" placeholder='{"learningRate":0.001,"epochs":3,"loraR":16,"zeroStage":2}' />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="trainVisible = false">取 消</el-button>
        <el-button type="primary" :loading="trainSaving" @click="doSaveTrain">保 存</el-button>
      </template>
    </el-dialog>

    <!-- 日志抽屉 -->
    <el-drawer v-model="logVisible" :title="`执行日志 · ${logRow?.name || ''}`" size="55%">
      <pre class="ops-pre">{{ logText || '（暂无日志）' }}</pre>
    </el-drawer>

    <!-- 评估编辑 -->
    <el-dialog v-model="evalVisible" title="新建评估" width="640px">
      <el-form label-width="110px">
        <el-form-item label="评估名"><el-input v-model="evalForm.name" /></el-form-item>
        <el-form-item label="被测模型ID"><el-input-number v-model="evalForm.modelId" :min="1" /></el-form-item>
        <el-form-item label="样本 JSON">
          <el-input v-model="evalForm.samplesJSON" type="textarea" :rows="5" placeholder='[{"question":"什么是 SFT","expected":"有监督微调"}]' />
        </el-form-item>
        <el-form-item label="指标 JSON">
          <el-input v-model="evalForm.metricJSON" type="textarea" :rows="3" placeholder='{"type":"contains"} 或 {"type":"keyword","keywords":["SFT","LoRA"]}' />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="evalVisible = false">取 消</el-button>
        <el-button type="primary" :loading="evalSaving" @click="doSaveEval">保 存</el-button>
      </template>
    </el-dialog>

    <!-- 评估明细 -->
    <el-dialog v-model="evalResultVisible" :title="`评估明细 · ${evalRow?.name || ''}`" width="860px">
      <el-table :data="evalResults" stripe size="small">
        <el-table-column prop="idx" label="#" width="50" />
        <el-table-column prop="question" label="问题" min-width="150" show-overflow-tooltip />
        <el-table-column prop="actual" label="实际输出" min-width="180" show-overflow-tooltip />
        <el-table-column label="得分" width="80">
          <template #default="{ row }">{{ row.score.toFixed?.(2) ?? row.score }}</template>
        </el-table-column>
        <el-table-column label="通过" width="70">
          <template #default="{ row }">
            <el-tag size="small" :type="row.pass ? 'success' : 'danger'">{{ row.pass ? '✓' : '✗' }}</el-tag>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>

    <!-- 产物登记 -->
    <el-dialog v-model="artVisible" title="登记产物" width="520px">
      <el-form label-width="100px">
        <el-form-item label="产物名"><el-input v-model="artForm.name" /></el-form-item>
        <el-form-item label="来源类型">
          <el-select v-model="artForm.jobType" style="width: 200px">
            <el-option v-for="t in ['sft_full','sft_lora','reward','rl','compress']" :key="t" :label="t" :value="t" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="artForm.jobType === 'compress'" label="压缩方法">
          <el-select v-model="artForm.method" style="width: 200px">
            <el-option v-for="m in ['quantize','prune','distill']" :key="m" :label="m" :value="m" />
          </el-select>
        </el-form-item>
        <el-form-item label="路径"><el-input v-model="artForm.path" placeholder="/data/outputs/..." /></el-form-item>
        <el-form-item label="备注"><el-input v-model="artForm.notes" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="artVisible = false">取 消</el-button>
        <el-button type="primary" :loading="artSaving" @click="doSaveArtifact">登 记</el-button>
      </template>
    </el-dialog>

    <!-- 注册服务 -->
    <el-dialog v-model="regVisible" title="产物注册为模型服务" width="520px">
      <el-form label-width="100px">
        <el-form-item label="能力标签">
          <el-select v-model="regForm.capability" style="width: 200px">
            <el-option label="对话" value="chat" />
            <el-option label="向量" value="embed" />
            <el-option label="重排" value="rerank" />
          </el-select>
        </el-form-item>
        <el-form-item label="服务地址"><el-input v-model="regForm.baseUrl" placeholder="http://宿主:端口/v1" /></el-form-item>
        <el-form-item label="API Key"><el-input v-model="regForm.apiKey" type="password" show-password /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="regVisible = false">取 消</el-button>
        <el-button type="primary" :loading="regSaving" @click="doRegisterModel">注 册</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
  import { ref, reactive, onMounted } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import {
    getTrainJobs, saveTrainJob, transitionTrain, dryRunTrain, startTrain, stopTrain, getTrainLog, deleteTrain,
    getEvalRuns, saveEvalRun, runEval, getEvalResults,
    getArtifacts, registerArtifact, registerArtifactModel, deleteArtifact,
  } from '@/plugin/llmops/api/train'

  const trainTypes = ['sft_full', 'sft_lora', 'reward', 'rl', 'eval', 'compress']
  const typeNames = { sft_full: 'SFT 全参', sft_lora: 'SFT LoRA', reward: '奖励模型', rl: '强化学习', eval: '评估', compress: '压缩' }
  const typeLabel = (t) => typeNames[t] || t
  const statusTag = (s) => ({ draft: 'info', queued: 'warning', running: 'primary', succeeded: 'success', failed: 'danger', stopped: 'info', pending: 'info', done: 'success' }[s] || 'info')
  const fmtTime = (t) => (t ? String(t).replace('T', ' ').slice(0, 19) : '—')

  // 训练任务
  const jobs = ref([])
  const jobLoading = ref(false)
  const jobQuery = reactive({ type: '', status: '' })
  const loadJobs = async () => {
    jobLoading.value = true
    try {
      const res = await getTrainJobs({ page: 1, pageSize: 50, type: jobQuery.type, status: jobQuery.status })
      jobs.value = res.data?.list || []
    } finally {
      jobLoading.value = false
    }
  }
  const trainVisible = ref(false)
  const trainSaving = ref(false)
  const trainForm = reactive({ id: 0, name: '', type: 'sft_lora', baseModel: '', datasetId: 0, preJobId: 0, hyperJSON: '' })
  const openTrainEdit = (row) => {
    Object.assign(trainForm, {
      id: row?.ID || 0, name: row?.name || '', type: row?.type || 'sft_lora', baseModel: row?.baseModel || '',
      datasetId: row?.datasetId || 0, preJobId: row?.preJobId || 0, hyperJSON: row?.hyperJSON || '',
    })
    trainVisible.value = true
  }
  const doSaveTrain = async () => {
    trainSaving.value = true
    try {
      await saveTrainJob(trainForm)
      ElMessage.success('保存成功')
      trainVisible.value = false
      loadJobs()
    } finally {
      trainSaving.value = false
    }
  }
  const doDryRun = async (row) => {
    await dryRunTrain(row.ID)
    ElMessage.success('渲染完成，查看日志')
    openLog(row)
  }
  const doStart = async (row) => {
    await startTrain(row.ID)
    ElMessage.success('已启动（dry-run 执行器）')
    loadJobs()
  }
  const doStop = (row) => {
    ElMessageBox.prompt('停止原因', `停止 ${row.name}`, { type: 'warning' })
      .then(async ({ value }) => {
        await stopTrain({ id: row.ID, reason: value || '手动停止' })
        ElMessage.success('已停止')
        loadJobs()
      })
      .catch(() => {})
  }
  const logVisible = ref(false)
  const logRow = ref(null)
  const logText = ref('')
  const openLog = async (row) => {
    logRow.value = row
    const res = await getTrainLog({ id: row.ID })
    logText.value = res.data?.log || ''
    logVisible.value = true
  }
  const onDeleteJob = (row) => {
    ElMessageBox.confirm(`确认删除任务「${row.name}」？`, '提示', { type: 'warning' })
      .then(async () => {
        await deleteTrain({ id: row.ID })
        ElMessage.success('删除成功')
        loadJobs()
      })
      .catch(() => {})
  }
  void transitionTrain

  // 评估
  const evals = ref([])
  const evalLoading = ref(false)
  const loadEvals = async () => {
    evalLoading.value = true
    try {
      const res = await getEvalRuns({ page: 1, pageSize: 50 })
      evals.value = (res.data?.list || []).map((e) => ({ ...e, _running: false }))
    } finally {
      evalLoading.value = false
    }
  }
  const evalVisible = ref(false)
  const evalSaving = ref(false)
  const evalForm = reactive({ name: '', modelId: 1, samplesJSON: '', metricJSON: '' })
  const doSaveEval = async () => {
    evalSaving.value = true
    try {
      await saveEvalRun(evalForm)
      ElMessage.success('保存成功')
      evalVisible.value = false
      loadEvals()
    } finally {
      evalSaving.value = false
    }
  }
  const doRunEval = async (row) => {
    row._running = true
    try {
      const res = await runEval(row.ID)
      ElMessage.success(`评估完成：${res.data?.passCount}/${res.data?.sampleCount} 通过`)
      loadEvals()
    } finally {
      row._running = false
    }
  }
  const evalResultVisible = ref(false)
  const evalRow = ref(null)
  const evalResults = ref([])
  const openEvalResults = async (row) => {
    evalRow.value = row
    const res = await getEvalResults({ runId: row.ID })
    evalResults.value = res.data || []
    evalResultVisible.value = true
  }

  // 产物
  const artifacts = ref([])
  const artLoading = ref(false)
  const loadArtifacts = async () => {
    artLoading.value = true
    try {
      const res = await getArtifacts()
      artifacts.value = res.data || []
    } finally {
      artLoading.value = false
    }
  }
  const artVisible = ref(false)
  const artSaving = ref(false)
  const artForm = reactive({ name: '', jobType: 'compress', method: 'quantize', path: '', notes: '' })
  const doSaveArtifact = async () => {
    artSaving.value = true
    try {
      await registerArtifact(artForm)
      ElMessage.success('登记成功')
      artVisible.value = false
      loadArtifacts()
    } finally {
      artSaving.value = false
    }
  }
  const regVisible = ref(false)
  const regSaving = ref(false)
  const regRow = ref(null)
  const regForm = reactive({ capability: 'chat', baseUrl: '', apiKey: '' })
  const openRegisterModel = (row) => {
    regRow.value = row
    regVisible.value = true
  }
  const doRegisterModel = async () => {
    regSaving.value = true
    try {
      await registerArtifactModel({ ...regForm, artifactId: regRow.value.ID })
      ElMessage.success('注册成功，可在模型管理中查看')
      regVisible.value = false
    } finally {
      regSaving.value = false
    }
  }
  const onDeleteArtifact = (row) => {
    ElMessageBox.confirm(`确认删除产物「${row.name}」？`, '提示', { type: 'warning' })
      .then(async () => {
        await deleteArtifact({ id: row.ID })
        ElMessage.success('删除成功')
        loadArtifacts()
      })
      .catch(() => {})
  }

  onMounted(() => {
    loadJobs()
    loadEvals()
    loadArtifacts()
  })
</script>
