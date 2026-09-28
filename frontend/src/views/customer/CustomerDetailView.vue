<template>
  <div class="page" v-loading="loading">
    <el-page-header @back="goBack" :title="pageTitle" />
    <el-card class="detail-card">
      <el-form :model="form" label-width="100px" :disabled="!isEditable">
        <el-row :gutter="20">
          <el-col :span="8"><el-form-item label="ID"><el-input v-model="form.id" disabled /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="编码"><el-input v-model="form.code" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="名称" required><el-input v-model="form.name" /></el-form-item></el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="8"><el-form-item label="类型" required>
            <el-select v-model="form.type"><el-option label="客户" value="customer" /><el-option label="供应商" value="supplier" /><el-option label="两者都是" value="both" /></el-select>
          </el-form-item></el-col>
          <el-col :span="8"><el-form-item label="联系人"><el-input v-model="form.contact" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="电话"><el-input v-model="form.phone" /></el-form-item></el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="8"><el-form-item label="邮箱"><el-input v-model="form.email" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="信用额度"><el-input-number v-model="form.creditLimit" :precision="2" :min="0" style="width:100%" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="账期"><el-input-number v-model="form.creditDays" :min="0" style="width:100%" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="欠款余额"><el-input :model-value="form.balance != null ? `¥${Number(form.balance).toFixed(2)}` : '-'" disabled /></el-form-item></el-col>
        </el-row>
        <el-form-item label="地址"><el-input v-model="form.address" /></el-form-item>
        <el-form-item label="状态"><el-switch v-model="form.status" :active-value="1" :inactive-value="0" /></el-form-item>
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
import api from '@/api/client'
import { ElMessage } from 'element-plus'

const route = useRoute()
const router = useRouter()
const id = route.params.id as string
const routeMode = (route.query.mode as string) || 'view'

const form = reactive<Record<string, any>>({
  id: undefined, name: '', code: '', type: 'customer', contact: '', phone: '',
  email: '', address: '', creditLimit: 0, creditDays: 0, status: 1,
})
const loading = ref(false)
const saving = ref(false)
const mode = ref<'view' | 'edit' | 'create'>('view')
const isEditable = computed(() => mode.value === 'edit' || mode.value === 'create')

const pageTitle = computed(() => {
  if (id === 'new') return '新增客户'
  if (mode.value === 'edit') return '编辑客户'
  return '客户详情'
})

function setMode(m: 'view' | 'edit' | 'create') {
  mode.value = m
}

function goBack() {
  router.push('/customers')
}

async function loadDetail(customerId: string) {
  loading.value = true
  try {
    const res = await api.get(`/api/v1/customers/${customerId}`)
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
    const res = id === 'new'
      ? await api.post('/api/v1/customers', payload)
      : await api.put(`/api/v1/customers/${id}`, payload)
    if (res.data.code === 0 || res.data.code === 200) {
      ElMessage.success(id === 'new' ? '创建成功' : '保存成功')
      if (id === 'new') {
        router.push('/customers')
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
