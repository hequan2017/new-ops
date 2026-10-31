<template>
  <div>
    <div class="ops-card">
      <el-form label-width="90px">
        <el-row :gutter="12">
          <el-col :span="16">
            <el-form-item label="目标主机">
              <el-select
                v-model="form.hostIds"
                multiple
                filterable
                placeholder="选择主机（按数据权限过滤）"
                style="width: 100%"
              >
                <el-option
                  v-for="h in hostOptions"
                  :key="h.ID"
                  :label="`${h.hostname}（${h.ip}）`"
                  :value="h.ID"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="统一凭据">
              <el-select v-model="form.credentialId" clearable placeholder="留空用主机绑定凭据" style="width: 100%">
                <el-option
                  v-for="c in credOptions"
                  :key="c.ID"
                  :label="`${c.name}（${c.username || 'SSH'}）`"
                  :value="c.ID"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="脚本库">
              <el-select v-model="form.scriptId" clearable filterable placeholder="选脚本填充命令（优先于手输）" style="width: 100%" @change="onScriptPick">
                <el-option
                  v-for="s in scriptOptions"
                  :key="s.ID"
                  :label="`${s.name}（${s.language} v${s.version}）`"
                  :value="s.ID"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="变量组">
              <el-select v-model="form.variableGroupId" clearable placeholder="留空则按主机所在资产组自动注入变量" style="width: 100%">
                <el-option v-for="v in varGroupOptions" :key="v.ID" :label="v.name" :value="v.ID" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="命令">
              <el-input
                v-model="form.command"
                type="textarea"
                :rows="3"
                :placeholder="form.scriptId ? '已选脚本，将以脚本内容执行（可再选变量组渲染）' : '如：df -h || uptime'"
              />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="并发上限">
              <el-input-number v-model="form.concurrency" :min="1" :max="50" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="超时(秒)">
              <el-input-number v-model="form.timeoutSec" :min="5" :max="600" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="执行后端">
              <el-radio-group v-model="form.execBackend">
                <el-radio-button value="ssh">SSH 直连</el-radio-button>
                <el-radio-button value="agent">Agent 通道</el-radio-button>
              </el-radio-group>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label-width="0">
              <el-button type="primary" :loading="submitting" @click="onSubmitExec">
                批量执行（{{ form.hostIds.length }} 台）
              </el-button>
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
    </div>

    <div class="ops-card" style="margin-top: 12px">
      <div class="file-card-title">文件分发（Agent 通道）</div>
      <el-form label-width="110px">
        <el-row :gutter="12">
          <el-col :span="8">
            <el-form-item label="Server 文件">
              <el-input v-model="fileForm.localPath" placeholder="server 本地绝对路径，如 /opt/pkg/app.tar.gz" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="目标路径">
              <el-input v-model="fileForm.remotePath" placeholder="目标机绝对路径，如 /tmp/app.tar.gz" />
            </el-form-item>
          </el-col>
          <el-col :span="4">
            <el-form-item label="权限">
              <el-input v-model="fileForm.mode" placeholder="0644" />
            </el-form-item>
          </el-col>
          <el-col :span="4">
            <el-form-item label-width="0">
              <el-button type="primary" :loading="filePushing" :disabled="!fileHostIds.length" @click="onPushFile">
                分发（{{ fileHostIds.length }} 台）
              </el-button>
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="目标主机">
              <el-select
                v-model="fileHostIds"
                multiple
                filterable
                placeholder="仅列出 Agent 在线的主机（勾选后经 WS 通道分发）"
                style="width: 100%"
              >
                <el-option v-for="h in agentOnlineHosts" :key="h.ID" :label="`${h.hostname}（${h.ip}）`" :value="h.ID" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>

      <el-table v-if="fileResults.length" :data="fileResults" stripe size="small">
        <el-table-column prop="hostname" label="主机" width="200" show-overflow-tooltip />
        <el-table-column label="结果" width="90">
          <template #default="{ row }">
            <el-tag :type="row.ok ? 'success' : 'danger'" size="small">{{ row.ok ? '成功' : '失败' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="耗时" width="100">
          <template #default="{ row }">{{ row.elapsedMs }}ms</template>
        </el-table-column>
        <el-table-column label="详情">
          <template #default="{ row }">
            <span :class="row.ok ? 'ops-text-muted' : 'ops-text-danger'">{{ row.detail }}</span>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <div class="ops-card" style="margin-top: 12px">
      <el-form inline>
        <el-form-item label="状态">
          <el-select v-model="searchStatus" clearable style="width: 140px" @change="onSearch">
            <el-option v-for="s in ['执行中', '已完成', '已取消']" :key="s" :label="s" :value="s" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button :loading="listLoading" @click="getList">刷 新</el-button>
        </el-form-item>
      </el-form>

      <el-table
        :data="tableData"
        v-loading="listLoading"
        stripe
        border
        @header-dragend="onHeaderDragend"
      >
        <el-table-column prop="ID" label="批次" :width="colW('ID', 70)" />
        <el-table-column prop="command" label="命令" :width="colW('command')" min-width="200" show-overflow-tooltip />
        <el-table-column label="进度" :width="colW('进度', 130)">
          <template #default="{ row }">
            <span :class="row.failed > 0 ? 'ops-text-danger' : ''">
              {{ row.success + row.failed }}/{{ row.total }}
            </span>
            <span class="ops-text-muted">（成功 {{ row.success }} / 失败 {{ row.failed }}）</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" :width="colW('状态', 90)">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="concurrency" label="并发" :width="colW('concurrency', 60)" />
        <el-table-column prop="timeoutSec" label="超时s" :width="colW('timeoutSec', 70)" />
        <el-table-column prop="operator" label="操作人" :width="colW('operator', 100)" />
        <el-table-column label="开始时间" :width="colW('开始时间', 160)">
          <template #default="{ row }">
            {{ fmtTime(row.startedAt) }}
          </template>
        </el-table-column>
        <el-table-column label="耗时" width="90">
          <template #default="{ row }">
            <span v-if="row.finishedAt && row.startedAt" :class="row.failed > 0 ? 'ops-text-danger' : ''">
              {{ batchDuration(row) }}
            </span>
            <span v-else style="color: var(--el-text-color-placeholder)">—</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openDetail(row)">详情</el-button>
            <el-button v-if="row.status === '执行中'" link type="danger" @click="onCancel(row)">取消</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="ops-pagination">
        <el-pagination
          :current-page="page"
          :page-size="pageSize"
          :page-sizes="[10, 20, 50]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @current-change="handleCurrentChange"
          @size-change="handleSizeChange"
        />
      </div>
    </div>

    <el-drawer v-model="detailVisible" :title="`批次 #${detailRecord?.ID || ''} 执行结果`" size="65%">
      <el-descriptions :column="3" size="small" border style="margin-bottom: 12px">
        <el-descriptions-item label="命令" :span="3">{{ detailRecord?.command }}</el-descriptions-item>
        <el-descriptions-item label="状态">{{ detailRecord?.status }}</el-descriptions-item>
        <el-descriptions-item label="执行后端">{{ detailRecord?.execBackend || 'ssh' }}</el-descriptions-item>
        <el-descriptions-item label="成功">{{ detailRecord?.success }}</el-descriptions-item>
        <el-descriptions-item label="失败">{{ detailRecord?.failed }}</el-descriptions-item>
      </el-descriptions>
      <div style="margin-bottom: 8px">
        <span style="font-size: 13px; color: var(--el-text-color-secondary); margin-right: 8px">结果过滤</span>
        <el-select
          v-model="detailStatusFilter"
          size="small"
          style="width: 120px"
          placeholder="全部"
          @change="reloadDetailResults"
        >
          <el-option label="全部" value="" />
          <el-option v-for="s in ['成功', '失败', '超时', '取消', '跳过']" :key="s" :label="s" :value="s" />
        </el-select>
      </div>
      <el-table :data="detailResults" stripe>
        <el-table-column prop="hostname" label="主机" width="140" show-overflow-tooltip />
        <el-table-column prop="ip" label="IP" width="130" />
        <el-table-column label="结果" width="80">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="耗时" width="90">
          <template #default="{ row }">{{ row.elapsedMs }}ms</template>
        </el-table-column>
        <el-table-column label="输出">
          <template #default="{ row }">
            <el-collapse v-if="row.output || row.error">
              <el-collapse-item name="1">
                <template #title>
                  <span>查看</span>
                  <el-button
                    link
                    type="primary"
                    size="small"
                    style="margin-left: 8px"
                    @click.stop="copyResultOutput(row)"
                  >复制</el-button>
                </template>
                <pre class="ops-pre">{{ row.error ? 'ERROR: ' + row.error : row.output }}</pre>
              </el-collapse-item>
            </el-collapse>
            <span v-else style="color: var(--el-text-color-placeholder); font-size: 12px">（无输出）</span>
          </template>
        </el-table-column>
      </el-table>
    </el-drawer>
  </div>
</template>

<script setup>
  import { ref, reactive, computed, onMounted, onUnmounted } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { createBatchExec, cancelBatchExec, getBatchList, getBatchDetail } from '@/plugin/job/api/batchExec'
  import { getScriptList, getVarGroupList } from '@/plugin/job/api/jobScript'
  import { useColWidthMemory } from '@/hooks/useColWidth'
  import { fmtTime } from '@/utils/format'

  // 批次列表列宽拖拽记忆
  const { colW, onHeaderDragend } = useColWidthMemory('colWidth:jobBatch')

  // 批次耗时（结束-开始，秒级显示；<60s 显示 Ns，≥60s N 分 N 秒）
  const batchDuration = (row) => {
    const ms = new Date(row.finishedAt).getTime() - new Date(row.startedAt).getTime()
    if (!(ms > 0)) return '—'
    const s = Math.round(ms / 1000)
    if (s < 60) return `${s}s`
    return `${Math.floor(s / 60)} 分 ${s % 60} 秒`
  }
  import { useHostOptions } from '@/plugin/asset/composables/useHostOptions'
  import { getCredentialList } from '@/plugin/asset/api/credential'
  import { getAgentInstanceList, pushAgentFile } from '@/plugin/agent/api/agentInstance'

  defineOptions({ name: 'jobBatchExec' })

  // 发起表单
  const form = reactive({
    hostIds: [],
    command: '',
    scriptId: null,
    variableGroupId: null,
    concurrency: 10,
    timeoutSec: 30,
    credentialId: null,
    execBackend: 'ssh'
  })
  const { hostOptions, fetchHostOptions } = useHostOptions()
  const credOptions = ref([])
  const scriptOptions = ref([])
  const varGroupOptions = ref([])
  const submitting = ref(false)

  const loadOptions = async () => {
    const [, c, s, v] = await Promise.all([
      fetchHostOptions(),
      getCredentialList({ page: 1, pageSize: 200 }),
      getScriptList({}),
      getVarGroupList({})
    ])
    if (c.code === 0) {
      credOptions.value = (c.data.list || c.data || []).filter(
        (x) => x.type === 'ssh_password' || x.type === 'ssh_key'
      )
    }
    if (s.code === 0) scriptOptions.value = s.data || []
    if (v.code === 0) varGroupOptions.value = v.data || []
  }

  // 选中脚本时把内容回填到命令框（可预览修改，实际执行以脚本内容为准）
  const onScriptPick = (id) => {
    const sc = scriptOptions.value.find((x) => x.ID === id)
    if (sc) form.command = sc.content
  }

  // 文件分发（Agent 通道，M9 A2）：仅可选 Agent 在线的主机，逐台并发调 /agent/file/push
  const fileForm = reactive({ localPath: '', remotePath: '', mode: '' })
  const fileHostIds = ref([])
  const filePushing = ref(false)
  const fileResults = ref([])
  const agentOnlineIDs = ref(new Set())
  const agentOnlineHosts = computed(() => hostOptions.value.filter((h) => agentOnlineIDs.value.has(h.ID)))

  const loadAgentOnline = async () => {
    const res = await getAgentInstanceList()
    if (res.code === 0) {
      agentOnlineIDs.value = new Set((res.data || []).filter((x) => x.status === '在线').map((x) => x.assetId))
    }
  }

  const onPushFile = async () => {
    if (!fileForm.localPath.trim() || !fileForm.remotePath.trim()) {
      ElMessage.warning('请填写 server 文件路径与目标路径')
      return
    }
    await ElMessageBox.confirm(
      `确定把 ${fileForm.localPath} 分发到 ${fileHostIds.value.length} 台主机的 ${fileForm.remotePath} 吗？同名文件将被覆盖。`,
      '文件分发确认',
      { confirmButtonText: '分发', cancelButtonText: '取消', type: 'warning' }
    )
    filePushing.value = true
    fileResults.value = []
    const hostMap = Object.fromEntries(hostOptions.value.map((h) => [h.ID, h]))
    try {
      await Promise.all(
        fileHostIds.value.map(async (id) => {
          const h = hostMap[id] || {}
          const name = `${h.hostname || id}（${h.ip || '-'}）`
          const started = Date.now()
          try {
            const res = await pushAgentFile({
              assetId: id,
              localPath: fileForm.localPath.trim(),
              remotePath: fileForm.remotePath.trim(),
              mode: fileForm.mode.trim() || undefined
            })
            fileResults.value.push({
              hostname: name,
              ok: res.code === 0,
              elapsedMs: res.code === 0 ? res.data.elapsedMs : Date.now() - started,
              detail: res.code === 0 ? '已写入（SHA256 校验通过）' : res.msg
            })
          } catch (e) {
            fileResults.value.push({ hostname: name, ok: false, elapsedMs: Date.now() - started, detail: String(e) })
          }
        })
      )
    } finally {
      filePushing.value = false
    }
  }

  const onSubmitExec = async () => {
    if (!form.hostIds.length || (!form.command.trim() && !form.scriptId)) {
      ElMessage.warning('请选择主机并输入命令（或从脚本库选择）')
      return
    }
    await ElMessageBox.confirm(
      `确定在 ${form.hostIds.length} 台主机上执行该命令吗？`,
      '批量执行确认',
      { confirmButtonText: '执行', cancelButtonText: '取消', type: 'warning' }
    )
    submitting.value = true
    try {
      const res = await createBatchExec({
        hostIds: form.hostIds,
        command: form.command,
        scriptId: form.scriptId || undefined,
        variableGroupId: form.variableGroupId || undefined,
        concurrency: form.concurrency,
        timeoutSec: form.timeoutSec,
        credentialId: form.credentialId || undefined,
        execBackend: form.execBackend
      })
      if (res.code === 0) {
        ElMessage.success(`批次 #${res.data.ID} 已创建，后台执行中`)
        getList()
      }
    } finally {
      submitting.value = false
    }
  }

  // 批次列表
  const tableData = ref([])
  const page = ref(1)
  const pageSize = ref(10)
  const total = ref(0)
  const listLoading = ref(false)
  const searchStatus = ref('')
  let pollTimer = null

  const getList = async () => {
    listLoading.value = true
    try {
      const res = await getBatchList(
        { page: page.value, pageSize: pageSize.value },
        { status: searchStatus.value || undefined }
      )
      if (res.code === 0) {
        tableData.value = res.data.list || []
        total.value = res.data.total || 0
      }
    } finally {
      listLoading.value = false
    }
  }

  const onSearch = () => {
    page.value = 1
    getList()
  }
  const handleCurrentChange = (v) => {
    page.value = v
    getList()
  }
  const handleSizeChange = (v) => {
    pageSize.value = v
    onSearch()
  }

  // 单定时器：当前页存在执行中批次时静默刷新
  const pollTick = async () => {
    if (!tableData.value.some((r) => r.status === '执行中')) {
      return
    }
    const res = await getBatchList(
      { page: page.value, pageSize: pageSize.value },
      { status: searchStatus.value || undefined }
    )
    if (res.code === 0) {
      tableData.value = res.data.list || []
      total.value = res.data.total || 0
    }
  }

  const onCancel = (row) => {
    ElMessageBox.confirm(`确定取消批次 #${row.ID} 吗？未完成的主机将被中断。`, '提示', {
      confirmButtonText: '取消批次',
      cancelButtonText: '返回',
      type: 'warning'
    }).then(async () => {
      const res = await cancelBatchExec({ id: row.ID })
      if (res.code === 0) {
        ElMessage.success(res.msg)
        getList()
      }
    })
  }

  // 详情抽屉
  const detailVisible = ref(false)
  const detailRecord = ref(null)
  const detailResults = ref([])
  const detailStatusFilter = ref('') // 每主机结果状态过滤（空=全部，服务端过滤口径）

  const loadDetailResults = async (id) => {
    const res = await getBatchDetail({
      id,
      resultStatus: detailStatusFilter.value || undefined
    })
    if (res.code === 0) {
      detailRecord.value = res.data.record
      detailResults.value = res.data.results || []
    }
  }

  const openDetail = async (row) => {
    detailStatusFilter.value = ''
    await loadDetailResults(row.ID)
    detailVisible.value = true
  }

  const reloadDetailResults = () => {
    if (detailRecord.value) loadDetailResults(detailRecord.value.ID)
  }

  // 复制单主机执行输出（带主机名/IP/状态上下文，便于贴工单）
  const copyResultOutput = (row) => {
    const parts = [`[${row.hostname} (${row.ip})] 结果: ${row.status}`]
    if (row.command) parts.push(`命令: ${row.command}`)
    if (row.error) parts.push(`ERROR: ${row.error}`)
    if (row.output) parts.push(row.output)
    navigator.clipboard?.writeText(parts.join('\n'))
    ElMessage.success(`已复制 ${row.hostname} 的执行输出`)
  }

  const statusTagType = (s) =>
    ({ 成功: 'success', 已完成: 'success', 失败: 'danger', 超时: 'warning', 已取消: 'info', 取消: 'info' }[s] || 'info')

  onMounted(() => {
    loadOptions()
    loadAgentOnline()
    getList()
    pollTimer = setInterval(pollTick, 3500)
  })
  onUnmounted(() => {
    if (pollTimer) clearInterval(pollTimer)
  })
</script>

<style scoped>
  .file-card-title {
    font-weight: 600;
    margin-bottom: 12px;
  }

  .ops-pre {
    margin: 0;
    max-height: 240px;
    overflow: auto;
    white-space: pre-wrap;
    word-break: break-all;
    font-size: 12px;
    background: var(--el-fill-color-light);
    padding: 8px;
    border-radius: 4px;
  }
</style>
