<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">批号跟踪详情</h2>
      <div>
        <el-button @click="doExport">导出</el-button>
        <el-button type="primary" @click="search">查询</el-button>
      </div>
    </div>
    <el-card>
      <el-form inline class="search-form">
        <el-form-item label="时间">
          <el-date-picker v-model="dateRange" type="daterange" value-format="YYYY-MM-DD" start-placeholder="开始时间" end-placeholder="结束时间" style="width: 260px" @change="onDateChange" />
        </el-form-item>
        <el-form-item label="批号"><el-input v-model="filters.batchNo" placeholder="请输入批号" clearable /></el-form-item>
        <el-form-item label="商品"><el-input v-model="filters.keyword" placeholder="商品名称/编号" clearable /></el-form-item>
        <el-form-item label="仓库">
          <RemoteSelect v-model="filters.warehouseId" api-url="/api/v1/warehouses" placeholder="全部" />
        </el-form-item>
      </el-form>
      <el-table :data="list" v-loading="loading" border stripe show-summary>
        <el-table-column prop="billType" label="单据类型" width="100" />
        <el-table-column prop="counterpart" label="往来单位" width="120" />
        <el-table-column prop="handlerName" label="经手人" width="90" />
        <el-table-column prop="productName" label="商品名称" min-width="150" />
        <el-table-column prop="specification" label="规格" width="100" />
        <el-table-column prop="unit" label="单位" width="64" />
        <el-table-column prop="warehouseName" label="仓库" width="100" />
        <el-table-column prop="batchNo" label="批号" width="110" />
        <el-table-column prop="produceDate" label="生产日期" width="100" />
        <el-table-column prop="shelfLifeDays" label="保质期(天)" width="90" align="right" />
        <el-table-column prop="status" label="状态" width="80" align="center" />
        <el-table-column prop="quantity" label="数量" width="90" align="right" />
        <el-table-column label="成本单价" width="90" align="right">
          <template #default="{ row }">{{ row.costPrice?.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="成本金额" width="110" align="right">
          <template #default="{ row }">{{ row.costAmount?.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column prop="billNo" label="单号" width="150" />
        <el-table-column label="录单时间" width="160">
          <template #default="{ row }">{{ row.createdAt?.replace('T', ' ').slice(0, 19) }}</template>
        </el-table-column>
        <template #empty>暂无数据</template>
      </el-table>
      <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" :page-sizes="[30,50,100]" layout="total, sizes, prev, pager, next" @size-change="() => load(1)" @current-change="load" class="pagination" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import RemoteSelect from '@/components/RemoteSelect.vue'
import { useReport } from '@/composables/useReport'

interface Row {
  billType: string; counterpart: string; handlerName: string; productName: string; specification: string
  unit: string; warehouseName: string; batchNo: string; produceDate?: string; shelfLifeDays: number
  status: string; quantity: number; costPrice: number; costAmount: number; billNo: string; createdAt: string
}

const { list, total, loading, page, pageSize, filters, load, search, exportCsv } = useReport<Row>(
  '/api/v1/inventory-reports/batch-trace',
  { batchNo: '', keyword: '', warehouseId: undefined, startDate: '', endDate: '' },
)
const dateRange = ref<[string, string] | null>(null)

function onDateChange() {
  filters.value.startDate = dateRange.value?.[0] || ''
  filters.value.endDate = dateRange.value?.[1] || ''
}

function doExport() {
  exportCsv('批号跟踪详情', [
    { key: 'billType', label: '单据类型' }, { key: 'counterpart', label: '往来单位' },
    { key: 'productName', label: '商品名称' }, { key: 'batchNo', label: '批号' },
    { key: 'quantity', label: '数量' }, { key: 'costPrice', label: '成本单价' }, { key: 'costAmount', label: '成本金额' },
    { key: 'billNo', label: '单号' }, { key: 'createdAt', label: '录单时间' },
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
