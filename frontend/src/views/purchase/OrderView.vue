<template>
  <div class="page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>采购订单</span>
          <div>
            <el-button type="warning" @click="openReplenish">补货建议</el-button>
            <el-button type="primary" @click="goCreate">新增订单</el-button>
          </div>
        </div>
      </template>
      <el-form :model="searchForm" inline class="search-form">
        <el-form-item label="订单编号">
          <el-input v-model="searchForm.orderNo" placeholder="订单编号" clearable />
        </el-form-item>
        <el-form-item label="供应商">
          <RemoteSelect v-model="searchForm.supplierId" api-url="/api/v1/customers" :params="{ type: 'supplier' }" placeholder="供应商" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" placeholder="状态" clearable>
            <el-option label="草稿" value="draft" />
            <el-option label="已确认" value="confirmed" />
            <el-option label="已取消" value="cancelled" />
            <el-option label="已完成" value="completed" />
          </el-select>
        </el-form-item>
        <el-form-item label="日期">
          <el-date-picker v-model="dateRange" type="daterange" value-format="YYYY-MM-DD" start-placeholder="开始" end-placeholder="结束" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">搜索</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
      <el-table :data="list" v-loading="loading" border stripe>
        <el-table-column prop="orderNo" label="订单编号" min-width="150" />
        <el-table-column prop="orderDate" label="订单日期" width="120" />
        <el-table-column prop="supplierName" label="供应商" min-width="150">
          <template #default="{ row }">{{ row.supplierName || row.supplierId }}</template>
        </el-table-column>
        <el-table-column prop="totalAmount" label="总金额" width="120">
          <template #default="{ row }">{{ row.totalAmount?.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)">{{ statusLabel(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="goDetail(row.id)">查看</el-button>
            <el-button v-if="row.status === 'draft'" type="warning" size="small" @click="goEdit(row.id)">编辑</el-button>
            <template v-if="row.status === 'draft'">
              <el-button type="success" size="small" @click="handleConfirm(row)">确认</el-button>
              <el-button type="danger" size="small" @click="handleCancel(row)">取消</el-button>
            </template>
            <el-button v-if="row.status === 'draft'" type="danger" size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-pagination v-model:current-page="pagination.page" v-model:page-size="pagination.pageSize" :total="total" :page-sizes="[10,20,50]" layout="total, sizes, prev, pager, next" @size-change="handleSizeChange" @current-change="handleCurrentChange" class="pagination" />
    </el-card>
    <el-dialog v-model="replenishVisible" title="补货建议" width="820px">
      <el-table :data="replenishList" v-loading="replenishLoading" border stripe @selection-change="onReplenishSelection">
        <el-table-column type="selection" width="50" />
        <el-table-column prop="productCode" label="商品编码" width="120" />
        <el-table-column prop="productName" label="商品" min-width="160" />
        <el-table-column prop="unit" label="单位" width="80" />
        <el-table-column prop="currentStock" label="现存量" width="100" />
        <el-table-column prop="minStock" label="最低库存" width="100" />
        <el-table-column prop="suggestQty" label="建议采购量" width="120" />
      </el-table>
      <el-empty v-if="!replenishLoading && !replenishList.length" description="暂无补货建议" />
      <template #footer>
        <el-button @click="replenishVisible = false">取消</el-button>
        <el-button type="primary" :disabled="!replenishSelection.length" :loading="generating" @click="generateOrder">
          生成采购订单{{ replenishSelection.length ? `（已选 ${replenishSelection.length} 项）` : '' }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useCrud } from '@/composables/useCrud'
import RemoteSelect from '@/components/RemoteSelect.vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '@/api/client'

const router = useRouter()
const dateRange = ref<[string, string] | null>(null)

const crud = useCrud<any>({
  baseUrl: '/api/v1/purchase-orders',
  defaultForm: () => ({}),
})

const { list, total, loading, searchForm, pagination, fetchList, handleSearch, handleReset, handleSizeChange, handleCurrentChange, handleDelete: crudDelete } = crud

watch(dateRange, (val) => {
  if (val) {
    searchForm.value.startDate = val[0]
    searchForm.value.endDate = val[1]
  } else {
    searchForm.value.startDate = undefined
    searchForm.value.endDate = undefined
  }
})

function statusType(status: string) {
  return { draft: 'info', confirmed: 'success', cancelled: 'danger', completed: 'primary' }[status] || 'warning'
}
function statusLabel(status: string) {
  return { draft: '草稿', confirmed: '已确认', cancelled: '已取消', completed: '已完成' }[status] || status
}

function goCreate() { router.push('/purchase-orders/new') }
function goDetail(id: number) { router.push(`/purchase-orders/${id}`) }
function goEdit(id: number) { router.push(`/purchase-orders/${id}?mode=edit`) }

// 补货建议
const replenishVisible = ref(false)
const replenishLoading = ref(false)
const replenishList = ref<any[]>([])
const replenishSelection = ref<any[]>([])
const generating = ref(false)

function onReplenishSelection(rows: any[]) {
  replenishSelection.value = rows
}

async function openReplenish() {
  replenishVisible.value = true
  replenishLoading.value = true
  replenishSelection.value = []
  try {
    const res = await api.get('/api/v1/purchase-orders/replenish-suggestions', { params: { pageSize: 999 } })
    if (res.data.code === 0 || res.data.code === 200) {
      replenishList.value = res.data.data?.list || []
    }
  } catch {
    // 拦截器已提示错误
  } finally {
    replenishLoading.value = false
  }
}

async function generateOrder() {
  generating.value = true
  try {
    const payload = {
      items: replenishSelection.value.map(r => ({ productId: r.productId, quantity: r.suggestQty, price: 0 })),
    }
    const res = await api.post('/api/v1/purchase-orders', payload)
    if (res.data.code === 0 || res.data.code === 200) {
      ElMessage.success('采购订单已生成（草稿）')
      replenishVisible.value = false
      fetchList()
    }
  } catch {
    // 拦截器已提示错误
  } finally {
    generating.value = false
  }
}

async function handleConfirm(row: any) {
  try {
    await ElMessageBox.confirm('确认该采购订单？', '提示', { type: 'warning' })
    const res = await api.put(`/api/v1/purchase-orders/${row.id}/confirm`)
    if (res.data.code === 0 || res.data.code === 200) { ElMessage.success('确认成功'); fetchList() }
  } catch (e: any) { if (e !== 'cancel') ElMessage.error(e.message || '操作失败') }
}

async function handleCancel(row: any) {
  try {
    await ElMessageBox.confirm('取消该采购订单？', '提示', { type: 'warning' })
    const res = await api.put(`/api/v1/purchase-orders/${row.id}/cancel`)
    if (res.data.code === 0 || res.data.code === 200) { ElMessage.success('取消成功'); fetchList() }
  } catch (e: any) { if (e !== 'cancel') ElMessage.error(e.message || '操作失败') }
}

async function handleDelete(row: any) { await crudDelete(row) }

fetchList()
</script>

<style scoped>
.page { padding: 20px; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
.search-form { margin-bottom: 16px; }
.pagination { margin-top: 16px; justify-content: flex-end; }
</style>
