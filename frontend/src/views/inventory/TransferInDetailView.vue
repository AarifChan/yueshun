<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">调拨入库单</h2>
      <el-tag v-if="form.status === 'completed'" type="success">已过账</el-tag>
      <el-tag v-else-if="form.id" type="info">草稿</el-tag>
    </div>

    <div class="form-panel" v-loading="loading">
      <el-form :model="form" label-width="90px" :disabled="!editable" class="bill-form">
        <el-row :gutter="16">
          <el-col :span="8">
            <el-form-item label="调拨出库单" required>
              <el-select
                v-model="form.outId" filterable remote :remote-method="searchOutBills" :loading="outLoading"
                placeholder="选择已过账的调拨出库单" style="width: 100%" :disabled="!!billId"
                @focus="searchOutBills('')" @change="loadOutItems"
              >
                <el-option v-for="b in outBillOptions" :key="b.id" :label="`${b.billNo}（${b.fromWarehouseName} → ${b.toWarehouseName}）`" :value="b.id" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="5">
            <el-form-item label="调出仓库"><el-input :model-value="form.fromWarehouseName" disabled /></el-form-item>
          </el-col>
          <el-col :span="5">
            <el-form-item label="调入仓库"><el-input :model-value="form.toWarehouseName" disabled /></el-form-item>
          </el-col>
          <el-col :span="6">
            <el-form-item label="单据备注"><el-input v-model="form.remark" placeholder="请输入" :disabled="!editable" /></el-form-item>
          </el-col>
        </el-row>
      </el-form>
    </div>

    <div class="items-panel">
      <div class="items-toolbar">
        <span class="items-title">入库明细（按出库单行确认实收数量）</span>
      </div>
      <el-table :data="items" border size="small">
        <el-table-column type="index" width="50" />
        <el-table-column label="商品名称" min-width="220">
          <template #default="{ row }">{{ row.productName }}</template>
        </el-table-column>
        <el-table-column label="商品规格" width="130">
          <template #default="{ row }">{{ row.specification || '-' }}</template>
        </el-table-column>
        <el-table-column label="单位" width="80">
          <template #default="{ row }">{{ row.unit || '-' }}</template>
        </el-table-column>
        <el-table-column label="出库数量" width="100" align="right">
          <template #default="{ row }">{{ row.outQty }}</template>
        </el-table-column>
        <el-table-column label="已入库数量" width="110" align="right">
          <template #default="{ row }">{{ row.receivedQty }}</template>
        </el-table-column>
        <el-table-column label="本次入库数量" width="150">
          <template #default="{ row }">
            <el-input-number v-if="editable" v-model="row.quantity" :min="0" :precision="2" controls-position="right" style="width: 100%" />
            <span v-else>{{ row.quantity }}</span>
          </template>
        </el-table-column>
        <el-table-column label="单价" width="120">
          <template #default="{ row }">
            <el-input-number v-if="editable" v-model="row.price" :min="0" :precision="4" controls-position="right" style="width: 100%" />
            <span v-else>{{ row.price }}</span>
          </template>
        </el-table-column>
        <el-table-column label="金额" width="110" align="right">
          <template #default="{ row }">{{ ((row.quantity || 0) * (row.price || 0)).toFixed(2) }}</template>
        </el-table-column>
      </el-table>
      <div class="summary">
        <span>合计数量: {{ totalQty }}</span>
        <span>合计金额: ¥ {{ totalAmount }}</span>
      </div>
    </div>

    <div class="footer-bar">
      <span class="operator">制单人：{{ operatorName }}</span>
      <div class="footer-actions">
        <el-button @click="goBack">返 回</el-button>
        <el-button v-if="editable" type="primary" :loading="saving" @click="save">保存并过账</el-button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import api from '@/api/client'
import { useAuthStore } from '@/stores/auth'

interface ItemRow {
  productId: number; productName?: string; specification?: string; unit?: string
  outQty: number; receivedQty: number; quantity: number; price: number
}

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const billId = computed(() => (route.params.id && route.params.id !== 'new' ? Number(route.params.id) : null))
const editable = ref(true)
const loading = ref(false)
const saving = ref(false)

const form = reactive({
  id: 0,
  outId: undefined as number | undefined,
  fromWarehouseName: '',
  toWarehouseName: '',
  remark: '',
  status: 'draft',
})
const items = ref<ItemRow[]>([])
const outBillOptions = ref<any[]>([])
const outLoading = ref(false)
const operatorName = computed(() => authStore.user?.name || authStore.user?.username || '-')
const totalQty = computed(() => items.value.reduce((s, i) => s + (i.quantity || 0), 0))
const totalAmount = computed(() => items.value.reduce((s, i) => s + (i.quantity || 0) * (i.price || 0), 0).toFixed(2))

async function searchOutBills(keyword: string) {
  outLoading.value = true
  try {
    const res = await api.get('/api/v1/transfer-outs', { params: { keyword, status: 'completed', pageSize: 50 } })
    if (res.data.code === 0 || res.data.code === 200) outBillOptions.value = res.data.data?.list ?? []
  } finally { outLoading.value = false }
}

async function loadOutItems(outId: number) {
  if (!outId) return
  const bill = outBillOptions.value.find((b) => b.id === outId)
  if (bill) {
    form.fromWarehouseName = bill.fromWarehouseName
    form.toWarehouseName = bill.toWarehouseName
  }
  const res = await api.get(`/api/v1/transfer-outs/${outId}`)
  if (res.data.code === 0 || res.data.code === 200) {
    const data = res.data.data
    items.value = (data.items ?? [])
      .filter((it: any) => it.quantity - (it.receivedQty || 0) > 0)
      .map((it: any) => ({
        productId: it.productId, productName: it.product?.name, specification: it.product?.specification,
        unit: it.product?.unit, outQty: it.quantity, receivedQty: it.receivedQty || 0,
        quantity: it.quantity - (it.receivedQty || 0), price: it.price,
      }))
  }
}

async function loadDetail(id: number) {
  loading.value = true
  try {
    const res = await api.get(`/api/v1/transfer-ins/${id}`)
    if (res.data.code === 0 || res.data.code === 200) {
      const data = res.data.data
      Object.assign(form, {
        id: data.id, outId: data.outId, remark: data.remark, status: data.status,
      })
      items.value = (data.items ?? []).map((it: any) => ({
        productId: it.productId, productName: it.product?.name, specification: it.product?.specification,
        unit: it.product?.unit, outQty: 0, receivedQty: 0, quantity: it.quantity, price: it.price,
      }))
      if (data.status !== 'draft') editable.value = false
    }
  } finally { loading.value = false }
}

async function save() {
  if (!form.outId) { ElMessage.warning('请选择调拨出库单'); return }
  const validItems = items.value.filter((i) => i.quantity > 0)
  if (!validItems.length) { ElMessage.warning('请填写本次入库数量'); return }
  saving.value = true
  try {
    const payload = {
      outId: form.outId, remark: form.remark,
      items: validItems.map((i) => ({ productId: i.productId, quantity: i.quantity, price: i.price || 0 })),
    }
    const res = await api.post('/api/v1/transfer-ins', payload)
    if (res.data.code === 0 || res.data.code === 200) {
      const id = res.data.data?.id
      if (id) await api.put(`/api/v1/transfer-ins/${id}/complete`)
      ElMessage.success('已过账')
      router.back()
    }
  } finally { saving.value = false }
}

function goBack() { router.back() }

onMounted(() => { if (billId.value) loadDetail(billId.value) })
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
.summary span { margin-left: 24px; }
.footer-bar { display: flex; justify-content: space-between; align-items: center; background: var(--el-bg-color); border-radius: 8px; padding: 12px 16px; }
.operator { color: var(--el-text-color-secondary); }
.footer-actions { display: flex; gap: 8px; }
</style>
