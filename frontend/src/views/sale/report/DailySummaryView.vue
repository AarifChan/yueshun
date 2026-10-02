<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">日销售统计</h2>
      <el-button @click="doExport">导出</el-button>
    </div>
    <el-card>
      <el-form inline>
        <el-form-item label="时间">
          <el-date-picker v-model="dateRange" type="daterange" value-format="YYYY-MM-DD" start-placeholder="开始时间" end-placeholder="结束时间" @change="onDateChange" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="search">查询</el-button>
          <el-button @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>
      <el-table :data="list" v-loading="loading" border stripe show-summary>
        <el-table-column prop="day" label="日期" width="120" fixed="left" />
        <el-table-column label="销售金额" width="120" align="right">
          <template #default="{ row }">{{ row.saleAmount?.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="退货金额" width="120" align="right">
          <template #default="{ row }">{{ row.returnAmount?.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="销售总额" width="120" align="right">
          <template #default="{ row }">{{ row.totalAmount?.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="收款总额" width="120" align="right">
          <template #default="{ row }">{{ row.receipt?.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="欠款总额" width="120" align="right">
          <template #default="{ row }">{{ row.debt?.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="优惠金额" width="110" align="right">
          <template #default="{ row }">{{ row.discount?.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="利润" width="120" align="right">
          <template #default="{ row }">{{ row.profit?.toFixed(2) }}</template>
        </el-table-column>
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
  day: string; saleAmount: number; returnAmount: number; totalAmount: number
  receipt: number; debt: number; discount: number; profit: number
}

const { list, total, loading, page, pageSize, filters, load, search, reset, exportCsv } = useReport<Row>(
  '/api/v1/sale-reports/daily-summary', { startDate: '', endDate: '' })
const dateRange = ref<[string, string] | null>(null)

function onDateChange() {
  filters.value.startDate = dateRange.value?.[0] || ''
  filters.value.endDate = dateRange.value?.[1] || ''
}
function onReset() { dateRange.value = null; reset() }
function doExport() {
  exportCsv('日销售统计', [
    { key: 'day', label: '日期' }, { key: 'saleAmount', label: '销售金额' },
    { key: 'returnAmount', label: '退货金额' }, { key: 'totalAmount', label: '销售总额' },
    { key: 'receipt', label: '收款总额' }, { key: 'debt', label: '欠款总额' },
    { key: 'discount', label: '优惠金额' }, { key: 'profit', label: '利润' },
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
