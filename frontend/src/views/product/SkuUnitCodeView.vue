<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">规格单位条码</h2>
      <el-button @click="exportCsv('规格单位条码', exportCols)">导出</el-button>
    </div>
    <el-card>
      <el-form inline @submit.prevent>
        <el-form-item label="商品">
          <el-input v-model="filters.keyword" placeholder="名称/编号/条码" clearable style="width: 220px" @keyup.enter="search" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="search">查询</el-button>
          <el-button @click="reset">重置</el-button>
        </el-form-item>
      </el-form>
      <el-table :data="list" v-loading="loading" border stripe>
        <el-table-column prop="code" label="商品编号" width="110" />
        <el-table-column prop="name" label="商品名称" min-width="180" show-overflow-tooltip />
        <el-table-column prop="spec" label="规格" width="120" />
        <el-table-column prop="unit" label="基本单位" width="90" />
        <el-table-column label="条码" min-width="180">
          <template #default="{ row }">
            <div v-if="editId === row.id" class="barcode-edit">
              <el-input v-model="editBarcode" placeholder="请输入条码" style="width: 160px" />
              <el-button link type="success" @click="saveBarcode(row)">保存</el-button>
              <el-button link @click="editId = 0">取消</el-button>
            </div>
            <div v-else class="barcode-view">
              <span :class="{ empty: !row.barcode }">{{ row.barcode || '未设置' }}</span>
              <el-button link type="primary" @click="startEdit(row)">编辑</el-button>
            </div>
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
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import api from '@/api/client'
import { useReport } from '@/composables/useReport'

interface Row {
  id: number; code: string; barcode: string; name: string; spec: string; unit: string
}

const { list, total, loading, page, pageSize, filters, load, search, reset, exportCsv } =
  useReport<Row>('/api/v1/product-barcodes', { keyword: '' })

const exportCols = [
  { key: 'code', label: '商品编号' },
  { key: 'name', label: '商品名称' },
  { key: 'spec', label: '规格' },
  { key: 'unit', label: '基本单位' },
  { key: 'barcode', label: '条码' },
]

const editId = ref(0)
const editBarcode = ref('')

function startEdit(row: Row) {
  editId.value = row.id
  editBarcode.value = row.barcode || ''
}

async function saveBarcode(row: Row) {
  const res = await api.put(`/api/v1/product-barcodes/${row.id}`, { barcode: editBarcode.value })
  if (res.data.code === 0 || res.data.code === 200) {
    ElMessage.success('已保存')
    editId.value = 0
    load()
  } else {
    ElMessage.error(res.data.message || '保存失败')
  }
}

onMounted(() => load(1))
</script>

<style scoped>
.page { display: flex; flex-direction: column; gap: 12px; }
.page-header { display: flex; justify-content: space-between; align-items: center; }
.page-title { font-size: 18px; font-weight: 600; margin: 0; }
.barcode-view, .barcode-edit { display: flex; align-items: center; gap: 8px; }
.empty { color: #909399; }
</style>
