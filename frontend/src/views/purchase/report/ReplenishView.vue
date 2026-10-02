<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">智能补货</h2>
      <div>
        <el-button @click="load">刷新</el-button>
        <el-button type="primary" :disabled="!selected.length" @click="orderDialogVisible = true">生成采购订单</el-button>
      </div>
    </div>
    <el-card>
      <el-alert type="info" :closable="false" title="根据商品库存上下限自动计算建议补货数量（当前库存低于最低库存的商品），勾选后可一键生成采购订单草稿。" class="tip" />
      <el-table :data="list" v-loading="loading" border stripe @selection-change="(rows: Row[]) => selected = rows">
        <el-table-column type="selection" width="50" />
        <el-table-column prop="productCode" label="商品编号" width="110" />
        <el-table-column prop="productName" label="商品名称" min-width="150" />
        <el-table-column prop="unit" label="单位" width="70" />
        <el-table-column label="当前库存" width="100" align="right">
          <template #default="{ row }">
            <span class="low">{{ row.currentStock }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="minStock" label="最低库存" width="100" align="right" />
        <el-table-column prop="maxStock" label="最高库存" width="100" align="right" />
        <el-table-column label="建议补货数量" width="120" align="right">
          <template #default="{ row }">
            <el-input-number v-model="row.suggestQty" :min="0" size="small" style="width: 110px" />
          </template>
        </el-table-column>
        <template #empty>暂无需要补货的商品</template>
      </el-table>
    </el-card>

    <el-dialog v-model="orderDialogVisible" title="生成采购订单" width="480px">
      <el-form label-width="90px">
        <el-form-item label="供应商" required>
          <el-select v-model="orderForm.supplierId" filterable placeholder="选择供应商" style="width: 100%">
            <el-option v-for="s in suppliers" :key="s.id" :label="s.name" :value="s.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="入库仓库" required>
          <el-select v-model="orderForm.warehouseId" filterable placeholder="选择仓库" style="width: 100%">
            <el-option v-for="w in warehouses" :key="w.id" :label="w.name" :value="w.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="备注"><el-input v-model="orderForm.remark" type="textarea" placeholder="智能补货生成" /></el-form-item>
        <el-form-item label="商品明细">
          <div class="items-preview">已选 {{ selected.length }} 个商品，合计 {{ totalQty }} 件</div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="orderDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="creating" @click="createOrder">生成订单</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import api from '@/api/client'

interface Row {
  productId: number; productCode: string; productName: string; unit: string
  currentStock: number; minStock: number; maxStock: number; suggestQty: number
}
interface Opt { id: number; name: string }

const router = useRouter()
const list = ref<Row[]>([])
const loading = ref(false)
const selected = ref<Row[]>([])
const orderDialogVisible = ref(false)
const creating = ref(false)
const orderForm = ref({ supplierId: null as number | null, warehouseId: null as number | null, remark: '智能补货生成' })
const suppliers = ref<Opt[]>([])
const warehouses = ref<Opt[]>([])

const totalQty = computed(() => selected.value.reduce((s, r) => s + (r.suggestQty || 0), 0))

async function load() {
  loading.value = true
  try {
    const res = await api.get('/api/v1/purchase-orders/replenish-suggestions')
    if (res.data.code === 0 || res.data.code === 200) {
      list.value = res.data.data?.list ?? []
    }
  } finally {
    loading.value = false
  }
}

async function loadOptions() {
  const [s, w] = await Promise.all([
    api.get('/api/v1/suppliers', { params: { page: 1, pageSize: 200 } }),
    api.get('/api/v1/warehouses', { params: { page: 1, pageSize: 100 } }),
  ])
  if (s.data.code === 0 || s.data.code === 200) suppliers.value = s.data.data?.list ?? []
  if (w.data.code === 0 || w.data.code === 200) warehouses.value = w.data.data?.list ?? []
}

async function createOrder() {
  if (!orderForm.value.supplierId) { ElMessage.warning('请选择供应商'); return }
  if (!orderForm.value.warehouseId) { ElMessage.warning('请选择入库仓库'); return }
  const items = selected.value.filter((r) => r.suggestQty > 0).map((r) => ({
    productId: r.productId, quantity: r.suggestQty, price: 0, remark: '',
  }))
  if (!items.length) { ElMessage.warning('建议补货数量需大于0'); return }
  creating.value = true
  try {
    const res = await api.post('/api/v1/purchase-orders', {
      supplierId: orderForm.value.supplierId,
      warehouseId: orderForm.value.warehouseId,
      orderDate: new Date().toISOString().slice(0, 10),
      remark: orderForm.value.remark,
      items,
    })
    if (res.data.code === 0 || res.data.code === 200) {
      ElMessage.success('采购订单已生成（草稿）')
      orderDialogVisible.value = false
      router.push('/purchase-orders')
    } else {
      ElMessage.error(res.data.message || '生成失败')
    }
  } finally {
    creating.value = false
  }
}

onMounted(() => { load(); loadOptions() })
</script>

<style scoped>
.page { display: flex; flex-direction: column; gap: 12px; }
.page-header { display: flex; justify-content: space-between; align-items: center; }
.page-title { font-size: 18px; font-weight: 600; margin: 0; }
.tip { margin-bottom: 12px; }
.low { color: var(--el-color-danger); font-weight: 600; }
.items-preview { color: var(--el-text-color-secondary); }
</style>
