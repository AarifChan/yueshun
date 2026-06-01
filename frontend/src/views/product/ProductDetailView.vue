<template>
  <div class="page" v-loading="loading">
    <el-page-header @back="goBack" :title="pageTitle" />
    <el-card class="detail-card">
      <el-form :model="form" label-width="100px" :disabled="!isEditable">
        <el-row :gutter="20">
          <el-col :span="8"><el-form-item label="ID"><el-input v-model="form.id" disabled /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="编码"><el-input v-model="form.code" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="条码"><el-input v-model="form.barcode" /></el-form-item></el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="8"><el-form-item label="名称" required><el-input v-model="form.name" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="分类" required><RemoteSelect v-model="form.categoryId" api-url="/api/v1/products/categories" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="品牌"><RemoteSelect v-model="form.brandId" api-url="/api/v1/products/brands" /></el-form-item></el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="8"><el-form-item label="规格"><el-input v-model="form.specification" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="单位" required><el-input v-model="form.unit" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="状态"><el-switch v-model="form.status" :active-value="1" :inactive-value="0" /></el-form-item></el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="8"><el-form-item label="采购价"><el-input-number v-model="form.purchasePrice" :precision="2" :min="0" style="width:100%" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="零售价"><el-input-number v-model="form.retailPrice" :precision="2" :min="0" style="width:100%" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="批发价"><el-input-number v-model="form.wholesalePrice" :precision="2" :min="0" style="width:100%" /></el-form-item></el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="8"><el-form-item label="最低库存"><el-input-number v-model="form.minStock" :min="0" style="width:100%" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="最高库存"><el-input-number v-model="form.maxStock" :min="0" style="width:100%" /></el-form-item></el-col>
        </el-row>
        <el-form-item label="描述"><el-input v-model="form.description" type="textarea" /></el-form-item>
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
  id: undefined, name: '', code: '', barcode: '', categoryId: undefined, brandId: undefined,
  specification: '', unit: '', purchasePrice: 0, retailPrice: 0, wholesalePrice: 0,
  minStock: 0, maxStock: 0, description: '', status: 1,
})
const loading = ref(false)
const saving = ref(false)
const mode = ref<'view' | 'edit' | 'create'>('view')
const isEditable = computed(() => mode.value === 'edit' || mode.value === 'create')

const pageTitle = computed(() => {
  if (id === 'new') return '新增商品'
  if (mode.value === 'edit') return '编辑商品'
  return '商品详情'
})

function setMode(m: 'view' | 'edit' | 'create') {
  mode.value = m
}

function goBack() {
  router.push('/products')
}

async function loadDetail(productId: string) {
  loading.value = true
  try {
    const res = await api.get(`/api/v1/products/${productId}`)
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
    delete payload.categoryName
    delete payload.brandName
    const res = id === 'new'
      ? await api.post('/api/v1/products', payload)
      : await api.put(`/api/v1/products/${id}`, payload)
    if (res.data.code === 0 || res.data.code === 200) {
      ElMessage.success(id === 'new' ? '创建成功' : '保存成功')
      if (id === 'new') {
        router.push('/products')
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
