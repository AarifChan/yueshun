<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">客户公海</h2>
      <div>
        <el-button @click="doExport">导出</el-button>
        <el-button type="warning" @click="recycleDialogVisible = true">回收客户到公海</el-button>
      </div>
    </div>
    <el-card>
      <el-tabs v-model="tab" @tab-change="onTabChange">
        <el-tab-pane label="全部公海客户" name="all" />
        <el-tab-pane label="规定时间未有交易" name="noTrade" />
        <el-tab-pane label="待领取/分配" name="pending" />
        <el-tab-pane label="已领取/分配" name="claimed" />
      </el-tabs>
      <el-form inline>
        <el-form-item label="客户">
          <el-input v-model="filters.keyword" placeholder="客户/编号/联系人/手机号" clearable @keyup.enter="search" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="search">搜索</el-button>
        </el-form-item>
      </el-form>
      <el-table :data="list" v-loading="loading" border stripe>
        <el-table-column prop="name" label="客户名称" min-width="130" fixed="left" />
        <el-table-column prop="code" label="客户编号" width="100" />
        <el-table-column prop="contact" label="联系人" width="90" />
        <el-table-column prop="category" label="客户分类" width="100" />
        <el-table-column prop="region" label="销售区域" width="100" />
        <el-table-column prop="address" label="公司地址" min-width="140" show-overflow-tooltip />
        <el-table-column prop="recycleCount" label="被回收次数" width="90" align="right" />
        <el-table-column prop="claimCount" label="被领取/分配次数" width="120" align="right" />
        <el-table-column prop="lastRecycleAt" label="本次回收时间" width="140" />
        <el-table-column prop="lastRecycleType" label="本次回收方式" width="100" />
        <el-table-column prop="lastRecycleReason" label="本次回收原因" width="120" show-overflow-tooltip />
        <el-table-column prop="lastFollower" label="最后跟进人" width="90" />
        <el-table-column prop="ownerName" label="当前负责人" width="90" />
        <el-table-column label="操作" width="130" fixed="right">
          <template #default="{ row }">
            <template v-if="row.seaStatus === 'public'">
              <el-button link type="primary" @click="claim(row)">领取</el-button>
              <el-button link type="warning" @click="openAssign(row)">分配</el-button>
            </template>
            <el-tag v-else type="success">已领取</el-tag>
          </template>
        </el-table-column>
        <template #empty>暂无数据</template>
      </el-table>
      <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" layout="total, prev, pager, next" @current-change="load" class="pagination" />
    </el-card>

    <el-dialog v-model="recycleDialogVisible" title="回收客户到公海" width="480px">
      <el-form label-width="90px">
        <el-form-item label="客户" required>
          <el-select v-model="recycleForm.customerId" filterable remote :remote-method="searchCustomers" placeholder="输入客户名称搜索" style="width: 100%">
            <el-option v-for="cu in customerOptions" :key="cu.id" :label="cu.name" :value="cu.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="回收原因">
          <el-input v-model="recycleForm.reason" type="textarea" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="recycleDialogVisible = false">取消</el-button>
        <el-button type="warning" @click="recycle">回收</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="assignDialogVisible" title="分配客户" width="420px">
      <el-form label-width="90px">
        <el-form-item label="客户"><span>{{ assignRow?.name }}</span></el-form-item>
        <el-form-item label="分配给" required>
          <el-select v-model="assignEmployeeId" filterable style="width: 100%">
            <el-option v-for="e in employees" :key="e.id" :label="e.name" :value="e.id" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="assignDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="assign">确定</el-button>
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
  id: number; name: string; code: string; contact: string; category: string; region: string
  address: string; recycleCount: number; claimCount: number
  lastRecycleAt: string; lastRecycleType: string; lastRecycleReason: string
  lastFollower: string; ownerName: string; seaStatus: string
}
interface Opt { id: number; name: string }

const { list, total, loading, page, pageSize, filters, load, search, exportCsv } = useReport<Row>(
  '/api/v1/crm/sea/customers', { tab: 'all', keyword: '' })
const tab = ref('all')

const recycleDialogVisible = ref(false)
const recycleForm = ref({ customerId: null as number | null, reason: '' })
const customerOptions = ref<Opt[]>([])
const assignDialogVisible = ref(false)
const assignRow = ref<Row | null>(null)
const assignEmployeeId = ref<number | null>(null)
const employees = ref<Opt[]>([])

function syncTab() { filters.value.tab = tab.value }
function onTabChange() { syncTab(); load(1) }

async function searchCustomers(kw: string) {
  const res = await api.get('/api/v1/customers', { params: { page: 1, pageSize: 20, keyword: kw } })
  if (res.data.code === 0 || res.data.code === 200) {
    customerOptions.value = res.data.data?.list ?? []
  }
}

async function recycle() {
  if (!recycleForm.value.customerId) { ElMessage.warning('请选择客户'); return }
  const res = await api.post('/api/v1/crm/sea/recycle', recycleForm.value)
  if (res.data.code === 0 || res.data.code === 200) {
    ElMessage.success('已回收到公海')
    recycleDialogVisible.value = false
    recycleForm.value = { customerId: null, reason: '' }
    load()
  } else {
    ElMessage.error(res.data.message || '回收失败')
  }
}

async function claim(row: Row) {
  await ElMessageBox.confirm(`确定领取客户「${row.name}」？`, '提示', { type: 'info' })
  const res = await api.post('/api/v1/crm/sea/claim', { customerId: row.id })
  if (res.data.code === 0 || res.data.code === 200) {
    ElMessage.success('领取成功')
    load()
  } else {
    ElMessage.error(res.data.message || '领取失败')
  }
}

function openAssign(row: Row) {
  assignRow.value = row
  assignEmployeeId.value = null
  assignDialogVisible.value = true
}

async function assign() {
  if (!assignEmployeeId.value) { ElMessage.warning('请选择职员'); return }
  const res = await api.post('/api/v1/crm/sea/assign', { customerId: assignRow.value!.id, employeeId: assignEmployeeId.value })
  if (res.data.code === 0 || res.data.code === 200) {
    ElMessage.success('分配成功')
    assignDialogVisible.value = false
    load()
  } else {
    ElMessage.error(res.data.message || '分配失败')
  }
}

function doExport() {
  exportCsv('客户公海', [
    { key: 'name', label: '客户名称' }, { key: 'code', label: '客户编号' },
    { key: 'contact', label: '联系人' }, { key: 'category', label: '客户分类' },
    { key: 'region', label: '销售区域' }, { key: 'recycleCount', label: '被回收次数' },
    { key: 'claimCount', label: '被领取次数' }, { key: 'lastRecycleAt', label: '本次回收时间' },
    { key: 'lastRecycleType', label: '本次回收方式' }, { key: 'lastRecycleReason', label: '本次回收原因' },
    { key: 'lastFollower', label: '最后跟进人' },
  ])
}

onMounted(async () => {
  syncTab()
  load(1)
  const res = await api.get('/api/v1/employees', { params: { page: 1, pageSize: 200 } })
  if (res.data.code === 0 || res.data.code === 200) employees.value = res.data.data?.list ?? []
})
</script>

<style scoped>
.page { display: flex; flex-direction: column; gap: 12px; }
.page-header { display: flex; justify-content: space-between; align-items: center; }
.page-title { font-size: 18px; font-weight: 600; margin: 0; }
.pagination { margin-top: 16px; justify-content: flex-end; }
</style>
