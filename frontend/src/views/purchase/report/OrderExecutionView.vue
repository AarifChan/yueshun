<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">采购订单执行明细</h2>
      <el-button @click="doExport">导出</el-button>
    </div>
    <el-card>
      <el-form inline>
        <el-form-item label="时间">
          <el-date-picker v-model="dateRange" type="daterange" value-format="YYYY-MM-DD" start-placeholder="开始时间" end-placeholder="结束时间" @change="onDateChange" />
        </el-form-item>
        <el-form-item label="商品/单号"><el-input v-model="filters.keyword" placeholder="商品名称/单据编号" clearable /></el-form-item>
        <el-form-item label="供应商"><el-input v-model="filters.supplierKeyword" placeholder="供应商" clearable /></el-form-item>
        <el-form-item label="单据状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 120px">
            <el-option label="已确认" value="confirmed" />
            <el-option label="部分入库" value="partial" />
            <el-option label="已完成" value="completed" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="search">查询</el-button>
          <el-button @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>
      <el-table :data="list" v-loading="loading" border stripe>
        <el-table-column prop="orderNo" label="单据编号" width="160" fixed="left" />
        <el-table-column label="录单时间" width="160">
          <template #default="{ row }">{{ row.createdAt?.replace('T', ' ').slice(0, 19) }}</template>
        </el-table-column>
        <el-table-column prop="supplier" label="供应商名称" width="130" />
        <el-table-column prop="statusLabel" label="单据状态" width="90" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 'completed' ? 'success' : 'warning'">{{ row.statusLabel }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="product" label="商品名称" min-width="130" />
        <el-table-column prop="spec" label="规格" width="100" />
        <el-table-column prop="warehouse" label="入库仓库" width="100" />
        <el-table-column prop="unit" label="单位" width="70" />
        <el-table-column label="订货数量" width="90" align="right">
          <template #default="{ row }">{{ row.orderQty }}</template>
        </el-table-column>
        <el-table-column label="订货单价" width="100" align="right">
          <template #default="{ row }">{{ row.price?.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="订货金额" width="110" align="right">
          <template #default="{ row }">{{ row.amount?.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="入库数量" width="90" align="right">
          <template #default="{ row }">{{ row.inQty }}</template>
        </el-table-column>
        <el-table-column label="待入库数量" width="100" align="right">
          <template #default="{ row }">{{ row.pendingQty }}</template>
        </el-table-column>
        <template #empty>暂无搜索结果</template>
      </el-table>
      <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" layout="total, prev, pager, next" @current-change="load" class="pagination" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useReport } from '@/composables/useReport'

interface Row {
  orderNo: string; createdAt: string; supplier: string; status: string; statusLabel: string
  product: string; spec: string; warehouse: string; unit: string
  orderQty: number; price: number; amount: number; inQty: number; pendingQty: number
}

const { list, total, loading, page, pageSize, filters, load, search, reset, exportCsv } = useReport<Row>(
  '/api/v1/purchase-reports/order-execution',
  { startDate: '', endDate: '', keyword: '', supplierKeyword: '', status: '' })
const dateRange = ref<[string, string] | null>(null)

function onDateChange() {
  filters.value.startDate = dateRange.value?.[0] || ''
  filters.value.endDate = dateRange.value?.[1] || ''
}
function onReset() { dateRange.value = null; reset() }
function doExport() {
  exportCsv('采购订单执行明细', [
    { key: 'orderNo', label: '单据编号' }, { key: 'createdAt', label: '录单时间' },
    { key: 'supplier', label: '供应商名称' }, { key: 'statusLabel', label: '单据状态' },
    { key: 'product', label: '商品名称' }, { key: 'spec', label: '规格' },
    { key: 'warehouse', label: '入库仓库' }, { key: 'unit', label: '单位' },
    { key: 'orderQty', label: '订货数量' }, { key: 'price', label: '订货单价' },
    { key: 'amount', label: '订货金额' }, { key: 'inQty', label: '入库数量' },
    { key: 'pendingQty', label: '待入库数量' },
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
