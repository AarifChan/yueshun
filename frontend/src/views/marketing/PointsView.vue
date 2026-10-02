<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">积分</h2>
    </div>
    <el-card>
      <el-tabs v-model="tab">
        <el-tab-pane label="赠送规则设置" name="rules" />
        <el-tab-pane label="金额抵扣设置" name="deduct" />
        <el-tab-pane label="商品兑换设置" name="exchange" />
        <el-tab-pane label="客户积分状况" name="customers" />
      </el-tabs>

      <!-- 赠送规则设置 -->
      <template v-if="tab === 'rules'">
        <el-form :model="grantSettings" label-width="140px" style="max-width: 560px">
          <el-form-item label="首次登录积分">
            <el-input-number v-model="grantSettings.firstLogin" :min="0" />
          </el-form-item>
          <el-form-item label="每日登录积分">
            <el-input-number v-model="grantSettings.dailyLogin" :min="0" />
          </el-form-item>
          <el-form-item label="下单赠送积分">
            <el-input-number v-model="grantSettings.orderGrant" :min="0" />
          </el-form-item>
          <el-form-item label="积分清零设置">
            <el-switch v-model="grantSettings.clearEnabled" active-text="启用清零" />
          </el-form-item>
          <el-form-item v-if="grantSettings.clearEnabled" label="清零周期">
            <el-select v-model="grantSettings.clearCycle" style="width: 200px">
              <el-option label="每年" value="year" />
              <el-option label="每季度" value="quarter" />
              <el-option label="每月" value="month" />
            </el-select>
          </el-form-item>
          <el-form-item v-if="grantSettings.clearEnabled" label="清零日期">
            <el-input v-model="grantSettings.clearDate" placeholder="如 12-31" style="width: 200px" />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="saveGrantSettings">保存</el-button>
          </el-form-item>
        </el-form>

        <el-divider />
        <div class="section-header">
          <h3>赠送规则</h3>
          <el-button size="small" type="primary" @click="openRuleDialog()">新增规则</el-button>
        </div>
        <el-table :data="rules" v-loading="rulesLoading" border stripe>
          <el-table-column label="赠送条件" width="110">
            <template #default="{ row }">{{ conditionLabel(row.condition) }}</template>
          </el-table-column>
          <el-table-column prop="points" label="单笔赠送积分" width="110" align="right" />
          <el-table-column label="商品范围" width="100">
            <template #default="{ row }">{{ row.productScope === 'part' ? '部分商品' : '全部' }}</template>
          </el-table-column>
          <el-table-column label="客户范围" width="100">
            <template #default="{ row }">{{ row.customerScope === 'part' ? '部分客户' : '全部' }}</template>
          </el-table-column>
          <el-table-column label="分类" width="110">
            <template #default="{ row }">{{ categoryName(row.categoryId) }}</template>
          </el-table-column>
          <el-table-column label="品牌" width="110">
            <template #default="{ row }">{{ brandName(row.brandId) }}</template>
          </el-table-column>
          <el-table-column prop="specDate" label="指定日期" width="110" />
          <el-table-column prop="ruleText" label="赠送规则" min-width="150" show-overflow-tooltip />
          <el-table-column label="启用" width="80" align="center">
            <template #default="{ row }">
              <el-tag :type="row.enabled ? 'success' : 'info'">{{ row.enabled ? '启用' : '停用' }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="sort" label="排序" width="70" align="right" />
          <el-table-column label="操作" width="130" fixed="right">
            <template #default="{ row }">
              <el-button link type="primary" @click="openRuleDialog(row)">编辑</el-button>
              <el-button link type="danger" @click="removeRule(row)">删除</el-button>
            </template>
          </el-table-column>
          <template #empty>暂无数据</template>
        </el-table>
      </template>

      <!-- 金额抵扣设置 -->
      <template v-else-if="tab === 'deduct'">
        <el-form :model="deductSettings" label-width="160px" style="max-width: 560px">
          <el-form-item label="启用积分抵扣">
            <el-switch v-model="deductSettings.enabled" />
          </el-form-item>
          <el-form-item label="每积分抵扣金额(元)">
            <el-input-number v-model="deductSettings.perPoint" :min="0" :precision="4" :step="0.01" />
          </el-form-item>
          <el-form-item label="单笔最高抵扣比例(%)">
            <el-input-number v-model="deductSettings.maxPercent" :min="0" :max="100" />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="saveDeductSettings">保存</el-button>
          </el-form-item>
        </el-form>
      </template>

      <!-- 商品兑换设置 -->
      <template v-else-if="tab === 'exchange'">
        <div class="section-header">
          <h3>积分兑换商品</h3>
          <el-button size="small" type="primary" @click="openExchangeDialog()">新增兑换</el-button>
        </div>
        <el-table :data="exchanges" v-loading="exchangeLoading" border stripe>
          <el-table-column prop="code" label="商品编号" width="110" />
          <el-table-column prop="product" label="商品名称" min-width="160" />
          <el-table-column prop="spec" label="规格" width="110" />
          <el-table-column prop="unit" label="单位" width="70" />
          <el-table-column prop="points" label="兑换积分" width="100" align="right" />
          <el-table-column label="启用" width="80" align="center">
            <template #default="{ row }">
              <el-tag :type="row.enabled ? 'success' : 'info'">{{ row.enabled ? '启用' : '停用' }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="sort" label="排序" width="70" align="right" />
          <el-table-column label="操作" width="130" fixed="right">
            <template #default="{ row }">
              <el-button link type="primary" @click="openExchangeDialog(row)">编辑</el-button>
              <el-button link type="danger" @click="removeExchange(row)">删除</el-button>
            </template>
          </el-table-column>
          <template #empty>暂无数据</template>
        </el-table>
      </template>

      <!-- 客户积分状况 -->
      <template v-else>
        <el-form inline @submit.prevent>
          <el-form-item label="客户">
            <el-input v-model="custFilters.keyword" placeholder="客户名称/编号" clearable style="width: 180px" />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="custSearch">搜索</el-button>
            <el-button type="warning" @click="openAdjust()">手工调整积分</el-button>
          </el-form-item>
        </el-form>
        <el-table :data="custList" v-loading="custLoading" border stripe>
          <el-table-column prop="code" label="客户编号" width="110" />
          <el-table-column prop="customer" label="客户名称" min-width="160" />
          <el-table-column prop="points" label="当前积分" width="120" align="right" sortable />
          <el-table-column label="操作" width="120" fixed="right">
            <template #default="{ row }">
              <el-button link type="primary" @click="openFlows(row)">积分明细</el-button>
            </template>
          </el-table-column>
          <template #empty>暂无数据</template>
        </el-table>
        <el-pagination
          v-model:current-page="custPage" v-model:page-size="custPageSize"
          :total="custTotal" :page-sizes="[30, 50, 100]"
          layout="total, sizes, prev, pager, next" style="margin-top: 12px; justify-content: flex-end"
          @current-change="loadCustomers()" @size-change="loadCustomers(1)" />
      </template>
    </el-card>

    <!-- 规则编辑 -->
    <el-dialog v-model="ruleVisible" :title="ruleForm.id ? '编辑规则' : '新增规则'" width="520px">
      <el-form :model="ruleForm" label-width="110px">
        <el-form-item label="赠送条件" required>
          <el-select v-model="ruleForm.condition" style="width: 100%">
            <el-option label="首次登录" value="first_login" />
            <el-option label="每日登录" value="daily_login" />
            <el-option label="下单赠送" value="order" />
          </el-select>
        </el-form-item>
        <el-form-item label="单笔赠送积分" required>
          <el-input-number v-model="ruleForm.points" :min="0" style="width: 100%" />
        </el-form-item>
        <el-form-item label="商品范围">
          <el-radio-group v-model="ruleForm.productScope">
            <el-radio value="all">全部商品</el-radio>
            <el-radio value="part">部分商品</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="客户范围">
          <el-radio-group v-model="ruleForm.customerScope">
            <el-radio value="all">全部客户</el-radio>
            <el-radio value="part">部分客户</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="指定日期">
          <el-input v-model="ruleForm.specDate" placeholder="如 2026-10-01，可留空" />
        </el-form-item>
        <el-form-item label="赠送规则">
          <el-input v-model="ruleForm.ruleText" placeholder="规则说明" />
        </el-form-item>
        <el-form-item label="是否启用"><el-switch v-model="ruleForm.enabled" /></el-form-item>
        <el-form-item label="排序"><el-input-number v-model="ruleForm.sort" :min="0" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="ruleVisible = false">取消</el-button>
        <el-button type="primary" @click="saveRule">保存</el-button>
      </template>
    </el-dialog>

    <!-- 兑换编辑 -->
    <el-dialog v-model="exchangeVisible" :title="exchangeForm.id ? '编辑兑换' : '新增兑换'" width="480px">
      <el-form :model="exchangeForm" label-width="100px">
        <el-form-item label="商品" required>
          <el-select v-model="exchangeForm.productId" filterable remote :remote-method="searchProducts" :disabled="!!exchangeForm.id" style="width: 100%" placeholder="输入商品名称搜索">
            <el-option v-for="p in productOptions" :key="p.id" :label="`${p.name}（${p.code || '无编号'}）`" :value="p.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="兑换积分" required>
          <el-input-number v-model="exchangeForm.points" :min="1" style="width: 100%" />
        </el-form-item>
        <el-form-item label="是否启用"><el-switch v-model="exchangeForm.enabled" /></el-form-item>
        <el-form-item label="排序"><el-input-number v-model="exchangeForm.sort" :min="0" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="exchangeVisible = false">取消</el-button>
        <el-button type="primary" @click="saveExchange">保存</el-button>
      </template>
    </el-dialog>

    <!-- 手工调整 -->
    <el-dialog v-model="adjustVisible" title="手工调整积分" width="480px">
      <el-form label-width="90px">
        <el-form-item label="客户" required>
          <el-select v-model="adjustForm.customerId" filterable remote :remote-method="searchCustomers" style="width: 100%" placeholder="输入客户名称搜索">
            <el-option v-for="c in customerOptions" :key="c.id" :label="c.name" :value="c.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="积分变动" required>
          <el-input-number v-model="adjustForm.change" style="width: 100%" />
          <div class="hint">正数为增加，负数为扣减</div>
        </el-form-item>
        <el-form-item label="备注"><el-input v-model="adjustForm.remark" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="adjustVisible = false">取消</el-button>
        <el-button type="primary" @click="saveAdjust">保存</el-button>
      </template>
    </el-dialog>

    <!-- 积分明细 -->
    <el-dialog v-model="flowsVisible" :title="`积分明细：${flowsCustomer?.customer ?? ''}`" width="640px">
      <el-table :data="flows" v-loading="flowsLoading" border stripe size="small">
        <el-table-column label="时间" width="160">
          <template #default="{ row }">{{ formatTime(row.createdAt) }}</template>
        </el-table-column>
        <el-table-column label="类型" width="100">
          <template #default="{ row }">{{ flowTypeLabel(row.type) }}</template>
        </el-table-column>
        <el-table-column prop="change" label="积分变动" width="100" align="right">
          <template #default="{ row }">
            <span :style="{ color: row.change >= 0 ? '#67c23a' : '#f56c6c' }">{{ row.change > 0 ? '+' : '' }}{{ row.change }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="remark" label="备注" min-width="140" />
        <template #empty>暂无数据</template>
      </el-table>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '@/api/client'

interface Rule {
  id: number; condition: string; points: number; productScope: string; customerScope: string
  categoryId: number; brandId: number; specDate: string; ruleText: string; enabled: boolean; sort: number
}
interface Exchange { id: number; productId: number; product: string; code: string; spec: string; unit: string; points: number; enabled: boolean; sort: number }
interface CustPoint { customerId: number; customer: string; code: string; points: number }
interface Flow { id: number; change: number; type: string; remark: string; createdAt: string }
interface Option { id: number; name: string; code?: string }

const tab = ref('rules')

// ---- 赠送规则设置（存 company_settings）----
const grantSettings = ref({ firstLogin: 0, dailyLogin: 0, orderGrant: 0, clearEnabled: false, clearCycle: 'year', clearDate: '' })
const deductSettings = ref({ enabled: false, perPoint: 0, maxPercent: 0 })

async function loadSettings() {
  const res = await api.get('/api/v1/settings')
  if (res.data.code === 0 || res.data.code === 200) {
    const s = res.data.data || {}
    grantSettings.value = {
      firstLogin: Number(s['point.firstLogin'] || 0),
      dailyLogin: Number(s['point.dailyLogin'] || 0),
      orderGrant: Number(s['point.orderGrant'] || 0),
      clearEnabled: s['point.clearEnabled'] === '1',
      clearCycle: s['point.clearCycle'] || 'year',
      clearDate: s['point.clearDate'] || '',
    }
    deductSettings.value = {
      enabled: s['point.deductEnabled'] === '1',
      perPoint: Number(s['point.deductPerPoint'] || 0),
      maxPercent: Number(s['point.deductMaxPercent'] || 0),
    }
  }
}

async function saveGrantSettings() {
  const res = await api.put('/api/v1/settings', {
    'point.firstLogin': String(grantSettings.value.firstLogin),
    'point.dailyLogin': String(grantSettings.value.dailyLogin),
    'point.orderGrant': String(grantSettings.value.orderGrant),
    'point.clearEnabled': grantSettings.value.clearEnabled ? '1' : '0',
    'point.clearCycle': grantSettings.value.clearCycle,
    'point.clearDate': grantSettings.value.clearDate,
  })
  if (res.data.code === 0 || res.data.code === 200) ElMessage.success('已保存')
}

async function saveDeductSettings() {
  const res = await api.put('/api/v1/settings', {
    'point.deductEnabled': deductSettings.value.enabled ? '1' : '0',
    'point.deductPerPoint': String(deductSettings.value.perPoint),
    'point.deductMaxPercent': String(deductSettings.value.maxPercent),
  })
  if (res.data.code === 0 || res.data.code === 200) ElMessage.success('已保存')
}

// ---- 赠送规则 CRUD ----
const rules = ref<Rule[]>([])
const rulesLoading = ref(false)
const categories = ref<Option[]>([])
const brands = ref<Option[]>([])

function conditionLabel(v: string) {
  return v === 'first_login' ? '首次登录' : v === 'daily_login' ? '每日登录' : '下单赠送'
}
function categoryName(id: number) {
  return id ? categories.value.find((c) => c.id === id)?.name ?? '-' : '全部'
}
function brandName(id: number) {
  return id ? brands.value.find((b) => b.id === id)?.name ?? '-' : '全部'
}

async function loadRules() {
  rulesLoading.value = true
  try {
    const res = await api.get('/api/v1/point-rules')
    if (res.data.code === 0 || res.data.code === 200) rules.value = res.data.data?.list ?? []
  } finally {
    rulesLoading.value = false
  }
}

const ruleVisible = ref(false)
const ruleForm = ref<Rule>({ id: 0, condition: 'order', points: 0, productScope: 'all', customerScope: 'all', categoryId: 0, brandId: 0, specDate: '', ruleText: '', enabled: true, sort: 0 })

function openRuleDialog(row?: Rule) {
  ruleForm.value = row ? { ...row } : { id: 0, condition: 'order', points: 0, productScope: 'all', customerScope: 'all', categoryId: 0, brandId: 0, specDate: '', ruleText: '', enabled: true, sort: 0 }
  ruleVisible.value = true
}

async function saveRule() {
  const res = ruleForm.value.id
    ? await api.put(`/api/v1/point-rules/${ruleForm.value.id}`, ruleForm.value)
    : await api.post('/api/v1/point-rules', ruleForm.value)
  if (res.data.code === 0 || res.data.code === 200) {
    ElMessage.success('保存成功')
    ruleVisible.value = false
    loadRules()
  } else {
    ElMessage.error(res.data.message || '保存失败')
  }
}

async function removeRule(row: Rule) {
  await ElMessageBox.confirm('确定删除该规则？', '提示', { type: 'warning' })
  const res = await api.delete(`/api/v1/point-rules/${row.id}`)
  if (res.data.code === 0 || res.data.code === 200) {
    ElMessage.success('已删除')
    loadRules()
  }
}

// ---- 商品兑换 ----
const exchanges = ref<Exchange[]>([])
const exchangeLoading = ref(false)
const exchangeVisible = ref(false)
const exchangeForm = ref({ id: 0, productId: undefined as number | undefined, points: 100, enabled: true, sort: 0 })
const productOptions = ref<Option[]>([])

async function loadExchanges() {
  exchangeLoading.value = true
  try {
    const res = await api.get('/api/v1/point-exchanges')
    if (res.data.code === 0 || res.data.code === 200) exchanges.value = res.data.data?.list ?? []
  } finally {
    exchangeLoading.value = false
  }
}

async function searchProducts(kw: string) {
  const res = await api.get('/api/v1/products', { params: { page: 1, pageSize: 50, keyword: kw } })
  if (res.data.code === 0 || res.data.code === 200) productOptions.value = res.data.data?.list ?? []
}

function openExchangeDialog(row?: Exchange) {
  if (row) {
    exchangeForm.value = { id: row.id, productId: row.productId, points: row.points, enabled: row.enabled, sort: row.sort }
    productOptions.value = [{ id: row.productId, name: row.product, code: row.code }]
  } else {
    exchangeForm.value = { id: 0, productId: undefined, points: 100, enabled: true, sort: 0 }
    searchProducts('')
  }
  exchangeVisible.value = true
}

async function saveExchange() {
  if (!exchangeForm.value.productId) { ElMessage.warning('请选择商品'); return }
  const res = exchangeForm.value.id
    ? await api.put(`/api/v1/point-exchanges/${exchangeForm.value.id}`, exchangeForm.value)
    : await api.post('/api/v1/point-exchanges', exchangeForm.value)
  if (res.data.code === 0 || res.data.code === 200) {
    ElMessage.success('保存成功')
    exchangeVisible.value = false
    loadExchanges()
  } else {
    ElMessage.error(res.data.message || '保存失败')
  }
}

async function removeExchange(row: Exchange) {
  await ElMessageBox.confirm(`确定删除「${row.product}」的兑换设置？`, '提示', { type: 'warning' })
  const res = await api.delete(`/api/v1/point-exchanges/${row.id}`)
  if (res.data.code === 0 || res.data.code === 200) {
    ElMessage.success('已删除')
    loadExchanges()
  }
}

// ---- 客户积分 ----
const custList = ref<CustPoint[]>([])
const custTotal = ref(0)
const custPage = ref(1)
const custPageSize = ref(30)
const custLoading = ref(false)
const custFilters = ref({ keyword: '' })

async function loadCustomers(p = custPage.value) {
  custPage.value = p
  custLoading.value = true
  try {
    const res = await api.get('/api/v1/customer-points', { params: { page: custPage.value, pageSize: custPageSize.value, keyword: custFilters.value.keyword } })
    if (res.data.code === 0 || res.data.code === 200) {
      custList.value = res.data.data?.list ?? []
      custTotal.value = res.data.data?.total ?? 0
    }
  } finally {
    custLoading.value = false
  }
}

function custSearch() { loadCustomers(1) }

// ---- 手工调整 ----
const adjustVisible = ref(false)
const adjustForm = ref({ customerId: undefined as number | undefined, change: 0, remark: '' })
const customerOptions = ref<Option[]>([])

async function searchCustomers(kw: string) {
  const res = await api.get('/api/v1/customers', { params: { page: 1, pageSize: 50, keyword: kw } })
  if (res.data.code === 0 || res.data.code === 200) customerOptions.value = res.data.data?.list ?? []
}

function openAdjust() {
  adjustForm.value = { customerId: undefined, change: 0, remark: '' }
  searchCustomers('')
  adjustVisible.value = true
}

async function saveAdjust() {
  if (!adjustForm.value.customerId || !adjustForm.value.change) { ElMessage.warning('请选择客户并填写积分变动'); return }
  const res = await api.post('/api/v1/point-flows', adjustForm.value)
  if (res.data.code === 0 || res.data.code === 200) {
    ElMessage.success('已调整')
    adjustVisible.value = false
    loadCustomers()
  } else {
    ElMessage.error(res.data.message || '调整失败')
  }
}

// ---- 积分明细 ----
const flowsVisible = ref(false)
const flows = ref<Flow[]>([])
const flowsLoading = ref(false)
const flowsCustomer = ref<CustPoint | null>(null)

function flowTypeLabel(t: string) {
  const map: Record<string, string> = { grant: '赠送', deduct: '抵扣', exchange: '兑换', clear: '清零', manual: '手工调整' }
  return map[t] ?? t
}

async function openFlows(row: CustPoint) {
  flowsCustomer.value = row
  flowsVisible.value = true
  flowsLoading.value = true
  try {
    const res = await api.get('/api/v1/point-flows', { params: { page: 1, pageSize: 100, customerId: row.customerId } })
    if (res.data.code === 0 || res.data.code === 200) flows.value = res.data.data?.list ?? []
  } finally {
    flowsLoading.value = false
  }
}

function formatTime(t: string) {
  return t ? new Date(t).toLocaleString('zh-CN', { hour12: false }) : ''
}

onMounted(() => {
  loadSettings()
  loadRules()
  loadExchanges()
  loadCustomers(1)
})
</script>

<style scoped>
.page { display: flex; flex-direction: column; gap: 12px; }
.page-header { display: flex; justify-content: space-between; align-items: center; }
.page-title { font-size: 18px; font-weight: 600; margin: 0; }
.section-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 10px; }
.section-header h3 { margin: 0; font-size: 15px; }
.hint { font-size: 12px; color: #909399; line-height: 1.4; }
</style>
