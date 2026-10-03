<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">采购价格跟踪</h2>
      <div class="actions">
        <el-button @click="importVisible = true">导入</el-button>
        <el-button type="primary" @click="openCreate">新增价格跟踪</el-button>
        <el-button :disabled="!selectedRows.length" @click="exportSelectedCsv('采购价格跟踪', exportCols, selectedRows)">导出选中</el-button>
        <el-button @click="exportCsv('采购价格跟踪', exportCols)">导出</el-button>
      </div>
    </div>
    <el-card>
      <el-form inline @submit.prevent>
        <el-form-item label="日期">
          <el-date-picker v-model="range" type="daterange" value-format="YYYY-MM-DD" start-placeholder="开始日期" end-placeholder="结束日期" style="width: 260px" />
        </el-form-item>
        <el-form-item label="分类">
          <el-tree-select
            v-model="filters.categoryId"
            :data="categoryTree"
            :props="{ label: 'name', children: 'children' }"
            node-key="id"
            check-strictly
            clearable
            placeholder="商品分类"
            style="width: 160px"
          />
        </el-form-item>
        <el-form-item label="商品">
          <el-input v-model="filters.keyword" placeholder="商品名称/编号/条码/规格/关键字" clearable style="width: 220px" />
        </el-form-item>
        <el-form-item label="商品单位">
          <el-select v-model="filters.unit" placeholder="商品单位" clearable style="width: 120px">
            <el-option v-for="u in unitOptions" :key="u.id" :label="u.name" :value="u.name" />
          </el-select>
        </el-form-item>
        <el-form-item label="供应商">
          <el-input v-model="filters.supplierKeyword" placeholder="供应商名称" clearable style="width: 150px" />
        </el-form-item>
        <el-form-item label="商品状态">
          <el-select v-model="filters.status" placeholder="商品状态" clearable style="width: 110px">
            <el-option label="已启用" value="1" />
            <el-option label="已禁用" value="0" />
          </el-select>
        </el-form-item>
        <el-form-item label="品牌">
          <el-select v-model="filters.brandId" placeholder="品牌" clearable style="width: 130px">
            <el-option v-for="b in brandOptions" :key="b.id" :label="b.name" :value="b.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="商品标签">
          <el-select v-model="filters.tagId" placeholder="商品标签" clearable style="width: 130px">
            <el-option v-for="t in tagOptions" :key="t.id" :label="t.name" :value="t.id" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="search">查询</el-button>
          <el-button @click="reset">重置</el-button>
        </el-form-item>
      </el-form>
      <el-table :data="list" v-loading="loading" border stripe @selection-change="(rows: Row[]) => (selectedRows = rows)">
        <el-table-column type="selection" width="45" />
        <el-table-column prop="productCode" label="商品编号" width="110" />
        <el-table-column prop="product" label="商品名称" min-width="160" show-overflow-tooltip />
        <el-table-column prop="spec" label="规格" width="110" />
        <el-table-column prop="unit" label="单位" width="70" />
        <el-table-column prop="supplier" label="最近采购供应商" min-width="150" show-overflow-tooltip />
        <el-table-column prop="price" label="最近采购价" width="110" align="right">
          <template #default="{ row }">{{ row.price?.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column prop="quantity" label="数量" width="90" align="right" />
        <el-table-column prop="lastDate" label="最近采购日期" width="120" />
        <el-table-column label="操作" width="80" fixed="right">
          <template #default="{ row }">
            <el-button v-if="row.source === 'manual'" link type="danger" @click="removeTrack(row)">删除</el-button>
            <span v-else>-</span>
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

    <el-dialog v-model="createVisible" title="新增价格跟踪" width="520px" :close-on-click-modal="false">
      <el-form :model="createForm" label-width="90px">
        <el-form-item label="供应商" required>
          <el-select v-model="createForm.supplierId" filterable remote :remote-method="searchSuppliers" style="width: 100%" placeholder="输入供应商名称搜索">
            <el-option v-for="s in supplierOptions" :key="s.id" :label="s.name" :value="s.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="商品" required>
          <el-select v-model="createForm.productId" filterable remote :remote-method="searchProducts" style="width: 100%" placeholder="输入商品名称/编号搜索">
            <el-option v-for="p in productOptions" :key="p.id" :label="`${p.name}（${p.code || '无编号'}）`" :value="p.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="采购价">
          <el-input-number v-model="createForm.price" :min="0" :precision="2" style="width: 100%" />
        </el-form-item>
        <el-form-item label="数量">
          <el-input-number v-model="createForm.quantity" :min="0" :precision="2" style="width: 100%" />
        </el-form-item>
        <el-form-item label="日期">
          <el-date-picker v-model="createForm.trackDate" type="date" value-format="YYYY-MM-DD" style="width: 100%" />
        </el-form-item>
        <el-form-item label="备注"><el-input v-model="createForm.remark" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createVisible = false">取消</el-button>
        <el-button type="primary" :loading="creating" @click="submitCreate">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="importVisible" title="采购价格跟踪导入" width="520px" :close-on-click-modal="false" @closed="resetImport">
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
            <div class="el-upload__tip">支持 .xlsx / .xls；表头需包含「供应商名称、商品编号/商品名称」，可选「采购价、数量、日期、备注」</div>
          </template>
        </el-upload>
      </template>
      <template v-else>
        <div class="import-summary">
          <el-tag type="success">新增 {{ importResult.created }}</el-tag>
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
import { computed, onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { UploadFilled } from '@element-plus/icons-vue'
import type { UploadFile } from 'element-plus'
import api from '@/api/client'
import { useReport } from '@/composables/useReport'
import { fetchCategoryTree, type CategoryNode } from '@/api/category'
import { fetchBrandOptions } from '@/api/productDetail'
import { fetchTags, type ProductTag } from '@/api/tag'
import { fetchUnits, type GoodsUnit } from '@/api/unit'
import { ensureXlsxFile } from '@/utils/spreadsheet'

interface Row {
  id: number; source: string
  productId: number; product: string; productCode: string; barcode: string; spec: string; unit: string
  supplierId: number; supplier: string; price: number; quantity: number; lastDate: string
}

const { list, total, loading, page, pageSize, filters, load, search, reset, exportCsv, exportSelectedCsv } =
  useReport<Row>('/api/v1/product-reports/purchase-price-track', {
    startDate: '', endDate: '', keyword: '', supplierKeyword: '',
    categoryId: undefined, brandId: undefined, tagId: undefined, unit: '', status: '',
  })

const selectedRows = ref<Row[]>([])

const range = computed({
  get: () => (filters.value.startDate && filters.value.endDate ? [filters.value.startDate, filters.value.endDate] : null),
  set: (v) => {
    filters.value.startDate = v?.[0] ?? ''
    filters.value.endDate = v?.[1] ?? ''
  },
})

const exportCols = [
  { key: 'productCode', label: '商品编号' },
  { key: 'product', label: '商品名称' },
  { key: 'spec', label: '规格' },
  { key: 'unit', label: '单位' },
  { key: 'supplier', label: '最近采购供应商' },
  { key: 'price', label: '最近采购价' },
  { key: 'quantity', label: '数量' },
  { key: 'lastDate', label: '最近采购日期' },
]

const categoryTree = ref<CategoryNode[]>([])
const brandOptions = ref<{ id: number; name: string }[]>([])
const tagOptions = ref<ProductTag[]>([])
const unitOptions = ref<GoodsUnit[]>([])

function ok(res: any) {
  return res.data.code === 0 || res.data.code === 200
}

async function loadFilterOptions() {
  try {
    const res = await fetchCategoryTree()
    if (ok(res)) categoryTree.value = Array.isArray(res.data.data) ? res.data.data : []
  } catch { /* 选项加载失败不阻塞列表 */ }
  try {
    const res = await fetchBrandOptions()
    if (ok(res)) brandOptions.value = Array.isArray(res.data.data) ? res.data.data : (res.data.data?.list ?? [])
  } catch { /* 同上 */ }
  try {
    const res = await fetchTags()
    if (ok(res)) tagOptions.value = Array.isArray(res.data.data) ? res.data.data : (res.data.data?.list ?? [])
  } catch { /* 同上 */ }
  try {
    const res = await fetchUnits({})
    if (ok(res)) unitOptions.value = res.data.data?.list ?? []
  } catch { /* 同上 */ }
}

// ==================== 新增价格跟踪 ====================

const createVisible = ref(false)
const creating = ref(false)
const createForm = ref({
  supplierId: undefined as number | undefined,
  productId: undefined as number | undefined,
  price: 0,
  quantity: 1,
  trackDate: '',
  remark: '',
})
const supplierOptions = ref<{ id: number; name: string }[]>([])
const productOptions = ref<{ id: number; name: string; code?: string }[]>([])

async function searchSuppliers(kw: string) {
  const res = await api.get('/api/v1/suppliers', { params: { page: 1, pageSize: 50, keyword: kw } })
  if (ok(res)) supplierOptions.value = res.data.data?.list ?? []
}

async function searchProducts(kw: string) {
  const res = await api.get('/api/v1/products', { params: { page: 1, pageSize: 50, keyword: kw } })
  if (ok(res)) productOptions.value = res.data.data?.list ?? []
}

function openCreate() {
  createForm.value = { supplierId: undefined, productId: undefined, price: 0, quantity: 1, trackDate: '', remark: '' }
  searchSuppliers('')
  searchProducts('')
  createVisible.value = true
}

async function submitCreate() {
  if (!createForm.value.supplierId || !createForm.value.productId) {
    ElMessage.warning('请选择供应商和商品')
    return
  }
  creating.value = true
  try {
    const res = await api.post('/api/v1/product-reports/purchase-price-track', createForm.value)
    if (ok(res)) {
      ElMessage.success('已保存')
      createVisible.value = false
      load(1)
    } else {
      ElMessage.error(res.data.message || '保存失败')
    }
  } finally {
    creating.value = false
  }
}

async function removeTrack(row: Row) {
  try {
    await ElMessageBox.confirm(`确定删除「${row.supplier} - ${row.product}」的这条价格跟踪？`, '提示', { type: 'warning' })
  } catch {
    return
  }
  const res = await api.delete(`/api/v1/product-reports/purchase-price-track/${row.id}`)
  if (ok(res)) {
    ElMessage.success('已删除')
    load()
  } else {
    ElMessage.error(res.data.message || '删除失败')
  }
}

// ==================== 导入 ====================

interface ImportResult {
  created: number; updated: number; failed: number
  errors: { row: number; code: string; reason: string }[]
}

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
    const res = await api.post('/api/v1/product-reports/purchase-price-track/import', fd)
    if (ok(res)) {
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

onMounted(() => {
  load(1)
  loadFilterOptions()
})
</script>

<style scoped>
.page { display: flex; flex-direction: column; gap: 12px; }
.page-header { display: flex; justify-content: space-between; align-items: center; }
.page-title { font-size: 18px; font-weight: 600; margin: 0; }
.actions { display: flex; gap: 8px; }
.import-summary { display: flex; gap: 12px; margin-bottom: 12px; }
</style>
