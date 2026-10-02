<template>
  <div class="page">
    <el-card>
      <template #header>
        <div class="card-header">
          <el-tabs v-model="bizType" @tab-change="onTabChange">
            <el-tab-pane label="客户应收预收期初" name="customer" />
            <el-tab-pane label="供应商应付预付期初" name="supplier" />
          </el-tabs>
          <el-button type="primary" @click="openCreate">新增期初</el-button>
        </div>
      </template>
      <el-form :model="searchForm" inline class="search-form">
        <el-form-item :label="bizType === 'customer' ? '客户' : '供应商'">
          <el-input v-model="searchForm.keyword" placeholder="名称/编号" clearable />
        </el-form-item>
        <el-form-item><el-button type="primary" @click="handleSearch">搜索</el-button><el-button @click="handleReset">重置</el-button></el-form-item>
      </el-form>
      <el-table :data="list" v-loading="loading" border stripe>
        <el-table-column prop="targetCode" label="编号" width="120" />
        <el-table-column prop="targetName" :label="bizType === 'customer' ? '客户名称' : '供应商名称'" min-width="180" />
        <el-table-column :label="bizType === 'customer' ? '应收期初' : '应付期初'" width="140" align="right">
          <template #default="{ row }">{{ row.receivable?.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column :label="bizType === 'customer' ? '预收期初' : '预付期初'" width="140" align="right">
          <template #default="{ row }">{{ row.advance?.toFixed(2) }}</template>
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
      <el-form :model="form" label-width="110px">
        <el-form-item :label="bizType === 'customer' ? '客户' : '供应商'" required>
          <el-select v-model="form.targetId" filterable remote :remote-method="searchTargets" placeholder="请选择" style="width: 100%" :disabled="isEdit">
            <el-option v-for="t in targetOptions" :key="t.id" :label="`${t.name}${t.code ? '(' + t.code + ')' : ''}`" :value="t.id" />
          </el-select>
        </el-form-item>
        <el-form-item :label="bizType === 'customer' ? '应收期初' : '应付期初'">
          <el-input-number v-model="form.receivable" :min="0" :precision="2" style="width: 100%" />
        </el-form-item>
        <el-form-item :label="bizType === 'customer' ? '预收期初' : '预付期初'">
          <el-input-number v-model="form.advance" :min="0" :precision="2" style="width: 100%" />
        </el-form-item>
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

interface InitialBalance {
  id: number; bizType: string; targetId: number; receivable: number; advance: number
  targetName?: string; targetCode?: string
}
interface Option { id: number; name: string; code?: string }

const bizType = ref<'customer' | 'supplier'>('customer')
const targetOptions = ref<Option[]>([])

const {
  list, total, loading, dialogVisible, dialogTitle, form, isEdit, searchForm, pagination,
  fetchList, openCreate: openCreateBase, openEdit: openEditBase,
  handleSizeChange, handleCurrentChange, handleSearch, handleReset,
} = useCrud<InitialBalance>({
  baseUrl: '/base-data/initial-balances',
  defaultForm: () => ({ targetId: undefined, receivable: 0, advance: 0 }),
})

const listUrl = '/base-data/initial-balances'

async function searchTargets(kw: string) {
  const url = bizType.value === 'customer' ? '/customers' : '/suppliers'
  const res = await api.get(url, { params: { page: 1, pageSize: 20, keyword: kw || '' } })
  if (res.data.code === 0 || res.data.code === 200) {
    const data = res.data.data
    targetOptions.value = Array.isArray(data) ? data : (data?.list ?? [])
  }
}

function onTabChange() {
  targetOptions.value = []
  searchTargets('')
  pagination.value.page = 1
  fetchList({ bizType: bizType.value })
}

function openCreate() { openCreateBase() }
function openEdit(row: InitialBalance) {
  if (row.targetName && !targetOptions.value.find(t => t.id === row.targetId)) {
    targetOptions.value = [...targetOptions.value, { id: row.targetId, name: row.targetName, code: row.targetCode }]
  }
  openEditBase(row)
}

async function handleSubmit() {
  if (!form.value.targetId) {
    ElMessage.warning('请选择往来单位')
    return
  }
  const res = await api.put(listUrl, { ...form.value, bizType: bizType.value })
  if (res.data.code === 0 || res.data.code === 200) {
    ElMessage.success('保存成功')
    dialogVisible.value = false
    await fetchList({ bizType: bizType.value })
  } else {
    ElMessage.error(res.data.message || '保存失败')
  }
}

onMounted(() => { searchTargets(''); fetchList({ bizType: bizType.value }) })
</script>

<style scoped>
.card-header { display: flex; justify-content: space-between; align-items: center; }
.card-header :deep(.el-tabs) { flex: 1; }
.card-header :deep(.el-tabs__header) { margin-bottom: 0; }
.pagination { margin-top: 16px; justify-content: flex-end; }
</style>
