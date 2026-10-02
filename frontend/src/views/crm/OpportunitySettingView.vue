<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">商机设置</h2>
      <el-button type="primary" @click="openDialog()">新增模板</el-button>
    </div>
    <el-card>
      <el-table :data="list" v-loading="loading" border stripe>
        <el-table-column prop="name" label="模板名称" min-width="140" />
        <el-table-column label="模板状态" width="90" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'">{{ row.status === 1 ? '启用' : '停用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="商机阶段" min-width="200">
          <template #default="{ row }">
            <el-tag v-for="s in stages(row.stages)" :key="s" size="small" class="stage-tag">{{ s }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="可使用人" min-width="140">
          <template #default="{ row }">{{ row.userIds ? userNames(row.userIds) : '全部' }}</template>
        </el-table-column>
        <el-table-column prop="description" label="模板说明" min-width="140" show-overflow-tooltip />
        <el-table-column label="创建时间" width="160">
          <template #default="{ row }">{{ row.createdAt?.replace('T', ' ').slice(0, 19) }}</template>
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
        <el-form-item label="模板名称" required><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="商机阶段">
          <el-select v-model="formStages" multiple filterable allow-create default-first-option
            placeholder="输入阶段名后回车添加，如：初步接洽" style="width: 100%">
          </el-select>
        </el-form-item>
        <el-form-item label="可使用人">
          <el-select v-model="formUserIds" multiple filterable placeholder="留空表示全部" style="width: 100%">
            <el-option v-for="e in employees" :key="e.id" :label="e.name" :value="e.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="模板说明"><el-input v-model="form.description" type="textarea" /></el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="form.status" :active-value="1" :inactive-value="0" active-text="启用" />
        </el-form-item>
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
  id: number; name: string; stages: string; userIds: string
  description: string; status: number; sort: number; createdAt: string
}
interface Opt { id: number; name: string }

const list = ref<Tpl[]>([])
const loading = ref(false)
const dialogVisible = ref(false)
const form = ref<Tpl>({ id: 0, name: '', stages: '', userIds: '', description: '', status: 1, sort: 0, createdAt: '' })
const formStages = ref<string[]>([])
const formUserIds = ref<number[]>([])
const employees = ref<Opt[]>([])

const stages = (s: string) => (s ? s.split(',').filter(Boolean) : [])
function userNames(ids: string) {
  return ids.split(',').map((s) => employees.value.find((e) => e.id === Number(s))?.name || s).join('、')
}

async function load() {
  loading.value = true
  try {
    const res = await api.get('/api/v1/opportunity-templates')
    if (res.data.code === 0 || res.data.code === 200) list.value = res.data.data?.list ?? []
  } finally {
    loading.value = false
  }
}

function openDialog(row?: Tpl) {
  form.value = row ? { ...row } : { id: 0, name: '', stages: '初步接洽,需求确认,方案报价,商务谈判,赢单,输单', userIds: '', description: '', status: 1, sort: 0, createdAt: '' }
  formStages.value = row ? stages(row.stages) : stages(form.value.stages)
  formUserIds.value = row?.userIds ? row.userIds.split(',').map(Number).filter(Boolean) : []
  dialogVisible.value = true
}

async function save() {
  if (!form.value.name) { ElMessage.warning('请填写模板名称'); return }
  const payload = { ...form.value, stages: formStages.value.join(','), userIds: formUserIds.value.join(',') }
  const res = form.value.id
    ? await api.put(`/api/v1/opportunity-templates/${form.value.id}`, payload)
    : await api.post('/api/v1/opportunity-templates', payload)
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
  const res = await api.delete(`/api/v1/opportunity-templates/${row.id}`)
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
.stage-tag { margin-right: 6px; }
</style>
