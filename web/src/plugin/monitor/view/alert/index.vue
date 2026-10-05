<template>
  <div>
    <div class="ops-card">
      <el-tabs v-model="tab">
        <el-tab-pane label="告警规则" name="rules">
          <div class="ops-btn-list">
            <el-button type="primary" @click="openDialog()">新建规则</el-button>
            <el-button @click="loadRules">刷 新</el-button>
          </div>
          <el-table :data="rules" v-loading="rulesLoading" stripe>
            <el-table-column prop="ID" label="ID" width="55" />
            <el-table-column label="资产" width="120">
              <template #default="{ row }">{{ assetLabel(row.assetId) }}</template>
            </el-table-column>
            <el-table-column label="类型" width="80">
              <template #default="{ row }">
                <el-tag size="small" :type="row.type === 'port' ? 'warning' : 'primary'">{{ row.type }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="条件" min-width="180">
              <template #default="{ row }">
                <span v-if="row.type === 'metric'">
                  {{ row.metricName }} {{ row.operator }} {{ row.threshold }}，连续 {{ row.duration || 1 }} 次
                </span>
                <span v-else>端口 {{ row.port }} 可达性</span>
              </template>
            </el-table-column>
            <el-table-column label="静默" width="80">
              <template #default="{ row }">{{ row.silenceMin || 30 }}min</template>
            </el-table-column>
            <el-table-column label="钉钉" width="70">
              <template #default="{ row }">
                <el-tag v-if="row.webhookUrl" type="success" size="small">已配</el-tag>
                <span v-else>—</span>
              </template>
            </el-table-column>
            <el-table-column label="启用" width="70">
              <template #default="{ row }">
                <el-switch :model-value="row.enabled" @change="toggleRule(row, $event)" />
              </template>
            </el-table-column>
            <el-table-column label="操作" width="130" fixed="right">
              <template #default="{ row }">
                <el-button link type="primary" @click="openDialog(row)">编辑</el-button>
                <el-button link type="danger" @click="onDelete(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <el-tab-pane label="告警事件" name="events">
          <el-table :data="events" v-loading="eventsLoading" stripe>
            <el-table-column prop="ID" label="ID" width="55" />
            <el-table-column prop="summary" label="摘要" min-width="280" show-overflow-tooltip />
            <el-table-column label="触发时间" width="160">
              <template #default="{ row }">{{ (row.firedAt || '').replace('T', ' ').slice(0, 19) }}</template>
            </el-table-column>
            <el-table-column label="恢复时间" width="160">
              <template #default="{ row }">{{ (row.resolvedAt || '未恢复').toString().replace('T', ' ').slice(0, 19) }}</template>
            </el-table-column>
            <el-table-column label="通知" width="70">
              <template #default="{ row }">
                <el-tag v-if="row.notified" type="success" size="small">已推</el-tag>
                <span v-else>—</span>
              </template>
            </el-table-column>
          </el-table>
          <div class="ops-pagination">
            <el-pagination
              :current-page="page"
              :page-size="pageSize"
              :total="total"
              layout="total, prev, pager, next"
              @current-change="loadEvents"
            />
          </div>
        </el-tab-pane>
      </el-tabs>
    </div>

    <el-dialog v-model="dialogVisible" :title="form.ID ? '编辑规则' : '新建规则'" width="560px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="90px">
        <el-form-item label="资产" prop="assetId">
          <el-select v-model="form.assetId" filterable style="width: 100%">
            <el-option v-for="h in hostOptions" :key="h.ID" :label="`${h.hostname}（${h.ip}）`" :value="h.ID" />
          </el-select>
        </el-form-item>
        <el-form-item label="类型">
          <el-radio-group v-model="form.type">
            <el-radio value="metric">指标阈值</el-radio>
            <el-radio value="port">端口探活</el-radio>
          </el-radio-group>
        </el-form-item>
        <template v-if="form.type === 'metric'">
          <el-form-item label="指标">
            <el-select v-model="form.metricName" style="width: 200px">
              <el-option v-for="m in ['cpu_percent', 'mem_percent', 'disk_percent', 'load1']" :key="m" :label="m" :value="m" />
            </el-select>
          </el-form-item>
          <el-form-item label="条件">
            <div style="display: flex; gap: 8px">
              <el-select v-model="form.operator" style="width: 80px">
                <el-option label=">" value=">" />
                <el-option label="<" value="<" />
              </el-select>
              <el-input-number v-model="form.threshold" :min="0" :max="100000" style="flex: 1" />
              连续
              <el-input-number v-model="form.duration" :min="1" :max="10" style="width: 90px" /> 次
            </div>
          </el-form-item>
        </template>
        <el-form-item v-else label="端口">
          <el-input-number v-model="form.port" :min="1" :max="65535" style="width: 160px" />
        </el-form-item>
        <el-form-item label="静默分钟">
          <el-input-number v-model="form.silenceMin" :min="1" :max="1440" style="width: 140px" />
        </el-form-item>
        <el-form-item label="钉钉Webhook">
          <el-input v-model="form.webhookUrl" placeholder="https://oapi.dingtalk.com/robot/send?access_token=…（选配）" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.notes" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取 消</el-button>
        <el-button type="primary" @click="submitForm">确 定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
  import { ref, reactive, onMounted } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import {
    createAlertRule, updateAlertRule, deleteAlertRule, getAlertRules, getAlertEvents
  } from '@/plugin/monitor/api/alert'
  import { getAssetHostList } from '@/plugin/asset/api/assetHost'

  defineOptions({ name: 'monitorAlert' })

  const tab = ref('rules')
  const rules = ref([])
  const rulesLoading = ref(false)
  const hostOptions = ref([])

  const assetLabel = (id) => {
    const h = hostOptions.value.find((x) => x.ID === id)
    return h ? `${h.hostname}（${h.ip}）` : `#${id}`
  }

  const loadRules = async () => {
    rulesLoading.value = true
    try {
      const res = await getAlertRules()
      if (res.code === 0) rules.value = res.data || []
    } finally {
      rulesLoading.value = false
    }
  }

  const toggleRule = async (row, enabled) => {
    const res = await updateAlertRule({ ...row, enabled })
    if (res.code === 0) {
      row.enabled = enabled
      ElMessage.success(enabled ? '已启用' : '已停用')
    }
  }

  const dialogVisible = ref(false)
  const formRef = ref(null)
  const emptyForm = () => ({
    ID: 0, assetId: null, type: 'metric', metricName: 'cpu_percent',
    operator: '>', threshold: 80, duration: 1, port: 22,
    silenceMin: 30, webhookUrl: '', notes: ''
  })
  const form = reactive(emptyForm())
  const formRules = { assetId: [{ required: true, message: '请选择资产', trigger: 'change' }] }

  const openDialog = (row) => {
    Object.assign(form, emptyForm())
    if (row) Object.assign(form, row)
    dialogVisible.value = true
  }

  const submitForm = () => {
    formRef.value.validate(async (valid) => {
      if (!valid) return
      const payload = { ...form, webhookUrl: form.webhookUrl || '' }
      const res = form.ID ? await updateAlertRule(payload) : await createAlertRule(payload)
      if (res.code === 0) {
        ElMessage.success(res.msg)
        dialogVisible.value = false
        loadRules()
      }
    })
  }

  const onDelete = (row) => {
    ElMessageBox.confirm('确定删除该告警规则吗？', '提示', {
      confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning'
    }).then(async () => {
      const res = await deleteAlertRule({ id: row.ID })
      if (res.code === 0) {
        ElMessage.success('已删除')
        loadRules()
      }
    })
  }

  // 事件页签
  const events = ref([])
  const eventsLoading = ref(false)
  const page = ref(1)
  const pageSize = 15
  const total = ref(0)

  const loadEvents = async (p) => {
    if (p) page.value = p
    eventsLoading.value = true
    try {
      const res = await getAlertEvents({ page: page.value, pageSize })
      if (res.code === 0) {
        events.value = res.data.list || []
        total.value = res.data.total || 0
      }
    } finally {
      eventsLoading.value = false
    }
  }

  onMounted(async () => {
    const h = await getAssetHostList({ page: 1, pageSize: 500 })
    if (h.code === 0) hostOptions.value = h.data.list || []
    loadRules()
    loadEvents()
  })
</script>
