<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">往来余额查询</h2>
      <el-button @click="doExport">导出</el-button>
    </div>
    <el-card>
      <el-form inline>
        <el-form-item label="时间">
          <el-date-picker v-model="dateRange" type="daterange" value-format="YYYY-MM-DD" start-placeholder="开始时间" end-placeholder="结束时间" @change="onDateChange" />
        </el-form-item>
        <el-form-item label="往来单位"><el-input v-model="filters.keyword" placeholder="往来单位名称" clearable /></el-form-item>
        <el-form-item label="单位类型">
          <el-select v-model="filters.unitType" placeholder="全部" clearable style="width: 120px">
            <el-option label="客户" value="customer" />
            <el-option label="供应商" value="supplier" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="search">查询</el-button>
          <el-button @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>
      <el-table :data="list" v-loading="loading" border stripe show-summary>
        <el-table-column prop="name" label="往来单位名称" min-width="150" fixed="left" />
        <el-table-column prop="typeLabel" label="单位类型" width="90" align="center">
          <template #default="{ row }">
            <el-tag :type="row.type === 'customer' ? 'primary' : 'warning'">{{ row.typeLabel }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="应收" width="110" align="right"><template #default="{ row }">{{ row.receivable ? row.receivable.toFixed(2) : '' }}</template></el-table-column>
        <el-table-column label="预收" width="110" align="right"><template #default="{ row }">{{ row.advance ? row.advance.toFixed(2) : '' }}</template></el-table-column>
        <el-table-column label="应付" width="110" align="right"><template #default="{ row }">{{ row.payable ? row.payable.toFixed(2) : '' }}</template></el-table-column>
        <el-table-column label="预付" width="110" align="right"><template #default="{ row }">{{ row.prepay ? row.prepay.toFixed(2) : '' }}</template></el-table-column>
        <el-table-column label="往来余额" width="120" align="right">
          <template #default="{ row }">
            <span :class="{ debt: row.balance > 0 }">{{ row.balance?.toFixed(2) }}</span>
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
  name: string; type: string; typeLabel: string
  receivable: number; advance: number; payable: number; prepay: number; balance: number
}

const { list, total, loading, page, pageSize, filters, load, search, reset, exportCsv } = useReport<Row>(
  '/api/v1/finance-reports/balance', { startDate: '', endDate: '', keyword: '', unitType: '' })
const dateRange = ref<[string, string] | null>(null)

function onDateChange() {
  filters.value.startDate = dateRange.value?.[0] || ''
  filters.value.endDate = dateRange.value?.[1] || ''
}
function onReset() { dateRange.value = null; reset() }
function doExport() {
  exportCsv('往来余额查询', [
    { key: 'name', label: '往来单位名称' }, { key: 'typeLabel', label: '单位类型' },
    { key: 'receivable', label: '应收' }, { key: 'advance', label: '预收' },
    { key: 'payable', label: '应付' }, { key: 'prepay', label: '预付' }, { key: 'balance', label: '往来余额' },
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
