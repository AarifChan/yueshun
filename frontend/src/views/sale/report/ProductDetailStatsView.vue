<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">商品销售明细统计</h2>
      <el-button @click="doExport">导出</el-button>
    </div>
    <el-card>
      <el-form inline>
        <el-form-item label="时间">
          <el-date-picker v-model="dateRange" type="daterange" value-format="YYYY-MM-DD" start-placeholder="开始时间" end-placeholder="结束时间" @change="onDateChange" />
        </el-form-item>
        <el-form-item label="商品"><el-input v-model="filters.keyword" placeholder="商品名称" clearable /></el-form-item>
        <el-form-item label="客户"><el-input v-model="filters.customerKeyword" placeholder="客户名称" clearable /></el-form-item>
        <el-form-item>
          <el-button type="primary" @click="search">查询</el-button>
          <el-button @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>
      <el-table :data="list" v-loading="loading" border stripe>
        <el-table-column prop="billType" label="单据类型" width="100" fixed="left" />
        <el-table-column prop="billNo" label="单据编号" width="160" />
        <el-table-column prop="customer" label="客户" width="120" />
        <el-table-column prop="handler" label="经手人" width="90" />
        <el-table-column prop="product" label="商品名称" min-width="130" />
        <el-table-column prop="spec" label="规格" width="100" />
        <el-table-column prop="warehouse" label="出库仓库" width="100" />
        <el-table-column label="销售数量" width="90" align="right">
          <template #default="{ row }">{{ row.qty }}</template>
        </el-table-column>
        <el-table-column prop="unit" label="单位" width="70" />
        <el-table-column label="优惠后单价" width="100" align="right">
          <template #default="{ row }">{{ row.price?.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="优惠后金额" width="110" align="right">
          <template #default="{ row }">{{ row.amount?.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="成本金额" width="110" align="right">
          <template #default="{ row }">{{ row.costAmount?.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="毛利" width="100" align="right">
          <template #default="{ row }">{{ row.profit?.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="毛利率(%)" width="90" align="right">
          <template #default="{ row }">{{ row.profitPct?.toFixed(1) }}</template>
        </el-table-column>
        <el-table-column prop="remark" label="明细备注" width="120" show-overflow-tooltip />
        <el-table-column label="录单时间" width="160">
          <template #default="{ row }">{{ row.createdAt?.replace('T', ' ').slice(0, 19) }}</template>
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
  billType: string; billNo: string; customer: string; handler: string
  product: string; spec: string; warehouse: string; qty: number; unit: string
  price: number; amount: number; costPrice: number; costAmount: number
  profit: number; profitPct: number; remark: string; createdAt: string
}

const { list, total, loading, page, pageSize, filters, load, search, reset, exportCsv } = useReport<Row>(
  '/api/v1/sale-reports/product-detail', { startDate: '', endDate: '', keyword: '', customerKeyword: '' })
const dateRange = ref<[string, string] | null>(null)

function onDateChange() {
  filters.value.startDate = dateRange.value?.[0] || ''
  filters.value.endDate = dateRange.value?.[1] || ''
}
function onReset() { dateRange.value = null; reset() }
function doExport() {
  exportCsv('商品销售明细统计', [
    { key: 'billType', label: '单据类型' }, { key: 'billNo', label: '单据编号' },
    { key: 'customer', label: '客户' }, { key: 'handler', label: '经手人' },
    { key: 'product', label: '商品名称' }, { key: 'spec', label: '规格' },
    { key: 'warehouse', label: '出库仓库' }, { key: 'qty', label: '销售数量' },
    { key: 'unit', label: '单位' }, { key: 'price', label: '优惠后单价' },
    { key: 'amount', label: '优惠后金额' }, { key: 'costAmount', label: '成本金额' },
    { key: 'profit', label: '毛利' }, { key: 'profitPct', label: '毛利率(%)' },
    { key: 'remark', label: '明细备注' }, { key: 'createdAt', label: '录单时间' },
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
