<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">商城装修接收</h2>
      <el-button v-if="tab === 'pending'" type="primary" :disabled="!list.length" @click="receiveAll">一键接收</el-button>
    </div>
    <el-card>
      <el-tabs v-model="tab" @tab-change="onTab">
        <el-tab-pane label="自动接收" name="pending" />
        <el-tab-pane label="已关闭" name="closed" />
      </el-tabs>
      <el-table :data="list" v-loading="loading" border stripe>
        <el-table-column type="index" label="序" width="60" />
        <el-table-column prop="title" label="装修方案" min-width="180" show-overflow-tooltip />
        <el-table-column prop="sourceName" label="来源" min-width="150">
          <template #default="{ row }">{{ row.sourceName || '-' }}</template>
        </el-table-column>
        <el-table-column label="推送时间" width="160">
          <template #default="{ row }">{{ formatTime(row.createdAt) }}</template>
        </el-table-column>
        <el-table-column label="接收时间" width="160">
          <template #default="{ row }">{{ formatTime(row.receivedAt) }}</template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 'received' ? 'success' : row.status === 'pending' ? 'warning' : 'info'">
              {{ row.status === 'received' ? '已接收' : row.status === 'pending' ? '待接收' : '已关闭' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="140" fixed="right">
          <template #default="{ row }">
            <template v-if="row.status === 'pending'">
              <el-button link type="success" @click="receive(row)">接收</el-button>
              <el-button link type="warning" @click="close(row)">关闭</el-button>
            </template>
          </template>
        </el-table-column>
        <template #empty>暂无数据</template>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '@/api/client'

interface Row {
  id: number; title: string; sourceName: string; status: string
  receivedAt: string | null; createdAt: string
}

const tab = ref('pending')
const list = ref<Row[]>([])
const loading = ref(false)

async function load() {
  loading.value = true
  try {
    const res = await api.get('/api/v1/received-decorations', { params: { status: tab.value } })
    if (res.data.code === 0 || res.data.code === 200) list.value = res.data.data?.list ?? []
  } finally {
    loading.value = false
  }
}

function onTab() { load() }

function formatTime(t: string | null) {
  return t ? new Date(t).toLocaleString('zh-CN', { hour12: false }) : '-'
}

async function receive(row: Row) {
  await ElMessageBox.confirm(`确认接收装修方案「${row.title}」？接收后将应用到商城。`, '确认接收', { type: 'warning' })
  const res = await api.put(`/api/v1/received-decorations/${row.id}/receive`)
  if (res.data.code === 0 || res.data.code === 200) {
    ElMessage.success('已接收并应用')
    load()
  } else {
    ElMessage.error(res.data.message || '接收失败')
  }
}

async function close(row: Row) {
  await ElMessageBox.confirm(`确定关闭装修方案「${row.title}」？`, '提示', { type: 'warning' })
  const res = await api.put(`/api/v1/received-decorations/${row.id}/close`)
  if (res.data.code === 0 || res.data.code === 200) {
    ElMessage.success('已关闭')
    load()
  }
}

async function receiveAll() {
  await ElMessageBox.confirm(`确定一键接收全部 ${list.value.length} 个待接收装修方案？`, '一键接收', { type: 'warning' })
  const res = await api.post('/api/v1/received-decorations/receive-all')
  if (res.data.code === 0 || res.data.code === 200) {
    ElMessage.success(`已接收 ${res.data.data?.received ?? 0} 个方案`)
    load()
  } else {
    ElMessage.error(res.data.message || '接收失败')
  }
}

onMounted(load)
</script>

<style scoped>
.page { display: flex; flex-direction: column; gap: 12px; }
.page-header { display: flex; justify-content: space-between; align-items: center; }
.page-title { font-size: 18px; font-weight: 600; margin: 0; }
</style>
