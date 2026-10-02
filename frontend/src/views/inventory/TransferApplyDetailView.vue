<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">调拨申请单</h2>
      <el-tag :type="statusType(form.status)">{{ statusLabel(form.status) }}</el-tag>
    </div>

    <div class="form-panel" v-loading="loading">
      <el-form :model="form" label-width="80px" :disabled="!editable" class="bill-form">
        <el-row :gutter="16">
          <el-col :span="6">
            <el-form-item label="调出仓库" required>
              <RemoteSelect v-model="form.fromWarehouseId" api-url="/api/v1/warehouses" placeholder="请选择" />
            </el-form-item>
          </el-col>
          <el-col :span="6">
            <el-form-item label="调入仓库" required>
              <RemoteSelect v-model="form.toWarehouseId" api-url="/api/v1/warehouses" placeholder="请选择" />
            </el-form-item>
          </el-col>
          <el-col :span="6">
            <el-form-item label="录单时间">
              <el-date-picker v-model="form.billDate" type="date" value-format="YYYY-MM-DD" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="6">
            <el-form-item label="经手人">
              <RemoteSelect v-model="form.handlerId" api-url="/api/v1/employees" placeholder="请选择" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row>
          <el-col :span="12">
            <el-form-item label="单据备注"><el-input v-model="form.remark" placeholder="请输入" /></el-form-item>
          </el-col>
        </el-row>
      </el-form>
    </div>

    <div class="items-panel">
      <div class="items-toolbar">
        <span class="items-title">商品明细</span>
        <el-button v-if="editable" type="primary" size="small" @click="addRow">添加行</el-button>
      </div>
      <el-table :data="items" border size="small">
        <el-table-column type="index" width="50" />
        <el-table-column label="商品名称" min-width="240">
          <template #default="{ row }">
            <el-select
              v-if="editable"
              v-model="row.productId" filterable remote :remote-method="searchProducts" :loading="productLoading"
              placeholder="请输入商品名称/编码/条码/规格" style="width: 100%"
              @focus="searchProducts('')" @change="(val: number) => handleProductChange(row, val)"
            >
              <el-option v-for="p in productOptions" :key="p.id" :label="`${p.name}${p.specification ? ' / ' + p.specification : ''}`" :value="p.id" />
            </el-select>
            <span v-else>{{ row.productName }}</span>
          </template>
        </el-table-column>
        <el-table-column label="商品规格" width="130">
          <template #default="{ row }">{{ row.specification || '-' }}</template>
        </el-table-column>
        <el-table-column label="单位" width="80">
          <template #default="{ row }">{{ row.unit || '-' }}</template>
        </el-table-column>
        <el-table-column label="数量" width="140">
          <template #default="{ row }">
            <el-input-number v-if="editable" v-model="row.quantity" :min="0" :precision="2" controls-position="right" style="width: 100%" />
            <span v-else>{{ row.quantity }}</span>
          </template>
        </el-table-column>
        <el-table-column label="备注" min-width="140">
          <template #default="{ row }">
            <el-input v-if="editable" v-model="row.remark" size="small" />
            <span v-else>{{ row.remark }}</span>
          </template>
        </el-table-column>
        <el-table-column v-if="editable" label="操作" width="70" fixed="right">
          <template #default="{ $index }">
            <el-button type="danger" link size="small" @click="items.splice($index, 1)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <div class="summary"><span>合计数量: {{ totalQty }}</span></div>
    </div>

    <div class="footer-bar">
      <span class="operator">制单人：{{ operatorName }}</span>
      <div class="footer-actions">
        <el-button @click="goBack">返 回</el-button>
        <template v-if="editable">
          <el-button :loading="saving" @click="save(false)">存入草稿</el-button>
          <el-button type="primary" :loading="saving" @click="save(true)">保存并提交审核</el-button>
        </template>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import dayjs from 'dayjs'
import { ElMessage } from 'element-plus'
import RemoteSelect from '@/components/RemoteSelect.vue'
import api from '@/api/client'
import { useAuthStore } from '@/stores/auth'

interface ItemRow { productId?: number; productName?: string; specification?: string; unit?: string; quantity: number; remark?: string }

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const billId = computed(() => (route.params.id && route.params.id !== 'new' ? Number(route.params.id) : null))
const editable = ref(true)
const loading = ref(false)
const saving = ref(false)

const form = reactive({
  id: 0,
  fromWarehouseId: undefined as number | undefined,
  toWarehouseId: undefined as number | undefined,
  billDate: dayjs().format('YYYY-MM-DD'),
  handlerId: authStore.user?.id as number | undefined,
  remark: '',
  status: 'draft',
})
const items = ref<ItemRow[]>([])
const operatorName = computed(() => authStore.user?.name || authStore.user?.username || '-')
const productOptions = ref<any[]>([])
const productLoading = ref(false)
const totalQty = computed(() => items.value.reduce((s, i) => s + (i.quantity || 0), 0))

const STATUS_MAP: Record<string, [string, string]> = {
  draft: ['待审核', 'info'], approved: ['待出库', 'warning'], shipped: ['待入库', 'warning'], completed: ['已完成', 'success'], cancelled: ['已取消', 'danger'],
}
function statusLabel(s: string) { return STATUS_MAP[s]?.[0] || s }
function statusType(s: string) { return (STATUS_MAP[s]?.[1] || 'info') as 'info' | 'warning' | 'success' | 'danger' }

async function searchProducts(keyword: string) {
  productLoading.value = true
  try {
    const res = await api.get('/api/v1/products', { params: { keyword, pageSize: 50 } })
    if (res.data.code === 0 || res.data.code === 200) productOptions.value = res.data.data?.list ?? res.data.data?.items ?? []
  } finally { productLoading.value = false }
}

function handleProductChange(row: ItemRow, productId: number) {
  const p = productOptions.value.find((x) => x.id === productId)
  if (p) { row.productName = p.name; row.specification = p.specification; row.unit = p.unit }
}

function addRow() { items.value.push({ quantity: 1 }) }

async function loadDetail(id: number) {
  loading.value = true
  try {
    const res = await api.get(`/api/v1/transfer-applies/${id}`)
    if (res.data.code === 0 || res.data.code === 200) {
      const data = res.data.data
      Object.assign(form, {
        id: data.id, fromWarehouseId: data.fromWarehouseId, toWarehouseId: data.toWarehouseId,
        billDate: data.billDate ? dayjs(data.billDate).format('YYYY-MM-DD') : dayjs().format('YYYY-MM-DD'),
        handlerId: data.handlerId || undefined, remark: data.remark, status: data.status,
      })
      items.value = (data.items ?? []).map((it: any) => ({
        productId: it.productId, productName: it.product?.name, specification: it.product?.specification,
        unit: it.product?.unit, quantity: it.quantity, remark: it.remark,
      }))
      if (data.status !== 'draft') editable.value = false
    }
  } finally { loading.value = false }
}

async function save(submit: boolean) {
  if (!form.fromWarehouseId || !form.toWarehouseId) { ElMessage.warning('请选择调出/调入仓库'); return }
  const validItems = items.value.filter((i) => i.productId)
  if (!validItems.length) { ElMessage.warning('请至少添加一行商品'); return }
  saving.value = true
  try {
    const payload = {
      fromWarehouseId: form.fromWarehouseId, toWarehouseId: form.toWarehouseId,
      billDate: form.billDate, handlerId: form.handlerId, remark: form.remark,
      items: validItems.map((i) => ({ productId: i.productId!, quantity: i.quantity, remark: i.remark })),
    }
    let id = billId.value
    const res = id ? await api.put(`/api/v1/transfer-applies/${id}`, payload) : await api.post('/api/v1/transfer-applies', payload)
    if (res.data.code === 0 || res.data.code === 200) {
      if (!id) id = res.data.data?.id
      if (submit && id) await api.put(`/api/v1/transfer-applies/${id}/approve`)
      ElMessage.success(submit ? '已提交审核' : '已存入草稿')
      router.back()
    }
  } finally { saving.value = false }
}

function goBack() { router.back() }

onMounted(() => { billId.value ? loadDetail(billId.value) : addRow() })
</script>

<style scoped>
.page { display: flex; flex-direction: column; height: 100%; gap: 12px; }
.page-header { display: flex; align-items: center; gap: 12px; }
.page-title { font-size: 18px; font-weight: 600; margin: 0; }
.form-panel { background: var(--el-bg-color); border-radius: 8px; padding: 16px 16px 0; }
.items-panel { background: var(--el-bg-color); border-radius: 8px; padding: 16px; flex: 1; min-height: 0; overflow: auto; }
.items-toolbar { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px; }
.items-title { font-weight: 600; }
.summary { margin-top: 8px; text-align: right; font-weight: bold; }
.footer-bar { display: flex; justify-content: space-between; align-items: center; background: var(--el-bg-color); border-radius: 8px; padding: 12px 16px; }
.operator { color: var(--el-text-color-secondary); }
.footer-actions { display: flex; gap: 8px; }
</style>
