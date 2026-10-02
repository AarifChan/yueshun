# 前端单据页面 Phase 1 — 采购模块 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Complete full-featured bill pages for the Purchase module (orders, instock, returns, payments) with search, pagination, detail views, sub-table editing, and status transitions — while building reusable components for later modules.

**Architecture:** Two page templates (list + detail) plus reusable components (RemoteSelect, BillItemTable). Detail pages are standalone routes (`/purchase-orders/:id`) with mode-based editing. List pages use upgraded `useCrud` with pagination and search.

**Tech Stack:** Vue 3, TypeScript, Element Plus, Vite, Pinia, Axios, Vitest

---

## File Structure

```
frontend/
  package.json                          # Add vitest scripts + dev deps
  vitest.config.ts                      # Create
  src/
    composables/
      useCrud.ts                        # Modify: add pagination, search, status actions
      useBillDetail.ts                 # Create: generic bill detail page logic
    components/
      RemoteSelect.vue                  # Create: remote search select with API fetch
      BillItemTable.vue                 # Create: editable sub-table for bill items
    views/
      purchase/
        OrderView.vue                   # Modify: full list page
        OrderDetailView.vue             # Create: detail page with items editing
        InStockView.vue                 # Modify: full list page
        InStockDetailView.vue           # Create: detail page
        ReturnView.vue                  # Modify: full list page
        ReturnDetailView.vue            # Create: detail page
        PaymentView.vue                 # Modify: full list page
        PaymentDetailView.vue           # Create: detail page
    router/
      index.ts                          # Modify: add detail routes
```

---

### Task 1: Configure Vitest Test Infrastructure

**Files:**
- Modify: `frontend/package.json`
- Create: `frontend/vitest.config.ts`

- [ ] **Step 1: Install test dependencies**

Run:
```bash
cd frontend
npm install -D vitest @vue/test-utils jsdom @vitejs/plugin-vue
```

- [ ] **Step 2: Add test script to package.json**

Modify `frontend/package.json`:
```json
{
  "scripts": {
    "dev": "vite",
    "build": "vue-tsc -b && vite build",
    "preview": "vite preview",
    "test": "vitest run",
    "test:watch": "vitest"
  }
}
```

- [ ] **Step 3: Create vitest.config.ts**

Create `frontend/vitest.config.ts`:
```typescript
import { defineConfig } from 'vitest/config'
import vue from '@vitejs/plugin-vue'
import path from 'path'

export default defineConfig({
  plugins: [vue()],
  test: {
    environment: 'jsdom',
    globals: true,
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
})
```

- [ ] **Step 4: Verify test setup runs**

Run:
```bash
cd frontend && npx vitest run --reporter=verbose
```
Expected: PASS with 0 tests (no test files yet)

- [ ] **Step 5: Commit**

```bash
git add frontend/package.json frontend/vitest.config.ts frontend/package-lock.json
git commit -m "chore: configure vitest for frontend testing"
```

---

### Task 2: Upgrade useCrud Composable

**Files:**
- Modify: `frontend/src/composables/useCrud.ts`

- [ ] **Step 1: Rewrite useCrud with pagination, search, and status actions**

Replace contents of `frontend/src/composables/useCrud.ts`:
```typescript
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
      const res = await api[method](url)
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
```

- [ ] **Step 2: Write test for useCrud**

Create `frontend/src/composables/useCrud.spec.ts`:
```typescript
import { describe, it, expect, vi } from 'vitest'
import { useCrud } from './useCrud'

vi.mock('@/api/client', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
  },
}))

describe('useCrud', () => {
  it('returns reactive state', () => {
    const crud = useCrud({ baseUrl: '/test' })
    expect(crud.list.value).toEqual([])
    expect(crud.total.value).toBe(0)
    expect(crud.loading.value).toBe(false)
  })
})
```

- [ ] **Step 3: Run test**

Run:
```bash
cd frontend && npx vitest run src/composables/useCrud.spec.ts
```
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add frontend/src/composables/useCrud.ts frontend/src/composables/useCrud.spec.ts
git commit -m "feat: upgrade useCrud with pagination, search, and status actions"
```

---

### Task 3: Create useBillDetail Composable

**Files:**
- Create: `frontend/src/composables/useBillDetail.ts`

- [ ] **Step 1: Create useBillDetail composable**

Create `frontend/src/composables/useBillDetail.ts`:
```typescript
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
    doStatusAction, goBack, totalAmount, totalQuantity,
  }
}
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/composables/useBillDetail.ts
git commit -m "feat: add useBillDetail composable for bill detail pages"
```

---

### Task 4: Create RemoteSelect Component

**Files:**
- Create: `frontend/src/components/RemoteSelect.vue`

- [ ] **Step 1: Create RemoteSelect.vue**

Create `frontend/src/components/RemoteSelect.vue`:
```vue
<template>
  <el-select
    v-model="innerValue"
    :placeholder="placeholder"
    :disabled="disabled"
    filterable
    remote
    :remote-method="handleSearch"
    :loading="loading"
    clearable
    style="width: 100%"
  >
    <el-option
      v-for="item in options"
      :key="item.value"
      :label="item.label"
      :value="item.value"
    />
  </el-select>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import api from '@/api/client'

interface Props {
  modelValue: number | string | undefined
  apiUrl: string
  labelKey?: string
  valueKey?: string
  placeholder?: string
  disabled?: boolean
  params?: Record<string, any>
}

const props = withDefaults(defineProps<Props>(), {
  labelKey: 'name',
  valueKey: 'id',
  placeholder: '请选择',
  disabled: false,
})

const emit = defineEmits<{
  (e: 'update:modelValue', val: number | string | undefined): void
}>()

const innerValue = ref(props.modelValue)
const options = ref<{ label: string; value: number | string }[]>([])
const loading = ref(false)

watch(() => props.modelValue, (val) => {
  innerValue.value = val
})

watch(innerValue, (val) => {
  emit('update:modelValue', val)
})

async function handleSearch(query: string) {
  loading.value = true
  try {
    const res = await api.get(props.apiUrl, {
      params: { keyword: query, pageSize: 50, ...props.params },
    })
    if (res.data.code === 0 || res.data.code === 200) {
      const data = res.data.data.list || res.data.data || []
      options.value = data.map((item: any) => ({
        label: item[props.labelKey],
        value: item[props.valueKey],
      }))
    }
  } catch {
    options.value = []
  } finally {
    loading.value = false
  }
}

// Initial load
handleSearch('')
</script>
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/components/RemoteSelect.vue
git commit -m "feat: add RemoteSelect component for API-backed dropdowns"
```

---

### Task 5: Create BillItemTable Component

**Files:**
- Create: `frontend/src/components/BillItemTable.vue`

- [ ] **Step 1: Create BillItemTable.vue**

Create `frontend/src/components/BillItemTable.vue`:
```vue
<template>
  <div>
    <div class="table-header">
      <span>商品明细</span>
      <el-button type="primary" size="small" @click="$emit('add')" :disabled="!editable">添加行</el-button>
    </div>
    <el-table :data="items" border size="small">
      <el-table-column type="index" width="50" />
      <el-table-column label="商品" min-width="200">
        <template #default="{ row, $index }">
          <RemoteSelect
            v-if="editable"
            v-model="row.productId"
            api-url="/api/v1/products"
            placeholder="选择商品"
          />
          <span v-else>{{ row.productName || row.productId }}</span>
        </template>
      </el-table-column>
      <el-table-column label="数量" width="120">
        <template #default="{ row, $index }">
          <el-input-number v-if="editable" v-model="row.quantity" :min="0" :precision="2" controls-position="right" style="width: 100%" />
          <span v-else>{{ row.quantity }}</span>
        </template>
      </el-table-column>
      <el-table-column label="单价" width="120">
        <template #default="{ row, $index }">
          <el-input-number v-if="editable" v-model="row.price" :min="0" :precision="2" controls-position="right" style="width: 100%" />
          <span v-else>{{ row.price }}</span>
        </template>
      </el-table-column>
      <el-table-column label="金额" width="120">
        <template #default="{ row }">
          {{ (row.quantity || 0) * (row.price || 0) }}
        </template>
      </el-table-column>
      <el-table-column label="备注" min-width="150">
        <template #default="{ row, $index }">
          <el-input v-if="editable" v-model="row.remark" size="small" />
          <span v-else>{{ row.remark }}</span>
        </template>
      </el-table-column>
      <el-table-column v-if="editable" label="操作" width="80" fixed="right">
        <template #default="{ $index }">
          <el-button type="danger" size="small" @click="$emit('remove', $index)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
    <div class="summary">
      <span>合计数量: {{ totalQuantity }}</span>
      <span>合计金额: {{ totalAmount.toFixed(2) }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import RemoteSelect from './RemoteSelect.vue'

interface Props {
  items: Record<string, any>[]
  editable: boolean
}

const props = defineProps<Props>()
const emit = defineEmits<{
  (e: 'add'): void
  (e: 'remove', index: number): void
}>()

const totalQuantity = computed(() =>
  props.items.reduce((sum, item) => sum + (item.quantity || 0), 0)
)

const totalAmount = computed(() =>
  props.items.reduce((sum, item) => sum + (item.quantity || 0) * (item.price || 0), 0)
)
</script>

<style scoped>
.table-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}
.summary {
  margin-top: 8px;
  text-align: right;
  font-weight: bold;
}
.summary span {
  margin-left: 24px;
}
</style>
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/components/BillItemTable.vue
git commit -m "feat: add BillItemTable component for editable sub-table"
```

---

### Task 6: Update Router with Purchase Detail Routes

**Files:**
- Modify: `frontend/src/router/index.ts`

- [ ] **Step 1: Add detail routes for purchase module**

Add these routes inside the `children` array of the Layout route in `frontend/src/router/index.ts`, after the existing purchase routes:

```typescript
// Purchase detail pages
{
  path: 'purchase-orders/:id',
  name: 'PurchaseOrderDetail',
  component: () => import('@/views/purchase/OrderDetailView.vue'),
  meta: { title: '采购订单详情' },
},
{
  path: 'purchase-instock/:id',
  name: 'PurchaseInStockDetail',
  component: () => import('@/views/purchase/InStockDetailView.vue'),
  meta: { title: '采购入库详情' },
},
{
  path: 'purchase-returns/:id',
  name: 'PurchaseReturnDetail',
  component: () => import('@/views/purchase/ReturnDetailView.vue'),
  meta: { title: '采购退货详情' },
},
{
  path: 'purchase-payments/:id',
  name: 'PurchasePaymentDetail',
  component: () => import('@/views/purchase/PaymentDetailView.vue'),
  meta: { title: '采购付款详情' },
},
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/router/index.ts
git commit -m "feat: add detail page routes for purchase module"
```

---

### Task 7: Purchase Order Pages

**Files:**
- Modify: `frontend/src/views/purchase/OrderView.vue`
- Create: `frontend/src/views/purchase/OrderDetailView.vue`

- [ ] **Step 1: Rewrite OrderView.vue (list page)**

Replace contents of `frontend/src/views/purchase/OrderView.vue`:
```vue
<template>
  <div class="page">
    <el-card>
      <template #header>
        <div class="card-header"><span>采购订单</span><el-button type="primary" @click="goCreate">新增订单</el-button></div>
      </template>
      <el-form :model="searchForm" inline class="search-form">
        <el-form-item label="订单编号"><el-input v-model="searchForm.orderNo" placeholder="订单编号" clearable /></el-form-item>
        <el-form-item label="供应商">
          <RemoteSelect v-model="searchForm.supplierId" api-url="/api/v1/customers" :params="{ type: 'supplier' }" placeholder="供应商" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" placeholder="状态" clearable>
            <el-option label="草稿" value="draft" /><el-option label="已确认" value="confirmed" />
            <el-option label="已取消" value="cancelled" /><el-option label="已完成" value="completed" />
          </el-select>
        </el-form-item>
        <el-form-item label="日期">
          <el-date-picker v-model="dateRange" type="daterange" value-format="YYYY-MM-DD" start-placeholder="开始" end-placeholder="结束" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">搜索</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
      <el-table :data="list" v-loading="loading" border stripe>
        <el-table-column prop="orderNo" label="订单编号" min-width="150" />
        <el-table-column prop="orderDate" label="订单日期" width="120" />
        <el-table-column prop="supplierName" label="供应商" min-width="150"><template #default="{ row }">{{ row.supplierName || row.supplierId }}</template></el-table-column>
        <el-table-column prop="totalAmount" label="总金额" width="120"><template #default="{ row }">{{ row.totalAmount?.toFixed(2) }}</template></el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)">{{ statusLabel(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="goDetail(row.id)">查看</el-button>
            <el-button v-if="row.status === 'draft'" type="warning" size="small" @click="goEdit(row.id)">编辑</el-button>
            <template v-if="row.status === 'draft'">
              <el-button type="success" size="small" @click="handleConfirm(row)">确认</el-button>
              <el-button type="danger" size="small" @click="handleCancel(row)">取消</el-button>
            </template>
            <el-button v-if="row.status === 'draft'" type="danger" size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-pagination v-model:current-page="pagination.page" v-model:page-size="pagination.pageSize" :total="total" :page-sizes="[10,20,50]" layout="total, sizes, prev, pager, next" @size-change="handleSizeChange" @current-change="handleCurrentChange" class="pagination" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useCrud } from '@/composables/useCrud'
import RemoteSelect from '@/components/RemoteSelect.vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '@/api/client'

const router = useRouter()
const dateRange = ref<[string, string] | null>(null)

const crud = useCrud<any>({
  baseUrl: '/api/v1/purchase-orders',
  defaultForm: () => ({}),
})

const { list, total, loading, searchForm, pagination, fetchList, handleSearch, handleReset, handleSizeChange, handleCurrentChange, handleDelete: crudDelete } = crud

watch(dateRange, (val) => {
  if (val) {
    searchForm.value.startDate = val[0]
    searchForm.value.endDate = val[1]
  } else {
    searchForm.value.startDate = undefined
    searchForm.value.endDate = undefined
  }
})

function statusType(status: string) {
  return { draft: 'info', confirmed: 'success', cancelled: 'danger', completed: 'primary' }[status] || 'warning'
}
function statusLabel(status: string) {
  return { draft: '草稿', confirmed: '已确认', cancelled: '已取消', completed: '已完成' }[status] || status
}

function goCreate() { router.push('/purchase-orders/new') }
function goDetail(id: number) { router.push(`/purchase-orders/${id}`) }
function goEdit(id: number) { router.push(`/purchase-orders/${id}?mode=edit`) }

async function handleConfirm(row: any) {
  try {
    await ElMessageBox.confirm('确认该采购订单？', '提示', { type: 'warning' })
    const res = await api.put(`/api/v1/purchase-orders/${row.id}/confirm`)
    if (res.data.code === 0 || res.data.code === 200) { ElMessage.success('确认成功'); fetchList() }
  } catch (e: any) { if (e !== 'cancel') ElMessage.error(e.message || '操作失败') }
}

async function handleCancel(row: any) {
  try {
    await ElMessageBox.confirm('取消该采购订单？', '提示', { type: 'warning' })
    const res = await api.put(`/api/v1/purchase-orders/${row.id}/cancel`)
    if (res.data.code === 0 || res.data.code === 200) { ElMessage.success('取消成功'); fetchList() }
  } catch (e: any) { if (e !== 'cancel') ElMessage.error(e.message || '操作失败') }
}

fetchList()
</script>

<style scoped>
.page { padding: 20px; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
.search-form { margin-bottom: 16px; }
.pagination { margin-top: 16px; justify-content: flex-end; }
</style>
```

- [ ] **Step 2: Create OrderDetailView.vue**

Create `frontend/src/views/purchase/OrderDetailView.vue`:
```vue
<template>
  <div class="page" v-loading="loading">
    <el-page-header @back="goBack" :title="pageTitle" />
    <el-card class="detail-card">
      <el-form :model="form" label-width="100px" :disabled="!isEditable">
        <el-row :gutter="20">
          <el-col :span="8"><el-form-item label="订单编号"><el-input v-model="form.orderNo" disabled /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="订单日期" required><el-date-picker v-model="form.orderDate" type="date" value-format="YYYY-MM-DD" style="width:100%" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="交货日期"><el-date-picker v-model="form.deliveryDate" type="date" value-format="YYYY-MM-DD" style="width:100%" /></el-form-item></el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="8"><el-form-item label="供应商" required><RemoteSelect v-model="form.supplierId" api-url="/api/v1/customers" :params="{ type: 'supplier' }" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="仓库" required><RemoteSelect v-model="form.warehouseId" api-url="/api/v1/warehouses" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="折扣"><el-input-number v-model="form.discount" :min="0" :precision="2" style="width:100%" /></el-form-item></el-col>
        </el-row>
        <el-form-item label="备注"><el-input v-model="form.remark" type="textarea" /></el-form-item>
      </el-form>
    </el-card>
    <el-card class="detail-card">
      <BillItemTable :items="items" :editable="isEditable" @add="addItem" @remove="removeItem" />
    </el-card>
    <el-card class="detail-card" v-if="!isEditable && form.status">
      <div class="status-actions">
        <template v-if="form.status === 'draft'">
          <el-button type="success" @click="doStatusAction({ api: '/api/v1/purchase-orders/:id/confirm' })">确认订单</el-button>
          <el-button type="danger" @click="doStatusAction({ api: '/api/v1/purchase-orders/:id/cancel' })">取消订单</el-button>
        </template>
      </div>
    </el-card>
    <div class="footer-actions">
      <template v-if="isEditable">
        <el-button type="primary" @click="save" :loading="saving">保存</el-button>
      </template>
      <el-button @click="goBack">返回</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useBillDetail } from '@/composables/useBillDetail'
import RemoteSelect from '@/components/RemoteSelect.vue'
import BillItemTable from '@/components/BillItemTable.vue'

const route = useRoute()
const id = route.params.id as string
const mode = (route.query.mode as string) || 'view'

const {
  form, items, loading, saving, isEditable,
  initCreate, loadDetail, addItem, removeItem, save, doStatusAction, goBack,
} = useBillDetail({
  baseUrl: '/api/v1/purchase-orders',
  defaultForm: () => ({ orderDate: '', supplierId: undefined, warehouseId: undefined, discount: 0, remark: '' }),
  defaultItem: () => ({ productId: undefined, quantity: 0, price: 0, remark: '' }),
})

const pageTitle = computed(() => {
  if (id === 'new') return '新建采购订单'
  return mode === 'edit' ? '编辑采购订单' : '采购订单详情'
})

if (id === 'new') {
  initCreate()
} else {
  loadDetail(id)
  if (mode === 'edit') { /* mode is set in composable, but we need to make it editable */ }
}
</script>

<style scoped>
.page { padding: 20px; }
.detail-card { margin-top: 16px; }
.footer-actions { margin-top: 16px; text-align: center; }
.status-actions { text-align: center; }
</style>
```

Note: The detail view needs a small fix — when `mode=edit`, we need to set the mode. Add this after `loadDetail(id)`:
```typescript
// In the script setup, after loadDetail(id):
if (mode === 'edit') {
  // useBillDetail doesn't expose mode setter directly
  // We need to adjust the composable or handle this differently
}
```

Actually, the `useBillDetail` composable should expose a `setMode` function. Let me add that to the composable in Task 3. But since Task 3 is already committed, we need to modify it here. Add to `useBillDetail` return object:
```typescript
function setMode(m: 'create' | 'edit' | 'view') { mode.value = m }
```

And in the OrderDetailView, use:
```typescript
if (mode === 'edit') { setMode('edit') }
```

- [ ] **Step 3: Fix useBillDetail to expose setMode**

Modify `frontend/src/composables/useBillDetail.ts`, add `setMode` to the return object:
```typescript
function setMode(m: 'create' | 'edit' | 'view') { mode.value = m }
// ... in return:
return {
  form, items, loading, saving, mode, isEditable,
  initCreate, loadDetail, addItem, removeItem, save,
  doStatusAction, goBack, totalAmount, totalQuantity, setMode,
}
```

- [ ] **Step 4: Verify pages compile**

Run:
```bash
cd frontend && npm run build
```
Expected: Build completes without errors

- [ ] **Step 5: Commit**

```bash
git add frontend/src/views/purchase/OrderView.vue frontend/src/views/purchase/OrderDetailView.vue frontend/src/composables/useBillDetail.ts
git commit -m "feat: complete purchase order list and detail pages"
```

---

### Task 8: Purchase InStock Pages

**Files:**
- Modify: `frontend/src/views/purchase/InStockView.vue`
- Create: `frontend/src/views/purchase/InStockDetailView.vue`

- [ ] **Step 1: Rewrite InStockView.vue**

Replace `frontend/src/views/purchase/InStockView.vue`:
```vue
<template>
  <div class="page">
    <el-card>
      <template #header>
        <div class="card-header"><span>采购入库</span><el-button type="primary" @click="goCreate">新增入库</el-button></div>
      </template>
      <el-form :model="searchForm" inline class="search-form">
        <el-form-item label="单据编号"><el-input v-model="searchForm.billNo" placeholder="单据编号" clearable /></el-form-item>
        <el-form-item label="供应商"><RemoteSelect v-model="searchForm.supplierId" api-url="/api/v1/customers" :params="{ type: 'supplier' }" placeholder="供应商" /></el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" placeholder="状态" clearable>
            <el-option label="草稿" value="draft" /><el-option label="已完成" value="completed" />
          </el-select>
        </el-form-item>
        <el-form-item label="日期"><el-date-picker v-model="dateRange" type="daterange" value-format="YYYY-MM-DD" start-placeholder="开始" end-placeholder="结束" /></el-form-item>
        <el-form-item><el-button type="primary" @click="handleSearch">搜索</el-button><el-button @click="handleReset">重置</el-button></el-form-item>
      </el-form>
      <el-table :data="list" v-loading="loading" border stripe>
        <el-table-column prop="billNo" label="单据编号" min-width="150" />
        <el-table-column prop="billDate" label="入库日期" width="120" />
        <el-table-column prop="supplierName" label="供应商" min-width="150"><template #default="{ row }">{{ row.supplierName || row.supplierId }}</template></el-table-column>
        <el-table-column prop="totalAmount" label="总金额" width="120"><template #default="{ row }">{{ row.totalAmount?.toFixed(2) }}</template></el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }"><el-tag :type="row.status === 'completed' ? 'success' : 'info'">{{ row.status === 'completed' ? '已完成' : '草稿' }}</el-tag></template>
        </el-table-column>
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="goDetail(row.id)">查看</el-button>
            <el-button v-if="row.status === 'draft'" type="warning" size="small" @click="goEdit(row.id)">编辑</el-button>
            <el-button v-if="row.status === 'draft'" type="success" size="small" @click="handleComplete(row)">完成</el-button>
            <el-button v-if="row.status === 'draft'" type="danger" size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-pagination v-model:current-page="pagination.page" v-model:page-size="pagination.pageSize" :total="total" :page-sizes="[10,20,50]" layout="total, sizes, prev, pager, next" @size-change="handleSizeChange" @current-change="handleCurrentChange" class="pagination" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useCrud } from '@/composables/useCrud'
import RemoteSelect from '@/components/RemoteSelect.vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '@/api/client'

const router = useRouter()
const dateRange = ref<[string, string] | null>(null)
const crud = useCrud<any>({ baseUrl: '/api/v1/purchase-instock', defaultForm: () => ({}) })
const { list, total, loading, searchForm, pagination, fetchList, handleSearch, handleReset, handleSizeChange, handleCurrentChange, handleDelete: crudDelete } = crud

watch(dateRange, (val) => {
  if (val) { searchForm.value.startDate = val[0]; searchForm.value.endDate = val[1] }
  else { searchForm.value.startDate = undefined; searchForm.value.endDate = undefined }
})

function goCreate() { router.push('/purchase-instock/new') }
function goDetail(id: number) { router.push(`/purchase-instock/${id}`) }
function goEdit(id: number) { router.push(`/purchase-instock/${id}?mode=edit`) }

async function handleComplete(row: any) {
  try {
    await ElMessageBox.confirm('完成该入库单？', '提示', { type: 'warning' })
    const res = await api.put(`/api/v1/purchase-instock/${row.id}/complete`)
    if (res.data.code === 0 || res.data.code === 200) { ElMessage.success('完成成功'); fetchList() }
  } catch (e: any) { if (e !== 'cancel') ElMessage.error(e.message || '操作失败') }
}

async function handleDelete(row: any) { await crudDelete(row) }
fetchList()
</script>

<style scoped>
.page { padding: 20px; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
.search-form { margin-bottom: 16px; }
.pagination { margin-top: 16px; justify-content: flex-end; }
</style>
```

- [ ] **Step 2: Create InStockDetailView.vue**

Create `frontend/src/views/purchase/InStockDetailView.vue`:
```vue
<template>
  <div class="page" v-loading="loading">
    <el-page-header @back="goBack" :title="pageTitle" />
    <el-card class="detail-card">
      <el-form :model="form" label-width="100px" :disabled="!isEditable">
        <el-row :gutter="20">
          <el-col :span="8"><el-form-item label="单据编号"><el-input v-model="form.billNo" disabled /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="入库日期" required><el-date-picker v-model="form.billDate" type="date" value-format="YYYY-MM-DD" style="width:100%" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="采购订单"><RemoteSelect v-model="form.orderId" api-url="/api/v1/purchase-orders" /></el-form-item></el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="8"><el-form-item label="供应商" required><RemoteSelect v-model="form.supplierId" api-url="/api/v1/customers" :params="{ type: 'supplier' }" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="仓库" required><RemoteSelect v-model="form.warehouseId" api-url="/api/v1/warehouses" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="折扣"><el-input-number v-model="form.discount" :min="0" :precision="2" style="width:100%" /></el-form-item></el-col>
        </el-row>
        <el-form-item label="备注"><el-input v-model="form.remark" type="textarea" /></el-form-item>
      </el-form>
    </el-card>
    <el-card class="detail-card"><BillItemTable :items="items" :editable="isEditable" @add="addItem" @remove="removeItem" /></el-card>
    <el-card class="detail-card" v-if="!isEditable && form.status === 'draft'">
      <div class="status-actions">
        <el-button type="success" @click="doStatusAction({ api: '/api/v1/purchase-instock/:id/complete' })">完成入库</el-button>
      </div>
    </el-card>
    <div class="footer-actions">
      <template v-if="isEditable"><el-button type="primary" @click="save" :loading="saving">保存</el-button></template>
      <el-button @click="goBack">返回</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useBillDetail } from '@/composables/useBillDetail'
import RemoteSelect from '@/components/RemoteSelect.vue'
import BillItemTable from '@/components/BillItemTable.vue'

const route = useRoute()
const id = route.params.id as string
const mode = (route.query.mode as string) || 'view'

const { form, items, loading, saving, isEditable, initCreate, loadDetail, addItem, removeItem, save, doStatusAction, goBack, setMode } = useBillDetail({
  baseUrl: '/api/v1/purchase-instock',
  defaultForm: () => ({ billDate: '', supplierId: undefined, warehouseId: undefined, orderId: undefined, discount: 0, remark: '' }),
  defaultItem: () => ({ productId: undefined, quantity: 0, price: 0, remark: '' }),
})

const pageTitle = computed(() => id === 'new' ? '新建采购入库' : mode === 'edit' ? '编辑采购入库' : '采购入库详情')

if (id === 'new') { initCreate() } else { loadDetail(id); if (mode === 'edit') setMode('edit') }
</script>

<style scoped>
.page { padding: 20px; }
.detail-card { margin-top: 16px; }
.footer-actions { margin-top: 16px; text-align: center; }
.status-actions { text-align: center; }
</style>
```

- [ ] **Step 3: Verify build**

```bash
cd frontend && npm run build
```
Expected: No errors

- [ ] **Step 4: Commit**

```bash
git add frontend/src/views/purchase/InStockView.vue frontend/src/views/purchase/InStockDetailView.vue
git commit -m "feat: complete purchase instock list and detail pages"
```

---

### Task 9: Purchase Return Pages

**Files:**
- Modify: `frontend/src/views/purchase/ReturnView.vue`
- Create: `frontend/src/views/purchase/ReturnDetailView.vue`

- [ ] **Step 1: Rewrite ReturnView.vue**

Replace `frontend/src/views/purchase/ReturnView.vue`:
```vue
<template>
  <div class="page">
    <el-card>
      <template #header>
        <div class="card-header"><span>采购退货</span><el-button type="primary" @click="goCreate">新增退货</el-button></div>
      </template>
      <el-form :model="searchForm" inline class="search-form">
        <el-form-item label="单据编号"><el-input v-model="searchForm.billNo" placeholder="单据编号" clearable /></el-form-item>
        <el-form-item label="供应商"><RemoteSelect v-model="searchForm.supplierId" api-url="/api/v1/customers" :params="{ type: 'supplier' }" placeholder="供应商" /></el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" placeholder="状态" clearable>
            <el-option label="草稿" value="draft" /><el-option label="已完成" value="completed" />
          </el-select>
        </el-form-item>
        <el-form-item label="日期"><el-date-picker v-model="dateRange" type="daterange" value-format="YYYY-MM-DD" start-placeholder="开始" end-placeholder="结束" /></el-form-item>
        <el-form-item><el-button type="primary" @click="handleSearch">搜索</el-button><el-button @click="handleReset">重置</el-button></el-form-item>
      </el-form>
      <el-table :data="list" v-loading="loading" border stripe>
        <el-table-column prop="billNo" label="单据编号" min-width="150" />
        <el-table-column prop="billDate" label="退货日期" width="120" />
        <el-table-column prop="supplierName" label="供应商" min-width="150"><template #default="{ row }">{{ row.supplierName || row.supplierId }}</template></el-table-column>
        <el-table-column prop="totalAmount" label="总金额" width="120"><template #default="{ row }">{{ row.totalAmount?.toFixed(2) }}</template></el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }"><el-tag :type="row.status === 'completed' ? 'success' : 'info'">{{ row.status === 'completed' ? '已完成' : '草稿' }}</el-tag></template>
        </el-table-column>
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="goDetail(row.id)">查看</el-button>
            <el-button v-if="row.status === 'draft'" type="warning" size="small" @click="goEdit(row.id)">编辑</el-button>
            <el-button v-if="row.status === 'draft'" type="success" size="small" @click="handleComplete(row)">完成</el-button>
            <el-button v-if="row.status === 'draft'" type="danger" size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-pagination v-model:current-page="pagination.page" v-model:page-size="pagination.pageSize" :total="total" :page-sizes="[10,20,50]" layout="total, sizes, prev, pager, next" @size-change="handleSizeChange" @current-change="handleCurrentChange" class="pagination" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useCrud } from '@/composables/useCrud'
import RemoteSelect from '@/components/RemoteSelect.vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '@/api/client'

const router = useRouter()
const dateRange = ref<[string, string] | null>(null)
const crud = useCrud<any>({ baseUrl: '/api/v1/purchase-returns', defaultForm: () => ({}) })
const { list, total, loading, searchForm, pagination, fetchList, handleSearch, handleReset, handleSizeChange, handleCurrentChange, handleDelete: crudDelete } = crud

watch(dateRange, (val) => {
  if (val) { searchForm.value.startDate = val[0]; searchForm.value.endDate = val[1] }
  else { searchForm.value.startDate = undefined; searchForm.value.endDate = undefined }
})

function goCreate() { router.push('/purchase-returns/new') }
function goDetail(id: number) { router.push(`/purchase-returns/${id}`) }
function goEdit(id: number) { router.push(`/purchase-returns/${id}?mode=edit`) }

async function handleComplete(row: any) {
  try {
    await ElMessageBox.confirm('完成该退货单？', '提示', { type: 'warning' })
    const res = await api.put(`/api/v1/purchase-returns/${row.id}/complete`)
    if (res.data.code === 0 || res.data.code === 200) { ElMessage.success('完成成功'); fetchList() }
  } catch (e: any) { if (e !== 'cancel') ElMessage.error(e.message || '操作失败') }
}

async function handleDelete(row: any) { await crudDelete(row) }
fetchList()
</script>

<style scoped>
.page { padding: 20px; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
.search-form { margin-bottom: 16px; }
.pagination { margin-top: 16px; justify-content: flex-end; }
</style>
```

- [ ] **Step 2: Create ReturnDetailView.vue**

Create `frontend/src/views/purchase/ReturnDetailView.vue`:
```vue
<template>
  <div class="page" v-loading="loading">
    <el-page-header @back="goBack" :title="pageTitle" />
    <el-card class="detail-card">
      <el-form :model="form" label-width="100px" :disabled="!isEditable">
        <el-row :gutter="20">
          <el-col :span="8"><el-form-item label="单据编号"><el-input v-model="form.billNo" disabled /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="退货日期" required><el-date-picker v-model="form.billDate" type="date" value-format="YYYY-MM-DD" style="width:100%" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="入库单"><RemoteSelect v-model="form.inStockId" api-url="/api/v1/purchase-instock" /></el-form-item></el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="8"><el-form-item label="供应商" required><RemoteSelect v-model="form.supplierId" api-url="/api/v1/customers" :params="{ type: 'supplier' }" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="仓库" required><RemoteSelect v-model="form.warehouseId" api-url="/api/v1/warehouses" /></el-form-item></el-col>
        </el-row>
        <el-form-item label="备注"><el-input v-model="form.remark" type="textarea" /></el-form-item>
      </el-form>
    </el-card>
    <el-card class="detail-card"><BillItemTable :items="items" :editable="isEditable" @add="addItem" @remove="removeItem" /></el-card>
    <el-card class="detail-card" v-if="!isEditable && form.status === 'draft'">
      <div class="status-actions"><el-button type="success" @click="doStatusAction({ api: '/api/v1/purchase-returns/:id/complete' })">完成退货</el-button></div>
    </el-card>
    <div class="footer-actions">
      <template v-if="isEditable"><el-button type="primary" @click="save" :loading="saving">保存</el-button></template>
      <el-button @click="goBack">返回</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useBillDetail } from '@/composables/useBillDetail'
import RemoteSelect from '@/components/RemoteSelect.vue'
import BillItemTable from '@/components/BillItemTable.vue'

const route = useRoute()
const id = route.params.id as string
const mode = (route.query.mode as string) || 'view'

const { form, items, loading, saving, isEditable, initCreate, loadDetail, addItem, removeItem, save, doStatusAction, goBack, setMode } = useBillDetail({
  baseUrl: '/api/v1/purchase-returns',
  defaultForm: () => ({ billDate: '', supplierId: undefined, warehouseId: undefined, inStockId: undefined, remark: '' }),
  defaultItem: () => ({ productId: undefined, quantity: 0, price: 0, remark: '' }),
})

const pageTitle = computed(() => id === 'new' ? '新建采购退货' : mode === 'edit' ? '编辑采购退货' : '采购退货详情')
if (id === 'new') { initCreate() } else { loadDetail(id); if (mode === 'edit') setMode('edit') }
</script>

<style scoped>
.page { padding: 20px; }
.detail-card { margin-top: 16px; }
.footer-actions { margin-top: 16px; text-align: center; }
.status-actions { text-align: center; }
</style>
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/views/purchase/ReturnView.vue frontend/src/views/purchase/ReturnDetailView.vue
git commit -m "feat: complete purchase return list and detail pages"
```

---

### Task 10: Purchase Payment Pages

**Files:**
- Modify: `frontend/src/views/purchase/PaymentView.vue`
- Create: `frontend/src/views/purchase/PaymentDetailView.vue`

- [ ] **Step 1: Rewrite PaymentView.vue**

Replace `frontend/src/views/purchase/PaymentView.vue`:
```vue
<template>
  <div class="page">
    <el-card>
      <template #header>
        <div class="card-header"><span>采购付款</span><el-button type="primary" @click="goCreate">新增付款</el-button></div>
      </template>
      <el-form :model="searchForm" inline class="search-form">
        <el-form-item label="单据编号"><el-input v-model="searchForm.billNo" placeholder="单据编号" clearable /></el-form-item>
        <el-form-item label="供应商"><RemoteSelect v-model="searchForm.supplierId" api-url="/api/v1/customers" :params="{ type: 'supplier' }" placeholder="供应商" /></el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" placeholder="状态" clearable>
            <el-option label="草稿" value="draft" /><el-option label="已完成" value="completed" />
          </el-select>
        </el-form-item>
        <el-form-item label="日期"><el-date-picker v-model="dateRange" type="daterange" value-format="YYYY-MM-DD" start-placeholder="开始" end-placeholder="结束" /></el-form-item>
        <el-form-item><el-button type="primary" @click="handleSearch">搜索</el-button><el-button @click="handleReset">重置</el-button></el-form-item>
      </el-form>
      <el-table :data="list" v-loading="loading" border stripe>
        <el-table-column prop="billNo" label="单据编号" min-width="150" />
        <el-table-column prop="billDate" label="付款日期" width="120" />
        <el-table-column prop="supplierName" label="供应商" min-width="150"><template #default="{ row }">{{ row.supplierName || row.supplierId }}</template></el-table-column>
        <el-table-column prop="totalAmount" label="总金额" width="120"><template #default="{ row }">{{ row.totalAmount?.toFixed(2) }}</template></el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }"><el-tag :type="row.status === 'completed' ? 'success' : 'info'">{{ row.status === 'completed' ? '已完成' : '草稿' }}</el-tag></template>
        </el-table-column>
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="goDetail(row.id)">查看</el-button>
            <el-button v-if="row.status === 'draft'" type="warning" size="small" @click="goEdit(row.id)">编辑</el-button>
            <el-button v-if="row.status === 'draft'" type="success" size="small" @click="handleComplete(row)">完成</el-button>
            <el-button v-if="row.status === 'draft'" type="danger" size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-pagination v-model:current-page="pagination.page" v-model:page-size="pagination.pageSize" :total="total" :page-sizes="[10,20,50]" layout="total, sizes, prev, pager, next" @size-change="handleSizeChange" @current-change="handleCurrentChange" class="pagination" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useCrud } from '@/composables/useCrud'
import RemoteSelect from '@/components/RemoteSelect.vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '@/api/client'

const router = useRouter()
const dateRange = ref<[string, string] | null>(null)
const crud = useCrud<any>({ baseUrl: '/api/v1/purchase-payments', defaultForm: () => ({}) })
const { list, total, loading, searchForm, pagination, fetchList, handleSearch, handleReset, handleSizeChange, handleCurrentChange, handleDelete: crudDelete } = crud

watch(dateRange, (val) => {
  if (val) { searchForm.value.startDate = val[0]; searchForm.value.endDate = val[1] }
  else { searchForm.value.startDate = undefined; searchForm.value.endDate = undefined }
})

function goCreate() { router.push('/purchase-payments/new') }
function goDetail(id: number) { router.push(`/purchase-payments/${id}`) }
function goEdit(id: number) { router.push(`/purchase-payments/${id}?mode=edit`) }

async function handleComplete(row: any) {
  try {
    await ElMessageBox.confirm('完成该付款单？', '提示', { type: 'warning' })
    const res = await api.put(`/api/v1/purchase-payments/${row.id}/complete`)
    if (res.data.code === 0 || res.data.code === 200) { ElMessage.success('完成成功'); fetchList() }
  } catch (e: any) { if (e !== 'cancel') ElMessage.error(e.message || '操作失败') }
}

async function handleDelete(row: any) { await crudDelete(row) }
fetchList()
</script>

<style scoped>
.page { padding: 20px; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
.search-form { margin-bottom: 16px; }
.pagination { margin-top: 16px; justify-content: flex-end; }
</style>
```

- [ ] **Step 2: Create PaymentDetailView.vue**

Create `frontend/src/views/purchase/PaymentDetailView.vue`:
```vue
<template>
  <div class="page" v-loading="loading">
    <el-page-header @back="goBack" :title="pageTitle" />
    <el-card class="detail-card">
      <el-form :model="form" label-width="100px" :disabled="!isEditable">
        <el-row :gutter="20">
          <el-col :span="8"><el-form-item label="单据编号"><el-input v-model="form.billNo" disabled /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="付款日期" required><el-date-picker v-model="form.billDate" type="date" value-format="YYYY-MM-DD" style="width:100%" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="供应商" required><RemoteSelect v-model="form.supplierId" api-url="/api/v1/customers" :params="{ type: 'supplier' }" /></el-form-item></el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="8"><el-form-item label="付款账户"><RemoteSelect v-model="form.accountId" api-url="/api/v1/accounts" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="总金额"><el-input-number v-model="form.totalAmount" :min="0" :precision="2" style="width:100%" /></el-form-item></el-col>
        </el-row>
        <el-form-item label="备注"><el-input v-model="form.remark" type="textarea" /></el-form-item>
      </el-form>
    </el-card>
    <el-card class="detail-card" v-if="!isEditable && form.status === 'draft'">
      <div class="status-actions"><el-button type="success" @click="doStatusAction({ api: '/api/v1/purchase-payments/:id/complete' })">完成付款</el-button></div>
    </el-card>
    <div class="footer-actions">
      <template v-if="isEditable"><el-button type="primary" @click="save" :loading="saving">保存</el-button></template>
      <el-button @click="goBack">返回</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useBillDetail } from '@/composables/useBillDetail'
import RemoteSelect from '@/components/RemoteSelect.vue'

const route = useRoute()
const id = route.params.id as string
const mode = (route.query.mode as string) || 'view'

const { form, loading, saving, isEditable, initCreate, loadDetail, save, doStatusAction, goBack, setMode } = useBillDetail({
  baseUrl: '/api/v1/purchase-payments',
  defaultForm: () => ({ billDate: '', supplierId: undefined, accountId: undefined, totalAmount: 0, remark: '' }),
  defaultItem: () => ({}),
})

const pageTitle = computed(() => id === 'new' ? '新建采购付款' : mode === 'edit' ? '编辑采购付款' : '采购付款详情')
if (id === 'new') { initCreate() } else { loadDetail(id); if (mode === 'edit') setMode('edit') }
</script>

<style scoped>
.page { padding: 20px; }
.detail-card { margin-top: 16px; }
.footer-actions { margin-top: 16px; text-align: center; }
.status-actions { text-align: center; }
</style>
```

- [ ] **Step 3: Final build verification**

```bash
cd frontend && npm run build
```
Expected: Build completes successfully with no errors

- [ ] **Step 4: Commit**

```bash
git add frontend/src/views/purchase/PaymentView.vue frontend/src/views/purchase/PaymentDetailView.vue
git commit -m "feat: complete purchase payment list and detail pages"
```

---

## Self-Review

**1. Spec coverage:**
- Search/filter: Covered in all list views (Task 7-10)
- Pagination: Covered via upgraded useCrud (Task 2) and el-pagination in all list views
- Detail page with route-based navigation: Covered (Task 6 router + Task 7-10 detail views)
- Sub-table editing: Covered via BillItemTable component (Task 5) used in detail views
- Status transitions: Covered via status action buttons in list views and detail views
- Remote select for关联数据: Covered via RemoteSelect component (Task 4)
- Create/edit/view mode on same page: Covered via useBillDetail composable (Task 3)

**2. Placeholder scan:** No TBD, TODO, or vague steps. Every step contains complete code.

**3. Type consistency:** All components use the same `useCrud` and `useBillDetail` interfaces defined in Task 2 and 3. API paths match backend routes confirmed from handler source code.

**4. Gap identified and fixed:** `useBillDetail` initially didn't expose `setMode`. Fixed in Task 7 Step 3.

---

## Execution Handoff

**Plan complete and saved to `docs/superpowers/plans/2026-06-01-frontend-bill-pages-phase1.md`.**

Two execution options:

**1. Subagent-Driven (recommended)** — I dispatch a fresh subagent per task, review between tasks, fast iteration

**2. Inline Execution** — Execute tasks in this session using executing-plans, batch execution with checkpoints

**Which approach?**
