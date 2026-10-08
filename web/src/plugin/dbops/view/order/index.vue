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
          <el-button @click="openIncConfig">审核引擎</el-button>
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
        <el-table-column label="操作" width="260" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openDetail(row)">明细</el-button>
            <el-button v-if="row.status === '待审核'" link type="warning" @click="doAudit(row)">审核</el-button>
            <el-button v-if="row.status === '审核通过'" link type="success" @click="doExecute(row)">执行</el-button>
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
      <template v-if="auditRows(detail?.auditResult).length">
        <h4 style="margin: 12px 0 6px">审核结果</h4>
        <el-table :data="auditRows(detail?.auditResult)" size="small" border>
          <el-table-column prop="stage" label="阶段" width="90" />
          <el-table-column label="级别" width="70">
            <template #default="{ row }">
              <el-tag :type="row.errlevel >= 2 ? 'danger' : row.errlevel === 1 ? 'warning' : 'success'" size="small">
                {{ row.errlevel >= 2 ? '错误' : row.errlevel === 1 ? '警告' : '通过' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="sqlstatement" label="SQL" min-width="200" show-overflow-tooltip />
          <el-table-column prop="errormessage" label="消息" min-width="180" show-overflow-tooltip />
        </el-table>
      </template>
      <template v-if="execPayload(detail?.execResult)">
        <h4 style="margin: 12px 0 6px">执行结果</h4>
        <el-table :data="execPayload(detail?.execResult).rows || []" size="small" border>
          <el-table-column prop="stage" label="阶段" width="90" />
          <el-table-column label="级别" width="70">
            <template #default="{ row }">
              <el-tag :type="row.errlevel >= 2 ? 'danger' : row.errlevel === 1 ? 'warning' : 'success'" size="small">
                {{ row.errlevel >= 2 ? '错误' : row.errlevel === 1 ? '警告' : '成功' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="sqlstatement" label="SQL" min-width="200" show-overflow-tooltip />
          <el-table-column prop="errormessage" label="消息" min-width="160" show-overflow-tooltip />
        </el-table>
        <template v-if="(execPayload(detail?.execResult).rollbacks || []).length">
          <h4 style="margin: 12px 0 6px">备份回滚语句</h4>
          <pre class="ops-pre">{{ (execPayload(detail?.execResult).rollbacks || []).map((r) => r.rollback).join('\n') }}</pre>
        </template>
      </template>
    </el-drawer>

    <el-dialog v-model="incVisible" title="审核引擎（goInception）配置" width="560px">
      <el-form label-width="110px">
        <el-form-item label="引擎地址">
          <el-input v-model="incForm.host" placeholder="goInception 主机" style="width: 220px" />
          <el-input-number v-model="incForm.port" :min="1" :max="65535" controls-position="right"
            style="width: 130px; margin-left: 8px" />
        </el-form-item>
        <el-form-item label="备份库地址">
          <el-input v-model="incForm.backupHost" placeholder="备份 MySQL 主机" style="width: 220px" />
          <el-input-number v-model="incForm.backupPort" :min="1" :max="65535" controls-position="right"
            style="width: 130px; margin-left: 8px" />
        </el-form-item>
        <el-form-item label="备份库用户">
          <el-input v-model="incForm.backupUser" placeholder="root" style="width: 220px" />
        </el-form-item>
        <el-form-item label="备份库密码">
          <el-input v-model="incForm.backupPass" type="password" show-password
            placeholder="留空表示不修改" style="width: 220px" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="incForm.notes" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="doTestInc">测试连通</el-button>
        <el-button @click="incVisible = false">取 消</el-button>
        <el-button type="primary" @click="submitIncConfig">保 存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
  import { ref, reactive, onMounted } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import {
    createOrder, auditOrder, executeOrder, cancelOrder, getOrderList, getInstanceList,
    getInceptionConfig, saveInceptionConfig, testInception
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
      '将调用 goInception 审核；有错误级别（errlevel≥2）的工单会保持待审核。继续吗？', 'SQL 审核',
      { confirmButtonText: '审核', cancelButtonText: '取消', type: 'info' }
    ).then(async () => {
      const res = await auditOrder({ id: row.ID })
      if (res.code === 0) {
        ElMessage.success('审核完成')
        getList()
      }
    })
  }

  const doExecute = (row) => {
    ElMessageBox.confirm(
      `将在实例「${row.instanceName}」上执行工单「${row.title}」的 SQL，执行自动生成备份回滚语句。确定执行吗？`,
      'SQL 上线执行', { confirmButtonText: '执行', cancelButtonText: '取消', type: 'warning' }
    ).then(async () => {
      const res = await executeOrder({ id: row.ID })
      if (res.code === 0) {
        ElMessage.success(`执行完成：${res.data.status}`)
        getList()
      }
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

  // 审核结果 JSON 解析（容错：非 JSON 原样空）
  const parseJson = (s) => {
    try { return JSON.parse(s) } catch (e) { return null }
  }
  const auditRows = (s) => parseJson(s) || []
  const execPayload = (s) => parseJson(s) || null

  const detailVisible = ref(false)
  const detail = ref(null)
  const openDetail = (row) => {
    detail.value = row
    detailVisible.value = true
  }

  // 审核引擎配置（密码留空不改）
  const incVisible = ref(false)
  const incForm = reactive({ host: '', port: 4000, backupHost: '', backupPort: 3306, backupUser: '', backupPass: '', notes: '' })

  const openIncConfig = async () => {
    const res = await getInceptionConfig()
    if (res.code === 0) {
      Object.assign(incForm, {
        host: res.data.host || '', port: res.data.port || 4000,
        backupHost: res.data.backupHost || '', backupPort: res.data.backupPort || 3306,
        backupUser: res.data.backupUser || '', backupPass: '', notes: res.data.notes || ''
      })
    }
    incVisible.value = true
  }

  const doTestInc = async () => {
    const res = await testInception()
    if (res.code === 0) ElMessage.success(res.msg || '引擎可达')
  }

  const submitIncConfig = async () => {
    const res = await saveInceptionConfig({ ...incForm })
    if (res.code === 0) {
      ElMessage.success('已保存')
      incVisible.value = false
    }
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
