<template>
  <div class="page">
    <el-card>
      <template #header>
        <div class="card-header"><span>发货方式</span><el-button type="primary" @click="openCreate">添加发货方式</el-button></div>
      </template>
      <el-form :model="searchForm" inline class="search-form">
        <el-form-item label="名称"><el-input v-model="searchForm.keyword" placeholder="名称" clearable /></el-form-item>
        <el-form-item><el-button type="primary" @click="handleSearch">搜索</el-button><el-button @click="handleReset">重置</el-button></el-form-item>
      </el-form>
      <el-table :data="list" v-loading="loading" border stripe>
        <el-table-column prop="name" label="发货方式" min-width="160" />
        <el-table-column label="物流" width="120" align="center">
          <template #default="{ row }"><el-tag :type="row.needLogistic ? 'primary' : 'info'">{{ row.needLogistic ? '需要物流' : '无需物流' }}</el-tag></template>
        </el-table-column>
        <el-table-column prop="sort" label="排序" width="80" align="center" />
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }"><el-tag :type="row.status === 1 ? 'success' : 'danger'">{{ row.status === 1 ? '启用' : '停用' }}</el-tag></template>
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
      <el-form :model="form" label-width="90px">
        <el-form-item label="发货方式" required><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="需要物流"><el-switch v-model="form.needLogistic" /></el-form-item>
        <el-form-item label="排序"><el-input-number v-model="form.sort" :min="0" /></el-form-item>
        <el-form-item label="状态"><el-switch v-model="form.status" :active-value="1" :inactive-value="0" /></el-form-item>
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

interface DeliveryMethod { id: number; name: string; needLogistic: boolean; sort: number; status: number }

const {
  list, total, loading, dialogVisible, dialogTitle, form, searchForm, pagination,
  openCreate, openEdit, handleSubmit, handleDelete,
  handleSizeChange, handleCurrentChange, handleSearch, handleReset,
} = useCrud<DeliveryMethod>({
  baseUrl: '/base-data/delivery-methods',
  defaultForm: () => ({ name: '', needLogistic: false, sort: 0, status: 1 }),
})
</script>

<style scoped>
.card-header { display: flex; justify-content: space-between; align-items: center; }
.pagination { margin-top: 16px; justify-content: flex-end; }
</style>
