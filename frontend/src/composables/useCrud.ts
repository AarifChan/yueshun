import { ref, computed } from 'vue'
import api from '@/api/client'
import { ElMessage, ElMessageBox } from 'element-plus'

export interface CrudOptions<T> {
  baseUrl: string
  listUrl?: string
  createUrl?: string
  updateUrl?: string
  deleteUrl?: string
  defaultForm?: () => Partial<T>
  statusActions?: Record<string, { label: string; api: string; method?: string; confirm?: string }[]>
}

export function useCrud<T extends Record<string, any>>(options: CrudOptions<T>) {
  const { baseUrl, listUrl, createUrl, updateUrl, deleteUrl, defaultForm, statusActions } = options

  const list = ref<T[]>([])
  const total = ref(0)
  const loading = ref(false)
  const dialogVisible = ref(false)
  const dialogTitle = ref('')
  const form = ref<Partial<T>>(defaultForm ? defaultForm() : {})
  const isEdit = ref(false)
  const currentId = ref<number | string | null>(null)
  const searchForm = ref<Record<string, any>>({})
  const pagination = ref({ page: 1, pageSize: 20 })

  const queryParams = computed(() => ({
    page: pagination.value.page,
    pageSize: pagination.value.pageSize,
    ...searchForm.value,
  }))

  async function fetchList(params?: Record<string, any>) {
    loading.value = true
    try {
      const res = await api.get(listUrl || baseUrl, {
        params: { ...queryParams.value, ...params },
      })
      if (res.data.code === 0 || res.data.code === 200) {
        const data = res.data.data
        list.value = data.list || data.items || data || []
        total.value = data.total || data.length || 0
      }
    } catch (error: any) {
      ElMessage.error(error.message || '获取列表失败')
    } finally {
      loading.value = false
    }
  }

  function openCreate() {
    form.value = defaultForm ? defaultForm() : {}
    isEdit.value = false
    currentId.value = null
    dialogTitle.value = '新增'
    dialogVisible.value = true
  }

  function openEdit(row: T) {
    form.value = { ...row }
    isEdit.value = true
    currentId.value = row.id
    dialogTitle.value = '编辑'
    dialogVisible.value = true
  }

  async function handleSubmit() {
    try {
      const url = isEdit.value
        ? `${updateUrl || baseUrl}/${currentId.value}`
        : createUrl || baseUrl
      const method = isEdit.value ? 'put' : 'post'
      const res = await api[method](url, form.value)
      if (res.data.code === 0 || res.data.code === 200) {
        ElMessage.success(`${dialogTitle.value}成功`)
        dialogVisible.value = false
        await fetchList()
      } else {
        ElMessage.error(res.data.message || '操作失败')
      }
    } catch (error: any) {
      ElMessage.error(error.message || '操作失败')
    }
  }

  async function handleDelete(row: T) {
    try {
      await ElMessageBox.confirm('确认删除该记录？', '提示', { type: 'warning' })
      const res = await api.delete(`${deleteUrl || baseUrl}/${row.id}`)
      if (res.data.code === 0 || res.data.code === 200) {
        ElMessage.success('删除成功')
        await fetchList()
      } else {
        ElMessage.error(res.data.message || '删除失败')
      }
    } catch (error: any) {
      if (error !== 'cancel') {
        ElMessage.error(error.message || '删除失败')
      }
    }
  }

  async function handleStatusAction(row: T, action: { api: string; method?: string; confirm?: string }) {
    if (action.confirm) {
      try {
        await ElMessageBox.confirm(action.confirm, '提示', { type: 'warning' })
      } catch {
        return
      }
    }
    try {
      const method = (action.method || 'put').toLowerCase()
      const url = action.api.replace(':id', String(row.id))
      let res
      if (method === 'post') {
        res = await api.post(url)
      } else if (method === 'delete') {
        res = await api.delete(url)
      } else {
        res = await api.put(url)
      }
      if (res.data.code === 0 || res.data.code === 200) {
        ElMessage.success('操作成功')
        await fetchList()
      } else {
        ElMessage.error(res.data.message || '操作失败')
      }
    } catch (error: any) {
      ElMessage.error(error.message || '操作失败')
    }
  }

  function getStatusActions(row: T) {
    if (!statusActions || !row.status) return []
    return statusActions[row.status] || []
  }

  function handleSizeChange(size: number) {
    pagination.value.pageSize = size
    pagination.value.page = 1
    fetchList()
  }

  function handleCurrentChange(page: number) {
    pagination.value.page = page
    fetchList()
  }

  function handleSearch() {
    pagination.value.page = 1
    fetchList()
  }

  function handleReset() {
    searchForm.value = {}
    pagination.value.page = 1
    fetchList()
  }

  return {
    list, total, loading,
    dialogVisible, dialogTitle, form, isEdit, currentId,
    searchForm, pagination, queryParams,
    fetchList, openCreate, openEdit, handleSubmit, handleDelete,
    handleStatusAction, getStatusActions,
    handleSizeChange, handleCurrentChange, handleSearch, handleReset,
  }
}
