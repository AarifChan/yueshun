<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">客户预收查询</h2>
      <el-button @click="doExport">导出</el-button>
    </div>
    <el-card>
      <el-form inline>
        <el-form-item label="客户名称"><el-input v-model="filters.customerKeyword" placeholder="客户名称/编号" clearable /></el-form-item>
        <el-form-item>
          <el-button type="primary" @click="search">查询</el-button>
          <el-button @click="reset">清空</el-button>
        </el-form-item>
      </el-form>
      <el-table :data="list" v-loading="loading" border stripe show-summary>
        <el-table-column prop="code" label="客户编号" width="110" />
        <el-table-column prop="name" label="客户名称" min-width="150" />
        <el-table-column label="期初预收" width="120" align="right"><template #default="{ row }">{{ row.initAdvance?.toFixed(2) }}</template></el-table-column>
        <el-table-column label="本期预收" width="120" align="right"><template #default="{ row }">{{ row.periodAdvance?.toFixed(2) }}</template></el-table-column>
        <el-table-column label="预收余额" width="120" align="right"><template #default="{ row }">{{ row.advanceBal?.toFixed(2) }}</template></el-table-column>
        <template #empty>暂无数据</template>
      </el-table>
      <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" layout="total, prev, pager, next" @current-change="load" class="pagination" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useReport } from '@/composables/useReport'

interface Row {
  customerId: number; code: string; name: string
  initAdvance: number; periodAdvance: number; advanceBal: number
}

const { list, total, loading, page, pageSize, filters, load, search, reset, exportCsv } = useReport<Row>(
  '/api/v1/finance-reports/customer-advance', { customerKeyword: '' })

function doExport() {
  exportCsv('客户预收查询', [
    { key: 'code', label: '客户编号' }, { key: 'name', label: '客户名称' },
    { key: 'initAdvance', label: '期初预收' }, { key: 'periodAdvance', label: '本期预收' },
    { key: 'advanceBal', label: '预收余额' },
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
