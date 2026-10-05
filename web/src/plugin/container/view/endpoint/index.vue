<template>
  <div>
    <div class="ops-card">
      <el-form inline>
        <el-form-item label="关键字">
          <el-input v-model="keyword" placeholder="名称/地址/备注" clearable style="width: 180px" @keyup.enter="getList" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="status" clearable style="width: 110px" @change="getList">
            <el-option v-for="s in ['在线', '离线', '未知']" :key="s" :label="s" :value="s" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="getList">查 询</el-button>
          <el-button @click="openDialog()">新建接入点</el-button>
        </el-form-item>
      </el-form>

      <el-table :data="tableData" v-loading="loading" stripe>
        <el-table-column prop="ID" label="ID" width="60" />
        <el-table-column prop="name" label="名称" min-width="130" show-overflow-tooltip />
        <el-table-column prop="addr" label="地址" min-width="220" show-overflow-tooltip />
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="{ 在线: 'success', 离线: 'danger' }[row.status] || 'info'" size="small">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="dockerVersion" label="API版本" width="100" />
        <el-table-column label="最近巡检" width="160">
          <template #default="{ row }">
            {{ (row.lastCheckAt || '—').replace('T', ' ').slice(0, 19) }}
          </template>
        </el-table-column>
        <el-table-column prop="notes" label="备注" min-width="120" show-overflow-tooltip />
        <el-table-column label="操作" width="230" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openDialog(row)">编辑</el-button>
            <el-button link type="success" :loading="checkingId === row.ID" @click="onCheck(row)">巡检</el-button>
            <el-button link type="primary" @click="openContainers(row)">容器</el-button>
            <el-button link type="danger" @click="onDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <el-dialog v-model="dialogVisible" :title="form.ID ? '编辑接入点' : '新建接入点'" width="620px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="form.name" placeholder="如 本机Docker" />
        </el-form-item>
        <el-form-item label="地址" prop="addr">
          <el-input v-model="form.addr" placeholder="unix:///var/run/docker.sock 或 tcp://host:2376" />
        </el-form-item>
        <el-form-item label="TLS 凭据">
          <el-select v-model="form.tlsCredentialId" clearable placeholder="tcp+tls 时选择 docker_tls 凭据" style="width: 100%">
            <el-option
              v-for="c in tlsCredOptions"
              :key="c.ID"
              :label="`${c.name}（TLS）`"
              :value="c.ID"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.notes" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取 消</el-button>
        <el-button type="primary" @click="submitForm">确 定</el-button>
      </template>
    </el-dialog>

    <el-drawer v-model="ctVisible" :title="`容器 · ${ctEndpoint?.name || ''}`" size="70%">
      <el-form inline>
        <el-form-item>
          <el-checkbox v-model="ctAll" @change="loadContainers">含已停止</el-checkbox>
        </el-form-item>
        <el-form-item>
          <el-button :loading="ctLoading" @click="loadContainers">刷 新</el-button>
          <el-button type="primary" @click="createVisible = true">创建容器</el-button>
        </el-form-item>
      </el-form>
      <el-table :data="ctList" v-loading="ctLoading" stripe size="small">
        <el-table-column prop="id" label="ID" width="100" />
        <el-table-column label="名称" min-width="140">
          <template #default="{ row }">{{ prettyName(row.names) }}</template>
        </el-table-column>
        <el-table-column prop="image" label="镜像" min-width="160" show-overflow-tooltip />
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="row.state === 'running' ? 'success' : 'info'" size="small">{{ row.state }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="详情" min-width="120" show-overflow-tooltip />
        <el-table-column prop="ports" label="端口" min-width="140" show-overflow-tooltip />
        <el-table-column label="操作" width="290" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openContainerLogs(row)">日志</el-button>
            <el-button v-if="row.state === 'running'" link type="success" @click="openContainerShell(row)">终端</el-button>
            <el-button v-if="row.state !== 'running'" link type="success" @click="doAction(row, 'start')">启动</el-button>
            <el-button v-if="row.state === 'running'" link type="warning" @click="doAction(row, 'stop')">停止</el-button>
            <el-button v-if="row.state === 'running'" link type="primary" @click="doAction(row, 'restart')">重启</el-button>
            <el-button link type="danger" @click="doRemove(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-drawer>

    <el-dialog v-model="createVisible" title="创建容器" width="680px" append-to-body>
      <el-form ref="createFormRef" :model="createForm" :rules="createRules" label-width="90px">
        <el-row :gutter="12">
          <el-col :span="12">
            <el-form-item label="容器名">
              <el-input v-model="createForm.name" placeholder="留空自动分配" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="镜像" prop="image">
              <el-input v-model="createForm.image" placeholder="如 nginx:alpine" />
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="端口映射">
              <el-input v-model="portsText" type="textarea" :rows="2" placeholder="每行一条：8080:80/tcp" />
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="环境变量">
              <el-input v-model="envsText" type="textarea" :rows="2" placeholder="每行一条：KEY=value" />
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="挂载">
              <el-input v-model="mountsText" type="textarea" :rows="2" placeholder="每行一条：/host:/ct:rw" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="CPU核">
              <el-input-number v-model="createForm.cpuCores" :min="0" :step="0.5" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="内存MB">
              <el-input-number v-model="createForm.memoryMb" :min="0" :step="128" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="重启策略">
              <el-select v-model="createForm.restartPolicy" style="width: 100%">
                <el-option v-for="p in ['no', 'always', 'unless-stopped', 'on-failure']" :key="p" :label="p" :value="p" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="启动命令">
              <el-input v-model="cmdText" placeholder="空格分隔，如 sleep 3600（可选）" />
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label-width="90px">
              <el-checkbox v-model="createForm.startNow">创建后立即启动</el-checkbox>
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取 消</el-button>
        <el-button type="primary" :loading="creating" @click="submitCreate">创 建</el-button>
      </template>
    </el-dialog>

    <el-drawer v-model="shellVisible" :title="`终端 · ${shellName}`" size="70%" append-to-body>
      <ContainerShell
        v-if="shellVisible"
        :endpoint-id="ctEndpoint?.ID || 0"
        :container-id="shellCid"
        :label="shellName"
      />
    </el-drawer>

    <el-drawer v-model="logDrawerVisible" :title="`日志 · ${shellName}`" size="70%" append-to-body>
      <ContainerLogs
        v-if="logDrawerVisible"
        :endpoint-id="ctEndpoint?.ID || 0"
        :container-id="shellCid"
      />
    </el-drawer>
  </div>
</template>

<script setup>
  import { ref, reactive, onMounted, onUnmounted } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import {
    createEndpoint, updateEndpoint, deleteEndpoint, getEndpointList, checkEndpoint
  } from '@/plugin/container/api/dockerEndpoint'
  import { getContainerList, containerAction, createContainer } from '@/plugin/container/api/container'
  import ContainerShell from '@/plugin/container/components/ContainerShell.vue'
  import ContainerLogs from '@/plugin/container/components/ContainerLogs.vue'
  import { getCredentialList } from '@/plugin/asset/api/credential'

  defineOptions({ name: 'containerEndpoint' })

  const keyword = ref('')
  const status = ref('')
  const tableData = ref([])
  const loading = ref(false)
  const tlsCredOptions = ref([])
  const checkingId = ref(0)
  let pollTimer = null

  const loadCreds = async () => {
    const res = await getCredentialList()
    if (res.code === 0) {
      tlsCredOptions.value = (res.data || []).filter((x) => x.type === 'docker_tls')
    }
  }

  const getList = async () => {
    loading.value = true
    try {
      const res = await getEndpointList({ keyword: keyword.value || undefined, status: status.value || undefined })
      if (res.code === 0) tableData.value = res.data || []
    } finally {
      loading.value = false
    }
  }

  const onCheck = async (row) => {
    checkingId.value = row.ID
    try {
      const res = await checkEndpoint({ id: row.ID })
      if (res.code === 0) {
        ElMessage.success(`巡检通过：API ${res.data.version}`)
      }
      getList()
    } finally {
      checkingId.value = 0
    }
  }

  const dialogVisible = ref(false)
  const formRef = ref(null)
  const emptyForm = () => ({ ID: 0, name: '', addr: '', tlsCredentialId: null, notes: '' })
  const form = reactive(emptyForm())
  const rules = {
    name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
    addr: [{ required: true, message: '请输入地址', trigger: 'blur' }]
  }

  const openDialog = (row) => {
    Object.assign(form, emptyForm())
    if (row) Object.assign(form, row)
    dialogVisible.value = true
  }

  const submitForm = () => {
    formRef.value.validate(async (valid) => {
      if (!valid) return
      const res = form.ID ? await updateEndpoint(form) : await createEndpoint(form)
      if (res.code === 0) {
        ElMessage.success(res.msg)
        dialogVisible.value = false
        getList()
      }
    })
  }

  const onDelete = (row) => {
    ElMessageBox.confirm(`确定删除接入点「${row.name}」吗？`, '提示', {
      confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning'
    }).then(async () => {
      const res = await deleteEndpoint({ id: row.ID })
      if (res.code === 0) {
        ElMessage.success('删除成功')
        getList()
      }
    })
  }

  // ---------- 容器抽屉 ----------
  const ctVisible = ref(false)
  const ctEndpoint = ref(null)
  const ctList = ref([])
  const ctLoading = ref(false)
  const ctAll = ref(true)

  const prettyName = (names) => (names && names.length ? names[0].replace(/^\//, '') : '—')

  const openContainers = (row) => {
    ctEndpoint.value = row
    ctVisible.value = true
    loadContainers()
  }

  const loadContainers = async () => {
    if (!ctEndpoint.value) return
    ctLoading.value = true
    try {
      const res = await getContainerList({ endpointId: ctEndpoint.value.ID, all: ctAll.value })
      if (res.code === 0) ctList.value = res.data || []
    } finally {
      ctLoading.value = false
    }
  }

  const doAction = async (row, action) => {
    const res = await containerAction({
      endpointId: ctEndpoint.value.ID,
      id: row.id,
      action
    })
    if (res.code === 0) {
      ElMessage.success(`${action} 成功`)
      loadContainers()
    }
  }

  const doRemove = (row) => {
    ElMessageBox.confirm(`确定删除容器 ${prettyName(row.names)} 吗？`, '危险操作', {
      confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning'
    }).then(async () => {
      const res = await containerAction({
        endpointId: ctEndpoint.value.ID,
        id: row.id,
        action: 'remove',
        force: true
      })
      if (res.code === 0) {
        ElMessage.success('已删除')
        loadContainers()
      }
    })
  }

  // ---------- 创建容器 ----------
  const createVisible = ref(false)
  const creating = ref(false)
  const createFormRef = ref(null)
  const createForm = reactive({
    name: '',
    image: '',
    cpuCores: 0,
    memoryMb: 0,
    restartPolicy: 'no',
    startNow: true
  })
  const portsText = ref('')
  const envsText = ref('')
  const mountsText = ref('')
  const cmdText = ref('')
  const createRules = { image: [{ required: true, message: '请输入镜像', trigger: 'blur' }] }

  const lines = (s) => s.split('\n').map((x) => x.trim()).filter(Boolean)

  const submitCreate = () => {
    createFormRef.value.validate(async (valid) => {
      if (!valid) return
      creating.value = true
      try {
        const res = await createContainer({
          endpointId: ctEndpoint.value.ID,
          name: createForm.name,
          image: createForm.image,
          command: cmdText.value ? cmdText.value.split(/\s+/) : [],
          ports: lines(portsText.value),
          envs: lines(envsText.value),
          mounts: lines(mountsText.value),
          cpuCores: createForm.cpuCores,
          memoryMb: createForm.memoryMb,
          restartPolicy: createForm.restartPolicy,
          startNow: createForm.startNow
        })
        if (res.code === 0) {
          ElMessage.success(`已创建：${res.data.id}`)
          createVisible.value = false
          loadContainers()
        }
      } finally {
        creating.value = false
      }
    })
  }

  // ---------- 容器终端/日志 ----------
  const shellVisible = ref(false)
  const logDrawerVisible = ref(false)
  const shellCid = ref('')
  const shellName = ref('')

  const openContainerShell = (row) => {
    shellCid.value = row.id
    shellName.value = prettyName(row.names)
    shellVisible.value = true
  }
  const openContainerLogs = (row) => {
    shellCid.value = row.id
    shellName.value = prettyName(row.names)
    logDrawerVisible.value = true
  }

  onMounted(() => {
    loadCreds()
    getList()
    pollTimer = setInterval(getList, 35000) // 跟随 30s 巡检周期轻刷新
  })
  onUnmounted(() => {
    if (pollTimer) clearInterval(pollTimer)
  })
</script>
