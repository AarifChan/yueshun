<template>
  <div class="page">
    <el-card class="no-print">
      <template #header>
        <div class="card-header">
          <span>条码打印</span>
          <div>
            <el-button @click="selected = []">清空选择</el-button>
            <el-button type="primary" :disabled="!selected.length" @click="doPrint">打印标签（{{ selected.length }}）</el-button>
          </div>
        </div>
      </template>
      <el-form :model="searchForm" inline class="search-form">
        <el-form-item label="关键词"><el-input v-model="searchForm.keyword" placeholder="名称/编码/条码" clearable /></el-form-item>
        <el-form-item><el-button type="primary" @click="handleSearch">搜索</el-button><el-button @click="handleReset">重置</el-button></el-form-item>
      </el-form>
      <el-table :data="list" v-loading="loading" border stripe @selection-change="(rows: any[]) => (selected = rows)">
        <el-table-column type="selection" width="45" :selectable="(row: any) => !!codeOf(row)" />
        <el-table-column prop="name" label="商品名称" min-width="160" />
        <el-table-column prop="code" label="编码" min-width="110" />
        <el-table-column prop="barcode" label="条码" min-width="130">
          <template #default="{ row }">{{ row.barcode || '-' }}</template>
        </el-table-column>
        <el-table-column label="打印码" min-width="130">
          <template #default="{ row }">{{ codeOf(row) || '无可用条码' }}</template>
        </el-table-column>
        <el-table-column prop="retailPrice" label="零售价" width="100" align="right" />
      </el-table>
      <el-pagination v-model:current-page="pagination.page" v-model:page-size="pagination.pageSize" :total="total" :page-sizes="[10,20,50]" layout="total, sizes, prev, pager, next" @size-change="handleSizeChange" @current-change="handleCurrentChange" class="pagination" />
    </el-card>

    <div v-if="selected.length" class="preview no-print">
      <h3>标签预览</h3>
      <div class="label-grid">
        <div v-for="item in selected" :key="item.id" class="label-card">
          <p class="label-name">{{ item.name }}</p>
          <canvas :ref="(el) => setCanvas(el, item.id)" class="label-canvas" />
          <p class="label-code">{{ codeOf(item) }}</p>
          <p v-if="item.retailPrice" class="label-price">¥{{ Number(item.retailPrice).toFixed(2) }}</p>
        </div>
      </div>
    </div>

    <!-- 打印区域：仅打印时可见 -->
    <div v-if="selected.length" class="print-area">
      <div v-for="item in selected" :key="item.id" class="print-label">
        <p class="label-name">{{ item.name }}</p>
        <canvas :ref="(el) => setPrintCanvas(el, item.id)" />
        <p class="label-code">{{ codeOf(item) }}</p>
        <p v-if="item.retailPrice" class="label-price">¥{{ Number(item.retailPrice).toFixed(2) }}</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, nextTick, watch } from 'vue'
import { useCrud } from '@/composables/useCrud'
import { drawCode39, isCode39Supported, sanitizeCode39 } from '@/utils/barcode'

interface Product { id: number; name: string; code: string; barcode: string; retailPrice: number }

const crud = useCrud<Product>({ baseUrl: '/api/v1/products' })
const { list, total, loading, searchForm, pagination, fetchList, handleSearch, handleReset, handleSizeChange, handleCurrentChange } = crud
fetchList()

const selected = ref<Product[]>([])
const previewCanvases = new Map<number, HTMLCanvasElement>()
const printCanvases = new Map<number, HTMLCanvasElement>()

function codeOf(item: Product): string {
  return item.barcode || item.code || ''
}

function setCanvas(el: unknown, id: number) {
  if (el) previewCanvases.set(id, el as HTMLCanvasElement)
  else previewCanvases.delete(id)
}

function setPrintCanvas(el: unknown, id: number) {
  if (el) printCanvases.set(id, el as HTMLCanvasElement)
  else printCanvases.delete(id)
}

function paint() {
  for (const item of selected.value) {
    const code = codeOf(item)
    const warn = isCode39Supported(code)
    const text = warn ? code : sanitizeCode39(code)
    const preview = previewCanvases.get(item.id)
    if (preview) drawCode39(preview, text, { height: 56, narrow: 1.5 })
    const printEl = printCanvases.get(item.id)
    if (printEl) drawCode39(printEl, text, { height: 70, narrow: 2 })
  }
}

async function doPrint() {
  await nextTick()
  paint()
  await nextTick()
  window.print()
}

// 选中变化时重绘预览
watch(selected, async () => {
  await nextTick()
  paint()
}, { deep: true })
</script>

<style scoped>
.page { padding: 20px; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
.search-form { margin-bottom: 16px; }
.pagination { margin-top: 16px; justify-content: flex-end; }
.preview { margin-top: 16px; }
.preview h3 { margin: 0 0 12px; font-size: 15px; color: #303133; }
.label-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}
.label-card {
  width: 240px;
  padding: 12px;
  background: #fff;
  border: 1px dashed #dcdfe6;
  border-radius: 6px;
  text-align: center;
}
.label-name {
  margin: 0 0 8px;
  font-size: 14px;
  font-weight: 600;
  color: #303133;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.label-canvas { max-width: 100%; }
.label-code { margin: 6px 0 0; font-size: 13px; letter-spacing: 1px; color: #606266; }
.label-price { margin: 2px 0 0; font-size: 13px; color: #f56c6c; }
.print-area { display: none; }
.print-label {
  width: 300px;
  padding: 10px;
  text-align: center;
  page-break-inside: avoid;
}
@media print {
  .no-print { display: none !important; }
  .print-area {
    display: flex !important;
    flex-wrap: wrap;
    gap: 8px;
  }
  .page { padding: 0; }
}
</style>
