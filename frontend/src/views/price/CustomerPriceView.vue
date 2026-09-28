<template>
  <div class="page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>客户单独定价</span>
          <el-button type="primary" :disabled="!customerId" @click="openCreate">新增专属价格</el-button>
        </div>
      </template>

      <el-form inline class="search-form">
        <el-form-item label="客户">
          <RemoteSelect v-model="customerId" api-url="/api/v1/customers" placeholder="选择客户" style="width: 260px" @update:model-value="fetchPrices" />
        </el-form-item>
      </el-form>

      <el-table :data="prices" v-loading="loading" border stripe>
        <template #empty>
          <el-empty :description="customerId ? '该客户暂无专属价格' : '请先选择客户'" :image-size="80" />
        </template>
        <el-table-column prop="productId" label="商品" min-width="200">
          <template #default="{ row }">{{ productName(row.productId) }}</template>
        </el-table-column>
        <el-table-column label="单位" width="120">
          <template #default="{ row }">{{ row.unitId ? `单位#${row.unitId}` : '主单位' }}</template>
        </el-table-column>
        <el-table-column prop="price" label="专属价" width="140" align="right">
          <template #default="{ row }">¥{{ Number(row.price).toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button type="danger" size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" title="新增专属价格" width="480px">
      <el-form label-width="80px">
        <el-form-item label="商品" required>
          <RemoteSelect v-model="form.productId" api-url="/api/v1/products" placeholder="选择商品" style="width: 100%" />
        </el-form-item>
        <el-form-item label="专属价" required>
          <el-input-number v-model="form.price" :min="0" :precision="2" :controls="false" style="width: 100%" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import api from '@/api/client'
import { ElMessage, ElMessageBox } from 'element-plus'
import RemoteSelect from '@/components/RemoteSelect.vue'

interface CustomerPrice { id: number; productId: number; unitId: number; price: number }

const customerId = ref<number | undefined>()
const prices = ref<CustomerPrice[]>([])
const products = ref<Map<number, string>>(new Map())
const loading = ref(false)
const saving = ref(false)
const dialogVisible = ref(false)
const form = reactive({ productId: undefined as number | undefined, price: 0 })

function productName(id: number) {
  return products.value.get(id) || `商品#${id}`
}

async function loadProducts() {
  if (products.value.size) return
  const res = await api.get('/api/v1/products', { params: { page: 1, pageSize: 500 } })
  const data = res.data.data
  const rows = Array.isArray(data) ? data : (data?.list ?? [])
  products.value = new Map(rows.map((p: any) => [p.id, p.name]))
}

async function fetchPrices() {
  if (!customerId.value) {
    prices.value = []
    return
  }
  loading.value = true
  try {
    await loadProducts()
    const res = await api.get(`/api/v1/prices/customers/${customerId.value}`)
    prices.value = res.data.data || []
  } catch (error: any) {
    ElMessage.error(error.message || '加载失败')
  } finally {
    loading.value = false
  }
}

function openCreate() {
  form.productId = undefined
  form.price = 0
  dialogVisible.value = true
}

async function handleSubmit() {
  if (!form.productId || !customerId.value) {
    ElMessage.warning('请选择商品并填写价格')
    return
  }
  saving.value = true
  try {
    const res = await api.post(`/api/v1/prices/customers/${customerId.value}`, {
      prices: [{ productId: form.productId, price: form.price }],
    })
    if (res.data.code === 200) {
      ElMessage.success('设置成功')
      dialogVisible.value = false
      await fetchPrices()
    }
  } catch (error: any) {
    ElMessage.error(error.message || '设置失败')
  } finally {
    saving.value = false
  }
}

async function handleDelete(row: CustomerPrice) {
  try {
    await ElMessageBox.confirm(`确认删除商品「${productName(row.productId)}」的专属价格？`, '提示', { type: 'warning' })
  } catch {
    return
  }
  const res = await api.delete(`/api/v1/prices/customers/${customerId.value}/products/${row.productId}`)
  if (res.data.code === 200) {
    ElMessage.success('删除成功')
    await fetchPrices()
  }
}
</script>

<style scoped>
.page { padding: 20px; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
.search-form { margin-bottom: 16px; }
</style>
