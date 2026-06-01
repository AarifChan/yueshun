<template>
  <div class="page">
    <el-card>
      <template #header>
        <div class="card-header"><span>合同管理</span><el-button type="primary" @click="openCreate">新增合同</el-button></div>
      </template>
      <el-form :model="searchForm" inline class="search-form">
        <el-form-item label="名称"><el-input v-model="searchForm.name" placeholder="名称" clearable /></el-form-item>
        <el-form-item label="客户"><RemoteSelect v-model="searchForm.customerId" api-url="/api/v1/customers" placeholder="客户" /></el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" placeholder="状态" clearable>
            <el-option label="草稿" value="draft" /><el-option label="生效" value="active" /><el-option label="终止" value="terminated" />
          </el-select>
        </el-form-item>
        <el-form-item><el-button type="primary" @click="handleSearch">搜索</el-button><el-button @click="handleReset">重置</el-button></el-form-item>
      </el-form>
      <el-table :data="list" v-loading="loading" border stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="name" label="名称" min-width="150" />
        <el-table-column prop="customerName" label="客户" min-width="150"><template #default="{ row }">{{ row.customerName || row.customerId }}</template></el-table-column>
        <el-table-column prop="amount" label="金额" width="120"><template #default="{ row }">¥{{ row.amount?.toFixed(2) }}</template></el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }"><el-tag :type="row.status === 'active' ? 'success' : row.status === 'terminated' ? 'danger' : 'info'">{{ row.status === 'active' ? '生效' : row.status === 'terminated' ? '终止' : '草稿' }}</el-tag></template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="openEdit(row)">编辑</el-button>
            <el-button type="danger" size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-pagination v-model:current-page="pagination.page" v-model:page-size="pagination.pageSize" :total="total" :page-sizes="[10,20,50]" layout="total, sizes, prev, pager, next" @size-change="handleSizeChange" @current-change="handleCurrentChange" class="pagination" />
    </el-card>
    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="500px">
      <el-form :model="form" label-width="80px">
        <el-form-item label="名称" required><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="客户"><RemoteSelect v-model="form.customerId" api-url="/api/v1/customers" /></el-form-item>
        <el-form-item label="金额"><el-input-number v-model="form.amount" :precision="2" :min="0" /></el-form-item>
        <el-form-item label="状态">
          <el-select v-model="form.status"><el-option label="草稿" value="draft" /><el-option label="生效" value="active" /><el-option label="终止" value="terminated" /></el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { useCrud } from '@/composables/useCrud'
import RemoteSelect from '@/components/RemoteSelect.vue'

interface Contract { id: number; name: string; customerId?: number; customerName?: string; amount: number; status: string }

const crud = useCrud<Contract>({ baseUrl: '/api/v1/contracts', defaultForm: () => ({ status: 'draft' }) })
const { list, total, loading, dialogVisible, dialogTitle, form, searchForm, pagination, fetchList, openCreate, openEdit, handleSubmit, handleDelete, handleSearch, handleReset, handleSizeChange, handleCurrentChange } = crud
fetchList()
</script>

<style scoped>
.page { padding: 20px; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
.search-form { margin-bottom: 16px; }
.pagination { margin-top: 16px; justify-content: flex-end; }
</style>
