<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">应收查询</h2>
      <el-button @click="doExport">导出</el-button>
    </div>
    <el-card>
      <el-form inline>
        <el-form-item label="时间">
          <el-date-picker v-model="dateRange" type="daterange" value-format="YYYY-MM-DD" start-placeholder="开始时间" end-placeholder="结束时间" @change="onDateChange" />
        </el-form-item>
        <el-form-item label="客户名称"><el-input v-model="filters.customerKeyword" placeholder="客户名称/编号" clearable /></el-form-item>
        <el-form-item><el-checkbox v-model="filters.includeZero">显示欠款总额为0的客户</el-checkbox></el-form-item>
        <el-form-item>
          <el-button type="primary" @click="search">查询</el-button>
          <el-button @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>
      <el-table :data="list" v-loading="loading" border stripe show-summary>
        <el-table-column prop="code" label="客户编号" width="100" fixed="left" />
        <el-table-column prop="name" label="客户名称" min-width="140" />
        <el-table-column label="期初应收" width="110" align="right"><template #default="{ row }">{{ row.initReceivable?.toFixed(2) }}</template></el-table-column>
        <el-table-column label="本期应收" width="110" align="right"><template #default="{ row }">{{ row.periodSale?.toFixed(2) }}</template></el-table-column>
        <el-table-column label="应收余额" width="110" align="right"><template #default="{ row }">{{ row.receivableBal?.toFixed(2) }}</template></el-table-column>
        <el-table-column label="期初预收" width="110" align="right"><template #default="{ row }">{{ row.initAdvance?.toFixed(2) }}</template></el-table-column>
        <el-table-column label="预收余额" width="110" align="right"><template #default="{ row }">{{ row.advanceBal?.toFixed(2) }}</template></el-table-column>
        <el-table-column label="欠款总额" width="120" align="right">
          <template #default="{ row }">
            <span :class="{ debt: row.debt > 0 }">{{ row.debt?.toFixed(2) }}</span>
          </template>
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
  customerId: number; code: string; name: string
  initReceivable: number; periodSale: number; periodReceipt: number; receivableBal: number
  initAdvance: number; periodAdvance: number; advanceBal: number; debt: number
}

const { list, total, loading, page, pageSize, filters, load, search, reset, exportCsv } = useReport<Row>(
  '/api/v1/finance-reports/receivable', { startDate: '', endDate: '', customerKeyword: '', includeZero: false })
const dateRange = ref<[string, string] | null>(null)

function onDateChange() {
  filters.value.startDate = dateRange.value?.[0] || ''
  filters.value.endDate = dateRange.value?.[1] || ''
}
function onReset() { dateRange.value = null; reset() }
function doExport() {
  exportCsv('应收查询', [
    { key: 'code', label: '客户编号' }, { key: 'name', label: '客户名称' },
    { key: 'initReceivable', label: '期初应收' }, { key: 'periodSale', label: '本期应收' },
    { key: 'receivableBal', label: '应收余额' }, { key: 'initAdvance', label: '期初预收' },
    { key: 'advanceBal', label: '预收余额' }, { key: 'debt', label: '欠款总额' },
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
