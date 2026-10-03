<template>
  <div>
    <el-card shadow="never">
      <div class="ops-btn-list">
        <el-input
          v-model="username"
          placeholder="操作用户"
          clearable
          style="width: 200px"
          @keyup.enter="getList"
          @clear="getList"
        />
        <el-select v-model="status" placeholder="状态" clearable style="width: 140px" @change="getList">
          <el-option label="进行中" value="进行中" />
          <el-option label="已结束" value="已结束" />
        </el-select>
        <el-button type="primary" icon="search" @click="getList">查询</el-button>
      </div>

      <el-table :data="list" style="width: 100%">
        <el-table-column prop="ID" label="会话ID" width="80" />
        <el-table-column prop="hostname" label="主机" min-width="130" />
        <el-table-column prop="ip" label="IP" min-width="120" />
        <el-table-column prop="username" label="操作用户" min-width="100" />
        <el-table-column prop="clientIp" label="来源IP" min-width="120" />
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="row.status === '进行中' ? 'success' : 'info'">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="指纹" min-width="140">
          <template #default="{ row }">
            <span v-if="row.fingerprint" class="text-xs">{{ row.fingerprint.slice(0, 20) }}…</span>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="startedAt" label="开始时间" min-width="160">
          <template #default="{ row }">{{ (row.startedAt || '').replace('T', ' ').slice(0, 19) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="110" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" icon="video-play" @click="openReplay(row)">回放</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="ops-pagination">
        <el-pagination
          :current-page="page"
          :page-size="pageSize"
          :total="total"
          layout="total, prev, pager, next"
          @current-change="(v) => ((page = v), getList())"
        />
      </div>
    </el-card>

    <el-drawer v-model="replayVisible" :title="`会话回放 · #${replayId} ${replayHost}`" size="70%">
      <el-tabs v-model="replayTab">
        <el-tab-pane label="终端回放" name="stream">
          <div ref="replayEl" class="rounded bg-[#0b1021] p-2" style="height: 60vh" />
        </el-tab-pane>
        <el-tab-pane label="命令列表" name="commands">
          <el-table :data="commands" style="width: 100%" size="small">
            <el-table-column prop="seq" label="序号" width="80" />
            <el-table-column prop="command" label="命令" min-width="300" />
          </el-table>
        </el-tab-pane>
      </el-tabs>
    </el-drawer>
  </div>
</template>

<script setup>
  import { getTermSessionList, getTermSessionStreams, getTermSessionCommands } from '@/plugin/term/api/termSession'
  import { Terminal } from '@xterm/xterm'
  import '@xterm/xterm/css/xterm.css'
  import { nextTick, onMounted, ref } from 'vue'

  defineOptions({ name: 'TermAudit' })

  const username = ref('')
  const status = ref('')
  const list = ref([])
  const page = ref(1)
  const pageSize = ref(20)
  const total = ref(0)

  const getList = async () => {
    const res = await getTermSessionList({
      page: page.value,
      pageSize: pageSize.value,
      username: username.value || undefined,
      status: status.value || undefined
    })
    if (res.code === 0) {
      list.value = res.data.list || []
      total.value = res.data.total || 0
    }
  }

  // 回放
  const replayVisible = ref(false)
  const replayId = ref(0)
  const replayHost = ref('')
  const replayTab = ref('stream')
  const replayEl = ref(null)
  const commands = ref([])

  const openReplay = async (row) => {
    replayId.value = row.ID
    replayHost.value = `${row.hostname}（${row.ip}）`
    replayVisible.value = true
    const [streamsRes, cmdsRes] = await Promise.all([
      getTermSessionStreams({ id: row.ID }),
      getTermSessionCommands({ id: row.ID })
    ])
    commands.value = cmdsRes.data || []
    await nextTick()
    const term = new Terminal({
      fontSize: 12,
      disableStdin: true,
      theme: { background: '#0b1021', foreground: '#e2e8f0' },
      scrollback: 10000
    })
    term.open(replayEl.value)
    for (const s of streamsRes.data || []) {
      if (s.direction === 0) {
        term.write(`\x1b[33m〔输入〕${s.payload.replace(/\r?\n/g, '')}\x1b[0m`)
      } else {
        term.write(s.payload)
      }
    }
    if (!streamsRes.data?.length) term.writeln('\x1b[33m（该会话无流数据）\x1b[0m')
  }

  onMounted(getList)
</script>
