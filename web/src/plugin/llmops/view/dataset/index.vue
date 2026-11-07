<template>
  <div>
    <div class="ops-card">
      <el-form inline>
        <el-form-item>
          <el-button type="primary" size="small" @click="openEdit()">新建数据集</el-button>
          <el-button size="small" @click="loadDatasets">刷 新</el-button>
        </el-form-item>
        <el-form-item label="用途">
          <el-select v-model="queryPurpose" clearable placeholder="全部" style="width: 130px" @change="loadDatasets">
            <el-option v-for="p in purposes" :key="p.value" :label="p.label" :value="p.value" />
          </el-select>
        </el-form-item>
      </el-form>
      <el-table :data="datasets" v-loading="loading" stripe size="small">
        <el-table-column prop="name" label="数据集" min-width="170" show-overflow-tooltip />
        <el-table-column label="用途" width="90">
          <template #default="{ row }">
            <el-tag size="small" :type="purposeTag(row.purpose)">{{ purposeLabel(row.purpose) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="format" label="格式" width="70" />
        <el-table-column prop="sampleCount" label="样本数" width="80" />
        <el-table-column prop="version" label="版本" width="60" />
        <el-table-column prop="license" label="许可证" width="110" show-overflow-tooltip />
        <el-table-column label="引用" width="110">
          <template #default="{ row }">
            <el-button link type="info" size="small" @click="showRefs(row)">查看引用</el-button>
          </template>
        </el-table-column>
        <el-table-column prop="desc" label="描述" min-width="140" show-overflow-tooltip />
        <el-table-column label="操作" width="210" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="openEdit(row)">编辑</el-button>
            <el-button link type="success" size="small" @click="openPreview(row)">预览</el-button>
            <el-button link type="warning" size="small" @click="openImport(row)">导入</el-button>
            <el-button link type="info" size="small" @click="doExport(row)">下载</el-button>
            <el-button link type="danger" size="small" @click="onDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 编辑 -->
    <el-dialog v-model="editVisible" :title="form.id ? '编辑数据集' : '新建数据集'" width="720px">
      <el-form label-width="90px">
        <el-form-item label="名称"><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="用途">
          <el-select v-model="form.purpose" style="width: 200px">
            <el-option v-for="p in purposes" :key="p.value" :label="p.label" :value="p.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="格式">
          <el-radio-group v-model="form.format">
            <el-radio-button value="jsonl">JSONL</el-radio-button>
            <el-radio-button value="csv">CSV</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="版本"><el-input v-model="form.version" placeholder="v1" style="width: 120px" /></el-form-item>
        <el-form-item label="许可证"><el-input v-model="form.license" placeholder="如 Apache-2.0" /></el-form-item>
        <el-form-item label="内容">
          <el-input v-model="form.content" type="textarea" :rows="8"
            :placeholder="schemaHint(form.purpose)" />
          <div style="font-size: 12px; color: #909399; margin-top: 4px">
            预计样本数：{{ samplePreview }}（保存时按格式精确计数）
          </div>
        </el-form-item>
        <el-form-item label="描述"><el-input v-model="form.desc" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editVisible = false">取 消</el-button>
        <el-button type="primary" :loading="saving" @click="doSave">保 存</el-button>
      </template>
    </el-dialog>

    <!-- 导入 -->
    <el-dialog v-model="importVisible" title="导入数据（JSONL 逐行校验）" width="720px">
      <el-alert v-if="importRow" :title="`目标：${importRow.name}（${importRow.purpose}，${importRow.format}）`" type="info" :closable="false" style="margin-bottom: 8px" />
      <el-input v-model="importContent" type="textarea" :rows="8" placeholder="粘贴 JSONL 内容，逐行校验后仅有效行入库" />
      <el-table v-if="importErrors.length" :data="importErrors" stripe size="small" style="margin-top: 8px">
        <el-table-column prop="line" label="行号" width="70" />
        <el-table-column prop="reason" label="原因" min-width="200" />
      </el-table>
      <template #footer>
        <el-button @click="importVisible = false">关 闭</el-button>
        <el-button type="primary" :loading="importing" @click="doImport">导 入</el-button>
      </template>
    </el-dialog>

    <!-- 样本预览 -->
    <el-drawer v-model="previewVisible" :title="`样本预览 · ${previewRow?.name || ''}`" size="55%">
      <el-alert v-if="previewRow" :title="`${previewRow.sampleCount} 样本 · ${previewRow.format} · ${purposeLabel(previewRow.purpose)}`" type="info" :closable="false" style="margin-bottom: 10px" />
      <pre class="ops-pre">{{ previewText }}</pre>
    </el-drawer>

    <!-- 引用 -->
    <el-dialog v-model="refsVisible" :title="`引用情况 · ${refsRow?.name || ''}`" width="420px">
      <el-descriptions :column="1" border>
        <el-descriptions-item label="训练任务引用">{{ refs?.trainJobs ?? '—' }}</el-descriptions-item>
        <el-descriptions-item label="评估任务引用">{{ refs?.evalRuns ?? '—' }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup>
  import { ref, reactive, computed, onMounted } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import {
    getDatasets, saveDataset, deleteDataset, importDataset, exportDataset, getDatasetRefs,
  } from '@/plugin/llmops/api/dataset'

  const purposes = [
    { value: 'sft', label: 'SFT 训练' }, { value: 'rlhf', label: 'RLHF 偏好' },
    { value: 'eval', label: '评测' }, { value: 'exam', label: '考试题集' },
  ]
  const purposeLabel = (p) => (purposes.find((x) => x.value === p) || {}).label || p
  const purposeTag = (p) => ({ sft: 'primary', rlhf: 'warning', eval: 'success', exam: 'danger' }[p] || 'info')

  const schemaHints = {
    sft: '每行 {"instruction":"指令","output":"回答"}',
    rlhf: '每行 {"prompt":"问题","chosen":"优选回答","rejected":"拒绝回答"}',
    eval: '每行 {"question":"问题","expected":"期望答案"}',
    exam: '每行 {"type":"single","stem":"题干","options":["A","B"],"answer":"A"}',
  }
  const schemaHint = (p) => schemaHints[p] || '每行一个 JSON 对象'

  const datasets = ref([])
  const loading = ref(false)
  const queryPurpose = ref('')
  const loadDatasets = async () => {
    loading.value = true
    try {
      const res = await getDatasets({ purpose: queryPurpose.value })
      datasets.value = res.data || []
    } finally {
      loading.value = false
    }
  }

  // 编辑
  const editVisible = ref(false)
  const saving = ref(false)
  const form = reactive({ id: 0, name: '', purpose: 'sft', format: 'jsonl', version: 'v1', license: '', content: '', desc: '' })
  const samplePreview = computed(() => {
    if (!form.content.trim()) return 0
    if (form.format === 'csv') {
      return Math.max(0, form.content.split('\n').filter((l) => l.trim()).length - 1)
    }
    return form.content.split('\n').filter((l) => l.trim()).length
  })
  const openEdit = (row) => {
    Object.assign(form, {
      id: row?.ID || 0, name: row?.name || '', purpose: row?.purpose || 'sft',
      format: row?.format || 'jsonl', version: row?.version || 'v1', license: row?.license || '',
      content: row?.content || '', desc: row?.desc || '',
    })
    editVisible.value = true
  }
  const doSave = async () => {
    saving.value = true
    try {
      await saveDataset(form)
      ElMessage.success('保存成功')
      editVisible.value = false
      loadDatasets()
    } finally {
      saving.value = false
    }
  }
  const onDelete = (row) => {
    ElMessageBox.confirm(`确认删除数据集「${row.name}」？`, '提示', { type: 'warning' })
      .then(async () => {
        await deleteDataset({ id: row.ID })
        ElMessage.success('删除成功')
        loadDatasets()
      })
      .catch(() => {})
  }

  // 导入
  const importVisible = ref(false)
  const importing = ref(false)
  const importRow = ref(null)
  const importContent = ref('')
  const importErrors = ref([])
  const openImport = (row) => {
    importRow.value = row
    importContent.value = ''
    importErrors.value = []
    importVisible.value = true
  }
  const doImport = async () => {
    importing.value = true
    try {
      const res = await importDataset({ id: importRow.value.ID, content: importContent.value })
      importErrors.value = res.data?.errors || []
      ElMessage.success(`导入 ${res.data?.valid ?? 0} 条有效样本${importErrors.value.length ? `，${importErrors.value.length} 条坏行` : ''}`)
      loadDatasets()
    } finally {
      importing.value = false
    }
  }

  // 预览
  const previewVisible = ref(false)
  const previewRow = ref(null)
  const previewText = computed(() => {
    if (!previewRow.value?.content) return '（空）'
    const lines = previewRow.value.content.split('\n').filter((l) => l.trim())
    const head = lines.slice(0, 20).map((l) => {
      try { return JSON.stringify(JSON.parse(l), null, 2) } catch (e) { return l }
    })
    return head.join('\n---\n') + (lines.length > 20 ? `\n…共 ${lines.length} 行` : '')
  })
  const openPreview = (row) => {
    previewRow.value = row
    previewVisible.value = true
  }

  // 引用
  const refsVisible = ref(false)
  const refsRow = ref(null)
  const refs = ref(null)
  const showRefs = async (row) => {
    refsRow.value = row
    const res = await getDatasetRefs({ id: row.ID })
    refs.value = res.data
    refsVisible.value = true
  }

  // 导出下载
  const doExport = async (row) => {
    const res = await exportDataset({ id: row.ID })
    const blob = new Blob([res.data?.content || ''], { type: 'text/plain;charset=utf-8' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = res.data?.filename || `${row.name}.jsonl`
    a.click()
    URL.revokeObjectURL(url)
  }

  onMounted(loadDatasets)
</script>
