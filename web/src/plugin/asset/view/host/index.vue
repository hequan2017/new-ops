<template>
  <div>
    <div class="ops-search-box">
      <el-form
        ref="elSearchFormRef"
        :inline="true"
        :model="searchInfo"
        @keyup.enter="onSubmit"
      >
        <el-form-item label="关键字">
          <el-input
            v-model="searchInfo.keyword"
            placeholder="主机名 / IP / SN / 负责人"
            clearable
          />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchInfo.status" placeholder="全部" clearable style="width: 140px">
            <el-option
              v-for="s in statusOptions"
              :key="s"
              :label="s"
              :value="s"
            />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" icon="search" @click="onSubmit">查询</el-button>
          <el-button icon="refresh" @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>
    </div>

    <div class="ops-table-box">
      <div class="ops-btn-list">
        <el-button type="primary" icon="plus" @click="openDialog">新增主机</el-button>
        <el-upload
          :show-file-list="false"
          accept=".xlsx"
          :http-request="onImport"
        >
          <el-button type="success" icon="upload">导入 Excel</el-button>
        </el-upload>
        <el-button icon="download" @click="onExport">导出 Excel</el-button>
        <el-button
          type="danger"
          icon="delete"
          :disabled="!multipleSelection.length"
          @click="onDeleteBatch"
        >批量删除</el-button>
      </div>

      <el-table
        :data="tableData"
        style="width: 100%"
        tooltip-effect="dark"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="55" />
        <el-table-column prop="hostname" label="主机名" min-width="140" />
        <el-table-column prop="ip" label="内网IP" min-width="130" />
        <el-table-column label="配置" min-width="150">
          <template #default="{ row }">
            {{ row.cpuCores || '-' }}C / {{ row.memGb || '-' }}G / {{ row.diskGb || '-' }}G
          </template>
        </el-table-column>
        <el-table-column prop="os" label="操作系统" min-width="120">
          <template #default="{ row }">
            {{ row.os ? `${row.os} ${row.osVersion || ''}` : '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="vendor" label="厂商" min-width="100">
          <template #default="{ row }">{{ row.vendor || '-' }}</template>
        </el-table-column>
        <el-table-column prop="room.name" label="机房" min-width="100">
          <template #default="{ row }">{{ row.room?.name || '-' }}</template>
        </el-table-column>
        <el-table-column prop="owner" label="负责人" min-width="90">
          <template #default="{ row }">{{ row.owner || '-' }}</template>
        </el-table-column>
        <el-table-column label="最近采集" width="150">
          <template #default="{ row }">
            {{ row.lastCollectAt ? row.lastCollectAt.replace('T', ' ').slice(0, 19) : '未采集' }}
          </template>
        </el-table-column>
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)">{{ row.status || '-' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" icon="edit" @click="openDialog(row)">编辑</el-button>
            <el-button link type="success" icon="aim" @click="openCollect(row)">采集</el-button>
            <el-button link type="warning" icon="clock" @click="openHistory(row)">历史</el-button>
            <el-button link type="danger" icon="delete" @click="onDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="ops-pagination">
        <el-pagination
          :current-page="page"
          :page-size="pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @current-change="handleCurrentChange"
          @size-change="handleSizeChange"
        />
      </div>
    </div>

    <el-dialog
      v-model="dialogFormVisible"
      :title="form.ID ? '编辑主机' : '新增主机'"
      width="640px"
    >
      <el-form ref="formRef" :model="form" :rules="rules" label-width="90px">
        <el-row :gutter="12">
          <el-col :span="12">
            <el-form-item label="主机名" prop="hostname">
              <el-input v-model="form.hostname" placeholder="如 web01" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="内网IP" prop="ip">
              <el-input v-model="form.ip" placeholder="如 192.168.1.10" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="操作系统">
              <el-input v-model="form.os" placeholder="如 CentOS / Ubuntu" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="系统版本">
              <el-input v-model="form.osVersion" placeholder="如 7.9 / 22.04" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="CPU核数">
              <el-input-number v-model="form.cpuCores" :min="0" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="内存GB">
              <el-input-number v-model="form.memGb" :min="0" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="磁盘GB">
              <el-input-number v-model="form.diskGb" :min="0" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="SN">
              <el-input v-model="form.sn" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="厂商">
              <el-input v-model="form.vendor" placeholder="如 Dell / 阿里云" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="负责人">
              <el-input v-model="form.owner" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="状态">
              <el-select v-model="form.status" style="width: 100%">
                <el-option v-for="s in statusOptions" :key="s" :label="s" :value="s" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="备注">
              <el-input v-model="form.notes" type="textarea" :rows="2" />
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
      <template #footer>
        <el-button @click="closeDialog">取 消</el-button>
        <el-button type="primary" @click="submitForm">确 定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="collectVisible" title="SSH 现场采集" width="480px">
      <el-form label-width="90px">
        <el-form-item label="目标主机">
          <el-input :model-value="collectTarget" disabled />
        </el-form-item>
        <el-form-item label="SSH 凭据">
          <el-select v-model="collectCredId" placeholder="选择凭据保险库中的 SSH 凭据" style="width: 100%">
            <el-option
              v-for="c in credList"
              :key="c.ID"
              :label="`${c.name}（${typeLabel(c.type)}${c.username ? ' / ' + c.username : ''}）`"
              :value="c.ID"
            />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="collectVisible = false">取 消</el-button>
        <el-button type="primary" :loading="collecting" @click="submitCollect">开始采集</el-button>
      </template>
    </el-dialog>

    <el-drawer v-model="historyVisible" :title="`变更历史 · ${historyHost}`" size="480px">
      <el-timeline>
        <el-timeline-item
          v-for="h in historyList"
          :key="h.ID"
          :timestamp="`${h.CreatedAt} · ${h.operator || '-'}`"
          :type="h.action === '删除' ? 'danger' : h.action === '更新' ? 'warning' : 'success'"
        >
          <b>{{ h.action }}</b>
          <details style="margin-top: 4px">
            <summary style="cursor: pointer; font-size: 12px; color: var(--el-color-info)">快照</summary>
            <pre style="white-space: pre-wrap; font-size: 12px">{{ pretty(h.snapshot) }}</pre>
          </details>
        </el-timeline-item>
      </el-timeline>
      <el-empty v-if="!historyList.length" description="暂无历史" />
    </el-drawer>
  </div>
</template>

<script setup>
  import {
    createAssetHost,
    deleteAssetHost,
    deleteAssetHostByIds,
    updateAssetHost,
    findAssetHost,
    getAssetHostList,
    getAssetHostHistory,
    importAssetHost,
    collectAssetHost
  } from '@/plugin/asset/api/assetHost'
  import { getCredentialList } from '@/plugin/asset/api/credential'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { reactive, ref } from 'vue'
  import { useUserStore } from '@/pinia/modules/user'

  defineOptions({ name: 'AssetHost' })

  const statusOptions = ['运行中', '停机', '维护中', '已报废']

  const statusTagType = (status) => {
    switch (status) {
      case '运行中':
        return 'success'
      case '停机':
        return 'info'
      case '维护中':
        return 'warning'
      case '已报废':
        return 'danger'
      default:
        return 'info'
    }
  }

  // 列表查询
  const searchInfo = reactive({ keyword: '', status: '' })
  const tableData = ref([])
  const page = ref(1)
  const pageSize = ref(10)
  const total = ref(0)

  const getTableData = async () => {
    const res = await getAssetHostList({
      page: page.value,
      pageSize: pageSize.value,
      keyword: searchInfo.keyword || undefined,
      status: searchInfo.status || undefined
    })
    if (res.code === 0) {
      tableData.value = res.data.list || []
      total.value = res.data.total || 0
    }
  }

  const onSubmit = () => {
    page.value = 1
    getTableData()
  }

  const onReset = () => {
    searchInfo.keyword = ''
    searchInfo.status = ''
    onSubmit()
  }

  const handleCurrentChange = (val) => {
    page.value = val
    getTableData()
  }

  const handleSizeChange = (val) => {
    pageSize.value = val
    onSubmit()
  }

  // 删除
  const onDelete = (row) => {
    ElMessageBox.confirm('确定删除该主机资产吗？', '提示', {
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      type: 'warning'
    }).then(async () => {
      const res = await deleteAssetHost({ id: row.ID })
      if (res.code === 0) {
        ElMessage.success('删除成功')
        getTableData()
      }
    })
  }

  // 批量删除
  const multipleSelection = ref([])
  const handleSelectionChange = (val) => {
    multipleSelection.value = val
  }
  const onDeleteBatch = () => {
    ElMessageBox.confirm(
      `确定删除选中的 ${multipleSelection.value.length} 台主机资产吗？`,
      '提示',
      { confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning' }
    ).then(async () => {
      const ids = multipleSelection.value.map((i) => i.ID)
      const res = await deleteAssetHostByIds({ ids })
      if (res.code === 0) {
        ElMessage.success('批量删除成功')
        getTableData()
      }
    })
  }

  // 弹窗
  const dialogFormVisible = ref(false)
  const formRef = ref(null)
  const emptyForm = () => ({
    ID: 0,
    hostname: '',
    ip: '',
    os: '',
    osVersion: '',
    cpuCores: 0,
    memGb: 0,
    diskGb: 0,
    sn: '',
    vendor: '',
    owner: '',
    status: '运行中',
    notes: ''
  })
  const form = reactive(emptyForm())

  const rules = {
    hostname: [{ required: true, message: '请输入主机名', trigger: 'blur' }],
    ip: [
      { required: true, message: '请输入内网IP', trigger: 'blur' },
      {
        pattern: /^(\d{1,3}\.){3}\d{1,3}$/,
        message: 'IP 格式不正确',
        trigger: 'blur'
      }
    ]
  }

  const openDialog = async (row) => {
    Object.assign(form, emptyForm())
    if (row && row.ID) {
      const res = await findAssetHost({ id: row.ID })
      if (res.code === 0) {
        Object.assign(form, res.data)
      }
    }
    dialogFormVisible.value = true
  }

  const closeDialog = () => {
    dialogFormVisible.value = false
  }

  const submitForm = () => {
    formRef.value.validate(async (valid) => {
      if (!valid) return
      let res
      if (form.ID) {
        res = await updateAssetHost(form)
      } else {
        res = await createAssetHost(form)
      }
      if (res.code === 0) {
        ElMessage.success(form.ID ? '更新成功' : '创建成功')
        closeDialog()
        getTableData()
      }
    })
  }

  // SSH 现场采集
  const collectVisible = ref(false)
  const collecting = ref(false)
  const collectTarget = ref('')
  const collectHostId = ref(0)
  const collectCredId = ref(undefined)
  const credList = ref([])
  const typeLabel = (v) => ({
    ssh_password: 'SSH 密码', ssh_key: 'SSH 私钥', cloud_ak: '云平台 AccessKey',
    docker_tls: 'Docker TLS', kubeconfig: 'kubeconfig'
  }[v] || v)

  const openCollect = async (row) => {
    collectHostId.value = row.ID
    collectTarget.value = `${row.hostname}（${row.ip}）`
    const res = await getCredentialList()
    if (res.code === 0) {
      credList.value = (res.data || []).filter(
        (c) => c.type === 'ssh_password' || c.type === 'ssh_key'
      )
    }
    collectVisible.value = true
  }

  const submitCollect = async () => {
    if (!collectCredId.value) {
      ElMessage.warning('请选择 SSH 凭据')
      return
    }
    collecting.value = true
    try {
      const res = await collectAssetHost({
        ID: collectHostId.value,
        credentialId: collectCredId.value
      })
      if (res.code === 0) {
        ElMessage.success(
          `采集成功：${res.data.os || '?'} ${res.data.osVersion || ''} / ${res.data.cpuCores || '?'}C / ${res.data.memGb || '?'}G / ${res.data.diskGb || '?'}G`
        )
        collectVisible.value = false
        getTableData()
      }
    } finally {
      collecting.value = false
    }
  }

  // 变更历史抽屉
  const historyVisible = ref(false)
  const historyList = ref([])
  const historyHost = ref('')

  const pretty = (snapshot) => {
    try {
      return JSON.stringify(JSON.parse(snapshot), null, 2)
    } catch (e) {
      return snapshot
    }
  }

  const openHistory = async (row) => {
    historyHost.value = `${row.hostname}（${row.ip}）`
    const res = await getAssetHostHistory({ id: row.ID, page: 1, pageSize: 50 })
    if (res.code === 0) {
      historyList.value = res.data.list || []
      historyVisible.value = true
    }
  }

  // Excel 导入
  const onImport = async (opt) => {
    const res = await importAssetHost(opt.file)
    if (res.code === 0) {
      const d = res.data
      ElMessage.success(`导入完成：新建 ${d.created}，更新 ${d.updated}，失败 ${(d.failed || []).length}`)
      getTableData()
    }
  }

  // Excel 导出（fetch 直取，绕过 JSON 拦截器，手动带 token）
  const onExport = async () => {
    const userStore = useUserStore()
    const res = await fetch(import.meta.env.VITE_BASE_API + '/asset/host/export', {
      headers: { 'x-token': userStore.token }
    })
    if (!res.ok) {
      ElMessage.error('导出失败')
      return
    }
    const blob = await res.blob()
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = 'asset_hosts.xlsx'
    a.click()
    URL.revokeObjectURL(url)
  }

  getTableData()
</script>
