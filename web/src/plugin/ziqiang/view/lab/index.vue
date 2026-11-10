<template>
  <div class="ops-card">
    <el-tabs v-model="tab">
      <!-- 语料 -->
      <el-tab-pane label="语料" name="corpus">
        <el-form inline>
          <el-form-item label="类型">
            <el-select v-model="corpusQuery.type" clearable placeholder="全部" style="width: 110px" @change="loadCorpora">
              <el-option v-for="t in textTypes" :key="t" :label="t" :value="t" />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-button size="small" @click="importVisible = true">批量导入</el-button>
            <el-button size="small" @click="loadCorpora">刷 新</el-button>
          </el-form-item>
        </el-form>
        <el-table :data="corpora" v-loading="corpusLoading" stripe size="small">
          <el-table-column label="类型" width="90">
            <template #default="{ row }"><el-tag size="small">{{ row.type }}</el-tag></template>
          </el-table-column>
          <el-table-column prop="title" label="标题" min-width="150" show-overflow-tooltip />
          <el-table-column prop="source" label="来源" width="120" show-overflow-tooltip />
          <el-table-column label="AIGC" width="70">
            <template #default="{ row }">{{ row.aigcFlag ? '是' : '否' }}</template>
          </el-table-column>
          <el-table-column label="法律状态" width="90">
            <template #default="{ row }"><el-tag size="small" :type="row.legalStatus === '有效' ? 'success' : 'info'">{{ row.legalStatus }}</el-tag></template>
          </el-table-column>
          <el-table-column prop="content" label="内容" min-width="220" show-overflow-tooltip />
        </el-table>

        <el-divider content-position="left">图像鉴权语料</el-divider>
        <el-form inline>
          <el-form-item label="图库">
            <el-input v-model="imageQuery.gallery" placeholder="来源图库" style="width: 140px" clearable @keyup.enter="loadImages" />
            <el-button size="small" style="margin-left: 8px" @click="loadImages">搜 索</el-button>
          </el-form-item>
          <el-form-item label="哈希比对">
            <el-input v-model="compareForm.hashA" placeholder="哈希A" style="width: 140px" />
            <el-input v-model="compareForm.hashB" placeholder="哈希B" style="width: 140px" />
            <el-button size="small" style="margin-left: 8px" :loading="compareLoading" @click="doCompare">比 对</el-button>
          </el-form-item>
        </el-form>
        <el-table :data="images" v-loading="imageLoading" stripe size="small">
          <el-table-column prop="name" label="名称" min-width="140" show-overflow-tooltip />
          <el-table-column prop="hashHex" label="aHash" width="140" />
          <el-table-column prop="watermark" label="水印" width="110" show-overflow-tooltip />
          <el-table-column prop="gallery" label="图库" width="110" show-overflow-tooltip />
          <el-table-column prop="creationTime" label="创作时间" width="110" />
        </el-table>
      </el-tab-pane>

      <!-- 图谱 -->
      <el-tab-pane label="图谱" name="graph">
        <el-form inline>
          <el-form-item label="实体名"><el-input v-model="entityForm.name" style="width: 140px" /></el-form-item>
          <el-form-item label="分类">
            <el-select v-model="entityForm.category" style="width: 110px">
              <el-option v-for="c in entityCats" :key="c" :label="c" :value="c" />
            </el-select>
          </el-form-item>
          <el-form-item label="类型"><el-input v-model="entityForm.kind" style="width: 120px" placeholder="专利号/模型架构…" /></el-form-item>
          <el-form-item><el-button type="primary" size="small" @click="doSaveEntity">加 实 体</el-button></el-form-item>
          <el-form-item label="边">
            <el-input v-model.number="edgeForm.fromId" placeholder="起点ID" style="width: 90px" />
            <el-input v-model.number="edgeForm.toId" placeholder="终点ID" style="width: 90px" />
            <el-input v-model="edgeForm.relation" placeholder="关系" style="width: 100px" />
            <el-button size="small" style="margin-left: 8px" @click="doSaveEdge">加 边</el-button>
          </el-form-item>
        </el-form>
        <el-form inline>
          <el-form-item label="起点ID"><el-input v-model.number="bfsForm.startId" style="width: 90px" /></el-form-item>
          <el-form-item label="跳数"><el-input-number v-model="bfsForm.maxHops" :min="1" :max="6" /></el-form-item>
          <el-form-item><el-button type="success" size="small" :loading="bfsLoading" @click="doBfs">多跳推理</el-button></el-form-item>
        </el-form>
        <el-table :data="entities" v-loading="entityLoading" stripe size="small" style="margin-bottom: 10px">
          <el-table-column prop="id" label="ID" width="60" />
          <el-table-column prop="name" label="实体名" min-width="150" show-overflow-tooltip />
          <el-table-column label="分类" width="90">
            <template #default="{ row }"><el-tag size="small">{{ row.category }}</el-tag></template>
          </el-table-column>
          <el-table-column prop="kind" label="类型" width="110" />
          <el-table-column prop="desc" label="说明" min-width="160" show-overflow-tooltip />
        </el-table>
        <el-descriptions v-if="bfsPaths" :column="1" border size="small">
          <el-descriptions-item label="推理路径">{{ JSON.stringify(bfsPaths.map(p => p.nodeIds)) }}</el-descriptions-item>
        </el-descriptions>
      </el-tab-pane>

      <!-- 知识库 -->
      <el-tab-pane label="知识库" name="kb">
        <el-form inline>
          <el-form-item label="辖区">
            <el-select v-model="kbQuery.jurisdiction" clearable placeholder="全部" style="width: 100px" @change="loadRegulations">
              <el-option v-for="j in jurisdictions" :key="j" :label="j" :value="j" />
            </el-select>
          </el-form-item>
          <el-form-item label="检索词">
            <el-input v-model="kbQuery.query" placeholder="法规全文检索" style="width: 180px" />
            <el-button type="primary" size="small" style="margin-left: 8px" :loading="searchLoading" @click="doRegSearch">检 索</el-button>
          </el-form-item>
          <el-form-item>
            <el-button size="small" :loading="verifyLoading" @click="doVerify">链完整性校验</el-button>
          </el-form-item>
        </el-form>
        <el-table :data="regulations" v-loading="regLoading" stripe size="small" style="margin-bottom: 10px">
          <el-table-column label="辖区" width="80">
            <template #default="{ row }"><el-tag size="small">{{ row.jurisdiction }}</el-tag></template>
          </el-table-column>
          <el-table-column label="分类" width="100">
            <template #default="{ row }">{{ row.category }}</template>
          </el-table-column>
          <el-table-column prop="title" label="法规名" min-width="200" show-overflow-tooltip />
          <el-table-column prop="effectiveDate" label="生效日期" width="110" />
        </el-table>
        <el-divider content-position="left">区块链存证</el-divider>
        <el-form inline>
          <el-form-item label="业务号"><el-input v-model="evidenceForm.bizKey" style="width: 140px" /></el-form-item>
          <el-form-item label="原文"><el-input v-model="evidenceForm.payload" style="width: 260px" /></el-form-item>
          <el-form-item><el-button type="primary" size="small" :loading="evidenceLoading" @click="doAppendEvidence">追加存证</el-button></el-form-item>
        </el-form>
        <el-descriptions v-if="verifyReport" :column="2" border size="small" style="margin-bottom: 10px">
          <el-descriptions-item label="完整性">
            <el-tag size="small" :type="verifyReport.ok ? 'success' : 'danger'">{{ verifyReport.ok ? '通过' : '存在篡改' }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="块数">{{ verifyReport.blocks }}</el-descriptions-item>
          <el-descriptions-item label="问题" :span="2">{{ JSON.stringify(verifyReport.issues) }}</el-descriptions-item>
        </el-descriptions>
        <el-table :data="evidences" v-loading="evidenceListLoading" stripe size="small">
          <el-table-column prop="height" label="块高" width="70" />
          <el-table-column prop="bizKey" label="业务号" width="120" show-overflow-tooltip />
          <el-table-column prop="dataHash" label="数据哈希" min-width="200" show-overflow-tooltip />
          <el-table-column prop="blockHash" label="块哈希" min-width="200" show-overflow-tooltip />
          <el-table-column prop="timestamp" label="时间戳" width="120" />
        </el-table>
      </el-tab-pane>

      <!-- 问答台 -->
      <el-tab-pane label="问答台" name="qa">
        <el-form label-width="90px" size="small">
          <el-form-item label="查询文本">
            <el-input v-model="qaForm.text" type="textarea" :rows="3" placeholder="粘贴疑似侵权文本…" style="width: 560px" />
          </el-form-item>
          <el-form-item label="图像哈希">
            <el-input v-model="qaForm.imageHash" placeholder="aHash 16 位 hex（可选）" style="width: 220px" />
            <span style="margin-left: 12px; color: #909399">文本+图像混合 → bigram+aHash 融合排序</span>
          </el-form-item>
          <el-form-item>
            <el-button type="danger" :loading="qaLoading" @click="doQa">侵权风险评估</el-button>
          </el-form-item>
        </el-form>
        <el-descriptions v-if="qaReport" :column="3" border size="small" style="margin-bottom: 10px">
          <el-descriptions-item label="风险等级">
            <el-tag size="small" :type="qaReport.riskLevel === '高' ? 'danger' : qaReport.riskLevel === '中' ? 'warning' : 'success'">{{ qaReport.riskLevel }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="风险分">{{ qaReport.riskScore }}</el-descriptions-item>
          <el-descriptions-item label="图像维度">{{ qaReport.imageProvided ? '已提供' : '未提供' }}</el-descriptions-item>
          <el-descriptions-item label="结论" :span="3">{{ qaReport.verdict }}</el-descriptions-item>
          <el-descriptions-item label="证据" :span="3">{{ JSON.stringify(qaReport.evidence) }}</el-descriptions-item>
          <el-descriptions-item label="免责" :span="3">{{ qaReport.disclaimer }}</el-descriptions-item>
        </el-descriptions>
        <el-table v-if="qaReport" :data="qaReport.topMatches" stripe size="small">
          <el-table-column prop="corpusId" label="语料ID" width="80" />
          <el-table-column prop="title" label="相似专利" min-width="180" show-overflow-tooltip />
          <el-table-column prop="legalStatus" label="法律状态" width="90">
            <template #default="{ row }"><el-tag size="small" :type="row.legalStatus === '有效' ? 'success' : 'info'">{{ row.legalStatus }}</el-tag></template>
          </el-table-column>
          <el-table-column prop="textScore" label="文本分" width="90" />
          <el-table-column prop="imageScore" label="图像分" width="90" />
          <el-table-column prop="fusedScore" label="融合分" width="90" />
        </el-table>

        <el-divider content-position="left">侵权检测提示词</el-divider>
        <el-form inline>
          <el-form-item label="场景">
            <el-select v-model="promptQuery.scene" clearable placeholder="全部" style="width: 130px" @change="loadPrompts">
              <el-option label="多模态双验证" value="多模态双验证" />
              <el-option label="引导式对话" value="引导式对话" />
            </el-select>
          </el-form-item>
          <el-form-item label="模板"><el-input v-model="renderForm.name" placeholder="模板名" style="width: 200px" /></el-form-item>
          <el-form-item>
            <el-button size="small" :loading="renderLoading" @click="doRender">渲 染</el-button>
          </el-form-item>
        </el-form>
        <el-table :data="prompts" stripe size="small" style="margin-bottom: 10px">
          <el-table-column prop="name" label="模板名" min-width="170" show-overflow-tooltip />
          <el-table-column label="场景" width="110">
            <template #default="{ row }"><el-tag size="small">{{ row.scene }}</el-tag></template>
          </el-table-column>
          <el-table-column prop="step" label="步号" width="60" />
          <el-table-column prop="vars" label="占位变量" min-width="180" show-overflow-tooltip />
        </el-table>
        <div v-if="rendered" class="ops-pre">{{ rendered }}</div>
      </el-tab-pane>
    </el-tabs>

    <!-- 批量导入对话框 -->
    <el-dialog v-model="importVisible" title="批量导入文本语料" width="720px">
      <el-form label-width="80px">
        <el-form-item label="格式">
          <el-radio-group v-model="importForm.format">
            <el-radio value="jsonl">JSONL</el-radio>
            <el-radio value="csv">CSV</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="内容">
          <el-input v-model="importForm.content" type="textarea" :rows="8" placeholder='JSONL 每行 {"type":"专利","title":"…","aigcFlag":false,"content":"…"}；CSV 列序 type,title,aigcFlag,content' />
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
  </div>
</template>

<script setup>
  import { ref, reactive, onMounted } from 'vue'
  import { ElMessage } from 'element-plus'
  import {
    getTextCorpora, importTextCorpora,
    getImageCorpora, compareImageHash,
    getEntities, saveEntity, saveEdge, graphBfs,
    getRegulations, searchRegulations,
    appendEvidence, verifyEvidence, getEvidences,
    getPrompts, renderPrompt, qaPatent,
  } from '@/plugin/ziqiang/api/ziqiang'

  const tab = ref('corpus')
  const textTypes = ['专利', '开源协议', '论文', '合同']
  const entityCats = ['法律', '技术', '权利主体']
  const jurisdictions = ['中国', '美国', '欧盟']

  // 语料
  const corpora = ref([])
  const corpusLoading = ref(false)
  const corpusQuery = reactive({ type: '' })
  const importVisible = ref(false)
  const importForm = reactive({ format: 'jsonl', content: '' })
  const importLoading = ref(false)
  const importResult = ref(null)

  const loadCorpora = async () => {
    corpusLoading.value = true
    try {
      const res = await getTextCorpora({ ...corpusQuery, pageSize: 30 })
      corpora.value = (res.data && res.data.list) || []
    } finally {
      corpusLoading.value = false
    }
  }
  const doImport = async () => {
    importLoading.value = true
    try {
      const res = await importTextCorpora(importForm)
      importResult.value = res.data
      ElMessage.success(`导入 ${res.data.inserted} 条`)
      loadCorpora()
    } finally {
      importLoading.value = false
    }
  }

  // 图像语料
  const images = ref([])
  const imageLoading = ref(false)
  const imageQuery = reactive({ gallery: '' })
  const compareForm = reactive({ hashA: '', hashB: '' })
  const compareLoading = ref(false)

  const loadImages = async () => {
    imageLoading.value = true
    try {
      const res = await getImageCorpora({ ...imageQuery, pageSize: 20 })
      images.value = (res.data && res.data.list) || []
    } finally {
      imageLoading.value = false
    }
  }
  const doCompare = async () => {
    compareLoading.value = true
    try {
      const res = await compareImageHash({ hashA: compareForm.hashA, hashB: compareForm.hashB })
      ElMessage.success(`汉明距离 ${res.data.distance}，相似度 ${res.data.similarity}`)
    } finally {
      compareLoading.value = false
    }
  }

  // 图谱
  const entities = ref([])
  const entityLoading = ref(false)
  const entityForm = reactive({ name: '', category: '法律', kind: '', desc: '' })
  const edgeForm = reactive({ fromId: undefined, toId: undefined, relation: '引用' })
  const bfsForm = reactive({ startId: undefined, maxHops: 2 })
  const bfsLoading = ref(false)
  const bfsPaths = ref(null)

  const loadEntities = async () => {
    entityLoading.value = true
    try {
      const res = await getEntities({})
      entities.value = res.data || []
    } finally {
      entityLoading.value = false
    }
  }
  const doSaveEntity = async () => {
    await saveEntity(entityForm)
    ElMessage.success('实体已保存')
    loadEntities()
  }
  const doSaveEdge = async () => {
    await saveEdge(edgeForm)
    ElMessage.success('关系边已保存')
  }
  const doBfs = async () => {
    bfsLoading.value = true
    try {
      const res = await graphBfs({ startId: bfsForm.startId, maxHops: bfsForm.maxHops })
      bfsPaths.value = res.data || []
    } finally {
      bfsLoading.value = false
    }
  }

  // 知识库
  const regulations = ref([])
  const regLoading = ref(false)
  const kbQuery = reactive({ jurisdiction: '', query: '' })
  const searchLoading = ref(false)
  const evidences = ref([])
  const evidenceListLoading = ref(false)
  const evidenceForm = reactive({ bizKey: '', payload: '' })
  const evidenceLoading = ref(false)
  const verifyLoading = ref(false)
  const verifyReport = ref(null)

  const loadRegulations = async () => {
    regLoading.value = true
    try {
      const res = await getRegulations({ ...kbQuery, pageSize: 20 })
      regulations.value = (res.data && res.data.list) || []
    } finally {
      regLoading.value = false
    }
  }
  const doRegSearch = async () => {
    searchLoading.value = true
    try {
      const res = await searchRegulations({ query: kbQuery.query, jurisdiction: kbQuery.jurisdiction, topK: 5 })
      regulations.value = (res.data || []).map(h => ({ ...h, score: h.score }))
    } finally {
      searchLoading.value = false
    }
  }
  const loadEvidences = async () => {
    evidenceListLoading.value = true
    try {
      const res = await getEvidences({ pageSize: 20 })
      evidences.value = (res.data && res.data.list) || []
    } finally {
      evidenceListLoading.value = false
    }
  }
  const doAppendEvidence = async () => {
    evidenceLoading.value = true
    try {
      const res = await appendEvidence(evidenceForm)
      ElMessage.success(`已上链 块高 ${res.data.height}`)
      loadEvidences()
    } finally {
      evidenceLoading.value = false
    }
  }
  const doVerify = async () => {
    verifyLoading.value = true
    try {
      const res = await verifyEvidence()
      verifyReport.value = res.data
    } finally {
      verifyLoading.value = false
    }
  }

  // 问答台
  const qaForm = reactive({ text: '', imageHash: '' })
  const qaLoading = ref(false)
  const qaReport = ref(null)
  const prompts = ref([])
  const promptQuery = reactive({ scene: '' })
  const renderForm = reactive({ name: '', vars: {} })
  const renderLoading = ref(false)
  const rendered = ref('')

  const doQa = async () => {
    qaLoading.value = true
    try {
      const res = await qaPatent({ text: qaForm.text, imageHash: qaForm.imageHash || undefined, topK: 5 })
      qaReport.value = res.data
    } finally {
      qaLoading.value = false
    }
  }
  const loadPrompts = async () => {
    const res = await getPrompts({ ...promptQuery })
    prompts.value = res.data || []
  }
  const doRender = async () => {
    renderLoading.value = true
    try {
      const res = await renderPrompt({ name: renderForm.name, vars: renderForm.vars })
      rendered.value = res.data.rendered
    } finally {
      renderLoading.value = false
    }
  }

  onMounted(() => {
    loadCorpora()
    loadImages()
    loadEntities()
    loadRegulations()
    loadEvidences()
    loadPrompts()
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
