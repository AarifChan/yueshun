<template>
  <div class="page">
    <el-card>
      <template #header>
        <div class="card-header"><span>库存预警</span><el-button type="primary" @click="openCreate">新增预警</el-button></div>
      </template>
      <el-form :model="searchForm" inline class="search-form">
        <el-form-item label="仓库"><RemoteSelect v-model="searchForm.warehouseId" api-url="/api/v1/warehouses" placeholder="仓库" /></el-form-item>
        <el-form-item label="商品"><RemoteSelect v-model="searchForm.productId" api-url="/api/v1/products" placeholder="商品" /></el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" placeholder="状态" clearable>
            <el-option label="正常" value="normal" /><el-option label="预警" value="warning" />
          </el-select>
        </el-form-item>
        <el-form-item><el-button type="primary" @click="handleSearch">搜索</el-button><el-button @click="handleReset">重置</el-button></el-form-item>
      </el-form>
      <el-table :data="list" v-loading="loading" border stripe>
        <el-table-column prop="warehouseName" label="仓库" min-width="150"><template #default="{ row }">{{ row.warehouseName || row.warehouseId }}</template></el-table-column>
        <el-table-column prop="productName" label="商品" min-width="150"><template #default="{ row }">{{ row.productName || row.productId }}</template></el-table-column>
        <el-table-column prop="currentStock" label="当前库存" width="120" />
        <el-table-column prop="minStock" label="最低库存" width="120" />
        <el-table-column prop="maxStock" label="最高库存" width="120" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }"><el-tag :type="row.status === 'normal' ? 'success' : 'danger'">{{ row.status === 'normal' ? '正常' : '预警' }}</el-tag></template>
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
      <el-form :model="form" label-width="100px">
        <el-form-item label="仓库" required><RemoteSelect v-model="form.warehouseId" api-url="/api/v1/warehouses" /></el-form-item>
        <el-form-item label="商品" required><RemoteSelect v-model="form.productId" api-url="/api/v1/products" /></el-form-item>
        <el-form-item label="最低库存"><el-input-number v-model="form.minStock" :min="0" /></el-form-item>
        <el-form-item label="最高库存"><el-input-number v-model="form.maxStock" :min="0" /></el-form-item>
        <el-form-item label="备注"><el-input v-model="form.remark" type="textarea" /></el-form-item>
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

const crud = useCrud<any>({ baseUrl: '/api/v1/inventory-warnings', defaultForm: () => ({}) })
const { list, total, loading, dialogVisible, dialogTitle, form, searchForm, pagination, fetchList, openCreate, openEdit, handleSubmit, handleDelete, handleSearch, handleReset, handleSizeChange, handleCurrentChange } = crud
fetchList()
</script>

<style scoped>
.page { padding: 20px; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
.search-form { margin-bottom: 16px; }
.pagination { margin-top: 16px; justify-content: flex-end; }
</style>
