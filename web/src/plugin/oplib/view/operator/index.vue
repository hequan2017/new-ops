<template>
  <div>
    <!-- 算子 -->
    <div class="ops-card">
      <el-form inline>
        <el-form-item><h4 style="margin: 0">算子</h4></el-form-item>
        <el-form-item>
          <el-button type="primary" size="small" @click="openOpEdit()">新建算子</el-button>
          <el-button size="small" @click="loadOps">刷 新</el-button>
        </el-form-item>
        <el-form-item label="分类">
          <el-select v-model="opQuery.category" clearable placeholder="全部" style="width: 130px" @change="loadOps">
            <el-option v-for="c in categories" :key="c" :label="c" :value="c" />
          </el-select>
        </el-form-item>
      </el-form>
      <el-table :data="ops" v-loading="opLoading" stripe size="small">
        <el-table-column prop="name" label="算子名" width="170" show-overflow-tooltip />
        <el-table-column label="分类" width="100">
          <template #default="{ row }"><el-tag size="small">{{ row.category }}</el-tag></template>
        </el-table-column>
        <el-table-column prop="implType" label="实现" width="80" />
        <el-table-column prop="version" label="版本" width="70" />
        <el-table-column prop="desc" label="描述" min-width="180" show-overflow-tooltip />
        <el-table-column label="操作" width="170" fixed="right">
          <template #default="{ row }">
            <el-button link type="success" size="small" @click="openRun(row)">执行</el-button>
            <el-button link type="info" size="small" @click="openLogs(row)">留痕</el-button>
            <el-button link type="primary" size="small" @click="openOpEdit(row)">编辑</el-button>
            <el-button link type="danger" size="small" @click="onDeleteOp(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 管道 -->
    <div class="ops-card" style="margin-top: 12px">
      <el-form inline>
        <el-form-item><h4 style="margin: 0">算子管道（DAG 编排）</h4></el-form-item>
        <el-form-item>
          <el-button type="primary" size="small" @click="pipeVisible = true">新建管道</el-button>
          <el-button size="small" @click="loadPipes">刷 新</el-button>
        </el-form-item>
      </el-form>
      <el-table :data="pipes" v-loading="pipeLoading" stripe size="small">
        <el-table-column prop="name" label="管道名" width="170" show-overflow-tooltip />
        <el-table-column prop="desc" label="描述" min-width="200" show-overflow-tooltip />
        <el-table-column prop="status" label="状态" width="80" />
        <el-table-column label="操作" width="140" fixed="right">
          <template #default="{ row }">
            <el-button link type="success" size="small" @click="openPipeRun(row)">执行</el-button>
            <el-button link type="danger" size="small" @click="onDeletePipe(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 算子编辑 -->
    <el-dialog v-model="opVisible" :title="opForm.id ? '编辑算子' : '新建算子'" width="680px">
      <el-form label-width="100px">
        <el-form-item label="算子名"><el-input v-model="opForm.name" /></el-form-item>
        <el-form-item label="分类">
          <el-select v-model="opForm.category" style="width: 180px">
            <el-option v-for="c in categories" :key="c" :label="c" :value="c" />
          </el-select>
        </el-form-item>
        <el-form-item label="实现类型">
          <el-radio-group v-model="opForm.implType">
            <el-radio-button value="builtin">内置</el-radio-button>
            <el-radio-button value="script">脚本</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="入参 schema">
          <el-input v-model="opForm.inputSchema" type="textarea" :rows="3" placeholder='{"text":{"type":"string","required":true},"topN":{"type":"number"}}' />
        </el-form-item>
        <el-form-item label="出参 schema">
          <el-input v-model="opForm.outputSchema" type="textarea" :rows="2" />
        </el-form-item>
        <template v-if="opForm.implType === 'script'">
          <el-form-item label="解释器">
            <el-select v-model="opForm.scriptLang" style="width: 160px">
              <el-option label="python3" value="python3" />
              <el-option label="node" value="node" />
            </el-select>
          </el-form-item>
          <el-form-item label="脚本代码">
            <el-input v-model="opForm.scriptCode" type="textarea" :rows="6" placeholder="args 从 argv[1] 指向的 args.json 读取；输出 JSON" />
          </el-form-item>
        </template>
        <el-form-item label="版本"><el-input v-model="opForm.version" placeholder="v1" style="width: 120px" /></el-form-item>
        <el-form-item label="描述"><el-input v-model="opForm.desc" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="opVisible = false">取 消</el-button>
        <el-button type="primary" :loading="opSaving" @click="doSaveOp">保 存</el-button>
      </template>
    </el-dialog>

    <!-- 执行测试台 -->
    <el-dialog v-model="runVisible" :title="`执行 · ${runRow?.name || ''}`" width="640px">
      <el-input v-model="runArgs" type="textarea" :rows="5" placeholder='{"text":"输入内容"}' />
      <pre v-if="runResult" class="ops-pre" style="margin-top: 10px">{{ runResult }}</pre>
      <template #footer>
        <el-button @click="runVisible = false">关 闭</el-button>
        <el-button type="primary" :loading="runLoading" @click="doRun">执 行</el-button>
      </template>
    </el-dialog>

    <!-- 留痕抽屉 -->
    <el-drawer v-model="logsVisible" :title="`执行留痕 · ${logRow?.name || ''}`" size="60%">
      <el-table :data="logs" v-loading="logsLoading" stripe size="small">
        <el-table-column label="结果" width="70">
          <template #default="{ row }">
            <el-tag size="small" :type="row.ok ? 'success' : 'danger'">{{ row.ok ? '成功' : '失败' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="时延" width="80">
          <template #default="{ row }">{{ row.latencyMs }} ms</template>
        </el-table-column>
        <el-table-column prop="resultJSON" label="出参" min-width="220" show-overflow-tooltip />
        <el-table-column prop="errMsg" label="错误" min-width="140" show-overflow-tooltip />
      </el-table>
    </el-drawer>

    <!-- 管道编辑 -->
    <el-dialog v-model="pipeVisible" title="新建管道" width="680px">
      <el-form label-width="90px">
        <el-form-item label="管道名"><el-input v-model="pipeForm.name" /></el-form-item>
        <el-form-item label="描述"><el-input v-model="pipeForm.desc" /></el-form-item>
        <el-form-item label="编排定义">
          <el-input v-model="pipeForm.defJSON" type="textarea" :rows="8"
            placeholder='{"nodes":[{"id":"a","operator":"text_clean","args":{"text":"{input}"}},{"id":"b","operator":"keyword_extract"}],"edges":[["a","b"]]}' />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="pipeVisible = false">取 消</el-button>
        <el-button type="primary" :loading="pipeSaving" @click="doSavePipe">保 存</el-button>
      </template>
    </el-dialog>

    <!-- 管道执行 -->
    <el-dialog v-model="pipeRunVisible" :title="`执行管道 · ${pipeRunRow?.name || ''}`" width="700px">
      <el-input v-model="pipeInput" type="textarea" :rows="3" placeholder="管道输入文本" />
      <pre v-if="pipeResult" class="ops-pre" style="margin-top: 10px">{{ pipeResult }}</pre>
      <template #footer>
        <el-button @click="pipeRunVisible = false">关 闭</el-button>
        <el-button type="primary" :loading="pipeRunLoading" @click="doPipeRun">执 行</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
  import { ref, reactive, onMounted } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import {
    getOperators, saveOperator, runOperator, getRunLogs, deleteOperator,
    getPipelines, savePipeline, runPipeline, deletePipeline,
  } from '@/plugin/oplib/api/oplib'

  const categories = ['数据处理', '特征', '检索', '后处理', '评估']

  // 算子
  const ops = ref([])
  const opLoading = ref(false)
  const opQuery = reactive({ category: '' })
  const loadOps = async () => {
    opLoading.value = true
    try {
      const res = await getOperators(opQuery)
      ops.value = res.data || []
    } finally {
      opLoading.value = false
    }
  }
  const opVisible = ref(false)
  const opSaving = ref(false)
  const opForm = reactive({ id: 0, name: '', category: '数据处理', implType: 'builtin', inputSchema: '', outputSchema: '', scriptLang: 'python3', scriptCode: '', version: 'v1', desc: '' })
  const openOpEdit = (row) => {
    Object.assign(opForm, {
      id: row?.ID || 0, name: row?.name || '', category: row?.category || '数据处理',
      implType: row?.implType || 'builtin', inputSchema: row?.inputSchema || '', outputSchema: row?.outputSchema || '',
      scriptLang: row?.scriptLang || 'python3', scriptCode: row?.scriptCode || '',
      version: row?.version || 'v1', desc: row?.desc || '',
    })
    opVisible.value = true
  }
  const doSaveOp = async () => {
    opSaving.value = true
    try {
      await saveOperator(opForm)
      ElMessage.success('保存成功')
      opVisible.value = false
      loadOps()
    } finally {
      opSaving.value = false
    }
  }
  const runVisible = ref(false)
  const runRow = ref(null)
  const runArgs = ref('{}')
  const runResult = ref('')
  const runLoading = ref(false)
  const openRun = (row) => {
    runRow.value = row
    runArgs.value = '{}'
    runResult.value = ''
    runVisible.value = true
  }
  const doRun = async () => {
    runLoading.value = true
    try {
      const res = await runOperator({ id: runRow.value.ID, args: runArgs.value })
      runResult.value = JSON.stringify(res.data?.resultJSON ? JSON.parse(res.data.resultJSON) : res.data, null, 2)
    } finally {
      runLoading.value = false
    }
  }
  const logsVisible = ref(false)
  const logRow = ref(null)
  const logs = ref([])
  const logsLoading = ref(false)
  const openLogs = async (row) => {
    logRow.value = row
    logsVisible.value = true
    logsLoading.value = true
    try {
      const res = await getRunLogs({ operatorId: row.ID, page: 1, pageSize: 50 })
      logs.value = res.data?.list || []
    } finally {
      logsLoading.value = false
    }
  }
  const onDeleteOp = (row) => {
    ElMessageBox.confirm(`确认删除算子「${row.name}」？`, '提示', { type: 'warning' })
      .then(async () => {
        await deleteOperator({ id: row.ID })
        ElMessage.success('删除成功')
        loadOps()
      })
      .catch(() => {})
  }

  // 管道
  const pipes = ref([])
  const pipeLoading = ref(false)
  const loadPipes = async () => {
    pipeLoading.value = true
    try {
      const res = await getPipelines()
      pipes.value = res.data || []
    } finally {
      pipeLoading.value = false
    }
  }
  const pipeVisible = ref(false)
  const pipeSaving = ref(false)
  const pipeForm = reactive({ name: '', desc: '', defJSON: '' })
  const doSavePipe = async () => {
    pipeSaving.value = true
    try {
      await savePipeline(pipeForm)
      ElMessage.success('保存成功（拓扑校验通过）')
      pipeVisible.value = false
      loadPipes()
    } finally {
      pipeSaving.value = false
    }
  }
  const pipeRunVisible = ref(false)
  const pipeRunRow = ref(null)
  const pipeInput = ref('')
  const pipeResult = ref('')
  const pipeRunLoading = ref(false)
  const openPipeRun = (row) => {
    pipeRunRow.value = row
    pipeInput.value = ''
    pipeResult.value = ''
    pipeRunVisible.value = true
  }
  const doPipeRun = async () => {
    pipeRunLoading.value = true
    try {
      const res = await runPipeline({ id: pipeRunRow.value.ID, input: pipeInput.value })
      pipeResult.value = JSON.stringify(res.data, null, 2)
    } finally {
      pipeRunLoading.value = false
    }
  }
  const onDeletePipe = (row) => {
    ElMessageBox.confirm(`确认删除管道「${row.name}」？`, '提示', { type: 'warning' })
      .then(async () => {
        await deletePipeline({ id: row.ID })
        ElMessage.success('删除成功')
        loadPipes()
      })
      .catch(() => {})
  }

  onMounted(() => {
    loadOps()
    loadPipes()
  })
</script>
