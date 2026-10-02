<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">{{ title }}</h2>
      <div class="actions">
        <el-button @click="exportCsv(exportName, exportCols)">导出</el-button>
        <el-button type="primary" @click="openDialog()">{{ createLabel }}</el-button>
      </div>
    </div>
    <el-card>
      <el-form inline @submit.prevent>
        <el-form-item label="活动时间">
          <el-date-picker v-model="range" type="daterange" value-format="YYYY-MM-DD" start-placeholder="开始时间" end-placeholder="结束时间" style="width: 260px" />
        </el-form-item>
        <el-form-item label="活动名称">
          <el-input v-model="filters.name" placeholder="活动名称" clearable style="width: 160px" />
        </el-form-item>
        <el-form-item label="活动状态">
          <el-select v-model="filters.status" style="width: 120px">
            <el-option label="全部" value="all" />
            <el-option label="进行中" value="active" />
            <el-option label="未开始" value="pending" />
            <el-option label="已结束" value="finished" />
            <el-option label="已停用" value="disabled" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="!wholeOrder" label="商品">
          <el-input v-model="filters.productKeyword" placeholder="商品名称/编号" clearable style="width: 160px" />
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
        <el-table-column prop="name" label="活动名称" min-width="150" show-overflow-tooltip />
        <el-table-column label="活动状态" width="90" align="center">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)">{{ row.statusLabel }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="活动时间" width="200">
          <template #default="{ row }">{{ formatDate(row.startAt) }} ~ {{ formatDate(row.endAt) }}</template>
        </el-table-column>
        <el-table-column prop="remainDays" label="剩余天数" width="90" align="right" />
        <el-table-column v-if="showProductCount" prop="productCount" label="活动商品数量" width="110" align="right">
          <template #default="{ row }">{{ row.productCount || '全部' }}</template>
        </el-table-column>
        <el-table-column v-if="showGift" prop="giftQty" label="活动赠品数" width="100" align="right" />
        <el-table-column prop="customers" label="购买客户数" width="100" align="right" sortable />
        <el-table-column v-if="wholeOrder" prop="orderBills" label="单量" width="80" align="right" sortable />
        <el-table-column v-if="!wholeOrder" prop="orderQty" label="下单数量" width="90" align="right" sortable />
        <el-table-column v-if="!wholeOrder" prop="outQty" label="出库数量" width="90" align="right" />
        <el-table-column v-if="!wholeOrder" prop="returnQty" label="退货数量" width="90" align="right" />
        <el-table-column v-if="!wholeOrder" prop="actualQty" label="实销数量" width="90" align="right" />
        <el-table-column prop="orderAmount" label="下单金额" width="110" align="right" sortable>
          <template #default="{ row }">{{ row.orderAmount?.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column prop="outAmount" label="出库金额" width="110" align="right">
          <template #default="{ row }">{{ row.outAmount?.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column prop="returnAmount" label="退货金额" width="110" align="right">
          <template #default="{ row }">{{ row.returnAmount?.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column prop="actualAmount" label="实销金额" width="110" align="right">
          <template #default="{ row }">{{ row.actualAmount?.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column prop="sort" label="排序" width="70" align="right" />
        <el-table-column label="操作" width="130" fixed="right">
          <template #default="{ row }">
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
    </el-card>

    <el-dialog v-model="dialogVisible" :title="form.id ? '编辑活动' : createLabel" width="560px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="活动名称" required><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="活动时间" required>
          <el-date-picker v-model="form.range" type="datetimerange" value-format="YYYY-MM-DD HH:mm:ss" start-placeholder="开始时间" end-placeholder="结束时间" style="width: 100%" />
        </el-form-item>
        <el-form-item v-if="!wholeOrder" label="参与商品">
          <el-select v-model="form.productIdList" multiple filterable remote :remote-method="searchProducts" placeholder="不选为全部商品" style="width: 100%">
            <el-option v-for="p in productOptions" :key="p.id" :label="`${p.name}（${p.code || '无编号'}）`" :value="p.id" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="showGift" label="活动赠品数">
          <el-input-number v-model="form.giftQty" :min="0" style="width: 100%" />
        </el-form-item>
        <el-form-item label="是否启用"><el-switch v-model="form.enabled" /></el-form-item>
        <el-form-item label="排序"><el-input-number v-model="form.sort" :min="0" /></el-form-item>
        <el-form-item label="备注"><el-input v-model="form.remark" type="textarea" :rows="2" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="save">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '@/api/client'
import { useReport } from '@/composables/useReport'

const props = withDefaults(defineProps<{
  promoType: string
  title: string
  createLabel: string
  showGift?: boolean
  showProductCount?: boolean
  wholeOrder?: boolean
  exportName: string
}>(), { showGift: false, showProductCount: false, wholeOrder: false })

interface Row {
  id: number; name: string; createdAt: string; startAt: string; endAt: string
  status: string; statusLabel: string; remainDays: number
  productIds: string; productCount: number; giftQty: number; sort: number; remark: string
  customers: number; orderBills: number
  orderQty: number; orderAmount: number; outQty: number; outAmount: number
  returnQty: number; returnAmount: number; actualQty: number; actualAmount: number
}

interface ProductOption { id: number; name: string; code?: string }

const { list, total, loading, page, pageSize, filters, load, search, reset, exportCsv } =
  useReport<Row>('/api/v1/promotions', {
    type: props.promoType, name: '', status: 'all', startDate: '', endDate: '', productKeyword: '',
  })

const range = computed({
  get: () => (filters.value.startDate && filters.value.endDate ? [filters.value.startDate, filters.value.endDate] : null),
  set: (v) => {
    filters.value.startDate = v?.[0] ?? ''
    filters.value.endDate = v?.[1] ?? ''
  },
})

const exportCols = [
  { key: 'name', label: '活动名称' },
  { key: 'statusLabel', label: '活动状态' },
  { key: 'remainDays', label: '剩余天数' },
  { key: 'customers', label: '购买客户数' },
  { key: 'orderQty', label: '下单数量' },
  { key: 'outQty', label: '出库数量' },
  { key: 'returnQty', label: '退货数量' },
  { key: 'actualQty', label: '实销数量' },
  { key: 'orderAmount', label: '下单金额' },
  { key: 'outAmount', label: '出库金额' },
  { key: 'returnAmount', label: '退货金额' },
  { key: 'actualAmount', label: '实销金额' },
]

function statusTagType(s: string) {
  return s === 'active' ? 'success' : s === 'pending' ? 'warning' : s === 'finished' ? 'info' : 'danger'
}

function formatTime(t: string) {
  return t ? new Date(t).toLocaleString('zh-CN', { hour12: false }) : ''
}

function formatDate(t: string) {
  return t ? new Date(t).toLocaleDateString('zh-CN') : ''
}

const dialogVisible = ref(false)
const form = ref({
  id: 0, name: '', range: null as [string, string] | null,
  productIdList: [] as number[], giftQty: 0, enabled: true, sort: 0, remark: '',
})
const productOptions = ref<ProductOption[]>([])

async function searchProducts(kw: string) {
  const res = await api.get('/api/v1/products', { params: { page: 1, pageSize: 50, keyword: kw } })
  if (res.data.code === 0 || res.data.code === 200) productOptions.value = res.data.data?.list ?? []
}

function openDialog(row?: Row) {
  if (row) {
    form.value = {
      id: row.id, name: row.name,
      range: [row.startAt.replace('T', ' ').slice(0, 19), row.endAt.replace('T', ' ').slice(0, 19)],
      productIdList: row.productIds ? row.productIds.split(',').filter(Boolean).map(Number) : [],
      giftQty: row.giftQty, enabled: row.status !== 'disabled', sort: row.sort, remark: row.remark,
    }
    if (form.value.productIdList.length) searchProducts('')
  } else {
    form.value = { id: 0, name: '', range: null, productIdList: [], giftQty: 0, enabled: true, sort: 0, remark: '' }
  }
  dialogVisible.value = true
}

async function save() {
  if (!form.value.name) { ElMessage.warning('请填写活动名称'); return }
  if (!form.value.range?.[0] || !form.value.range?.[1]) { ElMessage.warning('请选择活动时间'); return }
  const payload = {
    type: props.promoType,
    name: form.value.name,
    startAt: form.value.range[0],
    endAt: form.value.range[1],
    productIds: form.value.productIdList.join(','),
    giftQty: form.value.giftQty,
    sort: form.value.sort,
    status: form.value.enabled ? 1 : 0,
    remark: form.value.remark,
  }
  const res = form.value.id
    ? await api.put(`/api/v1/promotions/${form.value.id}`, payload)
    : await api.post('/api/v1/promotions', payload)
  if (res.data.code === 0 || res.data.code === 200) {
    ElMessage.success('保存成功')
    dialogVisible.value = false
    load()
  } else {
    ElMessage.error(res.data.message || '保存失败')
  }
}

async function remove(row: Row) {
  await ElMessageBox.confirm(`确定删除活动「${row.name}」？`, '提示', { type: 'warning' })
  const res = await api.delete(`/api/v1/promotions/${row.id}`)
  if (res.data.code === 0 || res.data.code === 200) {
    ElMessage.success('已删除')
    load()
  }
}

onMounted(() => load(1))
</script>

<style scoped>
.page { display: flex; flex-direction: column; gap: 12px; }
.page-header { display: flex; justify-content: space-between; align-items: center; }
.page-title { font-size: 18px; font-weight: 600; margin: 0; }
.actions { display: flex; gap: 8px; }
</style>
