<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">订销退综合分析</h2>
      <el-button @click="doExport">导出</el-button>
    </div>
    <el-card>
      <el-form inline>
        <el-form-item label="时间">
          <el-date-picker v-model="dateRange" type="daterange" value-format="YYYY-MM-DD" start-placeholder="开始时间" end-placeholder="结束时间" @change="onDateChange" />
        </el-form-item>
        <el-form-item label="商品"><el-input v-model="filters.keyword" placeholder="商品名称/编号" clearable /></el-form-item>
        <el-form-item>
          <el-button type="primary" @click="search">查询</el-button>
          <el-button @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>
      <el-table :data="list" v-loading="loading" border stripe show-summary>
        <el-table-column prop="name" label="商品名称" min-width="130" fixed="left" />
        <el-table-column prop="spec" label="规格" width="100" />
        <el-table-column prop="unit" label="单位" width="70" />
        <el-table-column label="订货数量" width="90" align="right"><template #default="{ row }">{{ row.orderQty }}</template></el-table-column>
        <el-table-column label="订货金额" width="110" align="right"><template #default="{ row }">{{ row.orderAmount?.toFixed(2) }}</template></el-table-column>
        <el-table-column prop="orderCount" label="订单数" width="80" align="right" />
        <el-table-column label="销售数量" width="90" align="right"><template #default="{ row }">{{ row.saleQty }}</template></el-table-column>
        <el-table-column label="销售金额" width="110" align="right"><template #default="{ row }">{{ row.saleAmount?.toFixed(2) }}</template></el-table-column>
        <el-table-column prop="saleCount" label="销售单数" width="90" align="right" />
        <el-table-column label="退货数量" width="90" align="right"><template #default="{ row }">{{ row.returnQty }}</template></el-table-column>
        <el-table-column label="退货金额" width="110" align="right"><template #default="{ row }">{{ row.returnAmount?.toFixed(2) }}</template></el-table-column>
        <el-table-column prop="returnCount" label="退单数" width="80" align="right" />
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
  orderQty: number; orderAmount: number; orderCount: number
  saleQty: number; saleAmount: number; saleCount: number
  returnQty: number; returnAmount: number; returnCount: number
}

const { list, total, loading, page, pageSize, filters, load, search, reset, exportCsv } = useReport<Row>(
  '/api/v1/sale-reports/comprehensive', { startDate: '', endDate: '', keyword: '' })
const dateRange = ref<[string, string] | null>(null)

function onDateChange() {
  filters.value.startDate = dateRange.value?.[0] || ''
  filters.value.endDate = dateRange.value?.[1] || ''
}
function onReset() { dateRange.value = null; reset() }
function doExport() {
  exportCsv('订销退综合分析', [
    { key: 'name', label: '商品名称' }, { key: 'spec', label: '规格' }, { key: 'unit', label: '单位' },
    { key: 'orderQty', label: '订货数量' }, { key: 'orderAmount', label: '订货金额' }, { key: 'orderCount', label: '订单数' },
    { key: 'saleQty', label: '销售数量' }, { key: 'saleAmount', label: '销售金额' }, { key: 'saleCount', label: '销售单数' },
    { key: 'returnQty', label: '退货数量' }, { key: 'returnAmount', label: '退货金额' }, { key: 'returnCount', label: '退单数' },
    { key: 'code', label: '商品编号' },
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
