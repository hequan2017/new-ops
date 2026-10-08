<template>
  <div>
    <div class="ops-card">
      <el-form inline>
        <el-form-item>
          <el-button type="primary" @click="openEdit()">新建模板</el-button>
          <el-button @click="loadAll">刷 新</el-button>
        </el-form-item>
      </el-form>
      <el-table :data="templates" v-loading="loading" stripe>
        <el-table-column prop="name" label="模板名" min-width="140" />
        <el-table-column prop="category" label="分类" width="100">
          <template #default="{ row }">{{ row.category || '-' }}</template>
        </el-table-column>
        <el-table-column prop="description" label="描述" min-width="180" show-overflow-tooltip />
        <el-table-column label="参数" width="80">
          <template #default="{ row }">{{ parseSchema(row.paramsSchema).length }} 个</template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button link type="success" @click="openDeploy(row)">一键部署</el-button>
            <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
            <el-button link type="danger" @click="onDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <div class="ops-card" style="margin-top: 12px">
      <h4 style="margin: 0 0 10px">已部署实例（模板 → compose 项目追溯）</h4>
      <el-table :data="instances" v-loading="instLoading" stripe size="small">
        <el-table-column prop="projectName" label="项目" min-width="140" />
        <el-table-column prop="templateName" label="来源模板" width="140" />
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="row.composeStatus === '运行中' ? 'success' : row.composeStatus === '已删除' ? 'info' : 'warning'" size="small">
              {{ row.composeStatus }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createdBy" label="部署人" width="100" />
        <el-table-column label="部署时间" width="160">
          <template #default="{ row }">{{ (row.deployedAt || '').toString().replace('T', ' ').slice(0, 19) }}</template>
        </el-table-column>
        <el-table-column label="参数" min-width="180" show-overflow-tooltip>
          <template #default="{ row }">{{ row.params }}</template>
        </el-table-column>
        <el-table-column label="操作" width="90" fixed="right">
          <template #default="{ row }">
            <el-button v-if="row.composeProjectId" link type="primary"
              @click="$router.push({ name: 'containerCompose', query: { project: row.projectName } })">去管理</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <el-dialog v-model="editVisible" :title="form.ID ? '编辑模板' : '新建模板'" width="720px">
      <el-form label-width="90px">
        <el-form-item label="模板名">
          <el-input v-model="form.name" placeholder="如 nginx-web" />
        </el-form-item>
        <el-form-item label="分类">
          <el-input v-model="form.category" placeholder="web / db / 监控" style="width: 200px" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" />
        </el-form-item>
        <el-form-item label="参数定义">
          <el-input v-model="form.paramsSchema" type="textarea" :rows="4" :placeholder="schemaHint" />
          <div class="form-tip">JSON 数组；compose 内容中用两对花括号包 key 占位（如 two curly braces 包 image），部署时按此定义生成表单</div>
        </el-form-item>
        <el-form-item label="compose">
          <el-input v-model="form.composeTpl" type="textarea" :rows="10" :placeholder="composeHint" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.notes" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editVisible = false">取 消</el-button>
        <el-button type="primary" @click="submitEdit">保 存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="deployVisible" :title="`一键部署 · ${deployForm.templateName}`" width="560px">
      <el-form label-width="110px">
        <el-form-item label="Docker 接入点">
          <el-select v-model="deployForm.endpointId" style="width: 100%">
            <el-option v-for="ep in endpoints" :key="ep.ID" :label="`${ep.name}（${ep.addr}）`" :value="ep.ID" />
          </el-select>
        </el-form-item>
        <el-form-item label="项目名">
          <el-input v-model="deployForm.projectName" placeholder="小写字母数字与 -_（compose -p）" />
        </el-form-item>
        <el-form-item v-for="d in deployDefs" :key="d.key" :label="d.label || d.key">
          <el-input v-model="deployForm.params[d.key]" :placeholder="d.required ? '必填' : d.default || '选填'" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="deployVisible = false">取 消</el-button>
        <el-button type="primary" :loading="deploying" @click="submitDeploy">部 署</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
  import { ref, reactive, onMounted } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import {
    getAppTemplateList, saveAppTemplate, deleteAppTemplate, deployAppTemplate, getAppInstances
  } from '@/plugin/container/api/appTemplate'
  import { getEndpointList } from '@/plugin/container/api/dockerEndpoint'

  defineOptions({ name: 'containerAppTemplate' })

  const schemaHint =
    '[{"key":"image","label":"镜像","default":"nginx:latest","required":true},{"key":"port","label":"宿主端口","default":"8080"}]'
  const composeHint =
    'services:\n  web:\n    image: ' + '{{image}}' + '\n    ports:\n      - ' + "'{{port}}:80'"

  const templates = ref([])
  const instances = ref([])
  const loading = ref(false)
  const instLoading = ref(false)

  const parseSchema = (raw) => {
    try {
      const arr = JSON.parse(raw || '[]')
      return Array.isArray(arr) ? arr : []
    } catch (e) {
      return []
    }
  }

  const loadAll = async () => {
    loading.value = true
    try {
      const res = await getAppTemplateList()
      if (res.code === 0) templates.value = res.data || []
    } finally {
      loading.value = false
    }
    instLoading.value = true
    try {
      const res = await getAppInstances()
      if (res.code === 0) instances.value = res.data || []
    } finally {
      instLoading.value = false
    }
  }

  const editVisible = ref(false)
  const form = reactive({ ID: 0, name: '', category: '', description: '', composeTpl: '', paramsSchema: '', notes: '' })

  const openEdit = (row) => {
    Object.assign(form, {
      ID: row?.ID || 0, name: row?.name || '', category: row?.category || '',
      description: row?.description || '', composeTpl: row?.composeTpl || '',
      paramsSchema: row?.paramsSchema || '', notes: row?.notes || ''
    })
    editVisible.value = true
  }

  const submitEdit = async () => {
    if (!form.name.trim() || !form.composeTpl.trim()) {
      ElMessage.warning('模板名与 compose 内容必填')
      return
    }
    if (form.paramsSchema.trim()) {
      try {
        const arr = JSON.parse(form.paramsSchema)
        if (!Array.isArray(arr)) throw new Error('not array')
      } catch (e) {
        ElMessage.warning('参数定义必须为 JSON 数组')
        return
      }
    }
    const res = await saveAppTemplate({ ...form })
    if (res.code === 0) {
      ElMessage.success('已保存')
      editVisible.value = false
      loadAll()
    }
  }

  const onDelete = (row) => {
    ElMessageBox.confirm(`确定删除模板「${row.name}」吗？已部署实例不受影响。`, '提示', {
      confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning'
    }).then(async () => {
      const res = await deleteAppTemplate({ id: row.ID })
      if (res.code === 0) {
        ElMessage.success('已删除')
        loadAll()
      }
    })
  }

  const deployVisible = ref(false)
  const deploying = ref(false)
  const endpoints = ref([])
  const deployForm = reactive({ templateId: 0, templateName: '', endpointId: null, projectName: '', params: {} })
  const deployDefs = ref([])

  const openDeploy = async (row) => {
    const res = await getEndpointList()
    if (res.code === 0) endpoints.value = res.data || []
    deployForm.templateId = row.ID
    deployForm.templateName = row.name
    deployForm.endpointId = endpoints.value.length ? endpoints.value[0].ID : null
    deployForm.projectName = ''
    deployDefs.value = parseSchema(row.paramsSchema)
    const params = {}
    deployDefs.value.forEach((d) => {
      params[d.key] = d.default || ''
    })
    deployForm.params = params
    deployVisible.value = true
  }

  const submitDeploy = async () => {
    if (!deployForm.endpointId || !deployForm.projectName.trim()) {
      ElMessage.warning('请选择接入点并填写项目名')
      return
    }
    deploying.value = true
    try {
      const res = await deployAppTemplate({
        templateId: deployForm.templateId,
        endpointId: deployForm.endpointId,
        projectName: deployForm.projectName.trim(),
        params: deployForm.params
      })
      if (res.code === 0) {
        ElMessage.success('部署完成')
        deployVisible.value = false
        loadAll()
      }
    } finally {
      deploying.value = false
    }
  }

  onMounted(loadAll)
</script>

<style scoped>
  .form-tip {
    font-size: 12px;
    color: var(--el-text-color-secondary);
    line-height: 1.4;
    margin-top: 4px;
  }
</style>
