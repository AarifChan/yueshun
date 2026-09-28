<template>
  <div class="page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>客户管理</span>
          <div>
            <el-button type="warning" @click="openChurn">流失预警</el-button>
            <el-button type="primary" @click="goCreate">新增客户</el-button>
          </div>
        </div>
      </template>
      <el-form :model="searchForm" inline class="search-form">
        <el-form-item label="关键词"><el-input v-model="searchForm.keyword" placeholder="名称/编码/电话" clearable /></el-form-item>
        <el-form-item label="类型">
          <el-select v-model="searchForm.type" placeholder="类型" clearable>
            <el-option label="客户" value="customer" /><el-option label="供应商" value="supplier" /><el-option label="两者都是" value="both" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" placeholder="状态" clearable>
            <el-option label="启用" :value="1" /><el-option label="禁用" :value="0" />
          </el-select>
        </el-form-item>
        <el-form-item><el-button type="primary" @click="handleSearch">搜索</el-button><el-button @click="handleReset">重置</el-button></el-form-item>
      </el-form>
      <el-table :data="list" v-loading="loading" border stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="name" label="名称" min-width="150" />
        <el-table-column prop="code" label="编码" min-width="120" />
        <el-table-column prop="type" label="类型" width="100">
          <template #default="{ row }"><el-tag>{{ { customer: '客户', supplier: '供应商', both: '两者' }[row.type] || row.type }}</el-tag></template>
        </el-table-column>
        <el-table-column prop="contact" label="联系人" width="120" />
        <el-table-column prop="phone" label="电话" width="120" />
        <el-table-column prop="balance" label="欠款余额" width="120"><template #default="{ row }">{{ row.balance != null ? `¥${Number(row.balance).toFixed(2)}` : '-' }}</template></el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }"><el-tag :type="row.status === 1 ? 'success' : 'danger'">{{ row.status === 1 ? '启用' : '禁用' }}</el-tag></template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="goDetail(row.id)">查看</el-button>
            <el-button type="warning" size="small" @click="goEdit(row.id)">编辑</el-button>
            <el-button type="danger" size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-pagination v-model:current-page="pagination.page" v-model:page-size="pagination.pageSize" :total="total" :page-sizes="[10,20,50]" layout="total, sizes, prev, pager, next" @size-change="handleSizeChange" @current-change="handleCurrentChange" class="pagination" />
    </el-card>
    <el-dialog v-model="churnVisible" title="客户流失预警" width="820px">
      <div class="churn-toolbar">
        <span>未下单天数 ≥</span>
        <el-input-number v-model="churnDays" :min="1" :max="3650" />
        <el-button type="primary" :loading="churnLoading" @click="loadChurn">查询</el-button>
      </div>
      <el-table :data="churnList" v-loading="churnLoading" border stripe>
        <el-table-column prop="code" label="客户编码" width="120" />
        <el-table-column prop="name" label="名称" min-width="140" />
        <el-table-column prop="phone" label="电话" width="130" />
        <el-table-column label="最近下单时间" width="180">
          <template #default="{ row }">{{ formatDate(row.lastOrderAt) }}</template>
        </el-table-column>
        <el-table-column label="未下单天数" width="110">
          <template #default="{ row }">{{ row.daysSinceOrder ?? '从未下单' }}</template>
        </el-table-column>
      </el-table>
      <el-empty v-if="!churnLoading && !churnList.length" description="暂无流失风险客户" />
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useCrud } from '@/composables/useCrud'
import api from '@/api/client'

interface Customer { id: number; name: string; code: string; type: string; contact: string; phone: string; status: number }

const router = useRouter()
const crud = useCrud<Customer>({ baseUrl: '/api/v1/customers', defaultForm: () => ({ status: 1, type: 'customer' }) })
const { list, total, loading, searchForm, pagination, fetchList, handleDelete, handleSizeChange, handleCurrentChange, handleSearch, handleReset } = crud
fetchList()

function goCreate() { router.push('/customers/new') }
function goDetail(id: number) { router.push(`/customers/${id}`) }
function goEdit(id: number) { router.push(`/customers/${id}?mode=edit`) }

// 客户流失预警
const churnVisible = ref(false)
const churnDays = ref(90)
const churnLoading = ref(false)
const churnList = ref<Record<string, any>[]>([])

function openChurn() {
  churnVisible.value = true
  loadChurn()
}

async function loadChurn() {
  churnLoading.value = true
  try {
    const res = await api.get('/api/v1/customers/churn-risk', { params: { days: churnDays.value, pageSize: 999 } })
    if (res.data.code === 0 || res.data.code === 200) {
      churnList.value = res.data.data?.list || []
    }
  } catch {
    // 拦截器已提示错误
  } finally {
    churnLoading.value = false
  }
}

function formatDate(v: string | number | null | undefined) {
  if (!v) return '-'
  const d = new Date(typeof v === 'number' ? v : String(v).replace(' ', 'T'))
  return isNaN(d.getTime()) ? String(v) : d.toLocaleString()
}
</script>

<style scoped>
.page { padding: 20px; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
.search-form { margin-bottom: 16px; }
.pagination { margin-top: 16px; justify-content: flex-end; }
.churn-toolbar { display: flex; align-items: center; gap: 8px; margin-bottom: 12px; }
</style>
