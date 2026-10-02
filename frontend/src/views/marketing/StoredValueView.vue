<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">预收款储值</h2>
      <div class="actions">
        <el-button type="primary" @click="openRecord()">新增储值</el-button>
      </div>
    </div>
    <el-card>
      <el-tabs v-model="tab">
        <el-tab-pane label="储值记录" name="records" />
        <el-tab-pane label="储值规则" name="rules" />
      </el-tabs>

      <!-- 储值记录 -->
      <template v-if="tab === 'records'">
        <el-form inline @submit.prevent>
          <el-form-item label="客户">
            <el-input v-model="recordKeyword" placeholder="客户名称" clearable style="width: 180px" />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="loadRecords(1)">搜索</el-button>
          </el-form-item>
        </el-form>
        <el-table :data="records" v-loading="recordLoading" border stripe>
          <el-table-column label="储值时间" width="160">
            <template #default="{ row }">{{ formatTime(row.createdAt) }}</template>
          </el-table-column>
          <el-table-column prop="customer" label="客户" min-width="150" />
          <el-table-column prop="rule" label="储值规则" width="140">
            <template #default="{ row }">{{ row.rule || '自定义' }}</template>
          </el-table-column>
          <el-table-column prop="amount" label="储值金额" width="110" align="right">
            <template #default="{ row }">{{ row.amount?.toFixed(2) }}</template>
          </el-table-column>
          <el-table-column prop="giftAmount" label="赠送金额" width="110" align="right">
            <template #default="{ row }">{{ row.giftAmount?.toFixed(2) }}</template>
          </el-table-column>
          <el-table-column prop="account" label="收款账户" width="130">
            <template #default="{ row }">{{ row.account || '-' }}</template>
          </el-table-column>
          <el-table-column prop="remark" label="备注" min-width="120" show-overflow-tooltip />
          <template #empty>
            <el-empty description="空空如也，赶快创建～" />
          </template>
        </el-table>
        <el-pagination
          v-model:current-page="recordPage" v-model:page-size="recordPageSize"
          :total="recordTotal" :page-sizes="[30, 50, 100]"
          layout="total, sizes, prev, pager, next" style="margin-top: 12px; justify-content: flex-end"
          @current-change="loadRecords()" @size-change="loadRecords(1)" />
      </template>

      <!-- 储值规则 -->
      <template v-else>
        <div class="section-header">
          <h3>储值规则</h3>
          <el-button size="small" type="primary" @click="openRule()">新增储值规则</el-button>
        </div>
        <el-table :data="rules" v-loading="ruleLoading" border stripe>
          <el-table-column prop="name" label="规则名称" min-width="150" />
          <el-table-column prop="amount" label="储值金额" width="120" align="right">
            <template #default="{ row }">{{ row.amount?.toFixed(2) }}</template>
          </el-table-column>
          <el-table-column prop="giftAmount" label="赠送金额" width="120" align="right">
            <template #default="{ row }">{{ row.giftAmount?.toFixed(2) }}</template>
          </el-table-column>
          <el-table-column label="状态" width="90" align="center">
            <template #default="{ row }">
              <el-tag :type="row.status === 1 ? 'success' : 'info'">{{ row.status === 1 ? '启用' : '停用' }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="sort" label="排序" width="70" align="right" />
          <el-table-column prop="remark" label="备注" min-width="120" />
          <el-table-column label="操作" width="130" fixed="right">
            <template #default="{ row }">
              <el-button link type="primary" @click="openRule(row)">编辑</el-button>
              <el-button link type="danger" @click="removeRule(row)">删除</el-button>
            </template>
          </el-table-column>
          <template #empty>
            <el-empty description="空空如也，赶快创建～" />
          </template>
        </el-table>
      </template>
    </el-card>

    <!-- 新增储值 -->
    <el-dialog v-model="recordVisible" title="新增储值" width="520px">
      <el-form :model="recordForm" label-width="90px">
        <el-form-item label="客户" required>
          <el-select v-model="recordForm.customerId" filterable remote :remote-method="searchCustomers" style="width: 100%" placeholder="输入客户名称搜索">
            <el-option v-for="c in customerOptions" :key="c.id" :label="c.name" :value="c.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="储值规则">
          <el-select v-model="recordForm.ruleId" clearable placeholder="选择规则自动带出金额" style="width: 100%" @change="onRuleChange">
            <el-option v-for="r in enabledRules" :key="r.id" :label="`${r.name}（储${r.amount.toFixed(2)}赠${r.giftAmount.toFixed(2)}）`" :value="r.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="储值金额" required>
          <el-input-number v-model="recordForm.amount" :min="0.01" :precision="2" style="width: 100%" />
        </el-form-item>
        <el-form-item label="赠送金额">
          <el-input-number v-model="recordForm.giftAmount" :min="0" :precision="2" style="width: 100%" />
        </el-form-item>
        <el-form-item label="收款账户">
          <el-select v-model="recordForm.accountId" clearable style="width: 100%">
            <el-option v-for="a in accounts" :key="a.id" :label="a.name" :value="a.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="备注"><el-input v-model="recordForm.remark" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="recordVisible = false">取消</el-button>
        <el-button type="primary" @click="saveRecord">保存</el-button>
      </template>
    </el-dialog>

    <!-- 规则编辑 -->
    <el-dialog v-model="ruleVisible" :title="ruleForm.id ? '编辑规则' : '新增储值规则'" width="480px">
      <el-form :model="ruleForm" label-width="90px">
        <el-form-item label="规则名称" required><el-input v-model="ruleForm.name" /></el-form-item>
        <el-form-item label="储值金额" required>
          <el-input-number v-model="ruleForm.amount" :min="0.01" :precision="2" style="width: 100%" />
        </el-form-item>
        <el-form-item label="赠送金额">
          <el-input-number v-model="ruleForm.giftAmount" :min="0" :precision="2" style="width: 100%" />
        </el-form-item>
        <el-form-item label="是否启用"><el-switch v-model="ruleForm.enabled" /></el-form-item>
        <el-form-item label="排序"><el-input-number v-model="ruleForm.sort" :min="0" /></el-form-item>
        <el-form-item label="备注"><el-input v-model="ruleForm.remark" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="ruleVisible = false">取消</el-button>
        <el-button type="primary" @click="saveRule">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '@/api/client'

interface Record { id: number; customerId: number; customer: string; ruleId: number; rule: string; amount: number; giftAmount: number; accountId: number; account: string; remark: string; createdAt: string }
interface Rule { id: number; name: string; amount: number; giftAmount: number; status: number; sort: number; remark: string }
interface Option { id: number; name: string }

const tab = ref('records')

// ---- 储值记录 ----
const records = ref<Record[]>([])
const recordTotal = ref(0)
const recordPage = ref(1)
const recordPageSize = ref(30)
const recordLoading = ref(false)
const recordKeyword = ref('')

async function loadRecords(p = recordPage.value) {
  recordPage.value = p
  recordLoading.value = true
  try {
    const res = await api.get('/api/v1/stored-value-records', { params: { page: recordPage.value, pageSize: recordPageSize.value, keyword: recordKeyword.value } })
    if (res.data.code === 0 || res.data.code === 200) {
      records.value = res.data.data?.list ?? []
      recordTotal.value = res.data.data?.total ?? 0
    }
  } finally {
    recordLoading.value = false
  }
}

function formatTime(t: string) {
  return t ? new Date(t).toLocaleString('zh-CN', { hour12: false }) : ''
}

// ---- 储值规则 ----
const rules = ref<Rule[]>([])
const ruleLoading = ref(false)
const enabledRules = computed(() => rules.value.filter((r) => r.status === 1))

async function loadRules() {
  ruleLoading.value = true
  try {
    const res = await api.get('/api/v1/stored-value-rules')
    if (res.data.code === 0 || res.data.code === 200) rules.value = res.data.data?.list ?? []
  } finally {
    ruleLoading.value = false
  }
}

const ruleVisible = ref(false)
const ruleForm = ref({ id: 0, name: '', amount: 0, giftAmount: 0, enabled: true, sort: 0, remark: '' })

function openRule(row?: Rule) {
  ruleForm.value = row
    ? { id: row.id, name: row.name, amount: row.amount, giftAmount: row.giftAmount, enabled: row.status === 1, sort: row.sort, remark: row.remark }
    : { id: 0, name: '', amount: 0, giftAmount: 0, enabled: true, sort: 0, remark: '' }
  ruleVisible.value = true
}

async function saveRule() {
  if (!ruleForm.value.name) { ElMessage.warning('请填写规则名称'); return }
  const payload = { name: ruleForm.value.name, amount: ruleForm.value.amount, giftAmount: ruleForm.value.giftAmount, status: ruleForm.value.enabled ? 1 : 0, sort: ruleForm.value.sort, remark: ruleForm.value.remark }
  const res = ruleForm.value.id
    ? await api.put(`/api/v1/stored-value-rules/${ruleForm.value.id}`, payload)
    : await api.post('/api/v1/stored-value-rules', payload)
  if (res.data.code === 0 || res.data.code === 200) {
    ElMessage.success('保存成功')
    ruleVisible.value = false
    loadRules()
  } else {
    ElMessage.error(res.data.message || '保存失败')
  }
}

async function removeRule(row: Rule) {
  await ElMessageBox.confirm(`确定删除规则「${row.name}」？`, '提示', { type: 'warning' })
  const res = await api.delete(`/api/v1/stored-value-rules/${row.id}`)
  if (res.data.code === 0 || res.data.code === 200) {
    ElMessage.success('已删除')
    loadRules()
  }
}

// ---- 新增储值 ----
const recordVisible = ref(false)
const recordForm = ref({ customerId: undefined as number | undefined, ruleId: undefined as number | undefined, amount: 0, giftAmount: 0, accountId: undefined as number | undefined, remark: '' })
const customerOptions = ref<Option[]>([])
const accounts = ref<Option[]>([])

async function searchCustomers(kw: string) {
  const res = await api.get('/api/v1/customers', { params: { page: 1, pageSize: 50, keyword: kw } })
  if (res.data.code === 0 || res.data.code === 200) customerOptions.value = res.data.data?.list ?? []
}

async function loadAccounts() {
  const res = await api.get('/api/v1/accounts', { params: { page: 1, pageSize: 500 } })
  if (res.data.code === 0 || res.data.code === 200) accounts.value = res.data.data?.list ?? []
}

function onRuleChange(ruleId: number | undefined) {
  const rule = rules.value.find((r) => r.id === ruleId)
  if (rule) {
    recordForm.value.amount = rule.amount
    recordForm.value.giftAmount = rule.giftAmount
  }
}

function openRecord() {
  recordForm.value = { customerId: undefined, ruleId: undefined, amount: 0, giftAmount: 0, accountId: undefined, remark: '' }
  searchCustomers('')
  recordVisible.value = true
}

async function saveRecord() {
  if (!recordForm.value.customerId) { ElMessage.warning('请选择客户'); return }
  if (!recordForm.value.amount) { ElMessage.warning('请填写储值金额'); return }
  const res = await api.post('/api/v1/stored-value-records', recordForm.value)
  if (res.data.code === 0 || res.data.code === 200) {
    ElMessage.success('储值成功')
    recordVisible.value = false
    tab.value = 'records'
    loadRecords()
  } else {
    ElMessage.error(res.data.message || '储值失败')
  }
}

onMounted(() => {
  loadRecords(1)
  loadRules()
  loadAccounts()
})
</script>

<style scoped>
.page { display: flex; flex-direction: column; gap: 12px; }
.page-header { display: flex; justify-content: space-between; align-items: center; }
.page-title { font-size: 18px; font-weight: 600; margin: 0; }
.actions { display: flex; gap: 8px; }
.section-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 10px; }
.section-header h3 { margin: 0; font-size: 15px; }
</style>
