<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">商品近效期预警</h2>
      <div>
        <el-button @click="doExport">导出</el-button>
        <el-button type="primary" @click="search">查询</el-button>
      </div>
    </div>
    <el-card>
      <el-form inline class="search-form">
        <el-form-item label="商品/批号">
          <el-input v-model="filters.keyword" placeholder="批号/商品名称/编号/条码/规格" clearable style="width: 220px" />
        </el-form-item>
        <el-form-item label="仓库">
          <RemoteSelect v-model="filters.warehouseId" api-url="/api/v1/warehouses" placeholder="全部" />
        </el-form-item>
        <el-form-item label="预警天数">
          <el-input-number v-model="filters.days" :min="1" :max="365" style="width: 110px" />
        </el-form-item>
      </el-form>
      <el-table :data="list" v-loading="loading" border stripe>
        <el-table-column prop="productName" label="商品名称" min-width="150" fixed="left" />
        <el-table-column prop="specification" label="规格" width="100" />
        <el-table-column prop="batchNo" label="批号" width="110" />
        <el-table-column prop="produceDate" label="生产日期" width="100" />
        <el-table-column prop="shelfLifeDays" label="保质期(天)" width="90" align="right" />
        <el-table-column prop="expiryDate" label="到期日期" width="100" />
        <el-table-column label="剩余有效天数" width="110" align="right">
          <template #default="{ row }">
            <span :class="{ 'warn-text': row.remainDays <= 30 }">{{ row.remainDays }}</span>
          </template>
        </el-table-column>
        <el-table-column label="过期天数" width="90" align="right">
          <template #default="{ row }">
            <span v-if="row.expiredDays > 0" class="expired-text">{{ row.expiredDays }}</span>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="unit" label="单位" width="64" />
        <el-table-column prop="warehouseName" label="仓库" width="100" />
        <el-table-column prop="quantity" label="数量" width="90" align="right" />
        <el-table-column label="成本价" width="90" align="right">
          <template #default="{ row }">{{ row.costPrice?.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="成本金额" width="110" align="right">
          <template #default="{ row }">{{ row.costAmount?.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column prop="supplierName" label="批次供应商" width="110" />
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
  id: number; productName: string; specification: string; batchNo: string
  produceDate?: string; shelfLifeDays: number; expiryDate?: string
  remainDays: number; expiredDays: number
  unit: string; warehouseName: string; quantity: number; costPrice: number; costAmount: number
  supplierName: string; code: string
}

const { list, total, loading, page, pageSize, filters, load, search, exportCsv } = useReport<Row>(
  '/api/v1/inventory-reports/expiry-warning',
  { keyword: '', warehouseId: undefined, days: 90 },
)

function doExport() {
  exportCsv('商品近效期预警', [
    { key: 'productName', label: '商品名称' }, { key: 'batchNo', label: '批号' },
    { key: 'produceDate', label: '生产日期' }, { key: 'shelfLifeDays', label: '保质期' },
    { key: 'expiryDate', label: '到期日期' }, { key: 'remainDays', label: '剩余有效天数' },
    { key: 'expiredDays', label: '过期天数' }, { key: 'warehouseName', label: '仓库' },
    { key: 'quantity', label: '数量' }, { key: 'costAmount', label: '成本金额' },
  ])
}

onMounted(() => load(1))
</script>

<style scoped>
.page { display: flex; flex-direction: column; gap: 12px; }
.page-header { display: flex; justify-content: space-between; align-items: center; }
.page-title { font-size: 18px; font-weight: 600; margin: 0; }
.warn-text { color: var(--el-color-warning); font-weight: 600; }
.expired-text { color: var(--el-color-danger); font-weight: 600; }
.pagination { margin-top: 16px; justify-content: flex-end; }
</style>
