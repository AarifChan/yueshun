<template>
  <div class="page">
    <el-card>
      <template #header>
        <div class="card-header"><span>审批记录</span></div>
      </template>
      <el-form :model="searchForm" inline class="search-form">
        <el-form-item label="流程"><RemoteSelect v-model="searchForm.processId" api-url="/api/v1/approval-processes" placeholder="流程" /></el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" placeholder="状态" clearable>
            <el-option label="待审批" value="pending" /><el-option label="已通过" value="approved" /><el-option label="已驳回" value="rejected" />
          </el-select>
        </el-form-item>
        <el-form-item><el-button type="primary" @click="handleSearch">搜索</el-button><el-button @click="handleReset">重置</el-button></el-form-item>
      </el-form>
      <el-table :data="list" v-loading="loading" border stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="processName" label="流程" min-width="150"><template #default="{ row }">{{ row.processName || row.processId }}</template></el-table-column>
        <el-table-column prop="applicantName" label="申请人" min-width="120"><template #default="{ row }">{{ row.applicantName || row.applicantId }}</template></el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'approved' ? 'success' : row.status === 'rejected' ? 'danger' : 'warning'">{{ row.status === 'approved' ? '已通过' : row.status === 'rejected' ? '已驳回' : '待审批' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="申请时间" width="160" />
      </el-table>
      <el-pagination v-model:current-page="pagination.page" v-model:page-size="pagination.pageSize" :total="total" :page-sizes="[10,20,50]" layout="total, sizes, prev, pager, next" @size-change="handleSizeChange" @current-change="handleCurrentChange" class="pagination" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { useCrud } from '@/composables/useCrud'
import RemoteSelect from '@/components/RemoteSelect.vue'

interface Record { id: number; processId: number; processName?: string; applicantId: number; applicantName?: string; status: string; createdAt: string }

const crud = useCrud<Record>({ baseUrl: '/api/v1/approval-records' })
const { list, total, loading, searchForm, pagination, fetchList, handleSearch, handleReset, handleSizeChange, handleCurrentChange } = crud
fetchList()
</script>

<style scoped>
.page { padding: 20px; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
.search-form { margin-bottom: 16px; }
.pagination { margin-top: 16px; justify-content: flex-end; }
</style>
