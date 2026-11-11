<template>
  <div class="ops-card">
    <el-tabs v-model="tab">
      <!-- 制度库 -->
      <el-tab-pane label="制度库" name="docs">
        <el-form inline>
          <el-form-item label="密级">
            <el-select v-model="docQuery.securityLevel" clearable placeholder="全部" style="width: 90px" @change="loadDocs">
              <el-option v-for="l in levels" :key="l" :label="l" :value="l" />
            </el-select>
          </el-form-item>
          <el-form-item><el-button size="small" @click="importVisible = true">导入制度</el-button></el-form-item>
          <el-form-item><el-button size="small" @click="loadDocs">刷 新</el-button></el-form-item>
        </el-form>
        <el-table :data="docs" v-loading="docLoading" stripe size="small">
          <el-table-column prop="name" label="制度名" min-width="200" show-overflow-tooltip />
          <el-table-column label="格式" width="80">
            <template #default="{ row }"><el-tag size="small">{{ row.format }}</el-tag></template>
          </el-table-column>
          <el-table-column label="密级" width="80">
            <template #default="{ row }">
              <el-tag size="small" :type="row.securityLevel === '涉密' ? 'danger' : row.securityLevel === '内部' ? 'warning' : 'info'">{{ row.securityLevel }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="SM4 口径" width="90">
            <template #default="{ row }">{{ row.sm4Mark ? '是' : '否' }}</template>
          </el-table-column>
          <el-table-column prop="permScope" label="权限范围" min-width="130" show-overflow-tooltip />
          <el-table-column label="操作" width="150">
            <template #default="{ row }">
              <el-button size="small" text type="primary" @click="doClean(row)">清 洗</el-button>
              <el-button size="small" text type="warning" @click="goLabel(row)">标 注</el-button>
            </template>
          </el-table-column>
        </el-table>
        <el-descriptions v-if="cleanReport" :column="3" border size="small" style="margin-top: 10px">
          <el-descriptions-item label="条数">{{ cleanReport.clauses }}</el-descriptions-item>
          <el-descriptions-item label="问题">{{ (cleanReport.issues || []).length }}</el-descriptions-item>
          <el-descriptions-item label="结论">{{ cleanReport.ok ? '通过' : '需复核' }}</el-descriptions-item>
          <el-descriptions-item label="明细" :span="3">{{ JSON.stringify(cleanReport.issues) }}</el-descriptions-item>
        </el-descriptions>
      </el-tab-pane>

      <!-- 标注 -->
      <el-tab-pane label="标注" name="label">
        <el-form inline>
          <el-form-item label="制度ID"><el-input-number v-model="labelDocId" :min="1" /></el-form-item>
          <el-form-item><el-button size="small" :loading="extractLoading" @click="doExtract">实体抽取</el-button></el-form-item>
          <el-form-item><el-button type="primary" size="small" :loading="preLabelLoading" @click="doPreLabel">规则预标</el-button></el-form-item>
        </el-form>
        <el-divider content-position="left">实体</el-divider>
        <el-table :data="entities" v-loading="entityLoading" stripe size="small" style="margin-bottom: 10px">
          <el-table-column prop="name" label="实体名" min-width="200" show-overflow-tooltip />
          <el-table-column label="类型" width="100">
            <template #default="{ row }"><el-tag size="small">{{ row.eType }}</el-tag></template>
          </el-table-column>
          <el-table-column prop="source" label="来源" width="90" />
        </el-table>
        <el-divider content-position="left">分词标注（点击改标）</el-divider>
        <div style="line-height: 2.2">
          <el-tag
            v-for="tl in tokenLabels" :key="tl.id"
            size="small" :type="tl.label === 'O' ? 'info' : tl.label === '机构' ? 'success' : 'warning'"
            style="margin: 2px; cursor: pointer" @click="cycleLabel(tl)"
          >{{ tl.token }}<span style="opacity: .6">{{ tl.label === 'O' ? '' : '·' + tl.label }}</span></el-tag>
        </div>
      </el-tab-pane>

      <!-- 鉴权 -->
      <el-tab-pane label="鉴权" name="acl">
        <el-form inline>
          <el-form-item label="角色">
            <el-select v-model="aclQuery.roleName" style="width: 160px" allow-create filterable>
              <el-option v-for="r in roles" :key="r" :label="r" :value="r" />
            </el-select>
          </el-form-item>
          <el-form-item label="关键字"><el-input v-model="aclQuery.keyword" placeholder="制度名" style="width: 160px" clearable /></el-form-item>
          <el-form-item><el-button type="primary" size="small" :loading="aclLoading" @click="doRoleSearch">按角色检索</el-button></el-form-item>
        </el-form>
        <el-descriptions :column="1" border size="small" style="margin-bottom: 10px">
          <el-descriptions-item label="说明">角色×分类 grant 控制：无授权/超密级/停用均不可见</el-descriptions-item>
        </el-descriptions>
        <el-table :data="aclResults" v-loading="aclResultLoading" stripe size="small">
          <el-table-column prop="name" label="可见制度" min-width="200" show-overflow-tooltip />
          <el-table-column label="密级" width="80">
            <template #default="{ row }">
              <el-tag size="small" :type="row.securityLevel === '涉密' ? 'danger' : row.securityLevel === '内部' ? 'warning' : 'info'">{{ row.securityLevel }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="format" label="格式" width="80" />
        </el-table>
      </el-tab-pane>

      <!-- 问答台 -->
      <el-tab-pane label="问答台" name="qa">
        <el-form label-width="80px" style="max-width: 680px">
          <el-form-item label="提问角色">
            <el-select v-model="qaForm.roleName" style="width: 180px" allow-create filterable>
              <el-option v-for="r in roles" :key="r" :label="r" :value="r" />
            </el-select>
          </el-form-item>
          <el-form-item label="问题">
            <el-input v-model="qaForm.question" type="textarea" :rows="3" placeholder="如：横向课题经费预算如何调整？" />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" :loading="qaLoading" @click="doQa">提 问</el-button>
          </el-form-item>
        </el-form>
        <el-descriptions v-if="qaLog" :column="3" border size="small" style="margin-top: 10px">
          <el-descriptions-item label="来源"><el-tag size="small">{{ qaLog.source }}</el-tag></el-descriptions-item>
          <el-descriptions-item label="时延">{{ qaLog.latencyMs }}ms</el-descriptions-item>
          <el-descriptions-item label="角色">{{ qaLog.roleName }}</el-descriptions-item>
          <el-descriptions-item label="回答" :span="3"><div class="ops-pre">{{ qaLog.answer }}</div></el-descriptions-item>
        </el-descriptions>
        <el-divider content-position="left">提示词模板</el-divider>
        <el-table :data="prompts" stripe size="small">
          <el-table-column prop="name" label="模板名" min-width="160" />
          <el-table-column prop="vars" label="占位变量" min-width="200" show-overflow-tooltip />
        </el-table>
      </el-tab-pane>
    </el-tabs>

    <!-- 导入对话框 -->
    <el-dialog v-model="importVisible" title="导入制度文档" width="720px">
      <el-form label-width="90px">
        <el-form-item label="制度名"><el-input v-model="importForm.name" /></el-form-item>
        <el-form-item label="格式">
          <el-select v-model="importForm.format" style="width: 130px">
            <el-option v-for="f in ['TXT','MD','Word','PDF','HTML']" :key="f" :label="f" :value="f" />
          </el-select>
        </el-form-item>
        <el-form-item label="密级">
          <el-select v-model="importForm.securityLevel" style="width: 130px">
            <el-option v-for="l in levels" :key="l" :label="l" :value="l" />
          </el-select>
        </el-form-item>
        <el-form-item label="权限范围"><el-input v-model="importForm.permScope" placeholder="角色/单位，逗号分隔" /></el-form-item>
        <el-form-item label="原文">
          <el-input v-model="importForm.rawText" type="textarea" :rows="8" placeholder="TXT/MD 粘贴原文；Word/PDF/HTML 仅编目可留空" />
        </el-form-item>
      </el-form>
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
    getPolicyDocs, importPolicyDoc, cleanPolicyDoc,
    extractEntities, getEntities, preLabel, getTokenLabels, updateTokenLabel,
    searchByRole, getPolicyPrompts, policyQa,
  } from '@/plugin/policyreg/api/policyreg'

  const tab = ref('docs')
  const levels = ['公开', '内部', '涉密']
  const roles = ['教师', '科研人员', '行政人员', '审计员', '学生', '访客']

  // 制度库
  const docs = ref([])
  const docLoading = ref(false)
  const docQuery = reactive({ securityLevel: '' })
  const importVisible = ref(false)
  const importForm = reactive({ name: '', format: 'TXT', securityLevel: '公开', permScope: '', rawText: '' })
  const importLoading = ref(false)
  const cleanReport = ref(null)

  const loadDocs = async () => {
    docLoading.value = true
    try {
      const res = await getPolicyDocs({ ...docQuery, pageSize: 30 })
      docs.value = (res.data && res.data.list) || []
    } finally {
      docLoading.value = false
    }
  }
  const doImport = async () => {
    importLoading.value = true
    try {
      const res = await importPolicyDoc(importForm)
      ElMessage.success(`已导入 ${res.data.name}`)
      importVisible.value = false
      loadDocs()
    } finally {
      importLoading.value = false
    }
  }
  const doClean = async (row) => {
    const res = await cleanPolicyDoc({ id: row.id })
    cleanReport.value = res.data
  }

  // 标注
  const labelDocId = ref(1)
  const entities = ref([])
  const entityLoading = ref(false)
  const extractLoading = ref(false)
  const tokenLabels = ref([])
  const preLabelLoading = ref(false)

  const doExtract = async () => {
    extractLoading.value = true
    try {
      const res = await extractEntities({ docId: labelDocId.value })
      entities.value = res.data || []
      ElMessage.success(`抽取 ${entities.value.length} 个实体`)
    } finally {
      extractLoading.value = false
    }
  }
  const doPreLabel = async () => {
    preLabelLoading.value = true
    try {
      await preLabel({ docId: labelDocId.value })
      const res = await getTokenLabels({ docId: labelDocId.value })
      tokenLabels.value = res.data || []
      ElMessage.success(`预标 ${tokenLabels.value.length} 词`)
    } finally {
      preLabelLoading.value = false
    }
  }
  const cycleLabels = ['O', '机构', '文件号', '日期', '专有名词']
  const cycleLabel = async (tl) => {
    const idx = cycleLabels.indexOf(tl.label)
    const next = cycleLabels[(idx + 1) % cycleLabels.length]
    await updateTokenLabel({ id: tl.id, label: next })
    tl.label = next
    tl.source = '人工'
  }
  const goLabel = (row) => {
    labelDocId.value = row.id
    tab.value = 'label'
  }

  // 鉴权
  const aclQuery = reactive({ roleName: '教师', keyword: '' })
  const aclLoading = ref(false)
  const aclResults = ref([])
  const aclResultLoading = ref(false)
  const doRoleSearch = async () => {
    aclResultLoading.value = true
    try {
      const res = await searchByRole({ ...aclQuery, pageSize: 30 })
      aclResults.value = (res.data && res.data.list) || []
    } finally {
      aclResultLoading.value = false
    }
  }

  // 问答台
  const qaForm = reactive({ roleName: '教师', question: '' })
  const qaLoading = ref(false)
  const qaLog = ref(null)
  const prompts = ref([])
  const doQa = async () => {
    qaLoading.value = true
    try {
      const res = await policyQa({ ...qaForm })
      qaLog.value = res.data
    } finally {
      qaLoading.value = false
    }
  }
  const loadPrompts = async () => {
    const res = await getPolicyPrompts()
    prompts.value = res.data || []
  }

  onMounted(() => {
    loadDocs()
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
