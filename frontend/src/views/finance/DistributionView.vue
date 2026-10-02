<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">{{ title }}</h2>
      <el-button @click="doExport">导出</el-button>
    </div>
    <el-card>
      <el-form inline>
        <el-form-item label="拆分维度">
          <el-select v-model="filters.dimension" style="width: 130px" @change="search">
            <el-option label="经手人" value="handler" />
            <el-option label="往来单位" value="counterpart" />
            <el-option label="部门" value="dept" />
          </el-select>
        </el-form-item>
        <el-form-item label="时间">
          <el-date-picker v-model="dateRange" type="daterange" value-format="YYYY-MM-DD" start-placeholder="开始时间" end-placeholder="结束时间" @change="onDateChange" />
        </el-form-item>
        <el-form-item :label="itemLabel"><el-input v-model="filters.itemKeyword" :placeholder="itemLabel + '名称'" clearable /></el-form-item>
        <el-form-item>
          <el-button type="primary" @click="search">查询</el-button>
        </el-form-item>
      </el-form>
      <el-table :data="list" v-loading="loading" border stripe show-summary>
        <el-table-column prop="name" :label="dimensionLabel" min-width="150" fixed="left" />
        <el-table-column :label="'本期' + itemLabel" width="140" align="right"><template #default="{ row }">{{ row.period?.toFixed(2) }}</template></el-table-column>
        <el-table-column :label="'累计' + itemLabel" width="140" align="right"><template #default="{ row }">{{ row.total?.toFixed(2) }}</template></el-table-column>
        <el-table-column :label="'本期' + itemLabel + '占比(%)'" width="170" align="right"><template #default="{ row }">{{ row.periodPct?.toFixed(1) }}</template></el-table-column>
        <el-table-column :label="'累计' + itemLabel + '占比(%)'" width="170" align="right"><template #default="{ row }">{{ row.totalPct?.toFixed(1) }}</template></el-table-column>
        <template #empty>暂无数据</template>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useReport } from '@/composables/useReport'

const props = defineProps<{
  title: string
  apiUrl: string
  itemLabel: string
}>()

interface Row { key: string; name: string; period: number; total: number; periodPct: number; totalPct: number }

const { list, loading, filters, load, search, exportCsv } = useReport<Row>(
  props.apiUrl, { startDate: '', endDate: '', dimension: 'handler', itemKeyword: '' })
const dateRange = ref<[string, string] | null>(null)

const dimensionLabel = computed(() =>
  filters.value.dimension === 'dept' ? '部门' : filters.value.dimension === 'counterpart' ? '往来单位' : '职员')

function onDateChange() {
  filters.value.startDate = dateRange.value?.[0] || ''
  filters.value.endDate = dateRange.value?.[1] || ''
}
function doExport() {
  exportCsv(props.title, [
    { key: 'name', label: dimensionLabel.value },
    { key: 'period', label: '本期' + props.itemLabel }, { key: 'total', label: '累计' + props.itemLabel },
    { key: 'periodPct', label: '本期占比(%)' }, { key: 'totalPct', label: '累计占比(%)' },
  ])
}
onMounted(() => load(1))
</script>

<style scoped>
.page { display: flex; flex-direction: column; gap: 12px; }
.page-header { display: flex; justify-content: space-between; align-items: center; }
.page-title { font-size: 18px; font-weight: 600; margin: 0; }
</style>
