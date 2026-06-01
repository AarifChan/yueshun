<template>
  <div class="page">
    <el-card>
      <template #header>
        <div class="card-header"><span>客户管理</span><el-button type="primary" @click="goCreate">新增客户</el-button></div>
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
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useCrud } from '@/composables/useCrud'

interface Customer { id: number; name: string; code: string; type: string; contact: string; phone: string; status: number }

const router = useRouter()
const crud = useCrud<Customer>({ baseUrl: '/api/v1/customers', defaultForm: () => ({ status: 1, type: 'customer' }) })
const { list, total, loading, searchForm, pagination, fetchList, handleDelete, handleSizeChange, handleCurrentChange, handleSearch, handleReset } = crud
fetchList()

function goCreate() { router.push('/customers/new') }
function goDetail(id: number) { router.push(`/customers/${id}`) }
function goEdit(id: number) { router.push(`/customers/${id}?mode=edit`) }
</script>

<style scoped>
.page { padding: 20px; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
.search-form { margin-bottom: 16px; }
.pagination { margin-top: 16px; justify-content: flex-end; }
</style>
