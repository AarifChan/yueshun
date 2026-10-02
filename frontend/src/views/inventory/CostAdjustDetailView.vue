<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">成本调价单</h2>
      <el-tag v-if="form.status === 'completed'" type="success">已过账</el-tag>
      <el-tag v-else-if="form.id" type="info">草稿</el-tag>
    </div>

    <div class="form-panel" v-loading="loading">
      <el-form :model="form" label-width="80px" :disabled="!editable" class="bill-form">
        <el-row :gutter="16">
          <el-col :span="6">
            <el-form-item label="仓库" required>
              <RemoteSelect v-model="form.warehouseId" api-url="/api/v1/warehouses" placeholder="请选择" />
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
          <el-col :span="6">
            <el-form-item label="关联单据"><el-input v-model="form.refBillNo" placeholder="关联单据编号" clearable /></el-form-item>
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
        <span class="items-title">调价明细</span>
        <el-button v-if="editable" type="primary" size="small" @click="addRow">添加行</el-button>
      </div>
      <el-table :data="items" border size="small">
        <el-table-column type="index" width="50" />
        <el-table-column label="商品名称" min-width="220">
          <template #default="{ row }">
            <el-select
              v-if="editable"
              v-model="row.productId"
              filterable remote :remote-method="searchProducts" :loading="productLoading"
              placeholder="请输入商品名称/编码/条码/规格" style="width: 100%"
              @focus="searchProducts('')"
              @change="(val: number) => handleProductChange(row, val)"
            >
              <el-option v-for="p in productOptions" :key="p.id" :label="`${p.name}${p.specification ? ' / ' + p.specification : ''}`" :value="p.id" />
            </el-select>
            <span v-else>{{ row.productName }}</span>
          </template>
        </el-table-column>
        <el-table-column label="商品规格" width="120">
          <template #default="{ row }">{{ row.specification || '-' }}</template>
        </el-table-column>
        <el-table-column label="单位" width="80">
          <template #default="{ row }">{{ row.unit || '-' }}</template>
        </el-table-column>
        <el-table-column label="库存数量" width="100" align="right">
          <template #default="{ row }">{{ row.stockQty ?? '-' }}</template>
        </el-table-column>
        <el-table-column label="调价数量" width="130">
          <template #default="{ row }">
            <el-input-number v-if="editable" v-model="row.quantity" :min="0" :precision="2" controls-position="right" style="width: 100%" />
            <span v-else>{{ row.quantity }}</span>
          </template>
        </el-table-column>
        <el-table-column label="原成本价" width="130">
          <template #default="{ row }">
            <el-input-number v-if="editable" v-model="row.oldPrice" :min="0" :precision="4" controls-position="right" style="width: 100%" />
            <span v-else>{{ row.oldPrice }}</span>
          </template>
        </el-table-column>
        <el-table-column label="新成本价" width="130">
          <template #default="{ row }">
            <el-input-number v-if="editable" v-model="row.newPrice" :min="0" :precision="4" controls-position="right" style="width: 100%" />
            <span v-else>{{ row.newPrice }}</span>
          </template>
        </el-table-column>
        <el-table-column label="调价差额" width="110" align="right">
          <template #default="{ row }">{{ diffOf(row).toFixed(2) }}</template>
        </el-table-column>
        <el-table-column v-if="editable" label="操作" width="70" fixed="right">
          <template #default="{ $index }">
            <el-button type="danger" link size="small" @click="items.splice($index, 1)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <div class="summary"><span>调价差额合计: {{ totalDiff }}</span></div>
    </div>

    <div class="footer-bar">
      <span class="operator">制单人：{{ operatorName }}</span>
      <div class="footer-actions">
        <el-button @click="goBack">返 回</el-button>
        <template v-if="editable">
          <el-button :loading="saving" @click="save(false)">存入草稿</el-button>
          <el-button type="primary" :loading="saving" @click="save(true)">保存并过账</el-button>
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

interface ItemRow {
  productId?: number; productName?: string; specification?: string; unit?: string
  stockQty?: number; quantity: number; oldPrice: number; newPrice: number; remark?: string
}

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const billId = computed(() => {
  const id = route.params.id
  return id && id !== 'new' ? Number(id) : null
})
const editable = ref(true)
const loading = ref(false)
const saving = ref(false)

const form = reactive({
  id: 0,
  warehouseId: undefined as number | undefined,
  billDate: dayjs().format('YYYY-MM-DD'),
  handlerId: authStore.user?.id as number | undefined,
  refBillNo: '',
  remark: '',
  status: 'draft',
})

const items = ref<ItemRow[]>([])
const operatorName = computed(() => authStore.user?.name || authStore.user?.username || '-')
const productOptions = ref<any[]>([])
const productLoading = ref(false)

function diffOf(row: ItemRow) { return ((row.newPrice || 0) - (row.oldPrice || 0)) * (row.quantity || 0) }
const totalDiff = computed(() => items.value.reduce((s, i) => s + diffOf(i), 0).toFixed(2))

async function searchProducts(keyword: string) {
  productLoading.value = true
  try {
    const res = await api.get('/api/v1/products', { params: { keyword, pageSize: 50 } })
    if (res.data.code === 0 || res.data.code === 200) {
      productOptions.value = res.data.data?.list ?? res.data.data?.items ?? []
    }
  } finally { productLoading.value = false }
}

async function handleProductChange(row: ItemRow, productId: number) {
  const p = productOptions.value.find((x) => x.id === productId)
  if (p) {
    row.productName = p.name
    row.specification = p.specification
    row.unit = p.unit
    row.oldPrice = p.purchasePrice || 0
    if (!row.newPrice) row.newPrice = row.oldPrice
  }
  if (form.warehouseId && row.productId) {
    try {
      const res = await api.get('/api/v1/stocks', { params: { warehouseId: form.warehouseId, productId: row.productId, pageSize: 1 } })
      if (res.data.code === 0 || res.data.code === 200) {
        const list = res.data.data?.list ?? []
        row.stockQty = list.length ? list[0].quantity : 0
        if (list.length && list[0].costPrice) row.oldPrice = list[0].costPrice
      }
    } catch { /* 忽略 */ }
  }
}

function addRow() { items.value.push({ quantity: 1, oldPrice: 0, newPrice: 0 }) }

async function loadDetail(id: number) {
  loading.value = true
  try {
    const res = await api.get(`/api/v1/cost-adjusts/${id}`)
    if (res.data.code === 0 || res.data.code === 200) {
      const data = res.data.data
      Object.assign(form, {
        id: data.id,
        warehouseId: data.warehouseId,
        billDate: data.billDate ? dayjs(data.billDate).format('YYYY-MM-DD') : dayjs().format('YYYY-MM-DD'),
        handlerId: data.handlerId || undefined,
        refBillNo: data.refBillNo,
        remark: data.remark,
        status: data.status,
      })
      items.value = (data.items ?? []).map((it: any) => ({
        productId: it.productId,
        productName: it.product?.name,
        specification: it.product?.specification,
        unit: it.product?.unit,
        quantity: it.quantity, oldPrice: it.oldPrice, newPrice: it.newPrice, remark: it.remark,
      }))
      if (data.status !== 'draft') editable.value = false
    }
  } finally { loading.value = false }
}

async function save(complete: boolean) {
  if (!form.warehouseId) { ElMessage.warning('请选择仓库'); return }
  const validItems = items.value.filter((i) => i.productId)
  if (!validItems.length) { ElMessage.warning('请至少添加一行商品'); return }
  saving.value = true
  try {
    const payload = {
      warehouseId: form.warehouseId,
      billDate: form.billDate,
      handlerId: form.handlerId,
      refBillNo: form.refBillNo,
      remark: form.remark,
      items: validItems.map((i) => ({ productId: i.productId!, quantity: i.quantity, oldPrice: i.oldPrice, newPrice: i.newPrice, remark: i.remark })),
    }
    let id = billId.value
    const res = id
      ? await api.put(`/api/v1/cost-adjusts/${id}`, payload)
      : await api.post('/api/v1/cost-adjusts', payload)
    if (res.data.code === 0 || res.data.code === 200) {
      if (!id) id = res.data.data?.id
      if (complete && id) {
        await api.put(`/api/v1/cost-adjusts/${id}/complete`)
      }
      ElMessage.success(complete ? '已过账' : '已存入草稿')
      router.back()
    }
  } finally { saving.value = false }
}

function goBack() { router.back() }

onMounted(() => {
  if (billId.value) loadDetail(billId.value)
  else addRow()
})
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
