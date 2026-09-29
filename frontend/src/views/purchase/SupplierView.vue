<template>
  <div class="page">
    <div class="page-header">
      <span class="page-title">供应商管理</span>
      <el-button type="primary" :icon="Plus" @click="openCreate">新增</el-button>
    </div>
    <el-card>
      <div class="filter-bar">
        <el-input
          v-model="keyword"
          placeholder="名称/编码/联系人/电话"
          clearable
          class="filter-input"
          @keyup.enter="handleSearch"
          @clear="handleSearch"
        />
        <el-button type="primary" @click="handleSearch">查 询</el-button>
      </div>
      <el-table :data="list" v-loading="loading" row-key="id" border>
        <el-table-column label="操作" width="70" align="center">
          <template #default="{ row }">
            <el-dropdown trigger="click" @command="(cmd: string) => handleCommand(cmd, row)">
              <el-button text :icon="More" />
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="edit">编辑</el-dropdown-item>
                  <el-dropdown-item command="delete"><span class="danger-text">删除</span></el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </template>
        </el-table-column>
        <el-table-column prop="name" label="供应商名称" min-width="180" />
        <el-table-column label="联系人" min-width="120">
          <template #default="{ row }">{{ row.contact || '-' }}</template>
        </el-table-column>
        <el-table-column label="电话" min-width="140">
          <template #default="{ row }">{{ row.phone || '-' }}</template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.status === 1" type="success">已启用</el-tag>
            <el-tag v-else type="info">已禁用</el-tag>
          </template>
        </el-table-column>
      </el-table>
      <div class="pagination-bar">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :total="total"
          :page-sizes="[20, 50, 100]"
          layout="total, prev, pager, next, sizes"
          background
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑供应商' : '新增供应商'" width="480px" destroy-on-close>
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="100px">
        <el-form-item label="供应商名称" prop="name">
          <el-input v-model="form.name" placeholder="供应商名称" maxlength="50" show-word-limit />
        </el-form-item>
        <el-form-item label="编码">
          <el-input v-model="form.code" placeholder="留空自动生成" maxlength="50" />
        </el-form-item>
        <el-form-item label="联系人">
          <el-input v-model="form.contact" placeholder="联系人" maxlength="50" />
        </el-form-item>
        <el-form-item label="电话">
          <el-input v-model="form.phone" placeholder="电话" maxlength="30" />
        </el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="form.status" :active-value="1" :inactive-value="0" active-text="启用" inactive-text="禁用" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance } from 'element-plus'
import { More, Plus } from '@element-plus/icons-vue'
import {
  createSupplier,
  deleteSupplier,
  fetchSuppliers,
  updateSupplier,
  type Supplier,
  type SupplierPayload,
} from '@/api/supplier'

const list = ref<Supplier[]>([])
const loading = ref(false)
const keyword = ref('')
const total = ref(0)
const pagination = reactive({ page: 1, pageSize: 20 })

function ok(res: any) {
  return res.data.code === 0 || res.data.code === 200
}

async function fetchList() {
  loading.value = true
  try {
    const res = await fetchSuppliers({
      keyword: keyword.value || undefined,
      page: pagination.page,
      pageSize: pagination.pageSize,
    })
    if (ok(res)) {
      list.value = res.data.data?.list || []
      total.value = res.data.data?.total ?? list.value.length
    } else {
      ElMessage.error(res.data.message || '获取供应商失败')
    }
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  pagination.page = 1
  fetchList()
}

function handleSizeChange(size: number) {
  pagination.pageSize = size
  pagination.page = 1
  fetchList()
}

function handleCurrentChange(page: number) {
  pagination.page = page
  fetchList()
}

const dialogVisible = ref(false)
const isEdit = ref(false)
const currentId = ref<number | null>(null)
const submitting = ref(false)
const formRef = ref<FormInstance>()
const form = reactive({ name: '', code: '', contact: '', phone: '', status: 1 })
const formRules = {
  name: [{ required: true, message: '请输入供应商名称', trigger: 'blur' }],
}

function openCreate() {
  Object.assign(form, { name: '', code: '', contact: '', phone: '', status: 1 })
  isEdit.value = false
  currentId.value = null
  dialogVisible.value = true
}

function openEdit(row: Supplier) {
  Object.assign(form, {
    name: row.name,
    code: row.code || '',
    contact: row.contact || '',
    phone: row.phone || '',
    status: row.status ?? 1,
  })
  isEdit.value = true
  currentId.value = row.id
  dialogVisible.value = true
}

function handleCommand(cmd: string, row: Supplier) {
  if (cmd === 'edit') openEdit(row)
  else if (cmd === 'delete') handleDelete(row)
}

async function handleSubmit() {
  await formRef.value?.validate()
  submitting.value = true
  try {
    const payload: SupplierPayload = {
      name: form.name,
      code: form.code,
      contact: form.contact,
      phone: form.phone,
      status: form.status,
    }
    const res = isEdit.value && currentId.value !== null
      ? await updateSupplier(currentId.value, payload)
      : await createSupplier(payload)
    if (ok(res)) {
      ElMessage.success(isEdit.value ? '编辑成功' : '新增成功')
      dialogVisible.value = false
      fetchList()
    } else {
      ElMessage.error(res.data.message || '操作失败')
    }
  } finally {
    submitting.value = false
  }
}

async function handleDelete(row: Supplier) {
  try {
    await ElMessageBox.confirm(`确认删除供应商“${row.name}”？`, '提示', { type: 'warning' })
  } catch {
    return
  }
  const res = await deleteSupplier(row.id)
  if (ok(res)) {
    ElMessage.success('删除成功')
    fetchList()
  } else {
    ElMessage.error(res.data.message || '删除失败')
  }
}

onMounted(fetchList)
</script>

<style scoped>
.page { padding: 20px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.page-title { font-size: 18px; font-weight: 600; }
.filter-bar { display: flex; gap: 8px; margin-bottom: 16px; }
.filter-input { width: 260px; }
.pagination-bar { display: flex; justify-content: flex-end; margin-top: 16px; }
.danger-text { color: var(--el-color-danger); }
</style>
