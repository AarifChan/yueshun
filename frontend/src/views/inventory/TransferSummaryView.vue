<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">调拨汇总表</h2>
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
        <el-form-item label="商品"><el-input v-model="filters.keyword" placeholder="商品名称/编号" clearable /></el-form-item>
        <el-form-item label="调拨类型">
          <el-select v-model="filters.transferType" placeholder="全部" clearable style="width: 120px">
            <el-option label="同价调拨" value="same" /><el-option label="变价调拨" value="diff" />
          </el-select>
        </el-form-item>
        <el-form-item label="调出仓库">
          <RemoteSelect v-model="filters.fromWarehouseId" api-url="/api/v1/warehouses" placeholder="全部" />
        </el-form-item>
        <el-form-item label="调入仓库">
          <RemoteSelect v-model="filters.toWarehouseId" api-url="/api/v1/warehouses" placeholder="全部" />
        </el-form-item>
      </el-form>
      <el-table :data="list" v-loading="loading" border stripe show-summary>
        <el-table-column prop="productName" label="商品名称" min-width="150" fixed="left" />
        <el-table-column prop="specification" label="规格" width="100" />
        <el-table-column prop="unit" label="单位" width="64" />
        <el-table-column prop="outQty" label="调出数量" width="90" align="right" />
        <el-table-column prop="inQty" label="调入数量" width="90" align="right" />
        <el-table-column label="调出成本均价" width="110" align="right">
          <template #default="{ row }">{{ row.outCostPrice?.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="调出成本金额" width="120" align="right">
          <template #default="{ row }">{{ row.outCostAmt?.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="调入均价" width="100" align="right">
          <template #default="{ row }">{{ row.inPrice?.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="调入金额" width="110" align="right">
          <template #default="{ row }">{{ row.inAmt?.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="差异金额" width="100" align="right">
          <template #default="{ row }">{{ row.diffAmt?.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column prop="code" label="商品编号" width="110" />
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
  productId: number; productName: string; specification: string; unit: string; code: string
  outQty: number; inQty: number; outCostPrice: number; outCostAmt: number; inPrice: number; inAmt: number; diffAmt: number
}

const { list, total, loading, page, pageSize, filters, load, search, exportCsv } = useReport<Row>(
  '/api/v1/inventory-reports/transfer-summary',
  { keyword: '', startDate: '', endDate: '', transferType: '', fromWarehouseId: undefined, toWarehouseId: undefined },
)
const dateRange = ref<[string, string] | null>(null)

function onDateChange() {
  filters.value.startDate = dateRange.value?.[0] || ''
  filters.value.endDate = dateRange.value?.[1] || ''
}

function doExport() {
  exportCsv('调拨汇总表', [
    { key: 'productName', label: '商品名称' }, { key: 'specification', label: '规格' }, { key: 'unit', label: '单位' },
    { key: 'outQty', label: '调出数量' }, { key: 'inQty', label: '调入数量' },
    { key: 'outCostPrice', label: '调出成本均价' }, { key: 'outCostAmt', label: '调出成本金额' },
    { key: 'inPrice', label: '调入均价' }, { key: 'inAmt', label: '调入金额' }, { key: 'diffAmt', label: '差异金额' },
    { key: 'code', label: '商品编号' },
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
