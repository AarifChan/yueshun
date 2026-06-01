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
          <el-col :span="8"><el-form-item label="客户"><RemoteSelect v-model="form.customerId" api-url="/api/v1/customers" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="金额"><el-input-number v-model="form.amount" :precision="2" :min="0" style="width:100%" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="状态">
            <el-select v-model="form.status">
              <el-option label="草稿" value="draft" /><el-option label="生效" value="active" /><el-option label="终止" value="terminated" />
            </el-select>
          </el-form-item></el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="8"><el-form-item label="签约日期"><el-date-picker v-model="form.signDate" type="date" value-format="YYYY-MM-DD" style="width:100%" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="到期日期"><el-date-picker v-model="form.expireDate" type="date" value-format="YYYY-MM-DD" style="width:100%" /></el-form-item></el-col>
        </el-row>
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
  id: undefined, name: '', code: '', customerId: undefined, amount: 0,
  status: 'draft', signDate: '', expireDate: '', remark: '',
})
const loading = ref(false)
const saving = ref(false)
const mode = ref<'view' | 'edit' | 'create'>('view')
const isEditable = computed(() => mode.value === 'edit' || mode.value === 'create')

const pageTitle = computed(() => {
  if (id === 'new') return '新增合同'
  if (mode.value === 'edit') return '编辑合同'
  return '合同详情'
})

function setMode(m: 'view' | 'edit' | 'create') {
  mode.value = m
}

function goBack() {
  router.push('/contracts')
}

async function loadDetail(contractId: string) {
  loading.value = true
  try {
    const res = await api.get(`/api/v1/contracts/${contractId}`)
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
      ? await api.post('/api/v1/contracts', payload)
      : await api.put(`/api/v1/contracts/${id}`, payload)
    if (res.data.code === 0 || res.data.code === 200) {
      ElMessage.success(id === 'new' ? '创建成功' : '保存成功')
      if (id === 'new') {
        router.push('/contracts')
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
