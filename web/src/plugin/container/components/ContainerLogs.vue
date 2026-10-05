<template>
  <div>
    <div class="ops-text-muted" style="margin-bottom: 8px">
      {{ statusText }}
    </div>
    <div ref="boxEl" class="log-box"><pre ref="preEl">{{ text }}</pre></div>
  </div>
</template>

<script setup>
  import { ref, onMounted, onBeforeUnmount } from 'vue'
  import { useUserStore } from '@/pinia/modules/user'

  const props = defineProps({
    endpointId: { type: Number, required: true },
    containerId: { type: String, required: true },
    tail: { type: Number, default: 300 }
  })

  const text = ref('')
  const statusText = ref('连接中…')
  const boxEl = ref(null)
  const preEl = ref(null)
  let ws = null
  let closed = false
  let follow = true

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

  onMounted(() => {
    const proto = location.protocol === 'https:' ? 'wss' : 'ws'
    const base = import.meta.env.VITE_BASE_API || '/api'
    const qs = new URLSearchParams({
      token: useUserStore().token,
      endpointId: String(props.endpointId),
      id: props.containerId,
      tail: String(props.tail)
    })
    ws = new WebSocket(`${proto}://${location.host}${base}/container/container/logws?${qs}`)
    ws.binaryType = 'arraybuffer'

    ws.onopen = () => {
      statusText.value = 'docker logs -f 跟随中（滚动到上方暂停，拉到底恢复）'
    }
    ws.onmessage = (ev) => {
      const chunk = ev.data instanceof ArrayBuffer ? new TextDecoder().decode(ev.data) : ev.data
      text.value += chunk
      if (text.value.length > 512 * 1024) {
        text.value = text.value.slice(-256 * 1024)
      }
      if (follow && boxEl.value) boxEl.value.scrollTop = boxEl.value.scrollHeight
    }
    ws.onclose = (ev) => {
      if (closed) return
      statusText.value = ev.code === 4000 ? `连接断开：${ev.reason || ''}` : `连接已关闭 code=${ev.code}`
    }
    ws.onerror = () => {
      statusText.value = '连接错误'
    }
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
    height: calc(100vh - 180px);
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
