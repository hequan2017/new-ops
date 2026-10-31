<template>
  <div>
    <div style="margin-bottom: 8px; display: flex; gap: 8px; align-items: center">
      <el-radio-group v-model="hours" size="small" @change="load">
        <el-radio-button :value="1">1h</el-radio-button>
        <el-radio-button :value="6">6h</el-radio-button>
        <el-radio-button :value="24">24h</el-radio-button>
        <el-radio-button :value="168">7d</el-radio-button>
      </el-radio-group>
      <el-button size="small" :loading="loading" @click="load">刷 新</el-button>
      <el-button size="small" type="primary" plain @click="collectAndLoad">立即采集</el-button>
      <span class="ops-text-muted" style="font-size: 12px">
        {{ points > 0 ? `${points} 个采样点（每 5 分钟自动采集）` : '暂无数据，点击「立即采集」生成首批指标' }}
      </span>
    </div>
    <div ref="chartEl" style="height: 340px; width: 100%" />
  </div>
</template>

<script setup>
  import { ref, onMounted, onBeforeUnmount, nextTick, watch } from 'vue'
  import * as echarts from 'echarts'
  import { ElMessage } from 'element-plus'
  import { getMetrics, collectNow } from '@/plugin/monitor/api/monitor'
  import { useAppStore } from '@/pinia/modules/app'

  const props = defineProps({
    assetId: { type: Number, required: true },
    label: { type: String, default: '' }
  })

  const chartEl = ref(null)
  const hours = ref(24)
  const loading = ref(false)
  const points = ref(0)
  let chart = null

  const SERIES = [
    { name: 'cpu_percent', color: '#38bdf8' },
    { name: 'mem_percent', color: '#a78bfa' },
    { name: 'disk_percent', color: '#fbbf24' },
    { name: 'load1', color: '#34d399' },
    { name: 'net_rx_kbs', color: '#f472b6' },
    { name: 'net_tx_kbs', color: '#fb923c' }
  ]

  // 指标单位（tooltip 展示口径，与 server 采集字段一一对应）
  const UNIT_BY_SERIES = {
    cpu_percent: '%',
    mem_percent: '%',
    disk_percent: '%',
    load1: '',
    net_rx_kbs: ' KB/s',
    net_tx_kbs: ' KB/s'
  }

  // 暗色适配：轴/图例/tooltip 文字色跟随主题（ECharts 不吃 CSS 变量，需 JS 计算）
  const appStore = useAppStore()
  const chartText = () => (appStore.isDark ? '#cbd5e1' : '#475569')
  const chartAxis = () => (appStore.isDark ? 'rgba(255,255,255,0.12)' : 'rgba(148,163,184,0.25)')

  // 主题切换时用缓存数据重渲染（lastRows 由 render 维护）
  let lastRows = []
  watch(
    () => appStore.isDark,
    () => {
      if (chart) render(lastRows)
    }
  )

  const render = (rows) => {
    if (!chart) return
    lastRows = rows
    const byName = {}
    const tsSet = new Map()
    rows.forEach((r) => {
      byName[r.name] = byName[r.name] || []
      byName[r.name].push([r.ts, r.value])
      tsSet.set(r.ts, true)
    })
    points.value = tsSet.size
    chart.setOption({
      tooltip: {
        trigger: 'axis',
        backgroundColor: appStore.isDark ? 'rgba(15,23,42,0.95)' : '#ffffff',
        borderColor: chartAxis(),
        textStyle: { color: chartText() },
        formatter: (params) => {
          const lines = params.map((p) => {
            const unit = UNIT_BY_SERIES[p.seriesName]
            const v = Array.isArray(p.value) ? p.value[1] : p.value
            return `${p.marker}${p.seriesName}&nbsp;&nbsp;<b>${v ?? '—'}${unit ?? ''}</b>`
          })
          return `${params[0]?.axisValueLabel || ''}<br/>${lines.join('<br/>')}`
        }
      },
      legend: { data: SERIES.map((s) => s.name), top: 0, textStyle: { color: chartText() } },
      grid: { left: 48, right: 24, top: 32, bottom: 32 },
      xAxis: { type: 'time', axisLabel: { color: chartText() }, splitLine: { lineStyle: { color: chartAxis() } } },
      yAxis: [
        { type: 'value', max: 100, name: '%', axisLabel: { color: chartText() }, splitLine: { lineStyle: { color: chartAxis() } } },
        { type: 'value', name: 'load', position: 'right', splitLine: false, axisLabel: { color: chartText() } }
      ],
      series: SERIES.map((s, i) => ({
        name: s.name,
        type: 'line',
        showSymbol: false,
        smooth: true,
        yAxisIndex: s.name === 'load1' || s.name.startsWith('net_') ? 1 : 0,
        itemStyle: { color: s.color },
        data: byName[s.name] || []
      }))
    })
  }

  const load = async () => {
    loading.value = true
    try {
      const res = await getMetrics({ assetId: props.assetId, hours: hours.value })
      if (res.code === 0) render(res.data || [])
    } finally {
      loading.value = false
    }
  }

  const collectAndLoad = async () => {
    const res = await collectNow()
    if (res.code === 0) {
      ElMessage.success('采集已触发，5 秒后刷新')
      setTimeout(load, 5000)
    }
  }

  // 命名 handler 便于卸载移除（原匿名监听器泄漏且 dispose 后未置 null，触发时对已销毁实例调 resize）
  const onResize = () => chart && chart.resize()
  onMounted(async () => {
    await nextTick()
    chart = echarts.init(chartEl.value)
    window.addEventListener('resize', onResize)
    load()
  })
  onBeforeUnmount(() => {
    window.removeEventListener('resize', onResize)
    chart?.dispose()
    chart = null
  })
</script>
