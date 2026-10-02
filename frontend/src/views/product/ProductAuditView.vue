<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">商品审核</h2>
    </div>
    <el-card>
      <el-form inline @submit.prevent>
        <el-form-item label="审核状态">
          <el-radio-group v-model="filters.status" @change="search">
            <el-radio-button value="pending">待审核</el-radio-button>
            <el-radio-button value="approved">已审核</el-radio-button>
            <el-radio-button value="all">全部</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="商品">
          <el-input v-model="filters.keyword" placeholder="名称/编号/条码" clearable style="width: 200px" @keyup.enter="search" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="search">查询</el-button>
        </el-form-item>
      </el-form>
      <el-table :data="list" v-loading="loading" border stripe>
        <el-table-column prop="code" label="商品编号" width="110" />
        <el-table-column prop="barcode" label="条码" width="130" />
        <el-table-column prop="name" label="商品名称" min-width="180" show-overflow-tooltip />
        <el-table-column prop="spec" label="规格" width="110" />
        <el-table-column prop="unit" label="单位" width="70" />
        <el-table-column prop="category" label="分类" width="120" />
        <el-table-column label="审核状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.auditStatus === 'approved' ? 'success' : 'warning'">
              {{ row.auditStatus === 'approved' ? '已审核' : '待审核' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="更新时间" width="170">
          <template #default="{ row }">{{ formatTime(row.updatedAt) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button v-if="row.auditStatus !== 'approved'" link type="success" @click="approve(row)">审核通过</el-button>
            <el-button v-else link type="warning" @click="revoke(row)">反审核</el-button>
          </template>
        </el-table-column>
        <template #empty>暂无数据</template>
      </el-table>
      <el-pagination
        v-model:current-page="page" v-model:page-size="pageSize"
        :total="total" :page-sizes="[30, 50, 100]"
        layout="total, sizes, prev, pager, next" style="margin-top: 12px; justify-content: flex-end"
        @current-change="load()" @size-change="load(1)" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '@/api/client'
import { useReport } from '@/composables/useReport'

interface Row {
  id: number; code: string; barcode: string; name: string; spec: string; unit: string
  category: string; auditStatus: string; updatedAt: string
}

const { list, total, loading, page, pageSize, filters, load, search } =
  useReport<Row>('/api/v1/product-audit', { status: 'pending', keyword: '' })

function formatTime(t: string) {
  return t ? new Date(t).toLocaleString('zh-CN', { hour12: false }) : ''
}

async function approve(row: Row) {
  await ElMessageBox.confirm(`确定审核通过商品「${row.name}」？`, '审核确认', { type: 'warning' })
  const res = await api.put(`/api/v1/product-audit/${row.id}/approve`)
  if (res.data.code === 0 || res.data.code === 200) {
    ElMessage.success('已审核通过')
    load()
  } else {
    ElMessage.error(res.data.message || '操作失败')
  }
}

async function revoke(row: Row) {
  await ElMessageBox.confirm(`确定将商品「${row.name}」反审核为待审核？`, '反审核确认', { type: 'warning' })
  const res = await api.put(`/api/v1/product-audit/${row.id}/revoke`)
  if (res.data.code === 0 || res.data.code === 200) {
    ElMessage.success('已反审核')
    load()
  } else {
    ElMessage.error(res.data.message || '操作失败')
  }
}

onMounted(() => load(1))
</script>

<style scoped>
.page { display: flex; flex-direction: column; gap: 12px; }
.page-header { display: flex; justify-content: space-between; align-items: center; }
.page-title { font-size: 18px; font-weight: 600; margin: 0; }
</style>
