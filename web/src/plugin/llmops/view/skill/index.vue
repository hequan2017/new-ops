<template>
  <div>
    <!-- 技能 -->
    <div class="ops-card">
      <el-form inline>
        <el-form-item><h4 style="margin: 0">技能</h4></el-form-item>
        <el-form-item>
          <el-button type="primary" size="small" @click="openSkillEdit()">新建技能</el-button>
          <el-button type="success" size="small" @click="genVisible = true">零代码创建</el-button>
          <el-button size="small" @click="openExport">导出</el-button>
          <el-button size="small" @click="importVisible = true">导入</el-button>
          <el-button size="small" @click="loadSkills">刷 新</el-button>
        </el-form-item>
      </el-form>
      <el-table :data="skills" v-loading="skillLoading" stripe size="small" @selection-change="(v) => (skillSelection = v)">
        <el-table-column type="selection" width="42" />
        <el-table-column prop="name" label="技能名" width="140" show-overflow-tooltip />
        <el-table-column label="类型" width="80">
          <template #default="{ row }"><el-tag size="small">{{ row.type }}</el-tag></template>
        </el-table-column>
        <el-table-column prop="desc" label="描述" min-width="180" show-overflow-tooltip />
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag size="small" :type="row.status === 'enabled' ? 'success' : 'info'">{{ row.status === 'enabled' ? '启用' : '停用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="openSkillEdit(row)">编辑</el-button>
            <el-button link type="danger" size="small" @click="onDeleteSkill(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- MCP 工具 -->
    <div class="ops-card" style="margin-top: 12px">
      <el-form inline>
        <el-form-item><h4 style="margin: 0">MCP 工具（业务系统 API 封装）</h4></el-form-item>
        <el-form-item>
          <el-button type="primary" size="small" @click="mcpVisible = true">注册 API</el-button>
          <el-button size="small" @click="loadMcp">刷 新</el-button>
        </el-form-item>
      </el-form>
      <el-table :data="mcpTools" v-loading="mcpLoading" stripe size="small">
        <el-table-column prop="name" label="工具名" width="150" show-overflow-tooltip />
        <el-table-column prop="desc" label="描述" min-width="160" show-overflow-tooltip />
        <el-table-column prop="method" label="方法" width="70" />
        <el-table-column prop="endpoint" label="地址" min-width="200" show-overflow-tooltip />
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button link type="danger" size="small" @click="onDeleteMcp(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 连接器 -->
    <div class="ops-card" style="margin-top: 12px">
      <el-form inline>
        <el-form-item><h4 style="margin: 0">业务系统连接器（ERP/CRM/OA/SCM）</h4></el-form-item>
        <el-form-item>
          <el-button type="primary" size="small" @click="connVisible = true">新建连接器</el-button>
          <el-button size="small" @click="loadConnectors">刷 新</el-button>
        </el-form-item>
      </el-form>
      <el-table :data="connectors" v-loading="connLoading" stripe size="small">
        <el-table-column prop="name" label="名称" width="150" show-overflow-tooltip />
        <el-table-column label="类型" width="80">
          <template #default="{ row }"><el-tag size="small">{{ row.type }}</el-tag></template>
        </el-table-column>
        <el-table-column prop="baseUrl" label="地址" min-width="200" show-overflow-tooltip />
        <el-table-column prop="authType" label="鉴权" width="80" />
        <el-table-column label="操作" width="130" fixed="right">
          <template #default="{ row }">
            <el-button link type="success" size="small" @click="doTestConn(row)">测试</el-button>
            <el-button link type="danger" size="small" @click="onDeleteConn(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 令牌与审计 -->
    <div class="ops-card" style="margin-top: 12px">
      <el-form inline>
        <el-form-item><h4 style="margin: 0">调用令牌与审计</h4></el-form-item>
        <el-form-item>
          <el-button type="primary" size="small" @click="tokenVisible = true">创建令牌</el-button>
          <el-button size="small" @click="logsVisible = true; loadLogs()">审计日志</el-button>
          <el-button size="small" @click="loadTokens">刷 新</el-button>
        </el-form-item>
      </el-form>
      <el-table :data="tokens" v-loading="tokenLoading" stripe size="small">
        <el-table-column prop="name" label="令牌名" width="150" />
        <el-table-column prop="prefix" label="前缀" width="80" />
        <el-table-column prop="rateLimit" label="限流(/s)" width="90" />
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag size="small" :type="row.status === 'enabled' ? 'success' : 'danger'">{{ row.status === 'enabled' ? '启用' : '已吊销' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="90" fixed="right">
          <template #default="{ row }">
            <el-button link type="danger" size="small" :disabled="row.status !== 'enabled'" @click="doRevoke(row)">吊销</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 技能编辑 -->
    <el-dialog v-model="skillVisible" :title="skillForm.id ? '编辑技能' : '新建技能'" width="640px">
      <el-form label-width="100px">
        <el-form-item label="技能名"><el-input v-model="skillForm.name" /></el-form-item>
        <el-form-item label="类型">
          <el-select v-model="skillForm.type" style="width: 200px">
            <el-option label="扩展" value="extended" />
            <el-option label="自定义" value="custom" />
          </el-select>
        </el-form-item>
        <el-form-item label="描述"><el-input v-model="skillForm.desc" /></el-form-item>
        <el-form-item label="定义 JSON">
          <el-input v-model="skillForm.defJSON" type="textarea" :rows="6"
            placeholder='{"kind":"prompt","system":"系统指令","template":"处理 {input}，目标 {lang}","params":{"lang":{"desc":"目标语言","default":"英文"}}}' />
        </el-form-item>
        <el-form-item label="启用"><el-switch v-model="skillForm.enabled" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="skillVisible = false">取 消</el-button>
        <el-button type="primary" :loading="skillSaving" @click="doSaveSkill">保 存</el-button>
      </template>
    </el-dialog>

    <!-- 零代码 -->
    <el-dialog v-model="genVisible" title="零代码创建技能" width="560px">
      <el-input v-model="genDesc" type="textarea" :rows="3" placeholder="用自然语言描述技能需求，如：把输入的中文地址解析成省市区三级" />
      <pre v-if="genResult" class="ops-pre" style="margin-top: 10px">{{ genResult }}</pre>
      <template #footer>
        <el-button @click="genVisible = false">关 闭</el-button>
        <el-button type="primary" :loading="genLoading" @click="doGenerate">生 成</el-button>
      </template>
    </el-dialog>

    <!-- 导出 -->
    <el-dialog v-model="exportVisible" title="导出技能包" width="640px">
      <el-alert v-if="!skillSelection.length" title="请先在表格勾选技能" type="warning" :closable="false" />
      <pre v-else class="ops-pre">{{ exportPkg }}</pre>
      <template #footer>
        <el-button @click="exportVisible = false">关 闭</el-button>
        <el-button type="primary" :disabled="!skillSelection.length" @click="doExport">生 成 包</el-button>
      </template>
    </el-dialog>

    <!-- 导入 -->
    <el-dialog v-model="importVisible" title="导入技能包" width="640px">
      <el-input v-model="importRaw" type="textarea" :rows="8" placeholder="粘贴技能包 JSON" />
      <el-checkbox v-model="importOverwrite" style="margin-top: 8px">同名覆盖（默认跳过并提示冲突）</el-checkbox>
      <template #footer>
        <el-button @click="importVisible = false">取 消</el-button>
        <el-button type="primary" :loading="importLoading" @click="doImport">导 入</el-button>
      </template>
    </el-dialog>

    <!-- MCP 编辑 -->
    <el-dialog v-model="mcpVisible" title="注册业务 API 为 MCP 工具" width="620px">
      <el-form label-width="100px">
        <el-form-item label="工具名"><el-input v-model="mcpForm.name" /></el-form-item>
        <el-form-item label="描述"><el-input v-model="mcpForm.desc" /></el-form-item>
        <el-form-item label="方法">
          <el-select v-model="mcpForm.method" style="width: 120px">
            <el-option label="GET" value="GET" /><el-option label="POST" value="POST" />
          </el-select>
        </el-form-item>
        <el-form-item label="API 地址"><el-input v-model="mcpForm.endpoint" placeholder="http(s)://..." /></el-form-item>
        <el-form-item label="参数模板">
          <el-input v-model="mcpForm.argsSchema" type="textarea" :rows="3" placeholder='{"city":"城市名","date":"日期"}' />
        </el-form-item>
        <el-form-item label="固定头 JSON">
          <el-input v-model="mcpForm.headersJSON" type="textarea" :rows="2" placeholder='{"Authorization":"Bearer xxx"}（密文落库不回显）' />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="mcpVisible = false">取 消</el-button>
        <el-button type="primary" :loading="mcpSaving" @click="doSaveMcp">保 存</el-button>
      </template>
    </el-dialog>

    <!-- 连接器编辑 -->
    <el-dialog v-model="connVisible" title="新建连接器" width="600px">
      <el-form label-width="100px">
        <el-form-item label="名称"><el-input v-model="connForm.name" /></el-form-item>
        <el-form-item label="类型">
          <el-select v-model="connForm.type" style="width: 160px">
            <el-option v-for="t in ['erp','crm','oa','scm']" :key="t" :label="t.toUpperCase()" :value="t" />
          </el-select>
        </el-form-item>
        <el-form-item label="系统地址"><el-input v-model="connForm.baseUrl" placeholder="http(s)://..." /></el-form-item>
        <el-form-item label="鉴权方式">
          <el-select v-model="connForm.authType" style="width: 160px">
            <el-option v-for="a in ['none','bearer','basic','apikey']" :key="a" :label="a" :value="a" />
          </el-select>
        </el-form-item>
        <el-form-item label="鉴权值"><el-input v-model="connForm.authValue" type="password" show-password placeholder="none 可空" /></el-form-item>
        <el-form-item label="字段映射">
          <el-input v-model="connForm.fieldMap" type="textarea" :rows="3" placeholder='{"userName":"姓名"}' />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="connVisible = false">取 消</el-button>
        <el-button type="primary" :loading="connSaving" @click="doSaveConn">保 存</el-button>
      </template>
    </el-dialog>

    <!-- 令牌创建 -->
    <el-dialog v-model="tokenVisible" title="创建调用令牌" width="520px">
      <el-form label-width="100px">
        <el-form-item label="令牌名"><el-input v-model="tokenForm.name" /></el-form-item>
        <el-form-item label="限流(/s)"><el-input-number v-model="tokenForm.rateLimit" :min="1" :max="1000" /></el-form-item>
      </el-form>
      <el-alert v-if="tokenPlain" type="success" :closable="false" style="margin-top: 8px">
        <template #title>明文仅此一次：{{ tokenPlain }}</template>
      </el-alert>
      <template #footer>
        <el-button @click="tokenVisible = false">关 闭</el-button>
        <el-button type="primary" :loading="tokenSaving" @click="doCreateToken">创 建</el-button>
      </template>
    </el-dialog>

    <!-- 审计日志 -->
    <el-dialog v-model="logsVisible" title="技能调用审计日志" width="900px">
      <el-table :data="logs" v-loading="logsLoading" stripe size="small">
        <el-table-column prop="skill" label="技能/工具" width="150" show-overflow-tooltip />
        <el-table-column label="结果" width="70">
          <template #default="{ row }">
            <el-tag size="small" :type="row.ok ? 'success' : 'danger'">{{ row.ok ? '成功' : '失败' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="时延" width="80">
          <template #default="{ row }">{{ row.latencyMs }} ms</template>
        </el-table-column>
        <el-table-column prop="argsMasked" label="参数(脱敏)" min-width="200" show-overflow-tooltip />
        <el-table-column prop="errMsg" label="错误" min-width="160" show-overflow-tooltip />
      </el-table>
    </el-dialog>
  </div>
</template>

<script setup>
  import { ref, reactive, onMounted } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import {
    getSkills, saveSkill, deleteSkill, generateSkill, exportSkills, importSkills,
    getMcpTools, saveMcpTool, deleteMcpTool,
    getConnectors, saveConnector, testConnector, deleteConnector,
    getSkillTokens, createSkillToken, revokeSkillToken, getSkillCallLogs,
  } from '@/plugin/llmops/api/skill'

  // 技能
  const skills = ref([])
  const skillLoading = ref(false)
  const skillSelection = ref([])
  const loadSkills = async () => {
    skillLoading.value = true
    try {
      const res = await getSkills({})
      skills.value = res.data || []
    } finally {
      skillLoading.value = false
    }
  }
  const skillVisible = ref(false)
  const skillSaving = ref(false)
  const skillForm = reactive({ id: 0, name: '', type: 'custom', desc: '', defJSON: '', enabled: true })
  const openSkillEdit = (row) => {
    Object.assign(skillForm, {
      id: row?.ID || 0, name: row?.name || '', type: row?.type === 'builtin' ? 'custom' : row?.type || 'custom',
      desc: row?.desc || '', defJSON: row?.defJSON || '', enabled: row?.status !== 'disabled',
    })
    skillVisible.value = true
  }
  const doSaveSkill = async () => {
    skillSaving.value = true
    try {
      await saveSkill({ ...skillForm, status: skillForm.enabled ? 'enabled' : 'disabled' })
      ElMessage.success('保存成功')
      skillVisible.value = false
      loadSkills()
    } finally {
      skillSaving.value = false
    }
  }
  const onDeleteSkill = (row) => {
    ElMessageBox.confirm(`确认删除技能「${row.name}」？`, '提示', { type: 'warning' })
      .then(async () => {
        await deleteSkill({ id: row.ID })
        ElMessage.success('删除成功')
        loadSkills()
      })
      .catch(() => {})
  }

  // 零代码
  const genVisible = ref(false)
  const genLoading = ref(false)
  const genDesc = ref('')
  const genResult = ref('')
  const doGenerate = async () => {
    if (!genDesc.value) return
    genLoading.value = true
    try {
      const res = await generateSkill({ desc: genDesc.value })
      genResult.value = JSON.stringify(res.data, null, 2)
      ElMessage.success('生成成功并已入库')
      loadSkills()
    } finally {
      genLoading.value = false
    }
  }

  // 导出导入
  const exportVisible = ref(false)
  const exportPkg = ref('')
  const openExport = () => {
    exportPkg.value = ''
    exportVisible.value = true
  }
  const doExport = async () => {
    const ids = skillSelection.value.map((s) => s.ID)
    const res = await exportSkills({ ids })
    exportPkg.value = res.data || ''
  }
  const importVisible = ref(false)
  const importLoading = ref(false)
  const importRaw = ref('')
  const importOverwrite = ref(false)
  const doImport = async () => {
    importLoading.value = true
    try {
      const res = await importSkills({ raw: importRaw.value, overwrite: importOverwrite.value })
      ElMessage.success(`导入 ${res.data?.imported ?? 0} 个技能${res.data?.conflicts?.length ? '，冲突跳过：' + res.data.conflicts.join('、') : ''}`)
      importVisible.value = false
      importRaw.value = ''
      loadSkills()
    } finally {
      importLoading.value = false
    }
  }

  // MCP
  const mcpTools = ref([])
  const mcpLoading = ref(false)
  const loadMcp = async () => {
    mcpLoading.value = true
    try {
      const res = await getMcpTools()
      mcpTools.value = res.data || []
    } finally {
      mcpLoading.value = false
    }
  }
  const mcpVisible = ref(false)
  const mcpSaving = ref(false)
  const mcpForm = reactive({ name: '', desc: '', method: 'GET', endpoint: '', argsSchema: '', headersJSON: '' })
  const doSaveMcp = async () => {
    mcpSaving.value = true
    try {
      await saveMcpTool(mcpForm)
      ElMessage.success('保存成功')
      mcpVisible.value = false
      loadMcp()
    } finally {
      mcpSaving.value = false
    }
  }
  const onDeleteMcp = (row) => {
    ElMessageBox.confirm(`确认删除 MCP 工具「${row.name}」？`, '提示', { type: 'warning' })
      .then(async () => {
        await deleteMcpTool({ id: row.ID })
        ElMessage.success('删除成功')
        loadMcp()
      })
      .catch(() => {})
  }

  // 连接器
  const connectors = ref([])
  const connLoading = ref(false)
  const loadConnectors = async () => {
    connLoading.value = true
    try {
      const res = await getConnectors()
      connectors.value = res.data || []
    } finally {
      connLoading.value = false
    }
  }
  const connVisible = ref(false)
  const connSaving = ref(false)
  const connForm = reactive({ name: '', type: 'erp', baseUrl: '', authType: 'none', authValue: '', fieldMap: '' })
  const doSaveConn = async () => {
    connSaving.value = true
    try {
      await saveConnector(connForm)
      ElMessage.success('保存成功')
      connVisible.value = false
      loadConnectors()
    } finally {
      connSaving.value = false
    }
  }
  const doTestConn = async (row) => {
    await testConnector(row.ID)
    ElMessage.success('连通正常')
  }
  const onDeleteConn = (row) => {
    ElMessageBox.confirm(`确认删除连接器「${row.name}」？`, '提示', { type: 'warning' })
      .then(async () => {
        await deleteConnector({ id: row.ID })
        ElMessage.success('删除成功')
        loadConnectors()
      })
      .catch(() => {})
  }

  // 令牌与审计
  const tokens = ref([])
  const tokenLoading = ref(false)
  const loadTokens = async () => {
    tokenLoading.value = true
    try {
      const res = await getSkillTokens()
      tokens.value = res.data || []
    } finally {
      tokenLoading.value = false
    }
  }
  const tokenVisible = ref(false)
  const tokenSaving = ref(false)
  const tokenForm = reactive({ name: '', rateLimit: 10 })
  const tokenPlain = ref('')
  const doCreateToken = async () => {
    tokenSaving.value = true
    try {
      const res = await createSkillToken(tokenForm)
      tokenPlain.value = res.data?.plain || ''
      ElMessage.success('创建成功')
      tokenForm.name = ''
      loadTokens()
    } finally {
      tokenSaving.value = false
    }
  }
  const doRevoke = (row) => {
    ElMessageBox.confirm(`确认吊销令牌「${row.name}」？`, '提示', { type: 'warning' })
      .then(async () => {
        await revokeSkillToken(row.ID)
        ElMessage.success('已吊销')
        loadTokens()
      })
      .catch(() => {})
  }
  const logsVisible = ref(false)
  const logs = ref([])
  const logsLoading = ref(false)
  const loadLogs = async () => {
    logsLoading.value = true
    try {
      const res = await getSkillCallLogs({ page: 1, pageSize: 50 })
      logs.value = res.data?.list || []
    } finally {
      logsLoading.value = false
    }
  }

  onMounted(() => {
    loadSkills()
    loadMcp()
    loadConnectors()
    loadTokens()
  })
</script>
