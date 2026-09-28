<template>
  <div class="page">
    <el-card>
      <template #header>
        <div class="card-header"><span>库存查询</span></div>
      </template>
      <el-form :model="searchForm" inline class="search-form">
        <el-form-item label="仓库"><RemoteSelect v-model="searchForm.warehouseId" api-url="/api/v1/warehouses" placeholder="仓库" /></el-form-item>
        <el-form-item label="商品"><RemoteSelect v-model="searchForm.productId" api-url="/api/v1/products" placeholder="商品关键词搜索" /></el-form-item>
        <el-form-item><el-button type="primary" @click="handleSearch">搜索</el-button><el-button @click="handleReset">重置</el-button></el-form-item>
      </el-form>
      <el-table :data="list" v-loading="loading" border stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="productCode" label="商品编码" min-width="120"><template #default="{ row }">{{ row.productCode || '-' }}</template></el-table-column>
        <el-table-column prop="productName" label="商品" min-width="150"><template #default="{ row }">{{ row.productName || row.productId }}</template></el-table-column>
        <el-table-column prop="warehouseName" label="仓库" min-width="150"><template #default="{ row }">{{ row.warehouseName || row.warehouseId }}</template></el-table-column>
        <el-table-column prop="quantity" label="库存数量" width="120" />
      </el-table>
      <el-pagination v-model:current-page="pagination.page" v-model:page-size="pagination.pageSize" :total="total" :page-sizes="[10,20,50]" layout="total, sizes, prev, pager, next" @size-change="handleSizeChange" @current-change="handleCurrentChange" class="pagination" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { useCrud } from '@/composables/useCrud'
import RemoteSelect from '@/components/RemoteSelect.vue'

const crud = useCrud<any>({ baseUrl: '/api/v1/stocks', defaultForm: () => ({}) })
const { list, total, loading, searchForm, pagination, fetchList, handleSearch, handleReset, handleSizeChange, handleCurrentChange } = crud
fetchList()
</script>

<style scoped>
.page { padding: 20px; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
.search-form { margin-bottom: 16px; }
.pagination { margin-top: 16px; justify-content: flex-end; }
</style>
