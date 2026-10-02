<template>
  <div class="page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>素材库</span>
          <el-upload :action="uploadAction" :headers="uploadHeaders" :show-file-list="false" :on-success="onUploaded" accept="image/*,video/*">
            <el-button type="primary">上传素材</el-button>
          </el-upload>
        </div>
      </template>
      <el-form :model="searchForm" inline class="search-form">
        <el-form-item label="名称"><el-input v-model="searchForm.keyword" placeholder="名称" clearable /></el-form-item>
        <el-form-item label="类型">
          <el-select v-model="searchForm.type" placeholder="全部" clearable style="width: 120px">
            <el-option label="图片" value="image" /><el-option label="视频" value="video" /><el-option label="附件" value="file" />
          </el-select>
        </el-form-item>
        <el-form-item><el-button type="primary" @click="handleSearch">搜索</el-button><el-button @click="handleReset">重置</el-button></el-form-item>
      </el-form>
      <el-table :data="list" v-loading="loading" border stripe>
        <el-table-column label="预览" width="100" align="center">
          <template #default="{ row }">
            <el-image v-if="row.type === 'image'" :src="row.url" fit="cover" style="width: 48px; height: 48px" :preview-src-list="[row.url]" preview-teleported />
            <el-icon v-else-if="row.type === 'video'" :size="28"><VideoPlay /></el-icon>
            <el-icon v-else :size="28"><Document /></el-icon>
          </template>
        </el-table-column>
        <el-table-column prop="name" label="素材名称" min-width="180" />
        <el-table-column label="类型" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.type === 'image' ? 'success' : row.type === 'video' ? 'warning' : 'info'">{{ typeLabel(row.type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="大小" width="110" align="right">
          <template #default="{ row }">{{ formatSize(row.size) }}</template>
        </el-table-column>
        <el-table-column prop="createdAt" label="上传时间" width="170">
          <template #default="{ row }">{{ formatTime(row.createdAt) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button type="danger" size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
        <template #empty>暂无数据</template>
      </el-table>
      <el-pagination v-model:current-page="pagination.page" v-model:page-size="pagination.pageSize" :total="total" :page-sizes="[10,20,50]" layout="total, sizes, prev, pager, next" @size-change="handleSizeChange" @current-change="handleCurrentChange" class="pagination" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { VideoPlay, Document } from '@element-plus/icons-vue'
import api from '@/api/client'
import { useCrud } from '@/composables/useCrud'
import { useAuthStore } from '@/stores/auth'

interface Material { id: number; name: string; type: string; url: string; size: number; createdAt: string }

const auth = useAuthStore()
const uploadAction = '/api/v1/upload'
const uploadHeaders = computed(() => ({ Authorization: `Bearer ${auth.token}` }))

const {
  list, total, loading, searchForm, pagination,
  fetchList, handleDelete,
  handleSizeChange, handleCurrentChange, handleSearch, handleReset,
} = useCrud<Material>({ baseUrl: '/base-data/materials' })

function typeLabel(t: string): string {
  return t === 'image' ? '图片' : t === 'video' ? '视频' : '附件'
}

function formatSize(size: number): string {
  if (!size) return '-'
  if (size < 1024) return `${size} B`
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`
  return `${(size / 1024 / 1024).toFixed(1)} MB`
}

function formatTime(t: string): string {
  return t ? t.replace('T', ' ').slice(0, 19) : '-'
}

async function onUploaded(res: { code: number; data?: { url?: string }; message?: string }, file: { name: string; raw?: File }) {
  if (res.code !== 0 && res.code !== 200) return
  const url = res.data?.url || ''
  const isVideo = /\.(mp4|mov|webm|avi)$/i.test(file.name)
  await api.post('/base-data/materials', {
    name: file.name,
    type: isVideo ? 'video' : 'image',
    url,
    size: file.raw?.size || 0,
  })
  await fetchList()
}
</script>

<style scoped>
.card-header { display: flex; justify-content: space-between; align-items: center; }
.pagination { margin-top: 16px; justify-content: flex-end; }
</style>
