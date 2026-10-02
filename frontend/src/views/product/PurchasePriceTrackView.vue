<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">采购价格跟踪</h2>
      <el-button @click="exportCsv('采购价格跟踪', exportCols)">导出</el-button>
    </div>
    <el-card>
      <el-form inline @submit.prevent>
        <el-form-item label="日期">
          <el-date-picker v-model="range" type="daterange" value-format="YYYY-MM-DD" start-placeholder="开始日期" end-placeholder="结束日期" style="width: 260px" />
        </el-form-item>
        <el-form-item label="商品">
          <el-input v-model="filters.keyword" placeholder="名称/编号/规格" clearable style="width: 180px" />
        </el-form-item>
        <el-form-item label="供应商">
          <el-input v-model="filters.supplierKeyword" placeholder="供应商名称" clearable style="width: 160px" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="search">查询</el-button>
          <el-button @click="reset">重置</el-button>
        </el-form-item>
      </el-form>
      <el-table :data="list" v-loading="loading" border stripe>
        <el-table-column prop="productCode" label="商品编号" width="110" />
        <el-table-column prop="product" label="商品名称" min-width="160" show-overflow-tooltip />
        <el-table-column prop="spec" label="规格" width="110" />
        <el-table-column prop="unit" label="单位" width="70" />
        <el-table-column prop="supplier" label="最近采购供应商" min-width="150" show-overflow-tooltip />
        <el-table-column prop="price" label="最近采购价" width="110" align="right">
          <template #default="{ row }">{{ row.price?.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column prop="quantity" label="数量" width="90" align="right" />
        <el-table-column prop="lastDate" label="最近采购日期" width="120" />
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
import { computed, onMounted } from 'vue'
import { useReport } from '@/composables/useReport'

interface Row {
  productId: number; product: string; productCode: string; spec: string; unit: string
  supplierId: number; supplier: string; price: number; quantity: number; lastDate: string
}

const { list, total, loading, page, pageSize, filters, load, search, reset, exportCsv } =
  useReport<Row>('/api/v1/product-reports/purchase-price-track', {
    startDate: '', endDate: '', keyword: '', supplierKeyword: '',
  })

const range = computed({
  get: () => (filters.value.startDate && filters.value.endDate ? [filters.value.startDate, filters.value.endDate] : null),
  set: (v) => {
    filters.value.startDate = v?.[0] ?? ''
    filters.value.endDate = v?.[1] ?? ''
  },
})

const exportCols = [
  { key: 'productCode', label: '商品编号' },
  { key: 'product', label: '商品名称' },
  { key: 'spec', label: '规格' },
  { key: 'unit', label: '单位' },
  { key: 'supplier', label: '最近采购供应商' },
  { key: 'price', label: '最近采购价' },
  { key: 'quantity', label: '数量' },
  { key: 'lastDate', label: '最近采购日期' },
]

onMounted(() => load(1))
</script>

<style scoped>
.page { display: flex; flex-direction: column; gap: 12px; }
.page-header { display: flex; justify-content: space-between; align-items: center; }
.page-title { font-size: 18px; font-weight: 600; margin: 0; }
</style>
