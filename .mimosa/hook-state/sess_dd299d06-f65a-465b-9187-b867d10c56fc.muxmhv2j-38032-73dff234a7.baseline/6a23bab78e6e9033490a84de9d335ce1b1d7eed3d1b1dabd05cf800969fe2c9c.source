import service from '@/utils/request'

// 发起批量命令执行（异步，返回批次）
export const createBatchExec = (data) => {
  return service({ url: '/job/exec', method: 'post', data })
}

// 取消进行中的批次
export const cancelBatchExec = (params) => {
  return service({ url: '/job/exec/cancel', method: 'post', params })
}

// 批次分页列表（可带 status 过滤）
export const getBatchList = (data, params) => {
  return service({ url: '/job/exec/list', method: 'post', data, params })
}

// 批次详情与每主机结果
export const getBatchDetail = (params) => {
  return service({ url: '/job/exec/detail', method: 'get', params })
}
