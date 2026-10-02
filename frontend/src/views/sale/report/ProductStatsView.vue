<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">商品销售统计</h2>
      <el-button @click="doExport">导出</el-button>
    </div>
    <el-card>
      <el-form inline>
        <el-form-item label="时间">
          <el-date-picker v-model="dateRange" type="daterange" value-format="YYYY-MM-DD" start-placeholder="开始时间" end-placeholder="结束时间" @change="onDateChange" />
        </el-form-item>
        <el-form-item label="商品"><el-input v-model="filters.keyword" placeholder="商品名称/编号/规格" clearable /></el-form-item>
        <el-form-item>
          <el-button type="primary" @click="search">查询</el-button>
          <el-button @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>
      <el-table :data="list" v-loading="loading" border stripe show-summary>
        <el-table-column prop="name" label="商品名称" min-width="140" fixed="left" />
        <el-table-column prop="spec" label="规格" width="110" />
        <el-table-column label="实销数量" width="100" align="right">
          <template #default="{ row }">{{ row.qty }}</template>
        </el-table-column>
        <el-table-column prop="unit" label="单位" width="70" />
        <el-table-column label="优惠后金额" width="120" align="right">
          <template #default="{ row }">{{ row.amount?.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="销售成本" width="120" align="right">
          <template #default="{ row }">{{ row.cost?.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="销售毛利" width="120" align="right">
          <template #default="{ row }">{{ row.profit?.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="毛利率(%)" width="100" align="right">
          <template #default="{ row }">{{ row.profitPct?.toFixed(1) }}</template>
        </el-table-column>
        <el-table-column label="销售占比(%)" width="110" align="right">
          <template #default="{ row }">{{ row.salePct?.toFixed(1) }}</template>
        </el-table-column>
        <el-table-column prop="billCount" label="销售单数" width="90" align="right" />
        <el-table-column prop="code" label="商品编号" width="110" />
        <template #empty>暂无数据</template>
      </el-table>
      <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" layout="total, prev, pager, next" @current-change="load" class="pagination" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useReport } from '@/composables/useReport'

interface Row {
  productId: number; name: string; spec: string; unit: string; code: string
  qty: number; amount: number; cost: number; profit: number
  profitPct: number; salePct: number; billCount: number
}

const { list, total, loading, page, pageSize, filters, load, search, reset, exportCsv } = useReport<Row>(
  '/api/v1/sale-reports/product-stats', { startDate: '', endDate: '', keyword: '' })
const dateRange = ref<[string, string] | null>(null)

function onDateChange() {
  filters.value.startDate = dateRange.value?.[0] || ''
  filters.value.endDate = dateRange.value?.[1] || ''
}
function onReset() { dateRange.value = null; reset() }
function doExport() {
  exportCsv('商品销售统计', [
    { key: 'name', label: '商品名称' }, { key: 'spec', label: '规格' }, { key: 'qty', label: '实销数量' },
    { key: 'unit', label: '单位' }, { key: 'amount', label: '优惠后金额' }, { key: 'cost', label: '销售成本' },
    { key: 'profit', label: '销售毛利' }, { key: 'profitPct', label: '毛利率(%)' },
    { key: 'salePct', label: '销售占比(%)' }, { key: 'billCount', label: '销售单数' }, { key: 'code', label: '商品编号' },
  ])
}
onMounted(() => load(1))
</script>

<style scoped>
.page { display: flex; flex-direction: column; gap: 12px; }
.page-header { display: flex; justify-content: space-between; align-items: center; }
.page-title { font-size: 18px; font-weight: 600; margin: 0; }
.pagination { margin-top: 16px; justify-content: flex-end; }
</style>
