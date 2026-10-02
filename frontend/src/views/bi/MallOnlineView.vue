<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">客户商城上线</h2>
      <el-button @click="doExport">导出</el-button>
    </div>
    <el-card>
      <el-tabs v-model="tab" @tab-change="load(1)">
        <el-tab-pane label="商城上线汇总" name="summary" />
        <el-tab-pane label="上线客户明细" name="detail" />
      </el-tabs>
      <el-form inline @submit.prevent>
        <el-form-item label="日期">
          <el-date-picker v-model="range" type="daterange" value-format="YYYY-MM-DD" start-placeholder="开始日期" end-placeholder="结束日期" style="width: 260px" />
        </el-form-item>
        <el-form-item label="统计客户类型">
          <el-select v-model="filters.customerType" style="width: 160px">
            <el-option label="所有注册客户" value="all" />
            <el-option label="本期新注册客户" value="new" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="tab === 'summary'" label="统计维度">
          <el-select v-model="filters.dimension" style="width: 150px">
            <el-option label="按职员权限" value="employee" />
            <el-option label="按业务经理" value="manager" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="tab === 'detail'" label="客户">
          <el-input v-model="filters.keyword" placeholder="客户名称/编号" clearable style="width: 160px" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="search">查询</el-button>
        </el-form-item>
      </el-form>

      <!-- 汇总 -->
      <el-table v-if="tab === 'summary'" :data="list" v-loading="loading" border stripe>
        <el-table-column prop="dimName" label="职员" min-width="110" fixed />
        <el-table-column prop="customers" label="客户数" width="80" align="right" />
        <el-table-column prop="allMall" label="全商城下单" width="100" align="right" />
        <el-table-column prop="partMall" label="部分商城下单" width="110" align="right" />
        <el-table-column prop="allOff" label="全线下下单" width="100" align="right" />
        <el-table-column prop="notAllMall" label="非全商城下单" width="110" align="right" />
        <el-table-column label="商城单数占比" width="110" align="right">
          <template #default="{ row }">{{ row.mallRatio?.toFixed(1) }}%</template>
        </el-table-column>
        <el-table-column prop="mallBills" label="商城单数" width="90" align="right" />
        <el-table-column prop="totalBills" label="总单数" width="80" align="right" />
        <el-table-column prop="newAllMall" label="新增全商城下单" width="130" align="right" />
        <el-table-column prop="lostAllMall" label="掉线全商城下单" width="130" align="right" />
        <el-table-column prop="netAllMall" label="净增全商城下单" width="130" align="right" />
        <el-table-column prop="newReg" label="新注册" width="80" align="right" />
        <template #empty>暂无数据</template>
      </el-table>

      <!-- 明细 -->
      <el-table v-else :data="list" v-loading="loading" border stripe>
        <el-table-column prop="code" label="客户编号" width="110" />
        <el-table-column prop="customer" label="客户名称" min-width="160" show-overflow-tooltip />
        <el-table-column prop="dimName" label="职员" width="110" />
        <el-table-column prop="mallCnt" label="商城单数" width="90" align="right" />
        <el-table-column prop="offCnt" label="线下单数" width="90" align="right" />
        <el-table-column prop="class" label="下单分类" width="120" align="center">
          <template #default="{ row }">
            <el-tag :type="row.class === '全商城下单' ? 'success' : row.class === '部分商城下单' ? 'warning' : 'info'">{{ row.class }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="注册时间" width="110">
          <template #default="{ row }">{{ formatDate(row.createdAt) }}</template>
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
const summary = useReport<Row>('/api/v1/bi/mall-online/summary', {
  startDate: '', endDate: '', customerType: 'all', dimension: 'employee', keyword: '',
})
const detail = useReport<Row>('/api/v1/bi/mall-online/detail', {
  startDate: '', endDate: '', customerType: 'all', dimension: 'employee', keyword: '',
})

const active = computed(() => (tab.value === 'summary' ? summary : detail))
const list = computed(() => active.value.list.value)
const total = computed(() => active.value.total.value)
const loading = computed(() => active.value.loading.value)
const page = computed({
  get: () => active.value.page.value,
  set: (v) => { active.value.page.value = v },
})
const pageSize = computed({
  get: () => active.value.pageSize.value,
  set: (v) => { active.value.pageSize.value = v },
})
const filters = computed(() => active.value.filters.value)

const range = computed({
  get: () => (filters.value.startDate && filters.value.endDate ? [filters.value.startDate, filters.value.endDate] : null),
  set: (v) => {
    summary.filters.value.startDate = v?.[0] ?? ''
    summary.filters.value.endDate = v?.[1] ?? ''
    detail.filters.value.startDate = v?.[0] ?? ''
    detail.filters.value.endDate = v?.[1] ?? ''
  },
})

watch(() => filters.value.customerType, (v) => {
  summary.filters.value.customerType = v
  detail.filters.value.customerType = v
})

function load(p = 1) { active.value.load(p) }
function search() { load(1) }

function doExport() {
  const cols = tab.value === 'summary'
    ? [
        { key: 'dimName', label: '职员' }, { key: 'customers', label: '客户数' },
        { key: 'allMall', label: '全商城下单' }, { key: 'partMall', label: '部分商城下单' },
        { key: 'allOff', label: '全线下下单' }, { key: 'notAllMall', label: '非全商城下单' },
        { key: 'mallRatio', label: '商城单数占比%' }, { key: 'mallBills', label: '商城单数' },
        { key: 'totalBills', label: '总单数' }, { key: 'newAllMall', label: '新增全商城下单' },
        { key: 'lostAllMall', label: '掉线全商城下单' }, { key: 'netAllMall', label: '净增全商城下单' },
        { key: 'newReg', label: '新注册' },
      ]
    : [
        { key: 'code', label: '客户编号' }, { key: 'customer', label: '客户名称' },
        { key: 'dimName', label: '职员' }, { key: 'mallCnt', label: '商城单数' },
        { key: 'offCnt', label: '线下单数' }, { key: 'class', label: '下单分类' },
      ]
  active.value.exportCsv('客户商城上线', cols)
}

function formatDate(t: string) {
  return t ? new Date(t).toLocaleDateString('zh-CN') : '-'
}

onMounted(() => load(1))
</script>

<style scoped>
.page { display: flex; flex-direction: column; gap: 12px; }
.page-header { display: flex; justify-content: space-between; align-items: center; }
.page-title { font-size: 18px; font-weight: 600; margin: 0; }
</style>
