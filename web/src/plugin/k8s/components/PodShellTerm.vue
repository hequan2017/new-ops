<template>
  <div class="h-full">
    <div ref="termEl" class="pod-term" />
  </div>
</template>

<script setup>
  import { ref, onMounted, onBeforeUnmount } from 'vue'
  import { Terminal } from '@xterm/xterm'
  import { FitAddon } from '@xterm/addon-fit'
  import { useUserStore } from '@/pinia/modules/user'

  const props = defineProps({
    clusterId: { type: Number, required: true },
    namespace: { type: String, required: true },
    pod: { type: String, required: true },
    container: { type: String, default: '' }
  })

  const termEl = ref(null)
  let term = null
  let fit = null
  let ws = null
  let pingTimer = null
  let resizeObserver = null
  let closed = false

  const sendResize = () => {
    if (!ws || ws.readyState !== 1 || !fit) return
    const { cols, rows } = fit.proposeDimensions()
    if (cols > 2 && rows > 2) {
      ws.send(JSON.stringify({ type: 'resize', cols, rows }))
      fit.fit()
    }
  }

  const cleanup = () => {
    closed = true
    if (pingTimer) clearInterval(pingTimer)
    if (resizeObserver) resizeObserver.disconnect()
    if (ws) ws.close()
    if (term) term.dispose()
  }

  onMounted(() => {
    term = new Terminal({
      fontSize: 13,
      cursorBlink: true,
      theme: {
        background: '#0b1021',
        foreground: '#e2e8f0',
        cursor: '#38bdf8',
        selectionBackground: 'rgba(56,189,248,0.3)'
      },
      scrollback: 5000
    })
    fit = new FitAddon()
    term.loadAddon(fit)
    term.open(termEl.value)
    term.writeln('\x1b[36m[白泽] 连接 Pod ' + props.namespace + '/' + props.pod + (props.container ? ' · ' + props.container : '') + ' ...\x1b[0m')
    fit.fit()

    const proto = location.protocol === 'https:' ? 'wss' : 'ws'
    const base = import.meta.env.VITE_BASE_API || '/api'
    const qs = new URLSearchParams({
      token: useUserStore().token,
      clusterId: String(props.clusterId),
      namespace: props.namespace,
      pod: props.pod
    })
    if (props.container) qs.set('container', props.container)
    qs.set('cols', String(term.cols))
    qs.set('rows', String(term.rows))
    ws = new WebSocket(`${proto}://${location.host}${base}/k8s/pod/execws?${qs}`)
    ws.binaryType = 'arraybuffer'

    ws.onopen = () => {
      term.writeln('\x1b[32m[已连接 /bin/sh]\x1b[0m')
      sendResize()
      pingTimer = setInterval(() => {
        if (ws.readyState === 1) ws.send(JSON.stringify({ type: 'ping' }))
      }, 30000)
    }
    ws.onmessage = (ev) => {
      if (ev.data instanceof ArrayBuffer) {
        term.write(new Uint8Array(ev.data))
        return
      }
      term.write(ev.data)
    }
    ws.onclose = (ev) => {
      if (closed) return
      term.writeln(ev.code === 4000 ? `\x1b[31m[连接断开] ${ev.reason || ''}\x1b[0m` : `\x1b[31m[连接已断开 code=${ev.code}]\x1b[0m`)
    }
    ws.onerror = () => {
      term.writeln('\x1b[31m[连接错误]\x1b[0m')
    }
    term.onData((d) => {
      if (ws.readyState === 1) ws.send(d)
    })
    resizeObserver = new ResizeObserver(() => sendResize())
    resizeObserver.observe(termEl.value)
  })

  onBeforeUnmount(cleanup)
  defineExpose({ cleanup })
</script>

<style scoped>
.pod-term {
  height: calc(100vh - 200px);
}
</style>
