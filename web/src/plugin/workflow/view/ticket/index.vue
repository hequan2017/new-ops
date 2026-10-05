<template>
  <div>
    <div class="ops-card">
      <el-tabs v-model="tab">
        <el-tab-pane label="我的工单" name="tickets">
          <el-form inline>
            <el-form-item label="状态">
              <el-select v-model="ticketStatus" clearable style="width: 120px" @change="loadTickets">
                <el-option v-for="s in ['进行中', '已完成', '已驳回', '已取消']" :key="s" :label="s" :value="s" />
              </el-select>
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="openStart">发起工单</el-button>
              <el-button @click="loadTickets">刷 新</el-button>
            </el-form-item>
          </el-form>
          <el-table :data="tickets" v-loading="ticketsLoading" stripe>
            <el-table-column prop="ID" label="ID" width="60" />
            <el-table-column prop="title" label="标题" min-width="180" show-overflow-tooltip />
            <el-table-column prop="definitionName" label="类型" width="130" />
            <el-table-column prop="state" label="当前节点" width="110" />
            <el-table-column label="状态" width="90">
              <template #default="{ row }">
                <el-tag :type="statusType(row.status)" size="small">{{ row.status }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="creator" label="发起人" width="90" />
            <el-table-column label="发起时间" width="160">
              <template #default="{ row }">{{ (row.createdAt || '').replace('T', ' ').slice(0, 19) }}</template>
            </el-table-column>
            <el-table-column label="操作" width="200" fixed="right">
              <template #default="{ row }">
                <el-button link type="primary" @click="openLogs(row)">流转</el-button>
                <template v-if="row.status === '进行中'">
                  <el-button link type="success" @click="doAction(row, 'approve', '同意')">通过</el-button>
                  <el-button link type="danger" @click="doAction(row, 'reject', '驳回')">驳回</el-button>
                  <el-button link @click="doAction(row, 'cancel', '撤回')">撤回</el-button>
                </template>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <el-tab-pane label="工单定义" name="definitions">
          <div class="ops-btn-list">
            <el-button type="primary" @click="openDefDialog()">新建定义</el-button>
            <el-button @click="loadDefs">刷 新</el-button>
          </div>
          <el-table :data="defs" v-loading="defsLoading" stripe>
            <el-table-column prop="ID" label="ID" width="60" />
            <el-table-column prop="name" label="名称" min-width="140" />
            <el-table-column prop="startState" label="起始状态" width="100" />
            <el-table-column label="节点" min-width="200">
              <template #default="{ row }">
                <el-tag v-for="st in parseStates(row.states)" :key="st.name" size="small" style="margin-right: 4px"
                  :type="st.isApproval ? 'warning' : st.isFinal ? 'success' : 'info'">
                  {{ st.name }}{{ st.isApproval ? '审' : '' }}{{ st.isFinal ? '终' : '' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="130" fixed="right">
              <template #default="{ row }">
                <el-button link type="primary" @click="openDefDialog(row)">编辑</el-button>
                <el-button link type="danger" @click="onDeleteDef(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
      </el-tabs>
    </div>

    <el-dialog v-model="startVisible" title="发起工单" width="520px">
      <el-form label-width="80px">
        <el-form-item label="类型">
          <el-select v-model="startForm.definitionId" style="width: 100%">
            <el-option v-for="d in defs" :key="d.ID" :label="d.name" :value="d.ID" />
          </el-select>
        </el-form-item>
        <el-form-item label="标题">
          <el-input v-model="startForm.title" placeholder="如 发布 user-service v1.2" />
        </el-form-item>
        <el-form-item label="业务参数">
          <el-input v-model="startForm.params" type="textarea" :rows="3" placeholder='JSON，如 {"pipeline":"deploy-web"}' />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="startVisible = false">取 消</el-button>
        <el-button type="primary" @click="submitStart">发 起</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="defVisible" :title="defForm.ID ? '编辑定义' : '新建定义'" width="680px">
      <el-form label-width="90px">
        <el-form-item label="名称">
          <el-input v-model="defForm.name" placeholder="如 发布审批" />
        </el-form-item>
        <el-form-item label="起始状态">
          <el-input v-model="defForm.startState" placeholder="如 submitted" style="width: 200px" />
        </el-form-item>
        <el-form-item label="状态节点">
          <el-input v-model="defForm.states" type="textarea" :rows="4"
            placeholder='[{"name":"submitted"},{"name":"approving","isApproval":true},{"name":"archived","isFinal":true}]' />
        </el-form-item>
        <el-form-item label="迁移表">
          <el-input v-model="defForm.transitions" type="textarea" :rows="4"
            placeholder='[{"from":"submitted","action":"submit","to":"approving"},{"from":"approving","action":"approve","to":"archived"},{"from":"approving","action":"reject","to":"archived"}]' />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="defVisible = false">取 消</el-button>
        <el-button type="primary" @click="submitDef">保 存</el-button>
      </template>
    </el-dialog>

    <el-drawer v-model="logsVisible" :title="`流转记录 · ${logsTitle}`" size="45%">
      <el-timeline style="padding-left: 8px">
        <el-timeline-item v-for="l in logsList" :key="l.ID" :timestamp="(l.createdAt || '').replace('T', ' ').slice(0, 19)">
          <b>{{ l.action }}</b> {{ l.fromState || '∅' }} → {{ l.toState }}
          <span class="ops-text-muted">（{{ l.operator }}）</span>
          <div v-if="l.comment" class="ops-text-muted">{{ l.comment }}</div>
        </el-timeline-item>
      </el-timeline>
    </el-drawer>
  </div>
</template>

<script setup>
  import { ref, reactive, onMounted } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import {
    createDefinition, updateDefinition, deleteDefinition, getDefinitions,
    startInstance, submitAction, getInstances, getInstanceLogs
  } from '@/plugin/workflow/api/workflow'

  defineOptions({ name: 'workflowTicket' })

  const tab = ref('tickets')
  const tickets = ref([])
  const ticketsLoading = ref(false)
  const ticketStatus = ref('')

  const statusType = (s) =>
    ({ 进行中: 'primary', 已完成: 'success', 已驳回: 'danger', 已取消: 'info' }[s] || 'info')

  const loadTickets = async () => {
    ticketsLoading.value = true
    try {
      const res = await getInstances({ page: 1, pageSize: 50 }, { status: ticketStatus.value || undefined })
      if (res.code === 0) tickets.value = res.data.list || []
    } finally {
      ticketsLoading.value = false
    }
  }

  const doAction = (row, action, label) => {
    ElMessageBox.confirm(`确定对工单「${row.title}」执行【${label}】吗？`, '确认', {
      confirmButtonText: label, cancelButtonText: '取消', type: action === 'approve' ? 'success' : 'warning'
    }).then(async () => {
      const res = await submitAction({ id: row.ID, action, comment: label })
      if (res.code === 0) {
        ElMessage.success(res.msg)
        loadTickets()
      }
    })
  }

  // 定义页签
  const defs = ref([])
  const defsLoading = ref(false)

  const parseStates = (s) => {
    try {
      return JSON.parse(s || '[]')
    } catch {
      return []
    }
  }

  const loadDefs = async () => {
    defsLoading.value = true
    try {
      const res = await getDefinitions({})
      if (res.code === 0) defs.value = res.data || []
    } finally {
      defsLoading.value = false
    }
  }

  const defVisible = ref(false)
  const emptyDef = () => ({
    ID: 0, name: '', startState: 'submitted',
    states: '[{"name":"submitted"},{"name":"approving","isApproval":true},{"name":"archived","isFinal":true}]',
    transitions: '[{"from":"submitted","action":"submit","to":"approving"},{"from":"approving","action":"approve","to":"archived"},{"from":"approving","action":"reject","to":"archived"}]'
  })
  const defForm = reactive(emptyDef())

  const openDefDialog = (row) => {
    Object.assign(defForm, emptyDef())
    if (row) Object.assign(defForm, row)
    defVisible.value = true
  }

  const submitDef = async () => {
    const res = defForm.ID ? await updateDefinition(defForm) : await createDefinition(defForm)
    if (res.code === 0) {
      ElMessage.success(res.msg)
      defVisible.value = false
      loadDefs()
    }
  }

  const onDeleteDef = (row) => {
    ElMessageBox.confirm(`确定删除定义「${row.name}」吗？`, '提示', {
      confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning'
    }).then(async () => {
      const res = await deleteDefinition({ id: row.ID })
      if (res.code === 0) {
        ElMessage.success('已删除')
        loadDefs()
      }
    })
  }

  // 发起
  const startVisible = ref(false)
  const startForm = reactive({ definitionId: null, title: '', params: '' })

  const openStart = () => {
    startForm.definitionId = defs.value.length ? defs.value[0].ID : null
    startForm.title = ''
    startForm.params = ''
    startVisible.value = true
  }

  const submitStart = async () => {
    if (!startForm.definitionId || !startForm.title.trim()) {
      ElMessage.warning('请选择类型并填写标题')
      return
    }
    const res = await startInstance({ ...startForm })
    if (res.code === 0) {
      ElMessage.success(`工单 #${res.data.ID} 已发起`)
      startVisible.value = false
      loadTickets()
    }
  }

  // 流转记录
  const logsVisible = ref(false)
  const logsList = ref([])
  const logsTitle = ref('')
  const openLogs = async (row) => {
    logsTitle.value = row.title
    const res = await getInstanceLogs({ id: row.ID })
    if (res.code === 0) {
      logsList.value = res.data || []
      logsVisible.value = true
    }
  }

  onMounted(() => {
    loadDefs()
    loadTickets()
  })
</script>
