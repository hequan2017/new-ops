<template>
  <div>
    <el-card shadow="never">
      <warning-bar title="资产组用于数据权限划分：普通用户仅能看到其所在资产组内的主机，超级管理员可见全部。" />
      <div class="ops-btn-list">
        <el-button type="primary" icon="plus" @click="openDialog()">新增资产组</el-button>
      </div>

      <el-table :data="list" style="width: 100%">
        <el-table-column prop="name" label="资产组名称" min-width="160" />
        <el-table-column prop="hostNum" label="主机数" width="90" />
        <el-table-column prop="userNum" label="成员数" width="90" />
        <el-table-column prop="notes" label="备注" min-width="180">
          <template #default="{ row }">{{ row.notes || '-' }}</template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" icon="edit" @click="openDialog(row)">编辑</el-button>
            <el-button link type="danger" icon="delete" @click="onDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="form.ID ? '编辑资产组' : '新增资产组'" width="560px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="90px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="form.name" placeholder="如 华东资产组" />
        </el-form-item>
        <el-form-item label="成员用户">
          <el-select v-model="form.userIds" multiple filterable style="width: 100%" placeholder="选择可查看该组的用户">
            <el-option
              v-for="u in userList"
              :key="u.ID"
              :label="`${u.nickName || u.userName}（${u.userName}）`"
              :value="u.ID"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="主机">
          <el-select v-model="form.hostIds" multiple filterable style="width: 100%" placeholder="选择纳入该组的主机">
            <el-option
              v-for="h in hostList"
              :key="h.ID"
              :label="`${h.hostname}（${h.ip}）`"
              :value="h.ID"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.notes" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取 消</el-button>
        <el-button type="primary" @click="submitForm">确 定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
  import {
    createAssetGroup,
    deleteAssetGroup,
    updateAssetGroup,
    getAssetGroupList
  } from '@/plugin/asset/api/assetGroup'
  import { getAssetHostList } from '@/plugin/asset/api/assetHost'
  import { getUserList } from '@/api/user'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { onMounted, reactive, ref } from 'vue'

  defineOptions({ name: 'AssetGroup' })

  const list = ref([])
  const hostList = ref([])
  const userList = ref([])
  const dialogVisible = ref(false)
  const formRef = ref(null)
  const form = reactive({ ID: 0, name: '', notes: '', hostIds: [], userIds: [] })
  const rules = { name: [{ required: true, message: '请输入资产组名称', trigger: 'blur' }] }

  const getList = async () => {
    const res = await getAssetGroupList()
    if (res.code === 0) list.value = res.data || []
  }

  const getHosts = async () => {
    const res = await getAssetHostList({ page: 1, pageSize: 1000 })
    if (res.code === 0) hostList.value = res.data.list || []
  }

  const getUsers = async () => {
    const res = await getUserList({ page: 1, pageSize: 1000 })
    if (res.code === 0) userList.value = res.data.list || []
  }

  const openDialog = (row) => {
    Object.assign(form, { ID: 0, name: '', notes: '', hostIds: [], userIds: [] })
    if (row && row.ID) {
      Object.assign(form, row)
    }
    dialogVisible.value = true
  }

  const submitForm = () => {
    formRef.value.validate(async (valid) => {
      if (!valid) return
      const res = form.ID ? await updateAssetGroup(form) : await createAssetGroup(form)
      if (res.code === 0) {
        ElMessage.success(form.ID ? '更新成功' : '创建成功')
        dialogVisible.value = false
        getList()
      }
    })
  }

  const onDelete = (row) => {
    ElMessageBox.confirm(`确定删除资产组「${row.name}」吗？`, '提示', {
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      type: 'warning'
    }).then(async () => {
      const res = await deleteAssetGroup({ id: row.ID })
      if (res.code === 0) {
        ElMessage.success('删除成功')
        getList()
      }
    })
  }

  onMounted(() => {
    getList()
    getHosts()
    getUsers()
  })
</script>
