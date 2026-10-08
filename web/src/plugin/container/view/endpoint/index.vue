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
        <el-table-column label="操作" width="400" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openDialog(row)">编辑</el-button>
            <el-button link type="success" :loading="checkingId === row.ID" @click="onCheck(row)">巡检</el-button>
            <el-button link type="primary" @click="openContainers(row)">容器</el-button>
            <el-button link type="warning" @click="openEvents(row)">事件</el-button>
            <el-button link type="success" @click="openImages(row)">镜像</el-button>
            <el-button link type="warning" @click="openResources(row)">网络卷</el-button>
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
        <el-table-column label="操作" width="330" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openContainerLogs(row)">日志</el-button>
            <el-button v-if="row.state === 'running'" link type="success" @click="openContainerShell(row)">终端</el-button>
            <el-button v-if="row.state !== 'running'" link type="success" @click="doAction(row, 'start')">启动</el-button>
            <el-button v-if="row.state === 'running'" link type="warning" @click="doAction(row, 'stop')">停止</el-button>
            <el-button v-if="row.state === 'running'" link type="primary" @click="doAction(row, 'restart')">重启</el-button>
            <el-button link type="danger" @click="doRemove(row)">删除</el-button>
            <el-button link type="primary" @click="openStats(row)">统计</el-button>
            <el-button link type="warning" @click="openPortForwards(row)">端口</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-drawer>

    <el-drawer v-model="statsVisible" :title="`资源统计 · ${statsName}`" size="45%" append-to-body @closed="disposeHistChart">
      <el-descriptions :column="2" border v-if="statsData">
        <el-descriptions-item label="CPU 使用">{{ statsData.cpuPercent }}%</el-descriptions-item>
        <el-descriptions-item label="内存">{{ statsData.memUsedMb }} / {{ statsData.memLimitMb }} MB（{{ statsData.memPercent }}%）</el-descriptions-item>
        <el-descriptions-item label="网络接收">{{ statsData.netRxMb }} MB</el-descriptions-item>
        <el-descriptions-item label="网络发送">{{ statsData.netTxMb }} MB</el-descriptions-item>
        <el-descriptions-item label="块读">{{ statsData.blockReadMb }} MB</el-descriptions-item>
        <el-descriptions-item label="块写">{{ statsData.blockWriteMb }} MB</el-descriptions-item>
        <el-descriptions-item label="进程数">{{ statsData.pids }}</el-descriptions-item>
      </el-descriptions>
      <el-empty v-else description="加载中" />
      <el-divider content-position="left">历史趋势（5 分钟采样 · 7 天留存）</el-divider>
      <el-radio-group v-model="histHours" size="small" style="margin-bottom: 8px" @change="loadHistory">
        <el-radio-button :value="1">1h</el-radio-button>
        <el-radio-button :value="6">6h</el-radio-button>
        <el-radio-button :value="24">24h</el-radio-button>
        <el-radio-button :value="168">7d</el-radio-button>
      </el-radio-group>
      <div v-loading="histLoading" style="height: 240px">
        <div v-show="!histLoading" ref="histChartEl" style="height: 240px" />
        <el-empty v-if="!histLoading && histEmpty" description="暂无采样数据（采集循环每 5 分钟一轮）" :image-size="60" />
      </div>
    </el-drawer>

    <el-dialog v-model="pfVisible" :title="`端口转发规则 · ${pfName}`" width="640px" append-to-body>
      <el-alert
        type="warning"
        :closable="false"
        show-icon
        title="Docker 端口绑定不可在线修改：保存后将停止并按新规则重建容器（挂载卷数据保留，容器层内未落卷的数据将丢失），运行中的容器重建后自动拉起。"
        style="margin-bottom: 10px"
      />
      <el-table :data="pfRules" size="small" border>
        <el-table-column label="宿主IP" width="150">
          <template #default="{ row }">
            <el-input v-model="row.hostIp" placeholder="留空=0.0.0.0" size="small" />
          </template>
        </el-table-column>
        <el-table-column label="宿主端口" width="120">
          <template #default="{ row }">
            <el-input v-model="row.hostPort" placeholder="8080" size="small" />
          </template>
        </el-table-column>
        <el-table-column label="容器端口" width="120">
          <template #default="{ row }">
            <el-input-number v-model="row.containerPort" :min="1" :max="65535" controls-position="right" size="small" style="width: 100%" />
          </template>
        </el-table-column>
        <el-table-column label="协议" width="100">
          <template #default="{ row }">
            <el-select v-model="row.proto" size="small">
              <el-option label="tcp" value="tcp" />
              <el-option label="udp" value="udp" />
            </el-select>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="70">
          <template #default="{ $index }">
            <el-button link type="danger" @click="pfRules.splice($index, 1)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <div class="ops-btn-list" style="margin-top: 8px">
        <el-button size="small" icon="plus" @click="pfRules.push({ hostIp: '', hostPort: '', containerPort: 80, proto: 'tcp' })">添加规则</el-button>
      </div>
      <template #footer>
        <el-button @click="pfVisible = false">取 消</el-button>
        <el-button type="danger" :loading="pfSaving" @click="submitPortForwards">应用（重建容器）</el-button>
      </template>
    </el-dialog>

    <el-drawer v-model="resVisible" :title="`网络与卷 · ${ctEndpoint?.name || ''}`" size="65%">
      <el-tabs v-model="resTab">
        <el-tab-pane label="网络" name="networks">
          <el-form inline>
            <el-form-item>
              <el-input v-model="netForm.name" placeholder="网络名" style="width: 140px" />
              <el-input v-model="netForm.subnet" placeholder="子网（选配）如 172.30.0.0/16" style="width: 220px; margin-left: 6px" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="doCreateNetwork">创建</el-button>
              <el-button @click="loadNetworks">刷 新</el-button>
            </el-form-item>
          </el-form>
          <el-table :data="networks" v-loading="netLoading" stripe size="small">
            <el-table-column prop="name" label="名称" min-width="130" />
            <el-table-column prop="driver" label="驱动" width="90" />
            <el-table-column prop="subnet" label="子网" min-width="140" />
            <el-table-column label="操作" width="80">
              <template #default="{ row }">
                <el-button v-if="!row.builtIn" link type="danger" @click="doRemoveNetwork(row)">删除</el-button>
                <span v-else class="ops-text-muted">内置</span>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
        <el-tab-pane label="卷" name="volumes">
          <el-form inline>
            <el-form-item>
              <el-button :loading="volLoading" @click="loadVolumes">刷 新</el-button>
            </el-form-item>
          </el-form>
          <el-table :data="volumes" v-loading="volLoading" stripe size="small">
            <el-table-column prop="name" label="卷名" min-width="200" show-overflow-tooltip />
            <el-table-column prop="driver" label="驱动" width="100" />
            <el-table-column prop="mountpoint" label="挂载点" min-width="220" show-overflow-tooltip />
            <el-table-column label="操作" width="80">
              <template #default="{ row }">
                <el-button link type="danger" @click="doRemoveVolume(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
      </el-tabs>
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

    <el-drawer v-model="imgVisible" :title="`镜像 · ${ctEndpoint?.name || ''}`" size="65%">
      <el-form inline>
        <el-form-item>
          <el-input v-model="pullRef" placeholder="镜像引用，如 alpine:latest" style="width: 240px" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :disabled="!pullRef" @click="doPull">拉取</el-button>
          <el-button :loading="imgLoading" @click="loadImages">刷 新</el-button>
          <el-upload :show-file-list="false" accept=".tar" :http-request="onImportImage">
            <el-button type="success" plain>导入 tar</el-button>
          </el-upload>
        </el-form-item>
      </el-form>
      <el-table :data="imgList" v-loading="imgLoading" stripe size="small">
        <el-table-column label="标签" min-width="220">
          <template #default="{ row }">
            <el-tag v-for="t in row.tags.length ? row.tags : ['<none>']" :key="t" size="small" style="margin-right: 4px">{{ t }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="id" label="ID" width="130" />
        <el-table-column label="大小" width="100">
          <template #default="{ row }">{{ row.sizeMb }} MB</template>
        </el-table-column>
        <el-table-column label="创建时间" width="170">
          <template #default="{ row }">{{ new Date(row.createdAt * 1000).toLocaleString() }}</template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openTag(row)">打标</el-button>
            <el-button link @click="doExport(row)">导出</el-button>
            <el-button link type="danger" @click="doRemoveImage(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-drawer>

    <el-dialog v-model="tagVisible" title="镜像打标签" width="480px" append-to-body>
      <el-form label-width="80px">
        <el-form-item label="源引用">{{ tagSource }}</el-form-item>
        <el-form-item label="新标签">
          <el-input v-model="tagTarget" placeholder="如 registry.local/app:v2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="tagVisible = false">取 消</el-button>
        <el-button type="primary" @click="submitTag">确 定</el-button>
      </template>
    </el-dialog>

    <el-drawer v-model="eventVisible" :title="`容器事件 · ${eventEndpoint || '全部接入点'}`" size="60%">
      <el-table :data="eventList" v-loading="eventLoading" stripe size="small">
        <el-table-column label="时间" width="170">
          <template #default="{ row }">{{ (row.occurredAt || '').replace('T', ' ').slice(0, 19) }}</template>
        </el-table-column>
        <el-table-column prop="action" label="动作" width="110">
          <template #default="{ row }">
            <el-tag :type="{ start: 'success', die: 'danger', destroy: 'danger', stop: 'info' }[row.action] || 'warning'" size="small">
              {{ row.action }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="containerName" label="容器" min-width="150" show-overflow-tooltip />
        <el-table-column prop="containerId" label="ID" width="110" />
      </el-table>
    </el-drawer>
  </div>
</template>

<script setup>
  import { ref, reactive, onMounted, onUnmounted, nextTick } from 'vue'
  import * as echarts from 'echarts'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import {
    createEndpoint, updateEndpoint, deleteEndpoint, getEndpointList, checkEndpoint
  } from '@/plugin/container/api/dockerEndpoint'
  import { getContainerList, containerAction, createContainer, getImageList, pullImage, pullStatus, removeImage,
    tagImage, exportImage, importImage,
    getNetworkList, createNetwork, removeNetwork, getVolumeList, removeVolume, getContainerStats,
    getStatsHistory, listPortForwards, setPortForwards } from '@/plugin/container/api/container'
  import { getEventList } from '@/plugin/container/api/dockerEvent'
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

  // ---------- 镜像抽屉 ----------
  const imgVisible = ref(false)
  const imgList = ref([])
  const imgLoading = ref(false)
  const pullRef = ref('')
  let pullTimer = null

  const openImages = (row) => {
    ctEndpoint.value = row
    imgVisible.value = true
    loadImages()
  }

  const loadImages = async () => {
    if (!ctEndpoint.value) return
    imgLoading.value = true
    try {
      const res = await getImageList({ endpointId: ctEndpoint.value.ID })
      if (res.code === 0) imgList.value = res.data || []
    } finally {
      imgLoading.value = false
    }
  }

  const doPull = async () => {
    const ref = pullRef.value.trim()
    if (!ref) return
    const res = await pullImage({ endpointId: ctEndpoint.value.ID, ref })
    if (res.code === 0) {
      ElMessage.success('已开始拉取，完成后自动刷新')
      if (pullTimer) clearInterval(pullTimer)
      pullTimer = setInterval(async () => {
        const st = await pullStatus({ endpointId: ctEndpoint.value.ID, ref })
        if (st.code === 0 && String(st.data || '').startsWith('成功')) {
          clearInterval(pullTimer)
          pullTimer = null
          ElMessage.success(`拉取完成：${ref}`)
          loadImages()
        }
      }, 2000)
    }
  }

  const doRemoveImage = (row) => {
    const ref = row.tags.length ? row.tags[0] : row.id
    ElMessageBox.confirm(`确定删除镜像 ${ref} 吗？（force，容器占用将解除）`, '危险操作', {
      confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning'
    }).then(async () => {
      const res = await removeImage({ endpointId: ctEndpoint.value.ID, ref })
      if (res.code === 0) {
        ElMessage.success('已删除')
        loadImages()
      }
    })
  }

  // 打标签
  const tagVisible = ref(false)
  const tagSource = ref('')
  const tagTarget = ref('')

  const openTag = (row) => {
    tagSource.value = row.tags.length ? row.tags[0] : row.id
    tagTarget.value = ''
    tagVisible.value = true
  }

  const submitTag = async () => {
    if (!tagTarget.value.trim()) {
      ElMessage.warning('请输入新标签')
      return
    }
    const res = await tagImage({ endpointId: ctEndpoint.value.ID, source: tagSource.value, target: tagTarget.value })
    if (res.code === 0) {
      ElMessage.success('已打标')
      tagVisible.value = false
      loadImages()
    }
  }

  // 导出（blob 下载）
  const doExport = async (row) => {
    const ref = row.tags.length ? row.tags[0] : row.id
    const res = await exportImage({ endpointId: ctEndpoint.value.ID, ref })
    const url = URL.createObjectURL(new Blob([res.data || res]))
    const a = document.createElement('a')
    a.href = url
    a.download = ref.replace(/[/:]/g, '_') + '.tar'
    a.click()
    URL.revokeObjectURL(url)
  }

  // 导入
  const onImportImage = async (opt) => {
    const fd = new FormData()
    fd.append('endpointId', ctEndpoint.value.ID)
    fd.append('file', opt.file)
    const res = await importImage(fd)
    if (res.code === 0) {
      ElMessage.success(res.msg)
      loadImages()
    }
  }

  // ---------- 容器统计 ----------
  const statsVisible = ref(false)
  const statsName = ref('')
  const statsData = ref(null)

  // 统计历史趋势（ECharts 双轴：CPU% / 内存MB）
  const histHours = ref(24)
  const histLoading = ref(false)
  const histEmpty = ref(false)
  const histChartEl = ref(null)
  let histChart = null
  let histTarget = { id: '' }

  const openStats = async (row) => {
    statsName.value = prettyName(row.names)
    statsData.value = null
    statsVisible.value = true
    histTarget = { id: row.id }
    const res = await getContainerStats({ endpointId: ctEndpoint.value.ID, id: row.id })
    if (res.code === 0) statsData.value = res.data
    loadHistory()
  }

  const loadHistory = async () => {
    histLoading.value = true
    histEmpty.value = false
    try {
      const res = await getStatsHistory({
        endpointId: ctEndpoint.value.ID, id: histTarget.id, hours: histHours.value
      })
      if (res.code !== 0) return
      const list = res.data || []
      if (!list.length) {
        histEmpty.value = true
        if (histChart) histChart.clear()
        return
      }
      const times = list.map((x) => x.createdAt?.replace('T', ' ').slice(5, 16) || '')
      const cpu = list.map((x) => x.cpuPercent)
      const mem = list.map((x) => x.memUsedMb)
      await nextTick()
      if (!histChartEl.value) return
      if (!histChart) histChart = echarts.init(histChartEl.value)
      histChart.setOption({
        tooltip: { trigger: 'axis' },
        legend: { data: ['CPU %', '内存 MB'], top: 0 },
        grid: { left: 50, right: 50, top: 30, bottom: 40 },
        xAxis: { type: 'category', data: times, axisLabel: { fontSize: 10 } },
        yAxis: [
          { type: 'value', name: 'CPU %', axisLabel: { fontSize: 10 } },
          { type: 'value', name: '内存 MB', axisLabel: { fontSize: 10 } }
        ],
        series: [
          { name: 'CPU %', type: 'line', data: cpu, showSymbol: false, smooth: true, itemStyle: { color: '#409eff' } },
          { name: '内存 MB', type: 'line', yAxisIndex: 1, data: mem, showSymbol: false, smooth: true, areaStyle: { opacity: 0.15 }, itemStyle: { color: '#67c23a' } }
        ]
      }, true)
      histChart.resize()
    } finally {
      histLoading.value = false
    }
  }

  const disposeHistChart = () => {
    if (histChart) {
      histChart.dispose()
      histChart = null
    }
  }

  // ---------- 端口转发规则 ----------
  const pfVisible = ref(false)
  const pfName = ref('')
  const pfTarget = ref({ id: '' })
  const pfRules = ref([])
  const pfSaving = ref(false)

  const openPortForwards = async (row) => {
    pfName.value = prettyName(row.names)
    pfTarget.value = { id: row.id }
    pfRules.value = []
    pfVisible.value = true
    const res = await listPortForwards({ endpointId: ctEndpoint.value.ID, id: row.id })
    if (res.code === 0) {
      pfRules.value = (res.data || []).map((r) => ({
        hostIp: r.hostIp || '',
        hostPort: r.hostPort || '',
        containerPort: r.containerPort || 80,
        proto: r.proto || 'tcp'
      }))
    } else {
      ElMessage.error(res.msg || '端口规则获取失败')
    }
  }

  const submitPortForwards = () => {
    const rules = pfRules.value
    if (!rules.length) {
      ElMessage.warning('至少保留一条规则')
      return
    }
    ElMessageBox.confirm(
      '确认应用新端口规则吗？容器将被停止并按新规则重建（挂载卷数据保留，容器层数据丢失），运行中的容器自动拉起。',
      '危险操作',
      { confirmButtonText: '应用', cancelButtonText: '取消', type: 'error' }
    ).then(async () => {
      pfSaving.value = true
      try {
        const res = await setPortForwards({
          endpointId: ctEndpoint.value.ID,
          containerId: pfTarget.value.id,
          rules
        })
        if (res.code === 0) {
          ElMessage.success(res.msg || '端口规则已应用')
          pfVisible.value = false
          loadContainers()
        }
      } finally {
        pfSaving.value = false
      }
    })
  }

  // ---------- 网络与卷 ----------
  const resVisible = ref(false)
  const resTab = ref('networks')
  const networks = ref([])
  const netLoading = ref(false)
  const netForm = reactive({ name: '', subnet: '' })
  const volumes = ref([])
  const volLoading = ref(false)

  const openResources = (row) => {
    ctEndpoint.value = row
    resVisible.value = true
    loadNetworks()
    loadVolumes()
  }

  const loadNetworks = async () => {
    netLoading.value = true
    try {
      const res = await getNetworkList({ endpointId: ctEndpoint.value.ID })
      if (res.code === 0) networks.value = res.data || []
    } finally {
      netLoading.value = false
    }
  }

  const doCreateNetwork = async () => {
    if (!netForm.name.trim()) {
      ElMessage.warning('请输入网络名')
      return
    }
    const res = await createNetwork({ endpointId: ctEndpoint.value.ID, name: netForm.name, subnet: netForm.subnet || undefined })
    if (res.code === 0) {
      ElMessage.success('网络已创建')
      netForm.name = ''
      netForm.subnet = ''
      loadNetworks()
    }
  }

  const doRemoveNetwork = (row) => {
    ElMessageBox.confirm(`确定删除网络「${row.name}」吗？`, '提示', {
      confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning'
    }).then(async () => {
      const res = await removeNetwork({ endpointId: ctEndpoint.value.ID, name: row.name })
      if (res.code === 0) {
        ElMessage.success('已删除')
        loadNetworks()
      }
    })
  }

  const loadVolumes = async () => {
    volLoading.value = true
    try {
      const res = await getVolumeList({ endpointId: ctEndpoint.value.ID })
      if (res.code === 0) volumes.value = res.data || []
    } finally {
      volLoading.value = false
    }
  }

  const doRemoveVolume = (row) => {
    ElMessageBox.confirm(`确定删除卷「${row.name}」吗？被容器占用的卷将被 daemon 拒绝。`, '提示', {
      confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning'
    }).then(async () => {
      const res = await removeVolume({ endpointId: ctEndpoint.value.ID, name: row.name })
      if (res.code === 0) {
        ElMessage.success('已删除')
        loadVolumes()
      }
    })
  }

  // ---------- 容器事件 ----------
  const eventVisible = ref(false)
  const eventEndpoint = ref('')
  const eventList = ref([])
  const eventLoading = ref(false)

  const openEvents = async (row) => {
    eventEndpoint.value = row.name
    eventVisible.value = true
    eventLoading.value = true
    try {
      const res = await getEventList({ endpointId: row.ID })
      if (res.code === 0) eventList.value = res.data || []
    } finally {
      eventLoading.value = false
    }
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
