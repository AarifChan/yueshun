<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">客户标签</h2>
      <el-button type="primary" @click="openDialog()">新增标签</el-button>
    </div>
    <el-card>
      <el-form inline>
        <el-form-item label="标签名称"><el-input v-model="keyword" placeholder="标签名称" clearable @keyup.enter="load" /></el-form-item>
        <el-form-item><el-button type="primary" @click="load">搜索</el-button></el-form-item>
      </el-form>
      <el-table :data="list" v-loading="loading" border stripe>
        <el-table-column prop="name" label="标签名称" min-width="150" />
        <el-table-column prop="tagType" label="标签类型" width="110" />
        <el-table-column prop="customerCount" label="绑定客户数" width="110" align="right" />
        <el-table-column label="状态" width="90" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'">{{ row.status === 1 ? '启用' : '停用' }}</el-tag>
          </template>
        </el-table-column>
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

    <el-dialog v-model="dialogVisible" :title="form.id ? '编辑标签' : '新增标签'" width="420px">
      <el-form :model="form" label-width="90px">
        <el-form-item label="标签名称" required><el-input v-model="form.name" /></el-form-item>
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

interface Tag { id: number; name: string; status: number; tagType: string; customerCount: number; createdAt: string }

const list = ref<Tag[]>([])
const loading = ref(false)
const keyword = ref('')
const dialogVisible = ref(false)
const form = ref<Tag>({ id: 0, name: '', status: 1, tagType: '手动标签', customerCount: 0, createdAt: '' })

async function load() {
  loading.value = true
  try {
    const res = await api.get('/api/v1/crm/customer-tags', { params: { keyword: keyword.value } })
    if (res.data.code === 0 || res.data.code === 200) list.value = res.data.data?.list ?? []
  } finally {
    loading.value = false
  }
}

function openDialog(row?: Tag) {
  form.value = row ? { ...row } : { id: 0, name: '', status: 1, tagType: '手动标签', customerCount: 0, createdAt: '' }
  dialogVisible.value = true
}

async function save() {
  if (!form.value.name) { ElMessage.warning('请填写标签名称'); return }
  const res = form.value.id
    ? await api.put(`/api/v1/crm/customer-tags/${form.value.id}`, form.value)
    : await api.post('/api/v1/crm/customer-tags', form.value)
  if (res.data.code === 0 || res.data.code === 200) {
    ElMessage.success('保存成功')
    dialogVisible.value = false
    load()
  } else {
    ElMessage.error(res.data.message || '保存失败')
  }
}

async function remove(row: Tag) {
  await ElMessageBox.confirm(`确定删除标签「${row.name}」？`, '提示', { type: 'warning' })
  const res = await api.delete(`/api/v1/crm/customer-tags/${row.id}`)
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
