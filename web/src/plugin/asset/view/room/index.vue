<template>
  <div>
    <el-card shadow="never">
      <el-tabs v-model="activeTab">
        <!-- 机房 -->
        <el-tab-pane label="机房" name="room">
          <div class="ops-btn-list">
            <el-button type="primary" icon="plus" @click="openRoomDialog()">新增机房</el-button>
          </div>
          <el-table :data="roomList" style="width: 100%">
            <el-table-column prop="name" label="机房名称" min-width="140" />
            <el-table-column prop="region" label="区域" min-width="100">
              <template #default="{ row }">{{ row.region || '-' }}</template>
            </el-table-column>
            <el-table-column prop="address" label="地址" min-width="160">
              <template #default="{ row }">{{ row.address || '-' }}</template>
            </el-table-column>
            <el-table-column prop="notes" label="备注" min-width="140">
              <template #default="{ row }">{{ row.notes || '-' }}</template>
            </el-table-column>
            <el-table-column label="操作" width="150" fixed="right">
              <template #default="{ row }">
                <el-button link type="primary" icon="edit" @click="openRoomDialog(row)">编辑</el-button>
                <el-button link type="danger" icon="delete" @click="deleteRoom(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <!-- 机柜 -->
        <el-tab-pane label="机柜" name="rack">
          <div class="ops-btn-list">
            <el-button type="primary" icon="plus" @click="openRackDialog()">新增机柜</el-button>
            <el-select
              v-model="rackRoomFilter"
              placeholder="按机房过滤"
              clearable
              style="width: 200px"
              @change="getRacks"
            >
              <el-option v-for="r in roomList" :key="r.ID" :label="r.name" :value="r.ID" />
            </el-select>
          </div>
          <el-table :data="rackList" style="width: 100%">
            <el-table-column prop="name" label="机柜名称" min-width="140" />
            <el-table-column label="所属机房" min-width="140">
              <template #default="{ row }">{{ row.room?.name || '-' }}</template>
            </el-table-column>
            <el-table-column prop="totalU" label="总U数" width="100">
              <template #default="{ row }">{{ row.totalU || '-' }}</template>
            </el-table-column>
            <el-table-column prop="notes" label="备注" min-width="140">
              <template #default="{ row }">{{ row.notes || '-' }}</template>
            </el-table-column>
            <el-table-column label="操作" width="150" fixed="right">
              <template #default="{ row }">
                <el-button link type="primary" icon="edit" @click="openRackDialog(row)">编辑</el-button>
                <el-button link type="danger" icon="delete" @click="deleteRack(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <!-- 机房弹窗 -->
    <el-dialog v-model="roomDialogVisible" :title="roomForm.ID ? '编辑机房' : '新增机房'" width="480px">
      <el-form ref="roomFormRef" :model="roomForm" :rules="roomRules" label-width="80px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="roomForm.name" placeholder="如 北京一机房" />
        </el-form-item>
        <el-form-item label="区域">
          <el-input v-model="roomForm.region" placeholder="如 华北" />
        </el-form-item>
        <el-form-item label="地址">
          <el-input v-model="roomForm.address" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="roomForm.notes" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="roomDialogVisible = false">取 消</el-button>
        <el-button type="primary" @click="submitRoom">确 定</el-button>
      </template>
    </el-dialog>

    <!-- 机柜弹窗 -->
    <el-dialog v-model="rackDialogVisible" :title="rackForm.ID ? '编辑机柜' : '新增机柜'" width="480px">
      <el-form ref="rackFormRef" :model="rackForm" :rules="rackRules" label-width="80px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="rackForm.name" placeholder="如 A-01" />
        </el-form-item>
        <el-form-item label="所属机房" prop="roomId">
          <el-select v-model="rackForm.roomId" style="width: 100%">
            <el-option v-for="r in roomList" :key="r.ID" :label="r.name" :value="r.ID" />
          </el-select>
        </el-form-item>
        <el-form-item label="总U数">
          <el-input-number v-model="rackForm.totalU" :min="0" style="width: 100%" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="rackForm.notes" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="rackDialogVisible = false">取 消</el-button>
        <el-button type="primary" @click="submitRack">确 定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
  import {
    createAssetRoom,
    deleteAssetRoom,
    updateAssetRoom,
    getAssetRoomList,
    createAssetRack,
    deleteAssetRack,
    updateAssetRack,
    getAssetRackList
  } from '@/plugin/asset/api/assetDict'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { onMounted, reactive, ref } from 'vue'

  defineOptions({ name: 'AssetRoom' })

  const activeTab = ref('room')

  // 机房
  const roomList = ref([])
  const roomDialogVisible = ref(false)
  const roomFormRef = ref(null)
  const roomForm = reactive({ ID: 0, name: '', region: '', address: '', notes: '' })
  const roomRules = { name: [{ required: true, message: '请输入机房名称', trigger: 'blur' }] }

  const getRooms = async () => {
    const res = await getAssetRoomList()
    if (res.code === 0) roomList.value = res.data || []
  }

  const openRoomDialog = (row) => {
    Object.assign(roomForm, { ID: 0, name: '', region: '', address: '', notes: '' })
    if (row && row.ID) Object.assign(roomForm, row)
    roomDialogVisible.value = true
  }

  const submitRoom = () => {
    roomFormRef.value.validate(async (valid) => {
      if (!valid) return
      const res = roomForm.ID ? await updateAssetRoom(roomForm) : await createAssetRoom(roomForm)
      if (res.code === 0) {
        ElMessage.success(roomForm.ID ? '更新成功' : '创建成功')
        roomDialogVisible.value = false
        getRooms()
      }
    })
  }

  const deleteRoom = (row) => {
    ElMessageBox.confirm(`确定删除机房「${row.name}」吗？`, '提示', {
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      type: 'warning'
    }).then(async () => {
      const res = await deleteAssetRoom({ id: row.ID })
      if (res.code === 0) {
        ElMessage.success('删除成功')
        getRooms()
      }
    })
  }

  // 机柜
  const rackList = ref([])
  const rackDialogVisible = ref(false)
  const rackFormRef = ref(null)
  const rackRoomFilter = ref(undefined)
  const rackForm = reactive({ ID: 0, name: '', roomId: undefined, totalU: 0, notes: '' })
  const rackRules = {
    name: [{ required: true, message: '请输入机柜名称', trigger: 'blur' }],
    roomId: [{ required: true, message: '请选择机房', trigger: 'change' }]
  }

  const getRacks = async () => {
    const res = await getAssetRackList(
      rackRoomFilter.value ? { roomId: rackRoomFilter.value } : undefined
    )
    if (res.code === 0) rackList.value = res.data || []
  }

  const openRackDialog = (row) => {
    Object.assign(rackForm, { ID: 0, name: '', roomId: rackRoomFilter.value || undefined, totalU: 0, notes: '' })
    if (row && row.ID) Object.assign(rackForm, row)
    rackDialogVisible.value = true
  }

  const submitRack = () => {
    rackFormRef.value.validate(async (valid) => {
      if (!valid) return
      const res = rackForm.ID ? await updateAssetRack(rackForm) : await createAssetRack(rackForm)
      if (res.code === 0) {
        ElMessage.success(rackForm.ID ? '更新成功' : '创建成功')
        rackDialogVisible.value = false
        getRacks()
      }
    })
  }

  const deleteRack = (row) => {
    ElMessageBox.confirm(`确定删除机柜「${row.name}」吗？`, '提示', {
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      type: 'warning'
    }).then(async () => {
      const res = await deleteAssetRack({ id: row.ID })
      if (res.code === 0) {
        ElMessage.success('删除成功')
        getRacks()
      }
    })
  }

  onMounted(() => {
    getRooms()
    getRacks()
  })
</script>
