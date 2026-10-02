<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">优惠券</h2>
      <el-button v-if="tab === 'list'" type="primary" @click="openDialog()">新增优惠券</el-button>
    </div>
    <el-card>
      <el-tabs v-model="tab">
        <el-tab-pane label="优惠券列表" name="list" />
        <el-tab-pane label="客户优惠券领用统计" name="stats" />
        <el-tab-pane label="优惠券领取详情" name="received" />
        <el-tab-pane label="优惠券使用详情" name="used" />
      </el-tabs>

      <!-- 优惠券列表 -->
      <template v-if="tab === 'list'">
        <el-form inline @submit.prevent>
          <el-form-item label="创建时间">
            <el-date-picker v-model="range" type="daterange" value-format="YYYY-MM-DD" start-placeholder="开始时间" end-placeholder="结束时间" style="width: 260px" />
          </el-form-item>
          <el-form-item label="优惠券名称">
            <el-input v-model="filters.name" clearable style="width: 160px" />
          </el-form-item>
          <el-form-item label="活动状态">
            <el-select v-model="filters.status" style="width: 120px">
              <el-option label="全部" value="" />
              <el-option label="启用" value="enabled" />
              <el-option label="停用" value="disabled" />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="search">搜索</el-button>
            <el-button @click="reset">重置</el-button>
          </el-form-item>
        </el-form>
        <el-table :data="list" v-loading="loading" border stripe>
          <el-table-column label="创建时间" width="160">
            <template #default="{ row }">{{ formatTime(row.createdAt) }}</template>
          </el-table-column>
          <el-table-column prop="name" label="优惠券名称" min-width="150" show-overflow-tooltip />
          <el-table-column label="发放时间" width="160">
            <template #default="{ row }">{{ formatTime(row.createdAt) }}</template>
          </el-table-column>
          <el-table-column prop="faceValue" label="面值" width="90" align="right">
            <template #default="{ row }">{{ row.faceValue?.toFixed(2) }}</template>
          </el-table-column>
          <el-table-column label="使用门槛" width="110" align="right">
            <template #default="{ row }">{{ row.minAmount > 0 ? `满${row.minAmount.toFixed(2)}` : '无门槛' }}</template>
          </el-table-column>
          <el-table-column label="状态" width="80" align="center">
            <template #default="{ row }">
              <el-tag :type="row.status === 1 ? 'success' : 'info'">{{ row.status === 1 ? '启用' : '停用' }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="有效期" width="200">
            <template #default="{ row }">{{ formatDate(row.validStart) }} ~ {{ formatDate(row.validEnd) }}</template>
          </el-table-column>
          <el-table-column label="优惠券总量" width="100" align="right">
            <template #default="{ row }">{{ row.totalQty > 0 ? row.totalQty : '不限' }}</template>
          </el-table-column>
          <el-table-column prop="received" label="已领取" width="80" align="right" />
          <el-table-column prop="used" label="已使用" width="80" align="right" />
          <el-table-column prop="sort" label="排序" width="70" align="right" />
          <el-table-column label="操作" width="190" fixed="right">
            <template #default="{ row }">
              <el-button link type="success" @click="openGrant(row)">发放优惠券</el-button>
              <el-button link type="primary" @click="openDialog(row)">编辑</el-button>
              <el-button link type="danger" @click="remove(row)">删除</el-button>
            </template>
          </el-table-column>
          <template #empty>暂无搜索结果</template>
        </el-table>
        <el-pagination
          v-model:current-page="page" v-model:page-size="pageSize"
          :total="total" :page-sizes="[30, 50, 100]"
          layout="total, sizes, prev, pager, next" style="margin-top: 12px; justify-content: flex-end"
          @current-change="load()" @size-change="load(1)" />
      </template>

      <!-- 领用统计 -->
      <template v-else-if="tab === 'stats'">
        <el-table :data="statsList" v-loading="statsLoading" border stripe>
          <el-table-column prop="couponName" label="优惠券名称" min-width="160" />
          <el-table-column prop="faceValue" label="面值" width="100" align="right">
            <template #default="{ row }">{{ row.faceValue?.toFixed(2) }}</template>
          </el-table-column>
          <el-table-column prop="received" label="已领取" width="100" align="right" />
          <el-table-column prop="used" label="已使用" width="100" align="right" />
          <el-table-column label="使用率" width="100" align="right">
            <template #default="{ row }">{{ row.received > 0 ? ((row.used / row.received) * 100).toFixed(1) + '%' : '-' }}</template>
          </el-table-column>
          <template #empty>暂无数据</template>
        </el-table>
      </template>

      <!-- 领取/使用详情 -->
      <template v-else>
        <el-form inline @submit.prevent>
          <el-form-item label="客户">
            <el-input v-model="grantFilters.keyword" placeholder="客户名称/编号" clearable style="width: 180px" />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="loadGrants(1)">搜索</el-button>
          </el-form-item>
        </el-form>
        <el-table :data="grantList" v-loading="grantLoading" border stripe>
          <el-table-column prop="couponName" label="优惠券名称" min-width="150" />
          <el-table-column prop="faceValue" label="面值" width="90" align="right">
            <template #default="{ row }">{{ row.faceValue?.toFixed(2) }}</template>
          </el-table-column>
          <el-table-column prop="customer" label="客户" min-width="150" />
          <el-table-column label="领取时间" width="160">
            <template #default="{ row }">{{ formatTime(row.receivedAt) }}</template>
          </el-table-column>
          <el-table-column v-if="tab === 'used'" label="使用时间" width="160">
            <template #default="{ row }">{{ formatTime(row.usedAt) }}</template>
          </el-table-column>
          <el-table-column v-if="tab === 'used'" prop="refBillNo" label="使用单号" width="140" />
          <el-table-column label="状态" width="90" align="center">
            <template #default="{ row }">
              <el-tag :type="row.status === 'used' ? 'success' : 'info'">{{ row.status === 'used' ? '已使用' : '未使用' }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column v-if="tab === 'received'" label="操作" width="110" fixed="right">
            <template #default="{ row }">
              <el-button v-if="row.status === 'unused'" link type="success" @click="markUsed(row)">标记使用</el-button>
            </template>
          </el-table-column>
          <template #empty>暂无数据</template>
        </el-table>
        <el-pagination
          v-model:current-page="grantPage" v-model:page-size="grantPageSize"
          :total="grantTotal" :page-sizes="[30, 50, 100]"
          layout="total, sizes, prev, pager, next" style="margin-top: 12px; justify-content: flex-end"
          @current-change="loadGrants()" @size-change="loadGrants(1)" />
      </template>
    </el-card>

    <!-- 新增/编辑优惠券 -->
    <el-dialog v-model="dialogVisible" :title="form.id ? '编辑优惠券' : '新增优惠券'" width="520px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="优惠券名称" required><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="面值" required>
          <el-input-number v-model="form.faceValue" :min="0.01" :precision="2" style="width: 100%" />
        </el-form-item>
        <el-form-item label="使用门槛">
          <el-input-number v-model="form.minAmount" :min="0" :precision="2" style="width: 100%" />
          <div class="hint">满多少元可用，0 为无门槛</div>
        </el-form-item>
        <el-form-item label="发放总量">
          <el-input-number v-model="form.totalQty" :min="0" style="width: 100%" />
          <div class="hint">0 为不限量</div>
        </el-form-item>
        <el-form-item label="有效期">
          <el-date-picker v-model="form.validRange" type="daterange" value-format="YYYY-MM-DD" start-placeholder="开始日期" end-placeholder="结束日期" style="width: 100%" />
        </el-form-item>
        <el-form-item label="是否启用"><el-switch v-model="form.enabled" /></el-form-item>
        <el-form-item label="排序"><el-input-number v-model="form.sort" :min="0" /></el-form-item>
        <el-form-item label="备注"><el-input v-model="form.remark" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="save">保存</el-button>
      </template>
    </el-dialog>

    <!-- 发放优惠券 -->
    <el-dialog v-model="grantVisible" :title="`发放优惠券：${grantCoupon?.name ?? ''}`" width="520px">
      <el-form label-width="90px">
        <el-form-item label="客户" required>
          <el-select v-model="grantCustomerIds" multiple filterable remote :remote-method="searchCustomers" placeholder="输入客户名称搜索" style="width: 100%">
            <el-option v-for="c in customerOptions" :key="c.id" :label="c.name" :value="c.id" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="grantVisible = false">取消</el-button>
        <el-button type="primary" @click="doGrant">发放</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '@/api/client'
import { useReport } from '@/composables/useReport'

interface Coupon {
  id: number; name: string; createdAt: string; faceValue: number; minAmount: number
  totalQty: number; validStart: string | null; validEnd: string | null
  status: number; sort: number; remark: string; received: number; used: number
}

interface Grant {
  id: number; couponId: number; couponName: string; faceValue: number
  customerId: number; customer: string; status: string
  receivedAt: string; usedAt: string | null; refBillNo: string
}

interface CustomerOption { id: number; name: string }

const tab = ref('list')

const { list, total, loading, page, pageSize, filters, load, search, reset } =
  useReport<Coupon>('/api/v1/coupons', { name: '', status: '', startDate: '', endDate: '' })

const range = computed({
  get: () => (filters.value.startDate && filters.value.endDate ? [filters.value.startDate, filters.value.endDate] : null),
  set: (v) => {
    filters.value.startDate = v?.[0] ?? ''
    filters.value.endDate = v?.[1] ?? ''
  },
})

function formatTime(t: string | null) {
  return t ? new Date(t).toLocaleString('zh-CN', { hour12: false }) : '-'
}

function formatDate(t: string | null) {
  return t ? new Date(t).toLocaleDateString('zh-CN') : '-'
}

// ---- 领用统计 ----
const statsList = ref<Coupon[]>([])
const statsLoading = ref(false)
async function loadStats() {
  statsLoading.value = true
  try {
    const res = await api.get('/api/v1/coupons', { params: { page: 1, pageSize: 10000 } })
    if (res.data.code === 0 || res.data.code === 200) {
      statsList.value = (res.data.data?.list ?? []).filter((x: Coupon) => x.received > 0 || x.used > 0)
    }
  } finally {
    statsLoading.value = false
  }
}

// ---- 领取/使用详情 ----
const grantList = ref<Grant[]>([])
const grantTotal = ref(0)
const grantPage = ref(1)
const grantPageSize = ref(30)
const grantLoading = ref(false)
const grantFilters = ref({ keyword: '' })

async function loadGrants(p = grantPage.value) {
  grantPage.value = p
  grantLoading.value = true
  try {
    const res = await api.get('/api/v1/coupons/grants', {
      params: {
        page: grantPage.value, pageSize: grantPageSize.value,
        keyword: grantFilters.value.keyword,
        status: tab.value === 'used' ? 'used' : '',
      },
    })
    if (res.data.code === 0 || res.data.code === 200) {
      grantList.value = res.data.data?.list ?? []
      grantTotal.value = res.data.data?.total ?? 0
    }
  } finally {
    grantLoading.value = false
  }
}

watch(tab, (t) => {
  if (t === 'stats') loadStats()
  else if (t === 'received' || t === 'used') loadGrants(1)
})

// ---- 新增/编辑 ----
const dialogVisible = ref(false)
const form = ref({
  id: 0, name: '', faceValue: 0, minAmount: 0, totalQty: 0,
  validRange: null as [string, string] | null, enabled: true, sort: 0, remark: '',
})

function openDialog(row?: Coupon) {
  if (row) {
    form.value = {
      id: row.id, name: row.name, faceValue: row.faceValue, minAmount: row.minAmount,
      totalQty: row.totalQty,
      validRange: row.validStart && row.validEnd ? [row.validStart.slice(0, 10), row.validEnd.slice(0, 10)] : null,
      enabled: row.status === 1, sort: row.sort, remark: row.remark,
    }
  } else {
    form.value = { id: 0, name: '', faceValue: 0, minAmount: 0, totalQty: 0, validRange: null, enabled: true, sort: 0, remark: '' }
  }
  dialogVisible.value = true
}

async function save() {
  if (!form.value.name) { ElMessage.warning('请填写优惠券名称'); return }
  if (!form.value.faceValue) { ElMessage.warning('请填写面值'); return }
  const payload = {
    name: form.value.name,
    faceValue: form.value.faceValue,
    minAmount: form.value.minAmount,
    totalQty: form.value.totalQty,
    validStart: form.value.validRange?.[0] ? form.value.validRange[0] + 'T00:00:00Z' : null,
    validEnd: form.value.validRange?.[1] ? form.value.validRange[1] + 'T00:00:00Z' : null,
    status: form.value.enabled ? 1 : 0,
    sort: form.value.sort,
    remark: form.value.remark,
  }
  const res = form.value.id
    ? await api.put(`/api/v1/coupons/${form.value.id}`, payload)
    : await api.post('/api/v1/coupons', payload)
  if (res.data.code === 0 || res.data.code === 200) {
    ElMessage.success('保存成功')
    dialogVisible.value = false
    load()
  } else {
    ElMessage.error(res.data.message || '保存失败')
  }
}

async function remove(row: Coupon) {
  await ElMessageBox.confirm(`确定删除优惠券「${row.name}」？`, '提示', { type: 'warning' })
  const res = await api.delete(`/api/v1/coupons/${row.id}`)
  if (res.data.code === 0 || res.data.code === 200) {
    ElMessage.success('已删除')
    load()
  }
}

// ---- 发放 ----
const grantVisible = ref(false)
const grantCoupon = ref<Coupon | null>(null)
const grantCustomerIds = ref<number[]>([])
const customerOptions = ref<CustomerOption[]>([])

async function searchCustomers(kw: string) {
  const res = await api.get('/api/v1/customers', { params: { page: 1, pageSize: 50, keyword: kw } })
  if (res.data.code === 0 || res.data.code === 200) customerOptions.value = res.data.data?.list ?? []
}

function openGrant(row: Coupon) {
  grantCoupon.value = row
  grantCustomerIds.value = []
  searchCustomers('')
  grantVisible.value = true
}

async function doGrant() {
  if (!grantCustomerIds.value.length) { ElMessage.warning('请选择客户'); return }
  const res = await api.post(`/api/v1/coupons/${grantCoupon.value?.id}/grant`, { customerIds: grantCustomerIds.value })
  if (res.data.code === 0 || res.data.code === 200) {
    ElMessage.success(`已发放 ${res.data.data?.created ?? 0} 张`)
    grantVisible.value = false
    load()
  } else {
    ElMessage.error(res.data.message || '发放失败')
  }
}

async function markUsed(row: Grant) {
  await ElMessageBox.confirm(`确定将「${row.customer}」的该优惠券标记为已使用？`, '提示', { type: 'warning' })
  const res = await api.put(`/api/v1/coupons/grants/${row.id}/use`, {})
  if (res.data.code === 0 || res.data.code === 200) {
    ElMessage.success('已标记')
    loadGrants()
  }
}

onMounted(() => load(1))
</script>

<style scoped>
.page { display: flex; flex-direction: column; gap: 12px; }
.page-header { display: flex; justify-content: space-between; align-items: center; }
.page-title { font-size: 18px; font-weight: 600; margin: 0; }
.hint { font-size: 12px; color: #909399; line-height: 1.4; }
</style>
