<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">应用授权</h2>
      <el-button icon="Refresh" @click="load">刷新</el-button>
    </div>
    <el-card>
      <el-table :data="list" v-loading="loading" border stripe>
        <el-table-column type="index" label="序" width="60" />
        <el-table-column prop="appName" label="应用名称" min-width="140" />
        <el-table-column label="应用到期时间" width="160">
          <template #default="{ row }">{{ row.expireAt || '长期有效' }}</template>
        </el-table-column>
        <el-table-column prop="purchasedUsers" label="已购买用户数" width="130" align="center" />
        <el-table-column prop="grantedUsers" label="已授权用户数" width="130" align="center" />
        <el-table-column label="操作" width="150" fixed="right">
          <template #default>
            <el-button link type="primary" @click="goEmployeeRights">设置授权用户</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import api from '@/api/client'

interface Row { appName: string; expireAt: string; purchasedUsers: number; grantedUsers: number }

const router = useRouter()
const list = ref<Row[]>([])
const loading = ref(false)

async function load() {
  loading.value = true
  try {
    const res = await api.get('/api/v1/app-licenses')
    if (res.data.code === 0 || res.data.code === 200) list.value = res.data.data?.list ?? []
  } finally {
    loading.value = false
  }
}

function goEmployeeRights() {
  router.push('/setting/employee-rights')
}

onMounted(load)
</script>

<style scoped>
.page { display: flex; flex-direction: column; gap: 12px; }
.page-header { display: flex; justify-content: space-between; align-items: center; }
.page-title { font-size: 18px; font-weight: 600; margin: 0; }
</style>
