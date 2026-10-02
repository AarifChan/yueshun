<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">职员提成统计</h2>
      <el-button @click="doExport">导出</el-button>
    </div>
    <el-card>
      <el-form inline>
        <el-form-item label="提成方案" required>
          <el-select v-model="filters.planId" placeholder="请选择提成方案" style="width: 200px">
            <el-option v-for="p in plans" :key="p.id" :label="p.name" :value="p.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="时间">
          <el-date-picker v-model="dateRange" type="daterange" value-format="YYYY-MM-DD" start-placeholder="开始日期" end-placeholder="结束日期" @change="onDateChange" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="doQuery">立即查询</el-button>
        </el-form-item>
      </el-form>
      <el-alert v-if="!queried" type="info" :closable="false" title="若需计算请点击查询，查询需要时间" class="tip" />
      <el-table v-else :data="list" v-loading="loading" border stripe show-summary>
        <el-table-column prop="employeeId" label="职员编号" width="90" />
        <el-table-column prop="name" label="职员姓名" width="110" />
        <el-table-column prop="dept" label="部门" width="110" />
        <el-table-column prop="planName" label="生效方案" min-width="130" />
        <el-table-column label="提成合计" width="120" align="right">
          <template #default="{ row }">{{ row.commission?.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="销售金额" width="120" align="right">
          <template #default="{ row }">{{ row.saleAmount?.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="销售数量" width="100" align="right">
          <template #default="{ row }">{{ row.saleQty }}</template>
        </el-table-column>
        <el-table-column label="回款金额" width="120" align="right">
          <template #default="{ row }">{{ row.receipt?.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="销售毛利" width="120" align="right">
          <template #default="{ row }">{{ row.profit?.toFixed(2) }}</template>
        </el-table-column>
        <template #empty>暂无数据</template>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import api from '@/api/client'
import { useReport } from '@/composables/useReport'

interface Row {
  employeeId: number; name: string; dept: string; planName: string
  commission: number; saleAmount: number; saleQty: number; receipt: number; profit: number
}
interface Plan { id: number; name: string }

const { list, loading, filters, load, exportCsv } = useReport<Row>(
  '/api/v1/sale-reports/commission-stats',
  { planId: null as number | null, startDate: '', endDate: '' })
const dateRange = ref<[string, string] | null>(null)
const plans = ref<Plan[]>([])
const queried = ref(false)

function onDateChange() {
  filters.value.startDate = dateRange.value?.[0] || ''
  filters.value.endDate = dateRange.value?.[1] || ''
}

function doQuery() {
  if (!filters.value.planId) { ElMessage.warning('请选择提成方案'); return }
  queried.value = true
  load(1)
}

function doExport() {
  if (!queried.value) { ElMessage.warning('请先查询'); return }
  exportCsv('职员提成统计', [
    { key: 'employeeId', label: '职员编号' }, { key: 'name', label: '职员姓名' },
    { key: 'dept', label: '部门' }, { key: 'planName', label: '生效方案' },
    { key: 'commission', label: '提成合计' }, { key: 'saleAmount', label: '销售金额' },
    { key: 'saleQty', label: '销售数量' }, { key: 'receipt', label: '回款金额' }, { key: 'profit', label: '销售毛利' },
  ])
}

async function loadPlans() {
  const res = await api.get('/api/v1/commission-plans/all')
  if (res.data.code === 0 || res.data.code === 200) {
    plans.value = res.data.data?.list ?? []
  }
}
onMounted(loadPlans)
</script>

<style scoped>
.page { display: flex; flex-direction: column; gap: 12px; }
.page-header { display: flex; justify-content: space-between; align-items: center; }
.page-title { font-size: 18px; font-weight: 600; margin: 0; }
.tip { margin-bottom: 12px; }
</style>
