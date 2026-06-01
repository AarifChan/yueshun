<template>
  <div class="page">
    <el-card>
      <template #header>
        <div class="card-header"><span>商品管理</span><el-button type="primary" @click="openCreate">新增商品</el-button></div>
      </template>
      <el-form :model="searchForm" inline class="search-form">
        <el-form-item label="关键词"><el-input v-model="searchForm.keyword" placeholder="名称/编码/条码" clearable /></el-form-item>
        <el-form-item label="分类"><RemoteSelect v-model="searchForm.categoryId" api-url="/api/v1/products/categories" placeholder="分类" /></el-form-item>
        <el-form-item label="品牌"><RemoteSelect v-model="searchForm.brandId" api-url="/api/v1/products/brands" placeholder="品牌" /></el-form-item>
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
        <el-table-column prop="barcode" label="条码" width="120" />
        <el-table-column prop="categoryName" label="分类" width="120" />
        <el-table-column prop="brandName" label="品牌" width="120" />
        <el-table-column prop="unit" label="单位" width="80" />
        <el-table-column prop="retailPrice" label="零售价" width="100" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }"><el-tag :type="row.status === 1 ? 'success' : 'danger'">{{ row.status === 1 ? '启用' : '禁用' }}</el-tag></template>
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
    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="600px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="分类" required><RemoteSelect v-model="form.categoryId" api-url="/api/v1/products/categories" /></el-form-item>
        <el-form-item label="品牌"><RemoteSelect v-model="form.brandId" api-url="/api/v1/products/brands" /></el-form-item>
        <el-form-item label="名称" required><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="编码"><el-input v-model="form.code" /></el-form-item>
        <el-form-item label="条码"><el-input v-model="form.barcode" /></el-form-item>
        <el-form-item label="规格"><el-input v-model="form.specification" /></el-form-item>
        <el-form-item label="单位" required><el-input v-model="form.unit" /></el-form-item>
        <el-form-item label="采购价"><el-input-number v-model="form.purchasePrice" :precision="2" /></el-form-item>
        <el-form-item label="零售价"><el-input-number v-model="form.retailPrice" :precision="2" /></el-form-item>
        <el-form-item label="批发价"><el-input-number v-model="form.wholesalePrice" :precision="2" /></el-form-item>
        <el-form-item label="最低库存"><el-input-number v-model="form.minStock" /></el-form-item>
        <el-form-item label="最高库存"><el-input-number v-model="form.maxStock" /></el-form-item>
        <el-form-item label="描述"><el-input v-model="form.description" type="textarea" /></el-form-item>
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
import RemoteSelect from '@/components/RemoteSelect.vue'

interface Product { id: number; name: string; code: string; barcode: string; categoryName: string; brandName: string; unit: string; retailPrice: number; status: number }

const crud = useCrud<Product>({ baseUrl: '/api/v1/products', defaultForm: () => ({ status: 1 }) })
const { list, total, loading, dialogVisible, dialogTitle, form, searchForm, pagination, fetchList, openCreate, openEdit, handleSubmit, handleDelete, handleSizeChange, handleCurrentChange, handleSearch, handleReset } = crud
fetchList()
</script>

<style scoped>
.page { padding: 20px; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
.search-form { margin-bottom: 16px; }
.pagination { margin-top: 16px; justify-content: flex-end; }
</style>
