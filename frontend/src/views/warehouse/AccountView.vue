<template>
  <div class="page">
    <el-card>
      <template #header>
        <div class="card-header"><span>资金账户</span><el-button type="primary" @click="openCreate">新增账户</el-button></div>
      </template>
      <el-form :model="searchForm" inline class="search-form">
        <el-form-item label="名称"><el-input v-model="searchForm.name" placeholder="名称" clearable /></el-form-item>
        <el-form-item label="类型">
          <el-select v-model="searchForm.type" placeholder="类型" clearable>
            <el-option label="现金" value="cash" /><el-option label="银行" value="bank" /><el-option label="支付宝" value="alipay" /><el-option label="微信" value="wechat" />
          </el-select>
        </el-form-item>
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
        <el-table-column prop="type" label="类型" width="100" />
        <el-table-column prop="balance" label="余额" width="120"><template #default="{ row }">¥{{ row.balance?.toFixed(2) }}</template></el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }"><el-tag :type="row.status === 1 ? 'success' : 'danger'">{{ row.status === 1 ? '启用' : '禁用' }}</el-tag></template>
        </el-table-column>
        <el-table-column label="操作" width="240" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="openFlows(row)">流水</el-button>
            <el-button type="warning" size="small" @click="openEdit(row)">编辑</el-button>
            <el-button type="danger" size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-pagination v-model:current-page="pagination.page" v-model:page-size="pagination.pageSize" :total="total" :page-sizes="[10,20,50]" layout="total, sizes, prev, pager, next" @size-change="handleSizeChange" @current-change="handleCurrentChange" class="pagination" />
    </el-card>
    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="500px">
      <el-form :model="form" label-width="80px">
        <el-form-item label="名称" required><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="编码"><el-input v-model="form.code" /></el-form-item>
        <el-form-item label="类型" required>
          <el-select v-model="form.type">
            <el-option label="现金" value="cash" /><el-option label="银行" value="bank" /><el-option label="支付宝" value="alipay" /><el-option label="微信" value="wechat" />
          </el-select>
        </el-form-item>
        <el-form-item label="余额"><el-input-number v-model="form.balance" :precision="2" :min="0" /></el-form-item>
        <el-form-item label="状态"><el-switch v-model="form.status" :active-value="1" :inactive-value="0" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
    <el-dialog v-model="flowDialogVisible" :title="`账户流水 - ${flowAccount?.name || ''}`" width="800px">
      <el-table :data="flowList" v-loading="flowLoading" border stripe>
        <el-table-column prop="type" label="类型" width="100"><template #default="{ row }"><el-tag :type="isIncome(row.type) ? 'success' : 'danger'">{{ isIncome(row.type) ? '收入' : '支出' }}</el-tag></template></el-table-column>
        <el-table-column prop="amount" label="金额" width="120"><template #default="{ row }">¥{{ row.amount?.toFixed(2) }}</template></el-table-column>
        <el-table-column prop="refType" label="关联单据" width="150"><template #default="{ row }">{{ row.refType || '-' }}</template></el-table-column>
        <el-table-column prop="remark" label="备注" min-width="150"><template #default="{ row }">{{ row.remark || '-' }}</template></el-table-column>
        <el-table-column prop="createdAt" label="时间" width="180"><template #default="{ row }">{{ row.createdAt || '-' }}</template></el-table-column>
      </el-table>
      <el-pagination v-model:current-page="flowPagination.page" v-model:page-size="flowPagination.pageSize" :total="flowTotal" :page-sizes="[10,20,50]" layout="total, sizes, prev, pager, next" @size-change="fetchFlows" @current-change="fetchFlows" class="pagination" />
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import api from '@/api/client'
import { useCrud } from '@/composables/useCrud'

interface Account { id: number; name: string; code: string; type: string; balance: number; status: number }

const crud = useCrud<Account>({ baseUrl: '/api/v1/accounts', defaultForm: () => ({ status: 1, type: 'cash' }) })
const { list, total, loading, dialogVisible, dialogTitle, form, searchForm, pagination, fetchList, openCreate, openEdit, handleSubmit, handleDelete, handleSearch, handleReset, handleSizeChange, handleCurrentChange } = crud
fetchList()

const flowDialogVisible = ref(false)
const flowAccount = ref<Account | null>(null)
const flowList = ref<any[]>([])
const flowTotal = ref(0)
const flowLoading = ref(false)
const flowPagination = ref({ page: 1, pageSize: 20 })

function openFlows(row: Account) {
  flowAccount.value = row
  flowPagination.value = { page: 1, pageSize: 20 }
  flowDialogVisible.value = true
  fetchFlows()
}

function isIncome(type: string) {
  return type === 'income' || type === 'in'
}

async function fetchFlows() {
  if (!flowAccount.value) return
  flowLoading.value = true
  try {
    const res = await api.get(`/api/v1/accounts/${flowAccount.value.id}/flows`, {
      params: { page: flowPagination.value.page, pageSize: flowPagination.value.pageSize },
    })
    if (res.data.code === 0 || res.data.code === 200) {
      flowList.value = res.data.data.list || []
      flowTotal.value = res.data.data.total ?? 0
    }
  } catch (error: any) {
    ElMessage.error(error.message || '获取流水失败')
  } finally {
    flowLoading.value = false
  }
}
</script>

<style scoped>
.page { padding: 20px; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
.search-form { margin-bottom: 16px; }
.pagination { margin-top: 16px; justify-content: flex-end; }
</style>
