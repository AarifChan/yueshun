<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">汇报</h2>
      <div>
        <el-button @click="doExport">导出</el-button>
        <el-button type="primary" @click="openDialog">点我写日志</el-button>
      </div>
    </div>
    <el-card>
      <el-tabs v-model="tab" @tab-change="onTabChange">
        <el-tab-pane label="我写的" name="mine" />
        <el-tab-pane label="抄给我的" name="cc" />
        <el-tab-pane label="他人汇报" name="others" />
        <el-tab-pane label="汇报统计" name="stats" />
      </el-tabs>

      <template v-if="tab !== 'stats'">
        <el-form inline>
          <el-form-item label="类型">
            <el-select v-model="filters.type" placeholder="全部" clearable style="width: 110px" @change="search">
              <el-option label="日志" value="log" />
              <el-option label="周报" value="week" />
              <el-option label="月报" value="month" />
            </el-select>
          </el-form-item>
          <el-form-item label="日期">
            <el-date-picker v-model="filters.date" type="date" value-format="YYYY-MM-DD" placeholder="选择日期" @change="search" />
          </el-form-item>
          <el-form-item><el-button @click="onReset">清空</el-button></el-form-item>
        </el-form>
        <el-empty v-if="!loading && list.length === 0" description="今天还未写汇报" />
        <div v-for="r in list" :key="r.id" class="report-item">
          <div class="report-head">
            <el-tag size="small" :type="r.type === 'log' ? 'primary' : r.type === 'week' ? 'warning' : 'success'">{{ r.typeLabel }}</el-tag>
            <b>{{ r.authorName }}</b>
            <span class="date">{{ r.reportDate }}</span>
            <span v-if="r.ccNames?.length" class="cc">抄送：{{ r.ccNames.join('、') }}</span>
            <el-button v-if="tab === 'mine'" link type="danger" size="small" @click="remove(r)">删除</el-button>
          </div>
          <div class="report-content">{{ r.content }}</div>
        </div>
        <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" layout="total, prev, pager, next" @current-change="load" class="pagination" />
      </template>

      <template v-else>
        <el-form inline>
          <el-form-item label="月份">
            <el-date-picker v-model="statsMonth" type="month" value-format="YYYY-MM" @change="loadStats" />
          </el-form-item>
        </el-form>
        <el-table :data="statsList" v-loading="statsLoading" border stripe>
          <el-table-column prop="name" label="职员" width="140" />
          <el-table-column prop="logCount" label="日志数" width="110" align="right" />
          <el-table-column prop="weekCount" label="周报数" width="110" align="right" />
          <el-table-column prop="monthCount" label="月报数" width="110" align="right" />
          <el-table-column prop="total" label="合计" width="110" align="right" />
          <template #empty>暂无数据</template>
        </el-table>
      </template>
    </el-card>

    <el-dialog v-model="dialogVisible" title="新增汇报" width="560px">
      <el-form :model="form" label-width="90px">
        <el-form-item label="类型" required>
          <el-radio-group v-model="form.type">
            <el-radio-button value="log">日志</el-radio-button>
            <el-radio-button value="week">周报</el-radio-button>
            <el-radio-button value="month">月报</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="汇报日期" required>
          <el-date-picker v-model="form.reportDate" type="date" value-format="YYYY-MM-DD" style="width: 100%" />
        </el-form-item>
        <el-form-item label="内容" required>
          <el-input v-model="form.content" type="textarea" :rows="6" placeholder="今天做了什么、遇到的问题、明天计划…" />
        </el-form-item>
        <el-form-item label="抄送给">
          <el-select v-model="formCcIds" multiple filterable style="width: 100%">
            <el-option v-for="e in employees" :key="e.id" :label="e.name" :value="e.id" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="save">提交</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '@/api/client'
import { useReport } from '@/composables/useReport'

interface Row {
  id: number; type: string; typeLabel: string; reportDate: string; content: string
  authorId: number; authorName: string; ccNames: string[]; createdAt: string
}
interface StatRow { employeeId: number; name: string; logCount: number; weekCount: number; monthCount: number; total: number }
interface Opt { id: number; name: string }

const { list, total, loading, page, pageSize, filters, load, search, reset, exportCsv } = useReport<Row>(
  '/api/v1/work-reports', { tab: 'mine', type: '', date: '' })
const tab = ref('mine')

const statsList = ref<StatRow[]>([])
const statsLoading = ref(false)
const statsMonth = ref(new Date().toISOString().slice(0, 7))

const dialogVisible = ref(false)
const form = ref({ type: 'log', reportDate: new Date().toISOString().slice(0, 10), content: '' })
const formCcIds = ref<number[]>([])
const employees = ref<Opt[]>([])

function onTabChange() {
  if (tab.value === 'stats') { loadStats(); return }
  filters.value.tab = tab.value
  load(1)
}
function onReset() { reset() }

async function loadStats() {
  statsLoading.value = true
  try {
    const res = await api.get('/api/v1/work-reports/stats', { params: { month: statsMonth.value } })
    if (res.data.code === 0 || res.data.code === 200) statsList.value = res.data.data?.list ?? []
  } finally {
    statsLoading.value = false
  }
}

function openDialog() {
  form.value = { type: 'log', reportDate: new Date().toISOString().slice(0, 10), content: '' }
  formCcIds.value = []
  dialogVisible.value = true
}

async function save() {
  if (!form.value.content) { ElMessage.warning('请填写汇报内容'); return }
  const res = await api.post('/api/v1/work-reports', { ...form.value, ccIds: formCcIds.value.join(',') })
  if (res.data.code === 0 || res.data.code === 200) {
    ElMessage.success('汇报已提交')
    dialogVisible.value = false
    if (tab.value !== 'stats') load(1)
  } else {
    ElMessage.error(res.data.message || '提交失败')
  }
}

async function remove(r: Row) {
  await ElMessageBox.confirm('确定删除该汇报？', '提示', { type: 'warning' })
  const res = await api.delete(`/api/v1/work-reports/${r.id}`)
  if (res.data.code === 0 || res.data.code === 200) {
    ElMessage.success('已删除')
    load()
  }
}

function doExport() {
  exportCsv('汇报', [
    { key: 'typeLabel', label: '类型' }, { key: 'reportDate', label: '汇报日期' },
    { key: 'authorName', label: '作者' }, { key: 'content', label: '内容' },
  ])
}

onMounted(async () => {
  load(1)
  const res = await api.get('/api/v1/employees', { params: { page: 1, pageSize: 200 } })
  if (res.data.code === 0 || res.data.code === 200) employees.value = res.data.data?.list ?? []
})
</script>

<style scoped>
.page { display: flex; flex-direction: column; gap: 12px; }
.page-header { display: flex; justify-content: space-between; align-items: center; }
.page-title { font-size: 18px; font-weight: 600; margin: 0; }
.report-item { border: 1px solid var(--el-border-color-lighter); border-radius: 8px; padding: 12px 16px; margin-bottom: 12px; }
.report-head { display: flex; align-items: center; gap: 10px; margin-bottom: 8px; }
.report-head .date { color: var(--el-text-color-secondary); }
.report-head .cc { color: var(--el-text-color-secondary); font-size: 12px; }
.report-head .el-button { margin-left: auto; }
.report-content { white-space: pre-wrap; color: var(--el-text-color-regular); }
.pagination { margin-top: 16px; justify-content: flex-end; }
</style>
