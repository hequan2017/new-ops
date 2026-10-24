// 主机下拉数据源组合式函数：模块级缓存 + 失效刷新。
// 原本 7 个页面在各自 onMounted 重复全量请求（pageSize 500/1000 口径不一，超限静默缺项），
// 统一为一次请求 + 同会话复用；打开弹窗时可传 force 刷新。
import { ref } from 'vue'
import { getAssetHostList } from '@/plugin/asset/api/assetHost'

const hostOptions = ref([])
let loadingPromise = null

const fetchHostOptions = async (force = false) => {
  if (!force && (hostOptions.value.length || loadingPromise)) return loadingPromise
  loadingPromise = (async () => {
    const res = await getAssetHostList({ page: 1, pageSize: 1000 })
    if (res.code === 0) hostOptions.value = res.data?.list || []
    loadingPromise = null
  })()
  return loadingPromise
}

export const useHostOptions = () => {
  return { hostOptions, fetchHostOptions }
}
