<template>
  <div class="page">
    <el-card>
      <template #header>
        <div class="card-header"><span>商品库存期初</span><el-button type="primary" @click="openCreate">新增期初</el-button></div>
      </template>
      <el-form :model="searchForm" inline class="search-form">
        <el-form-item label="商品"><el-input v-model="searchForm.keyword" placeholder="商品名称/编号" clearable /></el-form-item>
        <el-form-item><el-button type="primary" @click="handleSearch">搜索</el-button><el-button @click="handleReset">重置</el-button></el-form-item>
      </el-form>
      <el-table :data="list" v-loading="loading" border stripe>
        <el-table-column prop="productCode" label="商品编号" width="120" />
        <el-table-column prop="productName" label="商品名称" min-width="160" />
        <el-table-column prop="warehouseName" label="仓库" min-width="120" />
        <el-table-column prop="quantity" label="期初数量" width="110" align="right" />
        <el-table-column prop="costPrice" label="成本单价" width="110" align="right">
          <template #default="{ row }">{{ row.costPrice?.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column prop="costAmount" label="成本金额" width="120" align="right">
          <template #default="{ row }">{{ row.costAmount?.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="openEdit(row)">编辑</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-pagination v-model:current-page="pagination.page" v-model:page-size="pagination.pageSize" :total="total" :page-sizes="[10,20,50]" layout="total, sizes, prev, pager, next" @size-change="handleSizeChange" @current-change="handleCurrentChange" class="pagination" />
    </el-card>
    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="500px">
      <el-form :model="form" label-width="90px">
        <el-form-item label="商品" required>
          <el-select v-model="form.productId" filterable remote :remote-method="searchProducts" placeholder="选择商品" style="width: 100%" :disabled="isEdit">
            <el-option v-for="p in productOptions" :key="p.id" :label="`${p.name}(${p.code})`" :value="p.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="仓库" required>
          <el-select v-model="form.warehouseId" placeholder="选择仓库" style="width: 100%" :disabled="isEdit">
            <el-option v-for="w in warehouseOptions" :key="w.id" :label="w.name" :value="w.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="期初数量"><el-input-number v-model="form.quantity" :min="0" :precision="2" style="width: 100%" /></el-form-item>
        <el-form-item label="成本单价"><el-input-number v-model="form.costPrice" :min="0" :precision="4" style="width: 100%" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import api from '@/api/client'
import { useCrud } from '@/composables/useCrud'

interface InitialStock {
  id: number; productId: number; warehouseId: number; quantity: number; costPrice: number; costAmount: number
  productName?: string; productCode?: string; warehouseName?: string
}
interface Option { id: number; name: string; code?: string }

const productOptions = ref<Option[]>([])
const warehouseOptions = ref<Option[]>([])

const {
  list, total, loading, dialogVisible, dialogTitle, form, isEdit, searchForm, pagination,
  fetchList, openCreate: openCreateBase, openEdit: openEditBase,
  handleSizeChange, handleCurrentChange, handleSearch, handleReset,
} = useCrud<InitialStock>({
  baseUrl: '/base-data/initial-stocks',
  defaultForm: () => ({ productId: undefined, warehouseId: undefined, quantity: 0, costPrice: 0 }),
})

async function searchProducts(kw: string) {
  const res = await api.get('/products', { params: { page: 1, pageSize: 20, keyword: kw || '' } })
  if (res.data.code === 0 || res.data.code === 200) {
    const data = res.data.data
    productOptions.value = Array.isArray(data) ? data : (data?.list ?? [])
  }
}

async function loadWarehouses() {
  const res = await api.get('/warehouses', { params: { page: 1, pageSize: 100 } })
  if (res.data.code === 0 || res.data.code === 200) {
    const data = res.data.data
    warehouseOptions.value = Array.isArray(data) ? data : (data?.list ?? [])
  }
}

function openCreate() { openCreateBase() }
function openEdit(row: InitialStock) { openEditBase(row) }

async function handleSubmit() {
  if (!form.value.productId || !form.value.warehouseId) {
    ElMessage.warning('请选择商品和仓库')
    return
  }
  const res = await api.put('/base-data/initial-stocks', form.value)
  if (res.data.code === 0 || res.data.code === 200) {
    ElMessage.success('保存成功')
    dialogVisible.value = false
    await fetchList()
  } else {
    ElMessage.error(res.data.message || '保存失败')
  }
}

onMounted(() => { searchProducts(''); loadWarehouses() })
</script>

<style scoped>
.card-header { display: flex; justify-content: space-between; align-items: center; }
.pagination { margin-top: 16px; justify-content: flex-end; }
</style>
