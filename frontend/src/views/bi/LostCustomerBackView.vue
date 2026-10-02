<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">流失拉回</h2>
      <el-button @click="doExport">导出</el-button>
    </div>
    <el-card>
      <el-tabs v-model="tab" @tab-change="load(1)">
        <el-tab-pane label="流失客户汇总" name="summary" />
        <el-tab-pane label="流失客户明细" name="detail" />
      </el-tabs>
      <el-alert type="info" :closable="false" style="margin-bottom: 10px"
        title="统计截止到所选区间开始日期的前一天 23:59，超过「流失天数」未交易即计为流失客户；区间内再次交易计为拉回。" />
      <el-form inline @submit.prevent>
        <el-form-item label="日期">
          <el-date-picker v-model="range" type="daterange" value-format="YYYY-MM-DD" start-placeholder="开始日期" end-placeholder="结束日期" style="width: 260px" />
        </el-form-item>
        <el-form-item label="流失天数">
          <el-input-number v-model="lossDays" :min="7" :max="365" style="width: 120px" />
        </el-form-item>
        <el-form-item v-if="tab === 'summary'" label="统计维度">
          <el-select v-model="dimension" style="width: 150px">
            <el-option label="按职员权限" value="employee" />
            <el-option label="按业务经理" value="manager" />
            <el-option label="按部门" value="dept" />
            <el-option label="按销售区域" value="region" />
            <el-option label="按客户分类" value="category" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="tab === 'detail'" label="客户">
          <el-input v-model="keyword" placeholder="客户名称/编号" clearable style="width: 150px" />
        </el-form-item>
        <el-form-item v-if="tab === 'detail'" label="仅看拉回">
          <el-switch v-model="onlyBack" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="search">查询</el-button>
        </el-form-item>
      </el-form>

      <el-table v-if="tab === 'summary'" :data="list" v-loading="loading" border stripe>
        <el-table-column type="index" label="序" width="60" />
        <el-table-column prop="dimName" label="维度" min-width="130" />
        <el-table-column prop="lost" label="流失客户数" width="110" align="right" />
        <el-table-column prop="back" label="拉回客户数" width="110" align="right" />
        <el-table-column label="拉回率" width="90" align="right">
          <template #default="{ row }">{{ row.backRate?.toFixed(1) }}%</template>
        </el-table-column>
        <el-table-column prop="backAmount" label="拉回交易金额" width="130" align="right">
          <template #default="{ row }">{{ row.backAmount?.toFixed(2) }}</template>
        </el-table-column>
        <template #empty>暂无数据</template>
      </el-table>

      <el-table v-else :data="list" v-loading="loading" border stripe>
        <el-table-column type="index" label="序" width="60" />
        <el-table-column prop="code" label="客户编号" width="110" />
        <el-table-column prop="customer" label="客户" min-width="160" show-overflow-tooltip />
        <el-table-column prop="lastBill" label="最后交易日期" width="120" />
        <el-table-column label="是否拉回" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.isBack ? 'success' : 'info'">{{ row.isBack ? '已拉回' : '未拉回' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="拉回日期" width="110">
          <template #default="{ row }">{{ row.backDate || '-' }}</template>
        </el-table-column>
        <el-table-column label="拉回单数" width="90" align="right">
          <template #default="{ row }">{{ row.backBills ?? '-' }}</template>
        </el-table-column>
        <el-table-column label="拉回金额" width="110" align="right">
          <template #default="{ row }">{{ row.backAmount != null ? row.backAmount.toFixed(2) : '-' }}</template>
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
import { computed, onMounted, ref } from 'vue'
import { useReport } from '@/composables/useReport'

type Row = Record<string, any>

const tab = ref('summary')
const defaults = { startDate: '', endDate: '', dimension: 'employee', keyword: '', lossDays: 90, onlyBack: '' }
const summary = useReport<Row>('/api/v1/bi/lost-customers/summary', { ...defaults })
const detail = useReport<Row>('/api/v1/bi/lost-customers/detail', { ...defaults })

const active = computed(() => (tab.value === 'summary' ? summary : detail))
const list = computed(() => active.value.list.value)
const total = computed(() => active.value.total.value)
const loading = computed(() => active.value.loading.value)
const page = computed({ get: () => active.value.page.value, set: (v) => { active.value.page.value = v } })
const pageSize = computed({ get: () => active.value.pageSize.value, set: (v) => { active.value.pageSize.value = v } })

const range = computed({
  get: () => (summary.filters.value.startDate && summary.filters.value.endDate ? [summary.filters.value.startDate, summary.filters.value.endDate] : null),
  set: (v) => {
    for (const r of [summary, detail]) {
      r.filters.value.startDate = v?.[0] ?? ''
      r.filters.value.endDate = v?.[1] ?? ''
    }
  },
})

const dimension = computed({
  get: () => summary.filters.value.dimension,
  set: (v) => { summary.filters.value.dimension = v },
})
const keyword = computed({
  get: () => detail.filters.value.keyword,
  set: (v) => { detail.filters.value.keyword = v },
})
const lossDays = computed({
  get: () => summary.filters.value.lossDays,
  set: (v) => {
    summary.filters.value.lossDays = v
    detail.filters.value.lossDays = v
  },
})
const onlyBack = computed({
  get: () => detail.filters.value.onlyBack === '1',
  set: (v) => { detail.filters.value.onlyBack = v ? '1' : '' },
})

function load(p = 1) { active.value.load(p) }
function search() { load(1) }

function doExport() {
  const cols = tab.value === 'summary'
    ? [
        { key: 'dimName', label: '维度' }, { key: 'lost', label: '流失客户数' },
        { key: 'back', label: '拉回客户数' }, { key: 'backRate', label: '拉回率%' },
        { key: 'backAmount', label: '拉回交易金额' },
      ]
    : [
        { key: 'code', label: '客户编号' }, { key: 'customer', label: '客户' },
        { key: 'lastBill', label: '最后交易日期' }, { key: 'backDate', label: '拉回日期' },
        { key: 'backBills', label: '拉回单数' }, { key: 'backAmount', label: '拉回金额' },
      ]
  active.value.exportCsv('流失拉回', cols)
}

onMounted(() => load(1))
</script>

<style scoped>
.page { display: flex; flex-direction: column; gap: 12px; }
.page-header { display: flex; justify-content: space-between; align-items: center; }
.page-title { font-size: 18px; font-weight: 600; margin: 0; }
</style>
