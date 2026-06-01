import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import api from '@/api/client'
import { ElMessage } from 'element-plus'

export interface BillDetailOptions {
  baseUrl: string
  defaultForm: () => Record<string, any>
  defaultItem: () => Record<string, any>
  statusActions?: Record<string, { label: string; api: string; method?: string }[]>
}

export function useBillDetail(options: BillDetailOptions) {
  const { baseUrl, defaultForm, defaultItem, statusActions } = options
  const router = useRouter()

  const form = ref<Record<string, any>>(defaultForm())
  const items = ref<Record<string, any>[]>([])
  const loading = ref(false)
  const saving = ref(false)
  const mode = ref<'create' | 'edit' | 'view'>('view')
  const isEditable = computed(() => mode.value === 'create' || mode.value === 'edit')

  function initCreate() {
    mode.value = 'create'
    form.value = defaultForm()
    items.value = []
  }

  async function loadDetail(id: string | number) {
    loading.value = true
    try {
      const res = await api.get(`${baseUrl}/${id}`)
      if (res.data.code === 0 || res.data.code === 200) {
        const data = res.data.data
        form.value = { ...data }
        items.value = data.items || data.details || []
      }
    } catch (error: any) {
      ElMessage.error(error.message || '加载失败')
    } finally {
      loading.value = false
    }
  }

  function setMode(m: 'create' | 'edit' | 'view') {
    mode.value = m
  }

  function addItem() {
    items.value.push(defaultItem())
  }

  function removeItem(index: number) {
    items.value.splice(index, 1)
  }

  async function save() {
    saving.value = true
    try {
      const payload = { ...form.value, items: items.value }
      delete payload.id
      delete payload.createdAt
      delete payload.updatedAt
      const res = mode.value === 'create'
        ? await api.post(baseUrl, payload)
        : await api.put(`${baseUrl}/${form.value.id}`, payload)
      if (res.data.code === 0 || res.data.code === 200) {
        ElMessage.success('保存成功')
        router.back()
      } else {
        ElMessage.error(res.data.message || '保存失败')
      }
    } catch (error: any) {
      ElMessage.error(error.message || '保存失败')
    } finally {
      saving.value = false
    }
  }

  async function doStatusAction(action: { api: string; method?: string }) {
    try {
      const method = (action.method || 'put').toLowerCase()
      const url = action.api.replace(':id', String(form.value.id))
      const res = await api[method](url)
      if (res.data.code === 0 || res.data.code === 200) {
        ElMessage.success('操作成功')
        await loadDetail(form.value.id)
      } else {
        ElMessage.error(res.data.message || '操作失败')
      }
    } catch (error: any) {
      ElMessage.error(error.message || '操作失败')
    }
  }

  function goBack() {
    router.back()
  }

  const totalAmount = computed(() => {
    return items.value.reduce((sum, item) => sum + (item.totalAmount || item.amount || 0), 0)
  })

  const totalQuantity = computed(() => {
    return items.value.reduce((sum, item) => sum + (item.quantity || 0), 0)
  })

  return {
    form, items, loading, saving, mode, isEditable,
    initCreate, loadDetail, addItem, removeItem, save,
    doStatusAction, goBack, totalAmount, totalQuantity, setMode,
  }
}
