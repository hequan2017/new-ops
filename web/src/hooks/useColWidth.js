// 表格列宽拖拽记忆：el-table 加 border 后可拖拽列宽，header-dragend 落 localStorage，
// 模板列上用 :width="colW('prop')" 回填（有记忆值时覆盖 min-width 的弹性布局）。
// 用法：const { colW, onHeaderDragend } = useColWidthMemory('key')；表格加 border 与 @header-dragend。
export const useColWidthMemory = (storageKey) => {
  const load = () => {
    try {
      const v = JSON.parse(localStorage.getItem(storageKey) || '{}')
      return v && typeof v === 'object' ? v : {}
    } catch {
      return {}
    }
  }
  const saved = load()

  // colW 返回记忆宽度；无记忆时返回 def（def 缺省 undefined → 不设 width，保持 min-width 弹性）
  const colW = (prop, def) => (saved[prop] >= 60 ? Number(saved[prop]) : def)

  const onHeaderDragend = (newWidth, _oldWidth, column) => {
    const key = column?.prop || column?.label
    if (!key) return
    saved[key] = Math.round(newWidth)
    try {
      localStorage.setItem(storageKey, JSON.stringify(saved))
    } catch {
      /* 配额满等异常忽略 */
    }
  }

  return { colW, onHeaderDragend }
}
