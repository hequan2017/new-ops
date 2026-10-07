import service from '@/utils/request'

// 主机性能指标序列
export const getMetrics = (params) => service({ url: '/monitor/metric/list', method: 'get', params })

// 手动触发全量采集
export const collectNow = () => service({ url: '/monitor/metric/collect', method: 'post' })
