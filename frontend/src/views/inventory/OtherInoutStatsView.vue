<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">其他出入库统计</h2>
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
        <el-form-item label="仓库">
          <RemoteSelect v-model="filters.warehouseId" api-url="/api/v1/warehouses" placeholder="全部" />
        </el-form-item>
      </el-form>
      <el-table :data="list" v-loading="loading" border stripe show-summary>
        <el-table-column prop="code" label="商品编号" width="110" />
        <el-table-column prop="productName" label="商品名称" min-width="160" />
        <el-table-column prop="unit" label="单位" width="70" />
        <el-table-column label="入库合计" align="center">
          <el-table-column prop="inQty" label="数量" width="90" align="right" />
          <el-table-column label="均价" width="90" align="right">
            <template #default="{ row }">{{ row.inPrice?.toFixed(2) }}</template>
          </el-table-column>
          <el-table-column label="金额" width="110" align="right">
            <template #default="{ row }">{{ row.inAmt?.toFixed(2) }}</template>
          </el-table-column>
        </el-table-column>
        <el-table-column label="出库合计" align="center">
          <el-table-column prop="outQty" label="数量" width="90" align="right" />
          <el-table-column label="均价" width="90" align="right">
            <template #default="{ row }">{{ row.outPrice?.toFixed(2) }}</template>
          </el-table-column>
          <el-table-column label="金额" width="110" align="right">
            <template #default="{ row }">{{ row.outAmt?.toFixed(2) }}</template>
          </el-table-column>
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
  productId: number; code: string; productName: string; unit: string
  inQty: number; inPrice: number; inAmt: number; outQty: number; outPrice: number; outAmt: number
}

const { list, total, loading, page, pageSize, filters, load, search, exportCsv } = useReport<Row>(
  '/api/v1/inventory-reports/other-inout-stats',
  { keyword: '', startDate: '', endDate: '', warehouseId: undefined },
)
const dateRange = ref<[string, string] | null>(null)

function onDateChange() {
  filters.value.startDate = dateRange.value?.[0] || ''
  filters.value.endDate = dateRange.value?.[1] || ''
}

function doExport() {
  exportCsv('其他出入库统计', [
    { key: 'code', label: '商品编号' }, { key: 'productName', label: '商品名称' }, { key: 'unit', label: '单位' },
    { key: 'inQty', label: '入库数量' }, { key: 'inPrice', label: '入库均价' }, { key: 'inAmt', label: '入库金额' },
    { key: 'outQty', label: '出库数量' }, { key: 'outPrice', label: '出库均价' }, { key: 'outAmt', label: '出库金额' },
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
