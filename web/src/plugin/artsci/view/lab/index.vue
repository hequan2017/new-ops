<template>
  <div class="ops-card">
    <el-tabs v-model="tab">
      <!-- 工具执行台 -->
      <el-tab-pane label="工具执行台" name="tool">
        <el-form inline>
          <el-form-item label="学科">
            <el-select v-model="toolQuery.subject" clearable placeholder="全部" style="width: 120px" @change="loadTools">
              <el-option v-for="s in subjects" :key="s" :label="s" :value="s" />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-button size="small" @click="loadTools">刷 新</el-button>
          </el-form-item>
        </el-form>
        <el-table :data="tools" v-loading="toolLoading" stripe size="small">
          <el-table-column prop="name" label="工具" width="190" show-overflow-tooltip />
          <el-table-column label="学科" width="80">
            <template #default="{ row }"><el-tag size="small">{{ row.subject }}</el-tag></template>
          </el-table-column>
          <el-table-column prop="module" label="模块" width="120" />
          <el-table-column prop="desc" label="描述" min-width="220" show-overflow-tooltip />
          <el-table-column label="操作" width="140" fixed="right">
            <template #default="{ row }">
              <el-button link type="success" size="small" @click="openRun(row)">执行</el-button>
              <el-button link type="info" size="small" @click="openLogs(row)">留痕</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- 作业批改 -->
      <el-tab-pane label="作业批改" name="grading">
        <el-form label-width="90px" style="max-width: 860px">
          <el-form-item label="作业标题"><el-input v-model="gradeForm.title" /></el-form-item>
          <el-form-item label="评分标准"><el-input v-model="gradeForm.rubric" type="textarea" :rows="2" placeholder="逗号/顿号分隔要点" /></el-form-item>
          <el-form-item label="学生作答"><el-input v-model="gradeForm.studentText" type="textarea" :rows="4" /></el-form-item>
          <el-form-item label="模型ID"><el-input-number v-model="gradeForm.modelId" :min="0" /><span class="hint">0=规则口径</span></el-form-item>
          <el-form-item>
            <el-button type="primary" :loading="gradeLoading" @click="doGrade">提交批改</el-button>
          </el-form-item>
        </el-form>
        <el-descriptions v-if="gradeResult" :column="1" border size="small" style="max-width: 860px; margin-bottom: 12px">
          <el-descriptions-item label="得分">{{ gradeResult.score }}</el-descriptions-item>
          <el-descriptions-item label="总评">{{ gradeResult.comment }}</el-descriptions-item>
          <el-descriptions-item label="改进建议">{{ gradeResult.suggestions }}</el-descriptions-item>
        </el-descriptions>
        <el-table :data="gradings" v-loading="gradingLoading" stripe size="small">
          <el-table-column prop="title" label="标题" min-width="160" show-overflow-tooltip />
          <el-table-column prop="score" label="得分" width="70" />
          <el-table-column label="引擎" width="70">
            <template #default="{ row }"><el-tag size="small">{{ row.engine }}</el-tag></template>
          </el-table-column>
          <el-table-column prop="comment" label="总评" min-width="200" show-overflow-tooltip />
        </el-table>
      </el-tab-pane>

      <!-- 出题与组卷 -->
      <el-tab-pane label="出题与组卷" name="question">
        <el-form inline>
          <el-form-item label="知识点"><el-input v-model="genForm.knowledgePoint" style="width: 160px" /></el-form-item>
          <el-form-item label="题型">
            <el-select v-model="genForm.qType" style="width: 100px">
              <el-option v-for="t in qTypes" :key="t" :label="t" :value="t" />
            </el-select>
          </el-form-item>
          <el-form-item label="模型ID"><el-input-number v-model="genForm.modelId" :min="0" /></el-form-item>
          <el-form-item>
            <el-button type="primary" size="small" :loading="genLoading" @click="doGenerate">生成题目</el-button>
            <el-button size="small" @click="paperVisible = true">规则组卷</el-button>
          </el-form-item>
        </el-form>
        <el-table :data="questions" v-loading="questionLoading" stripe size="small">
          <el-table-column prop="knowledgePoint" label="知识点" width="130" show-overflow-tooltip />
          <el-table-column prop="qType" label="题型" width="70" />
          <el-table-column prop="difficulty" label="难度" width="70" />
          <el-table-column prop="stem" label="题干" min-width="260" show-overflow-tooltip />
          <el-table-column label="来源" width="80">
            <template #default="{ row }"><el-tag size="small">{{ row.source }}</el-tag></template>
          </el-table-column>
        </el-table>
        <h5 style="margin: 12px 0 6px">试卷</h5>
        <el-table :data="papers" v-loading="paperLoading" stripe size="small">
          <el-table-column prop="name" label="试卷名" min-width="160" show-overflow-tooltip />
          <el-table-column prop="totalScore" label="总分" width="80" />
          <el-table-column prop="questionIds" label="题目ID" min-width="180" show-overflow-tooltip />
        </el-table>
      </el-tab-pane>

      <!-- 文生图场景 -->
      <el-tab-pane label="文生图场景" name="image">
        <el-form inline>
          <el-form-item label="场景">
            <el-select v-model="imgForm.scene" style="width: 160px">
              <el-option v-for="t in scenes" :key="t.scene" :label="t.scene" :value="t.scene" />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-input v-model="imgForm.inputText" type="textarea" :rows="2" placeholder="教学输入文本" style="width: 380px" />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" size="small" :loading="imgLoading" @click="doImageTask">生成</el-button>
          </el-form-item>
        </el-form>
        <el-table :data="imgTasks" v-loading="imgTaskLoading" stripe size="small">
          <el-table-column prop="scene" label="场景" width="130" />
          <el-table-column label="状态" width="80">
            <template #default="{ row }">
              <el-tag size="small" :type="row.status === 'done' ? 'success' : row.status === 'failed' ? 'danger' : 'info'">{{ row.status }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="errMsg" label="说明" min-width="200" show-overflow-tooltip />
          <el-table-column prop="imageUrl" label="图像" min-width="160" show-overflow-tooltip />
        </el-table>
      </el-tab-pane>

      <!-- 图像解读 -->
      <el-tab-pane label="图像解读" name="interpret">
        <el-form inline>
          <el-form-item>
            <el-button type="primary" size="small" @click="assetVisible = true">登记图像资产</el-button>
            <el-button size="small" @click="loadAssets">刷 新</el-button>
          </el-form-item>
        </el-form>
        <el-table :data="assets" v-loading="assetLoading" stripe size="small">
          <el-table-column prop="name" label="名称" min-width="160" show-overflow-tooltip />
          <el-table-column prop="mime" label="MIME" width="110" />
          <el-table-column prop="source" label="来源" width="90" />
          <el-table-column label="操作" width="90" fixed="right">
            <template #default="{ row }">
              <el-button link type="success" size="small" @click="openInterpret(row)">解读</el-button>
            </template>
          </el-table-column>
        </el-table>
        <h5 style="margin: 12px 0 6px">解读记录</h5>
        <el-table :data="interps" v-loading="interpLoading" stripe size="small">
          <el-table-column prop="subject" label="视角" width="120" show-overflow-tooltip />
          <el-table-column prop="surfaceDesc" label="表面细节" min-width="200" show-overflow-tooltip />
          <el-table-column prop="culturalMeaning" label="文化含义" min-width="220" show-overflow-tooltip />
          <el-table-column label="引擎" width="70">
            <template #default="{ row }"><el-tag size="small">{{ row.engine }}</el-tag></template>
          </el-table-column>
        </el-table>
      </el-tab-pane>
    </el-tabs>

    <!-- 执行对话框 -->
    <el-dialog v-model="runVisible" :title="`执行 · ${runRow?.name || ''}`" width="700px">
      <el-input v-model="runArgs" type="textarea" :rows="4" placeholder='参数 JSON，如 {"epsilon":0.02}' />
      <pre v-if="runResult" class="ops-pre" style="margin-top: 10px">{{ runResult }}</pre>
      <template #footer>
        <el-button @click="runVisible = false">关 闭</el-button>
        <el-button type="primary" :loading="runLoading" @click="doRun">执 行</el-button>
      </template>
    </el-dialog>

    <!-- 留痕对话框 -->
    <el-dialog v-model="logsVisible" title="执行留痕" width="760px">
      <el-table :data="runLogs" v-loading="logsLoading" stripe size="small" max-height="420">
        <el-table-column label="成功" width="70">
          <template #default="{ row }"><el-tag size="small" :type="row.ok ? 'success' : 'danger'">{{ row.ok ? '是' : '否' }}</el-tag></template>
        </el-table-column>
        <el-table-column prop="latencyMs" label="耗时(ms)" width="90" />
        <el-table-column prop="errMsg" label="错误" min-width="180" show-overflow-tooltip />
        <el-table-column prop="resultJSON" label="出参" min-width="240" show-overflow-tooltip />
      </el-table>
    </el-dialog>

    <!-- 组卷对话框 -->
    <el-dialog v-model="paperVisible" title="规则组卷" width="640px">
      <el-form label-width="90px">
        <el-form-item label="试卷名"><el-input v-model="paperForm.name" /></el-form-item>
        <el-form-item label="规则JSON">
          <el-input v-model="paperForm.rulesJSON" type="textarea" :rows="4" placeholder='[{"qType":"单选","count":5,"scorePer":2}]' />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="paperVisible = false">取 消</el-button>
        <el-button type="primary" :loading="paperSubmitLoading" @click="doBuildPaper">生成试卷</el-button>
      </template>
    </el-dialog>

    <!-- 资产登记对话框 -->
    <el-dialog v-model="assetVisible" title="登记图像资产" width="640px">
      <el-form label-width="90px">
        <el-form-item label="名称"><el-input v-model="assetForm.name" /></el-form-item>
        <el-form-item label="路径/URL"><el-input v-model="assetForm.path" /></el-form-item>
        <el-form-item label="MIME"><el-input v-model="assetForm.mime" placeholder="image/jpeg" /></el-form-item>
        <el-form-item label="元数据"><el-input v-model="assetForm.metaJSON" type="textarea" :rows="2" placeholder='{"创作者":"...","年代":"..."}' /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="assetVisible = false">取 消</el-button>
        <el-button type="primary" :loading="assetSubmitLoading" @click="doRegisterAsset">登 记</el-button>
      </template>
    </el-dialog>

    <!-- 解读对话框 -->
    <el-dialog v-model="interpVisible" :title="`解读 · ${interpRow?.name || ''}`" width="720px">
      <el-form label-width="90px">
        <el-form-item label="解读视角"><el-input v-model="interpForm.subject" placeholder="艺术作品/历史照片/..." /></el-form-item>
        <el-form-item label="模型ID"><el-input-number v-model="interpForm.modelId" :min="0" /><span class="hint">0=元数据口径</span></el-form-item>
      </el-form>
      <el-descriptions v-if="interpResult" :column="1" border size="small">
        <el-descriptions-item label="表面细节">{{ interpResult.surfaceDesc }}</el-descriptions-item>
        <el-descriptions-item label="文化含义">{{ interpResult.culturalMeaning }}</el-descriptions-item>
      </el-descriptions>
      <template #footer>
        <el-button @click="interpVisible = false">关 闭</el-button>
        <el-button type="primary" :loading="interpSubmitLoading" @click="doInterpret">解 读</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
  import { ref, reactive, onMounted } from 'vue'
  import { ElMessage } from 'element-plus'
  import {
    getTools, runTool, getRunLogs,
    gradeEssay, getGradings,
    generateQuestion, getQuestions, buildPaper, getPapers,
    getSceneTemplates, submitImageTask, getImageTasks,
    registerImageAsset, getImageAssets, interpretImage, getInterpretations,
  } from '@/plugin/artsci/api/artsci'

  const tab = ref('tool')
  const subjects = ['数学', '文科', '图像']
  const qTypes = ['单选', '多选', '判断', '简答']

  // 工具执行台
  const tools = ref([])
  const toolLoading = ref(false)
  const toolQuery = reactive({ subject: '' })
  const runVisible = ref(false)
  const runRow = ref(null)
  const runArgs = ref('{}')
  const runResult = ref('')
  const runLoading = ref(false)
  const logsVisible = ref(false)
  const runLogs = ref([])
  const logsLoading = ref(false)

  const loadTools = async () => {
    toolLoading.value = true
    try {
      const res = await getTools(toolQuery)
      tools.value = res.data || []
    } finally {
      toolLoading.value = false
    }
  }
  const openRun = (row) => {
    runRow.value = row
    runArgs.value = '{}'
    runResult.value = ''
    runVisible.value = true
  }
  const doRun = async () => {
    runLoading.value = true
    try {
      const res = await runTool({ id: runRow.value.ID, argsJSON: runArgs.value })
      runResult.value = JSON.stringify(res.data, null, 2)
    } finally {
      runLoading.value = false
    }
  }
  const openLogs = async (row) => {
    logsVisible.value = true
    logsLoading.value = true
    try {
      const res = await getRunLogs({ toolId: row.ID, pageSize: 50 })
      runLogs.value = (res.data && res.data.list) || []
    } finally {
      logsLoading.value = false
    }
  }

  // 作业批改
  const gradeForm = reactive({ title: '', rubric: '', studentText: '', modelId: 0 })
  const gradeLoading = ref(false)
  const gradeResult = ref(null)
  const gradings = ref([])
  const gradingLoading = ref(false)

  const doGrade = async () => {
    gradeLoading.value = true
    try {
      const res = await gradeEssay(gradeForm)
      gradeResult.value = res.data
      ElMessage.success('批改完成')
      loadGradings()
    } finally {
      gradeLoading.value = false
    }
  }
  const loadGradings = async () => {
    gradingLoading.value = true
    try {
      const res = await getGradings({ pageSize: 20 })
      gradings.value = (res.data && res.data.list) || []
    } finally {
      gradingLoading.value = false
    }
  }

  // 出题与组卷
  const genForm = reactive({ knowledgePoint: '', qType: '单选', modelId: 0 })
  const genLoading = ref(false)
  const questions = ref([])
  const questionLoading = ref(false)
  const papers = ref([])
  const paperLoading = ref(false)
  const paperVisible = ref(false)
  const paperForm = reactive({ name: '', rulesJSON: '[{"qType":"单选","count":5,"scorePer":2}]' })
  const paperSubmitLoading = ref(false)

  const loadQuestions = async () => {
    questionLoading.value = true
    try {
      const res = await getQuestions({ pageSize: 30 })
      questions.value = (res.data && res.data.list) || []
    } finally {
      questionLoading.value = false
    }
  }
  const doGenerate = async () => {
    genLoading.value = true
    try {
      await generateQuestion(genForm)
      ElMessage.success('生成成功')
      loadQuestions()
    } finally {
      genLoading.value = false
    }
  }
  const loadPapers = async () => {
    paperLoading.value = true
    try {
      const res = await getPapers({ pageSize: 20 })
      papers.value = (res.data && res.data.list) || []
    } finally {
      paperLoading.value = false
    }
  }
  const doBuildPaper = async () => {
    paperSubmitLoading.value = true
    try {
      await buildPaper(paperForm)
      ElMessage.success('组卷成功')
      paperVisible.value = false
      loadPapers()
    } finally {
      paperSubmitLoading.value = false
    }
  }

  // 文生图场景
  const scenes = ref([])
  const imgForm = reactive({ scene: '', inputText: '' })
  const imgLoading = ref(false)
  const imgTasks = ref([])
  const imgTaskLoading = ref(false)

  const loadScenes = async () => {
    const res = await getSceneTemplates()
    scenes.value = res.data || []
    if (scenes.value.length && !imgForm.scene) imgForm.scene = scenes.value[0].scene
  }
  const loadImgTasks = async () => {
    imgTaskLoading.value = true
    try {
      const res = await getImageTasks({ pageSize: 20 })
      imgTasks.value = (res.data && res.data.list) || []
    } finally {
      imgTaskLoading.value = false
    }
  }
  const doImageTask = async () => {
    imgLoading.value = true
    try {
      await submitImageTask(imgForm)
      ElMessage.success('任务完成')
      loadImgTasks()
    } finally {
      imgLoading.value = false
    }
  }

  // 图像解读
  const assets = ref([])
  const assetLoading = ref(false)
  const assetVisible = ref(false)
  const assetForm = reactive({ name: '', path: '', mime: 'image/jpeg', metaJSON: '{}' })
  const assetSubmitLoading = ref(false)
  const interpVisible = ref(false)
  const interpRow = ref(null)
  const interpForm = reactive({ modelId: 0 })
  const interpResult = ref(null)
  const interpSubmitLoading = ref(false)
  const interps = ref([])
  const interpLoading = ref(false)

  const loadAssets = async () => {
    assetLoading.value = true
    try {
      const res = await getImageAssets({ pageSize: 20 })
      assets.value = (res.data && res.data.list) || []
    } finally {
      assetLoading.value = false
    }
  }
  const doRegisterAsset = async () => {
    assetSubmitLoading.value = true
    try {
      await registerImageAsset(assetForm)
      ElMessage.success('登记成功')
      assetVisible.value = false
      loadAssets()
    } finally {
      assetSubmitLoading.value = false
    }
  }
  const openInterpret = (row) => {
    interpRow.value = row
    interpResult.value = null
    interpVisible.value = true
  }
  const doInterpret = async () => {
    interpSubmitLoading.value = true
    try {
      const res = await interpretImage({ assetId: interpRow.value.ID, subject: interpForm.subject ?? '', modelId: interpForm.modelId })
      interpResult.value = res.data
      loadInterps()
    } finally {
      interpSubmitLoading.value = false
    }
  }
  const loadInterps = async () => {
    interpLoading.value = true
    try {
      const res = await getInterpretations({ pageSize: 20 })
      interps.value = (res.data && res.data.list) || []
    } finally {
      interpLoading.value = false
    }
  }

  onMounted(() => {
    loadTools()
    loadGradings()
    loadQuestions()
    loadPapers()
    loadScenes()
    loadImgTasks()
    loadAssets()
    loadInterps()
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
  .hint {
    margin-left: 8px;
    color: #909399;
    font-size: 12px;
  }
</style>
