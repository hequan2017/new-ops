<template>
  <div>
    <div class="ops-card">
      <el-tabs v-model="tab" @tab-change="loadAll">
        <!-- 节点 -->
        <el-tab-pane label="算力节点" name="nodes">
          <div class="ops-btn-list">
            <el-button type="primary" @click="openNodeDialog()">注册节点</el-button>
            <el-button @click="loadNodes">刷 新</el-button>
          </div>
          <el-table :data="nodes" v-loading="nodesLoading" stripe>
            <el-table-column prop="name" label="节点" min-width="120" />
            <el-table-column prop="gpuModel" label="显卡" min-width="110" />
            <el-table-column label="GPU（已用/总）" width="130">
              <template #default="{ row }">{{ row.gpuUsed }} / {{ row.gpuTotal }}</template>
            </el-table-column>
            <el-table-column label="CPU（已用/总）" width="130">
              <template #default="{ row }">{{ row.cpuUsed }} / {{ row.cpuTotal }} 核</template>
            </el-table-column>
            <el-table-column label="内存（已用/总）" width="140">
              <template #default="{ row }">{{ row.memUsed }} / {{ row.memTotalGb }} GB</template>
            </el-table-column>
            <el-table-column label="状态" width="80">
              <template #default="{ row }">
                <el-tag :type="row.status === '在线' ? 'success' : 'info'" size="small">{{ row.status }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="130" fixed="right">
              <template #default="{ row }">
                <el-button link type="primary" @click="openNodeDialog(row)">编辑</el-button>
                <el-button link type="danger" @click="onDeleteNode(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <!-- 规格 -->
        <el-tab-pane label="产品规格" name="specs">
          <div class="ops-btn-list">
            <el-button type="primary" @click="openSpecDialog()">新建规格</el-button>
            <el-button @click="loadSpecs">刷 新</el-button>
          </div>
          <el-table :data="specs" v-loading="specsLoading" stripe>
            <el-table-column prop="name" label="规格" min-width="140" />
            <el-table-column label="配置" min-width="180">
              <template #default="{ row }">{{ row.gpuCount }} GPU / {{ row.cpuCores }} 核 / {{ row.memGb }} GB</template>
            </el-table-column>
            <el-table-column label="单价" width="110">
              <template #default="{ row }">￥{{ row.pricePerHour }}/小时</template>
            </el-table-column>
            <el-table-column prop="description" label="描述" min-width="160" show-overflow-tooltip />
            <el-table-column label="操作" width="80" fixed="right">
              <template #default="{ row }">
                <el-button link type="danger" @click="onDeleteSpec(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <!-- 实例 -->
        <el-tab-pane label="GPU 实例" name="instances">
          <div class="ops-btn-list">
            <el-button type="primary" @click="openStartDialog()">开通实例</el-button>
            <el-button @click="loadInstances">刷 新</el-button>
          </div>
          <el-table :data="instances" v-loading="instLoading" stripe>
            <el-table-column prop="ID" label="ID" width="60" />
            <el-table-column prop="name" label="实例名" min-width="140" show-overflow-tooltip />
            <el-table-column prop="nodeName" label="节点" width="120" />
            <el-table-column prop="specName" label="规格" width="120" />
            <el-table-column label="分配" min-width="160">
              <template #default="{ row }">{{ row.gpuAllocated }} GPU / {{ row.cpuAllocated }} 核 / {{ row.memAllocGb }} GB</template>
            </el-table-column>
            <el-table-column label="状态" width="90">
              <template #default="{ row }">
                <el-tag :type="{ 运行中: 'success', 创建失败: 'danger' }[row.status] || 'info'" size="small">{{ row.status }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="creator" label="开通人" width="90" />
            <el-table-column label="操作" width="90" fixed="right">
              <template #default="{ row }">
                <el-button v-if="row.status === '运行中'" link type="danger" @click="onRelease(row)">销毁</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
      </el-tabs>
    </div>

    <el-dialog v-model="nodeVisible" :title="nodeForm.ID ? '编辑节点' : '注册节点'" width="520px">
      <el-form label-width="90px">
        <el-form-item label="名称">
          <el-input v-model="nodeForm.name" placeholder="如 gpu-node-01" />
        </el-form-item>
        <el-form-item label="显卡型号">
          <el-input v-model="nodeForm.gpuModel" placeholder="如 A100 / RTX 4090" />
        </el-form-item>
        <el-form-item label="GPU 卡数">
          <el-input-number v-model="nodeForm.gpuTotal" :min="1" :max="64" style="width: 160px" />
        </el-form-item>
        <el-form-item label="CPU 核数">
          <el-input-number v-model="nodeForm.cpuTotal" :min="0" style="width: 160px" />
        </el-form-item>
        <el-form-item label="内存 GB">
          <el-input-number v-model="nodeForm.memTotalGb" :min="0" style="width: 160px" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="nodeForm.status" style="width: 160px">
            <el-option label="在线" value="在线" />
            <el-option label="离线" value="离线" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="nodeVisible = false">取 消</el-button>
        <el-button type="primary" @click="submitNode">保 存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="specVisible" title="新建规格" width="480px">
      <el-form label-width="90px">
        <el-form-item label="规格名">
          <el-input v-model="specForm.name" placeholder="如 1GPU-4C-16G" />
        </el-form-item>
        <el-form-item label="GPU 卡数">
          <el-input-number v-model="specForm.gpuCount" :min="1" :max="16" style="width: 160px" />
        </el-form-item>
        <el-form-item label="CPU 核数">
          <el-input-number v-model="specForm.cpuCores" :min="1" style="width: 160px" />
        </el-form-item>
        <el-form-item label="内存 GB">
          <el-input-number v-model="specForm.memGb" :min="1" style="width: 160px" />
        </el-form-item>
        <el-form-item label="单价/时">
          <el-input-number v-model="specForm.pricePerHour" :min="0" :precision="2" style="width: 160px" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="specVisible = false">取 消</el-button>
        <el-button type="primary" @click="submitSpec">保 存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="startVisible" title="开通 GPU 实例" width="480px">
      <el-form label-width="80px">
        <el-form-item label="节点">
          <el-select v-model="startForm.nodeId" style="width: 100%">
            <el-option
              v-for="n in nodes.filter((x) => x.status === '在线')"
              :key="n.ID"
              :label="`${n.name}（余 ${n.gpuTotal - n.gpuUsed} 卡）`"
              :value="n.ID"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="规格">
          <el-select v-model="startForm.specId" style="width: 100%">
            <el-option v-for="s in specs" :key="s.ID" :label="`${s.name}（${s.gpuCount} GPU）`" :value="s.ID" />
          </el-select>
        </el-form-item>
        <el-form-item label="实例名">
          <el-input v-model="startForm.name" placeholder="留空自动生成" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="startVisible = false">取 消</el-button>
        <el-button type="primary" :loading="starting" @click="submitStart">开 通</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
  import { ref, reactive, onMounted } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import {
    createGpuNode, updateGpuNode, deleteGpuNode, getGpuNodeList,
    createGpuSpec, deleteGpuSpec, getGpuSpecList,
    startGpuInstance, releaseGpuInstance, getGpuInstanceList
  } from '@/plugin/gpu/api/gpu'

  defineOptions({ name: 'gpuInstance' })

  const tab = ref('nodes')
  const nodes = ref([])
  const nodesLoading = ref(false)
  const specs = ref([])
  const specsLoading = ref(false)
  const instances = ref([])
  const instLoading = ref(false)

  const loadAll = () => {
    loadNodes()
    loadSpecs()
    loadInstances()
  }

  const loadNodes = async () => {
    nodesLoading.value = true
    try {
      const res = await getGpuNodeList()
      if (res.code === 0) nodes.value = res.data || []
    } finally {
      nodesLoading.value = false
    }
  }

  const nodeVisible = ref(false)
  const emptyNode = () => ({
    ID: 0, name: '', gpuModel: '', gpuTotal: 8, cpuTotal: 64, memTotalGb: 256, status: '在线', notes: ''
  })
  const nodeForm = reactive(emptyNode())

  const openNodeDialog = (row) => {
    Object.assign(nodeForm, emptyNode())
    if (row) Object.assign(nodeForm, row)
    nodeVisible.value = true
  }

  const submitNode = async () => {
    const res = nodeForm.ID ? await updateGpuNode(nodeForm) : await createGpuNode(nodeForm)
    if (res.code === 0) {
      ElMessage.success(res.msg)
      nodeVisible.value = false
      loadNodes()
    }
  }

  const onDeleteNode = (row) => {
    ElMessageBox.confirm(`确定删除节点「${row.name}」吗？`, '提示', {
      confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning'
    }).then(async () => {
      const res = await deleteGpuNode({ id: row.ID })
      if (res.code === 0) {
        ElMessage.success('已删除')
        loadNodes()
      }
    })
  }

  const loadSpecs = async () => {
    specsLoading.value = true
    try {
      const res = await getGpuSpecList()
      if (res.code === 0) specs.value = res.data || []
    } finally {
      specsLoading.value = false
    }
  }

  const specVisible = ref(false)
  const specForm = reactive({ name: '', gpuCount: 1, cpuCores: 8, memGb: 32, pricePerHour: 0 })

  const openSpecDialog = () => {
    Object.assign(specForm, { name: '', gpuCount: 1, cpuCores: 8, memGb: 32, pricePerHour: 0 })
    specVisible.value = true
  }

  const submitSpec = async () => {
    const res = await createGpuSpec(specForm)
    if (res.code === 0) {
      ElMessage.success('规格已创建')
      specVisible.value = false
      loadSpecs()
    }
  }

  const onDeleteSpec = (row) => {
    ElMessageBox.confirm(`确定删除规格「${row.name}」吗？`, '提示', {
      confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning'
    }).then(async () => {
      const res = await deleteGpuSpec({ id: row.ID })
      if (res.code === 0) {
        ElMessage.success('已删除')
        loadSpecs()
      }
    })
  }

  const loadInstances = async () => {
    instLoading.value = true
    try {
      const res = await getGpuInstanceList({})
      if (res.code === 0) instances.value = res.data || []
    } finally {
      instLoading.value = false
    }
  }

  const startVisible = ref(false)
  const starting = ref(false)
  const startForm = reactive({ nodeId: null, specId: null, name: '' })

  const openStartDialog = () => {
    startForm.nodeId = null
    startForm.specId = null
    startForm.name = ''
    startVisible.value = true
  }

  const submitStart = async () => {
    if (!startForm.nodeId || !startForm.specId) {
      ElMessage.warning('请选择节点与规格')
      return
    }
    starting.value = true
    try {
      const res = await startGpuInstance({ ...startForm })
      if (res.code === 0) {
        ElMessage.success(`实例 #${res.data.ID} 已开通`)
        startVisible.value = false
        loadInstances()
        loadNodes()
      }
    } finally {
      starting.value = false
    }
  }

  const onRelease = (row) => {
    ElMessageBox.confirm(`确定销毁实例「${row.name}」吗？配额将立即释放。`, '危险操作', {
      confirmButtonText: '销毁', cancelButtonText: '取消', type: 'warning'
    }).then(async () => {
      const res = await releaseGpuInstance({ id: row.ID })
      if (res.code === 0) {
        ElMessage.success('已销毁')
        loadInstances()
        loadNodes()
      }
    })
  }

  onMounted(loadAll)
</script>
