<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">分销</h2>
      <el-button v-if="tab === 'list'" type="primary" @click="openDialog()">新增分销商</el-button>
    </div>
    <el-card>
      <el-tabs v-model="tab">
        <el-tab-pane label="分销商列表" name="list" />
        <el-tab-pane label="佣金提现" name="withdrawals" />
        <el-tab-pane label="佣金设置" name="commission" />
        <el-tab-pane label="基础设置" name="base" />
      </el-tabs>

      <!-- 分销商列表 -->
      <template v-if="tab === 'list'">
        <el-form inline @submit.prevent>
          <el-form-item label="状态">
            <el-select v-model="filters.status" style="width: 120px" @change="search">
              <el-option label="全部" value="" />
              <el-option label="待审核" value="pending" />
              <el-option label="已通过" value="active" />
              <el-option label="已停用" value="disabled" />
            </el-select>
          </el-form-item>
          <el-form-item label="分销商">
            <el-input v-model="filters.keyword" placeholder="名称/手机号" clearable style="width: 180px" />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="search">搜索</el-button>
            <el-button @click="exportCsv('分销商列表', exportCols)">导出</el-button>
          </el-form-item>
        </el-form>
        <el-table :data="list" v-loading="loading" border stripe>
          <el-table-column prop="name" label="分销商名称" min-width="140" show-overflow-tooltip />
          <el-table-column prop="manager" label="业务经理" width="100" />
          <el-table-column prop="saleAmount" label="分销额" width="100" align="right">
            <template #default="{ row }">{{ row.saleAmount?.toFixed(2) }}</template>
          </el-table-column>
          <el-table-column prop="unpostedAmount" label="未入账金额" width="100" align="right">
            <template #default="{ row }">{{ row.unpostedAmount?.toFixed(2) }}</template>
          </el-table-column>
          <el-table-column prop="postedAmount" label="已入账金额" width="100" align="right">
            <template #default="{ row }">{{ row.postedAmount?.toFixed(2) }}</template>
          </el-table-column>
          <el-table-column prop="withdrawable" label="可提现佣金" width="100" align="right">
            <template #default="{ row }">{{ row.withdrawable?.toFixed(2) }}</template>
          </el-table-column>
          <el-table-column prop="withdrawing" label="提现中佣金" width="100" align="right">
            <template #default="{ row }">{{ row.withdrawing?.toFixed(2) }}</template>
          </el-table-column>
          <el-table-column prop="withdrawn" label="已提现佣金" width="100" align="right">
            <template #default="{ row }">{{ row.withdrawn?.toFixed(2) }}</template>
          </el-table-column>
          <el-table-column prop="commissionBalance" label="佣金余额" width="100" align="right">
            <template #default="{ row }">{{ row.commissionBalance?.toFixed(2) }}</template>
          </el-table-column>
          <el-table-column prop="undistributedDays" label="未分销天数" width="100" align="right" />
          <el-table-column label="状态" width="90" align="center">
            <template #default="{ row }">
              <el-tag :type="row.status === 'active' ? 'success' : row.status === 'pending' ? 'warning' : 'info'">
                {{ row.status === 'active' ? '已通过' : row.status === 'pending' ? '待审核' : '已停用' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="申请时间" width="110">
            <template #default="{ row }">{{ formatDate(row.appliedAt) }}</template>
          </el-table-column>
          <el-table-column label="通过时间" width="110">
            <template #default="{ row }">{{ formatDate(row.approvedAt) }}</template>
          </el-table-column>
          <el-table-column label="操作" width="170" fixed="right">
            <template #default="{ row }">
              <el-button v-if="row.status === 'pending'" link type="success" @click="approve(row)">通过</el-button>
              <el-button link type="primary" @click="openDialog(row)">编辑</el-button>
              <el-button v-if="row.status === 'active'" link type="warning" @click="disable(row)">停用</el-button>
              <el-button link type="danger" @click="remove(row)">删除</el-button>
            </template>
          </el-table-column>
          <template #empty>暂无数据</template>
        </el-table>
        <el-pagination
          v-model:current-page="page" v-model:page-size="pageSize"
          :total="total" :page-sizes="[30, 50, 100]"
          layout="total, sizes, prev, pager, next" style="margin-top: 12px; justify-content: flex-end"
          @current-change="load()" @size-change="load(1)" />
      </template>

      <!-- 佣金提现 -->
      <template v-else-if="tab === 'withdrawals'">
        <el-form inline @submit.prevent>
          <el-form-item label="状态">
            <el-select v-model="wStatus" style="width: 120px" @change="loadWithdrawals(1)">
              <el-option label="全部" value="" />
              <el-option label="提现中" value="pending" />
              <el-option label="已提现" value="paid" />
              <el-option label="已拒绝" value="rejected" />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="openWithdrawal()">新增提现</el-button>
          </el-form-item>
        </el-form>
        <el-table :data="wList" v-loading="wLoading" border stripe>
          <el-table-column prop="distributor" label="分销商" min-width="140" />
          <el-table-column prop="amount" label="提现金额" width="110" align="right">
            <template #default="{ row }">{{ row.amount?.toFixed(2) }}</template>
          </el-table-column>
          <el-table-column label="状态" width="90" align="center">
            <template #default="{ row }">
              <el-tag :type="row.status === 'paid' ? 'success' : row.status === 'pending' ? 'warning' : 'danger'">
                {{ row.status === 'paid' ? '已提现' : row.status === 'pending' ? '提现中' : '已拒绝' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="申请时间" width="160">
            <template #default="{ row }">{{ formatTime(row.appliedAt) }}</template>
          </el-table-column>
          <el-table-column label="打款时间" width="160">
            <template #default="{ row }">{{ formatTime(row.paidAt) }}</template>
          </el-table-column>
          <el-table-column prop="remark" label="备注" min-width="120" />
          <el-table-column label="操作" width="130" fixed="right">
            <template #default="{ row }">
              <template v-if="row.status === 'pending'">
                <el-button link type="success" @click="payWithdrawal(row)">确认打款</el-button>
                <el-button link type="danger" @click="rejectWithdrawal(row)">拒绝</el-button>
              </template>
            </template>
          </el-table-column>
          <template #empty>暂无数据</template>
        </el-table>
        <el-pagination
          v-model:current-page="wPage" v-model:page-size="wPageSize"
          :total="wTotal" :page-sizes="[30, 50, 100]"
          layout="total, sizes, prev, pager, next" style="margin-top: 12px; justify-content: flex-end"
          @current-change="loadWithdrawals()" @size-change="loadWithdrawals(1)" />
      </template>

      <!-- 佣金设置 -->
      <template v-else-if="tab === 'commission'">
        <el-form :model="commissionSettings" label-width="180px" style="max-width: 560px">
          <el-form-item label="一级佣金比例(%)">
            <el-input-number v-model="commissionSettings.level1Rate" :min="0" :max="100" :precision="2" />
          </el-form-item>
          <el-form-item label="二级佣金比例(%)">
            <el-input-number v-model="commissionSettings.level2Rate" :min="0" :max="100" :precision="2" />
          </el-form-item>
          <el-form-item label="佣金入账条件">
            <el-select v-model="commissionSettings.postCondition" style="width: 220px">
              <el-option label="出库后入账" value="out_stock" />
              <el-option label="签收后入账" value="signed" />
              <el-option label="付款后入账" value="paid" />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="saveCommission">保存</el-button>
          </el-form-item>
        </el-form>
      </template>

      <!-- 基础设置 -->
      <template v-else>
        <el-form :model="baseSettings" label-width="180px" style="max-width: 560px">
          <el-form-item label="启用分销">
            <el-switch v-model="baseSettings.enabled" />
          </el-form-item>
          <el-form-item label="分销商申请需审核">
            <el-switch v-model="baseSettings.needApprove" />
          </el-form-item>
          <el-form-item label="最低提现金额(元)">
            <el-input-number v-model="baseSettings.minWithdrawal" :min="0" :precision="2" />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="saveBase">保存</el-button>
          </el-form-item>
        </el-form>
      </template>
    </el-card>

    <!-- 新增/编辑分销商 -->
    <el-dialog v-model="dialogVisible" :title="form.id ? '编辑分销商' : '新增分销商'" width="480px">
      <el-form :model="form" label-width="90px">
        <el-form-item label="名称" required><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="手机号"><el-input v-model="form.phone" /></el-form-item>
        <el-form-item label="业务经理">
          <el-select v-model="form.managerId" filterable clearable style="width: 100%">
            <el-option v-for="e in employees" :key="e.id" :label="e.name" :value="e.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="关联客户">
          <el-select v-model="form.customerId" filterable remote clearable :remote-method="searchCustomers" style="width: 100%" placeholder="输入客户名称搜索">
            <el-option v-for="c in customerOptions" :key="c.id" :label="c.name" :value="c.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="备注"><el-input v-model="form.remark" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="save">保存</el-button>
      </template>
    </el-dialog>

    <!-- 新增提现 -->
    <el-dialog v-model="wVisible" title="新增提现" width="440px">
      <el-form label-width="90px">
        <el-form-item label="分销商" required>
          <el-select v-model="wForm.distributorId" filterable style="width: 100%">
            <el-option v-for="d in activeDistributors" :key="d.id" :label="d.name" :value="d.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="提现金额" required>
          <el-input-number v-model="wForm.amount" :min="0.01" :precision="2" style="width: 100%" />
        </el-form-item>
        <el-form-item label="备注"><el-input v-model="wForm.remark" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="wVisible = false">取消</el-button>
        <el-button type="primary" @click="saveWithdrawal">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '@/api/client'
import { useReport } from '@/composables/useReport'

interface Row {
  id: number; name: string; phone: string; managerId: number; manager: string; customerId: number
  status: string; appliedAt: string; approvedAt: string | null; remark: string
  saleAmount: number; unpostedAmount: number; postedAmount: number
  withdrawable: number; withdrawing: number; withdrawn: number; commissionBalance: number
  undistributedDays: number
}
interface Withdrawal { id: number; distributorId: number; distributor: string; amount: number; status: string; appliedAt: string; paidAt: string | null; remark: string }
interface Option { id: number; name: string }

const tab = ref('list')

const { list, total, loading, page, pageSize, filters, load, search, exportCsv } =
  useReport<Row>('/api/v1/distributors', { status: '', keyword: '' })

const exportCols = [
  { key: 'name', label: '分销商名称' },
  { key: 'manager', label: '业务经理' },
  { key: 'saleAmount', label: '分销额' },
  { key: 'withdrawable', label: '可提现佣金' },
  { key: 'withdrawing', label: '提现中佣金' },
  { key: 'withdrawn', label: '已提现佣金' },
  { key: 'undistributedDays', label: '未分销天数' },
]

function formatDate(t: string | null) {
  return t ? new Date(t).toLocaleDateString('zh-CN') : '-'
}
function formatTime(t: string | null) {
  return t ? new Date(t).toLocaleString('zh-CN', { hour12: false }) : '-'
}

// ---- 分销商 CRUD ----
const dialogVisible = ref(false)
const form = ref({ id: 0, name: '', phone: '', managerId: undefined as number | undefined, customerId: undefined as number | undefined, remark: '' })
const employees = ref<Option[]>([])
const customerOptions = ref<Option[]>([])

async function loadEmployees() {
  const res = await api.get('/api/v1/employees', { params: { page: 1, pageSize: 500 } })
  if (res.data.code === 0 || res.data.code === 200) employees.value = res.data.data?.list ?? []
}

async function searchCustomers(kw: string) {
  const res = await api.get('/api/v1/customers', { params: { page: 1, pageSize: 50, keyword: kw } })
  if (res.data.code === 0 || res.data.code === 200) customerOptions.value = res.data.data?.list ?? []
}

function openDialog(row?: Row) {
  form.value = row
    ? { id: row.id, name: row.name, phone: row.phone, managerId: row.managerId || undefined, customerId: row.customerId || undefined, remark: row.remark }
    : { id: 0, name: '', phone: '', managerId: undefined, customerId: undefined, remark: '' }
  dialogVisible.value = true
}

async function save() {
  if (!form.value.name) { ElMessage.warning('请填写分销商名称'); return }
  const res = form.value.id
    ? await api.put(`/api/v1/distributors/${form.value.id}`, form.value)
    : await api.post('/api/v1/distributors', form.value)
  if (res.data.code === 0 || res.data.code === 200) {
    ElMessage.success('保存成功')
    dialogVisible.value = false
    load()
  } else {
    ElMessage.error(res.data.message || '保存失败')
  }
}

async function approve(row: Row) {
  await ElMessageBox.confirm(`确定通过分销商「${row.name}」的审核？`, '审核', { type: 'warning' })
  const res = await api.put(`/api/v1/distributors/${row.id}/approve`)
  if (res.data.code === 0 || res.data.code === 200) {
    ElMessage.success('已通过')
    load()
  }
}

async function disable(row: Row) {
  await ElMessageBox.confirm(`确定停用分销商「${row.name}」？`, '提示', { type: 'warning' })
  const res = await api.put(`/api/v1/distributors/${row.id}/disable`)
  if (res.data.code === 0 || res.data.code === 200) {
    ElMessage.success('已停用')
    load()
  }
}

async function remove(row: Row) {
  await ElMessageBox.confirm(`确定删除分销商「${row.name}」？`, '提示', { type: 'warning' })
  const res = await api.delete(`/api/v1/distributors/${row.id}`)
  if (res.data.code === 0 || res.data.code === 200) {
    ElMessage.success('已删除')
    load()
  }
}

// ---- 佣金提现 ----
const wList = ref<Withdrawal[]>([])
const wTotal = ref(0)
const wPage = ref(1)
const wPageSize = ref(30)
const wLoading = ref(false)
const wStatus = ref('')

async function loadWithdrawals(p = wPage.value) {
  wPage.value = p
  wLoading.value = true
  try {
    const res = await api.get('/api/v1/distributor-withdrawals', { params: { page: wPage.value, pageSize: wPageSize.value, status: wStatus.value } })
    if (res.data.code === 0 || res.data.code === 200) {
      wList.value = res.data.data?.list ?? []
      wTotal.value = res.data.data?.total ?? 0
    }
  } finally {
    wLoading.value = false
  }
}

const wVisible = ref(false)
const wForm = ref({ distributorId: undefined as number | undefined, amount: 0, remark: '' })
const activeDistributors = ref<Option[]>([])

async function openWithdrawal() {
  wForm.value = { distributorId: undefined, amount: 0, remark: '' }
  const res = await api.get('/api/v1/distributors', { params: { page: 1, pageSize: 500, status: 'active' } })
  if (res.data.code === 0 || res.data.code === 200) activeDistributors.value = res.data.data?.list ?? []
  wVisible.value = true
}

async function saveWithdrawal() {
  if (!wForm.value.distributorId || !wForm.value.amount) { ElMessage.warning('请选择分销商并填写金额'); return }
  const res = await api.post('/api/v1/distributor-withdrawals', wForm.value)
  if (res.data.code === 0 || res.data.code === 200) {
    ElMessage.success('已提交')
    wVisible.value = false
    loadWithdrawals()
  } else {
    ElMessage.error(res.data.message || '提交失败')
  }
}

async function payWithdrawal(row: Withdrawal) {
  await ElMessageBox.confirm(`确定已为「${row.distributor}」打款 ${row.amount.toFixed(2)} 元？`, '确认打款', { type: 'warning' })
  const res = await api.put(`/api/v1/distributor-withdrawals/${row.id}/pay`)
  if (res.data.code === 0 || res.data.code === 200) {
    ElMessage.success('已确认')
    loadWithdrawals()
  }
}

async function rejectWithdrawal(row: Withdrawal) {
  await ElMessageBox.confirm(`确定拒绝「${row.distributor}」的提现申请？`, '提示', { type: 'warning' })
  const res = await api.put(`/api/v1/distributor-withdrawals/${row.id}/reject`)
  if (res.data.code === 0 || res.data.code === 200) {
    ElMessage.success('已拒绝')
    loadWithdrawals()
  }
}

// ---- 设置 ----
const commissionSettings = ref({ level1Rate: 0, level2Rate: 0, postCondition: 'out_stock' })
const baseSettings = ref({ enabled: false, needApprove: true, minWithdrawal: 0 })

async function loadSettings() {
  const res = await api.get('/api/v1/settings')
  if (res.data.code === 0 || res.data.code === 200) {
    const s = res.data.data || {}
    commissionSettings.value = {
      level1Rate: Number(s['dist.level1Rate'] || 0),
      level2Rate: Number(s['dist.level2Rate'] || 0),
      postCondition: s['dist.postCondition'] || 'out_stock',
    }
    baseSettings.value = {
      enabled: s['dist.enabled'] === '1',
      needApprove: s['dist.needApprove'] !== '0',
      minWithdrawal: Number(s['dist.minWithdrawal'] || 0),
    }
  }
}

async function saveCommission() {
  const res = await api.put('/api/v1/settings', {
    'dist.level1Rate': String(commissionSettings.value.level1Rate),
    'dist.level2Rate': String(commissionSettings.value.level2Rate),
    'dist.postCondition': commissionSettings.value.postCondition,
  })
  if (res.data.code === 0 || res.data.code === 200) ElMessage.success('已保存')
}

async function saveBase() {
  const res = await api.put('/api/v1/settings', {
    'dist.enabled': baseSettings.value.enabled ? '1' : '0',
    'dist.needApprove': baseSettings.value.needApprove ? '1' : '0',
    'dist.minWithdrawal': String(baseSettings.value.minWithdrawal),
  })
  if (res.data.code === 0 || res.data.code === 200) ElMessage.success('已保存')
}

watch(tab, (t) => {
  if (t === 'withdrawals') loadWithdrawals(1)
})

onMounted(() => {
  load(1)
  loadEmployees()
  loadSettings()
})
</script>

<style scoped>
.page { display: flex; flex-direction: column; gap: 12px; }
.page-header { display: flex; justify-content: space-between; align-items: center; }
.page-title { font-size: 18px; font-weight: 600; margin: 0; }
</style>
