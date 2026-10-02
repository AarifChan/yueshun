import * as XLSX from 'xlsx'

// ensureXlsxFile 老格式 .xls 在浏览器内转换为 .xlsx 后再上传
// （后端只解析 .xlsx：Go 的 xls 解析器对真实业务文件会静默丢数据）
export async function ensureXlsxFile(file: File): Promise<File> {
  if (!/\.xls$/i.test(file.name)) return file
  const wb = XLSX.read(await file.arrayBuffer(), { type: 'array' })
  const out = XLSX.write(wb, { type: 'array', bookType: 'xlsx' }) as ArrayBuffer
  return new File([out], file.name.replace(/\.xls$/i, '.xlsx'), {
    type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
  })
}
