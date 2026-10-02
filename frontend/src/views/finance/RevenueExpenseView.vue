<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">收入费用统计</h2>
      <el-button @click="doExport">导出</el-button>
    </div>
    <el-card>
      <el-form inline>
        <el-form-item label="视图">
          <el-radio-group v-model="filters.view" @change="search">
            <el-radio-button value="expense">按费用</el-radio-button>
            <el-radio-button value="income">按收入</el-radio-button>
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
        <el-table-column prop="code" :label="filters.view === 'expense' ? '费用编码' : '收入编码'" width="120" />
        <el-table-column prop="name" :label="filters.view === 'expense' ? '费用名称' : '收入名称'" min-width="160" />
        <el-table-column label="本期发生额" width="160" align="right"><template #default="{ row }">{{ row.period?.toFixed(2) }}</template></el-table-column>
        <el-table-column label="累计发生额" width="160" align="right"><template #default="{ row }">{{ row.total?.toFixed(2) }}</template></el-table-column>
        <template #empty>暂无数据</template>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useReport } from '@/composables/useReport'

interface Row { itemId: number; code: string; name: string; period: number; total: number }

const { list, loading, filters, load, search, exportCsv } = useReport<Row>(
  '/api/v1/finance-reports/revenue-expense', { startDate: '', endDate: '', view: 'expense' })
const dateRange = ref<[string, string] | null>(null)

function onDateChange() {
  filters.value.startDate = dateRange.value?.[0] || ''
  filters.value.endDate = dateRange.value?.[1] || ''
}
function doExport() {
  const label = filters.value.view === 'expense' ? '费用' : '收入'
  exportCsv('收入费用统计', [
    { key: 'code', label: label + '编码' }, { key: 'name', label: label + '名称' },
    { key: 'period', label: '本期发生额' }, { key: 'total', label: '累计发生额' },
  ])
}
onMounted(() => load(1))
</script>

<style scoped>
.page { display: flex; flex-direction: column; gap: 12px; }
.page-header { display: flex; justify-content: space-between; align-items: center; }
.page-title { font-size: 18px; font-weight: 600; margin: 0; }
</style>
