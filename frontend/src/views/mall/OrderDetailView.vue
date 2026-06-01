<template>
  <div class="page" v-loading="loading">
    <el-page-header @back="goBack" :title="pageTitle" />
    <el-card class="detail-card">
      <el-form :model="form" label-width="100px" :disabled="!isEditable">
        <el-row :gutter="20">
          <el-col :span="8"><el-form-item label="订单编号"><el-input v-model="form.orderNo" disabled /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="客户"><RemoteSelect v-model="form.customerId" api-url="/api/v1/customers" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="状态">
            <el-select v-model="form.status">
              <el-option label="待付款" value="pending" /><el-option label="已付款" value="paid" />
              <el-option label="已发货" value="shipped" /><el-option label="已完成" value="completed" />
            </el-select>
          </el-form-item></el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="8"><el-form-item label="总金额"><el-input-number v-model="form.totalAmount" :precision="2" :min="0" style="width:100%" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="运费"><el-input-number v-model="form.freight" :precision="2" :min="0" style="width:100%" /></el-form-item></el-col>
        </el-row>
        <el-form-item label="收货地址"><el-input v-model="form.address" type="textarea" /></el-form-item>
        <el-form-item label="备注"><el-input v-model="form.remark" type="textarea" /></el-form-item>
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
  id: undefined, orderNo: '', customerId: undefined, status: 'pending',
  totalAmount: 0, freight: 0, address: '', remark: '',
})
const loading = ref(false)
const saving = ref(false)
const mode = ref<'view' | 'edit' | 'create'>('view')
const isEditable = computed(() => mode.value === 'edit' || mode.value === 'create')

const pageTitle = computed(() => {
  if (id === 'new') return '新增商城订单'
  if (mode.value === 'edit') return '编辑商城订单'
  return '商城订单详情'
})

function setMode(m: 'view' | 'edit' | 'create') {
  mode.value = m
}

function goBack() {
  router.push('/mall-orders')
}

async function loadDetail(orderId: string) {
  loading.value = true
  try {
    const res = await api.get(`/api/v1/mall-orders/${orderId}`)
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
    delete payload.orderNo
    const res = id === 'new'
      ? await api.post('/api/v1/mall-orders', payload)
      : await api.put(`/api/v1/mall-orders/${id}`, payload)
    if (res.data.code === 0 || res.data.code === 200) {
      ElMessage.success(id === 'new' ? '创建成功' : '保存成功')
      if (id === 'new') {
        router.push('/mall-orders')
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
