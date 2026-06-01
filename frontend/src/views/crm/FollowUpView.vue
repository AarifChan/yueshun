<template>
  <div class="page">
    <el-card>
      <template #header>
        <div class="card-header"><span>跟进记录</span><el-button type="primary" @click="openCreate">新增跟进</el-button></div>
      </template>
      <el-form :model="searchForm" inline class="search-form">
        <el-form-item label="客户"><RemoteSelect v-model="searchForm.customerId" api-url="/api/v1/customers" placeholder="客户" /></el-form-item>
        <el-form-item label="类型"><el-input v-model="searchForm.type" placeholder="类型" clearable /></el-form-item>
        <el-form-item><el-button type="primary" @click="handleSearch">搜索</el-button><el-button @click="handleReset">重置</el-button></el-form-item>
      </el-form>
      <el-table :data="list" v-loading="loading" border stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="customerName" label="客户" min-width="150"><template #default="{ row }">{{ row.customerName || row.customerId }}</template></el-table-column>
        <el-table-column prop="type" label="类型" width="120" />
        <el-table-column prop="contactDate" label="联系日期" width="120" />
        <el-table-column prop="content" label="内容" min-width="200" show-overflow-tooltip />
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
        <el-form-item label="客户" required><RemoteSelect v-model="form.customerId" api-url="/api/v1/customers" /></el-form-item>
        <el-form-item label="类型"><el-input v-model="form.type" /></el-form-item>
        <el-form-item label="联系日期"><el-date-picker v-model="form.contactDate" type="date" value-format="YYYY-MM-DD" style="width:100%" /></el-form-item>
        <el-form-item label="内容"><el-input v-model="form.content" type="textarea" /></el-form-item>
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

interface FollowUp { id: number; customerId?: number; customerName?: string; type: string; contactDate: string; content: string }

const crud = useCrud<FollowUp>({ baseUrl: '/api/v1/follow-ups', defaultForm: () => ({}) })
const { list, total, loading, dialogVisible, dialogTitle, form, searchForm, pagination, fetchList, openCreate, openEdit, handleSubmit, handleDelete, handleSearch, handleReset, handleSizeChange, handleCurrentChange } = crud
fetchList()
</script>

<style scoped>
.page { padding: 20px; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
.search-form { margin-bottom: 16px; }
.pagination { margin-top: 16px; justify-content: flex-end; }
</style>
