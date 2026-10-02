<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">新客跟踪</h2>
      <el-radio-group v-model="filters.mode" @change="search">
        <el-radio-button value="total">看总数</el-radio-button>
        <el-radio-button value="new">看新增</el-radio-button>
      </el-radio-group>
    </div>
    <el-card>
      <el-tabs v-model="tab" @tab-change="load(1)">
        <el-tab-pane label="新客汇总" name="summary" />
        <el-tab-pane label="新客明细" name="detail" />
      </el-tabs>
      <el-form inline @submit.prevent>
        <el-form-item label="日期">
          <el-date-picker v-model="range" type="daterange" value-format="YYYY-MM-DD" start-placeholder="开始日期" end-placeholder="结束日期" style="width: 260px" />
        </el-form-item>
        <el-form-item v-if="tab === 'summary'" label="统计维度">
          <el-select v-model="filters.dimension" style="width: 150px">
            <el-option label="按职员权限" value="employee" />
            <el-option label="按业务经理" value="manager" />
            <el-option label="按部门" value="dept" />
            <el-option label="按销售区域" value="region" />
            <el-option label="按客户分类" value="category" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="tab === 'detail'" label="客户">
          <el-input v-model="filters.keyword" placeholder="客户名称/编号" clearable style="width: 160px" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="search">查询</el-button>
          <el-button @click="doExport">导出</el-button>
        </el-form-item>
      </el-form>

      <el-table v-if="tab === 'summary'" :data="list" v-loading="loading" border stripe>
        <el-table-column type="index" label="序" width="60" />
        <el-table-column prop="dimName" label="维度" min-width="120" />
        <el-table-column prop="customers" label="新客户数" width="100" align="right" />
        <el-table-column prop="notRepeat" label="未复购新客" width="110" align="right" />
        <el-table-column prop="repeated" label="已复购新客" width="110" align="right" />
        <el-table-column prop="retained" label="已留存新客" width="110" align="right" />
        <el-table-column prop="tradeAmount" label="交易金额" width="120" align="right">
          <template #default="{ row }">{{ row.tradeAmount?.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="留存率" width="90" align="right">
          <template #default="{ row }">{{ row.retainRate?.toFixed(1) }}%</template>
        </el-table-column>
        <template #empty>暂无数据</template>
      </el-table>

      <el-table v-else :data="list" v-loading="loading" border stripe>
        <el-table-column type="index" label="序" width="60" />
        <el-table-column prop="code" label="客户编号" width="110" />
        <el-table-column prop="customer" label="新客户" min-width="160" show-overflow-tooltip />
        <el-table-column label="注册日期" width="110">
          <template #default="{ row }">{{ row.created?.slice(0, 10) }}</template>
        </el-table-column>
        <el-table-column label="首单日期" width="110">
          <template #default="{ row }">{{ row.firstBill || '-' }}</template>
        </el-table-column>
        <el-table-column label="最近下单" width="110">
          <template #default="{ row }">{{ row.lastBill || '-' }}</template>
        </el-table-column>
        <el-table-column prop="billDays" label="交易天数" width="90" align="right" />
        <el-table-column prop="tradeAmount" label="交易金额" width="120" align="right">
          <template #default="{ row }">{{ row.tradeAmount?.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === '已留存' ? 'success' : row.status === '已复购' ? 'warning' : 'info'">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <template #empty>暂无数据</template>
      </el-table>

      <el-pagination
        v-model:current-page="page" v-model:page-size="pageSize"
        :total="total" :page-sizes="[30, 50, 100]"
        layout="total, sizes, prev, pager, next" style="margin-top: 12px; justify-content: flex-end"
        @current-change="load()" @size-change="load(1)" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useReport } from '@/composables/useReport'

type Row = Record<string, any>

const tab = ref('summary')
const defaults = { startDate: '', endDate: '', dimension: 'employee', keyword: '', mode: 'new' }
const summary = useReport<Row>('/api/v1/bi/new-customers/summary', { ...defaults })
const detail = useReport<Row>('/api/v1/bi/new-customers/detail', { ...defaults })

const active = computed(() => (tab.value === 'summary' ? summary : detail))
const list = computed(() => active.value.list.value)
const total = computed(() => active.value.total.value)
const loading = computed(() => active.value.loading.value)
const page = computed({ get: () => active.value.page.value, set: (v) => { active.value.page.value = v } })
const pageSize = computed({ get: () => active.value.pageSize.value, set: (v) => { active.value.pageSize.value = v } })
const filters = computed(() => summary.filters.value)

const range = computed({
  get: () => (summary.filters.value.startDate && summary.filters.value.endDate ? [summary.filters.value.startDate, summary.filters.value.endDate] : null),
  set: (v) => {
    for (const r of [summary, detail]) {
      r.filters.value.startDate = v?.[0] ?? ''
      r.filters.value.endDate = v?.[1] ?? ''
    }
  },
})

watch(() => summary.filters.value.mode, (v) => { detail.filters.value.mode = v })

function load(p = 1) { active.value.load(p) }
function search() { load(1) }

function doExport() {
  const cols = tab.value === 'summary'
    ? [
        { key: 'dimName', label: '维度' }, { key: 'customers', label: '新客户数' },
        { key: 'notRepeat', label: '未复购新客' }, { key: 'repeated', label: '已复购新客' },
        { key: 'retained', label: '已留存新客' }, { key: 'tradeAmount', label: '交易金额' },
        { key: 'retainRate', label: '留存率%' },
      ]
    : [
        { key: 'code', label: '客户编号' }, { key: 'customer', label: '新客户' },
        { key: 'created', label: '注册日期' }, { key: 'firstBill', label: '首单日期' },
        { key: 'lastBill', label: '最近下单' }, { key: 'billDays', label: '交易天数' },
        { key: 'tradeAmount', label: '交易金额' }, { key: 'status', label: '状态' },
      ]
  active.value.exportCsv('新客跟踪', cols)
}

onMounted(() => load(1))
</script>

<style scoped>
.page { display: flex; flex-direction: column; gap: 12px; }
.page-header { display: flex; justify-content: space-between; align-items: center; }
.page-title { font-size: 18px; font-weight: 600; margin: 0; }
</style>
