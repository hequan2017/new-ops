import fs from 'fs'

// 白泽本地 svgBuilder：替代 vite-auto-import-svg（该 npm 包会在生产构建时
// 向产物注入"未授权"徽标、回传 beacon 与域名校验脚本，并改写 keywords meta）。
// 本地实现仅保留合法功能：扫描 svg 目录 → 生成 symbol sprite 注入 index.html。
const svgTitle = /<svg([^>+].*?)>/
const clearHeightWidth = /(width|height)="([^>+].*?)"/g
const hasViewBox = /(viewBox="[^>+].*?")/g
const clearReturn = /(\r)|(\n)/g

function findSvgFile(dirs) {
  const svgRes = []
  for (const dir of dirs) {
    const dirents = fs.readdirSync(dir, { withFileTypes: true })
    for (const dirent of dirents) {
      let pluginName = ''
      if (dir.startsWith('./src/plugin')) {
        pluginName = `${dir.split('/')[3]}-`
      }
      if (dirent.isDirectory()) {
        svgRes.push(...findSvgFile([dir + dirent.name + '/']))
      } else if (dirent.name.endsWith('.svg')) {
        const svg = fs
          .readFileSync(dir + dirent.name)
          .toString()
          .replace(clearReturn, '')
          .replace(svgTitle, ($1, $2) => {
            let width = 0
            let height = 0
            let content = $2.replace(clearHeightWidth, (s1, s2, s3) => {
              if (s2 === 'width') {
                width = s3
              } else if (s2 === 'height') {
                height = s3
              }
              return ''
            })
            if (!hasViewBox.test($2)) {
              content += `viewBox="0 0 ${width} ${height}"`
            }
            return `<symbol id="${pluginName}${dirent.name.replace('.svg', '')}" ${content}>`
          })
          .replace('</svg>', '</symbol>')
        svgRes.push(svg)
      }
    }
  }
  return svgRes
}

export const svgBuilder = (dirs) => {
  let root
  return {
    name: 'svg-transform',
    configResolved(resolvedConfig) {
      root = resolvedConfig.root
    },
    transformIndexHtml(html) {
      const res = findSvgFile(dirs)
      return html.replace('<body>', `
<body>
  <svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" style="position: absolute; width: 0; height: 0">
    ${res.join('')}
  </svg>
`)
    }
  }
}
