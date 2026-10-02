<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">调拨申请单</h2>
      <el-button type="primary" :icon="Plus" @click="goCreate">新增调拨申请单</el-button>
    </div>
    <el-card>
      <div class="status-chips">
        <el-radio-group v-model="search.status" @change="load(1)">
          <el-radio-button value="">全部</el-radio-button>
          <el-radio-button value="draft">待审核</el-radio-button>
          <el-radio-button value="approved">待出库</el-radio-button>
          <el-radio-button value="shipped">待入库</el-radio-button>
          <el-radio-button value="completed">已完成</el-radio-button>
        </el-radio-group>
      </div>
      <el-form inline class="search-form">
        <el-form-item label="单号"><el-input v-model="search.keyword" placeholder="请输入调拨申请单号" clearable /></el-form-item>
        <el-form-item label="商品"><el-input v-model="search.productKw" placeholder="商品名称/编号" clearable /></el-form-item>
        <el-form-item label="调出仓库">
          <RemoteSelect v-model="search.fromWarehouseId" api-url="/api/v1/warehouses" placeholder="请选择" />
        </el-form-item>
        <el-form-item label="录单时间">
          <el-date-picker v-model="dateRange" type="daterange" value-format="YYYY-MM-DD" start-placeholder="开始时间" end-placeholder="结束时间" style="width: 260px" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="load(1)">查询</el-button>
          <el-button @click="reset">清空</el-button>
        </el-form-item>
      </el-form>
      <el-table :data="list" v-loading="loading" border stripe>
        <el-table-column label="操作" width="150" fixed="left">
          <template #default="{ row }">
            <el-button link type="primary" @click="goDetail(row)">查看</el-button>
            <el-button v-if="row.status === 'draft'" link type="success" @click="approve(row)">审核</el-button>
            <el-button v-if="row.status === 'draft'" link type="danger" @click="remove(row)">删除</el-button>
          </template>
        </el-table-column>
        <el-table-column prop="billNo" label="单号" width="160" />
        <el-table-column prop="createdAt" label="录单时间" width="170">
          <template #default="{ row }">{{ fmtTime(row.createdAt) }}</template>
        </el-table-column>
        <el-table-column prop="fromWarehouseName" label="调出仓库" min-width="120" />
        <el-table-column prop="toWarehouseName" label="调入仓库" min-width="120" />
        <el-table-column prop="totalQty" label="商品数量" width="100" align="right" />
        <el-table-column label="单据状态" width="100" align="center">
          <template #default="{ row }"><el-tag :type="statusType(row.status)">{{ statusLabel(row.status) }}</el-tag></template>
        </el-table-column>
        <el-table-column prop="handlerName" label="经手人" width="110" />
        <el-table-column prop="remark" label="备注" min-width="140" show-overflow-tooltip />
        <template #empty>暂无搜索结果</template>
      </el-table>
      <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" :page-sizes="[20,30,50]" layout="total, sizes, prev, pager, next" @size-change="() => load(1)" @current-change="load" class="pagination" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { Plus } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import RemoteSelect from '@/components/RemoteSelect.vue'
import api from '@/api/client'

interface Row { id: number; billNo: string; createdAt: string; fromWarehouseName: string; toWarehouseName: string; totalQty: number; status: string; handlerName: string; remark: string }

const router = useRouter()
const list = ref<Row[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(30)
const loading = ref(false)
const dateRange = ref<[string, string] | null>(null)
const search = reactive({ keyword: '', productKw: '', fromWarehouseId: undefined as number | undefined, status: '' })

const STATUS_MAP: Record<string, [string, string]> = {
  draft: ['待审核', 'info'], approved: ['待出库', 'warning'], shipped: ['待入库', 'warning'], completed: ['已完成', 'success'], cancelled: ['已取消', 'danger'],
}
function statusLabel(s: string) { return STATUS_MAP[s]?.[0] || s }
function statusType(s: string) { return (STATUS_MAP[s]?.[1] || 'info') as 'info' | 'warning' | 'success' | 'danger' }
function fmtTime(t: string) { return t ? t.replace('T', ' ').slice(0, 19) : '-' }

async function load(p = page.value) {
  page.value = p
  loading.value = true
  try {
    const res = await api.get('/api/v1/transfer-applies', {
      params: { page: page.value, pageSize: pageSize.value, ...search, startDate: dateRange.value?.[0] || '', endDate: dateRange.value?.[1] || '' },
    })
    if (res.data.code === 0 || res.data.code === 200) {
      list.value = res.data.data?.list ?? []
      total.value = res.data.data?.total ?? 0
    }
  } finally { loading.value = false }
}

function reset() {
  Object.assign(search, { keyword: '', productKw: '', fromWarehouseId: undefined, status: '' })
  dateRange.value = null
  load(1)
}

function goCreate() { router.push({ name: 'InventoryTransferApplyCreate' }) }
function goDetail(row: Row) { router.push({ name: 'InventoryTransferApplyDetail', params: { id: row.id } }) }

async function approve(row: Row) {
  await ElMessageBox.confirm(`确认审核调拨申请单 ${row.billNo}？`, '提示', { type: 'warning' })
  const res = await api.put(`/api/v1/transfer-applies/${row.id}/approve`)
  if (res.data.code === 0 || res.data.code === 200) {
    ElMessage.success('审核成功')
    load()
  }
}

async function remove(row: Row) {
  await ElMessageBox.confirm(`确认删除调拨申请单 ${row.billNo}？`, '提示', { type: 'warning' })
  const res = await api.delete(`/api/v1/transfer-applies/${row.id}`)
  if (res.data.code === 0 || res.data.code === 200) {
    ElMessage.success('删除成功')
    load()
  }
}

onMounted(() => load(1))
</script>

<style scoped>
.page { display: flex; flex-direction: column; gap: 12px; }
.page-header { display: flex; justify-content: space-between; align-items: center; }
.page-title { font-size: 18px; font-weight: 600; margin: 0; }
.status-chips { margin-bottom: 12px; }
.pagination { margin-top: 16px; justify-content: flex-end; }
</style>
