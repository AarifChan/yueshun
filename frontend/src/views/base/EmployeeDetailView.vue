<template>
  <div class="page" v-loading="loading">
    <el-page-header @back="goBack" :title="pageTitle" />
    <el-card class="detail-card">
      <el-form :model="form" label-width="100px" :disabled="!isEditable">
        <el-row :gutter="20">
          <el-col :span="8"><el-form-item label="ID"><el-input v-model="form.id" disabled /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="用户名" required><el-input v-model="form.username" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="姓名" required><el-input v-model="form.name" /></el-form-item></el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="8"><el-form-item label="部门" required><RemoteSelect v-model="form.deptId" api-url="/api/v1/departments" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="角色" required><RemoteSelect v-model="form.roleId" api-url="/api/v1/roles" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="状态"><el-switch v-model="form.status" :active-value="1" :inactive-value="0" /></el-form-item></el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="8"><el-form-item label="手机号"><el-input v-model="form.phone" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="邮箱"><el-input v-model="form.email" /></el-form-item></el-col>
          <el-col :span="8" v-if="id === 'new'"><el-form-item label="密码"><el-input v-model="form.password" type="password" /></el-form-item></el-col>
        </el-row>
      </el-form>
    </el-card>
    <div class="footer-actions">
      <template v-if="isEditable"><el-button type="primary" @click="save" :loading="saving">保存</el-button></template>
      <template v-if="!isEditable && id !== 'new'"><el-button type="primary" @click="setMode('edit')">编辑</el-button></template>
      <el-button @click="goBack">返回</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, reactive } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import RemoteSelect from '@/components/RemoteSelect.vue'
import api from '@/api/client'
import { ElMessage } from 'element-plus'

const route = useRoute()
const router = useRouter()
const id = route.params.id as string
const routeMode = (route.query.mode as string) || 'view'

const form = reactive<Record<string, any>>({
  id: undefined, username: '', name: '', phone: '', deptId: undefined,
  roleId: undefined, status: 1, email: '', password: '',
})
const loading = ref(false)
const saving = ref(false)
const mode = ref<'view' | 'edit' | 'create'>('view')
const isEditable = computed(() => mode.value === 'edit' || mode.value === 'create')

const pageTitle = computed(() => {
  if (id === 'new') return '新增职员'
  if (mode.value === 'edit') return '编辑职员'
  return '职员详情'
})

function setMode(m: 'view' | 'edit' | 'create') {
  mode.value = m
}

function goBack() {
  router.push('/employees')
}

async function loadDetail(employeeId: string) {
  loading.value = true
  try {
    const res = await api.get(`/api/v1/employees/${employeeId}`)
    if (res.data.code === 0 || res.data.code === 200) {
      Object.assign(form, res.data.data)
    }
  } catch (e: any) {
    ElMessage.error(e.message || '加载失败')
  } finally {
    loading.value = false
  }
}

async function save() {
  saving.value = true
  try {
    const payload = { ...form }
    delete payload.id
    delete payload.createdAt
    delete payload.updatedAt
    if (id !== 'new') delete payload.password
    const res = id === 'new'
      ? await api.post('/api/v1/employees', payload)
      : await api.put(`/api/v1/employees/${id}`, payload)
    if (res.data.code === 0 || res.data.code === 200) {
      ElMessage.success(id === 'new' ? '创建成功' : '保存成功')
      if (id === 'new') {
        router.push('/employees')
      } else {
        setMode('view')
      }
    }
  } catch (e: any) {
    ElMessage.error(e.message || '保存失败')
  } finally {
    saving.value = false
  }
}

if (id === 'new') {
  setMode('create')
} else {
  loadDetail(id)
  if (routeMode === 'edit') setMode('edit')
}
</script>

<style scoped>
.page { padding: 20px; }
.detail-card { margin-top: 16px; }
.footer-actions { margin-top: 16px; text-align: center; }
</style>
