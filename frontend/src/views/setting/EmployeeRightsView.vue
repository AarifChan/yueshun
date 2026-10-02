<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">职员权限</h2>
      <div>
        <el-input v-model="keyword" placeholder="搜索职员姓名/账号" clearable style="width: 220px; margin-right: 8px" @keyup.enter="search" />
        <el-button type="primary" icon="Search" @click="search">查询</el-button>
        <el-button icon="Refresh" @click="load">刷新</el-button>
      </div>
    </div>
    <el-alert type="info" :closable="false" style="margin-bottom: 12px"
      title="职员权限说明：权限组控制功能菜单与按钮权限（在「基础 → 角色权限」维护），此处维护各职员的数据管辖范围与许可账号类型。企微许可账号用于企业微信集成登录。" />
    <el-card>
      <el-table :data="list" v-loading="loading" border stripe>
        <el-table-column prop="name" label="职员" width="110" fixed="left" />
        <el-table-column label="已授权应用" min-width="150">
          <template #default="{ row }">{{ row.appScopes || '全部应用' }}</template>
        </el-table-column>
        <el-table-column prop="roleName" label="权限组" width="110">
          <template #default="{ row }">{{ row.roleName || '-' }}</template>
        </el-table-column>
        <el-table-column label="客户管辖范围" width="110" align="center">
          <template #default="{ row }">{{ scopeLabel(row.customerScope) }}</template>
        </el-table-column>
        <el-table-column label="职员管辖范围" width="110" align="center">
          <template #default="{ row }">{{ scopeLabel(row.employeeScope) }}</template>
        </el-table-column>
        <el-table-column label="仓库管辖范围" width="110" align="center">
          <template #default="{ row }">{{ row.warehouseScope ? `${row.warehouseScope.split(',').length} 个仓库` : '全部' }}</template>
        </el-table-column>
        <el-table-column label="供应商管辖范围" width="120" align="center">
          <template #default="{ row }">{{ scopeLabel(row.supplierScope) }}</template>
        </el-table-column>
        <el-table-column label="商品管辖范围" width="110" align="center">
          <template #default="{ row }">{{ scopeLabel(row.productScope) }}</template>
        </el-table-column>
        <el-table-column label="现金银行账户" width="110" align="center">
          <template #default="{ row }">{{ row.accountScope ? `${row.accountScope.split(',').length} 个账户` : '全部' }}</template>
        </el-table-column>
        <el-table-column label="允许登录时间" width="120" align="center">
          <template #default="{ row }">{{ row.loginTimeRange || '不限' }}</template>
        </el-table-column>
        <el-table-column label="许可账号类型" width="120" align="center">
          <template #default="{ row }">
            <el-tag :type="row.licenseType === 'wecom' ? 'success' : 'info'" size="small">
              {{ row.licenseType === 'wecom' ? '企微许可' : '普通账号' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="90" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openForm(row)">设置</el-button>
          </template>
        </el-table-column>
        <template #empty>暂无数据</template>
      </el-table>
      <el-pagination
        v-model:current-page="page"
        v-model:page-size="pageSize"
        :total="total"
        layout="total, prev, pager, next"
        style="margin-top: 12px; justify-content: flex-end"
        @current-change="load"
      />
    </el-card>

    <el-dialog v-model="formVisible" :title="`设置职员权限 - ${current?.name || ''}`" width="640px">
      <el-form label-width="150px">
        <el-form-item label="已授权应用">
          <el-select v-model="formAppScopes" multiple style="width: 100%" placeholder="不选表示全部应用">
            <el-option v-for="a in appOptions" :key="a" :label="a" :value="a" />
          </el-select>
        </el-form-item>
        <el-form-item label="客户管辖范围">
          <el-radio-group v-model="form.customerScope">
            <el-radio value="all">全部客户</el-radio>
            <el-radio value="self">仅自己负责</el-radio>
            <el-radio value="dept">本部门</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="职员管辖范围">
          <el-radio-group v-model="form.employeeScope">
            <el-radio value="all">全部职员</el-radio>
            <el-radio value="self">仅自己</el-radio>
            <el-radio value="dept">本部门</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="仓库管辖范围">
          <el-select v-model="formWarehouseIds" multiple style="width: 100%" placeholder="不选表示全部仓库">
            <el-option v-for="w in warehouses" :key="w.id" :label="w.name" :value="w.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="供应商管辖范围">
          <el-radio-group v-model="form.supplierScope">
            <el-radio value="all">全部供应商</el-radio>
            <el-radio value="self">仅自己负责</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="商品管辖范围">
          <el-radio-group v-model="form.productScope">
            <el-radio value="all">全部商品</el-radio>
            <el-radio value="part">部分分类商品</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="现金银行账户">
          <el-select v-model="formAccountIds" multiple style="width: 100%" placeholder="不选表示全部账户">
            <el-option v-for="a in accounts" :key="a.id" :label="a.name" :value="a.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="允许登录时间">
          <el-input v-model="form.loginTimeRange" placeholder="如 08:00-22:00，留空不限" style="width: 240px" />
        </el-form-item>
        <el-form-item label="许可账号类型">
          <el-radio-group v-model="form.licenseType">
            <el-radio value="normal">普通账号</el-radio>
            <el-radio value="wecom">企微许可账号</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="save">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import api from '@/api/client'

interface Row {
  employeeId: number; name: string; username: string; roleName: string
  appScopes: string; customerScope: string; employeeScope: string; warehouseScope: string
  supplierScope: string; productScope: string; accountScope: string; loginTimeRange: string; licenseType: string
}
interface Option { id: number; name: string }

const list = ref<Row[]>([])
const loading = ref(false)
const page = ref(1)
const pageSize = ref(30)
const total = ref(0)
const keyword = ref('')

const warehouses = ref<Option[]>([])
const accounts = ref<Option[]>([])
const appOptions = ['进销存', '订货商城', 'CRM', '营销', 'BI 分析', '生态互联']

const formVisible = ref(false)
const saving = ref(false)
const current = ref<Row | null>(null)
const form = ref({
  customerScope: 'all', employeeScope: 'all', supplierScope: 'all',
  productScope: 'all', loginTimeRange: '', licenseType: 'normal',
})
const formAppScopes = ref<string[]>([])
const formWarehouseIds = ref<number[]>([])
const formAccountIds = ref<number[]>([])

function scopeLabel(v: string) {
  return v === 'self' ? '仅自己' : v === 'dept' ? '本部门' : v === 'part' ? '部分' : '全部'
}

async function load() {
  loading.value = true
  try {
    const res = await api.get('/api/v1/employee-permissions', {
      params: { page: page.value, pageSize: pageSize.value, keyword: keyword.value || undefined },
    })
    if (res.data.code === 0 || res.data.code === 200) {
      list.value = res.data.data?.list ?? []
      total.value = res.data.data?.total ?? 0
    }
  } finally {
    loading.value = false
  }
}

function search() {
  page.value = 1
  load()
}

async function loadOptions() {
  const [w, a] = await Promise.all([
    api.get('/api/v1/warehouses', { params: { page: 1, pageSize: 200 } }),
    api.get('/api/v1/accounts', { params: { page: 1, pageSize: 200 } }),
  ])
  if (w.data.code === 0 || w.data.code === 200) warehouses.value = w.data.data?.list ?? []
  if (a.data.code === 0 || a.data.code === 200) accounts.value = a.data.data?.list ?? []
}

function openForm(row: Row) {
  current.value = row
  form.value = {
    customerScope: row.customerScope || 'all',
    employeeScope: row.employeeScope || 'all',
    supplierScope: row.supplierScope || 'all',
    productScope: row.productScope || 'all',
    loginTimeRange: row.loginTimeRange || '',
    licenseType: row.licenseType || 'normal',
  }
  formAppScopes.value = row.appScopes ? row.appScopes.split(',').filter(Boolean) : []
  formWarehouseIds.value = row.warehouseScope ? row.warehouseScope.split(',').filter(Boolean).map(Number) : []
  formAccountIds.value = row.accountScope ? row.accountScope.split(',').filter(Boolean).map(Number) : []
  formVisible.value = true
}

async function save() {
  if (!current.value) return
  saving.value = true
  try {
    const res = await api.put(`/api/v1/employee-permissions/${current.value.employeeId}`, {
      ...form.value,
      appScopes: formAppScopes.value.join(','),
      warehouseScope: formWarehouseIds.value.join(','),
      accountScope: formAccountIds.value.join(','),
    })
    if (res.data.code === 0 || res.data.code === 200) {
      ElMessage.success('已保存')
      formVisible.value = false
      load()
    } else {
      ElMessage.error(res.data.message || '保存失败')
    }
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  load()
  loadOptions()
})
</script>

<style scoped>
.page { display: flex; flex-direction: column; gap: 12px; }
.page-header { display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 8px; }
.page-title { font-size: 18px; font-weight: 600; margin: 0; }
</style>
