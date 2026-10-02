<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">库存批号统计</h2>
      <div>
        <el-button @click="doExport">导出</el-button>
        <el-button type="primary" @click="search">查询</el-button>
      </div>
    </div>
    <el-card>
      <el-form inline class="search-form">
        <el-form-item label="批号"><el-input v-model="filters.batchNo" placeholder="批号/批次号" clearable /></el-form-item>
        <el-form-item label="商品"><el-input v-model="filters.keyword" placeholder="商品名称/编号" clearable /></el-form-item>
        <el-form-item label="仓库">
          <RemoteSelect v-model="filters.warehouseId" api-url="/api/v1/warehouses" placeholder="全部" />
        </el-form-item>
        <el-form-item><el-checkbox v-model="filters.onlyNoBatch">仅显示无批次号的批次</el-checkbox></el-form-item>
      </el-form>
      <el-table :data="list" v-loading="loading" border stripe show-summary>
        <el-table-column prop="productName" label="商品名称" min-width="150" fixed="left" />
        <el-table-column prop="specification" label="规格" width="100" />
        <el-table-column prop="unit" label="单位" width="64" />
        <el-table-column prop="batchNo" label="批号" width="120" />
        <el-table-column prop="quantity" label="数量" width="90" align="right" />
        <el-table-column label="成本价" width="90" align="right">
          <template #default="{ row }">{{ row.costPrice?.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="成本金额" width="110" align="right">
          <template #default="{ row }">{{ row.costAmount?.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column prop="produceDate" label="生产日期" width="110" />
        <el-table-column prop="expiryDate" label="到期日期" width="110" />
        <el-table-column prop="shelfLifeDays" label="保质期(天)" width="90" align="right" />
        <el-table-column prop="inDate" label="入库日期" width="110" />
        <el-table-column prop="warehouseName" label="仓库" width="110" />
        <el-table-column prop="supplierName" label="批次供应商" width="120" />
        <el-table-column prop="code" label="商品编号" width="110" />
        <template #empty>暂无数据</template>
      </el-table>
      <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" :page-sizes="[30,50,100]" layout="total, sizes, prev, pager, next" @size-change="() => load(1)" @current-change="load" class="pagination" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import RemoteSelect from '@/components/RemoteSelect.vue'
import { useReport } from '@/composables/useReport'

interface Row {
  id: number; productName: string; specification: string; unit: string; batchNo: string
  quantity: number; costPrice: number; costAmount: number
  produceDate?: string; expiryDate?: string; shelfLifeDays: number; inDate?: string
  warehouseName: string; supplierName: string; code: string
}

const { list, total, loading, page, pageSize, filters, load, search, exportCsv } = useReport<Row>(
  '/api/v1/inventory-reports/batch-stats',
  { batchNo: '', keyword: '', warehouseId: undefined, onlyNoBatch: false },
)

function doExport() {
  exportCsv('库存批号统计', [
    { key: 'productName', label: '商品名称' }, { key: 'specification', label: '规格' }, { key: 'unit', label: '单位' },
    { key: 'batchNo', label: '批号' }, { key: 'quantity', label: '数量' },
    { key: 'costPrice', label: '成本价' }, { key: 'costAmount', label: '成本金额' },
    { key: 'produceDate', label: '生产日期' }, { key: 'expiryDate', label: '到期日期' },
    { key: 'shelfLifeDays', label: '保质期' }, { key: 'inDate', label: '入库日期' },
    { key: 'warehouseName', label: '仓库' }, { key: 'supplierName', label: '批次供应商' }, { key: 'code', label: '商品编号' },
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
