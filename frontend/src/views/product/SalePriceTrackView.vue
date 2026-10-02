<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">销售价格跟踪</h2>
      <el-button @click="exportCsv('销售价格跟踪', exportCols)">导出</el-button>
    </div>
    <el-card>
      <el-form inline @submit.prevent>
        <el-form-item label="日期">
          <el-date-picker v-model="range" type="daterange" value-format="YYYY-MM-DD" start-placeholder="开始日期" end-placeholder="结束日期" style="width: 260px" />
        </el-form-item>
        <el-form-item label="商品">
          <el-input v-model="filters.keyword" placeholder="名称/编号/规格" clearable style="width: 180px" />
        </el-form-item>
        <el-form-item label="客户">
          <el-input v-model="filters.customerKeyword" placeholder="客户名称/编号" clearable style="width: 160px" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="onSearch">查询</el-button>
          <el-button @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>
      <el-table :data="list" v-loading="loading" border stripe>
        <el-table-column prop="productCode" label="商品编号" width="110" />
        <el-table-column prop="product" label="商品名称" min-width="160" show-overflow-tooltip />
        <el-table-column prop="spec" label="规格" width="110" />
        <el-table-column prop="unit" label="单位" width="70" />
        <el-table-column prop="customerCode" label="客户编号" width="110" />
        <el-table-column prop="customer" label="最近销售客户" min-width="150" show-overflow-tooltip />
        <el-table-column prop="price" label="销售折前价" width="110" align="right">
          <template #default="{ row }">{{ row.price?.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="销售折扣" width="100" align="right">
          <template #default="{ row }">{{ discountOf(row) }}</template>
        </el-table-column>
        <el-table-column label="销售折后价" width="110" align="right">
          <template #default="{ row }">{{ netPriceOf(row) }}</template>
        </el-table-column>
        <el-table-column prop="quantity" label="数量" width="90" align="right" />
        <el-table-column prop="lastDate" label="最近销售日期" width="120" />
        <template #empty>暂无数据</template>
      </el-table>
      <el-pagination
        v-model:current-page="page" v-model:page-size="pageSize"
        :total="total" :page-sizes="[30, 50, 100]"
        layout="total, sizes, prev, pager, next" style="margin-top: 12px; justify-content: flex-end"
        @current-change="load()" @size-change="load(1)" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useReport } from '@/composables/useReport'

interface Row {
  productId: number; product: string; productCode: string; spec: string; unit: string
  customerId: number; customer: string; customerCode: string
  price: number; amount: number; quantity: number; lastDate: string
}

const { list, total, loading, page, pageSize, filters, load, search, reset, exportCsv } =
  useReport<Row>('/api/v1/product-reports/sale-price-track', {
    startDate: '', endDate: '', keyword: '', customerKeyword: '',
  })

const range = computed({
  get: () => (filters.value.startDate && filters.value.endDate ? [filters.value.startDate, filters.value.endDate] : null),
  set: (v) => {
    filters.value.startDate = v?.[0] ?? ''
    filters.value.endDate = v?.[1] ?? ''
  },
})

const rangeInit = ref(range.value)
void rangeInit

function onSearch() { search() }
function onReset() { reset() }

function netPriceOf(row: Row): string {
  if (!row.quantity) return row.price?.toFixed(2) ?? '0.00'
  return (row.amount / row.quantity).toFixed(2)
}

function discountOf(row: Row): string {
  if (!row.price || !row.quantity) return '-'
  const net = row.amount / row.quantity
  return `${((net / row.price) * 100).toFixed(0)}%`
}

const exportCols = [
  { key: 'productCode', label: '商品编号' },
  { key: 'product', label: '商品名称' },
  { key: 'spec', label: '规格' },
  { key: 'unit', label: '单位' },
  { key: 'customer', label: '最近销售客户' },
  { key: 'price', label: '销售折前价' },
  { key: 'quantity', label: '数量' },
  { key: 'amount', label: '销售金额' },
  { key: 'lastDate', label: '最近销售日期' },
]

onMounted(() => load(1))
</script>

<style scoped>
.page { display: flex; flex-direction: column; gap: 12px; }
.page-header { display: flex; justify-content: space-between; align-items: center; }
.page-title { font-size: 18px; font-weight: 600; margin: 0; }
</style>
