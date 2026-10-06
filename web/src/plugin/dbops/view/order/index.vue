<template>
  <div>
    <div class="ops-card">
      <el-form inline>
        <el-form-item label="发起人">
          <el-input v-model="creator" placeholder="操作人" clearable style="width: 140px" @keyup.enter="getList" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="status" clearable style="width: 120px" @change="getList">
            <el-option v-for="s in ['待审核', '审核通过', '执行中', '成功', '失败', '已取消']" :key="s" :label="s" :value="s" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="getList">查 询</el-button>
          <el-button type="success" @click="openCreate">新建工单</el-button>
          <el-button @click="getList">刷 新</el-button>
        </el-form-item>
      </el-form>

      <el-table :data="tableData" v-loading="loading" stripe>
        <el-table-column prop="ID" label="ID" width="60" />
        <el-table-column prop="title" label="标题" min-width="160" show-overflow-tooltip />
        <el-table-column prop="instanceName" label="实例" width="130" />
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)" size="small">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="creator" label="发起人" width="90" />
        <el-table-column label="发起时间" width="160">
          <template #default="{ row }">{{ (row.createdAt || '').replace('T', ' ').slice(0, 19) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openDetail(row)">明细</el-button>
            <el-button v-if="row.status === '待审核'" link type="warning" @click="doAudit(row)">审核</el-button>
            <el-button v-if="row.status === '待审核'" link type="danger" @click="doCancel(row)">取消</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="ops-pagination">
        <el-pagination
          :current-page="page"
          :page-size="pageSize"
          :total="total"
          layout="total, prev, pager, next"
          @current-change="loadList"
        />
      </div>
    </div>

    <el-dialog v-model="createVisible" title="新建 SQL 工单" width="640px">
      <el-form label-width="80px">
        <el-form-item label="实例">
          <el-select v-model="createForm.instanceId" style="width: 100%">
            <el-option v-for="i in instances" :key="i.ID" :label="`${i.name}（${i.host}:${i.port}）`" :value="i.ID" />
          </el-select>
        </el-form-item>
        <el-form-item label="标题">
          <el-input v-model="createForm.title" placeholder="如 user 表加索引" />
        </el-form-item>
        <el-form-item label="SQL">
          <el-input v-model="createForm.sqlText" type="textarea" :rows="8"
            placeholder="-- 支持 DDL/DML，每条以分号结尾&#10;ALTER TABLE users ADD INDEX idx_name(name);" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createVisible = false">取 消</el-button>
        <el-button type="primary" @click="submitCreate">提 交</el-button>
      </template>
    </el-dialog>

    <el-drawer v-model="detailVisible" :title="`工单明细 · ${detail?.title || ''}`" size="55%">
      <el-descriptions :column="2" border size="small" style="margin-bottom: 12px">
        <el-descriptions-item label="实例">{{ detail?.instanceName }}</el-descriptions-item>
        <el-descriptions-item label="状态">{{ detail?.status }}</el-descriptions-item>
        <el-descriptions-item label="发起人">{{ detail?.creator }}</el-descriptions-item>
        <el-descriptions-item label="结束时间">{{ (detail?.finishedAt || '—').toString().replace('T', ' ').slice(0, 19) }}</el-descriptions-item>
      </el-descriptions>
      <pre class="ops-pre">{{ detail?.sqlText }}</pre>
      <template v-if="detail?.auditResult || detail?.execResult">
        <h4 style="margin: 12px 0 6px">审核/执行结果</h4>
        <pre class="ops-pre">{{ detail?.auditResult }}{{ detail?.execResult }}</pre>
      </template>
    </el-drawer>
  </div>
</template>

<script setup>
  import { ref, reactive, onMounted } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import {
    createOrder, auditOrder, cancelOrder, getOrderList, getInstanceList
  } from '@/plugin/dbops/api/dbops'

  defineOptions({ name: 'dbopsOrder' })

  const creator = ref('')
  const status = ref('')
  const tableData = ref([])
  const loading = ref(false)
  const page = ref(1)
  const pageSize = 15
  const total = ref(0)
  const instances = ref([])

  const statusType = (s) =>
    ({ 待审核: 'warning', 审核通过: 'primary', 执行中: 'primary', 成功: 'success', 失败: 'danger', 已取消: 'info' }[s] || 'info')

  const getList = async (p) => {
    if (p) page.value = p
    loading.value = true
    try {
      const res = await getOrderList({
        page: page.value, pageSize,
        creator: creator.value || undefined
      })
      if (res.code === 0) {
        tableData.value = res.data.list || []
        total.value = res.data.total || 0
      }
    } finally {
      loading.value = false
    }
  }

  const createVisible = ref(false)
  const createForm = reactive({ instanceId: null, title: '', sqlText: '' })

  const openCreate = async () => {
    const res = await getInstanceList({})
    if (res.code === 0) instances.value = res.data || []
    createForm.instanceId = instances.value.length ? instances.value[0].ID : null
    createForm.title = ''
    createForm.sqlText = ''
    createVisible.value = true
  }

  const submitCreate = async () => {
    if (!createForm.instanceId || !createForm.title.trim() || !createForm.sqlText.trim()) {
      ElMessage.warning('请选择实例并填写标题与 SQL')
      return
    }
    const res = await createOrder({ ...createForm })
    if (res.code === 0) {
      ElMessage.success(`工单 #${res.data.ID} 已创建（待审核）`)
      createVisible.value = false
      getList()
    }
  }

  const doAudit = (row) => {
    ElMessageBox.confirm(
      '将调用 goInception 审核；审核引擎未配置时会明确提示。继续吗？', 'SQL 审核',
      { confirmButtonText: '审核', cancelButtonText: '取消', type: 'info' }
    ).then(async () => {
      const res = await auditOrder({ id: row.ID })
      if (res.code === 0) getList()
    })
  }

  const doCancel = (row) => {
    ElMessageBox.confirm(`确定取消工单「${row.title}」吗？`, '提示', {
      confirmButtonText: '取消工单', cancelButtonText: '返回', type: 'warning'
    }).then(async () => {
      const res = await cancelOrder({ id: row.ID })
      if (res.code === 0) {
        ElMessage.success('已取消')
        getList()
      }
    })
  }

  const detailVisible = ref(false)
  const detail = ref(null)
  const openDetail = (row) => {
    detail.value = row
    detailVisible.value = true
  }

  onMounted(getList)
</script>

<style scoped>
  .ops-pre {
    margin: 0;
    max-height: 320px;
    overflow: auto;
    white-space: pre-wrap;
    word-break: break-all;
    font-size: 12px;
    background: var(--el-fill-color-light);
    padding: 8px;
    border-radius: 4px;
  }
</style>
