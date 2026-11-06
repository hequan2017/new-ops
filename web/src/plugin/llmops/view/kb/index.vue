<template>
  <div style="display: flex; gap: 12px; align-items: flex-start">
    <!-- 左：知识库列表 -->
    <div class="ops-card" style="width: 340px; flex-shrink: 0">
      <el-form inline>
        <el-form-item>
          <el-button type="primary" size="small" @click="openKbEdit()">新建知识库</el-button>
          <el-button size="small" @click="loadKbs">刷 新</el-button>
        </el-form-item>
      </el-form>
      <el-table :data="kbs" v-loading="kbLoading" stripe size="small" highlight-current-row @current-change="selectKb">
        <el-table-column prop="name" label="知识库" min-width="120" show-overflow-tooltip />
        <el-table-column label="类型" width="80">
          <template #default="{ row }">
            <el-tag size="small" :type="typeTag(row.type)">{{ typeLabel(row.type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="110" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click.stop="openKbEdit(row)">编辑</el-button>
            <el-button link type="danger" size="small" @click.stop="onDeleteKb(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 右：文档 + 检索测试台 -->
    <div style="flex: 1; min-width: 0">
      <div class="ops-card" v-if="currentKb">
        <el-form inline>
          <el-form-item>
            <span style="font-weight: 600">{{ currentKb.name }}</span>
            <el-tag size="small" style="margin-left: 8px">{{ typeLabel(currentKb.type) }}</el-tag>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" size="small" @click="ingestVisible = true">文档入库</el-button>
            <el-button v-if="currentKb.type === 'vector'" size="small" @click="doIndex('vector')">向量索引全部</el-button>
            <el-button v-if="currentKb.type === 'fulltext'" size="small" @click="doIndex('fulltext')">全文索引全部</el-button>
          </el-form-item>
        </el-form>
        <el-table :data="docs" v-loading="docLoading" stripe size="small">
          <el-table-column prop="title" label="文档" min-width="180" show-overflow-tooltip />
          <el-table-column prop="assetId" label="来源资产" width="90">
            <template #default="{ row }">{{ row.assetId || '内联' }}</template>
          </el-table-column>
          <el-table-column label="状态" width="90">
            <template #default="{ row }">
              <el-tag size="small" :type="row.status === 'indexed' ? 'success' : 'info'">{{ row.status === 'indexed' ? '已索引' : '就绪' }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="chunkCount" label="块数" width="70" />
          <el-table-column label="摘要" width="90">
            <template #default="{ row }">
              <el-button link type="primary" size="small" @click="openSummary(row)">查看/生成</el-button>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="80" fixed="right">
            <template #default="{ row }">
              <el-button link type="danger" size="small" @click="onDeleteDoc(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
      <div class="ops-card" v-else style="color: #999">← 请选择左侧知识库</div>

      <div class="ops-card" style="margin-top: 12px">
        <h4 style="margin: 0 0 10px">检索测试台（hybrid 五路融合 / 单路由）</h4>
        <el-form inline>
          <el-form-item label="模式">
            <el-select v-model="searchForm.mode" style="width: 120px">
              <el-option label="hybrid 融合" value="hybrid" />
              <el-option v-for="m in ['vector', 'fulltext', 'qa', 'graph', 'summary']" :key="m" :label="m" :value="m" />
            </el-select>
          </el-form-item>
          <el-form-item label="topK">
            <el-input-number v-model="searchForm.topK" :min="1" :max="50" size="small" />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" :loading="searchLoading" @click="doSearch">检 索</el-button>
          </el-form-item>
        </el-form>
        <el-input v-model="searchForm.query" placeholder="输入检索问题（多库以全部启用库为目标）" @keyup.enter="doSearch" />
        <el-table v-if="searchHits.length" :data="searchHits" stripe size="small" style="margin-top: 10px">
          <el-table-column prop="route" label="路由" width="90">
            <template #default="{ row }"><el-tag size="small">{{ row.route }}</el-tag></template>
          </el-table-column>
          <el-table-column label="融合分" width="90">
            <template #default="{ row }">{{ row.rrf?.toFixed?.(4) ?? row.rrf }}</template>
          </el-table-column>
          <el-table-column prop="content" label="内容" min-width="300" show-overflow-tooltip />
        </el-table>
      </div>
    </div>

    <!-- 新建/编辑库 -->
    <el-dialog v-model="kbVisible" :title="kbForm.id ? '编辑知识库' : '新建知识库'" width="520px">
      <el-form label-width="100px">
        <el-form-item label="名称">
          <el-input v-model="kbForm.name" />
        </el-form-item>
        <el-form-item label="类型">
          <el-select v-model="kbForm.type" style="width: 220px">
            <el-option label="向量语义库" value="vector" />
            <el-option label="全文检索库" value="fulltext" />
            <el-option label="知识图谱库" value="graph" />
            <el-option label="问答对库" value="qa" />
          </el-select>
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="kbForm.desc" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="kbVisible = false">取 消</el-button>
        <el-button type="primary" :loading="kbSaving" @click="doSaveKb">保 存</el-button>
      </template>
    </el-dialog>

    <!-- 文档入库 -->
    <el-dialog v-model="ingestVisible" title="文档入库" width="560px">
      <el-form label-width="100px">
        <el-form-item label="来源方式">
          <el-radio-group v-model="ingestForm.useAsset">
            <el-radio-button :value="true">选择资产</el-radio-button>
            <el-radio-button :value="false">内联文本</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item v-if="ingestForm.useAsset" label="数据资产">
          <el-select v-model="ingestForm.assetId" filterable placeholder="选择已解析资产" style="width: 100%">
            <el-option v-for="a in assets" :key="a.ID" :label="`${a.name}（${a.dataType}）`" :value="a.ID" />
          </el-select>
        </el-form-item>
        <template v-else>
          <el-form-item label="标题">
            <el-input v-model="ingestForm.title" />
          </el-form-item>
          <el-form-item label="文本">
            <el-input v-model="ingestForm.text" type="textarea" :rows="6" />
          </el-form-item>
        </template>
      </el-form>
      <template #footer>
        <el-button @click="ingestVisible = false">取 消</el-button>
        <el-button type="primary" :loading="ingestLoading" @click="doIngest">入 库</el-button>
      </template>
    </el-dialog>

    <!-- 摘要 -->
    <el-dialog v-model="summaryVisible" title="文档摘要" width="620px">
      <el-alert v-if="summaryTask" :title="`状态：${summaryTask.status}${summaryTask.errMsg ? ' · ' + summaryTask.errMsg : ''}`" :type="summaryTask.status === 'done' ? 'success' : summaryTask.status === 'failed' ? 'error' : 'warning'" :closable="false" style="margin-bottom: 10px" />
      <pre v-if="summaryTask?.summary" class="ops-pre">{{ summaryTask.summary }}</pre>
      <el-form label-width="100px" style="margin-top: 10px">
        <el-form-item label="chat 模型ID">
          <el-input-number v-model="summaryModelId" :min="1" size="small" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="summaryVisible = false">关 闭</el-button>
        <el-button type="primary" :loading="summaryLoading" @click="doGenerateSummary">生成/重新生成</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
  import { ref, reactive, onMounted } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import {
    getKbList, saveKb, deleteKb, ingestDoc, getKbDocs, deleteKbDoc,
    indexVector, indexFulltext, unifiedSearch, generateSummary, getDocSummary, getAssetList,
  } from '@/plugin/llmops/api/kb'

  const typeNames = { vector: '向量', fulltext: '全文', graph: '图谱', qa: '问答' }
  const typeLabel = (t) => typeNames[t] || t
  const typeTag = (t) => ({ vector: 'primary', fulltext: 'success', graph: 'warning', qa: 'info' }[t] || 'info')

  const kbs = ref([])
  const kbLoading = ref(false)
  const currentKb = ref(null)
  const loadKbs = async () => {
    kbLoading.value = true
    try {
      const res = await getKbList()
      kbs.value = res.data || []
    } finally {
      kbLoading.value = false
    }
  }
  const selectKb = async (row) => {
    currentKb.value = row
    if (row) await loadDocs()
  }

  const docs = ref([])
  const docLoading = ref(false)
  const loadDocs = async () => {
    if (!currentKb.value) return
    docLoading.value = true
    try {
      const res = await getKbDocs({ kbId: currentKb.value.ID, page: 1, pageSize: 50 })
      docs.value = res.data?.list || []
    } finally {
      docLoading.value = false
    }
  }

  // 库编辑
  const kbVisible = ref(false)
  const kbSaving = ref(false)
  const kbForm = reactive({ id: 0, name: '', type: 'vector', desc: '' })
  const openKbEdit = (row) => {
    Object.assign(kbForm, { id: row?.ID || 0, name: row?.name || '', type: row?.type || 'vector', desc: row?.desc || '' })
    kbVisible.value = true
  }
  const doSaveKb = async () => {
    kbSaving.value = true
    try {
      await saveKb(kbForm)
      ElMessage.success('保存成功')
      kbVisible.value = false
      loadKbs()
    } finally {
      kbSaving.value = false
    }
  }
  const onDeleteKb = (row) => {
    ElMessageBox.confirm(`确认删除知识库「${row.name}」？（级联删除库内文档）`, '提示', { type: 'warning' })
      .then(async () => {
        await deleteKb({ id: row.ID })
        ElMessage.success('删除成功')
        if (currentKb.value?.ID === row.ID) currentKb.value = null
        loadKbs()
      })
      .catch(() => {})
  }

  // 入库
  const ingestVisible = ref(false)
  const ingestLoading = ref(false)
  const ingestForm = reactive({ useAsset: true, assetId: null, title: '', text: '' })
  const assets = ref([])
  const loadAssets = async () => {
    const res = await getAssetList({ page: 1, pageSize: 100, status: 'parsed' })
    assets.value = res.data?.list || []
  }
  const doIngest = async () => {
    ingestLoading.value = true
    try {
      const payload = { kbId: currentKb.value.ID }
      if (ingestForm.useAsset) {
        if (!ingestForm.assetId) {
          ElMessage.warning('请选择资产')
          return
        }
        payload.assetId = ingestForm.assetId
      } else {
        if (!ingestForm.text) {
          ElMessage.warning('请输入文本')
          return
        }
        payload.title = ingestForm.title
        payload.text = ingestForm.text
      }
      await ingestDoc(payload)
      ElMessage.success('入库成功')
      ingestVisible.value = false
      loadDocs()
    } finally {
      ingestLoading.value = false
    }
  }

  // 索引
  const doIndex = async (kind) => {
    const fn = kind === 'vector' ? indexVector : indexFulltext
    let ok = 0
    for (const d of docs.value) {
      try {
        await fn({ kbId: currentKb.value.ID, docId: d.ID })
        ok++
      } catch (e) {
        /* 单文档失败继续 */
      }
    }
    ElMessage.success(`索引完成 ${ok}/${docs.value.length}`)
    loadDocs()
  }

  // 删文档
  const onDeleteDoc = (row) => {
    ElMessageBox.confirm(`确认删除文档「${row.title}」？`, '提示', { type: 'warning' })
      .then(async () => {
        await deleteKbDoc({ id: row.ID })
        ElMessage.success('删除成功')
        loadDocs()
      })
      .catch(() => {})
  }

  // 检索测试台
  const searchForm = reactive({ mode: 'hybrid', topK: 5, query: '' })
  const searchHits = ref([])
  const searchLoading = ref(false)
  const doSearch = async () => {
    if (!searchForm.query) {
      ElMessage.warning('请输入检索问题')
      return
    }
    searchLoading.value = true
    try {
      const payload = { kbIds: kbs.value.map((k) => k.ID), query: searchForm.query, mode: searchForm.mode, topK: searchForm.topK }
      if (!payload.kbIds.length) {
        ElMessage.warning('请先创建知识库')
        return
      }
      const res = await unifiedSearch(payload)
      searchHits.value = res.data?.hits || []
      if (!searchHits.value.length) ElMessage.info('无命中')
    } finally {
      searchLoading.value = false
    }
  }

  // 摘要
  const summaryVisible = ref(false)
  const summaryTask = ref(null)
  const summaryModelId = ref(1)
  const summaryLoading = ref(false)
  const summaryDoc = ref(null)
  const openSummary = async (row) => {
    summaryDoc.value = row
    summaryVisible.value = true
    const res = await getDocSummary({ kbId: currentKb.value.ID, docId: row.ID }).catch(() => null)
    summaryTask.value = res?.data || null
  }
  const doGenerateSummary = async () => {
    summaryLoading.value = true
    try {
      await generateSummary({ kbId: currentKb.value.ID, docId: summaryDoc.value.ID, modelId: summaryModelId.value })
      ElMessage.success('任务已提交，稍后刷新查看')
      setTimeout(() => openSummary(summaryDoc.value), 3000)
    } finally {
      summaryLoading.value = false
    }
  }

  onMounted(() => {
    loadKbs()
    loadAssets()
  })
</script>
