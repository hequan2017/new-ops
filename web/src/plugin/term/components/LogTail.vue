<template>
  <div class="logtail">
    <div style="margin-bottom: 8px; display: flex; gap: 8px; align-items: center">
      <span class="ops-text-muted" style="font-size: 12px">初始行数</span>
      <el-select v-model="lineCount" size="small" style="width: 90px" :disabled="connected" @change="saveLines">
        <el-option v-for="n in [50, 100, 200, 500, 1000]" :key="n" :label="String(n)" :value="n" />
      </el-select>
      <span class="ops-text-muted" style="font-size: 12px">{{ statusText }}</span>
    </div>
    <div ref="boxEl" class="log-box"><pre ref="preEl">{{ text }}</pre></div>
  </div>
</template>

<script setup>
  import { ref, onMounted, onBeforeUnmount } from 'vue'
  import { useUserStore } from '@/pinia/modules/user'

  const props = defineProps({
    hostId: { type: Number, required: true },
    credentialId: { type: Number, required: true },
    path: { type: String, required: true },
    lines: { type: Number, default: 200 }
  })

  const text = ref('')
  const statusText = ref('连接中…')
  const boxEl = ref(null)
  const preEl = ref(null)
  let ws = null
  let closed = false
  let follow = true

  // 初始行数选择记忆（localStorage 持久化；连接中禁改，改行数需断开重连）
  const LS_LINES_KEY = 'logtail:lines'
  const connected = ref(false)
  const lineCount = ref(Number(localStorage.getItem(LS_LINES_KEY)) || props.lines || 200)
  const saveLines = (v) => {
    try {
      localStorage.setItem(LS_LINES_KEY, String(v))
    } catch (e) {
      /* 配额满忽略 */
    }
  }

  const cleanup = () => {
    closed = true
    if (ws) {
      try {
        if (ws.readyState === 1) ws.send(JSON.stringify({ type: 'close' }))
      } catch (e) {
        /* ignore */
      }
      ws.close()
    }
  }

  const scrollToBottom = () => {
    if (follow && boxEl.value) boxEl.value.scrollTop = boxEl.value.scrollHeight
  }

  onMounted(() => {
    const proto = location.protocol === 'https:' ? 'wss' : 'ws'
    const base = import.meta.env.VITE_BASE_API || '/api'
    const qs = new URLSearchParams({
      token: useUserStore().token,
      hostId: String(props.hostId),
      credentialId: String(props.credentialId),
      path: props.path,
      lines: String(lineCount.value)
    })
    ws = new WebSocket(`${proto}://${location.host}${base}/term/logtail?${qs}`)
    ws.binaryType = 'arraybuffer'

    ws.onopen = () => {
      connected.value = true
      statusText.value = `tail -f ${props.path}（滚动到上方即暂停跟随，拉到底恢复）`
    }
    ws.onmessage = (ev) => {
      const chunk = ev.data instanceof ArrayBuffer ? new TextDecoder().decode(ev.data) : ev.data
      text.value += chunk
      if (text.value.length > 512 * 1024) {
        text.value = text.value.slice(-256 * 1024) // 防长会话内存膨胀
      }
      scrollToBottom()
    }
    ws.onclose = (ev) => {
      connected.value = false
      if (closed) return
      statusText.value = ev.code === 4000 ? `连接断开：${ev.reason || ''}` : `连接已关闭 code=${ev.code}`
    }
    ws.onerror = () => {
      statusText.value = '连接错误（检查路径/凭据/网络）'
    }
    // 滚动跟随：用户上滚暂停，滚回底部恢复
    if (boxEl.value) {
      boxEl.value.addEventListener('scroll', () => {
        const el = boxEl.value
        follow = el.scrollHeight - el.scrollTop - el.clientHeight < 40
      })
    }
  })

  onBeforeUnmount(cleanup)
  defineExpose({ cleanup })
</script>

<style scoped>
  .log-box {
    height: calc(100vh - 240px);
    overflow: auto;
    background: #0b1021;
    color: #e2e8f0;
    border-radius: 6px;
    padding: 10px;
  }
  .log-box pre {
    margin: 0;
    font-size: 12px;
    white-space: pre-wrap;
    word-break: break-all;
    font-family: 'JetBrains Mono', Consolas, monospace;
  }
</style>
