<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">规格单位条码</h2>
      <div class="actions">
        <el-button :disabled="!selectedRows.length" @click="exportSelectedCsv('规格单位条码', exportCols, selectedRows)">导出选中</el-button>
        <el-button @click="exportCsv('规格单位条码', exportCols)">导出</el-button>
        <el-button type="primary" @click="importVisible = true">条码导入</el-button>
      </div>
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
      <el-table :data="list" v-loading="loading" border stripe @selection-change="(rows: Row[]) => (selectedRows = rows)">
        <el-table-column type="selection" width="45" />
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

    <el-dialog v-model="importVisible" title="条码导入" width="520px" :close-on-click-modal="false" @closed="resetImport">
      <template v-if="!importResult">
        <el-upload
          drag
          :auto-upload="false"
          :limit="1"
          accept=".xlsx,.xls"
          :on-change="handleFileChange"
          :on-remove="handleFileRemove"
          :file-list="importFileList"
        >
          <el-icon class="el-icon--upload"><UploadFilled /></el-icon>
          <div class="el-upload__text">将文件拖到此处，或<em>点击上传</em></div>
          <template #tip>
            <div class="el-upload__tip">支持 .xlsx / .xls；表头需包含「商品编号/商品名称、条码」，按编号优先匹配商品后更新条码</div>
          </template>
        </el-upload>
      </template>
      <template v-else>
        <div class="import-summary">
          <el-tag type="success">更新 {{ importResult.updated }}</el-tag>
          <el-tag type="danger">失败 {{ importResult.failed }}</el-tag>
        </div>
        <el-table v-if="importResult.errors.length" :data="importResult.errors.slice(0, 20)" border size="small" max-height="240">
          <el-table-column prop="row" label="行号" width="80" />
          <el-table-column prop="code" label="编号" width="140" />
          <el-table-column prop="reason" label="原因" min-width="180" />
        </el-table>
      </template>
      <template #footer>
        <template v-if="!importResult">
          <el-button @click="importVisible = false">取消</el-button>
          <el-button type="primary" :loading="importing" :disabled="!importFile" @click="startImport">开始导入</el-button>
        </template>
        <el-button v-else type="primary" @click="finishImport">完成</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { UploadFilled } from '@element-plus/icons-vue'
import type { UploadFile } from 'element-plus'
import api from '@/api/client'
import { useReport } from '@/composables/useReport'
import { ensureXlsxFile } from '@/utils/spreadsheet'

interface Row {
  id: number; code: string; barcode: string; name: string; spec: string; unit: string
}

interface ImportResult {
  created: number; updated: number; failed: number
  errors: { row: number; code: string; reason: string }[]
}

const { list, total, loading, page, pageSize, filters, load, search, reset, exportCsv, exportSelectedCsv } =
  useReport<Row>('/api/v1/product-barcodes', { keyword: '' })

const selectedRows = ref<Row[]>([])

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

// ==================== 条码导入 ====================

const importVisible = ref(false)
const importing = ref(false)
const importFile = ref<File | null>(null)
const importFileList = ref<UploadFile[]>([])
const importResult = ref<ImportResult | null>(null)

function handleFileChange(file: UploadFile) {
  importFileList.value = [file]
  importFile.value = file.raw ?? null
}

function handleFileRemove() {
  importFileList.value = []
  importFile.value = null
}

async function startImport() {
  if (!importFile.value) {
    ElMessage.warning('请先选择要导入的文件')
    return
  }
  importing.value = true
  try {
    const fd = new FormData()
    fd.append('file', await ensureXlsxFile(importFile.value))
    const res = await api.post('/api/v1/product-barcodes/import', fd)
    if (res.data.code === 0 || res.data.code === 200) {
      importResult.value = res.data.data
    } else {
      ElMessage.error(res.data.message || '导入失败')
    }
  } catch {
    // 拦截器已提示
  } finally {
    importing.value = false
  }
}

function finishImport() {
  importVisible.value = false
  load(1)
}

function resetImport() {
  importFile.value = null
  importFileList.value = []
  importing.value = false
  importResult.value = null
}

onMounted(() => load(1))
</script>

<style scoped>
.page { display: flex; flex-direction: column; gap: 12px; }
.page-header { display: flex; justify-content: space-between; align-items: center; }
.page-title { font-size: 18px; font-weight: 600; margin: 0; }
.actions { display: flex; gap: 8px; }
.barcode-view, .barcode-edit { display: flex; align-items: center; gap: 8px; }
.empty { color: #909399; }
.import-summary { display: flex; gap: 12px; margin-bottom: 12px; }
</style>
