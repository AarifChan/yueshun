<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">组装拆装单</h2>
      <el-button type="primary" :icon="Plus" @click="goCreate">新增组装拆装单</el-button>
    </div>
    <el-card>
      <el-form inline class="search-form">
        <el-form-item label="单号"><el-input v-model="search.keyword" placeholder="请输入组装拆装单号" clearable /></el-form-item>
        <el-form-item label="出库仓库">
          <RemoteSelect v-model="search.outWarehouseId" api-url="/api/v1/warehouses" placeholder="请选择" />
        </el-form-item>
        <el-form-item label="入库仓库">
          <RemoteSelect v-model="search.inWarehouseId" api-url="/api/v1/warehouses" placeholder="请选择" />
        </el-form-item>
        <el-form-item label="单据状态">
          <el-select v-model="search.status" placeholder="请选择" clearable style="width: 130px">
            <el-option label="草稿" value="draft" /><el-option label="已过账" value="completed" />
          </el-select>
        </el-form-item>
        <el-form-item><el-checkbox v-model="search.showCancelled">显示撤销单据</el-checkbox></el-form-item>
        <el-form-item>
          <el-button type="primary" @click="load(1)">查询</el-button>
          <el-button @click="reset">清空</el-button>
        </el-form-item>
      </el-form>
      <el-table :data="list" v-loading="loading" border stripe>
        <el-table-column label="操作" width="130" fixed="left">
          <template #default="{ row }">
            <el-button link type="primary" @click="goDetail(row)">查看</el-button>
            <el-button v-if="row.status === 'draft'" link type="danger" @click="remove(row)">删除</el-button>
          </template>
        </el-table-column>
        <el-table-column prop="billNo" label="单号" width="160" />
        <el-table-column prop="createdAt" label="录单时间" width="170">
          <template #default="{ row }">{{ fmtTime(row.createdAt) }}</template>
        </el-table-column>
        <el-table-column prop="outWarehouseName" label="出库" min-width="110" />
        <el-table-column prop="inWarehouseName" label="入库" min-width="110" />
        <el-table-column label="加工费用" width="110" align="right">
          <template #default="{ row }">{{ row.fee?.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="单据状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 'completed' ? 'success' : 'info'">{{ row.status === 'completed' ? '已过账' : '草稿' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="handlerName" label="经手人" width="100" />
        <el-table-column prop="remark" label="备注" min-width="130" show-overflow-tooltip />
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

interface Row {
  id: number; billNo: string; createdAt: string; outWarehouseName: string; inWarehouseName: string
  fee: number; status: string; handlerName: string; remark: string
}

const router = useRouter()
const list = ref<Row[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(30)
const loading = ref(false)
const search = reactive({
  keyword: '', outWarehouseId: undefined as number | undefined,
  inWarehouseId: undefined as number | undefined, status: '', showCancelled: false,
})

function fmtTime(t: string) { return t ? t.replace('T', ' ').slice(0, 19) : '-' }

async function load(p = page.value) {
  page.value = p
  loading.value = true
  try {
    const res = await api.get('/api/v1/assembly-orders', { params: { page: page.value, pageSize: pageSize.value, ...search } })
    if (res.data.code === 0 || res.data.code === 200) {
      list.value = res.data.data?.list ?? []
      total.value = res.data.data?.total ?? 0
    }
  } finally { loading.value = false }
}

function reset() {
  Object.assign(search, { keyword: '', outWarehouseId: undefined, inWarehouseId: undefined, status: '', showCancelled: false })
  load(1)
}

function goCreate() { router.push({ name: 'InventoryAssemblyOrderCreate' }) }
function goDetail(row: Row) { router.push({ name: 'InventoryAssemblyOrderDetail', params: { id: row.id } }) }

async function remove(row: Row) {
  await ElMessageBox.confirm(`确认删除组装拆装单 ${row.billNo}？`, '提示', { type: 'warning' })
  const res = await api.delete(`/api/v1/assembly-orders/${row.id}`)
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
.pagination { margin-top: 16px; justify-content: flex-end; }
</style>
