<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">账户收支统计</h2>
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
        <el-form-item label="账户名称"><el-input v-model="filters.accountKeyword" placeholder="账户名称" clearable /></el-form-item>
        <el-form-item>
          <el-button type="primary" @click="search">查询</el-button>
        </el-form-item>
      </el-form>
      <el-table :data="list" v-loading="loading" border stripe show-summary>
        <el-table-column prop="period" label="期间" width="120" fixed="left" />
        <el-table-column prop="account" label="账户名称" min-width="130" />
        <el-table-column label="收入" width="140" align="right"><template #default="{ row }">{{ row.income?.toFixed(2) }}</template></el-table-column>
        <el-table-column label="支出" width="140" align="right"><template #default="{ row }">{{ row.expense?.toFixed(2) }}</template></el-table-column>
        <el-table-column label="收支净额" width="140" align="right">
          <template #default="{ row }"><b>{{ row.net?.toFixed(2) }}</b></template>
        </el-table-column>
        <template #empty>暂无数据</template>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useReport } from '@/composables/useReport'

interface Row { period: string; account: string; income: number; expense: number; net: number }

const { list, loading, filters, load, search, exportCsv } = useReport<Row>(
  '/api/v1/finance-reports/account-income', { startDate: '', endDate: '', accountKeyword: '', view: 'day' })
const dateRange = ref<[string, string] | null>(null)

function onDateChange() {
  filters.value.startDate = dateRange.value?.[0] || ''
  filters.value.endDate = dateRange.value?.[1] || ''
}
function doExport() {
  exportCsv('账户收支统计', [
    { key: 'period', label: '期间' }, { key: 'account', label: '账户名称' },
    { key: 'income', label: '收入' }, { key: 'expense', label: '支出' }, { key: 'net', label: '收支净额' },
  ])
}
onMounted(() => load(1))
</script>

<style scoped>
.page { display: flex; flex-direction: column; gap: 12px; }
.page-header { display: flex; justify-content: space-between; align-items: center; }
.page-title { font-size: 18px; font-weight: 600; margin: 0; }
</style>
