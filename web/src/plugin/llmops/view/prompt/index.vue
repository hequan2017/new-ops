<template>
  <div>
    <!-- 模板 -->
    <div class="ops-card">
      <el-form inline>
        <el-form-item><h4 style="margin: 0">Prompt 模板</h4></el-form-item>
        <el-form-item>
          <el-button type="primary" size="small" @click="openEdit()">新建模板</el-button>
          <el-button size="small" @click="loadPrompts">刷 新</el-button>
        </el-form-item>
      </el-form>
      <el-table :data="prompts" v-loading="loading" stripe size="small">
        <el-table-column prop="name" label="模板名" width="170" show-overflow-tooltip />
        <el-table-column label="版本" width="70">
          <template #default="{ row }">v{{ row.version }}</template>
        </el-table-column>
        <el-table-column label="变量" min-width="200">
          <template #default="{ row }">
            <el-tag v-for="v in varsOf(row.content)" :key="v" size="small" style="margin-right: 4px">{{ v }}</el-tag>
            <span v-if="!varsOf(row.content).length" style="color: #999">无变量</span>
          </template>
        </el-table-column>
        <el-table-column prop="desc" label="描述" min-width="140" show-overflow-tooltip />
        <el-table-column label="操作" width="240" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="openEdit(row)">编辑</el-button>
            <el-button link type="success" size="small" @click="openRender(row)">渲染测试</el-button>
            <el-button link type="warning" size="small" @click="openOptimize(row)">优化</el-button>
            <el-button link type="info" size="small" @click="openVersions(row)">版本</el-button>
            <el-button link type="danger" size="small" @click="onDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 配套资产 -->
    <div class="ops-card" style="margin-top: 12px">
      <el-form inline>
        <el-form-item><h4 style="margin: 0">配套资产</h4></el-form-item>
        <el-form-item>
          <el-radio-group v-model="assetKind" size="small" @change="loadAssets">
            <el-radio-button value="rule">业务规则</el-radio-button>
            <el-radio-button value="term">行业术语库</el-radio-button>
            <el-radio-button value="cot">专家思维链</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" size="small" @click="openAssetEdit()">新增{{ kindLabel }}</el-button>
        </el-form-item>
      </el-form>
      <el-table :data="assets" v-loading="assetLoading" stripe size="small">
        <el-table-column prop="name" label="名称" width="180" show-overflow-tooltip />
        <el-table-column prop="content" label="内容" min-width="300" show-overflow-tooltip />
        <el-table-column label="引用模板" width="90">
          <template #default="{ row }">{{ row.templateId || '—' }}</template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="openAssetEdit(row)">编辑</el-button>
            <el-button link type="danger" size="small" @click="onDeleteAsset(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 编辑模板 -->
    <el-dialog v-model="editVisible" :title="form.id ? '编辑模板' : '新建模板'" width="680px">
      <el-form label-width="90px">
        <el-form-item label="模板名"><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="描述"><el-input v-model="form.desc" /></el-form-item>
        <el-form-item label="内容">
          <el-input v-model="form.content" type="textarea" :rows="7" placeholder="支持双花括号变量占位，如：把变量 text 翻译成变量 lang 对应语言" />
          <div v-if="form.content" style="font-size: 12px; color: #909399; margin-top: 4px">
            识别到变量：{{ varsOf(form.content).join('、') || '无' }}
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editVisible = false">取 消</el-button>
        <el-button type="primary" :loading="saving" @click="doSave">保 存</el-button>
      </template>
    </el-dialog>

    <!-- 渲染测试台 -->
    <el-dialog v-model="renderVisible" :title="`渲染测试 · ${renderRow?.name || ''}`" width="640px">
      <el-form label-width="110px">
        <el-form-item v-for="v in renderVars" :key="v" :label="v">
          <el-input v-model="renderInputs[v]" placeholder="变量值" />
        </el-form-item>
      </el-form>
      <pre v-if="rendered" class="ops-pre">{{ rendered }}</pre>
      <template #footer>
        <el-button @click="renderVisible = false">关 闭</el-button>
        <el-button type="primary" :loading="renderLoading" @click="doRender">渲 染</el-button>
      </template>
    </el-dialog>

    <!-- 优化对比 -->
    <el-dialog v-model="optVisible" :title="`Prompt 优化 · ${optRow?.name || ''}`" width="760px">
      <el-input v-model="optRequirement" placeholder="优化需求（空走通用优化），如：更简洁、面向客服场景" style="margin-bottom: 10px" />
      <div style="display: flex; gap: 10px">
        <div style="flex: 1">
          <h5 style="margin: 0 0 6px">原版</h5>
          <pre class="ops-pre">{{ optResult?.original || optRow?.content }}</pre>
        </div>
        <div style="flex: 1">
          <h5 style="margin: 0 0 6px">优化版</h5>
          <pre class="ops-pre">{{ optResult?.optimized || '（执行优化后显示）' }}</pre>
        </div>
      </div>
      <template #footer>
        <el-button @click="optVisible = false">关 闭</el-button>
        <el-button type="primary" :loading="optLoading" @click="doOptimize">执 行 优 化</el-button>
      </template>
    </el-dialog>

    <!-- 版本快照 -->
    <el-dialog v-model="verVisible" :title="`版本快照 · ${verRow?.name || ''}`" width="640px">
      <el-table :data="versions" stripe size="small">
        <el-table-column prop="version" label="版本" width="70" />
        <el-table-column prop="content" label="内容" min-width="300" show-overflow-tooltip />
      </el-table>
    </el-dialog>

    <!-- 资产编辑 -->
    <el-dialog v-model="assetVisible" :title="`编辑${kindLabel}`" width="560px">
      <el-form label-width="90px">
        <el-form-item label="名称"><el-input v-model="assetForm.name" /></el-form-item>
        <el-form-item label="内容"><el-input v-model="assetForm.content" type="textarea" :rows="5" /></el-form-item>
        <el-form-item label="引用模板ID"><el-input-number v-model="assetForm.templateId" :min="0" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="assetVisible = false">取 消</el-button>
        <el-button type="primary" :loading="assetSaving" @click="doSaveAsset">保 存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
  import { ref, reactive, computed, onMounted } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import {
    getPrompts, savePrompt, deletePrompt, renderPrompt, optimizePrompt, getPromptVersions,
    getPromptAssets, savePromptAsset, deletePromptAsset,
  } from '@/plugin/llmops/api/prompt'

  const reVar = /\{\{\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*\}\}/g
  const varsOf = (content) => {
    const seen = new Set()
    const out = []
    for (const m of (content || '').matchAll(reVar)) {
      if (!seen.has(m[1])) { seen.add(m[1]); out.push(m[1]) }
    }
    return out
  }

  // 模板
  const prompts = ref([])
  const loading = ref(false)
  const loadPrompts = async () => {
    loading.value = true
    try {
      const res = await getPrompts()
      prompts.value = res.data || []
    } finally {
      loading.value = false
    }
  }
  const editVisible = ref(false)
  const saving = ref(false)
  const form = reactive({ id: 0, name: '', desc: '', content: '' })
  const openEdit = (row) => {
    Object.assign(form, { id: row?.ID || 0, name: row?.name || '', desc: row?.desc || '', content: row?.content || '' })
    editVisible.value = true
  }
  const doSave = async () => {
    saving.value = true
    try {
      await savePrompt(form)
      ElMessage.success('保存成功')
      editVisible.value = false
      loadPrompts()
    } finally {
      saving.value = false
    }
  }
  const onDelete = (row) => {
    ElMessageBox.confirm(`确认删除模板「${row.name}」？（级联删除版本快照）`, '提示', { type: 'warning' })
      .then(async () => {
        await deletePrompt({ id: row.ID })
        ElMessage.success('删除成功')
        loadPrompts()
      })
      .catch(() => {})
  }

  // 渲染测试台
  const renderVisible = ref(false)
  const renderRow = ref(null)
  const renderVars = ref([])
  const renderInputs = reactive({})
  const rendered = ref('')
  const renderLoading = ref(false)
  const openRender = (row) => {
    renderRow.value = row
    renderVars.value = varsOf(row.content)
    Object.keys(renderInputs).forEach((k) => delete renderInputs[k])
    rendered.value = ''
    renderVisible.value = true
  }
  const doRender = async () => {
    renderLoading.value = true
    try {
      const res = await renderPrompt({ templateId: renderRow.value.ID, vars: { ...renderInputs } })
      rendered.value = res.data?.rendered || ''
    } finally {
      renderLoading.value = false
    }
  }

  // 优化对比
  const optVisible = ref(false)
  const optRow = ref(null)
  const optRequirement = ref('')
  const optResult = ref(null)
  const optLoading = ref(false)
  const openOptimize = (row) => {
    optRow.value = row
    optResult.value = null
    optRequirement.value = ''
    optVisible.value = true
  }
  const doOptimize = async () => {
    optLoading.value = true
    try {
      const res = await optimizePrompt({ items: [{ templateId: optRow.value.ID, requirement: optRequirement.value }] })
      optResult.value = res.data?.[0] || null
      if (optResult.value?.err) ElMessage.warning(optResult.value.err)
    } finally {
      optLoading.value = false
    }
  }

  // 版本
  const verVisible = ref(false)
  const verRow = ref(null)
  const versions = ref([])
  const openVersions = async (row) => {
    verRow.value = row
    const res = await getPromptVersions({ templateId: row.ID })
    versions.value = res.data || []
    verVisible.value = true
  }

  // 资产
  const kindNames = { rule: '业务规则', term: '行业术语', cot: '专家思维链' }
  const kindLabel = computed(() => kindNames[assetKind.value] || '资产')
  const assetKind = ref('rule')
  const assets = ref([])
  const assetLoading = ref(false)
  const loadAssets = async () => {
    assetLoading.value = true
    try {
      const res = await getPromptAssets({ kind: assetKind.value })
      assets.value = res.data || []
    } finally {
      assetLoading.value = false
    }
  }
  const assetVisible = ref(false)
  const assetSaving = ref(false)
  const assetForm = reactive({ id: 0, name: '', kind: 'rule', content: '', templateId: 0 })
  const openAssetEdit = (row) => {
    Object.assign(assetForm, {
      id: row?.ID || 0, name: row?.name || '', kind: assetKind.value,
      content: row?.content || '', templateId: row?.templateId || 0,
    })
    assetVisible.value = true
  }
  const doSaveAsset = async () => {
    assetSaving.value = true
    try {
      await savePromptAsset(assetForm)
      ElMessage.success('保存成功')
      assetVisible.value = false
      loadAssets()
    } finally {
      assetSaving.value = false
    }
  }
  const onDeleteAsset = (row) => {
    ElMessageBox.confirm(`确认删除「${row.name}」？`, '提示', { type: 'warning' })
      .then(async () => {
        await deletePromptAsset({ id: row.ID })
        ElMessage.success('删除成功')
        loadAssets()
      })
      .catch(() => {})
  }

  onMounted(() => {
    loadPrompts()
    loadAssets()
  })
</script>
