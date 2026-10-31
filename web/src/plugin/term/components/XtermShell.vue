<template>
  <div ref="wrapEl" class="relative h-full w-full overflow-hidden bg-[#0b1021]">
    <div ref="termEl" class="h-full w-full" />
  </div>
</template>

<script setup>
  // XtermShell：WebSSH 终端组件
  // WS 二进制帧 → xterm 输出；xterm 输入 → WS 文本帧；控制帧 JSON {type: resize/ping}
  import { onBeforeUnmount, onMounted, ref } from 'vue'
  import { ElMessage } from 'element-plus'
  import { Terminal } from '@xterm/xterm'
  import { FitAddon } from '@xterm/addon-fit'
  import '@xterm/xterm/css/xterm.css'
  import { useUserStore } from '@/pinia/modules/user'

  const props = defineProps({
    hostId: { type: Number, required: true },
    credentialId: { type: Number, required: true },
    hostLabel: { type: String, default: '' }
  })

  const wrapEl = ref(null)
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
    term.writeln('\x1b[36m[白泽] 正在连接 ' + (props.hostLabel || '主机') + ' ...\x1b[0m')
    fit.fit()

    const proto = location.protocol === 'https:' ? 'wss' : 'ws'
    const base = import.meta.env.VITE_BASE_API || '/api'
    const qs = new URLSearchParams({
      token: useUserStore().token,
      hostId: String(props.hostId),
      credentialId: String(props.credentialId),
      cols: String(term.cols),
      rows: String(term.rows)
    })
    ws = new WebSocket(`${proto}://${location.host}${base}/term/ws?${qs}`)
    ws.binaryType = 'arraybuffer'

    ws.onopen = () => {
      term.writeln('\x1b[32m[已连接]\x1b[0m')
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
      try {
        const msg = JSON.parse(ev.data)
        if (msg.type === 'pong') return
      } catch (e) {
        /* 非 JSON 输出按文本写 */
      }
      term.write(ev.data)
    }
    ws.onclose = (ev) => {
      if (closed) return
      // 关闭码分级（与 server webssh.go 口径一致）：1000=SSH 正常结束(exit)；
      // 4000=SSH 读异常；1006=WS 层异常断（网络中断/server 退出）
      if (ev.code === 1000) {
        term.writeln('\x1b[32m[SSH 会话已结束]\x1b[0m')
        ElMessage.info('SSH 会话已结束')
        return
      }
      if (ev.code === 4000) {
        term.writeln(`\x1b[31m[SSH 连接已断开] ${ev.reason || '远端连接中断'}\x1b[0m`)
        ElMessage.warning('SSH 连接异常断开，请检查主机网络后在会话列表重新连接')
        return
      }
      term.writeln(`\x1b[31m[连接已断开 code=${ev.code}]\x1b[0m`)
      ElMessage.warning(`终端连接已断开（code=${ev.code}），可在会话列表重新发起连接`)
    }
    ws.onerror = () => {
      term.writeln('\x1b[31m[连接错误]\x1b[0m')
    }

    term.onData((d) => {
      if (ws.readyState === 1) ws.send(d)
    })

    resizeObserver = new ResizeObserver(() => sendResize())
    resizeObserver.observe(wrapEl.value)
  })

  onBeforeUnmount(cleanup)

  defineExpose({ cleanup })
</script>
