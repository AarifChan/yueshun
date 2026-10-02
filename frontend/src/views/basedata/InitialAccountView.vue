<template>
  <div class="page">
    <el-card>
      <template #header>
        <div class="card-header"><span>现金银行期初</span><el-button type="primary" @click="openCreate">新增期初</el-button></div>
      </template>
      <el-form :model="searchForm" inline class="search-form">
        <el-form-item label="账户"><el-input v-model="searchForm.keyword" placeholder="账户名称" clearable /></el-form-item>
        <el-form-item><el-button type="primary" @click="handleSearch">搜索</el-button><el-button @click="handleReset">重置</el-button></el-form-item>
      </el-form>
      <el-table :data="list" v-loading="loading" border stripe>
        <el-table-column prop="accountName" label="账户名称" min-width="180" />
        <el-table-column label="账户类型" width="120" align="center">
          <template #default="{ row }"><el-tag>{{ typeLabel(row.accountType) }}</el-tag></template>
        </el-table-column>
        <el-table-column label="期初金额" width="140" align="right">
          <template #default="{ row }">{{ row.amount?.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="openEdit(row)">编辑</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-pagination v-model:current-page="pagination.page" v-model:page-size="pagination.pageSize" :total="total" :page-sizes="[10,20,50]" layout="total, sizes, prev, pager, next" @size-change="handleSizeChange" @current-change="handleCurrentChange" class="pagination" />
    </el-card>
    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="500px">
      <el-form :model="form" label-width="90px">
        <el-form-item label="账户" required>
          <el-select v-model="form.accountId" placeholder="选择账户" style="width: 100%" :disabled="isEdit">
            <el-option v-for="a in accountOptions" :key="a.id" :label="a.name" :value="a.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="期初金额"><el-input-number v-model="form.amount" :precision="2" style="width: 100%" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import api from '@/api/client'
import { useCrud } from '@/composables/useCrud'

interface InitialAccount { id: number; accountId: number; amount: number; accountName?: string; accountType?: string }
interface Option { id: number; name: string }

const accountOptions = ref<Option[]>([])

const {
  list, total, loading, dialogVisible, dialogTitle, form, isEdit, searchForm, pagination,
  fetchList, openCreate: openCreateBase, openEdit: openEditBase,
  handleSizeChange, handleCurrentChange, handleSearch, handleReset,
} = useCrud<InitialAccount>({
  baseUrl: '/base-data/initial-accounts',
  defaultForm: () => ({ accountId: undefined, amount: 0 }),
})

function typeLabel(t?: string): string {
  const map: Record<string, string> = { cash: '现金', bank: '银行', alipay: '支付宝', wechat: '微信' }
  return t ? (map[t] || t) : '-'
}

async function loadAccounts() {
  const res = await api.get('/accounts', { params: { page: 1, pageSize: 100 } })
  if (res.data.code === 0 || res.data.code === 200) {
    const data = res.data.data
    accountOptions.value = Array.isArray(data) ? data : (data?.list ?? [])
  }
}

function openCreate() { openCreateBase() }
function openEdit(row: InitialAccount) { openEditBase(row) }

async function handleSubmit() {
  if (!form.value.accountId) {
    ElMessage.warning('请选择账户')
    return
  }
  const res = await api.put('/base-data/initial-accounts', form.value)
  if (res.data.code === 0 || res.data.code === 200) {
    ElMessage.success('保存成功')
    dialogVisible.value = false
    await fetchList()
  } else {
    ElMessage.error(res.data.message || '保存失败')
  }
}

onMounted(loadAccounts)
</script>

<style scoped>
.card-header { display: flex; justify-content: space-between; align-items: center; }
.pagination { margin-top: 16px; justify-content: flex-end; }
</style>
