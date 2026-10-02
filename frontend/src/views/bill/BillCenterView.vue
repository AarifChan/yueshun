<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">单据中心</h2>
      <div>
        <el-button @click="doExport">导出</el-button>
        <el-button type="primary" @click="search">查询</el-button>
      </div>
    </div>
    <div class="body">
      <div class="sidebar">
        <div class="sidebar-title">单据目录</div>
        <el-menu :default-active="filters.billType" @select="onTypeSelect">
          <el-menu-item index="">全部单据</el-menu-item>
          <el-menu-item-group title="采购">
            <el-menu-item index="采购订单">采购订单</el-menu-item>
            <el-menu-item index="采购入库单">采购入库单</el-menu-item>
            <el-menu-item index="采购退货单">采购退货单</el-menu-item>
          </el-menu-item-group>
          <el-menu-item-group title="销售">
            <el-menu-item index="销售订单">销售订单</el-menu-item>
            <el-menu-item index="销售出库单">销售出库单</el-menu-item>
            <el-menu-item index="销售退货单">销售退货单</el-menu-item>
          </el-menu-item-group>
          <el-menu-item-group title="库存">
            <el-menu-item index="其他出库单">其他出库单</el-menu-item>
            <el-menu-item index="其他入库单">其他入库单</el-menu-item>
            <el-menu-item index="同价调拨单">同价调拨单</el-menu-item>
            <el-menu-item index="调拨出库单(同价)">调拨出库单(同价)</el-menu-item>
            <el-menu-item index="调拨入库单(同价)">调拨入库单(同价)</el-menu-item>
            <el-menu-item index="调拨出库单(变价)">调拨出库单(变价)</el-menu-item>
            <el-menu-item index="调拨入库单(变价)">调拨入库单(变价)</el-menu-item>
            <el-menu-item index="成本调价单">成本调价单</el-menu-item>
            <el-menu-item index="组装拆装单">组装拆装单</el-menu-item>
          </el-menu-item-group>
        </el-menu>
      </div>
      <el-card class="main">
        <el-form inline class="search-form">
          <el-form-item label="时间">
            <el-date-picker v-model="dateRange" type="daterange" value-format="YYYY-MM-DD" start-placeholder="开始时间" end-placeholder="结束时间" style="width: 260px" @change="onDateChange" />
          </el-form-item>
          <el-form-item label="单据编号"><el-input v-model="filters.keyword" placeholder="单据编号" clearable /></el-form-item>
          <el-form-item label="往来单位"><el-input v-model="filters.counterpart" placeholder="往来单位" clearable /></el-form-item>
          <el-form-item label="单据状态">
            <el-select v-model="filters.status" placeholder="全部" clearable style="width: 120px">
              <el-option label="草稿" value="draft" />
              <el-option label="待审核" value="pending" />
              <el-option label="已过账" value="completed" />
              <el-option label="已撤销" value="cancelled" />
            </el-select>
          </el-form-item>
          <el-form-item label="支付状态">
            <el-select v-model="filters.payStatus" placeholder="全部" clearable style="width: 120px">
              <el-option label="未支付" value="未支付" />
              <el-option label="部分支付" value="部分支付" />
              <el-option label="已支付" value="已支付" />
            </el-select>
          </el-form-item>
          <el-form-item><el-checkbox v-model="filters.onlyDraft">只显示草稿单</el-checkbox></el-form-item>
        </el-form>
        <el-table :data="list" v-loading="loading" border stripe height="100%">
          <el-table-column prop="billType" label="单据类型" width="130" fixed="left" />
          <el-table-column prop="statusLabel" label="单据状态" width="90" align="center">
            <template #default="{ row }">
              <el-tag :type="row.status === 'completed' ? 'success' : row.status === 'draft' ? 'info' : 'warning'">{{ row.statusLabel }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="counterpart" label="往来单位" width="130" />
          <el-table-column prop="handlerName" label="经手人" width="90" />
          <el-table-column prop="quantity" label="数量" width="90" align="right" />
          <el-table-column prop="inWarehouse" label="入库仓库" width="110" />
          <el-table-column prop="outWarehouse" label="出库仓库" width="110" />
          <el-table-column label="单据金额" width="110" align="right">
            <template #default="{ row }">{{ row.amount?.toFixed(2) }}</template>
          </el-table-column>
          <el-table-column label="收付款金额" width="110" align="right">
            <template #default="{ row }">{{ row.paidAmount?.toFixed(2) }}</template>
          </el-table-column>
          <el-table-column prop="payStatus" label="支付状态" width="90" align="center">
            <template #default="{ row }">
              <el-tag v-if="row.payStatus === '已支付'" type="success">已支付</el-tag>
              <el-tag v-else-if="row.payStatus === '部分支付'" type="warning">部分支付</el-tag>
              <el-tag v-else-if="row.payStatus === '未支付'" type="danger">未支付</el-tag>
              <span v-else>-</span>
            </template>
          </el-table-column>
          <el-table-column prop="remark" label="单据备注" min-width="130" show-overflow-tooltip />
          <el-table-column prop="printCount" label="打印次数" width="90" align="right" />
          <el-table-column prop="billNo" label="单号" width="160" />
          <el-table-column label="录单时间" width="160">
            <template #default="{ row }">{{ row.createdAt?.replace('T', ' ').slice(0, 19) }}</template>
          </el-table-column>
          <template #empty>暂无数据</template>
        </el-table>
        <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" :page-sizes="[30,50,100]" layout="total, sizes, prev, pager, next" @size-change="() => load(1)" @current-change="load" class="pagination" />
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useReport } from '@/composables/useReport'

interface Row {
  billType: string; billId: number; billNo: string; status: string; statusLabel: string
  counterpart: string; handlerName: string; quantity: number
  inWarehouse: string; outWarehouse: string
  amount: number; accountName: string; paidAmount: number; payStatus: string
  remark: string; printCount: number; createdAt: string
}

const { list, total, loading, page, pageSize, filters, load, search, exportCsv } = useReport<Row>(
  '/api/v1/bill-center',
  { billType: '', keyword: '', counterpart: '', status: '', payStatus: '', onlyDraft: false, startDate: '', endDate: '' },
)
const dateRange = ref<[string, string] | null>(null)

function onDateChange() {
  filters.value.startDate = dateRange.value?.[0] || ''
  filters.value.endDate = dateRange.value?.[1] || ''
}

function onTypeSelect(index: string) {
  filters.value.billType = index
  load(1)
}

function doExport() {
  exportCsv('单据中心', [
    { key: 'billType', label: '单据类型' }, { key: 'statusLabel', label: '单据状态' },
    { key: 'counterpart', label: '往来单位' }, { key: 'handlerName', label: '经手人' },
    { key: 'quantity', label: '数量' }, { key: 'inWarehouse', label: '入库仓库' },
    { key: 'outWarehouse', label: '出库仓库' }, { key: 'amount', label: '单据金额' },
    { key: 'paidAmount', label: '收付款金额' }, { key: 'payStatus', label: '支付状态' },
    { key: 'remark', label: '单据备注' }, { key: 'billNo', label: '单号' }, { key: 'createdAt', label: '录单时间' },
  ])
}

onMounted(() => load(1))
</script>

<style scoped>
.page { display: flex; flex-direction: column; gap: 12px; height: 100%; }
.page-header { display: flex; justify-content: space-between; align-items: center; }
.page-title { font-size: 18px; font-weight: 600; margin: 0; }
.body { display: flex; gap: 12px; flex: 1; min-height: 0; }
.sidebar { width: 200px; background: var(--el-bg-color); border-radius: 8px; overflow: auto; flex-shrink: 0; }
.sidebar-title { padding: 12px 16px; font-weight: 600; border-bottom: 1px solid var(--el-border-color-lighter); }
.sidebar :deep(.el-menu) { border-right: none; }
.main { flex: 1; min-width: 0; }
.pagination { margin-top: 16px; justify-content: flex-end; }
</style>
