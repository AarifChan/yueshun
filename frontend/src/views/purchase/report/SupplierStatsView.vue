<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">供应商采购统计</h2>
      <el-button @click="doExport">导出</el-button>
    </div>
    <el-card>
      <el-form inline>
        <el-form-item label="时间">
          <el-date-picker v-model="dateRange" type="daterange" value-format="YYYY-MM-DD" start-placeholder="开始时间" end-placeholder="结束时间" @change="onDateChange" />
        </el-form-item>
        <el-form-item label="供应商"><el-input v-model="filters.keyword" placeholder="供应商名称" clearable /></el-form-item>
        <el-form-item label="商品"><el-input v-model="filters.productKeyword" placeholder="商品名称" clearable /></el-form-item>
        <el-form-item><el-checkbox v-model="filters.includeZero">显示实购数量为0的供应商</el-checkbox></el-form-item>
        <el-form-item>
          <el-button type="primary" @click="search">查询</el-button>
          <el-button @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>
      <el-table :data="list" v-loading="loading" border stripe show-summary>
        <el-table-column prop="name" label="供应商名称" min-width="160" fixed="left" />
        <el-table-column label="实购数量" width="110" align="right">
          <template #default="{ row }">{{ row.qty }}</template>
        </el-table-column>
        <el-table-column label="折后金额" width="130" align="right">
          <template #default="{ row }">{{ row.amount?.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="赠品数量" width="100" align="right">
          <template #default="{ row }">{{ row.giftQty }}</template>
        </el-table-column>
        <template #empty>暂无搜索结果</template>
      </el-table>
      <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" layout="total, prev, pager, next" @current-change="load" class="pagination" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useReport } from '@/composables/useReport'

interface Row {
  supplierId: number; name: string; qty: number; amount: number; giftQty: number
}

const { list, total, loading, page, pageSize, filters, load, search, reset, exportCsv } = useReport<Row>(
  '/api/v1/purchase-reports/supplier-stats',
  { startDate: '', endDate: '', keyword: '', productKeyword: '', includeZero: false })
const dateRange = ref<[string, string] | null>(null)

function onDateChange() {
  filters.value.startDate = dateRange.value?.[0] || ''
  filters.value.endDate = dateRange.value?.[1] || ''
}
function onReset() { dateRange.value = null; reset() }
function doExport() {
  exportCsv('供应商采购统计', [
    { key: 'name', label: '供应商名称' }, { key: 'qty', label: '实购数量' },
    { key: 'amount', label: '折后金额' }, { key: 'giftQty', label: '赠品数量' },
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
