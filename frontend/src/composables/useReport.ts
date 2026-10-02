import { ref } from 'vue'
import api from '@/api/client'
import { ElMessage } from 'element-plus'
import { downloadCsv, type CsvCellFormatter, type CsvColumn } from '@/utils/csvExport'

/**
 * 报表页通用逻辑：筛选 + 分页 + CSV 导出
 */
export function useReport<T extends Record<string, any>>(url: string, defaultFilters: Record<string, any> = {}) {
  const list = ref<T[]>([])
  const total = ref(0)
  const loading = ref(false)
  const page = ref(1)
  const pageSize = ref(30)
  const filters = ref<Record<string, any>>({ ...defaultFilters })
  const extra = ref<Record<string, any>>({})

  async function load(p = page.value) {
    page.value = p
    loading.value = true
    try {
      const res = await api.get(url, {
        params: { page: page.value, pageSize: pageSize.value, ...filters.value },
      })
      if (res.data.code === 0 || res.data.code === 200) {
        const data = res.data.data || {}
        list.value = data.list ?? []
        total.value = data.total ?? 0
        const { list: _l, total: _t, page: _p, pageSize: _ps, ...rest } = data
        extra.value = rest
      }
    } finally {
      loading.value = false
    }
  }

  function search() { load(1) }

  function reset() {
    filters.value = { ...defaultFilters }
    load(1)
  }

  /** 导出当前筛选条件下全部数据为 CSV */
  async function exportCsv(filename: string, columns: CsvColumn[], format?: CsvCellFormatter<T>) {
    loading.value = true
    try {
      const res = await api.get(url, { params: { page: 1, pageSize: 10000, ...filters.value } })
      if (res.data.code !== 0 && res.data.code !== 200) return
      const rows: T[] = res.data.data?.list ?? []
      if (!downloadCsv(filename, columns, rows, format)) {
        ElMessage.warning('没有可导出的数据')
      }
    } finally {
      loading.value = false
    }
  }

  /** 导出选中行（列与「导出全部」一致） */
  function exportSelectedCsv(filename: string, columns: CsvColumn[], rows: T[], format?: CsvCellFormatter<T>) {
    if (!rows.length) {
      ElMessage.warning('请先勾选要导出的行')
      return
    }
    downloadCsv(filename, columns, rows, format)
  }

  return { list, total, loading, page, pageSize, filters, extra, load, search, reset, exportCsv, exportSelectedCsv }
}
