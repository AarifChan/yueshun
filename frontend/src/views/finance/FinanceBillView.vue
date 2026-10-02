<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">{{ title }}</h2>
      <el-button type="primary" @click="openDialog()">新增{{ title }}</el-button>
    </div>
    <el-card>
      <el-form inline>
        <el-form-item label="时间">
          <el-date-picker v-model="dateRange" type="daterange" value-format="YYYY-MM-DD" start-placeholder="开始时间" end-placeholder="结束时间" @change="onDateChange" />
        </el-form-item>
        <el-form-item label="搜索"><el-input v-model="filters.keyword" placeholder="单号/往来单位" clearable /></el-form-item>
        <el-form-item label="单据状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 120px">
            <el-option label="待审核" value="pending" />
            <el-option label="已完成" value="completed" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="search">查询</el-button>
          <el-button @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>
      <el-table :data="list" v-loading="loading" border stripe>
        <el-table-column prop="billNo" label="单据编号" width="150" fixed="left" />
        <el-table-column prop="billDate" label="单据日期" width="110" />
        <el-table-column prop="itemName" :label="itemLabel" width="120" />
        <el-table-column label="金额" width="120" align="right">
          <template #default="{ row }">{{ row.amount?.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column prop="accountName" label="收付款账户" width="120" />
        <el-table-column prop="counterpart" label="往来单位" width="130" />
        <el-table-column prop="handlerName" label="经手人" width="90" />
        <el-table-column prop="statusLabel" label="单据状态" width="90" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 'completed' ? 'success' : 'warning'">{{ row.statusLabel }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="remark" label="备注" min-width="120" show-overflow-tooltip />
        <el-table-column label="操作" width="170" fixed="right">
          <template #default="{ row }">
            <template v-if="row.status === 'pending'">
              <el-button link type="success" @click="complete(row)">完成</el-button>
              <el-button link type="primary" @click="openDialog(row)">编辑</el-button>
              <el-button link type="danger" @click="remove(row)">删除</el-button>
            </template>
          </template>
        </el-table-column>
        <template #empty>暂无数据</template>
      </el-table>
      <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" layout="total, prev, pager, next" @current-change="load" class="pagination" />
    </el-card>

    <el-dialog v-model="dialogVisible" :title="(form.id ? '编辑' : '新增') + title" width="520px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="单据日期" required>
          <el-date-picker v-model="form.billDate" type="date" value-format="YYYY-MM-DD" style="width: 100%" />
        </el-form-item>
        <el-form-item :label="itemLabel" required>
          <el-select v-model="form.itemId" filterable style="width: 100%">
            <el-option v-for="it in items" :key="it.id" :label="it.name" :value="it.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="金额" required>
          <el-input-number v-model="form.amount" :min="0.01" :precision="2" style="width: 100%" />
        </el-form-item>
        <el-form-item label="收付款账户" required>
          <el-select v-model="form.accountId" filterable style="width: 100%">
            <el-option v-for="a in accounts" :key="a.id" :label="a.name" :value="a.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="往来单位"><el-input v-model="form.counterpart" /></el-form-item>
        <el-form-item label="经手人">
          <el-select v-model="form.handlerId" filterable clearable style="width: 100%">
            <el-option v-for="e in employees" :key="e.id" :label="e.name" :value="e.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="备注"><el-input v-model="form.remark" type="textarea" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="save">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '@/api/client'
import { useReport } from '@/composables/useReport'

const props = defineProps<{
  title: string
  apiUrl: string
  itemType: 'income' | 'expense'
  itemLabel: string
}>()

interface Row {
  id: number; billNo: string; billDate: string; itemId: number; itemName: string
  amount: number; accountId: number; accountName: string; counterpart: string
  handlerId: number | null; handlerName: string; deptId: number
  status: string; statusLabel: string; remark: string
}
interface Opt { id: number; name: string }

const { list, total, loading, page, pageSize, filters, load, search, reset } = useReport<Row>(
  props.apiUrl, { startDate: '', endDate: '', keyword: '', status: '' })
const dateRange = ref<[string, string] | null>(null)

const dialogVisible = ref(false)
const saving = ref(false)
const form = ref({ id: 0, billDate: '', itemId: null as number | null, amount: 0, accountId: null as number | null, counterpart: '', handlerId: null as number | null, remark: '' })
const accounts = ref<Opt[]>([])
const items = ref<Opt[]>([])
const employees = ref<Opt[]>([])

function onDateChange() {
  filters.value.startDate = dateRange.value?.[0] || ''
  filters.value.endDate = dateRange.value?.[1] || ''
}
function onReset() { dateRange.value = null; reset() }

async function loadOptions() {
  const [a, it, e] = await Promise.all([
    api.get('/api/v1/accounts', { params: { page: 1, pageSize: 100 } }),
    api.get('/api/v1/income-expense-items', { params: { page: 1, pageSize: 200, type: props.itemType } }),
    api.get('/api/v1/employees', { params: { page: 1, pageSize: 200 } }),
  ])
  if (a.data.code === 0 || a.data.code === 200) accounts.value = a.data.data?.list ?? []
  if (it.data.code === 0 || it.data.code === 200) items.value = it.data.data?.list ?? []
  if (e.data.code === 0 || e.data.code === 200) employees.value = e.data.data?.list ?? []
}

function openDialog(row?: Row) {
  form.value = row
    ? { id: row.id, billDate: row.billDate, itemId: row.itemId, amount: row.amount, accountId: row.accountId, counterpart: row.counterpart, handlerId: row.handlerId, remark: row.remark }
    : { id: 0, billDate: new Date().toISOString().slice(0, 10), itemId: null, amount: 0, accountId: null, counterpart: '', handlerId: null, remark: '' }
  dialogVisible.value = true
}

async function save() {
  if (!form.value.billDate || !form.value.itemId || !form.value.accountId || !form.value.amount) {
    ElMessage.warning('请填写完整信息'); return
  }
  saving.value = true
  try {
    const res = form.value.id
      ? await api.put(`${props.apiUrl}/${form.value.id}`, form.value)
      : await api.post(props.apiUrl, form.value)
    if (res.data.code === 0 || res.data.code === 200) {
      ElMessage.success('保存成功')
      dialogVisible.value = false
      load()
    } else {
      ElMessage.error(res.data.message || '保存失败')
    }
  } finally {
    saving.value = false
  }
}

async function complete(row: Row) {
  await ElMessageBox.confirm(`确定完成单据「${row.billNo}」？完成后将${props.itemType === 'expense' ? '扣减' : '增加'}账户余额。`, '提示', { type: 'warning' })
  const res = await api.put(`${props.apiUrl}/${row.id}/complete`)
  if (res.data.code === 0 || res.data.code === 200) {
    ElMessage.success('已完成')
    load()
  } else {
    ElMessage.error(res.data.message || '操作失败')
  }
}

async function remove(row: Row) {
  await ElMessageBox.confirm(`确定删除单据「${row.billNo}」？`, '提示', { type: 'warning' })
  const res = await api.delete(`${props.apiUrl}/${row.id}`)
  if (res.data.code === 0 || res.data.code === 200) {
    ElMessage.success('已删除')
    load()
  }
}

onMounted(() => { load(1); loadOptions() })
</script>

<style scoped>
.page { display: flex; flex-direction: column; gap: 12px; }
.page-header { display: flex; justify-content: space-between; align-items: center; }
.page-title { font-size: 18px; font-weight: 600; margin: 0; }
.pagination { margin-top: 16px; justify-content: flex-end; }
</style>
