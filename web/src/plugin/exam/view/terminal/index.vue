<template>
  <div style="max-width: 960px; margin: 0 auto; padding: 20px">
    <!-- 状态一：终端注册 + 可考列表 -->
    <template v-if="phase === 'lobby'">
      <h2 style="text-align: center">考试终端</h2>
      <div class="ops-card">
        <el-form inline style="justify-content: center">
          <el-form-item label="终端机器码">
            <el-input v-model="nodeKey" placeholder="本机唯一标识" style="width: 240px" />
          </el-form-item>
          <el-form-item label="考点">
            <el-input v-model="region" placeholder="可选" style="width: 140px" />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" :loading="registering" @click="doRegister">注册终端</el-button>
          </el-form-item>
        </el-form>
        <el-empty v-if="!exams.length && registered" description="当前无可考场次" />
        <el-table v-if="exams.length" :data="exams" stripe>
          <el-table-column prop="name" label="考试" min-width="180" />
          <el-table-column label="时间窗" min-width="240">
            <template #default="{ row }">{{ fmtTime(row.startTime) }} ~ {{ fmtTime(row.endTime) }}</template>
          </el-table-column>
          <el-table-column prop="durationMin" label="时长(分)" width="90" />
          <el-table-column label="操作" width="110" fixed="right">
            <template #default="{ row }">
              <el-button type="primary" size="small" @click="chooseExam(row)">进入考试</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </template>

    <!-- 状态二：券码登录 -->
    <template v-else-if="phase === 'ticket'">
      <div class="ops-card" style="max-width: 480px; margin: 60px auto; text-align: center">
        <h3>{{ chosenExam?.name }}</h3>
        <p style="color: #999">请输入考试券码</p>
        <el-input v-model="ticket" size="large" placeholder="12 位券码" style="margin: 16px 0"
          @keyup.enter="doResume" />
        <el-button type="primary" size="large" style="width: 100%" :loading="resuming" @click="doResume">验券进场</el-button>
        <el-button link style="margin-top: 12px" @click="phase = 'lobby'">返回列表</el-button>
      </div>
    </template>

    <!-- 状态三：答题 -->
    <template v-else-if="phase === 'answering'">
      <div class="ops-card">
        <div style="display: flex; justify-content: space-between; align-items: center">
          <h3 style="margin: 0">{{ chosenExam?.name }}</h3>
          <div>
            <el-tag :type="remainSec < 300 ? 'danger' : 'primary'" size="large">{{ countdownText }}</el-tag>
            <el-button size="small" style="margin-left: 10px" @click="doSave(true)">手动保存</el-button>
            <el-button type="danger" size="small" @click="confirmSubmit">交 卷</el-button>
          </div>
        </div>
        <!-- 单题导航 -->
        <div style="margin: 14px 0; display: flex; gap: 6px; flex-wrap: wrap">
          <el-button v-for="(q, i) in questions" :key="q.ID" size="small"
            :type="i === curIdx ? 'primary' : answered(q.ID) ? 'success' : 'info'"
            @click="curIdx = i">{{ i + 1 }}</el-button>
        </div>
        <template v-if="curQ">
          <div style="margin: 10px 0">
            <el-tag size="small">{{ typeLabel(curQ.type) }}</el-tag>
            <el-tag size="small" type="info" style="margin-left: 6px">{{ curQ.score }} 分</el-tag>
            <p style="white-space: pre-wrap; font-size: 15px">{{ curQ.stem }}</p>
          </div>
          <!-- 单选/判断 -->
          <el-radio-group v-if="['single','judge'].includes(curQ.type)" v-model="curAnswer">
            <div style="display: flex; flex-direction: column; gap: 10px">
              <el-radio v-for="opt in optionsOf(curQ)" :key="opt" :value="opt" style="margin: 0">{{ opt }}</el-radio>
            </div>
          </el-radio-group>
          <!-- 多选 -->
          <el-checkbox-group v-else-if="curQ.type === 'multiple'" v-model="multiSel">
            <div style="display: flex; flex-direction: column; gap: 10px">
              <el-checkbox v-for="opt in optionsOf(curQ)" :key="opt" :value="opt">{{ opt }}</el-checkbox>
            </div>
          </el-checkbox-group>
          <!-- 简答 -->
          <el-input v-else-if="curQ.type === 'short'" v-model="curAnswer" type="textarea" :rows="6" />
          <!-- 实操 -->
          <div v-else-if="curQ.type === 'practice'">
            <el-input v-model="practiceInput" type="textarea" :rows="3" placeholder="实操输入（提交给执行通道）" />
            <el-form inline style="margin-top: 8px">
              <el-form-item label="通道">
                <el-select v-model="practiceChannel" size="small" style="width: 120px">
                  <el-option label="算子" value="operator" />
                  <el-option label="智能体" value="agent" />
                </el-select>
              </el-form-item>
              <el-form-item label="目标ID">
                <el-input v-model="practiceTarget" size="small" style="width: 120px" placeholder="算子ID/模型目标" />
              </el-form-item>
              <el-form-item>
                <el-button type="success" size="small" :loading="practicing" @click="doPractice">执行实操</el-button>
              </el-form-item>
            </el-form>
            <pre v-if="practiceOutput" class="ops-pre">{{ practiceOutput }}</pre>
          </div>
        </template>
        <div style="margin-top: 16px; display: flex; justify-content: space-between">
          <el-button :disabled="curIdx <= 0" @click="curIdx--">上一题</el-button>
          <span style="color: #999; font-size: 12px; align-self: center">答案每 30 秒自动保存</span>
          <el-button :disabled="curIdx >= questions.length - 1" @click="curIdx++">下一题</el-button>
        </div>
      </div>
    </template>

    <!-- 状态四：完成 -->
    <template v-else>
      <div class="ops-card" style="max-width: 480px; margin: 60px auto; text-align: center">
        <h2 style="color: #67c23a">交卷成功</h2>
        <p style="color: #999">成绩将在判分后公布</p>
        <el-button type="primary" @click="resetAll">返回列表</el-button>
      </div>
    </template>
  </div>
</template>

<script setup>
  import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import {
    terminalRegister, terminalHeartbeat, terminalExams, terminalResume,
    terminalSave, terminalSubmit, terminalPractice,
  } from '@/plugin/exam/api/terminal'

  const phase = ref('lobby')
  const nodeKey = ref(localStorage.getItem('exam-node-key') || '')
  const region = ref('')
  const registered = ref(false)
  const exams = ref([])
  const chosenExam = ref(null)
  const ticket = ref('')

  const questions = ref([])
  const sheet = ref(null)
  const answers = ref({})
  const remainSec = ref(0)
  const curIdx = ref(0)
  const multiSel = ref([])
  const practiceInput = ref('')
  const practiceChannel = ref('agent')
  const practiceTarget = ref('')
  const practiceOutput = ref('')
  const practicing = ref(false)

  const curQ = computed(() => questions.value[curIdx.value] || null)
  const curAnswer = computed({
    get: () => answers.value[String(curQ.value?.ID)] || '',
    set: (v) => { answers.value[String(curQ.value?.ID)] = v },
  })
  const answered = (id) => {
    const a = answers.value[String(id)]
    return a !== undefined && a !== ''
  }
  const optionsOf = (q) => {
    try { return JSON.parse(q.optionsJSON || '[]') } catch (e) { return [] }
  }
  const typeNames = { single: '单选', multiple: '多选', judge: '判断', short: '简答', practice: '实操' }
  const typeLabel = (t) => typeNames[t] || t
  const fmtTime = (t) => (t ? String(t).replace('T', ' ').slice(0, 16) : '—')
  const countdownText = computed(() => {
    const s = remainSec.value
    const h = Math.floor(s / 3600), m = Math.floor((s % 3600) / 60), ss = s % 60
    return `${String(h).padStart(2, '0')}:${String(m).padStart(2, '0')}:${String(ss).padStart(2, '0')}`
  })

  // 多选 checkbox → 逗号串同步
  watch(multiSel, (v) => {
    if (curQ.value?.type === 'multiple') {
      answers.value[String(curQ.value.ID)] = [...v].sort().join(',')
    }
  })
  watch(curIdx, () => {
    multiSel.value = (curAnswer.value || '').split(',').filter(Boolean)
    practiceOutput.value = ''
  })

  // 终端注册 + 心跳 + 列表
  const registering = ref(false)
  let heartbeatTimer = null
  const doRegister = async () => {
    if (!nodeKey.value) { ElMessage.warning('请输入机器码'); return }
    registering.value = true
    try {
      localStorage.setItem('exam-node-key', nodeKey.value)
      await terminalRegister({ nodeKey: nodeKey.value, region: region.value, name: '考试终端' })
      registered.value = true
      const res = await terminalExams(nodeKey.value)
      exams.value = res.data || []
      ElMessage.success('终端已注册')
      heartbeatTimer = setInterval(async () => {
        try { await terminalHeartbeat({ nodeKey: nodeKey.value, load: phase.value === 'answering' ? 1 : 0 }) } catch (e) { /* 断线容忍 */ }
      }, 60000)
    } finally {
      registering.value = false
    }
  }

  const chooseExam = (row) => {
    chosenExam.value = row
    phase.value = 'ticket'
  }

  // 券码恢复
  const resuming = ref(false)
  const doResume = async () => {
    if (!ticket.value) { ElMessage.warning('请输入券码'); return }
    resuming.value = true
    try {
      const res = await terminalResume({ ticket: ticket.value.trim().toUpperCase(), userKey: nodeKey.value })
      questions.value = res.data?.questions || []
      sheet.value = res.data?.sheet
      answers.value = res.data?.answers || {}
      remainSec.value = res.data?.remainSec || 0
      phase.value = 'answering'
      ElMessage.success('进场成功，祝考试顺利')
    } finally {
      resuming.value = false
    }
  }

  // 保存与自动保存
  const saving = ref(false)
  const doSave = async (manual) => {
    if (!sheet.value || phase.value !== 'answering') return
    saving.value = true
    try {
      const res = await terminalSave({ sheetId: sheet.value.ID, answers: answers.value, version: sheet.value.version })
      sheet.value = res.data
      if (manual) ElMessage.success('已保存')
    } catch (e) {
      if (manual) ElMessage.error('保存失败，请重试')
    } finally {
      saving.value = false
    }
  }
  let autoSaveTimer = null

  // 倒计时
  let tickTimer = null
  const startTick = () => {
    tickTimer = setInterval(() => {
      if (remainSec.value > 0) remainSec.value--
      if (remainSec.value <= 0 && phase.value === 'answering') {
        ElMessage.warning('时间到，自动交卷')
        doSubmit()
      }
    }, 1000)
  }

  // 实操
  const doPractice = async () => {
    if (!practiceTarget.value) { ElMessage.warning('请填写执行目标'); return }
    practicing.value = true
    try {
      const res = await terminalPractice({
        sheetId: sheet.value.ID, questionId: curQ.value.ID,
        channel: practiceChannel.value, target: practiceTarget.value, input: practiceInput.value || curQ.value.stem,
      })
      practiceOutput.value = res.data?.output || res.data?.errMsg || ''
      ElMessage.success(res.data?.ok ? '执行成功，输出已写入答卷' : '执行失败，见输出')
      await doSave(false)
    } finally {
      practicing.value = false
    }
  }

  // 交卷
  const confirmSubmit = () => {
    const unanswered = questions.value.filter((q) => !answered(q.ID)).length
    ElMessageBox.confirm(
      unanswered ? `还有 ${unanswered} 题未作答，确认交卷？` : '确认交卷？', '交卷确认', { type: 'warning' },
    )
      .then(() => doSubmit())
      .catch(() => {})
  }
  const doSubmit = async () => {
    await doSave(false)
    await terminalSubmit({ sheetId: sheet.value.ID })
    phase.value = 'done'
    clearInterval(tickTimer)
    clearInterval(autoSaveTimer)
  }
  const resetAll = () => {
    phase.value = 'lobby'
    sheet.value = null
    answers.value = {}
    questions.value = []
    ticket.value = ''
    doRegister()
  }

  onMounted(() => {
    if (nodeKey.value) doRegister()
    autoSaveTimer = setInterval(() => doSave(false), 30000)
    startTick()
  })
  onUnmounted(() => {
    clearInterval(heartbeatTimer)
    clearInterval(autoSaveTimer)
    clearInterval(tickTimer)
  })
  void saving
</script>
