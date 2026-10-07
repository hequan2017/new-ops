import fs from 'fs'
import path from 'path'
import chokidar from 'chokidar'

const toPascalCase = (str) => {
  return str.replace(/(^\w|-\w)/g, clearAndUpper)
}

const clearAndUpper = (text) => {
  return text.replace(/-/, '').toUpperCase()
}

// 递归获取目录下所有的 .vue 文件（base 为项目根，路径逐项解析校验，防止符号链接等引用越出项目边界）
const getAllVueFiles = (base, dir, fileList = []) => {
  const baseAbs = path.resolve(base)
  const files = fs.readdirSync(dir)
  files.forEach((file) => {
    const filePath = path.resolve(dir, file)
    const rel = path.relative(baseAbs, filePath)
    if (rel.startsWith('..') || path.isAbsolute(rel)) {
      return
    }
    if (fs.statSync(filePath).isDirectory()) {
      getAllVueFiles(base, filePath, fileList)
    } else if (filePath.endsWith('.vue')) {
      fileList.push(filePath)
    }
  })
  return fileList
}

// 从 .vue 文件内容中提取组件名称
const extractComponentName = (fileContent) => {
  const regex = /defineOptions\(\s*{\s*name:\s*["']([^"']+)["']/
  const match = fileContent.match(regex)
  return match ? match[1] : null
}

// Vite 插件定义
const vueFilePathPlugin = (outputFilePath) => {
  let root
  let isDev = false
  const generatePathNameMap = () => {
    const vueFiles = [
      ...getAllVueFiles(root, path.join(root, 'src/view')),
      ...getAllVueFiles(root, path.join(root, 'src/plugin'))
    ]
    const pathNameMap = vueFiles.reduce((acc, filePath) => {
      const content = fs.readFileSync(filePath, 'utf-8')
      const componentName = extractComponentName(content)
      let relativePath = '/' + path.relative(root, filePath).replace(/\\/g, '/')
      acc[relativePath] =
        componentName || toPascalCase(path.basename(filePath, '.vue'))
      return acc
    }, {})
    const outputContent = JSON.stringify(pathNameMap, null, 2)
    fs.writeFileSync(outputFilePath, outputContent)
  }

  const watchDirectoryChanges = () => {
    const watchDirectories = [
      path.join(root, 'src/view'),
      path.join(root, 'src/plugin')
    ]
    const watcher = chokidar.watch(watchDirectories, {
      persistent: true,
      ignoreInitial: true
    })
    watcher.on('all', () => {
      generatePathNameMap()
    })
  }

  return {
    name: 'vue-file-path-plugin',
    configResolved(resolvedConfig) {
      root = resolvedConfig.root
      if (resolvedConfig.mode === 'development') {
        isDev = true
      }
    },
    buildStart() {
      generatePathNameMap()
    },
    buildEnd() {
      if (isDev) {
        watchDirectoryChanges()
      }
    }
  }
}

export default vueFilePathPlugin
