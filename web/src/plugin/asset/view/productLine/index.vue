<template>
  <div>
    <el-card shadow="never">
      <div class="ops-btn-list">
        <el-input
          v-model="keyword"
          placeholder="产品线名称 / 负责人"
          clearable
          style="width: 240px"
          @keyup.enter="getLines"
          @clear="getLines"
        >
          <template #prefix>
            <el-icon><search /></el-icon>
          </template>
        </el-input>
        <el-button type="primary" icon="plus" @click="openDialog()">新增产品线</el-button>
      </div>

      <el-table :data="list" style="width: 100%">
        <el-table-column prop="name" label="产品线名称" min-width="160" />
        <el-table-column prop="level" label="等级" width="100">
          <template #default="{ row }">{{ row.level || '-' }}</template>
        </el-table-column>
        <el-table-column prop="owner" label="负责人" min-width="100">
          <template #default="{ row }">{{ row.owner || '-' }}</template>
        </el-table-column>
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

    <el-dialog v-model="dialogVisible" :title="form.ID ? '编辑产品线' : '新增产品线'" width="480px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="90px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="form.name" placeholder="如 核心业务" />
        </el-form-item>
        <el-form-item label="等级">
          <el-input v-model="form.level" placeholder="如 P0 / P1" />
        </el-form-item>
        <el-form-item label="负责人">
          <el-input v-model="form.owner" />
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
    createAssetProductLine,
    deleteAssetProductLine,
    updateAssetProductLine,
    getAssetProductLineList
  } from '@/plugin/asset/api/assetDict'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { onMounted, reactive, ref } from 'vue'

  defineOptions({ name: 'AssetProductLine' })

  const keyword = ref('')
  const list = ref([])
  const dialogVisible = ref(false)
  const formRef = ref(null)
  const form = reactive({ ID: 0, name: '', owner: '', level: '', notes: '' })
  const rules = { name: [{ required: true, message: '请输入产品线名称', trigger: 'blur' }] }

  const getLines = async () => {
    const res = await getAssetProductLineList(keyword.value || undefined)
    if (res.code === 0) list.value = res.data || []
  }

  const openDialog = (row) => {
    Object.assign(form, { ID: 0, name: '', owner: '', level: '', notes: '' })
    if (row && row.ID) Object.assign(form, row)
    dialogVisible.value = true
  }

  const submitForm = () => {
    formRef.value.validate(async (valid) => {
      if (!valid) return
      const res = form.ID ? await updateAssetProductLine(form) : await createAssetProductLine(form)
      if (res.code === 0) {
        ElMessage.success(form.ID ? '更新成功' : '创建成功')
        dialogVisible.value = false
        getLines()
      }
    })
  }

  const onDelete = (row) => {
    ElMessageBox.confirm(`确定删除产品线「${row.name}」吗？删除后主机关联将解除。`, '提示', {
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      type: 'warning'
    }).then(async () => {
      const res = await deleteAssetProductLine({ id: row.ID })
      if (res.code === 0) {
        ElMessage.success('删除成功')
        getLines()
      }
    })
  }

  onMounted(getLines)
</script>
