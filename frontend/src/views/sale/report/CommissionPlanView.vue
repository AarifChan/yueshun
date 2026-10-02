<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">提成方案列表</h2>
      <el-button type="primary" @click="openDialog()">新增提成方案</el-button>
    </div>
    <el-card>
      <el-form inline>
        <el-form-item label="方案名称"><el-input v-model="filters.keyword" placeholder="提成方案名称" clearable /></el-form-item>
        <el-form-item label="方案状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 120px">
            <el-option label="启用" value="1" />
            <el-option label="停用" value="0" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="search">搜索</el-button>
        </el-form-item>
      </el-form>
      <el-table :data="list" v-loading="loading" border stripe>
        <el-table-column type="index" label="序" width="60" />
        <el-table-column prop="name" label="方案名称" min-width="140" />
        <el-table-column label="生效日期" width="110">
          <template #default="{ row }">{{ row.startDate?.slice(0, 10) }}</template>
        </el-table-column>
        <el-table-column label="失效日期" width="110">
          <template #default="{ row }">{{ row.endDate ? row.endDate.slice(0, 10) : '长期' }}</template>
        </el-table-column>
        <el-table-column label="计算方式" width="140">
          <template #default="{ row }">{{ calcTypeLabel(row.calcType) }}</template>
        </el-table-column>
        <el-table-column label="比率/定额" width="100" align="right">
          <template #default="{ row }">{{ row.rate }}</template>
        </el-table-column>
        <el-table-column label="生效职员" min-width="140">
          <template #default="{ row }">{{ employeeNames(row.employeeIds) }}</template>
        </el-table-column>
        <el-table-column label="方案状态" width="90" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'">{{ row.status === 1 ? '启用' : '停用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="160">
          <template #default="{ row }">{{ row.createdAt?.replace('T', ' ').slice(0, 19) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="130" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openDialog(row)">编辑</el-button>
            <el-button link type="danger" @click="remove(row)">删除</el-button>
          </template>
        </el-table-column>
        <template #empty>暂无数据</template>
      </el-table>
      <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" layout="total, prev, pager, next" @current-change="load" class="pagination" />
    </el-card>

    <el-dialog v-model="dialogVisible" :title="form.id ? '编辑提成方案' : '新增提成方案'" width="560px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="方案名称" required><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="生效日期" required>
          <el-date-picker v-model="form.startDate" type="date" value-format="YYYY-MM-DD" />
        </el-form-item>
        <el-form-item label="失效日期">
          <el-date-picker v-model="form.endDate" type="date" value-format="YYYY-MM-DD" placeholder="留空表示长期有效" />
        </el-form-item>
        <el-form-item label="计算方式" required>
          <el-select v-model="form.calcType" style="width: 100%">
            <el-option label="按销售金额比例(%)" value="amount_percent" />
            <el-option label="按销售毛利比例(%)" value="profit_percent" />
            <el-option label="按销售数量定额(元/单位)" value="qty_fixed" />
          </el-select>
        </el-form-item>
        <el-form-item label="比率/定额" required>
          <el-input-number v-model="form.rate" :min="0" :precision="2" style="width: 100%" />
        </el-form-item>
        <el-form-item label="生效职员">
          <el-select v-model="formEmployeeIds" multiple placeholder="留空表示全部职员" style="width: 100%">
            <el-option v-for="e in employees" :key="e.id" :label="e.name" :value="e.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="form.status" :active-value="1" :inactive-value="0" active-text="启用" />
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

interface Plan {
  id: number; name: string; startDate: string; endDate: string | null
  calcType: string; rate: number; employeeIds: string; status: number; remark: string
  createdAt: string
}
interface Emp { id: number; name: string }

const { list, total, loading, page, pageSize, filters, load, search } = useReport<Plan>(
  '/api/v1/commission-plans', { keyword: '', status: '' })

const dialogVisible = ref(false)
const saving = ref(false)
const form = ref<Plan>({ id: 0, name: '', startDate: '', endDate: null, calcType: 'amount_percent', rate: 0, employeeIds: '', status: 1, remark: '', createdAt: '' })
const formEmployeeIds = ref<number[]>([])
const employees = ref<Emp[]>([])

function calcTypeLabel(t: string) {
  return t === 'profit_percent' ? '按销售毛利比例' : t === 'qty_fixed' ? '按销售数量定额' : '按销售金额比例'
}
function employeeNames(ids: string) {
  if (!ids) return '全部职员'
  return ids.split(',').map((s) => employees.value.find((e) => e.id === Number(s))?.name || s).join('、')
}

async function loadEmployees() {
  const res = await api.get('/api/v1/employees', { params: { page: 1, pageSize: 200 } })
  if (res.data.code === 0 || res.data.code === 200) {
    employees.value = res.data.data?.list ?? []
  }
}

function openDialog(row?: Plan) {
  if (row) {
    form.value = { ...row }
    formEmployeeIds.value = row.employeeIds ? row.employeeIds.split(',').map(Number).filter(Boolean) : []
  } else {
    form.value = { id: 0, name: '', startDate: '', endDate: null, calcType: 'amount_percent', rate: 0, employeeIds: '', status: 1, remark: '', createdAt: '' }
    formEmployeeIds.value = []
  }
  dialogVisible.value = true
}

async function save() {
  if (!form.value.name) { ElMessage.warning('请填写方案名称'); return }
  saving.value = true
  try {
    const payload = { ...form.value, employeeIds: formEmployeeIds.value.join(','), endDate: form.value.endDate || null }
    const res = form.value.id
      ? await api.put(`/api/v1/commission-plans/${form.value.id}`, payload)
      : await api.post('/api/v1/commission-plans', payload)
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

async function remove(row: Plan) {
  await ElMessageBox.confirm(`确定删除方案「${row.name}」？`, '提示', { type: 'warning' })
  const res = await api.delete(`/api/v1/commission-plans/${row.id}`)
  if (res.data.code === 0 || res.data.code === 200) {
    ElMessage.success('已删除')
    load()
  }
}

onMounted(() => { load(1); loadEmployees() })
</script>

<style scoped>
.page { display: flex; flex-direction: column; gap: 12px; }
.page-header { display: flex; justify-content: space-between; align-items: center; }
.page-title { font-size: 18px; font-weight: 600; margin: 0; }
.pagination { margin-top: 16px; justify-content: flex-end; }
</style>
