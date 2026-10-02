<template>
  <div class="page">
    <el-card>
      <template #header>
        <div class="card-header"><span>职员管理</span><el-button type="primary" @click="goCreate">新增职员</el-button></div>
      </template>
      <el-form :model="searchForm" inline class="search-form">
        <el-form-item label="关键词"><el-input v-model="searchForm.keyword" placeholder="用户名/姓名/手机号" clearable /></el-form-item>
        <el-form-item label="部门"><RemoteSelect v-model="searchForm.deptId" api-url="/api/v1/departments" placeholder="部门" /></el-form-item>
        <el-form-item label="角色"><RemoteSelect v-model="searchForm.roleId" api-url="/api/v1/roles" placeholder="角色" /></el-form-item>
        <el-form-item><el-button type="primary" @click="handleSearch">搜索</el-button><el-button @click="handleReset">重置</el-button></el-form-item>
      </el-form>
      <el-table :data="list" v-loading="loading" border stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="username" label="用户名" min-width="120" />
        <el-table-column prop="name" label="姓名" min-width="120" />
        <el-table-column prop="phone" label="手机号" width="120" />
        <el-table-column prop="deptName" label="部门" width="120"><template #default="{ row }">{{ row.deptName || row.deptId }}</template></el-table-column>
        <el-table-column prop="roleName" label="角色" width="120"><template #default="{ row }">{{ row.roleName || row.roleId }}</template></el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }"><el-tag :type="row.status === 1 ? 'success' : 'danger'">{{ row.status === 1 ? '启用' : '禁用' }}</el-tag></template>
        </el-table-column>
        <el-table-column label="企微绑定" width="150">
          <template #default="{ row }">
            <el-tag v-if="row.wecomUserId" type="success">{{ row.wecomUserId }}</el-tag>
            <el-tag v-else type="info">未绑定</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="260" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="goDetail(row.id)">查看</el-button>
            <el-button type="warning" size="small" @click="goEdit(row.id)">编辑</el-button>
            <el-button v-if="row.wecomUserId" type="warning" size="small" @click="handleUnbindWecom(row)">解绑企微</el-button>
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
import api from '@/api/client'
import { ElMessage, ElMessageBox } from 'element-plus'

interface Employee { id: number; username: string; name: string; phone: string; deptId?: number; deptName?: string; roleId?: number; roleName?: string; status: number; email: string; wecomUserId?: string }

const router = useRouter()
const crud = useCrud<Employee>({ baseUrl: '/api/v1/employees', defaultForm: () => ({ status: 1 }) })
const { list, total, loading, searchForm, pagination, fetchList, handleDelete, handleSearch, handleReset, handleSizeChange, handleCurrentChange } = crud
fetchList()

function goCreate() { router.push('/employees/new') }
function goDetail(id: number) { router.push(`/employees/${id}`) }
function goEdit(id: number) { router.push(`/employees/${id}?mode=edit`) }

async function handleUnbindWecom(row: Employee) {
  await ElMessageBox.confirm(`确认解除「${row.name}」的企业微信绑定？解除后其下次登录将重新按手机号绑定。`, '解绑确认', { type: 'warning' })
  await api.put(`/api/v1/employees/${row.id}`, { wecomUserId: '' })
  ElMessage.success('已解绑')
  fetchList()
}
</script>

<style scoped>
.page { padding: 20px; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
.search-form { margin-bottom: 16px; }
.pagination { margin-top: 16px; justify-content: flex-end; }
</style>
