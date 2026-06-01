<template>
  <div class="page">
    <el-card>
      <template #header>
        <div class="card-header"><span>库存盘点</span><el-button type="primary" @click="goCreate">新增盘点</el-button></div>
      </template>
      <el-form :model="searchForm" inline class="search-form">
        <el-form-item label="单据编号"><el-input v-model="searchForm.billNo" placeholder="单据编号" clearable /></el-form-item>
        <el-form-item label="仓库"><RemoteSelect v-model="searchForm.warehouseId" api-url="/api/v1/warehouses" placeholder="仓库" /></el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" placeholder="状态" clearable>
            <el-option label="草稿" value="draft" /><el-option label="已完成" value="completed" />
          </el-select>
        </el-form-item>
        <el-form-item label="日期"><el-date-picker v-model="dateRange" type="daterange" value-format="YYYY-MM-DD" start-placeholder="开始" end-placeholder="结束" /></el-form-item>
        <el-form-item><el-button type="primary" @click="handleSearch">搜索</el-button><el-button @click="handleReset">重置</el-button></el-form-item>
      </el-form>
      <el-table :data="list" v-loading="loading" border stripe>
        <el-table-column prop="billNo" label="单据编号" min-width="150" />
        <el-table-column prop="billDate" label="盘点日期" width="120" />
        <el-table-column prop="warehouseName" label="仓库" min-width="150"><template #default="{ row }">{{ row.warehouseName || row.warehouseId }}</template></el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }"><el-tag :type="row.status === 'completed' ? 'success' : 'info'">{{ row.status === 'completed' ? '已完成' : '草稿' }}</el-tag></template>
        </el-table-column>
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="goDetail(row.id)">查看</el-button>
            <el-button v-if="row.status === 'draft'" type="warning" size="small" @click="goEdit(row.id)">编辑</el-button>
            <el-button v-if="row.status === 'draft'" type="success" size="small" @click="handleComplete(row)">完成</el-button>
            <el-button v-if="row.status === 'draft'" type="danger" size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-pagination v-model:current-page="pagination.page" v-model:page-size="pagination.pageSize" :total="total" :page-sizes="[10,20,50]" layout="total, sizes, prev, pager, next" @size-change="handleSizeChange" @current-change="handleCurrentChange" class="pagination" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useCrud } from '@/composables/useCrud'
import RemoteSelect from '@/components/RemoteSelect.vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '@/api/client'

const router = useRouter()
const dateRange = ref<[string, string] | null>(null)
const crud = useCrud<any>({ baseUrl: '/api/v1/inventory-checks', defaultForm: () => ({}) })
const { list, total, loading, searchForm, pagination, fetchList, handleSearch, handleReset, handleSizeChange, handleCurrentChange, handleDelete: crudDelete } = crud

watch(dateRange, (val) => {
  if (val) { searchForm.value.startDate = val[0]; searchForm.value.endDate = val[1] }
  else { searchForm.value.startDate = undefined; searchForm.value.endDate = undefined }
})

function goCreate() { router.push('/inventory-checks/new') }
function goDetail(id: number) { router.push(`/inventory-checks/${id}`) }
function goEdit(id: number) { router.push(`/inventory-checks/${id}?mode=edit`) }

async function handleComplete(row: any) {
  try {
    await ElMessageBox.confirm('完成该盘点单？', '提示', { type: 'warning' })
    const res = await api.put(`/api/v1/inventory-checks/${row.id}/complete`)
    if (res.data.code === 0 || res.data.code === 200) { ElMessage.success('完成成功'); fetchList() }
  } catch (e: any) { if (e !== 'cancel') ElMessage.error(e.message || '操作失败') }
}

async function handleDelete(row: any) { await crudDelete(row) }
fetchList()
</script>

<style scoped>
.page { padding: 20px; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
.search-form { margin-bottom: 16px; }
.pagination { margin-top: 16px; justify-content: flex-end; }
</style>
