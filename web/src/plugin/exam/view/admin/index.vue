<template>
  <div>
    <el-tabs v-model="tab" type="border-card">
      <!-- 题库 -->
      <el-tab-pane label="题库" name="q">
        <el-form inline>
          <el-form-item>
            <el-button type="primary" size="small" @click="qVisible = true">新建考题</el-button>
            <el-button size="small" @click="importVisible = true">JSONL 导入</el-button>
            <el-button size="small" @click="loadQuestions">刷 新</el-button>
          </el-form-item>
          <el-form-item label="题型">
            <el-select v-model="qQuery.type" clearable placeholder="全部" style="width: 120px" @change="loadQuestions">
              <el-option v-for="t in typeNames" :key="t.value" :label="t.label" :value="t.value" />
            </el-select>
          </el-form-item>
        </el-form>
        <el-table :data="questions" v-loading="qLoading" stripe size="small">
          <el-table-column prop="stem" label="题干" min-width="240" show-overflow-tooltip />
          <el-table-column label="题型" width="80">
            <template #default="{ row }"><el-tag size="small">{{ typeLabel(row.type) }}</el-tag></template>
          </el-table-column>
          <el-table-column prop="knowledgePoint" label="知识点" width="110" show-overflow-tooltip />
          <el-table-column prop="score" label="分值" width="70" />
          <el-table-column prop="difficulty" label="难度" width="60" />
          <el-table-column label="操作" width="120" fixed="right">
            <template #default="{ row }">
              <el-button link type="primary" size="small" @click="openQEdit(row)">编辑</el-button>
              <el-button link type="danger" size="small" @click="onDeleteQ(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- 试卷 -->
      <el-tab-pane label="试卷" name="p">
        <el-form inline>
          <el-form-item>
            <el-button type="primary" size="small" @click="pVisible = true">新建试卷</el-button>
            <el-button size="small" @click="loadPapers">刷 新</el-button>
          </el-form-item>
        </el-form>
        <el-table :data="papers" v-loading="pLoading" stripe size="small">
          <el-table-column prop="name" label="试卷名" min-width="180" show-overflow-tooltip />
          <el-table-column label="总分" width="90">
            <template #default="{ row }">{{ row.totalScore || '—' }}</template>
          </el-table-column>
          <el-table-column label="状态" width="90">
            <template #default="{ row }">
              <el-tag size="small" :type="row.status === 'ready' ? 'success' : 'info'">{{ row.status === 'ready' ? '已组卷' : '草稿' }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="130" fixed="right">
            <template #default="{ row }">
              <el-button link type="success" size="small" @click="doCompose(row)">组卷</el-button>
              <el-button link type="danger" size="small" @click="onDeleteP(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- 场次 -->
      <el-tab-pane label="场次" name="e">
        <el-form inline>
          <el-form-item>
            <el-button type="primary" size="small" @click="eVisible = true">新建考试</el-button>
            <el-button size="small" @click="loadExams">刷 新</el-button>
          </el-form-item>
        </el-form>
        <el-table :data="exams" v-loading="eLoading" stripe size="small">
          <el-table-column prop="name" label="考试" min-width="160" show-overflow-tooltip />
          <el-table-column prop="paperId" label="试卷" width="70" />
          <el-table-column label="时间窗" min-width="240">
            <template #default="{ row }">{{ fmtTime(row.startTime) }} ~ {{ fmtTime(row.endTime) }}</template>
          </el-table-column>
          <el-table-column prop="durationMin" label="时长(分)" width="85" />
          <el-table-column label="状态" width="90">
            <template #default="{ row }">
              <el-tag size="small" :type="examTag(row.status)">{{ examLabel(row.status) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="200" fixed="right">
            <template #default="{ row }">
              <el-button v-if="row.status === 'draft'" link type="success" size="small" @click="doTrans(row, 'published')">发布</el-button>
              <el-button v-if="row.status === 'published'" link type="success" size="small" @click="doTrans(row, 'ongoing')">开始</el-button>
              <el-button v-if="row.status === 'ongoing'" link type="warning" size="small" @click="doTrans(row, 'finished')">结束</el-button>
              <el-button v-if="row.status === 'draft'" link type="primary" size="small" @click="openEEdit(row)">编辑</el-button>
              <el-button v-if="['draft','finished'].includes(row.status)" link type="danger" size="small" @click="onDeleteE(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- 考生与节点 -->
      <el-tab-pane label="考生" name="c">
        <el-form inline>
          <el-form-item>
            <el-button type="primary" size="small" @click="cVisible = true">名单导入</el-button>
            <el-button size="small" @click="loadCandidates(); loadNodes()">刷 新</el-button>
          </el-form-item>
          <el-form-item label="考试">
            <el-select v-model="candExamId" placeholder="选择考试" style="width: 200px" @change="loadCandidates">
              <el-option v-for="e in exams" :key="e.ID" :label="e.name" :value="e.ID" />
            </el-select>
          </el-form-item>
        </el-form>
        <el-table :data="candidates" v-loading="cLoading" stripe size="small">
          <el-table-column prop="name" label="考生" width="120" />
          <el-table-column prop="ticketCode" label="券码" width="150" />
          <el-table-column prop="userKey" label="标识" min-width="140" show-overflow-tooltip />
          <el-table-column label="状态" width="90">
            <template #default="{ row }">
              <el-tag size="small">{{ candLabel(row.status) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="90" fixed="right">
            <template #default="{ row }">
              <el-button link type="danger" size="small" @click="onDeleteC(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
        <h4 style="margin: 12px 0 6px">考试节点（分布式终端）</h4>
        <el-table :data="nodes" v-loading="nLoading" stripe size="small">
          <el-table-column prop="nodeKey" label="标识" min-width="140" show-overflow-tooltip />
          <el-table-column prop="region" label="考点" width="110" />
          <el-table-column prop="load" label="负载" width="70" />
          <el-table-column label="状态" width="90">
            <template #default="{ row }">
              <el-tag size="small" :type="row.status === 'online' ? 'success' : 'info'">{{ row.status }}</el-tag>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- 场次日历 -->
      <el-tab-pane label="日历" name="cal">
        <el-calendar v-model="calDate">
          <template #date-cell="{ data }">
            <div style="min-height: 42px">
              <div style="font-size: 12px">{{ data.day.slice(8) }}</div>
              <div v-for="e in examsOn(data.day)" :key="e.ID">
                <el-tag size="small" :type="examTag(e.status)" style="margin: 1px 0; max-width: 100%; overflow: hidden; text-overflow: ellipsis">
                  {{ e.name }}
                </el-tag>
              </div>
            </div>
          </template>
        </el-calendar>
        <el-alert type="info" :closable="false" title="按开考日期展示全部场次；点击格内考试名可在「场次」页签管理" style="margin-top: 8px" />
      </el-tab-pane>

      <!-- 成绩 -->
      <el-tab-pane label="成绩" name="s">
        <el-form inline>
          <el-form-item>
            <el-button type="primary" size="small" @click="pubVisible = true">判分并发布</el-button>
            <el-button size="small" @click="doExport">导出 CSV</el-button>
            <el-button type="warning" size="small" @click="simVisible = true">相似度抽查</el-button>
            <el-button size="small" @click="loadScores">刷 新</el-button>
          </el-form-item>
          <el-form-item label="考试">
            <el-select v-model="scoreExamId" placeholder="选择考试" style="width: 200px" @change="loadScores">
              <el-option v-for="e in exams" :key="e.ID" :label="e.name" :value="e.ID" />
            </el-select>
          </el-form-item>
        </el-form>
        <el-table :data="scores" v-loading="sLoading" stripe size="small">
          <el-table-column prop="candidateId" label="考生ID" width="90" />
          <el-table-column prop="total" label="总分" width="90" />
          <el-table-column prop="full" label="满分" width="90" />
          <el-table-column label="等级" width="100">
            <template #default="{ row }">
              <el-tag size="small" :type="levelTag(row.level)">{{ row.level }}</el-tag>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>
    </el-tabs>

    <!-- 考题编辑 -->
    <el-dialog v-model="qVisible" :title="qForm.id ? '编辑考题' : '新建考题'" width="640px">
      <el-form label-width="90px">
        <el-form-item label="题型">
          <el-select v-model="qForm.type" style="width: 200px">
            <el-option v-for="t in typeNames" :key="t.value" :label="t.label" :value="t.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="题干"><el-input v-model="qForm.stem" type="textarea" :rows="3" /></el-form-item>
        <el-form-item v-if="['single','multiple'].includes(qForm.type)" label="选项">
          <el-input v-model="qForm.optionsText" type="textarea" :rows="3" placeholder="每行一个选项" />
        </el-form-item>
        <el-form-item label="答案"><el-input v-model="qForm.answer" placeholder="如 A / A,B,D / 对 / 参考要点" /></el-form-item>
        <el-form-item label="分值"><el-input-number v-model="qForm.score" :min="1" /></el-form-item>
        <el-form-item label="知识点"><el-input v-model="qForm.knowledgePoint" /></el-form-item>
        <el-form-item label="难度"><el-input-number v-model="qForm.difficulty" :min="1" :max="5" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="qVisible = false">取 消</el-button>
        <el-button type="primary" :loading="qSaving" @click="doSaveQ">保 存</el-button>
      </template>
    </el-dialog>

    <!-- JSONL 导入 -->
    <el-dialog v-model="importVisible" title="JSONL 导入" width="560px">
      <input ref="qFileRef" type="file" accept=".jsonl,.json,.txt" />
      <template #footer>
        <el-button @click="importVisible = false">取 消</el-button>
        <el-button type="primary" :loading="importing" @click="doImport">导 入</el-button>
      </template>
    </el-dialog>

    <!-- 试卷编辑 -->
    <el-dialog v-model="pVisible" title="新建试卷" width="640px">
      <el-form label-width="90px">
        <el-form-item label="试卷名"><el-input v-model="pForm.name" /></el-form-item>
        <el-form-item label="抽题规则">
          <el-input v-model="pForm.ruleJSON" type="textarea" :rows="6"
            placeholder='{"sections":[{"type":"single","count":5,"knowledgePoint":"微调","maxDifficulty":3},{"type":"judge","count":3}]}' />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="pVisible = false">取 消</el-button>
        <el-button type="primary" :loading="pSaving" @click="doSaveP">保 存</el-button>
      </template>
    </el-dialog>

    <!-- 考试编辑 -->
    <el-dialog v-model="eVisible" :title="eForm.id ? '编辑考试' : '新建考试'" width="600px">
      <el-form label-width="100px">
        <el-form-item label="考试名"><el-input v-model="eForm.name" /></el-form-item>
        <el-form-item label="试卷ID"><el-input-number v-model="eForm.paperId" :min="1" /></el-form-item>
        <el-form-item label="开考时间"><el-date-picker v-model="eForm.startTime" type="datetime" style="width: 100%" /></el-form-item>
        <el-form-item label="截止时间"><el-date-picker v-model="eForm.endTime" type="datetime" style="width: 100%" /></el-form-item>
        <el-form-item label="时长(分)"><el-input-number v-model="eForm.durationMin" :min="10" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="eVisible = false">取 消</el-button>
        <el-button type="primary" :loading="eSaving" @click="doSaveE">保 存</el-button>
      </template>
    </el-dialog>

    <!-- 名单导入 -->
    <el-dialog v-model="cVisible" title="考生名单导入" width="560px">
      <el-form label-width="90px">
        <el-form-item label="考试">
          <el-select v-model="candExamId" placeholder="选择考试" style="width: 100%">
            <el-option v-for="e in exams" :key="e.ID" :label="e.name" :value="e.ID" />
          </el-select>
        </el-form-item>
        <el-form-item label="名单">
          <el-input v-model="candNamesText" type="textarea" :rows="5" placeholder="每行一个考生姓名" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="cVisible = false">取 消</el-button>
        <el-button type="primary" :loading="cSaving" @click="doImportC">导 入</el-button>
      </template>
    </el-dialog>

    <!-- 成绩发布 -->
    <el-dialog v-model="pubVisible" title="判分并发布成绩" width="520px">
      <el-form label-width="120px">
        <el-form-item label="考试">
          <el-select v-model="scoreExamId" placeholder="选择考试" style="width: 100%">
            <el-option v-for="e in exams" :key="e.ID" :label="e.name" :value="e.ID" />
          </el-select>
        </el-form-item>
        <el-form-item label="评判模型ID">
          <el-input-number v-model="pubForm.judgeModelId" :min="0" />
          <div style="font-size: 12px; color: #999">0 = 主观题不计分</div>
        </el-form-item>
        <el-form-item label="多选漏选给分"><el-switch v-model="pubForm.multiPartial" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="pubVisible = false">取 消</el-button>
        <el-button type="primary" :loading="pubLoading" @click="doPublish">发 布</el-button>
      </template>
    </el-dialog>

    <!-- 相似度抽查 -->
    <el-dialog v-model="simVisible" title="答案相似度抽查" width="640px">
      <el-form label-width="90px">
        <el-form-item label="考试">
          <el-select v-model="scoreExamId" placeholder="选择考试" style="width: 100%">
            <el-option v-for="e in exams" :key="e.ID" :label="e.name" :value="e.ID" />
          </el-select>
        </el-form-item>
        <el-form-item label="Top N"><el-input-number v-model="simSize" :min="1" :max="50" /></el-form-item>
      </el-form>
      <el-table v-if="simPairs.length" :data="simPairs" stripe size="small" style="margin-top: 8px">
        <el-table-column prop="sheetA" label="答卷A" width="90" />
        <el-table-column prop="sheetB" label="答卷B" width="90" />
        <el-table-column label="相似度" width="100">
          <template #default="{ row }">{{ (row.ratio * 100).toFixed(1) }}%</template>
        </el-table-column>
      </el-table>
      <template #footer>
        <el-button @click="simVisible = false">关 闭</el-button>
        <el-button type="primary" :loading="simLoading" @click="doSimCheck">抽 查</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
  import { ref, reactive, onMounted } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import {
    getQuestions, saveQuestion, deleteQuestion, importQuestions,
    getPapers, savePaper, composePaper, deletePaper,
    getExams, saveExam, transitionExam, deleteExam,
    getCandidates, importCandidates, deleteCandidate,
    publishScores, getScores, exportScores,
    getNodes, checkSimilarity,
  } from '@/plugin/exam/api/exam'

  const typeNames = [
    { value: 'single', label: '单选' }, { value: 'multiple', label: '多选' },
    { value: 'judge', label: '判断' }, { value: 'short', label: '简答' }, { value: 'practice', label: '实操' },
  ]
  const typeLabel = (t) => (typeNames.find((x) => x.value === t) || {}).label || t
  const examNames = { draft: '草稿', published: '已发布', ongoing: '进行中', finished: '已结束' }
  const examLabel = (s) => examNames[s] || s
  const examTag = (s) => ({ draft: 'info', published: 'warning', ongoing: 'success', finished: 'info' }[s] || 'info')
  const candNames = { pending: '待考试', entered: '已进场', submitted: '已交卷' }
  const candLabel = (s) => candNames[s] || s
  const levelTag = (l) => ({ 优秀: 'success', 良好: 'primary', 及格: 'warning', 不及格: 'danger' }[l] || 'info')
  const fmtTime = (t) => (t ? String(t).replace('T', ' ').slice(0, 16) : '—')

  const tab = ref('q')
  const calDate = ref(new Date())
  const dayKey = (t) => String(t).slice(0, 10)
  const examsOn = (day) => exams.value.filter((e) => dayKey(e.startTime) === day)

  // 题库
  const questions = ref([])
  const qLoading = ref(false)
  const qQuery = reactive({ type: '' })
  const loadQuestions = async () => {
    qLoading.value = true
    try {
      const res = await getQuestions({ page: 1, pageSize: 50, type: qQuery.type })
      questions.value = res.data?.list || []
    } finally {
      qLoading.value = false
    }
  }
  const qVisible = ref(false)
  const qSaving = ref(false)
  const qForm = reactive({ id: 0, type: 'single', stem: '', optionsText: '', answer: '', score: 1, knowledgePoint: '', difficulty: 2 })
  const openQEdit = (row) => {
    Object.assign(qForm, {
      id: row?.ID || 0, type: row?.type || 'single', stem: row?.stem || '',
      optionsText: (row?.optionsJSON ? JSON.parse(row.optionsJSON) : []).join('\n'),
      answer: row?.answer || '', score: row?.score || 1,
      knowledgePoint: row?.knowledgePoint || '', difficulty: row?.difficulty || 2,
    })
    qVisible.value = true
  }
  const doSaveQ = async () => {
    qSaving.value = true
    try {
      const options = qForm.optionsText.split('\n').map((s) => s.trim()).filter(Boolean)
      await saveQuestion({
        id: qForm.id, type: qForm.type, stem: qForm.stem,
        optionsJSON: options.length ? JSON.stringify(options) : '',
        answer: qForm.answer, score: qForm.score,
        knowledgePoint: qForm.knowledgePoint, difficulty: qForm.difficulty,
      })
      ElMessage.success('保存成功')
      qVisible.value = false
      loadQuestions()
    } finally {
      qSaving.value = false
    }
  }
  const onDeleteQ = (row) => {
    ElMessageBox.confirm('确认删除该考题？', '提示', { type: 'warning' })
      .then(async () => { await deleteQuestion({ id: row.ID }); ElMessage.success('删除成功'); loadQuestions() })
      .catch(() => {})
  }
  const importVisible = ref(false)
  const importing = ref(false)
  const qFileRef = ref(null)
  const doImport = async () => {
    const file = qFileRef.value?.files?.[0]
    if (!file) { ElMessage.warning('请选择 .jsonl 文件'); return }
    importing.value = true
    try {
      const res = await importQuestions(file)
      ElMessage.success(`导入 ${res.data?.imported ?? 0} 题（跳过 ${res.data?.skipped ?? 0}）`)
      importVisible.value = false
      loadQuestions()
    } finally {
      importing.value = false
    }
  }

  // 试卷
  const papers = ref([])
  const pLoading = ref(false)
  const loadPapers = async () => {
    pLoading.value = true
    try {
      const res = await getPapers()
      papers.value = res.data || []
    } finally {
      pLoading.value = false
    }
  }
  const pVisible = ref(false)
  const pSaving = ref(false)
  const pForm = reactive({ name: '', ruleJSON: '' })
  const doSaveP = async () => {
    pSaving.value = true
    try {
      await savePaper(pForm)
      ElMessage.success('保存成功')
      pVisible.value = false
      loadPapers()
    } finally {
      pSaving.value = false
    }
  }
  const doCompose = async (row) => {
    const res = await composePaper(row.ID)
    ElMessage.success(`组卷完成：${res.data?.questionIds?.split(',').length ?? 0} 题，总分 ${res.data?.totalScore}`)
    loadPapers()
  }
  const onDeleteP = (row) => {
    ElMessageBox.confirm(`确认删除试卷「${row.name}」？`, '提示', { type: 'warning' })
      .then(async () => { await deletePaper({ id: row.ID }); ElMessage.success('删除成功'); loadPapers() })
      .catch(() => {})
  }

  // 场次
  const exams = ref([])
  const eLoading = ref(false)
  const loadExams = async () => {
    eLoading.value = true
    try {
      const res = await getExams({})
      exams.value = res.data || []
    } finally {
      eLoading.value = false
    }
  }
  const eVisible = ref(false)
  const eSaving = ref(false)
  const eForm = reactive({ id: 0, name: '', paperId: 1, startTime: '', endTime: '', durationMin: 60 })
  const openEEdit = (row) => {
    Object.assign(eForm, {
      id: row?.ID || 0, name: row?.name || '', paperId: row?.paperId || 1,
      startTime: row?.startTime || '', endTime: row?.endTime || '', durationMin: row?.durationMin || 60,
    })
    eVisible.value = true
  }
  const doSaveE = async () => {
    eSaving.value = true
    try {
      await saveExam(eForm)
      ElMessage.success('保存成功')
      eVisible.value = false
      loadExams()
    } finally {
      eSaving.value = false
    }
  }
  const doTrans = (row, to) => {
    ElMessageBox.confirm(`确认将「${row.name}」流转为 ${examLabel(to)}？`, '提示', { type: 'warning' })
      .then(async () => {
        await transitionExam({ id: row.ID, to })
        ElMessage.success('流转成功')
        loadExams()
      })
      .catch(() => {})
  }
  const onDeleteE = (row) => {
    ElMessageBox.confirm(`确认删除考试「${row.name}」？`, '提示', { type: 'warning' })
      .then(async () => { await deleteExam({ id: row.ID }); ElMessage.success('删除成功'); loadExams() })
      .catch(() => {})
  }

  // 考生与节点
  const candidates = ref([])
  const cLoading = ref(false)
  const candExamId = ref(null)
  const loadCandidates = async () => {
    if (!candExamId.value) return
    cLoading.value = true
    try {
      const res = await getCandidates({ examId: candExamId.value, page: 1, pageSize: 50 })
      candidates.value = res.data?.list || []
    } finally {
      cLoading.value = false
    }
  }
  const cVisible = ref(false)
  const cSaving = ref(false)
  const candNamesText = ref('')
  const doImportC = async () => {
    cSaving.value = true
    try {
      const names = candNamesText.value.split('\n').map((s) => s.trim()).filter(Boolean)
      const res = await importCandidates({ examId: candExamId.value, names })
      ElMessage.success(`导入 ${res.data?.imported ?? 0} 人（跳过 ${res.data?.skipped ?? 0}）`)
      cVisible.value = false
      candNamesText.value = ''
      loadCandidates()
    } finally {
      cSaving.value = false
    }
  }
  const onDeleteC = (row) => {
    ElMessageBox.confirm(`确认删除考生「${row.name}」？`, '提示', { type: 'warning' })
      .then(async () => { await deleteCandidate({ id: row.ID }); ElMessage.success('删除成功'); loadCandidates() })
      .catch(() => {})
  }
  const nodes = ref([])
  const nLoading = ref(false)
  const loadNodes = async () => {
    nLoading.value = true
    try {
      const res = await getNodes()
      nodes.value = res.data || []
    } finally {
      nLoading.value = false
    }
  }

  // 成绩
  const scores = ref([])
  const sLoading = ref(false)
  const scoreExamId = ref(null)
  const loadScores = async () => {
    if (!scoreExamId.value) return
    sLoading.value = true
    try {
      const res = await getScores({ examId: scoreExamId.value, page: 1, pageSize: 50 })
      scores.value = res.data?.list || []
    } finally {
      sLoading.value = false
    }
  }
  const pubVisible = ref(false)
  const pubLoading = ref(false)
  const pubForm = reactive({ judgeModelId: 0, multiPartial: true })
  const doPublish = async () => {
    if (!scoreExamId.value) { ElMessage.warning('请选择考试'); return }
    pubLoading.value = true
    try {
      const res = await publishScores({ examId: scoreExamId.value, ...pubForm })
      ElMessage.success(`已发布 ${res.data?.published ?? 0} 份成绩`)
      pubVisible.value = false
      loadScores()
    } finally {
      pubLoading.value = false
    }
  }
  const doExport = async () => {
    if (!scoreExamId.value) { ElMessage.warning('请选择考试'); return }
    const res = await exportScores({ examId: scoreExamId.value })
    const blob = new Blob([`\ufeff${res.data}`], { type: 'text/csv;charset=utf-8' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `scores-${scoreExamId.value}.csv`
    a.click()
    URL.revokeObjectURL(url)
  }
  const simVisible = ref(false)
  const simLoading = ref(false)
  const simSize = ref(10)
  const simPairs = ref([])
  const doSimCheck = async () => {
    if (!scoreExamId.value) { ElMessage.warning('请选择考试'); return }
    simLoading.value = true
    try {
      const res = await checkSimilarity({ examId: scoreExamId.value, sampleSize: simSize.value })
      simPairs.value = res.data || []
      if (!simPairs.value.length) ElMessage.info('未发现可疑相似对')
    } finally {
      simLoading.value = false
    }
  }

  onMounted(() => {
    loadQuestions()
    loadPapers()
    loadExams()
    loadNodes()
  })
</script>
