<template>
  <div class="page">
    <el-card>
      <template #header>
        <div class="card-header"><span>商城订单</span></div>
      </template>
      <el-form :model="searchForm" inline class="search-form">
        <el-form-item label="订单编号"><el-input v-model="searchForm.orderNo" placeholder="订单编号" clearable /></el-form-item>
        <el-form-item label="客户"><RemoteSelect v-model="searchForm.customerId" api-url="/api/v1/customers" placeholder="客户" /></el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" placeholder="状态" clearable>
            <el-option label="待付款" value="pending" /><el-option label="已付款" value="paid" /><el-option label="已发货" value="shipped" /><el-option label="已完成" value="completed" />
          </el-select>
        </el-form-item>
        <el-form-item><el-button type="primary" @click="handleSearch">搜索</el-button><el-button @click="handleReset">重置</el-button></el-form-item>
      </el-form>
      <el-table :data="list" v-loading="loading" border stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="orderNo" label="订单编号" min-width="150" />
        <el-table-column prop="customerName" label="客户" min-width="120"><template #default="{ row }">{{ row.customerName || row.customerId }}</template></el-table-column>
        <el-table-column prop="totalAmount" label="总金额" width="120"><template #default="{ row }">¥{{ row.totalAmount?.toFixed(2) }}</template></el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'completed' ? 'success' : row.status === 'shipped' ? 'primary' : row.status === 'paid' ? 'warning' : 'info'">
              {{ row.status === 'completed' ? '已完成' : row.status === 'shipped' ? '已发货' : row.status === 'paid' ? '已付款' : '待付款' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="goDetail(row.id)">查看</el-button>
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

interface MallOrder { id: number; orderNo: string; customerId?: number; customerName?: string; totalAmount: number; status: string }

const router = useRouter()
const crud = useCrud<MallOrder>({ baseUrl: '/api/v1/mall-orders' })
const { list, total, loading, searchForm, pagination, fetchList, handleSearch, handleReset, handleSizeChange, handleCurrentChange } = crud
fetchList()

function goDetail(id: number) { router.push(`/mall-orders/${id}`) }
</script>

<style scoped>
.page { padding: 20px; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
.search-form { margin-bottom: 16px; }
.pagination { margin-top: 16px; justify-content: flex-end; }
</style>
