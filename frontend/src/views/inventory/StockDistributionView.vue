<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">库存分布表</h2>
      <div>
        <el-button @click="doExport">导出</el-button>
        <el-button type="primary" @click="search">查询</el-button>
      </div>
    </div>
    <el-card>
      <el-form inline class="search-form">
        <el-form-item label="商品">
          <el-input v-model="filters.keyword" placeholder="商品名称/编号/条码/规格/关键字" clearable style="width: 240px" />
        </el-form-item>
      </el-form>
      <el-table :data="list" v-loading="loading" border stripe show-summary>
        <el-table-column prop="productName" label="商品名称" min-width="160" fixed="left" />
        <el-table-column prop="specification" label="规格" width="110" />
        <el-table-column prop="unit" label="单位" width="70" />
        <el-table-column prop="totalQty" label="合计" width="90" align="right" />
        <el-table-column v-for="w in warehouses" :key="w.id" :label="w.name" width="100" align="right">
          <template #default="{ row }">{{ row.warehouses?.[w.name] ?? 0 }}</template>
        </el-table-column>
        <el-table-column prop="totalQty" label="库存总量" width="100" align="right" />
        <el-table-column label="成本均价" width="100" align="right">
          <template #default="{ row }">{{ row.costPrice?.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="库存总额" width="120" align="right">
          <template #default="{ row }">{{ row.amount?.toFixed(2) }}</template>
        </el-table-column>
        <template #empty>暂无数据</template>
      </el-table>
      <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" :page-sizes="[30,50,100]" layout="total, sizes, prev, pager, next" @size-change="() => load(1)" @current-change="load" class="pagination" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useReport } from '@/composables/useReport'

interface Row {
  productId: number; productName: string; code: string; specification: string; unit: string
  totalQty: number; costPrice: number; amount: number; warehouses?: Record<string, number>
}

const { list, total, loading, page, pageSize, filters, extra, load, search, exportCsv } = useReport<Row>(
  '/api/v1/inventory-reports/stock-distribution',
  { keyword: '' },
)

const warehouses = computed(() => (extra.value.warehouses as { id: number; name: string }[]) ?? [])

function doExport() {
  const cols = [
    { key: 'productName', label: '商品名称' }, { key: 'specification', label: '规格' },
    { key: 'unit', label: '单位' }, { key: 'totalQty', label: '库存总量' },
    { key: 'costPrice', label: '成本均价' }, { key: 'amount', label: '库存总额' },
  ]
  exportCsv('库存分布表', cols)
}

onMounted(() => load(1))
</script>

<style scoped>
.page { display: flex; flex-direction: column; gap: 12px; }
.page-header { display: flex; justify-content: space-between; align-items: center; }
.page-title { font-size: 18px; font-weight: 600; margin: 0; }
.pagination { margin-top: 16px; justify-content: flex-end; }
</style>
