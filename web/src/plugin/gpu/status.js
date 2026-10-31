// gpu 状态 → el-tag 映射（节点/实例两处共用，与 server plugin/gpu/model 常量口径一致）
export const nodeStatusType = (s) =>
  ({ 在线: 'success', 离线: 'info' }[s] || 'info')

// 「已销毁」是台账历史行，plain 描边弱化避免与运行中抢视觉
export const instanceStatusType = (s) =>
  ({ 运行中: 'success', 创建失败: 'danger' }[s] || 'info')
export const instanceStatusEffect = (s) => (s === '已销毁' ? 'plain' : 'light')
