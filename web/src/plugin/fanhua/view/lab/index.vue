<template>
  <div class="ops-card">
    <el-tabs v-model="tab">
      <!-- 语料库 -->
      <el-tab-pane label="语料库" name="corpus">
        <el-form inline>
          <el-form-item label="来源">
            <el-select v-model="corpusQuery.source" clearable placeholder="全部" style="width: 110px" @change="loadCorpora">
              <el-option v-for="s in sources" :key="s" :label="s" :value="s" />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-button size="small" @click="importVisible = true">批量导入</el-button>
            <el-button size="small" @click="loadCorpora">刷 新</el-button>
          </el-form-item>
        </el-form>
        <el-table :data="corpora" v-loading="corpusLoading" stripe size="small">
          <el-table-column label="来源" width="80">
            <template #default="{ row }"><el-tag size="small">{{ row.source }}</el-tag></template>
          </el-table-column>
          <el-table-column prop="title" label="标题" min-width="160" show-overflow-tooltip />
          <el-table-column prop="lang" label="语种" width="80" />
          <el-table-column prop="content" label="内容" min-width="260" show-overflow-tooltip />
        </el-table>
      </el-tab-pane>

      <!-- 平行语料 -->
      <el-tab-pane label="平行语料" name="parallel">
        <el-form inline>
          <el-form-item>
            <el-input v-model="paraQuery.keyword" placeholder="关键字" style="width: 200px" clearable @keyup.enter="loadParallels" />
            <el-button size="small" style="margin-left: 8px" @click="loadParallels">搜 索</el-button>
          </el-form-item>
          <el-form-item>
            <el-button type="success" size="small" @click="paraSearchVisible = true">对照检索</el-button>
          </el-form-item>
        </el-form>
        <el-table :data="parallels" v-loading="paraLoading" stripe size="small">
          <el-table-column prop="sh" label="沪语" min-width="180" show-overflow-tooltip />
          <el-table-column prop="pu" label="普通话" min-width="180" show-overflow-tooltip />
          <el-table-column prop="en" label="英语" min-width="140" show-overflow-tooltip />
          <el-table-column prop="source" label="来源" width="120" show-overflow-tooltip />
        </el-table>
      </el-tab-pane>

      <!-- 标注任务 -->
      <el-tab-pane label="标注任务" name="label">
        <el-form inline>
          <el-form-item label="训练占比"><el-input-number v-model="labelForm.trainRatio" :min="0.1" :max="0.9" :step="0.1" /></el-form-item>
          <el-form-item>
            <el-button type="primary" size="small" :loading="labelBuilding" @click="doBuildLabels">从语料库构建标注</el-button>
          </el-form-item>
        </el-form>
        <el-descriptions v-if="labelStats" :column="1" border size="small" style="margin-bottom: 10px">
          <el-descriptions-item label="分布（划分×情感）">{{ JSON.stringify(labelStats.distribution) }}</el-descriptions-item>
        </el-descriptions>
        <el-table :data="labels" v-loading="labelLoading" stripe size="small">
          <el-table-column label="划分" width="70">
            <template #default="{ row }"><el-tag size="small">{{ row.split }}</el-tag></template>
          </el-table-column>
          <el-table-column label="情感" width="70">
            <template #default="{ row }"><el-tag size="small" type="info">{{ row.sentiment || row.autoSentiment }}</el-tag></template>
          </el-table-column>
          <el-table-column prop="text" label="文本" min-width="300" show-overflow-tooltip />
          <el-table-column label="人工" width="60">
            <template #default="{ row }">{{ row.labeled ? '是' : '否' }}</template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- 知识库 -->
      <el-tab-pane label="知识库" name="culture">
        <el-form inline>
          <el-form-item label="分类">
            <el-select v-model="voiceQuery.category" clearable placeholder="全部" style="width: 120px" @change="loadVoices">
              <el-option v-for="c in voiceCats" :key="c" :label="c" :value="c" />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-input v-model="cultureQuery.query" placeholder="组合检索词" style="width: 200px" />
            <el-button type="primary" size="small" style="margin-left: 8px" :loading="cultureLoading" @click="doCultureSearch">检 索</el-button>
          </el-form-item>
        </el-form>
        <el-table :data="voices" v-loading="voiceLoading" stripe size="small">
          <el-table-column prop="title" label="标题" min-width="160" show-overflow-tooltip />
          <el-table-column label="分类" width="90">
            <template #default="{ row }"><el-tag size="small">{{ row.category }}</el-tag></template>
          </el-table-column>
          <el-table-column prop="speakerType" label="声部" width="70" />
          <el-table-column prop="content" label="内容" min-width="260" show-overflow-tooltip />
          <el-table-column prop="textSource" label="出处" width="150" show-overflow-tooltip />
        </el-table>
        <el-descriptions v-if="cultureResult" :column="1" border size="small" style="margin-top: 10px">
          <el-descriptions-item label="语音命中">{{ JSON.stringify(cultureResult.voiceHits && cultureResult.voiceHits.map(h => h.title)) }}</el-descriptions-item>
          <el-descriptions-item label="平行命中">{{ JSON.stringify(cultureResult.parallelHits && cultureResult.parallelHits.map(h => h.sh)) }}</el-descriptions-item>
        </el-descriptions>
      </el-tab-pane>

      <!-- 审查测试台 -->
      <el-tab-pane label="审查测试台" name="audit">
        <el-form inline>
          <el-form-item>
            <el-input v-model="auditText" type="textarea" :rows="2" placeholder="输入待审查文本（中英皆可）" style="width: 420px" />
            <el-button type="danger" size="small" style="margin-left: 8px" :loading="auditLoading" @click="doAudit">审 查</el-button>
          </el-form-item>
        </el-form>
        <el-descriptions v-if="auditResult" :column="3" border size="small" style="margin-bottom: 10px">
          <el-descriptions-item label="风险等级"><el-tag :type="auditResult.riskLevel === '高' ? 'danger' : auditResult.riskLevel === '中' ? 'warning' : 'success'" size="small">{{ auditResult.riskLevel }}</el-tag></el-descriptions-item>
          <el-descriptions-item label="健康度">{{ auditResult.health }}</el-descriptions-item>
          <el-descriptions-item label="预警">{{ auditResult.alarm ? '是' : '否' }}</el-descriptions-item>
        </el-descriptions>
        <el-form inline>
          <el-form-item label="等级">
            <el-select v-model="auditQuery.riskLevel" clearable placeholder="全部" style="width: 100px" @change="loadAuditLogs">
              <el-option label="高" value="高" /><el-option label="中" value="中" /><el-option label="低" value="低" />
            </el-select>
          </el-form-item>
          <el-form-item><el-checkbox v-model="auditQuery.alarmOnly" @change="loadAuditLogs">仅预警</el-checkbox></el-form-item>
        </el-form>
        <el-table :data="auditLogs" v-loading="auditLogLoading" stripe size="small">
          <el-table-column label="等级" width="70">
            <template #default="{ row }"><el-tag size="small" :type="row.riskLevel === '高' ? 'danger' : row.riskLevel === '中' ? 'warning' : 'success'">{{ row.riskLevel }}</el-tag></template>
          </el-table-column>
          <el-table-column prop="health" label="健康度" width="80" />
          <el-table-column prop="text" label="文本" min-width="280" show-overflow-tooltip />
          <el-table-column label="预警" width="70">
            <template #default="{ row }">{{ row.alarm ? '是' : '否' }}</template>
          </el-table-column>
          <el-table-column prop="latencyMs" label="时延ms" width="80" />
        </el-table>
      </el-tab-pane>
    </el-tabs>

    <!-- 导入对话框 -->
    <el-dialog v-model="importVisible" title="批量导入语料" width="720px">
      <el-form label-width="80px">
        <el-form-item label="格式">
          <el-radio-group v-model="importForm.format">
            <el-radio value="jsonl">JSONL</el-radio>
            <el-radio value="csv">CSV</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="内容">
          <el-input v-model="importForm.content" type="textarea" :rows="8" placeholder='JSONL 每行一条 {"source":"新闻","title":"…","lang":"沪语","content":"…"}；CSV 列序 source,title,lang,content' />
        </el-form-item>
      </el-form>
      <el-descriptions v-if="importResult" :column="1" border size="small">
        <el-descriptions-item label="结果">入库 {{ importResult.inserted }} 条，失败 {{ (importResult.failedRows || []).length }} 行</el-descriptions-item>
      </el-descriptions>
      <template #footer>
        <el-button @click="importVisible = false">关 闭</el-button>
        <el-button type="primary" :loading="importLoading" @click="doImport">导 入</el-button>
      </template>
    </el-dialog>

    <!-- 对照检索对话框 -->
    <el-dialog v-model="paraSearchVisible" title="普通话→沪语对照检索" width="680px">
      <el-input v-model="paraSearchQuery" placeholder="输入普通话句子" />
      <el-table :data="paraHits" size="small" stripe style="margin-top: 10px" max-height="320">
        <el-table-column prop="sh" label="沪语" min-width="160" />
        <el-table-column prop="pu" label="普通话" min-width="160" />
        <el-table-column prop="score" label="相似度" width="90" />
      </el-table>
      <template #footer>
        <el-button @click="paraSearchVisible = false">关 闭</el-button>
        <el-button type="primary" :loading="paraSearchLoading" @click="doParaSearch">检 索</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
  import { ref, reactive, onMounted } from 'vue'
  import { ElMessage } from 'element-plus'
  import {
    getCorpora, importCorpora,
    getParallels, searchParallel,
    buildLabels, getTextLabels, getLabelStats,
    getVoiceEntries, searchCulture,
    auditValues, getAuditLogs,
  } from '@/plugin/fanhua/api/fanhua'

  const tab = ref('corpus')
  const sources = ['新闻', '社交', '文学', '对照', '舆情', '城市']
  const voiceCats = ['沪剧台词', '弄堂俚语', '经典朗读']

  // 语料库
  const corpora = ref([])
  const corpusLoading = ref(false)
  const corpusQuery = reactive({ source: '' })
  const importVisible = ref(false)
  const importForm = reactive({ format: 'jsonl', content: '' })
  const importLoading = ref(false)
  const importResult = ref(null)

  const loadCorpora = async () => {
    corpusLoading.value = true
    try {
      const res = await getCorpora({ ...corpusQuery, pageSize: 30 })
      corpora.value = (res.data && res.data.list) || []
    } finally {
      corpusLoading.value = false
    }
  }
  const doImport = async () => {
    importLoading.value = true
    try {
      const res = await importCorpora(importForm)
      importResult.value = res.data
      ElMessage.success(`导入 ${res.data.inserted} 条`)
      loadCorpora()
    } finally {
      importLoading.value = false
    }
  }

  // 平行语料
  const parallels = ref([])
  const paraLoading = ref(false)
  const paraQuery = reactive({ keyword: '' })
  const paraSearchVisible = ref(false)
  const paraSearchQuery = ref('')
  const paraHits = ref([])
  const paraSearchLoading = ref(false)

  const loadParallels = async () => {
    paraLoading.value = true
    try {
      const res = await getParallels({ ...paraQuery, pageSize: 20 })
      parallels.value = (res.data && res.data.list) || []
    } finally {
      paraLoading.value = false
    }
  }
  const doParaSearch = async () => {
    paraSearchLoading.value = true
    try {
      const res = await searchParallel({ query: paraSearchQuery.value, topK: 5 })
      paraHits.value = res.data || []
    } finally {
      paraSearchLoading.value = false
    }
  }

  // 标注任务
  const labelForm = reactive({ trainRatio: 0.8, limit: 100 })
  const labelBuilding = ref(false)
  const labels = ref([])
  const labelLoading = ref(false)
  const labelStats = ref(null)

  const doBuildLabels = async () => {
    labelBuilding.value = true
    try {
      const res = await buildLabels(labelForm)
      ElMessage.success(`构建 ${res.data.created} 条`)
      loadLabels()
      loadLabelStats()
    } finally {
      labelBuilding.value = false
    }
  }
  const loadLabels = async () => {
    labelLoading.value = true
    try {
      const res = await getTextLabels({ pageSize: 20 })
      labels.value = (res.data && res.data.list) || []
    } finally {
      labelLoading.value = false
    }
  }
  const loadLabelStats = async () => {
    const res = await getLabelStats()
    labelStats.value = res.data
  }

  // 知识库
  const voices = ref([])
  const voiceLoading = ref(false)
  const voiceQuery = reactive({ category: '' })
  const cultureQuery = reactive({ query: '' })
  const cultureLoading = ref(false)
  const cultureResult = ref(null)

  const loadVoices = async () => {
    voiceLoading.value = true
    try {
      const res = await getVoiceEntries({ ...voiceQuery, pageSize: 20 })
      voices.value = (res.data && res.data.list) || []
    } finally {
      voiceLoading.value = false
    }
  }
  const doCultureSearch = async () => {
    cultureLoading.value = true
    try {
      const res = await searchCulture({ ...cultureQuery, topK: 5 })
      cultureResult.value = res.data
    } finally {
      cultureLoading.value = false
    }
  }

  // 审查测试台
  const auditText = ref('')
  const auditLoading = ref(false)
  const auditResult = ref(null)
  const auditLogs = ref([])
  const auditLogLoading = ref(false)
  const auditQuery = reactive({ riskLevel: '', alarmOnly: false })

  const doAudit = async () => {
    auditLoading.value = true
    try {
      const res = await auditValues({ text: auditText.value })
      auditResult.value = res.data
      loadAuditLogs()
    } finally {
      auditLoading.value = false
    }
  }
  const loadAuditLogs = async () => {
    auditLogLoading.value = true
    try {
      const params = { pageSize: 20 }
      if (auditQuery.riskLevel) params.riskLevel = auditQuery.riskLevel
      if (auditQuery.alarmOnly) params.alarmOnly = true
      const res = await getAuditLogs(params)
      auditLogs.value = (res.data && res.data.list) || []
    } finally {
      auditLogLoading.value = false
    }
  }

  onMounted(() => {
    loadCorpora()
    loadParallels()
    loadLabels()
    loadLabelStats()
    loadVoices()
    loadAuditLogs()
  })
</script>

<style scoped>
  .ops-pre {
    background: #f5f7fa;
    padding: 10px;
    border-radius: 4px;
    max-height: 320px;
    overflow: auto;
    font-size: 12px;
    white-space: pre-wrap;
    word-break: break-all;
  }
</style>
