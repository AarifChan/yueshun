<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">其他入库单列表</h2>
      <div class="page-actions">
        <el-button @click="importVisible = true"><el-icon><Upload /></el-icon>导入</el-button>
        <el-dropdown trigger="click" @command="handleExport">
          <el-button><el-icon><Download /></el-icon>导出<el-icon><ArrowDown /></el-icon></el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="list">导出列表</el-dropdown-item>
              <el-dropdown-item command="items">导出明细</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
        <el-button type="primary" @click="goCreate"><el-icon><Plus /></el-icon>新增入库单</el-button>
      </div>
    </div>

    <div class="date-quick-bar">
      <el-button
        v-for="p in primaryPresets"
        :key="p.key"
        :type="activePreset === p.key ? 'primary' : 'default'"
        link
        class="preset-btn"
        @click="applyPreset(p.key)"
      >{{ p.label }}</el-button>
      <el-dropdown trigger="click" @command="applyPreset">
        <el-button link class="preset-btn" :type="isMorePreset ? 'primary' : 'default'">
          更多<el-icon><ArrowDown /></el-icon>
        </el-button>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item v-for="p in morePresets" :key="p.key" :command="p.key">{{ p.label }}</el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
      <el-date-picker
        v-model="dateRange"
        type="daterange"
        range-separator="至"
        start-placeholder="开始日期"
        end-placeholder="结束日期"
        value-format="YYYY-MM-DD"
        class="date-range"
        @change="handleRangeChange"
      />
    </div>

    <div class="filter-row">
      <template v-for="key in shownFilters" :key="key">
        <el-input
          v-if="key === 'keyword'"
          v-model="filters.keyword"
          placeholder="请输入其他入库单号"
          clearable
          class="filter-item-lg"
          @keyup.enter="handleSearch"
        >
          <template #prefix><span class="filter-label">搜索</span></template>
        </el-input>
        <el-input
          v-else-if="key === 'productKw'"
          v-model="filters.productKw"
          placeholder="请输入商品名称/编号"
          clearable
          class="filter-item-lg"
          @keyup.enter="handleSearch"
        >
          <template #prefix><span class="filter-label">商品</span></template>
        </el-input>
        <RemoteSelect
          v-else-if="key === 'warehouseId'"
          v-model="filters.warehouseId"
          api-url="/api/v1/warehouses"
          placeholder="入库仓库"
          class="filter-item"
        />
        <el-select v-else-if="key === 'status'" v-model="filters.status" placeholder="单据状态" clearable class="filter-item">
          <el-option label="草稿" value="draft" />
          <el-option label="已过账" value="completed" />
        </el-select>
        <el-select v-else-if="key === 'inType'" v-model="filters.inType" placeholder="入库类型" clearable class="filter-item">
          <el-option v-for="t in OTHER_IN_TYPES" :key="t" :label="t" :value="t" />
        </el-select>
      </template>
      <div class="filter-actions">
        <el-link type="primary" :underline="false" @click="handleReset">清空</el-link>
        <el-tooltip content="筛选项设置" placement="top">
          <el-button :icon="Setting" @click="filterDialogVisible = true" />
        </el-tooltip>
        <el-button type="primary" :icon="Search" @click="handleSearch">查 询</el-button>
      </div>
    </div>

    <el-table :data="list" v-loading="loading" border stripe>
      <el-table-column type="selection" width="45" />
      <el-table-column label="操作" width="70" align="center">
        <template #default="{ row }">
          <el-dropdown trigger="click" @command="(cmd: string) => handleRowCommand(cmd, row)">
            <el-button link type="primary" class="row-more-btn"><el-icon><More /></el-icon></el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="edit">{{ row.status === 'draft' ? '编辑' : '查看' }}</el-dropdown-item>
                <el-dropdown-item v-if="row.status === 'draft'" command="complete">过账</el-dropdown-item>
                <el-dropdown-item v-if="row.status === 'draft'" command="delete">删除</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </template>
      </el-table-column>
      <el-table-column label="单号" min-width="160">
        <template #default="{ row }">
          <el-link type="primary" :underline="false" @click="goEdit(row)">{{ row.billNo }}</el-link>
        </template>
      </el-table-column>
      <el-table-column label="录单时间" width="160">
        <template #default="{ row }">{{ formatTime(row.createdAt) }}</template>
      </el-table-column>
      <el-table-column prop="warehouseName" label="入库仓库" min-width="110" show-overflow-tooltip />
      <el-table-column prop="inType" label="入库类型" width="100" />
      <el-table-column prop="counterpart" label="往来单位" min-width="120" show-overflow-tooltip>
        <template #default="{ row }">{{ row.counterpart || '-' }}</template>
      </el-table-column>
      <el-table-column prop="totalQty" label="数量合计" width="100" align="right" />
      <el-table-column label="金额合计" width="110" align="right">
        <template #default="{ row }">¥ {{ row.amount.toFixed(2) }}</template>
      </el-table-column>
      <el-table-column label="单据状态" width="95" align="center">
        <template #default="{ row }">
          <el-tag :type="row.status === 'completed' ? 'success' : 'info'" size="small">
            {{ row.status === 'completed' ? '已过账' : '草稿' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="operatorName" label="制单人" width="100">
        <template #default="{ row }">{{ row.operatorName || '-' }}</template>
      </el-table-column>
      <el-table-column prop="handlerName" label="经手人" width="100">
        <template #default="{ row }">{{ row.handlerName || '-' }}</template>
      </el-table-column>
      <el-table-column prop="businessManagerName" label="业务经理" width="100">
        <template #default="{ row }">{{ row.businessManagerName || '-' }}</template>
      </el-table-column>
      <el-table-column prop="remark" label="备注" min-width="120" show-overflow-tooltip>
        <template #default="{ row }">{{ row.remark || '-' }}</template>
      </el-table-column>
    </el-table>

    <div class="pagination-row">
      <span class="total-text">总{{ total }}条</span>
      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="total"
        :page-sizes="[20, 50, 100]"
        layout="prev, pager, next, sizes, jumper"
        background
        @current-change="fetchList"
        @size-change="handleSearch"
      />
    </div>

    <!-- 导入 -->
    <el-dialog v-model="importVisible" title="导入其他入库单" width="560px" :close-on-click-modal="false" @closed="resetImport">
      <template v-if="!importResult">
        <el-upload
          drag
          :auto-upload="false"
          :limit="1"
          accept=".xlsx,.xls"
          :on-change="handleImportFileChange"
          :on-remove="handleImportFileRemove"
          :file-list="importFileList"
        >
          <el-icon class="el-icon--upload"><UploadFilled /></el-icon>
          <div class="el-upload__text">将文件拖到此处，或<em>点击上传</em></div>
          <template #tip>
            <div class="el-upload__tip">支持 .xlsx / .xls，需包含 单据编号、商品名称、数量 等列</div>
          </template>
        </el-upload>
        <el-checkbox v-model="importApplyStock" class="apply-stock-checkbox">同时增加库存（过账单据写入库存）</el-checkbox>
      </template>
      <template v-else>
        <div class="import-summary">
          <el-tag type="success">成功新增 {{ importResult.created }} 张单</el-tag>
          <el-tag type="danger">跳过 {{ importResult.skipped }} 行</el-tag>
          <el-tag v-if="importResult.createdProducts?.length" type="warning">
            自动建档 {{ importResult.createdProducts.length }} 个商品
          </el-tag>
        </div>
        <div v-if="importResult.createdProducts?.length" class="billno-map">
          <p v-for="p in importResult.createdProducts" :key="p">新商品：{{ p }}</p>
        </div>
        <div v-if="billNoMapEntries.length" class="billno-map">
          <p v-for="[oldNo, newNo] in billNoMapEntries" :key="oldNo">单号冲突：{{ oldNo }} → {{ newNo }}</p>
        </div>
        <el-table v-if="importResult.errors && importResult.errors.length" :data="importResult.errors" border size="small" max-height="260">
          <el-table-column prop="row" label="行号" width="80" />
          <el-table-column prop="code" label="单据编号" width="160" />
          <el-table-column prop="reason" label="原因" min-width="200" />
        </el-table>
      </template>
      <template #footer>
        <template v-if="!importResult">
          <el-button @click="importVisible = false">取消</el-button>
          <el-button type="primary" :loading="importing" :disabled="!importFile" @click="startImport">开始导入</el-button>
        </template>
        <template v-else>
          <el-button type="primary" @click="importVisible = false">完成</el-button>
        </template>
      </template>
    </el-dialog>

    <!-- 筛选项设置 -->
    <el-dialog v-model="filterDialogVisible" title="筛选项设置" width="420px">
      <el-checkbox
        v-for="conf in FILTER_CONFS"
        :key="conf.key"
        :model-value="shownFilters.includes(conf.key)"
        :disabled="conf.key === 'keyword'"
        class="filter-conf-item"
        @change="toggleFilter(conf.key)"
      >{{ conf.label }}</el-checkbox>
      <template #footer>
        <el-link type="primary" :underline="false" @click="resetFilterConfs">恢复默认</el-link>
        <el-button type="primary" @click="filterDialogVisible = false">确 定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import dayjs from 'dayjs'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowDown, Download, More, Plus, Search, Setting, Upload, UploadFilled } from '@element-plus/icons-vue'
import type { UploadFile } from 'element-plus'
import RemoteSelect from '@/components/RemoteSelect.vue'
import {
  OTHER_IN_TYPES,
  completeOtherInStock,
  deleteOtherInStock,
  exportOtherInStockItems,
  exportOtherInStocks,
  fetchOtherInStocks,
  importOtherInStocks,
  type OtherInStock,
  type OtherInStockImportResult,
  type OtherInStockListQuery,
} from '@/api/otherInStock'

const router = useRouter()

const list = ref<OtherInStock[]>([])
const total = ref(0)
const loading = ref(false)
const pagination = reactive({ page: 1, pageSize: 20 })
const dateRange = ref<[string, string] | null>(null)
const activePreset = ref('')

const filters = reactive({
  keyword: '',
  productKw: '',
  warehouseId: undefined as number | undefined,
  status: '',
  inType: '',
})

// ---- 日期快捷筛选 ----
const primaryPresets = [
  { key: 'last7', label: '近7天' },
  { key: 'last15', label: '近15天' },
  { key: 'yesterday', label: '昨天' },
  { key: 'today', label: '今天' },
  { key: 'week', label: '本周' },
  { key: 'month', label: '本月' },
  { key: 'year', label: '本年' },
]
const morePresets = [
  { key: 'lastWeek', label: '上周' },
  { key: 'lastMonth', label: '上月' },
  { key: 'lastYear', label: '上年' },
  { key: 'last30', label: '近30天' },
  { key: 'last3m', label: '近三个月' },
  { key: 'last365', label: '近一年' },
]
const isMorePreset = computed(() => morePresets.some((p) => p.key === activePreset.value))

function presetRange(key: string): [string, string] | null {
  const t = dayjs()
  const fmt = (d: dayjs.Dayjs) => d.format('YYYY-MM-DD')
  switch (key) {
    case 'last7': return [fmt(t.subtract(6, 'day')), fmt(t)]
    case 'last15': return [fmt(t.subtract(14, 'day')), fmt(t)]
    case 'yesterday': return [fmt(t.subtract(1, 'day')), fmt(t.subtract(1, 'day'))]
    case 'today': return [fmt(t), fmt(t)]
    case 'week': return [fmt(t.startOf('week')), fmt(t)]
    case 'month': return [fmt(t.startOf('month')), fmt(t)]
    case 'year': return [fmt(t.startOf('year')), fmt(t)]
    case 'lastWeek': {
      const s = t.subtract(1, 'week').startOf('week')
      return [fmt(s), fmt(s.endOf('week'))]
    }
    case 'lastMonth': {
      const s = t.subtract(1, 'month').startOf('month')
      return [fmt(s), fmt(s.endOf('month'))]
    }
    case 'lastYear': {
      const s = t.subtract(1, 'year').startOf('year')
      return [fmt(s), fmt(s.endOf('year'))]
    }
    case 'last30': return [fmt(t.subtract(29, 'day')), fmt(t)]
    case 'last3m': return [fmt(t.subtract(3, 'month')), fmt(t)]
    case 'last365': return [fmt(t.subtract(1, 'year')), fmt(t)]
    default: return null
  }
}

function applyPreset(key: string) {
  activePreset.value = key
  dateRange.value = presetRange(key)
  handleSearch()
}

function handleRangeChange() {
  activePreset.value = ''
  handleSearch()
}

// ---- 筛选项设置 ----
const FILTER_CONFS = [
  { key: 'keyword', label: '搜索' },
  { key: 'productKw', label: '商品' },
  { key: 'warehouseId', label: '入库仓库' },
  { key: 'status', label: '单据状态' },
  { key: 'inType', label: '入库类型' },
] as const
const DEFAULT_FILTERS = ['keyword', 'productKw', 'warehouseId', 'status']
const FILTER_KEY = 'other-in-stock-filters'
const shownFilters = ref<string[]>([...DEFAULT_FILTERS])
const filterDialogVisible = ref(false)

function toggleFilter(key: string) {
  const i = shownFilters.value.indexOf(key)
  if (i >= 0) shownFilters.value.splice(i, 1)
  else shownFilters.value.push(key)
  localStorage.setItem(FILTER_KEY, JSON.stringify(shownFilters.value))
}

function resetFilterConfs() {
  shownFilters.value = [...DEFAULT_FILTERS]
  localStorage.removeItem(FILTER_KEY)
}

function loadFilterConfs() {
  try {
    const saved = localStorage.getItem(FILTER_KEY)
    if (saved) {
      const arr = JSON.parse(saved)
      if (Array.isArray(arr) && arr.length) shownFilters.value = arr
    }
  } catch {
    // 忽略损坏的本地配置
  }
}

// ---- 列表 ----
function buildQuery(): OtherInStockListQuery {
  const params: OtherInStockListQuery = {}
  if (filters.keyword) params.keyword = filters.keyword
  if (filters.productKw) params.productKw = filters.productKw
  if (filters.warehouseId) params.warehouseId = filters.warehouseId
  if (filters.status) params.status = filters.status
  if (filters.inType) params.inType = filters.inType
  if (dateRange.value?.[0]) params.startDate = dateRange.value[0]
  if (dateRange.value?.[1]) params.endDate = dateRange.value[1]
  return params
}

async function fetchList() {
  loading.value = true
  try {
    const res = await fetchOtherInStocks({
      ...buildQuery(),
      page: pagination.page,
      pageSize: pagination.pageSize,
    })
    if (res.data.code === 0 || res.data.code === 200) {
      list.value = res.data.data?.list ?? []
      total.value = res.data.data?.total ?? 0
    }
  } catch {
    // 拦截器已提示
  } finally {
    loading.value = false
  }
}

// ---- 导出 ----
const exporting = ref(false)

async function handleExport(cmd: string) {
  if (exporting.value) return
  exporting.value = true
  try {
    const isList = cmd === 'list'
    const res = isList
      ? await exportOtherInStocks(buildQuery())
      : await exportOtherInStockItems(buildQuery())
    const filename = `其他入库单${isList ? '列表' : '明细'}_${dayjs().format('YYYYMMDD')}.xlsx`
    const url = URL.createObjectURL(new Blob([res.data]))
    const a = document.createElement('a')
    a.href = url
    a.download = filename
    a.click()
    URL.revokeObjectURL(url)
  } catch {
    // 拦截器已提示
  } finally {
    exporting.value = false
  }
}

// ---- 导入 ----
const importVisible = ref(false)
const importing = ref(false)
const importFile = ref<File | null>(null)
const importFileList = ref<UploadFile[]>([])
const importApplyStock = ref(false)
const importResult = ref<OtherInStockImportResult | null>(null)
const billNoMapEntries = computed(() => Object.entries(importResult.value?.billNoMap ?? {}))

function handleImportFileChange(file: UploadFile) {
  importFile.value = file.raw ?? null
  importFileList.value = [file]
}

function handleImportFileRemove() {
  importFile.value = null
  importFileList.value = []
}

function resetImport() {
  importFile.value = null
  importFileList.value = []
  importApplyStock.value = false
  importResult.value = null
}

async function startImport() {
  if (!importFile.value || importing.value) return
  importing.value = true
  try {
    const res = await importOtherInStocks(importFile.value, importApplyStock.value)
    if (res.data.code === 0 || res.data.code === 200) {
      importResult.value = res.data.data
      ElMessage.success(`导入完成：新增 ${importResult.value.created} 张单，跳过 ${importResult.value.skipped} 行`)
      fetchList()
    }
  } catch {
    // 拦截器已提示
  } finally {
    importing.value = false
  }
}

function handleSearch() {
  pagination.page = 1
  fetchList()
}

function handleReset() {
  filters.keyword = ''
  filters.productKw = ''
  filters.warehouseId = undefined
  filters.status = ''
  filters.inType = ''
  dateRange.value = null
  activePreset.value = ''
  handleSearch()
}

function formatTime(val: string) {
  return val ? dayjs(val).format('YYYY-MM-DD HH:mm') : '-'
}

function goCreate() {
  router.push('/inventory/other-in/new')
}

function goEdit(row: OtherInStock) {
  router.push(`/inventory/other-in/${row.id}`)
}

function handleRowCommand(cmd: string, row: OtherInStock) {
  if (cmd === 'edit') {
    goEdit(row)
  } else if (cmd === 'complete') {
    ElMessageBox.confirm(`确认将单据 ${row.billNo} 过账？过账后库存增加且不可编辑。`, '过账确认', { type: 'warning' })
      .then(async () => {
        await completeOtherInStock(row.id)
        ElMessage.success('过账成功')
        fetchList()
      })
      .catch(() => {})
  } else if (cmd === 'delete') {
    ElMessageBox.confirm(`确认删除单据 ${row.billNo}？`, '删除确认', { type: 'warning' })
      .then(async () => {
        await deleteOtherInStock(row.id)
        ElMessage.success('删除成功')
        fetchList()
      })
      .catch(() => {})
  }
}

onMounted(() => {
  loadFilterConfs()
  fetchList()
})
</script>

<style scoped>
.page { display: flex; flex-direction: column; height: 100%; gap: 12px; }
.page-header { display: flex; justify-content: space-between; align-items: center; }
.page-title { font-size: 18px; font-weight: 600; margin: 0; }
.date-quick-bar {
  display: flex; align-items: center; gap: 4px;
  background: var(--el-bg-color); border-radius: 8px; padding: 8px 12px;
}
.preset-btn { margin-right: 8px; font-size: 14px; }
.date-range { margin-left: auto; }
.filter-row {
  display: flex; gap: 8px; align-items: center; flex-wrap: wrap;
  background: var(--el-bg-color); border-radius: 8px; padding: 12px;
}
.filter-label { color: var(--el-text-color-secondary); font-size: 13px; }
.filter-item { width: 160px; }
.filter-item-lg { width: 220px; }
.filter-actions { display: flex; align-items: center; gap: 10px; margin-left: auto; }
.row-more-btn { font-size: 16px; }
.pagination-row {
  display: flex; justify-content: space-between; align-items: center;
  background: var(--el-bg-color); border-radius: 8px; padding: 12px;
}
.total-text { color: var(--el-text-color-secondary); font-size: 13px; }
.filter-conf-item { display: flex; margin-right: 0; }
.apply-stock-checkbox { margin-top: 12px; }
.import-summary { display: flex; gap: 8px; margin-bottom: 12px; }
.billno-map { margin-bottom: 12px; font-size: 13px; color: var(--el-text-color-secondary); }
.billno-map p { margin: 2px 0; }
</style>
