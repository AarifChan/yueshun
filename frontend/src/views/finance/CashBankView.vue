<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">现金银行账户统计</h2>
      <el-button @click="doExport">导出</el-button>
    </div>
    <el-card>
      <el-form inline>
        <el-form-item label="时间">
          <el-date-picker v-model="dateRange" type="daterange" value-format="YYYY-MM-DD" start-placeholder="开始时间" end-placeholder="结束时间" @change="onDateChange" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="search">查询</el-button>
        </el-form-item>
      </el-form>
      <el-table :data="list" v-loading="loading" border stripe show-summary>
        <el-table-column prop="name" label="账户名称" min-width="140" fixed="left" />
        <el-table-column label="类型" width="90">
          <template #default="{ row }">{{ typeLabel(row.type) }}</template>
        </el-table-column>
        <el-table-column label="期初余额" width="130" align="right"><template #default="{ row }">{{ row.initBal?.toFixed(2) }}</template></el-table-column>
        <el-table-column label="本期收入" width="130" align="right"><template #default="{ row }">{{ row.income?.toFixed(2) }}</template></el-table-column>
        <el-table-column label="本期支出" width="130" align="right"><template #default="{ row }">{{ row.expense?.toFixed(2) }}</template></el-table-column>
        <el-table-column label="期末余额" width="130" align="right">
          <template #default="{ row }"><b>{{ row.finalBal?.toFixed(2) }}</b></template>
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
  accountId: number; name: string; type: string
  initBal: number; income: number; expense: number; finalBal: number
}

const { list, loading, filters, load, search, exportCsv } = useReport<Row>(
  '/api/v1/finance-reports/cash-bank', { startDate: '', endDate: '' })
const dateRange = ref<[string, string] | null>(null)

function typeLabel(t: string) {
  return t === 'cash' ? '现金' : t === 'bank' ? '银行' : t === 'alipay' ? '支付宝' : t === 'wechat' ? '微信' : t
}
function onDateChange() {
  filters.value.startDate = dateRange.value?.[0] || ''
  filters.value.endDate = dateRange.value?.[1] || ''
}
function doExport() {
  exportCsv('现金银行账户统计', [
    { key: 'name', label: '账户名称' }, { key: 'initBal', label: '期初余额' },
    { key: 'income', label: '本期收入' }, { key: 'expense', label: '本期支出' }, { key: 'finalBal', label: '期末余额' },
  ])
}
onMounted(() => load(1))
</script>

<style scoped>
.page { display: flex; flex-direction: column; gap: 12px; }
.page-header { display: flex; justify-content: space-between; align-items: center; }
.page-title { font-size: 18px; font-weight: 600; margin: 0; }
</style>
