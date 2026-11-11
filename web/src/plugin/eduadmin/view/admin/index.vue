<template>
  <div class="ops-card">
    <el-tabs v-model="tab">
      <!-- 数据接入 -->
      <el-tab-pane label="数据接入" name="ingest">
        <el-form inline>
          <el-form-item label="域">
            <el-select v-model="recordQuery.domain" clearable placeholder="全部" style="width: 100px" @change="loadRecords">
              <el-option v-for="d in domains" :key="d" :label="d" :value="d" />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-button size="small" @click="importVisible = true">批量导入</el-button>
            <el-button size="small" @click="loadRecords">刷 新</el-button>
          </el-form-item>
        </el-form>
        <el-table :data="records" v-loading="recordLoading" stripe size="small">
          <el-table-column label="域" width="80">
            <template #default="{ row }"><el-tag size="small">{{ row.domain }}</el-tag></template>
          </el-table-column>
          <el-table-column prop="externalKey" label="唯一键" width="140" show-overflow-tooltip />
          <el-table-column prop="source" label="来源" width="130" show-overflow-tooltip />
          <el-table-column prop="payloadJSON" label="负载" min-width="300" show-overflow-tooltip />
        </el-table>

        <el-divider content-position="left">清洗与脱敏</el-divider>
        <el-form inline>
          <el-form-item label="域">
            <el-select v-model="cleanForm.domain" style="width: 100px">
              <el-option v-for="d in domains" :key="d" :label="d" :value="d" />
            </el-select>
          </el-form-item>
          <el-form-item label="数值字段"><el-input v-model="cleanFields" placeholder="逗号分隔" style="width: 160px" /></el-form-item>
          <el-form-item label="策略">
            <el-select v-model="cleanForm.strategy" style="width: 90px">
              <el-option label="均值" value="均值" /><el-option label="前值" value="前值" />
            </el-select>
          </el-form-item>
          <el-form-item><el-checkbox v-model="cleanForm.dryRun">试运行</el-checkbox></el-form-item>
          <el-form-item><el-button type="primary" size="small" :loading="cleanLoading" @click="doClean">清 洗</el-button></el-form-item>
          <el-form-item label="脱敏原文"><el-input v-model="maskPayloadText" placeholder='{"idCard":"…"}' style="width: 220px" /></el-form-item>
          <el-form-item><el-button size="small" :loading="maskLoading" @click="doMask">脱敏预览</el-button></el-form-item>
        </el-form>
        <el-descriptions v-if="cleanResult" :column="3" border size="small">
          <el-descriptions-item label="扫描">{{ cleanResult.scanned }}</el-descriptions-item>
          <el-descriptions-item label="乱码/格式">{{ (cleanResult.mojibakeIds || []).length }}/{{ (cleanResult.formatErrIds || []).length }}</el-descriptions-item>
          <el-descriptions-item label="填充/异常">{{ cleanResult.filled }}/{{ (cleanResult.outlierIds || []).length }}</el-descriptions-item>
        </el-descriptions>
        <el-descriptions v-if="maskResult" :column="3" border size="small" style="margin-top: 8px">
          <el-descriptions-item label="分级"><el-tag size="small" :type="maskResult.level === '机密' ? 'danger' : maskResult.level === '敏感' ? 'warning' : 'info'">{{ maskResult.level }}</el-tag></el-descriptions-item>
          <el-descriptions-item label="哈希字段">{{ JSON.stringify(maskResult.hashedFields) }}</el-descriptions-item>
          <el-descriptions-item label="金额字段">{{ JSON.stringify(maskResult.amountFields) }}</el-descriptions-item>
        </el-descriptions>
      </el-tab-pane>

      <!-- 知识库 -->
      <el-tab-pane label="知识库" name="kb">
        <el-form inline>
          <el-form-item label="检索词">
            <el-input v-model="kbQuery" placeholder="六大知识库/审计语料" style="width: 200px" />
            <el-button type="primary" size="small" style="margin-left: 8px" :loading="kbLoading" @click="doKbSearch">检 索</el-button>
          </el-form-item>
        </el-form>
        <el-divider content-position="left">管理知识库（人事/财务/资产/后勤/决策/跨校）</el-divider>
        <el-table :data="policyHits" v-loading="kbLoading" stripe size="small">
          <el-table-column label="域" width="80">
            <template #default="{ row }"><el-tag size="small">{{ row.domain }}</el-tag></template>
          </el-table-column>
          <el-table-column prop="title" label="标题" min-width="180" show-overflow-tooltip />
          <el-table-column prop="content" label="内容" min-width="280" show-overflow-tooltip />
          <el-table-column prop="score" label="命中分" width="80" />
        </el-table>
        <el-divider content-position="left">审计知识语料（财务/采购/人事/资产/工程/数据安全）</el-divider>
        <el-table :data="auditHits" stripe size="small">
          <el-table-column label="域" width="90">
            <template #default="{ row }"><el-tag size="small" type="warning">{{ row.domain }}</el-tag></template>
          </el-table-column>
          <el-table-column prop="title" label="标题" min-width="170" show-overflow-tooltip />
          <el-table-column prop="riskPoints" label="风险要点" min-width="240" show-overflow-tooltip />
          <el-table-column prop="controlAdvice" label="内控建议" min-width="200" show-overflow-tooltip />
          <el-table-column prop="score" label="命中分" width="80" />
        </el-table>
      </el-tab-pane>

      <!-- 工具台 -->
      <el-tab-pane label="工具台" name="tools">
        <el-form inline>
          <el-form-item label="招聘岗位"><el-input-number v-model="screenPositionId" :min="1" /></el-form-item>
          <el-form-item><el-button size="small" :loading="screenLoading" @click="doScreen">候选人筛选</el-button></el-form-item>
          <el-form-item label="总预算(万)"><el-input-number v-model="budgetTotal" :min="1" /></el-form-item>
          <el-form-item><el-button size="small" :loading="budgetLoading" @click="doBudget">分配建议</el-button></el-form-item>
          <el-form-item><el-button type="danger" size="small" :loading="riskLoading" @click="doRiskScan">风险扫描</el-button></el-form-item>
        </el-form>
        <el-row :gutter="10">
          <el-col :span="8">
            <el-card shadow="never" header="候选人筛选（高/中/低）" size="small">
              <el-table :data="screenResult" size="small" max-height="240">
                <el-table-column prop="name" label="姓名" width="80" />
                <el-table-column prop="score" label="融合分" width="80" />
                <el-table-column label="档" width="60">
                  <template #default="{ row }"><el-tag size="small" :type="row.level === '高' ? 'success' : row.level === '中' ? 'warning' : 'info'">{{ row.level }}</el-tag></template>
                </el-table-column>
              </el-table>
            </el-card>
          </el-col>
          <el-col :span="8">
            <el-card shadow="never" header="预算分配建议" size="small">
              <el-table :data="budgetResult" size="small" max-height="240">
                <el-table-column prop="dept" label="部门" min-width="90" />
                <el-table-column prop="amount" label="金额(万)" width="90" />
                <el-table-column prop="ratio" label="占比" width="70" />
              </el-table>
            </el-card>
          </el-col>
          <el-col :span="8">
            <el-card shadow="never" header="财务风险预警" size="small">
              <el-table :data="riskResult" size="small" max-height="240">
                <el-table-column label="类型" width="90">
                  <template #default="{ row }">{{ row.type }}</template>
                </el-table-column>
                <el-table-column label="等级" width="60">
                  <template #default="{ row }"><el-tag size="small" :type="row.level === '高' ? 'danger' : 'warning'">{{ row.level }}</el-tag></template>
                </el-table-column>
                <el-table-column prop="detail" label="详情" min-width="130" show-overflow-tooltip />
              </el-table>
            </el-card>
          </el-col>
        </el-row>
        <el-divider content-position="left">报销单据（提交-审核中-支付/驳回）</el-divider>
        <el-table :data="invoices" v-loading="invoiceLoading" stripe size="small">
          <el-table-column prop="invoiceNo" label="发票号" width="130" />
          <el-table-column prop="applicant" label="申请人" width="90" />
          <el-table-column prop="category" label="类别" width="80" />
          <el-table-column prop="amount" label="金额" width="90" />
          <el-table-column label="状态" width="90">
            <template #default="{ row }">
              <el-tag size="small" :type="row.status === '支付' ? 'success' : row.status === '驳回' ? 'danger' : 'warning'">{{ row.status }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="rejectReason" label="驳回原因" min-width="150" show-overflow-tooltip />
        </el-table>
      </el-tab-pane>
    </el-tabs>

    <!-- 导入对话框 -->
    <el-dialog v-model="importVisible" title="批量导入数据记录" width="720px">
      <el-form label-width="80px">
        <el-form-item label="格式">
          <el-radio-group v-model="importForm.format">
            <el-radio value="json">JSON</el-radio>
            <el-radio value="csv">CSV</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="内容">
          <el-input v-model="importForm.content" type="textarea" :rows="8" placeholder='JSON: [{"domain":"人事","externalKey":"HR-1","payload":{...}}]；CSV 列序 domain,externalKey,source,payloadJSON' />
        </el-form-item>
      </el-form>
      <el-descriptions v-if="importResult" :column="2" border size="small">
        <el-descriptions-item label="入库">{{ importResult.inserted }}</el-descriptions-item>
        <el-descriptions-item label="去重">{{ importResult.deduped }}</el-descriptions-item>
        <el-descriptions-item label="失败行" :span="2">{{ JSON.stringify(importResult.failedRows) }}</el-descriptions-item>
      </el-descriptions>
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
    getEduRecords, importEduRecords, cleanEduRecords, maskPayload,
    searchEduPolicies, searchEduAudit,
    screenCandidates, suggestBudget, getEduInvoices, scanEduRisk,
  } from '@/plugin/eduadmin/api/eduadmin'

  const tab = ref('ingest')
  const domains = ['人事', '财务', '资产', '后勤']

  // 数据接入
  const records = ref([])
  const recordLoading = ref(false)
  const recordQuery = reactive({ domain: '' })
  const importVisible = ref(false)
  const importForm = reactive({ format: 'json', content: '' })
  const importLoading = ref(false)
  const importResult = ref(null)
  const cleanForm = reactive({ domain: '人事', strategy: '均值', dryRun: true })
  const cleanFields = ref('amount,salary')
  const cleanLoading = ref(false)
  const cleanResult = ref(null)
  const maskPayloadText = ref('')
  const maskLoading = ref(false)
  const maskResult = ref(null)

  const loadRecords = async () => {
    recordLoading.value = true
    try {
      const res = await getEduRecords({ ...recordQuery, pageSize: 30 })
      records.value = (res.data && res.data.list) || []
    } finally {
      recordLoading.value = false
    }
  }
  const doImport = async () => {
    importLoading.value = true
    try {
      const res = await importEduRecords(importForm)
      importResult.value = res.data
      ElMessage.success(`入库 ${res.data.inserted} 条`)
      loadRecords()
    } finally {
      importLoading.value = false
    }
  }
  const doClean = async () => {
    cleanLoading.value = true
    try {
      const res = await cleanEduRecords({ ...cleanForm, fields: cleanFields.value.split(',').map(s => s.trim()).filter(Boolean) })
      cleanResult.value = res.data
    } finally {
      cleanLoading.value = false
    }
  }
  const doMask = async () => {
    maskLoading.value = true
    try {
      const res = await maskPayload({ payload: maskPayloadText.value })
      maskResult.value = res.data
    } finally {
      maskLoading.value = false
    }
  }

  // 知识库
  const kbQuery = ref('')
  const kbLoading = ref(false)
  const policyHits = ref([])
  const auditHits = ref([])
  const doKbSearch = async () => {
    kbLoading.value = true
    try {
      const [p, a] = await Promise.all([
        searchEduPolicies({ query: kbQuery.value, topK: 5 }),
        searchEduAudit({ query: kbQuery.value, topK: 5 }),
      ])
      policyHits.value = p.data || []
      auditHits.value = a.data || []
    } finally {
      kbLoading.value = false
    }
  }

  // 工具台
  const screenPositionId = ref(1)
  const screenLoading = ref(false)
  const screenResult = ref([])
  const budgetTotal = ref(1000)
  const budgetLoading = ref(false)
  const budgetResult = ref([])
  const riskLoading = ref(false)
  const riskResult = ref([])
  const invoices = ref([])
  const invoiceLoading = ref(false)

  const doScreen = async () => {
    screenLoading.value = true
    try {
      const res = await screenCandidates({ positionId: screenPositionId.value })
      screenResult.value = (res.data || []).slice(0, 10)
    } finally {
      screenLoading.value = false
    }
  }
  const doBudget = async () => {
    budgetLoading.value = true
    try {
      const res = await suggestBudget({ total: budgetTotal.value })
      budgetResult.value = res.data || []
    } finally {
      budgetLoading.value = false
    }
  }
  const doRiskScan = async () => {
    riskLoading.value = true
    try {
      await scanEduRisk({ cashBalance: 100, cashFloor: 100 })
      const res = await getRiskAlerts({ pageSize: 20 })
      riskResult.value = (res.data && res.data.list) || []
      ElMessage.success('扫描完成')
    } finally {
      riskLoading.value = false
    }
  }
  const loadInvoices = async () => {
    invoiceLoading.value = true
    try {
      const res = await getEduInvoices({ pageSize: 20 })
      invoices.value = (res.data && res.data.list) || []
    } finally {
      invoiceLoading.value = false
    }
  }

  onMounted(() => {
    loadRecords()
    loadInvoices()
  })
</script>
