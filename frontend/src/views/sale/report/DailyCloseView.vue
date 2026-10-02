<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">日清日结</h2>
      <el-button @click="doExport">导出</el-button>
    </div>
    <el-card>
      <el-form inline>
        <el-form-item label="日期">
          <el-date-picker v-model="filters.date" type="date" value-format="YYYY-MM-DD" placeholder="选择日期" />
        </el-form-item>
        <el-form-item label="单据编号"><el-input v-model="filters.keyword" placeholder="单据编号" clearable /></el-form-item>
        <el-form-item>
          <el-button type="primary" @click="search">查询</el-button>
        </el-form-item>
      </el-form>
      <el-table :data="list" v-loading="loading" border stripe>
        <el-table-column label="录单时间" width="160">
          <template #default="{ row }">{{ row.createdAt?.replace('T', ' ').slice(0, 19) }}</template>
        </el-table-column>
        <el-table-column prop="billNo" label="单据编号" width="160" />
        <el-table-column prop="billType" label="单据类型" width="110" />
        <el-table-column prop="statusLabel" label="单据状态" width="90" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 'completed' ? 'success' : 'info'">{{ row.statusLabel }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="payStatus" label="支付状态" width="90" align="center" />
        <el-table-column prop="handlerName" label="经手人" width="90" />
        <el-table-column label="单据金额" width="110" align="right">
          <template #default="{ row }">{{ row.amount?.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column v-for="acc in extra.accounts || []" :key="acc" :label="acc" width="110" align="right">
          <template #default="{ row }">{{ row.accounts?.[acc] ? row.accounts[acc].toFixed(2) : '' }}</template>
        </el-table-column>
        <el-table-column label="应收余额" width="110" align="right">
          <template #default="{ row }">{{ row.balance ? row.balance.toFixed(2) : '' }}</template>
        </el-table-column>
        <template #empty>暂无数据</template>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useReport } from '@/composables/useReport'

interface Row {
  billType: string; billNo: string; status: string; statusLabel: string
  handlerName: string; payStatus: string; amount: number; balance: number
  accounts: Record<string, number>; createdAt: string
}

const today = new Date().toISOString().slice(0, 10)
const { list, loading, filters, extra, load, search, exportCsv } = useReport<Row>(
  '/api/v1/sale-reports/daily-close', { date: today, keyword: '' })

function doExport() {
  const cols = [
    { key: 'createdAt', label: '录单时间' }, { key: 'billNo', label: '单据编号' },
    { key: 'billType', label: '单据类型' }, { key: 'statusLabel', label: '单据状态' },
    { key: 'payStatus', label: '支付状态' }, { key: 'handlerName', label: '经手人' },
    { key: 'amount', label: '单据金额' }, { key: 'balance', label: '应收余额' },
  ]
  exportCsv('日清日结', cols)
}
onMounted(() => load(1))
</script>

<style scoped>
.page { display: flex; flex-direction: column; gap: 12px; }
.page-header { display: flex; justify-content: space-between; align-items: center; }
.page-title { font-size: 18px; font-weight: 600; margin: 0; }
</style>
