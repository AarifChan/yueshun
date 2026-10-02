<template>
  <div class="page">
    <el-card>
      <template #header>
        <div class="card-header"><span>出入库类型</span><el-button type="primary" @click="openCreate">新增出入库类型</el-button></div>
      </template>
      <el-form :model="searchForm" inline class="search-form">
        <el-form-item label="名称/编号"><el-input v-model="searchForm.keyword" placeholder="名称/编号" clearable /></el-form-item>
        <el-form-item label="类型属性">
          <el-select v-model="searchForm.type" placeholder="全部" clearable style="width: 120px">
            <el-option label="入库" value="in" /><el-option label="出库" value="out" />
          </el-select>
        </el-form-item>
        <el-form-item><el-button type="primary" @click="handleSearch">搜索</el-button><el-button @click="handleReset">重置</el-button></el-form-item>
      </el-form>
      <el-table :data="list" v-loading="loading" border stripe>
        <el-table-column prop="code" label="类型编号" width="120" />
        <el-table-column prop="name" label="类型名称" min-width="160" />
        <el-table-column label="类型属性" width="110" align="center">
          <template #default="{ row }"><el-tag :type="row.type === 'in' ? 'success' : 'warning'">{{ row.type === 'in' ? '入库' : '出库' }}</el-tag></template>
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
        <el-form-item label="类型名称" required><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="类型编号"><el-input v-model="form.code" /></el-form-item>
        <el-form-item label="类型属性" required>
          <el-radio-group v-model="form.type"><el-radio value="in">入库</el-radio><el-radio value="out">出库</el-radio></el-radio-group>
        </el-form-item>
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

interface StockOpType { id: number; name: string; code: string; type: string; sort: number; status: number }

const {
  list, total, loading, dialogVisible, dialogTitle, form, searchForm, pagination,
  openCreate, openEdit, handleSubmit, handleDelete,
  handleSizeChange, handleCurrentChange, handleSearch, handleReset,
} = useCrud<StockOpType>({
  baseUrl: '/base-data/stock-op-types',
  defaultForm: () => ({ name: '', code: '', type: 'in', sort: 0, status: 1 }),
})
</script>

<style scoped>
.card-header { display: flex; justify-content: space-between; align-items: center; }
.pagination { margin-top: 16px; justify-content: flex-end; }
</style>
