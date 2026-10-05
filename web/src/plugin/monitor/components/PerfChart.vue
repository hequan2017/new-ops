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
  import { ref, onMounted, onBeforeUnmount, nextTick } from 'vue'
  import * as echarts from 'echarts'
  import { ElMessage } from 'element-plus'
  import { getMetrics, collectNow } from '@/plugin/monitor/api/monitor'

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
    { name: 'load1', color: '#34d399' }
  ]

  const render = (rows) => {
    if (!chart) return
    const byName = {}
    const tsSet = new Map()
    rows.forEach((r) => {
      byName[r.name] = byName[r.name] || []
      byName[r.name].push([r.ts, r.value])
      tsSet.set(r.ts, true)
    })
    points.value = tsSet.size
    chart.setOption({
      tooltip: { trigger: 'axis' },
      legend: { data: SERIES.map((s) => s.name), top: 0 },
      grid: { left: 48, right: 24, top: 32, bottom: 32 },
      xAxis: { type: 'time' },
      yAxis: [
        { type: 'value', max: 100, name: '%' },
        { type: 'value', name: 'load', position: 'right', splitLine: false }
      ],
      series: SERIES.map((s, i) => ({
        name: s.name,
        type: 'line',
        showSymbol: false,
        smooth: true,
        yAxisIndex: s.name === 'load1' ? 1 : 0,
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

  onMounted(async () => {
    await nextTick()
    chart = echarts.init(chartEl.value)
    window.addEventListener('resize', () => chart && chart.resize())
    load()
  })
  onBeforeUnmount(() => {
    if (chart) chart.dispose()
  })
</script>
