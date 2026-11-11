<template>
  <div class="ops-card">
    <el-tabs v-model="tab">
      <!-- 身份对接 -->
      <el-tab-pane label="身份对接" name="fed">
        <el-form inline>
          <el-form-item>
            <el-button size="small" @click="loadFederations">刷 新</el-button>
          </el-form-item>
        </el-form>
        <el-table :data="federations" v-loading="fedLoading" stripe size="small" style="margin-bottom: 10px">
          <el-table-column prop="school" label="高校" min-width="120" />
          <el-table-column label="协议" width="110">
            <template #default="{ row }"><el-tag size="small">{{ row.protocol }}</el-tag></template>
          </el-table-column>
          <el-table-column prop="endpoint" label="端点" min-width="220" show-overflow-tooltip />
          <el-table-column label="状态" width="80">
            <template #default="{ row }">
              <el-tag size="small" :type="row.status === '启用' ? 'success' : 'info'">{{ row.status }}</el-tag>
            </template>
          </el-table-column>
        </el-table>
        <el-form inline>
          <el-form-item label="院校"><el-input v-model="syncUserQuery.school" placeholder="院校" style="width: 140px" clearable @keyup.enter="loadSyncUsers" /></el-form-item>
          <el-form-item label="身份">
            <el-select v-model="syncUserQuery.identityStatus" clearable placeholder="全部" style="width: 100px" @change="loadSyncUsers">
              <el-option v-for="s in identityStatuses" :key="s" :label="s" :value="s" />
            </el-select>
          </el-form-item>
        </el-form>
        <el-table :data="syncUsers" v-loading="userLoading" stripe size="small">
          <el-table-column prop="name" label="姓名" width="90" />
          <el-table-column prop="userNo" label="学号/工号" width="110" />
          <el-table-column prop="school" label="院校" min-width="120" show-overflow-tooltip />
          <el-table-column prop="dept" label="院系" min-width="120" show-overflow-tooltip />
          <el-table-column label="状态" width="80">
            <template #default="{ row }"><el-tag size="small" :type="row.identityStatus === '在读' || row.identityStatus === '在职' ? 'success' : 'info'">{{ row.identityStatus }}</el-tag></template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- 权限管理 -->
      <el-tab-pane label="权限管理" name="perm">
        <el-form label-width="80px" size="small" style="max-width: 680px">
          <el-form-item label="互认校验">
            <el-input v-model="trustForm.identity" placeholder="身份（如 上海大学教师）" style="width: 180px" />
            <el-input v-model="trustForm.resource" placeholder="资源（如 教学资源库）" style="width: 180px" />
            <el-button type="primary" size="small" style="margin-left: 8px" :loading="trustLoading" @click="doTrustCheck">校 验</el-button>
          </el-form-item>
          <el-form-item label="矩阵校验">
            <el-input v-model="matrixForm.school" placeholder="院校" style="width: 130px" />
            <el-input v-model="matrixForm.dept" placeholder="部门" style="width: 120px" />
            <el-input v-model="matrixForm.user" placeholder="用户" style="width: 110px" />
            <el-input v-model="matrixForm.scene" placeholder="场景" style="width: 120px" />
            <el-button type="primary" size="small" style="margin-left: 8px" :loading="matrixLoading" @click="doMatrixCheck">校 验</el-button>
          </el-form-item>
        </el-form>
        <el-descriptions v-if="checkResult" :column="2" border size="small" style="margin-bottom: 10px">
          <el-descriptions-item label="结论">
            <el-tag size="small" :type="checkResult.allowed ? 'success' : 'danger'">{{ checkResult.allowed ? '放行' : '拒绝' }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="原因">{{ checkResult.reason }}</el-descriptions-item>
        </el-descriptions>
        <el-divider content-position="left">互认规则库</el-divider>
        <el-table :data="trustRules" v-loading="ruleLoading" stripe size="small">
          <el-table-column prop="homeIdentity" label="本校身份" min-width="130" show-overflow-tooltip />
          <el-table-column prop="targetResource" label="目标资源" min-width="160" show-overflow-tooltip />
          <el-table-column prop="scope" label="互认范围" min-width="140" show-overflow-tooltip />
          <el-table-column label="allowed" width="80">
            <template #default="{ row }">
              <el-tag size="small" :type="row.allowed ? 'success' : 'danger'">{{ row.allowed ? '允许' : '拒绝' }}</el-tag>
            </template>
          </el-table-column>
        </el-table>
        <el-divider content-position="left">角色模板 ×8</el-divider>
        <el-table :data="roleTemplates" stripe size="small">
          <el-table-column prop="roleName" label="角色" min-width="150" />
          <el-table-column label="分类" width="90">
            <template #default="{ row }"><el-tag size="small">{{ row.category }}</el-tag></template>
          </el-table-column>
          <el-table-column prop="description" label="说明" min-width="220" show-overflow-tooltip />
        </el-table>
      </el-tab-pane>

      <!-- 分级扫描 -->
      <el-tab-pane label="分级扫描" name="classify">
        <el-form inline>
          <el-form-item>
            <el-input v-model="classifyInput" placeholder="资源名=内容，每行一条（如：成绩单导出,含身份证号）" style="width: 420px" type="textarea" :rows="3" />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" size="small" :loading="classifyLoading" @click="doClassify">批量扫描</el-button>
          </el-form-item>
          <el-form-item label="分级">
            <el-select v-model="classifyQuery.level" clearable placeholder="全部" style="width: 90px" @change="loadClassify">
              <el-option v-for="l in levels" :key="l" :label="l" :value="l" />
            </el-select>
          </el-form-item>
        </el-form>
        <el-table :data="classifyResults" v-loading="classifyResultLoading" stripe size="small">
          <el-table-column prop="resourceName" label="资源" min-width="200" show-overflow-tooltip />
          <el-table-column label="分类" width="80">
            <template #default="{ row }"><el-tag size="small">{{ row.category }}</el-tag></template>
          </el-table-column>
          <el-table-column label="分级" width="80">
            <template #default="{ row }">
              <el-tag size="small" :type="row.level === '敏感' ? 'danger' : row.level === '内部' ? 'warning' : 'info'">{{ row.level }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="hitRule" label="命中规则" min-width="130" show-overflow-tooltip />
          <el-table-column prop="scannedAt" label="扫描时间" width="140" />
        </el-table>
      </el-tab-pane>

      <!-- 监控看板 -->
      <el-tab-pane label="监控看板" name="monitor">
        <el-form inline>
          <el-form-item><el-button size="small" :loading="fiveLoading" @click="loadFiveDim">刷 新</el-button></el-form-item>
        </el-form>
        <el-descriptions v-if="fiveDim" :column="2" border size="small">
          <el-descriptions-item label="总体健康">
            <el-tag size="small" :type="fiveDim.healthy ? 'success' : 'danger'">{{ fiveDim.healthy ? '健康' : '存在未处置告警' }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="同步用户">{{ fiveDim.identityAuth.syncUsers }}（同步任务 {{ fiveDim.identityAuth.syncTasks24h }}）</el-descriptions-item>
          <el-descriptions-item label="权限变更">分配 {{ fiveDim.permChange.assignments }} · 矩阵授权 {{ fiveDim.permChange.matrixGrants }}</el-descriptions-item>
          <el-descriptions-item label="数据传输">对接 {{ fiveDim.dataTransfer.federations }} · 累计拉取 {{ fiveDim.dataTransfer.pulledTotal }}</el-descriptions-item>
          <el-descriptions-item label="资源使用">{{ fiveDim.resourceUsage.usageWindows }} 个小时窗口 · {{ fiveDim.resourceUsage.coreHours }} 核时</el-descriptions-item>
          <el-descriptions-item label="安全事件">访问告警 {{ fiveDim.securityEvents.accessAlerts }}（未处置 {{ fiveDim.securityEvents.unhandledAlerts }}）· 算力 {{ fiveDim.securityEvents.computeAlerts }}</el-descriptions-item>
        </el-descriptions>
        <el-divider content-position="left">阻断告警（响应 ≤5 分钟口径）</el-divider>
        <el-table :data="alerts" v-loading="alertLoading" stripe size="small">
          <el-table-column prop="userName" label="用户" width="100" />
          <el-table-column prop="school" label="院校" min-width="120" show-overflow-tooltip />
          <el-table-column prop="resource" label="资源" min-width="150" show-overflow-tooltip />
          <el-table-column prop="blockReason" label="阻断原因" min-width="200" show-overflow-tooltip />
          <el-table-column label="处置" width="70">
            <template #default="{ row }">{{ row.handled ? '已处置' : '待处置' }}</template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- 运维工单 -->
      <el-tab-pane label="运维工单" name="ops">
        <el-form inline>
          <el-form-item label="院校"><el-input v-model="opsForm.school" placeholder="报障院校" style="width: 130px" /></el-form-item>
          <el-form-item label="标题"><el-input v-model="opsForm.title" placeholder="故障标题" style="width: 200px" /></el-form-item>
          <el-form-item><el-button type="primary" size="small" :loading="opsCreating" @click="doCreateOps">提 交</el-button></el-form-item>
          <el-form-item><el-button size="small" @click="loadOps">刷 新</el-button></el-form-item>
        </el-form>
        <el-table :data="opsTickets" v-loading="opsLoading" stripe size="small">
          <el-table-column prop="ticketNo" label="工单号" width="120" />
          <el-table-column prop="school" label="院校" min-width="120" show-overflow-tooltip />
          <el-table-column prop="title" label="标题" min-width="160" show-overflow-tooltip />
          <el-table-column label="状态" width="80">
            <template #default="{ row }">
              <el-tag size="small" :type="row.status === '闭环' ? 'success' : 'warning'">{{ row.status }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="assignee" label="承接方" width="110" show-overflow-tooltip />
          <el-table-column label="操作" width="80">
            <template #default="{ row }">
              <el-button size="small" text type="primary" @click="showOpsMessages(row)">消息</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>
    </el-tabs>

    <!-- 协同消息对话框 -->
    <el-dialog v-model="msgVisible" :title="`协同消息 ${currentOpsNo}`" width="560px">
      <el-empty v-if="!opsMessages.length" description="暂无消息" :image-size="60" />
      <el-timeline v-else>
        <el-timeline-item v-for="m in opsMessages" :key="m.id" :timestamp="`${m.sentAt} · ${m.from}`">
          {{ m.content }}
        </el-timeline-item>
      </el-timeline>
    </el-dialog>
  </div>
</template>

<script setup>
  import { ref, reactive, onMounted } from 'vue'
  import { ElMessage } from 'element-plus'
  import {
    getFederations, getSyncUsers,
    getTrustRules, checkTrust, checkMatrix, getRoleTemplates,
    scanClassify, getClassifyResults,
    getFiveDim, getAccessAlerts,
    getOpsTickets, createOpsTicket, getOpsMessages,
  } from '@/plugin/crossauth/api/crossauth'

  const tab = ref('fed')
  const identityStatuses = ['在读', '在职', '离校', '离职']
  const levels = ['公开', '内部', '敏感']

  // 身份对接
  const federations = ref([])
  const fedLoading = ref(false)
  const syncUsers = ref([])
  const userLoading = ref(false)
  const syncUserQuery = reactive({ school: '', identityStatus: '' })

  const loadFederations = async () => {
    fedLoading.value = true
    try {
      const res = await getFederations()
      federations.value = res.data || []
    } finally {
      fedLoading.value = false
    }
  }
  const loadSyncUsers = async () => {
    userLoading.value = true
    try {
      const res = await getSyncUsers({ ...syncUserQuery, pageSize: 30 })
      syncUsers.value = (res.data && res.data.list) || []
    } finally {
      userLoading.value = false
    }
  }

  // 权限管理
  const trustForm = reactive({ identity: '', resource: '' })
  const matrixForm = reactive({ school: '', dept: '', user: '', scene: '' })
  const trustLoading = ref(false)
  const matrixLoading = ref(false)
  const checkResult = ref(null)
  const trustRules = ref([])
  const ruleLoading = ref(false)
  const roleTemplates = ref([])

  const loadTrust = async () => {
    ruleLoading.value = true
    try {
      const [r, t] = await Promise.all([getTrustRules({}), getRoleTemplates({})])
      trustRules.value = r.data || []
      roleTemplates.value = t.data || []
    } finally {
      ruleLoading.value = false
    }
  }
  const doTrustCheck = async () => {
    trustLoading.value = true
    try {
      const res = await checkTrust({ ...trustForm })
      checkResult.value = res.data
    } finally {
      trustLoading.value = false
    }
  }
  const doMatrixCheck = async () => {
    matrixLoading.value = true
    try {
      const res = await checkMatrix({ ...matrixForm })
      checkResult.value = res.data
    } finally {
      matrixLoading.value = false
    }
  }

  // 分级扫描
  const classifyInput = ref('')
  const classifyLoading = ref(false)
  const classifyResults = ref([])
  const classifyResultLoading = ref(false)
  const classifyQuery = reactive({ level: '' })

  const loadClassify = async () => {
    classifyResultLoading.value = true
    try {
      const res = await getClassifyResults({ ...classifyQuery })
      classifyResults.value = res.data || []
    } finally {
      classifyResultLoading.value = false
    }
  }
  const doClassify = async () => {
    const items = classifyInput.value
      .split('\n')
      .map((line) => line.split(','))
      .filter((parts) => parts[0] && parts[0].trim())
      .map((parts) => ({ name: parts[0].trim(), content: (parts[1] || '').trim() }))
    if (!items.length) {
      ElMessage.warning('请输入至少一条「资源名,内容」')
      return
    }
    classifyLoading.value = true
    try {
      const res = await scanClassify({ items })
      ElMessage.success(`扫描 ${res.data.length} 条`)
      loadClassify()
    } finally {
      classifyLoading.value = false
    }
  }

  // 监控看板
  const fiveDim = ref(null)
  const fiveLoading = ref(false)
  const alerts = ref([])
  const alertLoading = ref(false)

  const loadFiveDim = async () => {
    fiveLoading.value = true
    try {
      const [f, a] = await Promise.all([getFiveDim(), getAccessAlerts({ pageSize: 20 })])
      fiveDim.value = f.data
      alerts.value = a.data || []
    } finally {
      fiveLoading.value = false
    }
  }

  // 运维工单
  const opsTickets = ref([])
  const opsLoading = ref(false)
  const opsForm = reactive({ school: '', title: '' })
  const opsCreating = ref(false)
  const msgVisible = ref(false)
  const opsMessages = ref([])
  const currentOpsNo = ref('')

  const loadOps = async () => {
    opsLoading.value = true
    try {
      const res = await getOpsTickets({})
      opsTickets.value = res.data || []
    } finally {
      opsLoading.value = false
    }
  }
  const doCreateOps = async () => {
    opsCreating.value = true
    try {
      const res = await createOpsTicket({ ...opsForm })
      ElMessage.success(`已提交 ${res.data.ticketNo}`)
      opsForm.title = ''
      loadOps()
    } finally {
      opsCreating.value = false
    }
  }
  const showOpsMessages = async (row) => {
    currentOpsNo.value = row.ticketNo
    const res = await getOpsMessages({ ticketId: row.id })
    opsMessages.value = res.data || []
    msgVisible.value = true
  }

  onMounted(() => {
    loadFederations()
    loadSyncUsers()
    loadTrust()
    loadClassify()
    loadFiveDim()
    loadOps()
  })
</script>
