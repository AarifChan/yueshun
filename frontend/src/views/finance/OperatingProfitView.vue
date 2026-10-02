<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">经营利润统计</h2>
      <el-button @click="doExport">导出</el-button>
    </div>
    <el-card>
      <el-form inline>
        <el-form-item label="视图">
          <el-radio-group v-model="filters.view" @change="search">
            <el-radio-button value="day">按日查看</el-radio-button>
            <el-radio-button value="week">按周查看</el-radio-button>
            <el-radio-button value="month">按月查看</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="时间">
          <el-date-picker v-model="dateRange" type="daterange" value-format="YYYY-MM-DD" start-placeholder="开始时间" end-placeholder="结束时间" @change="onDateChange" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="search">查询</el-button>
        </el-form-item>
      </el-form>
      <el-table :data="list" v-loading="loading" border stripe show-summary>
        <el-table-column type="index" label="序" width="60" />
        <el-table-column prop="period" label="日期" width="120" fixed="left" />
        <el-table-column label="销售收入" width="130" align="right"><template #default="{ row }">{{ row.saleIncome?.toFixed(2) }}</template></el-table-column>
        <el-table-column label="其他收入" width="120" align="right"><template #default="{ row }">{{ row.otherIncome?.toFixed(2) }}</template></el-table-column>
        <el-table-column label="收入合计" width="130" align="right"><template #default="{ row }">{{ row.incomeTotal?.toFixed(2) }}</template></el-table-column>
        <el-table-column label="销售成本" width="130" align="right"><template #default="{ row }">{{ row.saleCost?.toFixed(2) }}</template></el-table-column>
        <el-table-column label="费用" width="120" align="right"><template #default="{ row }">{{ row.expense?.toFixed(2) }}</template></el-table-column>
        <el-table-column label="支出合计" width="130" align="right"><template #default="{ row }">{{ row.expenseTotal?.toFixed(2) }}</template></el-table-column>
        <el-table-column label="利润" width="140" align="right">
          <template #default="{ row }">
            <b :class="{ negative: row.profit < 0 }">{{ row.profit?.toFixed(2) }}</b>
          </template>
        </el-table-column>
        <template #empty>暂无数据</template>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useReport } from '@/composables/useReport'

interface Row {
  period: string; saleIncome: number; otherIncome: number; incomeTotal: number
  saleCost: number; expense: number; expenseTotal: number; profit: number
}

const { list, loading, filters, load, search, exportCsv } = useReport<Row>(
  '/api/v1/finance-reports/operating-profit', { startDate: '', endDate: '', view: 'day' })
const dateRange = ref<[string, string] | null>(null)

function onDateChange() {
  filters.value.startDate = dateRange.value?.[0] || ''
  filters.value.endDate = dateRange.value?.[1] || ''
}
function doExport() {
  exportCsv('经营利润统计', [
    { key: 'period', label: '日期' }, { key: 'saleIncome', label: '销售收入' },
    { key: 'otherIncome', label: '其他收入' }, { key: 'incomeTotal', label: '收入合计' },
    { key: 'saleCost', label: '销售成本' }, { key: 'expense', label: '费用' },
    { key: 'expenseTotal', label: '支出合计' }, { key: 'profit', label: '利润' },
  ])
}
onMounted(() => load(1))
</script>

<style scoped>
.page { display: flex; flex-direction: column; gap: 12px; }
.page-header { display: flex; justify-content: space-between; align-items: center; }
.page-title { font-size: 18px; font-weight: 600; margin: 0; }
.negative { color: var(--el-color-danger); }
</style>
