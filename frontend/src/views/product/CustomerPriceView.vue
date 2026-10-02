<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">客户单独定价</h2>
      <div class="actions">
        <el-button :disabled="!selectedRows.length" @click="exportSelectedCsv('客户单独定价', exportCols, selectedRows)">导出选中</el-button>
        <el-button @click="exportCsv('客户单独定价', exportCols)">导出</el-button>
        <el-button type="primary" @click="openDialog()">新增定价</el-button>
      </div>
    </div>
    <el-card>
      <el-form inline @submit.prevent>
        <el-form-item label="客户">
          <el-input v-model="filters.customerKeyword" placeholder="客户名称/编号/联系人/手机号" clearable style="width: 220px" />
        </el-form-item>
        <el-form-item label="商品">
          <el-input v-model="filters.productKeyword" placeholder="商品名称/编号" clearable style="width: 180px" />
        </el-form-item>
        <el-form-item label="商品分类">
          <el-tree-select
            v-model="filters.categoryId"
            :data="categoryTree"
            :props="{ label: 'name', children: 'children' }"
            node-key="id"
            check-strictly
            clearable
            placeholder="商品分类"
            style="width: 160px"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="search">查询</el-button>
          <el-button @click="reset">重置</el-button>
        </el-form-item>
      </el-form>
      <el-table :data="list" v-loading="loading" border stripe @selection-change="(rows: Row[]) => (selectedRows = rows)">
        <el-table-column type="selection" width="45" />
        <el-table-column prop="customerCode" label="客户编号" width="110" />
        <el-table-column prop="customer" label="客户名称" min-width="150" show-overflow-tooltip />
        <el-table-column prop="productCode" label="商品编号" width="110" />
        <el-table-column prop="product" label="商品名称" min-width="160" show-overflow-tooltip />
        <el-table-column prop="spec" label="规格" width="110" />
        <el-table-column prop="unit" label="单位" width="70" />
        <el-table-column prop="retailPrice" label="标准零售价" width="110" align="right">
          <template #default="{ row }">{{ row.retailPrice?.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column prop="price" label="客户定价" width="110" align="right">
          <template #default="{ row }">
            <span class="price">{{ row.price?.toFixed(2) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="remark" label="备注" min-width="120" show-overflow-tooltip />
        <el-table-column label="操作" width="130" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openDialog(row)">编辑</el-button>
            <el-button link type="danger" @click="remove(row)">删除</el-button>
          </template>
        </el-table-column>
        <template #empty>暂无数据</template>
      </el-table>
      <el-pagination
        v-model:current-page="page" v-model:page-size="pageSize"
        :total="total" :page-sizes="[30, 50, 100]"
        layout="total, sizes, prev, pager, next" style="margin-top: 12px; justify-content: flex-end"
        @current-change="load()" @size-change="load(1)" />
    </el-card>

    <el-dialog v-model="dialogVisible" :title="form.id ? '编辑定价' : '新增定价'" width="520px">
      <el-form :model="form" label-width="90px">
        <el-form-item label="客户" required>
          <el-select v-model="form.customerId" filterable remote :remote-method="searchCustomers" :disabled="!!form.id" style="width: 100%" placeholder="输入客户名称搜索">
            <el-option v-for="c in customerOptions" :key="c.id" :label="c.name" :value="c.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="商品" required>
          <el-select v-model="form.productId" filterable remote :remote-method="searchProducts" :disabled="!!form.id" style="width: 100%" placeholder="输入商品名称/编号搜索">
            <el-option v-for="p in productOptions" :key="p.id" :label="`${p.name}（${p.code || '无编号'}）`" :value="p.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="客户定价" required>
          <el-input-number v-model="form.price" :min="0" :precision="2" style="width: 100%" />
        </el-form-item>
        <el-form-item label="备注"><el-input v-model="form.remark" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="save">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '@/api/client'
import { useReport } from '@/composables/useReport'
import { fetchCategoryTree, type CategoryNode } from '@/api/category'

interface Row {
  id: number; customerId: number; customer: string; customerCode: string
  productId: number; product: string; productCode: string; spec: string; unit: string
  retailPrice: number; price: number; remark: string
}

interface Option { id: number; name: string; code?: string }

const { list, total, loading, page, pageSize, filters, load, search, reset, exportCsv, exportSelectedCsv } =
  useReport<Row>('/api/v1/customer-prices', { customerKeyword: '', productKeyword: '', categoryId: undefined })

const selectedRows = ref<Row[]>([])
const categoryTree = ref<CategoryNode[]>([])

async function loadCategoryTree() {
  try {
    const res = await fetchCategoryTree()
    if (res.data.code === 0 || res.data.code === 200) {
      categoryTree.value = Array.isArray(res.data.data) ? res.data.data : []
    }
  } catch { /* 选项加载失败不阻塞列表 */ }
}

const exportCols = [
  { key: 'customerCode', label: '客户编号' },
  { key: 'customer', label: '客户名称' },
  { key: 'productCode', label: '商品编号' },
  { key: 'product', label: '商品名称' },
  { key: 'spec', label: '规格' },
  { key: 'unit', label: '单位' },
  { key: 'retailPrice', label: '标准零售价' },
  { key: 'price', label: '客户定价' },
  { key: 'remark', label: '备注' },
]

const dialogVisible = ref(false)
const form = ref({ id: 0, customerId: undefined as number | undefined, productId: undefined as number | undefined, price: 0, remark: '' })
const customerOptions = ref<Option[]>([])
const productOptions = ref<Option[]>([])

async function searchCustomers(kw: string) {
  const res = await api.get('/api/v1/customers', { params: { page: 1, pageSize: 50, keyword: kw } })
  if (res.data.code === 0 || res.data.code === 200) customerOptions.value = res.data.data?.list ?? []
}

async function searchProducts(kw: string) {
  const res = await api.get('/api/v1/products', { params: { page: 1, pageSize: 50, keyword: kw } })
  if (res.data.code === 0 || res.data.code === 200) productOptions.value = res.data.data?.list ?? []
}

function openDialog(row?: Row) {
  if (row) {
    form.value = { id: row.id, customerId: row.customerId, productId: row.productId, price: row.price, remark: row.remark }
    customerOptions.value = [{ id: row.customerId, name: row.customer }]
    productOptions.value = [{ id: row.productId, name: row.product, code: row.productCode }]
  } else {
    form.value = { id: 0, customerId: undefined, productId: undefined, price: 0, remark: '' }
    searchCustomers('')
    searchProducts('')
  }
  dialogVisible.value = true
}

async function save() {
  if (!form.value.customerId || !form.value.productId) { ElMessage.warning('请选择客户和商品'); return }
  const res = form.value.id
    ? await api.put(`/api/v1/customer-prices/${form.value.id}`, { price: form.value.price, remark: form.value.remark })
    : await api.post('/api/v1/customer-prices', form.value)
  if (res.data.code === 0 || res.data.code === 200) {
    ElMessage.success('保存成功')
    dialogVisible.value = false
    load()
  } else {
    ElMessage.error(res.data.message || '保存失败')
  }
}

async function remove(row: Row) {
  await ElMessageBox.confirm(`确定删除「${row.customer} - ${row.product}」的定价？`, '提示', { type: 'warning' })
  const res = await api.delete(`/api/v1/customer-prices/${row.id}`)
  if (res.data.code === 0 || res.data.code === 200) {
    ElMessage.success('已删除')
    load()
  }
}

onMounted(() => {
  load(1)
  loadCategoryTree()
})
</script>

<style scoped>
.page { display: flex; flex-direction: column; gap: 12px; }
.page-header { display: flex; justify-content: space-between; align-items: center; }
.page-title { font-size: 18px; font-weight: 600; margin: 0; }
.actions { display: flex; gap: 8px; }
.price { color: #f56c6c; font-weight: 600; }
</style>
