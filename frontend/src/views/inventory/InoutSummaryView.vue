<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">商品进销存汇总</h2>
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
        <el-form-item label="商品">
          <el-input v-model="filters.keyword" placeholder="商品名称/编号/条码/规格" clearable style="width: 220px" />
        </el-form-item>
      </el-form>
      <el-table :data="list" v-loading="loading" border stripe>
        <el-table-column prop="name" label="商品名称" min-width="150" fixed="left" />
        <el-table-column prop="specification" label="规格" width="100" />
        <el-table-column prop="unit" label="单位" width="64" />
        <el-table-column label="此前" align="center">
          <el-table-column prop="priorQty" label="数量" width="90" align="right" />
          <el-table-column label="金额" width="100" align="right">
            <template #default="{ row }">{{ row.priorAmt?.toFixed(2) }}</template>
          </el-table-column>
        </el-table-column>
        <el-table-column label="本期入库" align="center">
          <el-table-column v-for="t in IN_TYPES" :key="t" :label="t" width="100" align="right">
            <template #default="{ row }">{{ row.inByType?.[t]?.Qty ?? 0 }}</template>
          </el-table-column>
          <el-table-column label="入库合计" width="100" align="right">
            <template #default="{ row }">{{ row.periodInQty }}</template>
          </el-table-column>
        </el-table-column>
        <el-table-column label="本期出库" align="center">
          <el-table-column v-for="t in OUT_TYPES" :key="t" :label="t" width="100" align="right">
            <template #default="{ row }">{{ row.outByType?.[t]?.Qty ?? 0 }}</template>
          </el-table-column>
          <el-table-column label="出库合计" width="100" align="right">
            <template #default="{ row }">{{ row.periodOutQty }}</template>
          </el-table-column>
        </el-table-column>
        <el-table-column label="结存" align="center" fixed="right">
          <el-table-column prop="balanceQty" label="数量" width="90" align="right" />
          <el-table-column label="金额" width="110" align="right">
            <template #default="{ row }">{{ row.balanceAmt?.toFixed(2) }}</template>
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
import { useReport } from '@/composables/useReport'

const IN_TYPES = ['采购入库单', '销售退货单', '其他入库单', '同价调拨单', '变价调拨单', '组装拆装单']
const OUT_TYPES = ['销售出库单', '采购退货单', '其他出库单', '同价调拨单', '变价调拨单', '组装拆装单']

interface Bucket { Qty: number; Amt: number }
interface Row {
  id: number; name: string; code: string; specification: string; unit: string
  priorQty: number; priorAmt: number
  inByType?: Record<string, Bucket>; outByType?: Record<string, Bucket>
  periodInQty: number; periodInAmt: number; periodOutQty: number; periodOutAmt: number
  balanceQty: number; balanceAmt: number
}

const { list, total, loading, page, pageSize, filters, load, search, exportCsv } = useReport<Row>(
  '/api/v1/inventory-reports/inout-summary',
  { keyword: '', startDate: '', endDate: '' },
)
const dateRange = ref<[string, string] | null>(null)

function onDateChange() {
  filters.value.startDate = dateRange.value?.[0] || ''
  filters.value.endDate = dateRange.value?.[1] || ''
}

function doExport() {
  exportCsv('商品进销存汇总', [
    { key: 'name', label: '商品名称' }, { key: 'specification', label: '规格' }, { key: 'unit', label: '单位' },
    { key: 'priorQty', label: '此前数量' }, { key: 'priorAmt', label: '此前金额' },
    { key: 'periodInQty', label: '本期入库数量' }, { key: 'periodInAmt', label: '本期入库金额' },
    { key: 'periodOutQty', label: '本期出库数量' }, { key: 'periodOutAmt', label: '本期出库金额' },
    { key: 'balanceQty', label: '结存数量' }, { key: 'balanceAmt', label: '结存金额' },
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
