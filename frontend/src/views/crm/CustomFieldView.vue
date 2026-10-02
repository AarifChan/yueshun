<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">客户自定义字段</h2>
      <el-button type="primary" @click="openDialog()">新增客户字段</el-button>
    </div>
    <el-card>
      <el-table :data="list" v-loading="loading" border stripe>
        <el-table-column prop="name" label="字段名称" min-width="140" />
        <el-table-column label="字段类型" width="110">
          <template #default="{ row }">{{ typeLabel(row.fieldType) }}</template>
        </el-table-column>
        <el-table-column prop="options" label="下拉选项" min-width="150" show-overflow-tooltip />
        <el-table-column label="是否启用" width="90" align="center">
          <template #default="{ row }">
            <el-tag :type="row.enabled ? 'success' : 'info'">{{ row.enabled ? '启用' : '停用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="是否必填" width="90" align="center">
          <template #default="{ row }">{{ row.required ? '是' : '否' }}</template>
        </el-table-column>
        <el-table-column prop="defaultValue" label="默认值" width="120" />
        <el-table-column prop="sort" label="排序" width="70" align="right" />
        <el-table-column label="操作" width="130" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openDialog(row)">编辑</el-button>
            <el-button link type="danger" @click="remove(row)">删除</el-button>
          </template>
        </el-table-column>
        <template #empty>暂无数据</template>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="form.id ? '编辑字段' : '新增客户字段'" width="480px">
      <el-form :model="form" label-width="90px">
        <el-form-item label="字段名称" required><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="字段类型" required>
          <el-select v-model="form.fieldType" style="width: 100%">
            <el-option label="文本" value="text" />
            <el-option label="数字" value="number" />
            <el-option label="日期" value="date" />
            <el-option label="下拉选择" value="select" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="form.fieldType === 'select'" label="下拉选项">
          <el-input v-model="form.options" placeholder="选项用英文逗号分隔，如：A,B,C" />
        </el-form-item>
        <el-form-item label="默认值"><el-input v-model="form.defaultValue" /></el-form-item>
        <el-form-item label="是否必填"><el-switch v-model="form.required" /></el-form-item>
        <el-form-item label="是否启用"><el-switch v-model="form.enabled" /></el-form-item>
        <el-form-item label="排序"><el-input-number v-model="form.sort" :min="0" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="save">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '@/api/client'

interface Field {
  id: number; name: string; fieldType: string; options: string
  enabled: boolean; required: boolean; defaultValue: string; sort: number
}

const list = ref<Field[]>([])
const loading = ref(false)
const dialogVisible = ref(false)
const form = ref<Field>({ id: 0, name: '', fieldType: 'text', options: '', enabled: true, required: false, defaultValue: '', sort: 0 })

function typeLabel(t: string) {
  return t === 'number' ? '数字' : t === 'date' ? '日期' : t === 'select' ? '下拉选择' : '文本'
}

async function load() {
  loading.value = true
  try {
    const res = await api.get('/api/v1/custom-fields')
    if (res.data.code === 0 || res.data.code === 200) list.value = res.data.data?.list ?? []
  } finally {
    loading.value = false
  }
}

function openDialog(row?: Field) {
  form.value = row ? { ...row } : { id: 0, name: '', fieldType: 'text', options: '', enabled: true, required: false, defaultValue: '', sort: 0 }
  dialogVisible.value = true
}

async function save() {
  if (!form.value.name) { ElMessage.warning('请填写字段名称'); return }
  const res = form.value.id
    ? await api.put(`/api/v1/custom-fields/${form.value.id}`, form.value)
    : await api.post('/api/v1/custom-fields', form.value)
  if (res.data.code === 0 || res.data.code === 200) {
    ElMessage.success('保存成功')
    dialogVisible.value = false
    load()
  } else {
    ElMessage.error(res.data.message || '保存失败')
  }
}

async function remove(row: Field) {
  await ElMessageBox.confirm(`确定删除字段「${row.name}」？`, '提示', { type: 'warning' })
  const res = await api.delete(`/api/v1/custom-fields/${row.id}`)
  if (res.data.code === 0 || res.data.code === 200) {
    ElMessage.success('已删除')
    load()
  }
}

onMounted(load)
</script>

<style scoped>
.page { display: flex; flex-direction: column; gap: 12px; }
.page-header { display: flex; justify-content: space-between; align-items: center; }
.page-title { font-size: 18px; font-weight: 600; margin: 0; }
</style>
