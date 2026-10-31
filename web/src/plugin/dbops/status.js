// dbops 状态 → el-tag 映射（工单/实例两视图共用，与 server plugin/dbops/model 常量口径一致）
export const orderStatusType = (s) =>
  ({ 待审核: 'warning', 审核通过: 'primary', 执行中: 'primary', 成功: 'success', 失败: 'danger', 已取消: 'info' }[s] || 'info')

// 「审核通过」是静态就绪态（等待执行），用 plain 描边与「执行中」的实心区分同色歧义
export const orderStatusEffect = (s) => (s === '审核通过' ? 'plain' : 'light')

export const instanceStatusType = (s) =>
  ({ 在线: 'success', 离线: 'danger' }[s] || 'info')
