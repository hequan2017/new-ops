<template>
  <div>
    <el-card shadow="never">
      <warning-bar title="凭据采用 AES-256-GCM 加密存储，主密钥由环境变量注入；任何接口均不回显明文，仅显示末 4 位用于核对。" />
      <div class="ops-btn-list">
        <el-input
          v-model="keyword"
          placeholder="名称 / 用户名 / 备注"
          clearable
          style="width: 240px"
          @keyup.enter="getList"
          @clear="getList"
        >
          <template #prefix><el-icon><search /></el-icon></template>
        </el-input>
        <el-button type="primary" icon="plus" @click="openDialog()">新增凭据</el-button>
      </div>

      <el-table :data="list" style="width: 100%">
        <el-table-column prop="name" label="名称" min-width="150" />
        <el-table-column label="类型" width="130">
          <template #default="{ row }">
            <el-tag :type="typeTag(row.type)" effect="plain">{{ typeLabel(row.type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="username" label="用户名" min-width="110">
          <template #default="{ row }">{{ row.username || '-' }}</template>
        </el-table-column>
        <el-table-column label="凭据" width="120">
          <template #default="{ row }">
            <span v-if="row.secretLast4">****{{ row.secretLast4 }}</span>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="refCount" label="引用" width="80" />
        <el-table-column prop="remark" label="备注" min-width="150">
          <template #default="{ row }">{{ row.remark || '-' }}</template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" icon="edit" @click="openDialog(row)">编辑</el-button>
            <el-button link type="danger" icon="delete" @click="onDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="form.ID ? '编辑凭据' : '新增凭据'" width="520px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="90px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="form.name" placeholder="如 web01-root" />
        </el-form-item>
        <el-form-item label="类型" prop="type">
          <el-select v-model="form.type" style="width: 100%" :disabled="!!form.ID">
            <el-option v-for="t in typeOptions" :key="t.value" :label="t.label" :value="t.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="用户名">
          <el-input v-model="form.username" placeholder="SSH 场景的登录用户" />
        </el-form-item>
        <el-form-item :label="form.ID ? '新凭据' : '凭据'" prop="secretNew">
          <el-input
            v-model="form.secretNew"
            type="textarea"
            :rows="form.type === 'ssh_key' || form.type === 'kubeconfig' ? 5 : 1"
            :placeholder="form.ID ? '留空则保留原凭据' : '密码 / 私钥 / AccessKeySecret 等敏感内容'"
          />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.remark" type="textarea" :rows="2" />
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
    createCredential,
    deleteCredential,
    updateCredential,
    getCredentialList
  } from '@/plugin/asset/api/credential'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { onMounted, reactive, ref } from 'vue'

  defineOptions({ name: 'AssetCredential' })

  const typeOptions = [
    { value: 'ssh_password', label: 'SSH 密码' },
    { value: 'ssh_key', label: 'SSH 私钥' },
    { value: 'cloud_ak', label: '云平台 AccessKey' },
    { value: 'docker_tls', label: 'Docker TLS' },
    { value: 'kubeconfig', label: 'kubeconfig' }
  ]
  const typeLabel = (v) => typeOptions.find((t) => t.value === v)?.label || v
  const typeTag = (v) => {
    switch (v) {
      case 'ssh_password': return 'info'
      case 'ssh_key': return 'success'
      case 'cloud_ak': return 'warning'
      case 'docker_tls': return ''
      case 'kubeconfig': return 'danger'
      default: return 'info'
    }
  }

  const keyword = ref('')
  const list = ref([])
  const dialogVisible = ref(false)
  const formRef = ref(null)
  const emptyForm = () => ({ ID: 0, name: '', type: 'ssh_password', username: '', secretNew: '', remark: '' })
  const form = reactive(emptyForm())
  const rules = {
    name: [{ required: true, message: '请输入凭据名称', trigger: 'blur' }],
    type: [{ required: true, message: '请选择类型', trigger: 'change' }],
    secretNew: [
      {
        validator: (rule, value, callback) => {
          if (!form.ID && !value) callback(new Error('请输入凭据内容'))
          else callback()
        },
        trigger: 'blur'
      }
    ]
  }

  const getList = async () => {
    const res = await getCredentialList(keyword.value || undefined)
    if (res.code === 0) list.value = res.data || []
  }

  const openDialog = (row) => {
    Object.assign(form, emptyForm())
    if (row && row.ID) {
      Object.assign(form, { ID: row.ID, name: row.name, type: row.type, username: row.username, remark: row.remark })
    }
    dialogVisible.value = true
  }

  const submitForm = () => {
    formRef.value.validate(async (valid) => {
      if (!valid) return
      let res
      if (form.ID) {
        res = await updateCredential({
          ID: form.ID, name: form.name, username: form.username,
          secret: form.secretNew || '', remark: form.remark
        })
      } else {
        res = await createCredential({
          name: form.name, type: form.type, username: form.username,
          secret: form.secretNew, remark: form.remark
        })
      }
      if (res.code === 0) {
        ElMessage.success(form.ID ? '更新成功' : '创建成功')
        dialogVisible.value = false
        getList()
      }
    })
  }

  const onDelete = (row) => {
    ElMessageBox.confirm(`确定删除凭据「${row.name}」吗？`, '提示', {
      confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning'
    }).then(async () => {
      const res = await deleteCredential({ id: row.ID })
      if (res.code === 0) {
        ElMessage.success('删除成功')
        getList()
      }
    })
  }

  onMounted(getList)
</script>
