<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">单据待结算查询</h2>
      <el-button @click="doExport">导出</el-button>
    </div>
    <el-card>
      <el-form inline>
        <el-form-item label="时间">
          <el-date-picker v-model="dateRange" type="daterange" value-format="YYYY-MM-DD" start-placeholder="开始时间" end-placeholder="结束时间" @change="onDateChange" />
        </el-form-item>
        <el-form-item label="单号/客户"><el-input v-model="filters.keyword" placeholder="单号/客户名称" clearable /></el-form-item>
        <el-form-item>
          <el-button type="primary" @click="search">查询</el-button>
          <el-button @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>
      <el-table :data="list" v-loading="loading" border stripe show-summary>
        <el-table-column prop="billNo" label="单号" width="160" fixed="left" />
        <el-table-column prop="billType" label="单据类型" width="100" />
        <el-table-column prop="customer" label="客户名称" min-width="140" />
        <el-table-column label="单据金额" width="120" align="right"><template #default="{ row }">{{ row.amount?.toFixed(2) }}</template></el-table-column>
        <el-table-column label="已收金额" width="120" align="right"><template #default="{ row }">{{ row.paid?.toFixed(2) }}</template></el-table-column>
        <el-table-column label="未结算金额" width="120" align="right">
          <template #default="{ row }"><span class="debt">{{ row.unsettled?.toFixed(2) }}</span></template>
        </el-table-column>
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
  billNo: string; billType: string; customer: string
  amount: number; paid: number; unsettled: number; createdAt: string
}

const { list, total, loading, page, pageSize, filters, load, search, reset, exportCsv } = useReport<Row>(
  '/api/v1/finance-reports/unsettled-bills', { startDate: '', endDate: '', keyword: '' })
const dateRange = ref<[string, string] | null>(null)

function onDateChange() {
  filters.value.startDate = dateRange.value?.[0] || ''
  filters.value.endDate = dateRange.value?.[1] || ''
}
function onReset() { dateRange.value = null; reset() }
function doExport() {
  exportCsv('单据待结算查询', [
    { key: 'billNo', label: '单号' }, { key: 'billType', label: '单据类型' },
    { key: 'customer', label: '客户名称' }, { key: 'amount', label: '单据金额' },
    { key: 'paid', label: '已收金额' }, { key: 'unsettled', label: '未结算金额' },
    { key: 'createdAt', label: '录单时间' },
  ])
}
onMounted(() => load(1))
</script>

<style scoped>
.page { display: flex; flex-direction: column; gap: 12px; }
.page-header { display: flex; justify-content: space-between; align-items: center; }
.page-title { font-size: 18px; font-weight: 600; margin: 0; }
.debt { color: var(--el-color-danger); font-weight: 600; }
.pagination { margin-top: 16px; justify-content: flex-end; }
</style>
