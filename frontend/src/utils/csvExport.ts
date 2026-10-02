/**
 * CSV 导出工具：列表页「导出全部 / 导出选中」共用。
 * 与旧版 useReport.exportCsv 行为一致：BOM + 逗号/引号/换行转义。
 */

export interface CsvColumn {
  key: string
  label: string
}

/** 单元格格式化：返回最终写入 CSV 的字符串 */
export type CsvCellFormatter<T> = (key: string, row: T) => string

function escapeCell(value: string): string {
  return value.includes(',') || value.includes('"') || value.includes('\n')
    ? `"${value.replace(/"/g, '""')}"`
    : value
}

function defaultCell(value: unknown): string {
  return value === null || value === undefined ? '' : String(value)
}

export function buildCsv<T extends Record<string, any>>(
  columns: CsvColumn[],
  rows: T[],
  format?: CsvCellFormatter<T>,
): string {
  const header = columns.map((c) => escapeCell(c.label)).join(',')
  const body = rows.map((r) =>
    columns.map((c) => escapeCell(format ? format(c.key, r) : defaultCell(r[c.key]))).join(','),
  )
  // U+FEFF BOM：让 Excel 正确识别 UTF-8
  return '﻿' + [header, ...body].join('\n')
}

/**
 * 生成并下载 CSV 文件。
 * @returns true 表示已触发下载；rows 为空时返回 false（由调用方提示）
 */
export function downloadCsv<T extends Record<string, any>>(
  filename: string,
  columns: CsvColumn[],
  rows: T[],
  format?: CsvCellFormatter<T>,
): boolean {
  if (!rows.length) return false
  const csv = buildCsv(columns, rows, format)
  const blob = new Blob([csv], { type: 'text/csv;charset=utf-8' })
  const a = document.createElement('a')
  a.href = URL.createObjectURL(blob)
  a.download = `${filename}_${new Date().toISOString().slice(0, 10)}.csv`
  a.click()
  URL.revokeObjectURL(a.href)
  return true
}
