<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">业务驾驶舱设置</h2>
      <el-button type="primary" :loading="saving" @click="saveAll">保存全部</el-button>
    </div>
    <el-card>
      <el-form inline>
        <el-form-item label="年份">
          <el-date-picker v-model="yearStr" type="year" value-format="YYYY" @change="load" />
        </el-form-item>
        <el-form-item label="职员"><el-input v-model="employeeKeyword" placeholder="职员名称" clearable @keyup.enter="load" /></el-form-item>
        <el-form-item><el-button type="primary" @click="load">搜索</el-button></el-form-item>
      </el-form>
      <el-table :data="list" v-loading="loading" border stripe max-height="600">
        <el-table-column type="index" label="序" width="50" fixed="left" />
        <el-table-column prop="dept" label="部门名称" width="110" fixed="left" />
        <el-table-column prop="name" label="职员名称" width="100" fixed="left" />
        <el-table-column v-for="m in 12" :key="m" :label="m + '月'" width="100" align="right">
          <template #default="{ row }">
            <el-input-number v-model="row.months[m - 1]" :min="0" :controls="false" size="small" style="width: 88px" @change="recalc(row)" />
          </template>
        </el-table-column>
        <el-table-column label="年度总计" width="120" align="right" fixed="right">
          <template #default="{ row }"><b>{{ row.total.toFixed(2) }}</b></template>
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

interface Row {
  employeeId: number; name: string; dept: string; planId: number
  months: number[]; total: number
}

const list = ref<Row[]>([])
const loading = ref(false)
const saving = ref(false)
const yearStr = ref(String(new Date().getFullYear()))
const employeeKeyword = ref('')

function recalc(row: Row) {
  row.total = row.months.reduce((s, v) => s + (v || 0), 0)
}

async function load() {
  loading.value = true
  try {
    const res = await api.get('/api/v1/sales-plans', {
      params: { year: yearStr.value, employeeKeyword: employeeKeyword.value },
    })
    if (res.data.code === 0 || res.data.code === 200) {
      list.value = (res.data.data?.list ?? []).map((r: Row) => ({ ...r, months: [...(r.months || Array(12).fill(0))] }))
    }
  } finally {
    loading.value = false
  }
}

async function saveAll() {
  saving.value = true
  try {
    let ok = 0
    for (const row of list.value) {
      if (row.total <= 0 && !row.planId) continue
      const res = await api.put('/api/v1/sales-plans', {
        year: Number(yearStr.value), employeeId: row.employeeId, months: row.months,
      })
      if (res.data.code === 0 || res.data.code === 200) ok++
    }
    ElMessage.success(`已保存 ${ok} 条计划`)
    load()
  } finally {
    saving.value = false
  }
}

onMounted(load)
</script>

<style scoped>
.page { display: flex; flex-direction: column; gap: 12px; }
.page-header { display: flex; justify-content: space-between; align-items: center; }
.page-title { font-size: 18px; font-weight: 600; margin: 0; }
</style>
