<template>
  <div>
    <div class="ops-card">
      <el-form inline>
        <el-form-item label="类型">
          <el-select v-model="query.dataType" clearable placeholder="全部" style="width: 150px">
            <el-option v-for="t in dataTypes" :key="t" :label="typeLabel(t)" :value="t" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="query.status" clearable placeholder="全部" style="width: 120px">
            <el-option v-for="s in statuses" :key="s.value" :label="s.label" :value="s.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="关键字">
          <el-input v-model="query.keyword" placeholder="名称/标签" clearable style="width: 160px" @keyup.enter="getList" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="getList">查 询</el-button>
          <el-button @click="resetQuery">重 置</el-button>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="uploadVisible = true">上传文档</el-button>
          <el-button @click="webVisible = true">网页采集</el-button>
          <el-button @click="qaVisible = true">导入问答对</el-button>
        </el-form-item>
      </el-form>

      <el-table :data="tableData" v-loading="loading" stripe>
        <el-table-column prop="name" label="名称" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">
            <el-link type="primary" @click="openDetail(row)">{{ row.name }}</el-link>
          </template>
        </el-table-column>
        <el-table-column label="类型" width="110">
          <template #default="{ row }">
            <el-tag size="small" :type="typeTag(row.dataType)">{{ typeLabel(row.dataType) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag size="small" :type="statusTag(row.status)">{{ statusLabel(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="source" label="来源" width="100" />
        <el-table-column label="大小" width="110">
          <template #default="{ row }">{{ fmtSize(row.size) }}</template>
        </el-table-column>
        <el-table-column prop="tags" label="标签" min-width="140" show-overflow-tooltip />
        <el-table-column label="创建时间" width="170">
          <template #default="{ row }">{{ fmtTime(row.CreatedAt) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="140" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openDetail(row)">详情</el-button>
            <el-button link type="danger" @click="onDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-pagination
        style="margin-top: 12px; justify-content: flex-end"
        v-model:current-page="query.page"
        v-model:page-size="query.pageSize"
        :total="total"
        :page-sizes="[10, 20, 50, 100]"
        layout="total, sizes, prev, pager, next"
        @current-change="getList"
        @size-change="getList"
      />
    </div>

    <!-- 详情抽屉 -->
    <el-drawer v-model="detailVisible" :title="detail?.name" size="55%">
      <template v-if="detail">
        <el-descriptions :column="2" border size="small">
          <el-descriptions-item label="类型">{{ typeLabel(detail.dataType) }}</el-descriptions-item>
          <el-descriptions-item label="状态">{{ statusLabel(detail.status) }}</el-descriptions-item>
          <el-descriptions-item label="来源">{{ detail.source }}</el-descriptions-item>
          <el-descriptions-item label="大小">{{ fmtSize(detail.size) }}</el-descriptions-item>
          <el-descriptions-item label="标签" :span="2">{{ detail.tags || '—' }}</el-descriptions-item>
        </el-descriptions>
        <h4 style="margin: 14px 0 8px">编目元数据</h4>
        <pre class="ops-pre">{{ prettyJSON(detail.metaJSON) }}</pre>
        <template v-if="detail.rawText">
          <h4 style="margin: 14px 0 8px">正文预览（前 3000 字）</h4>
          <pre class="ops-pre">{{ detail.rawText.slice(0, 3000) }}{{ detail.rawText.length > 3000 ? '…' : '' }}</pre>
        </template>
        <h4 style="margin: 14px 0 8px">生命周期事件</h4>
        <el-timeline>
          <el-timeline-item v-for="ev in events" :key="ev.ID" :timestamp="fmtTime(ev.CreatedAt)">
            <el-tag size="small">{{ ev.action }}</el-tag> {{ ev.detail }}
          </el-timeline-item>
        </el-timeline>
      </template>
    </el-drawer>

    <!-- 上传文档 -->
    <el-dialog v-model="uploadVisible" title="上传文档" width="480px">
      <el-form label-width="80px">
        <el-form-item label="文件">
          <input ref="docFileRef" type="file" accept=".txt,.md,.csv,.json,.yaml,.yml,.xml,.log,.html,.sql" />
        </el-form-item>
        <el-form-item label="资产名">
          <el-input v-model="uploadForm.name" placeholder="默认取文件名" />
        </el-form-item>
        <el-form-item label="标签">
          <el-input v-model="uploadForm.tags" placeholder="逗号分隔" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="uploadVisible = false">取 消</el-button>
        <el-button type="primary" :loading="uploadLoading" @click="doUpload">上 传</el-button>
      </template>
    </el-dialog>

    <!-- 网页采集 -->
    <el-dialog v-model="webVisible" title="网页采集" width="520px">
      <el-form label-width="80px">
        <el-form-item label="URL">
          <el-input v-model="webForm.url" placeholder="http(s):// 开头" />
        </el-form-item>
        <el-form-item label="资产名">
          <el-input v-model="webForm.name" placeholder="默认取页面标题" />
        </el-form-item>
        <el-form-item label="标签">
          <el-input v-model="webForm.tags" placeholder="逗号分隔" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="webVisible = false">取 消</el-button>
        <el-button type="primary" :loading="webLoading" @click="doImportWeb">采 集</el-button>
      </template>
    </el-dialog>

    <!-- 问答对导入 -->
    <el-dialog v-model="qaVisible" title="导入问答对" width="480px">
      <el-form label-width="80px">
        <el-form-item label="文件">
          <input ref="qaFileRef" type="file" accept=".json,.csv" />
        </el-form-item>
        <el-form-item label="资产名">
          <el-input v-model="qaForm.name" placeholder="默认「未命名文档」" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="qaVisible = false">取 消</el-button>
        <el-button type="primary" :loading="qaLoading" @click="doImportQa">导 入</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
  import { ref, reactive, onMounted } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import {
    getAssetList,
    getAsset,
    deleteAsset,
    importDoc,
    importQa,
    importWeb,
    getAssetEvents,
  } from '@/plugin/llmops/api/data'

  const dataTypes = ['doc', 'image', 'audio', 'video', 'webpage', 'qa', 'structured', 'semistructured']
  const typeNames = {
    doc: '文档', image: '图片', audio: '音频', video: '视频',
    webpage: '网页', qa: '问答对', structured: '结构化', semistructured: '半结构化',
  }
  const statuses = [
    { value: 'draft', label: '已采集' },
    { value: 'parsed', label: '已解析' },
    { value: 'cataloged', label: '已编目' },
    { value: 'offline', label: '已下线' },
  ]
  const typeLabel = (t) => typeNames[t] || t
  const typeTag = (t) => ({ doc: 'primary', qa: 'success', webpage: 'warning', structured: 'info', semistructured: 'info' }[t] || 'info')
  const statusLabel = (s) => (statuses.find((x) => x.value === s) || {}).label || s
  const statusTag = (s) => ({ draft: 'info', parsed: 'success', cataloged: 'primary', offline: 'danger' }[s] || 'info')
  const fmtSize = (n) => {
    if (!n) return '—'
    if (n < 1024) return n + ' B'
    if (n < 1048576) return (n / 1024).toFixed(1) + ' KB'
    if (n < 1073741824) return (n / 1048576).toFixed(1) + ' MB'
    return (n / 1073741824).toFixed(2) + ' GB'
  }
  const fmtTime = (t) => (t ? String(t).replace('T', ' ').slice(0, 19) : '—')
  const prettyJSON = (s) => {
    if (!s) return '—'
    try {
      return JSON.stringify(JSON.parse(s), null, 2)
    } catch (e) {
      return s
    }
  }

  const query = reactive({ page: 1, pageSize: 10, dataType: '', status: '', keyword: '' })
  const tableData = ref([])
  const total = ref(0)
  const loading = ref(false)

  const getList = async () => {
    loading.value = true
    try {
      const res = await getAssetList(query)
      tableData.value = res.data?.list || []
      total.value = res.data?.total || 0
    } finally {
      loading.value = false
    }
  }
  const resetQuery = () => {
    Object.assign(query, { page: 1, dataType: '', status: '', keyword: '' })
    getList()
  }

  const detailVisible = ref(false)
  const detail = ref(null)
  const events = ref([])
  const openDetail = async (row) => {
    const res = await getAsset({ id: row.ID })
    detail.value = res.data
    detailVisible.value = true
    const evRes = await getAssetEvents({ assetId: row.ID, page: 1, pageSize: 20 })
    events.value = evRes.data?.list || []
  }

  const onDelete = (row) => {
    ElMessageBox.confirm(`确认删除资产「${row.name}」？（软删除，可恢复）`, '提示', { type: 'warning' })
      .then(async () => {
        await deleteAsset({ id: row.ID })
        ElMessage.success('删除成功')
        getList()
      })
      .catch(() => {})
  }

  // 上传文档
  const uploadVisible = ref(false)
  const uploadLoading = ref(false)
  const docFileRef = ref(null)
  const uploadForm = reactive({ name: '', tags: '' })
  const doUpload = async () => {
    const file = docFileRef.value?.files?.[0]
    if (!file) {
      ElMessage.warning('请选择文件')
      return
    }
    uploadLoading.value = true
    try {
      await importDoc(file, uploadForm.name, uploadForm.tags)
      ElMessage.success('采集成功')
      uploadVisible.value = false
      uploadForm.name = ''
      uploadForm.tags = ''
      if (docFileRef.value) docFileRef.value.value = ''
      getList()
    } finally {
      uploadLoading.value = false
    }
  }

  // 网页采集
  const webVisible = ref(false)
  const webLoading = ref(false)
  const webForm = reactive({ url: '', name: '', tags: '' })
  const doImportWeb = async () => {
    if (!webForm.url) {
      ElMessage.warning('请输入 URL')
      return
    }
    webLoading.value = true
    try {
      await importWeb(webForm)
      ElMessage.success('采集成功')
      webVisible.value = false
      Object.assign(webForm, { url: '', name: '', tags: '' })
      getList()
    } finally {
      webLoading.value = false
    }
  }

  // 问答对导入
  const qaVisible = ref(false)
  const qaLoading = ref(false)
  const qaFileRef = ref(null)
  const qaForm = reactive({ name: '' })
  const doImportQa = async () => {
    const file = qaFileRef.value?.files?.[0]
    if (!file) {
      ElMessage.warning('请选择 .json 或 .csv 文件')
      return
    }
    qaLoading.value = true
    try {
      const res = await importQa(file, qaForm.name, '')
      ElMessage.success(`导入 ${res.data?.imported ?? 0} 条（跳过 ${res.data?.skipped ?? 0}，去重 ${res.data?.dup ?? 0}）`)
      qaVisible.value = false
      qaForm.name = ''
      if (qaFileRef.value) qaFileRef.value.value = ''
      getList()
    } finally {
      qaLoading.value = false
    }
  }

  onMounted(getList)
</script>
