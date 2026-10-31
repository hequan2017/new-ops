import { formatTimeToStr } from '@/utils/date'
import { getDict } from '@/utils/dictionary'
import { ref } from 'vue'
import { getUrl } from './image'

export const formatBoolean = (bool) => {
  if (bool !== null) {
    return bool ? '是' : '否'
  } else {
    return ''
  }
}
export const formatDate = (time) => {
  if (time !== null && time !== '') {
    var date = new Date(time)
    return formatTimeToStr(date, 'yyyy-MM-dd hh:mm:ss')
  } else {
    return ''
  }
}

// fmtTime 时间列统一格式化：ISO 字符串/时间戳/Date → 'yyyy-MM-dd hh:mm:ss'；
// 空值返回 fallback（默认 '—'，可传领域文案如 '未采集'/'未恢复'）；非时间字符串原样回显
export const fmtTime = (time, fallback = '—') => {
  if (time === null || time === undefined || time === '') return fallback
  const d = new Date(time)
  if (isNaN(d.getTime())) return String(time)
  return formatTimeToStr(d, 'yyyy-MM-dd hh:mm:ss')
}

// relTime 相对时间：刚刚 / N 分钟前 / N 小时前 / N 天前，超 7 天回退绝对时间；
// 未来时间回退绝对时间；空值返回空串
export const relTime = (time) => {
  if (time === null || time === undefined || time === '') return ''
  const t = new Date(time).getTime()
  if (isNaN(t)) return String(time)
  const diff = Date.now() - t
  if (diff <= 0) return fmtTime(time)
  const min = Math.floor(diff / 60000)
  if (min < 1) return '刚刚'
  if (min < 60) return `${min} 分钟前`
  const h = Math.floor(min / 60)
  if (h < 24) return `${h} 小时前`
  const d = Math.floor(h / 24)
  if (d < 7) return `${d} 天前`
  return fmtTime(time)
}

export const filterDict = (value, options) => {
  // 递归查找函数
  const findInOptions = (opts, targetValue) => {
    if (!opts || !Array.isArray(opts)) return null
    
    for (const item of opts) {
      if (item.value === targetValue) {
        return item
      }
      
      if (item.children && Array.isArray(item.children)) {
        const found = findInOptions(item.children, targetValue)
        if (found) return found
      }
    }
    
    return null
  }
  
  const rowLabel = findInOptions(options, value)
  return rowLabel && rowLabel.label
}

export const filterDataSource = (dataSource, value) => {
  // 递归查找函数
  const findInDataSource = (data, targetValue) => {
    if (!data || !Array.isArray(data)) return null
    
    for (const item of data) {
      // 检查当前项是否匹配
      if (item.value === targetValue) {
        return item
      }
      
      // 如果有children属性，递归查找
      if (item.children && Array.isArray(item.children)) {
        const found = findInDataSource(item.children, targetValue)
        if (found) return found
      }
    }
    
    return null
  }
  
  if (Array.isArray(value)) {
    return value.map((item) => {
      const rowLabel = findInDataSource(dataSource, item)
      return rowLabel?.label
    })
  }
  
  const rowLabel = findInDataSource(dataSource, value)
  return rowLabel?.label
}

export const getDictFunc = async (type) => {
  const dicts = await getDict(type)
  return dicts
}

export const ReturnArrImg = (arr) => {
  const imgArr = []
  if (arr instanceof Array) {
    // 如果是数组类型
    for (const arrKey in arr) {
        imgArr.push(getUrl(arr[arrKey]))
    }
  } else {
    imgArr.push(getUrl(arr))
  }
  return imgArr
}

export const returnArrImg = ReturnArrImg

export const onDownloadFile = (url) => {
  window.open(getUrl(url))
}
const colorToHex = (u) => {
  let e = u.replace('#', '').match(/../g)
  for (let t = 0; t < 3; t++) e[t] = parseInt(e[t], 16)
  return e
}

const hexToColor = (u, e, t) => {
  let a = [u.toString(16), e.toString(16), t.toString(16)]
  for (let n = 0; n < 3; n++) a[n].length === 1 && (a[n] = `0${a[n]}`)
  return `#${a.join('')}`
}
const generateAllColors = (u, e) => {
  let t = colorToHex(u)
  const target = [10, 10, 30]
  for (let a = 0; a < 3; a++) t[a] = Math.floor(t[a] * (1 - e) + target[a] * e)
  return hexToColor(t[0], t[1], t[2])
}

const generateAllLightColors = (u, e) => {
  let t = colorToHex(u)
  const target = [240, 248, 255] // RGB for blue white color
  for (let a = 0; a < 3; a++) t[a] = Math.floor(t[a] * (1 - e) + target[a] * e)
  return hexToColor(t[0], t[1], t[2])
}

function addOpacityToColor(u, opacity) {
  let t = colorToHex(u)
  return `rgba(${t[0]}, ${t[1]}, ${t[2]}, ${opacity})`
}

export const setBodyPrimaryColor = (primaryColor, darkMode) => {
  let fmtColorFunc = generateAllColors
  if (darkMode === 'light') {
    fmtColorFunc = generateAllLightColors
  }

  document.documentElement.style.setProperty('--el-color-primary', primaryColor)
  document.documentElement.style.setProperty(
    '--el-color-primary-bg',
    addOpacityToColor(primaryColor, 0.4)
  )
  for (let times = 1; times <= 2; times++) {
    document.documentElement.style.setProperty(
      `--el-color-primary-dark-${times}`,
      fmtColorFunc(primaryColor, times / 10)
    )
  }
  for (let times = 1; times <= 10; times++) {
    document.documentElement.style.setProperty(
      `--el-color-primary-light-${times}`,
      fmtColorFunc(primaryColor, times / 10)
    )
  }
  document.documentElement.style.setProperty(
    `--el-menu-hover-bg-color`,
    addOpacityToColor(primaryColor, 0.2)
  )
}

const baseUrl = ref(import.meta.env.VITE_BASE_API)

export const getBaseUrl = () => {
  return baseUrl.value === '/' ? '' : baseUrl.value
}

export const CreateUUID = () => {
  let d = new Date().getTime()
  if (window.performance && typeof window.performance.now === 'function') {
    d += performance.now()
  }
  return '00000000-0000-0000-0000-000000000000'.replace(/0/g, (c) => {
    const r = (d + Math.random() * 16) % 16 | 0 // d是随机种子
    d = Math.floor(d / 16)
    return (c === '0' ? r : (r & 0x3) | 0x8).toString(16)
  })
}
