<template>
  <div>
    <el-card shadow="never">
      <div class="ops-btn-list flex flex-wrap items-center gap-2">
        <el-select
          v-model="hostId"
          filterable
          placeholder="选择主机"
          style="width: 220px"
          @change="onHostChange"
        >
          <el-option
            v-for="h in hostList"
            :key="h.ID"
            :label="`${h.hostname}（${h.ip}）`"
            :value="h.ID"
          />
        </el-select>
        <el-select
          v-model="credentialId"
          filterable
          placeholder="选择 SSH 凭据"
          style="width: 220px"
          @change="refresh"
        >
          <el-option
            v-for="c in credList"
            :key="c.ID"
            :label="`${c.name}（${typeLabel(c.type)}${c.username ? ' / ' + c.username : ''}）`"
            :value="c.ID"
          />
        </el-select>
        <el-button icon="refresh" @click="refresh">刷新</el-button>
        <el-button icon="plus" :disabled="!ready" @click="mkdirDialog = true">新建目录</el-button>
        <el-upload
          :show-file-list="false"
          :disabled="!ready"
          :http-request="onUpload"
        >
          <el-button type="primary" icon="upload" :disabled="!ready">上传</el-button>
        </el-upload>
      </div>

      <div class="mb-2 flex items-center gap-1 text-sm">
        <span class="text-gray-500">路径：</span>
        <el-breadcrumb separator="/">
          <el-breadcrumb-item>
            <a class="cursor-pointer" @click="goto('/')">/</a>
          </el-breadcrumb-item>
          <el-breadcrumb-item v-for="(seg, i) in pathSegments" :key="i">
            <a class="cursor-pointer" @click="goto('/' + pathSegments.slice(0, i + 1).join('/'))">
              {{ seg }}
            </a>
          </el-breadcrumb-item>
        </el-breadcrumb>
        <span class="text-gray-400">{{ currentPath }}</span>
      </div>

      <el-table :data="entries" style="width: 100%" v-loading="loading">
        <el-table-column label="名称" min-width="220">
          <template #default="{ row }">
            <a
              v-if="row.isDir"
              class="cursor-pointer text-blue-500"
              @click="goto(currentPath === '/' ? '/' + row.name : currentPath + '/' + row.name)"
            >📁 {{ row.name }}</a>
            <span v-else>📄 {{ row.name }}</span>
          </template>
        </el-table-column>
        <el-table-column label="大小" width="120">
          <template #default="{ row }">{{ row.isDir ? '-' : humanSize(row.size) }}</template>
        </el-table-column>
        <el-table-column prop="mode" label="权限" width="130" />
        <el-table-column label="修改时间" min-width="150">
          <template #default="{ row }">{{ fmtTime(row.modTime) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button v-if="!row.isDir" link type="primary" icon="download" @click="download(row)">下载</el-button>
            <el-button link type="primary" icon="edit" @click="openRename(row)">重命名</el-button>
            <el-button link type="danger" icon="delete" @click="onDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="mkdirDialog" title="新建目录" width="440px">
      <el-input v-model="mkdirName" placeholder="目录名（在当前路径下创建）" />
      <template #footer>
        <el-button @click="mkdirDialog = false">取 消</el-button>
        <el-button type="primary" @click="submitMkdir">创 建</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="renameDialog" title="重命名" width="440px">
      <el-input v-model="renameNewName" />
      <template #footer>
        <el-button @click="renameDialog = false">取 消</el-button>
        <el-button type="primary" @click="submitRename">确 定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
  import {
    sftpList, sftpMkdir, sftpDelete, sftpRename, sftpUpload, sftpDownload
  } from '@/plugin/term/api/sftp'
  import { getAssetHostList } from '@/plugin/asset/api/assetHost'
  import { getCredentialList } from '@/plugin/asset/api/credential'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { computed, onMounted, ref } from 'vue'

  defineOptions({ name: 'TermSftp' })

  const hostId = ref(undefined)
  const credentialId = ref(undefined)
  const hostList = ref([])
  const credList = ref([])
  const entries = ref([])
  const loading = ref(false)
  const currentPath = ref('/root')

  const ready = computed(() => !!(hostId.value && credentialId.value))
  const pathSegments = computed(() =>
    currentPath.value.split('/').filter(Boolean)
  )

  const typeLabel = (v) => ({
    ssh_password: 'SSH 密码', ssh_key: 'SSH 私钥', cloud_ak: '云平台 AccessKey',
    docker_tls: 'Docker TLS', kubeconfig: 'kubeconfig'
  }[v] || v)

  const humanSize = (n) => {
    if (n < 1024) return n + ' B'
    if (n < 1024 * 1024) return (n / 1024).toFixed(1) + ' KB'
    if (n < 1024 * 1024 * 1024) return (n / 1024 / 1024).toFixed(1) + ' MB'
    return (n / 1024 / 1024 / 1024).toFixed(1) + ' GB'
  }
  const fmtTime = (t) => (t ? t.replace('T', ' ').slice(0, 19) : '-')

  const loadHosts = async () => {
    const res = await getAssetHostList({ page: 1, pageSize: 1000 })
    if (res.code === 0) hostList.value = res.data.list || []
  }
  const loadCreds = async () => {
    const res = await getCredentialList()
    if (res.code === 0) {
      credList.value = (res.data || []).filter(
        (c) => c.type === 'ssh_password' || c.type === 'ssh_key'
      )
    }
  }

  const onHostChange = (hid) => {
    const h = hostList.value.find((x) => x.ID === hid)
    if (h?.credentialId) credentialId.value = h.credentialId
    refresh()
  }

  const refresh = () => {
    if (!ready.value) return
    loading.value = true
    sftpList({ hostId: hostId.value, credentialId: credentialId.value, path: currentPath.value })
      .then((res) => {
        if (res.code === 0) entries.value = res.data || []
      })
      .finally(() => (loading.value = false))
  }

  const goto = (p) => {
    currentPath.value = p
    refresh()
  }

  const mkdirDialog = ref(false)
  const mkdirName = ref('')
  const submitMkdir = () => {
    if (!mkdirName.value) return
    const p = currentPath.value === '/' ? '/' + mkdirName.value : currentPath.value + '/' + mkdirName.value
    sftpMkdir({ hostId: hostId.value, credentialId: credentialId.value, path: p }).then((res) => {
      if (res.code === 0) {
        ElMessage.success('创建成功')
        mkdirDialog.value = false
        refresh()
      }
    })
  }

  const onDelete = (row) => {
    ElMessageBox.confirm(`确定删除「${row.name}」吗？${row.isDir ? '目录将被递归删除。' : ''}`, '提示', {
      confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning'
    }).then(async () => {
      const p = currentPath.value === '/' ? '/' + row.name : currentPath.value + '/' + row.name
      const res = await sftpDelete({ hostId: hostId.value, credentialId: credentialId.value, path: p })
      if (res.code === 0) {
        ElMessage.success('删除成功')
        refresh()
      }
    })
  }

  const renameDialog = ref(false)
  const renameNewName = ref('')
  let renameTarget = null
  const openRename = (row) => {
    renameTarget = row
    renameNewName.value = row.name
    renameDialog.value = true
  }
  const submitRename = () => {
    if (!renameNewName.value || !renameTarget) return
    const dir = currentPath.value
    const from = dir === '/' ? '/' + renameTarget.name : dir + '/' + renameTarget.name
    const to = dir === '/' ? '/' + renameNewName.value : dir + '/' + renameNewName.value
    sftpRename({ hostId: hostId.value, credentialId: credentialId.value, from, to }).then((res) => {
      if (res.code === 0) {
        ElMessage.success('重命名成功')
        renameDialog.value = false
        refresh()
      }
    })
  }

  const onUpload = (opt) => {
    const data = new FormData()
    data.append('hostId', hostId.value)
    data.append('credentialId', credentialId.value)
    data.append('path', currentPath.value)
    data.append('file', opt.file)
    sftpUpload(data).then((res) => {
      if (res.code === 0) {
        ElMessage.success('上传成功')
        refresh()
      }
    })
  }

  const download = (row) => {
    const p = currentPath.value === '/' ? '/' + row.name : currentPath.value + '/' + row.name
    sftpDownload({ hostId: hostId.value, credentialId: credentialId.value, path: p }).then((res) => {
      if (res.data) {
        const url = URL.createObjectURL(new Blob([res.data]))
        const a = document.createElement('a')
        a.href = url
        a.download = row.name
        a.click()
        URL.revokeObjectURL(url)
      }
    })
  }

  onMounted(() => {
    loadHosts()
    loadCreds()
  })
</script>
