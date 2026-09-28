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
          <el-col :span="8"><el-form-item label="客户" required><RemoteSelect v-model="form.customerId" api-url="/api/v1/customers" :params="{ type: 'customer' }" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="仓库" required><RemoteSelect v-model="form.warehouseId" api-url="/api/v1/warehouses" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="折扣"><el-input-number v-model="form.discount" :min="0" :precision="2" style="width:100%" /></el-form-item></el-col>
        </el-row>
        <el-form-item label="备注"><el-input v-model="form.remark" type="textarea" /></el-form-item>
      </el-form>
    </el-card>
    <el-card class="detail-card">
      <div class="history-bar">
        <el-button size="small" type="warning" plain :disabled="!form.customerId" @click="showHistoryPrice">查看历史售价</el-button>
        <span v-if="!form.customerId" class="tip">选择客户后可查看该客户各商品的历史售价</span>
      </div>
      <BillItemTable :items="items" :editable="isEditable" @add="addItem" @remove="removeItem" />
    </el-card>
    <el-card class="detail-card" v-if="!isEditable && form.status">
      <div class="status-actions">
        <template v-if="form.status === 'draft'">
          <el-button type="success" @click="doStatusAction({ api: '/api/v1/sales-orders/:id/confirm' })">确认订单</el-button>
          <el-button type="danger" @click="doStatusAction({ api: '/api/v1/sales-orders/:id/cancel' })">取消订单</el-button>
        </template>
      </div>
    </el-card>
    <div class="footer-actions">
      <template v-if="isEditable">
        <el-button type="primary" @click="save" :loading="saving">保存</el-button>
      </template>
      <el-button @click="goBack">返回</el-button>
    </div>
    <el-dialog v-model="historyVisible" title="历史售价" width="680px">
      <el-table :data="historyList" v-loading="historyLoading" border size="small">
        <el-table-column label="商品" min-width="160">
          <template #default="{ row }">{{ row.productName || row.productId }}</template>
        </el-table-column>
        <el-table-column label="历史价" width="140">
          <template #default="{ row }">{{ row.price != null ? `¥${Number(row.price).toFixed(2)}` : '无历史记录' }}</template>
        </el-table-column>
        <el-table-column label="单号" min-width="150">
          <template #default="{ row }">{{ row.orderNo || '-' }}</template>
        </el-table-column>
        <el-table-column label="日期" width="120">
          <template #default="{ row }">{{ row.date || '-' }}</template>
        </el-table-column>
      </el-table>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useBillDetail } from '@/composables/useBillDetail'
import RemoteSelect from '@/components/RemoteSelect.vue'
import BillItemTable from '@/components/BillItemTable.vue'
import api from '@/api/client'

const route = useRoute()
const id = route.params.id as string
const routeMode = (route.query.mode as string) || 'view'

const {
  form, items, loading, saving, isEditable,
  initCreate, loadDetail, addItem, removeItem, save, doStatusAction, goBack, setMode,
} = useBillDetail({
  baseUrl: '/api/v1/sales-orders',
  defaultForm: () => ({ orderDate: '', customerId: undefined, warehouseId: undefined, discount: 0, remark: '' }),
  defaultItem: () => ({ productId: undefined, quantity: 0, price: 0, remark: '' }),
})

const pageTitle = computed(() => {
  if (id === 'new') return '新建销售订单'
  return routeMode === 'edit' ? '编辑销售订单' : '销售订单详情'
})

if (id === 'new') {
  initCreate()
} else {
  loadDetail(id)
  if (routeMode === 'edit') setMode('edit')
}

// 历史售价（开单必看历销）
const historyVisible = ref(false)
const historyLoading = ref(false)
const historyList = ref<Record<string, any>[]>([])

async function showHistoryPrice() {
  const rows = items.value.filter(it => it.productId)
  if (!rows.length) {
    ElMessage.warning('请先添加商品明细')
    return
  }
  historyVisible.value = true
  historyLoading.value = true
  historyList.value = rows.map(it => ({ productId: it.productId, productName: it.productName }))
  try {
    await Promise.all(rows.map(async (it, idx) => {
      try {
        const res = await api.get('/api/v1/sales-orders/history-price', {
          params: { customerId: form.value.customerId, productId: it.productId },
        })
        if ((res.data.code === 0 || res.data.code === 200) && res.data.data) {
          historyList.value[idx] = { ...historyList.value[idx], ...res.data.data }
        }
      } catch {
        // 单个商品查不到历史价时保持“无历史记录”
      }
    }))
  } finally {
    historyLoading.value = false
  }
}
</script>

<style scoped>
.page { padding: 20px; }
.detail-card { margin-top: 16px; }
.history-bar { display: flex; align-items: center; gap: 12px; margin-bottom: 8px; }
.history-bar .tip { font-size: 12px; color: #909399; }
.footer-actions { margin-top: 16px; text-align: center; }
.status-actions { text-align: center; }
</style>
