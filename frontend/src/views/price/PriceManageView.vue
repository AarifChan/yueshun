<template>
  <div class="page">
    <div class="page-header">
      <span class="page-title">商品物价管理</span>
      <div>
        <el-button @click="openSettings">设 置</el-button>
        <el-button type="primary" @click="importVisible = true">批量导入</el-button>
      </div>
    </div>
    <el-alert type="info" :closable="false" class="tip"
      title="同时存在多个订货价时，价格优先级为：促销活动价＞跟踪价＞级别价" />

    <div class="main">
      <div class="category-panel">
        <div class="category-header">
          <span>商品分类</span>
          <el-link type="primary" :underline="false" @click="router.push('/categories')">编辑</el-link>
        </div>
        <el-tree :data="categoryTree" :props="{ label: 'name', children: 'children' }" node-key="id"
          highlight-current :expand-on-click-node="false" @node-click="handleCategoryClick" />
      </div>

      <div class="content">
        <div class="filter-row">
          <el-select v-model="filters.priceType" style="width: 130px">
            <el-option label="订货价" value="default" />
            <el-option label="批发价" value="wholesale" />
            <el-option label="零售价" value="retail" />
          </el-select>
          <el-select v-model="filters.priceOp" style="width: 100px">
            <el-option label="大于" value="gt" />
            <el-option label="小于" value="lt" />
          </el-select>
          <el-input v-model="filters.priceValue" placeholder="价格" style="width: 140px" clearable />
        </div>
        <div class="filter-row">
          <el-input v-model="filters.keyword" placeholder="商品名称/编号/条码/关键字" style="width: 240px" clearable />
          <el-select v-model="filters.status" placeholder="上下架状态" style="width: 130px" clearable>
            <el-option label="上架" :value="1" />
            <el-option label="下架" :value="0" />
          </el-select>
          <el-select v-model="filters.stockStatus" placeholder="库存状态" style="width: 130px" clearable>
            <el-option label="有库存" value="in" />
            <el-option label="无库存" value="out" />
          </el-select>
          <el-button type="primary" @click="handleSearch">查 询</el-button>
          <el-link type="primary" :underline="false" @click="handleReset">清空</el-link>
        </div>

        <el-table :data="list" v-loading="loading" border stripe>
          <el-table-column type="index" label="序号" width="70" :index="indexMethod" />
          <el-table-column type="selection" width="45" />
          <el-table-column label="操作" width="70">
            <template #default="{ row }">
              <el-dropdown trigger="click" @command="(cmd: string) => cmd === 'edit' && openEditPrice(row)">
                <el-button text size="small">···</el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item command="edit">修改价格</el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </template>
          </el-table-column>
          <el-table-column label="商品名称" min-width="180">
            <template #default="{ row }">
              <el-link type="primary" :underline="false" @click="router.push(`/products/${row.id}`)">{{ row.name }}</el-link>
            </template>
          </el-table-column>
          <el-table-column label="默认订货价" width="120">
            <template #default="{ row }">{{ formatPrice(row.defaultPrice) }}/{{ row.unit }}</template>
          </el-table-column>
          <el-table-column label="批发价" width="120">
            <template #default="{ row }">{{ formatPrice(row.wholesalePrice) }}/{{ row.unit }}</template>
          </el-table-column>
          <el-table-column label="零售价" width="120">
            <template #default="{ row }">{{ formatPrice(row.retailPrice) }}/{{ row.unit }}</template>
          </el-table-column>
        </el-table>

        <el-pagination v-model:current-page="pagination.page" v-model:page-size="pagination.pageSize" :total="total"
          :page-sizes="[30, 50, 100]" layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange" @current-change="handleCurrentChange" class="pagination" />
      </div>
    </div>

    <el-dialog v-model="editVisible" title="修改价格" width="420px">
      <el-form label-width="90px">
        <el-form-item label="商品名称"><span>{{ editRow?.name }}</span></el-form-item>
        <el-form-item label="默认订货价"><el-input-number v-model="editForm.defaultPrice" :min="0" :precision="2" /></el-form-item>
        <el-form-item label="批发价"><el-input-number v-model="editForm.wholesalePrice" :min="0" :precision="2" /></el-form-item>
        <el-form-item label="零售价"><el-input-number v-model="editForm.retailPrice" :min="0" :precision="2" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSavePrice">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="settingsVisible" title="设置" width="520px">
      <div class="settings-group">
        <div class="settings-title">价格换算设置</div>
        <el-checkbox v-model="settingsForm.autoBaseToAux">修改基本单位单价时自动换算辅助单位单价</el-checkbox>
        <el-checkbox v-model="settingsForm.autoAuxToOther">修改辅助单位单价时自动换算其他单位单价</el-checkbox>
      </div>
      <el-divider />
      <div class="settings-group">
        <el-checkbox v-model="settingsForm.salesRangeEnabled" class="settings-title">销量统计区间设置</el-checkbox>
        <div class="settings-body">
          <el-radio-group v-model="settingsForm.salesRangeType" @change="handleRangeTypeChange">
            <el-radio value="3m">近三个月</el-radio>
            <el-radio value="6m">近半年</el-radio>
            <el-radio value="1y">近一年</el-radio>
            <el-radio value="custom">自定义</el-radio>
          </el-radio-group>
          <el-date-picker v-model="settingsForm.salesRangeDate" type="daterange" value-format="YYYY-MM-DD"
            range-separator="至" start-placeholder="开始日期" end-placeholder="结束日期" class="date-range" />
        </div>
      </div>
      <template #footer>
        <el-button @click="settingsVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSaveSettings">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="importVisible" title="批量导入" width="520px" @closed="resetImport">
      <el-alert type="info" :closable="false" class="tip"
        title="请上传 .xlsx 文件，表头依次为：商品编号、批发价、零售价、默认订货价" />
      <el-upload drag :show-file-list="false" accept=".xlsx" :http-request="handleImportUpload">
        <el-icon class="el-icon--upload"><UploadFilled /></el-icon>
        <div class="el-upload__text">将文件拖到此处，或<em>点击上传</em></div>
      </el-upload>
      <div v-if="importResult" class="import-result">
        <el-alert :type="importResult.failed > 0 ? 'warning' : 'success'" :closable="false"
          :title="`导入完成：成功 ${importResult.updated} 条，失败 ${importResult.failed} 条`" />
        <el-table v-if="importResult.errors.length" :data="importResult.errors" size="small" border class="error-table">
          <el-table-column prop="row" label="行号" width="70" />
          <el-table-column prop="code" label="商品编号" width="120" />
          <el-table-column prop="reason" label="失败原因" />
        </el-table>
      </div>
      <template #footer>
        <el-button @click="importVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { UploadFilled } from '@element-plus/icons-vue'
import { fetchCategoryTree, type CategoryNode } from '@/api/category'
import {
  fetchPriceManageList, updatePriceManage, importPriceManage, fetchSettings, saveSettings,
  type PriceManageItem, type PriceImportResult,
} from '@/api/priceManage'

const router = useRouter()

const list = ref<PriceManageItem[]>([])
const total = ref(0)
const loading = ref(false)
const pagination = ref({ page: 1, pageSize: 30 })
const filters = ref({
  priceType: 'default',
  priceOp: 'gt',
  priceValue: '',
  keyword: '',
  status: undefined as number | undefined,
  stockStatus: '',
  categoryId: undefined as number | undefined,
})
const categoryTree = ref<CategoryNode[]>([])

function formatPrice(v: number) {
  return Number(v || 0).toString()
}

function indexMethod(index: number) {
  return (pagination.value.page - 1) * pagination.value.pageSize + index + 1
}

async function fetchList() {
  loading.value = true
  try {
    const res = await fetchPriceManageList({
      page: pagination.value.page,
      pageSize: pagination.value.pageSize,
      keyword: filters.value.keyword || undefined,
      categoryId: filters.value.categoryId,
      status: filters.value.status,
      stockStatus: filters.value.stockStatus || undefined,
      priceType: filters.value.priceType,
      priceOp: filters.value.priceOp,
      priceValue: filters.value.priceValue || undefined,
    })
    if (res.data.code === 0 || res.data.code === 200) {
      list.value = res.data.data?.list ?? []
      total.value = res.data.data?.total ?? 0
    }
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  pagination.value.page = 1
  fetchList()
}

function handleReset() {
  filters.value = { priceType: 'default', priceOp: 'gt', priceValue: '', keyword: '', status: undefined, stockStatus: '', categoryId: undefined }
  pagination.value.page = 1
  fetchList()
}

function handleSizeChange(size: number) {
  pagination.value.pageSize = size
  pagination.value.page = 1
  fetchList()
}

function handleCurrentChange(page: number) {
  pagination.value.page = page
  fetchList()
}

function handleCategoryClick(node: CategoryNode) {
  filters.value.categoryId = node.id
  pagination.value.page = 1
  fetchList()
}

async function loadCategoryTree() {
  const res = await fetchCategoryTree()
  if (res.data.code === 0 || res.data.code === 200) {
    categoryTree.value = res.data.data ?? []
  }
}

const editVisible = ref(false)
const editRow = ref<PriceManageItem | null>(null)
const editForm = ref({ defaultPrice: 0, wholesalePrice: 0, retailPrice: 0 })

function openEditPrice(row: PriceManageItem) {
  editRow.value = row
  editForm.value = { defaultPrice: row.defaultPrice, wholesalePrice: row.wholesalePrice, retailPrice: row.retailPrice }
  editVisible.value = true
}

async function handleSavePrice() {
  if (!editRow.value) return
  const res = await updatePriceManage(editRow.value.id, { ...editForm.value })
  if (res.data.code === 0 || res.data.code === 200) {
    ElMessage.success('价格更新成功')
    editVisible.value = false
    fetchList()
  } else {
    ElMessage.error(res.data.message || '更新失败')
  }
}

const settingsVisible = ref(false)
const settingsForm = ref({
  autoBaseToAux: false,
  autoAuxToOther: false,
  salesRangeEnabled: false,
  salesRangeType: '3m',
  salesRangeDate: [] as string[],
})

async function openSettings() {
  settingsVisible.value = true
  const res = await fetchSettings()
  if (res.data.code === 0 || res.data.code === 200) {
    const data = res.data.data || {}
    settingsForm.value.autoBaseToAux = data['priceManage.autoBaseToAux'] === '1'
    settingsForm.value.autoAuxToOther = data['priceManage.autoAuxToOther'] === '1'
    settingsForm.value.salesRangeEnabled = data['priceManage.salesRange.enabled'] === '1'
    settingsForm.value.salesRangeType = data['priceManage.salesRange.type'] || '3m'
    const start = data['priceManage.salesRange.start']
    const end = data['priceManage.salesRange.end']
    settingsForm.value.salesRangeDate = start && end ? [start, end] : []
  }
}

function formatDate(d: Date) {
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${d.getFullYear()}-${m}-${day}`
}

function handleRangeTypeChange(type: string) {
  const monthsMap: Record<string, number> = { '3m': 3, '6m': 6, '1y': 12 }
  const months = monthsMap[type]
  if (!months) return
  const end = new Date()
  const start = new Date()
  start.setMonth(start.getMonth() - months)
  settingsForm.value.salesRangeDate = [formatDate(start), formatDate(end)]
}

async function handleSaveSettings() {
  const [start, end] = settingsForm.value.salesRangeDate || []
  const res = await saveSettings({
    'priceManage.autoBaseToAux': settingsForm.value.autoBaseToAux ? '1' : '0',
    'priceManage.autoAuxToOther': settingsForm.value.autoAuxToOther ? '1' : '0',
    'priceManage.salesRange.enabled': settingsForm.value.salesRangeEnabled ? '1' : '0',
    'priceManage.salesRange.type': settingsForm.value.salesRangeType,
    'priceManage.salesRange.start': start || '',
    'priceManage.salesRange.end': end || '',
  })
  if (res.data.code === 0 || res.data.code === 200) {
    ElMessage.success('保存成功')
    settingsVisible.value = false
  } else {
    ElMessage.error(res.data.message || '保存失败')
  }
}

const importVisible = ref(false)
const importResult = ref<PriceImportResult | null>(null)

function resetImport() {
  importResult.value = null
}

async function handleImportUpload(options: any) {
  const res = await importPriceManage(options.file as File)
  if (res.data.code === 0 || res.data.code === 200) {
    importResult.value = res.data.data
    fetchList()
  } else {
    ElMessage.error(res.data.message || '导入失败')
  }
}

onMounted(() => {
  loadCategoryTree()
  fetchList()
})
</script>

<style scoped>
.page { padding: 20px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; }
.page-title { font-size: 18px; font-weight: 600; }
.tip { margin-bottom: 12px; }
.main { display: flex; gap: 16px; }
.category-panel { width: 180px; flex-shrink: 0; border: 1px solid var(--el-border-color); border-radius: 4px; padding: 8px; }
.category-header { display: flex; justify-content: space-between; align-items: center; padding: 4px 8px 8px; font-weight: 600; border-bottom: 1px solid var(--el-border-color-lighter); margin-bottom: 8px; }
.content { flex: 1; min-width: 0; }
.filter-row { display: flex; gap: 8px; align-items: center; margin-bottom: 12px; }
.pagination { margin-top: 16px; justify-content: flex-end; }
.settings-group { margin-bottom: 8px; }
.settings-title { font-weight: 600; display: block; margin-bottom: 8px; }
.settings-group .el-checkbox { display: flex; }
.settings-body { margin-top: 8px; }
.date-range { margin-top: 12px; width: 100%; }
.import-result { margin-top: 16px; }
.error-table { margin-top: 12px; }
</style>
