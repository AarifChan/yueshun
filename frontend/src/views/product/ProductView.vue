<template>
  <div class="page">
    <el-card>
      <template #header>
        <div class="card-header"><span>商品管理</span><el-button type="primary" @click="goCreate">新增商品</el-button></div>
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
        <el-table-column prop="totalStock" label="库存" width="100"><template #default="{ row }">{{ row.totalStock ?? '-' }}</template></el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }"><el-tag :type="row.status === 1 ? 'success' : 'danger'">{{ row.status === 1 ? '启用' : '禁用' }}</el-tag></template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
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
import RemoteSelect from '@/components/RemoteSelect.vue'

interface Product { id: number; name: string; code: string; barcode: string; categoryName: string; brandName: string; unit: string; retailPrice: number; status: number }

const router = useRouter()
const crud = useCrud<Product>({ baseUrl: '/api/v1/products', defaultForm: () => ({ status: 1 }) })
const { list, total, loading, searchForm, pagination, fetchList, handleDelete, handleSizeChange, handleCurrentChange, handleSearch, handleReset } = crud
fetchList()

function goCreate() { router.push('/products/new') }
function goDetail(id: number) { router.push(`/products/${id}`) }
function goEdit(id: number) { router.push(`/products/${id}?mode=edit`) }
</script>

<style scoped>
.page { padding: 20px; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
.search-form { margin-bottom: 16px; }
.pagination { margin-top: 16px; justify-content: flex-end; }
</style>
