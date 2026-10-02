<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">汇报模板</h2>
      <el-button type="primary" @click="openDialog()">新增模板</el-button>
    </div>
    <el-card>
      <el-tabs v-model="statusTab" @tab-change="load">
        <el-tab-pane label="汇报模板" name="1" />
        <el-tab-pane label="已停用模板" name="0" />
      </el-tabs>
      <el-table :data="list" v-loading="loading" border stripe>
        <el-table-column label="模板分类" width="100">
          <template #default="{ row }">{{ categoryLabel(row.category) }}</template>
        </el-table-column>
        <el-table-column prop="name" label="模板名称" min-width="150" />
        <el-table-column label="可使用人" min-width="150">
          <template #default="{ row }">{{ row.userIds ? userNames(row.userIds) : '全部' }}</template>
        </el-table-column>
        <el-table-column prop="description" label="模板说明" min-width="160" show-overflow-tooltip />
        <el-table-column label="创建时间" width="160">
          <template #default="{ row }">{{ row.createdAt?.replace('T', ' ').slice(0, 19) }}</template>
        </el-table-column>
        <el-table-column label="更新时间" width="160">
          <template #default="{ row }">{{ row.updatedAt?.replace('T', ' ').slice(0, 19) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="130" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openDialog(row)">编辑</el-button>
            <el-button link type="danger" @click="remove(row)">删除</el-button>
          </template>
        </el-table-column>
        <template #empty>暂无数据</template>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="form.id ? '编辑模板' : '新增模板'" width="520px">
      <el-form :model="form" label-width="90px">
        <el-form-item label="模板分类" required>
          <el-radio-group v-model="form.category">
            <el-radio-button value="log">日志</el-radio-button>
            <el-radio-button value="week">周报</el-radio-button>
            <el-radio-button value="month">月报</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="模板名称" required><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="模板说明"><el-input v-model="form.description" type="textarea" /></el-form-item>
        <el-form-item label="可使用人">
          <el-select v-model="formUserIds" multiple filterable placeholder="留空表示全部" style="width: 100%">
            <el-option v-for="e in employees" :key="e.id" :label="e.name" :value="e.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="form.status" :active-value="1" :inactive-value="0" active-text="启用" />
        </el-form-item>
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

interface Tpl {
  id: number; category: string; name: string; description: string
  userIds: string; status: number; sort: number; createdAt: string; updatedAt: string
}
interface Opt { id: number; name: string }

const list = ref<Tpl[]>([])
const loading = ref(false)
const statusTab = ref('1')
const dialogVisible = ref(false)
const form = ref<Tpl>({ id: 0, category: 'log', name: '', description: '', userIds: '', status: 1, sort: 0, createdAt: '', updatedAt: '' })
const formUserIds = ref<number[]>([])
const employees = ref<Opt[]>([])

function categoryLabel(t: string) {
  return t === 'week' ? '周报' : t === 'month' ? '月报' : '日志'
}
function userNames(ids: string) {
  return ids.split(',').map((s) => employees.value.find((e) => e.id === Number(s))?.name || s).join('、')
}

async function load() {
  loading.value = true
  try {
    const res = await api.get('/api/v1/report-templates', { params: { status: statusTab.value } })
    if (res.data.code === 0 || res.data.code === 200) list.value = res.data.data?.list ?? []
  } finally {
    loading.value = false
  }
}

function openDialog(row?: Tpl) {
  form.value = row ? { ...row } : { id: 0, category: 'log', name: '', description: '', userIds: '', status: 1, sort: 0, createdAt: '', updatedAt: '' }
  formUserIds.value = row?.userIds ? row.userIds.split(',').map(Number).filter(Boolean) : []
  dialogVisible.value = true
}

async function save() {
  if (!form.value.name) { ElMessage.warning('请填写模板名称'); return }
  const payload = { ...form.value, userIds: formUserIds.value.join(',') }
  const res = form.value.id
    ? await api.put(`/api/v1/report-templates/${form.value.id}`, payload)
    : await api.post('/api/v1/report-templates', payload)
  if (res.data.code === 0 || res.data.code === 200) {
    ElMessage.success('保存成功')
    dialogVisible.value = false
    load()
  } else {
    ElMessage.error(res.data.message || '保存失败')
  }
}

async function remove(row: Tpl) {
  await ElMessageBox.confirm(`确定删除模板「${row.name}」？`, '提示', { type: 'warning' })
  const res = await api.delete(`/api/v1/report-templates/${row.id}`)
  if (res.data.code === 0 || res.data.code === 200) {
    ElMessage.success('已删除')
    load()
  }
}

onMounted(async () => {
  load()
  const res = await api.get('/api/v1/employees', { params: { page: 1, pageSize: 200 } })
  if (res.data.code === 0 || res.data.code === 200) employees.value = res.data.data?.list ?? []
})
</script>

<style scoped>
.page { display: flex; flex-direction: column; gap: 12px; }
.page-header { display: flex; justify-content: space-between; align-items: center; }
.page-title { font-size: 18px; font-weight: 600; margin: 0; }
</style>
