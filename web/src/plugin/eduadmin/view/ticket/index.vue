<template>
  <div class="ops-card">
    <el-tabs v-model="tab">
      <!-- 提交 -->
      <el-tab-pane label="提交工单" name="submit">
        <el-form label-width="90px" style="max-width: 640px">
          <el-form-item label="渠道">
            <el-radio-group v-model="form.channel">
              <el-radio value="网页">网页</el-radio>
              <el-radio value="企微">企微</el-radio>
              <el-radio value="小程序">小程序</el-radio>
            </el-radio-group>
          </el-form-item>
          <el-form-item label="诉求类型">
            <el-select v-model="form.demandType" style="width: 200px" allow-create filterable>
              <el-option v-for="t in demandTypes" :key="t" :label="t" :value="t" />
            </el-select>
          </el-form-item>
          <el-form-item label="时限类别">
            <el-select v-model="form.deadlineClass" style="width: 200px">
              <el-option label="简单（24h）" value="简单24h" />
              <el-option label="一般（3 工作日）" value="一般3工作日" />
              <el-option label="复杂（7 工作日）" value="复杂7工作日" />
            </el-select>
          </el-form-item>
          <el-form-item label="诉求内容">
            <el-input v-model="form.content" type="textarea" :rows="4" placeholder="请描述您的诉求…" />
          </el-form-item>
          <el-form-item label="附件">
            <el-input v-model="form.attachmentMeta" placeholder="附件元数据（名称/大小/哈希，可空）" />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" :loading="submitting" @click="doSubmit">提 交</el-button>
          </el-form-item>
        </el-form>
      </el-tab-pane>

      <!-- 工单列表 -->
      <el-tab-pane label="工单列表" name="list">
        <el-form inline>
          <el-form-item label="状态">
            <el-select v-model="listQuery.status" clearable placeholder="全部" style="width: 110px" @change="loadTickets">
              <el-option v-for="s in statuses" :key="s" :label="s" :value="s" />
            </el-select>
          </el-form-item>
          <el-form-item label="诉求">
            <el-input v-model="listQuery.demandType" placeholder="类型关键字" style="width: 130px" clearable @keyup.enter="loadTickets" />
          </el-form-item>
          <el-form-item>
            <el-button size="small" @click="loadTickets">刷 新</el-button>
          </el-form-item>
        </el-form>
        <el-table :data="tickets" v-loading="ticketLoading" stripe size="small">
          <el-table-column prop="ticketNo" label="编号" width="150" />
          <el-table-column label="渠道" width="80">
            <template #default="{ row }"><el-tag size="small">{{ row.channel }}</el-tag></template>
          </el-table-column>
          <el-table-column prop="demandType" label="诉求类型" width="100" />
          <el-table-column prop="content" label="内容" min-width="200" show-overflow-tooltip />
          <el-table-column prop="deadlineClass" label="时限" width="110" />
          <el-table-column label="状态" width="80">
            <template #default="{ row }">
              <el-tag size="small" :type="row.status === '归档' ? 'info' : row.status === '办结' || row.status === '评价' ? 'success' : 'warning'">{{ row.status }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="180">
            <template #default="{ row }">
              <el-button size="small" text type="primary" @click="showTimeline(row)">进度</el-button>
              <el-button size="small" text type="warning" @click="doAutoDispatch(row)">智能分派</el-button>
              <el-button size="small" text type="success" @click="doAdvance(row)">推进</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- 统计 -->
      <el-tab-pane label="统计" name="stats">
        <el-form inline>
          <el-form-item><el-button size="small" @click="loadStats">刷 新</el-button></el-form-item>
          <el-form-item><el-button size="small" type="primary" @click="doExport">导出 CSV</el-button></el-form-item>
        </el-form>
        <el-descriptions v-if="stats" :column="4" border size="small">
          <el-descriptions-item label="工单总量">{{ stats.total }}</el-descriptions-item>
          <el-descriptions-item label="已办结">{{ stats.closed }}</el-descriptions-item>
          <el-descriptions-item label="办结率">{{ (stats.closeRate * 100).toFixed(0) }}%</el-descriptions-item>
          <el-descriptions-item label="超时率">{{ (stats.overdueRate * 100).toFixed(0) }}%</el-descriptions-item>
          <el-descriptions-item label="临近督办">{{ stats.urgedNear }}</el-descriptions-item>
          <el-descriptions-item label="超时督办">{{ stats.urgedOver }}</el-descriptions-item>
          <el-descriptions-item label="评价数">{{ stats.reviews }}</el-descriptions-item>
          <el-descriptions-item label="满意率">{{ (stats.goodRate * 100).toFixed(0) }}%</el-descriptions-item>
        </el-descriptions>
        <el-row :gutter="10" style="margin-top: 10px">
          <el-col :span="12">
            <el-card shadow="never" header="高频诉求类型 Top5" size="small">
              <el-table :data="stats ? stats.topTypes : []" size="small">
                <el-table-column prop="type" label="类型" min-width="120" />
                <el-table-column prop="count" label="数量" width="80" />
              </el-table>
            </el-card>
          </el-col>
          <el-col :span="12">
            <el-card shadow="never" header="各部门办理效率" size="small">
              <el-table :data="stats ? stats.deptEfficiency : []" size="small">
                <el-table-column prop="dept" label="部门" min-width="140" />
                <el-table-column prop="dispatch" label="分派数" width="90" />
              </el-table>
            </el-card>
          </el-col>
        </el-row>
      </el-tab-pane>
    </el-tabs>

    <!-- 进度时间线对话框 -->
    <el-dialog v-model="timelineVisible" :title="`工单进度 ${currentTicketNo}`" width="560px">
      <el-empty v-if="!timelines.length" description="暂无进展记录" :image-size="60" />
      <el-timeline v-else>
        <el-timeline-item v-for="t in timelines" :key="t.id" :timestamp="`${t.occurredAt} · ${t.dept} / ${t.operator}`">
          {{ t.content }}
        </el-timeline-item>
      </el-timeline>
    </el-dialog>
  </div>
</template>

<script setup>
  import { ref, reactive, onMounted } from 'vue'
  import { ElMessage } from 'element-plus'
  import {
    getTickets, createTicket, advanceTicket, autoDispatch,
    getTimelines, getTicketStats, exportTicketStats,
  } from '@/plugin/eduadmin/api/eduadmin'

  const tab = ref('submit')
  const demandTypes = ['教务', '后勤', '财务', '信息化', '人事']
  const statuses = ['受理', '分派', '办理', '办结', '评价', '归档']

  // 提交
  const form = reactive({ channel: '网页', demandType: '教务', deadlineClass: '一般3工作日', content: '', attachmentMeta: '' })
  const submitting = ref(false)

  const doSubmit = async () => {
    submitting.value = true
    try {
      const res = await createTicket(form)
      ElMessage.success(`已受理 ${res.data.ticketNo}`)
      form.content = ''
      form.attachmentMeta = ''
      loadTickets()
    } finally {
      submitting.value = false
    }
  }

  // 工单列表
  const tickets = ref([])
  const ticketLoading = ref(false)
  const listQuery = reactive({ status: '', demandType: '' })

  const loadTickets = async () => {
    ticketLoading.value = true
    try {
      const res = await getTickets({ ...listQuery, pageSize: 30 })
      tickets.value = (res.data && res.data.list) || []
    } finally {
      ticketLoading.value = false
    }
  }
  const doAdvance = async (row) => {
    await advanceTicket({ id: row.id })
    ElMessage.success(`${row.ticketNo} → 已推进`)
    loadTickets()
  }
  const doAutoDispatch = async (row) => {
    await autoDispatch({ ticketId: row.id })
    ElMessage.success(`${row.ticketNo} 分派完成`)
    loadTickets()
  }

  // 时间线
  const timelineVisible = ref(false)
  const timelines = ref([])
  const currentTicketNo = ref('')
  const showTimeline = async (row) => {
    currentTicketNo.value = row.ticketNo
    const res = await getTimelines({ ticketId: row.id })
    timelines.value = res.data || []
    timelineVisible.value = true
  }

  // 统计
  const stats = ref(null)
  const loadStats = async () => {
    const res = await getTicketStats()
    stats.value = res.data
  }
  const doExport = async () => {
    const res = await exportTicketStats()
    ElMessage.info('CSV 已生成，见返回数据')
    console.log(res.data)
  }

  onMounted(() => {
    loadTickets()
    loadStats()
  })
</script>
