<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">商品铺市</h2>
      <el-button @click="doExport">导出</el-button>
    </div>
    <el-card>
      <el-tabs v-model="tab" @tab-change="load(1)">
        <el-tab-pane label="商品铺市汇总" name="summary" />
        <el-tab-pane label="商品铺市明细" name="detail" />
      </el-tabs>
      <el-form inline @submit.prevent>
        <el-form-item label="日期">
          <el-date-picker v-model="range" type="daterange" value-format="YYYY-MM-DD" start-placeholder="开始日期" end-placeholder="结束日期" style="width: 260px" />
        </el-form-item>
        <el-form-item label="商品">
          <el-input v-model="productKeyword" placeholder="名称/编号/规格" clearable style="width: 170px" />
        </el-form-item>
        <el-form-item v-if="tab === 'detail'" label="客户">
          <el-input v-model="keyword" placeholder="客户名称/编号" clearable style="width: 150px" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="search">查询</el-button>
        </el-form-item>
      </el-form>

      <el-table v-if="tab === 'summary'" :data="list" v-loading="loading" border stripe>
        <el-table-column type="index" label="序" width="60" />
        <el-table-column prop="productCode" label="商品编号" width="110" />
        <el-table-column prop="product" label="商品名称" min-width="160" show-overflow-tooltip />
        <el-table-column prop="spec" label="规格" width="110" />
        <el-table-column prop="unit" label="单位" width="70" />
        <el-table-column prop="total" label="累计铺市客户" width="120" align="right" sortable />
        <el-table-column prop="initial" label="期初铺市" width="100" align="right" />
        <el-table-column prop="newIn" label="本期新拿" width="100" align="right" />
        <el-table-column prop="period" label="本期铺市" width="100" align="right" />
        <template #empty>暂无数据</template>
      </el-table>

      <el-table v-else :data="list" v-loading="loading" border stripe>
        <el-table-column type="index" label="序" width="60" />
        <el-table-column prop="productCode" label="商品编号" width="110" />
        <el-table-column prop="product" label="商品名称" min-width="150" show-overflow-tooltip />
        <el-table-column prop="spec" label="规格" width="100" />
        <el-table-column prop="code" label="客户编号" width="110" />
        <el-table-column prop="customer" label="铺市客户" min-width="150" show-overflow-tooltip />
        <el-table-column prop="firstDate" label="首次拿货日期" width="120" />
        <el-table-column prop="qty" label="本期数量" width="100" align="right" />
        <el-table-column prop="amount" label="本期金额" width="110" align="right">
          <template #default="{ row }">{{ row.amount?.toFixed(2) }}</template>
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
const defaults = { startDate: '', endDate: '', keyword: '', productKeyword: '' }
const summary = useReport<Row>('/api/v1/bi/product-distribution/summary', { ...defaults })
const detail = useReport<Row>('/api/v1/bi/product-distribution/detail', { ...defaults })

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

const keyword = computed({
  get: () => detail.filters.value.keyword,
  set: (v) => { detail.filters.value.keyword = v },
})
const productKeyword = computed({
  get: () => summary.filters.value.keyword,
  set: (v) => {
    summary.filters.value.keyword = v
    detail.filters.value.productKeyword = v
  },
})

function load(p = 1) { active.value.load(p) }
function search() { load(1) }

function doExport() {
  const cols = tab.value === 'summary'
    ? [
        { key: 'productCode', label: '商品编号' }, { key: 'product', label: '商品名称' },
        { key: 'spec', label: '规格' }, { key: 'total', label: '累计铺市客户' },
        { key: 'initial', label: '期初铺市' }, { key: 'newIn', label: '本期新拿' },
        { key: 'period', label: '本期铺市' },
      ]
    : [
        { key: 'productCode', label: '商品编号' }, { key: 'product', label: '商品名称' },
        { key: 'code', label: '客户编号' }, { key: 'customer', label: '铺市客户' },
        { key: 'firstDate', label: '首次拿货日期' }, { key: 'qty', label: '本期数量' },
        { key: 'amount', label: '本期金额' },
      ]
  active.value.exportCsv('商品铺市', cols)
}

onMounted(() => load(1))
</script>

<style scoped>
.page { display: flex; flex-direction: column; gap: 12px; }
.page-header { display: flex; justify-content: space-between; align-items: center; }
.page-title { font-size: 18px; font-weight: 600; margin: 0; }
</style>
