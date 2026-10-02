<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">商品出入库流水</h2>
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
        <el-form-item label="单据编号"><el-input v-model="filters.billNo" placeholder="单据编号" clearable /></el-form-item>
        <el-form-item label="商品"><el-input v-model="filters.keyword" placeholder="商品名称" clearable /></el-form-item>
        <el-form-item label="单据类型">
          <el-select v-model="filters.billType" placeholder="全部" clearable style="width: 140px">
            <el-option v-for="t in BILL_TYPES" :key="t" :label="t" :value="t" />
          </el-select>
        </el-form-item>
      </el-form>
      <el-table :data="list" v-loading="loading" border stripe show-summary>
        <el-table-column prop="billType" label="单据类型" width="110" />
        <el-table-column prop="productName" label="商品名称" min-width="160" />
        <el-table-column prop="unit" label="单位" width="70" />
        <el-table-column label="批号" width="110">
          <template #default="{ row }">{{ row.batchNo || '-' }}</template>
        </el-table-column>
        <el-table-column prop="inQty" label="入库数量" width="100" align="right">
          <template #default="{ row }">{{ row.inQty || '-' }}</template>
        </el-table-column>
        <el-table-column label="入库金额" width="110" align="right">
          <template #default="{ row }">{{ row.inAmt ? row.inAmt.toFixed(2) : '-' }}</template>
        </el-table-column>
        <el-table-column prop="outQty" label="出库数量" width="100" align="right">
          <template #default="{ row }">{{ row.outQty || '-' }}</template>
        </el-table-column>
        <el-table-column label="出库金额" width="110" align="right">
          <template #default="{ row }">{{ row.outAmt ? row.outAmt.toFixed(2) : '-' }}</template>
        </el-table-column>
        <el-table-column prop="billNo" label="单号" width="160" />
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
import { useReport } from '@/composables/useReport'

const BILL_TYPES = [
  '采购入库单', '销售退货单', '其他入库单', '调拨入库单', '同价调拨单',
  '销售出库单', '采购退货单', '其他出库单', '调拨出库单', '组装拆装单',
]

interface Row {
  billType: string; productId: number; productName: string; unit: string; batchNo?: string
  inQty: number; inAmt: number; outQty: number; outAmt: number; billNo: string; createdAt: string
}

const { list, total, loading, page, pageSize, filters, load, search, exportCsv } = useReport<Row>(
  '/api/v1/inventory-reports/inout-flow',
  { billNo: '', keyword: '', billType: '', startDate: '', endDate: '' },
)
const dateRange = ref<[string, string] | null>(null)

function onDateChange() {
  filters.value.startDate = dateRange.value?.[0] || ''
  filters.value.endDate = dateRange.value?.[1] || ''
}

function doExport() {
  exportCsv('商品出入库流水', [
    { key: 'billType', label: '单据类型' }, { key: 'productName', label: '商品名称' },
    { key: 'unit', label: '单位' }, { key: 'batchNo', label: '批号' },
    { key: 'inQty', label: '入库数量' }, { key: 'inAmt', label: '入库金额' },
    { key: 'outQty', label: '出库数量' }, { key: 'outAmt', label: '出库金额' },
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
