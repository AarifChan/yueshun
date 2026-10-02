<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">库存状况表</h2>
      <div class="page-actions">
        <el-button @click="openImportDialog">导 入</el-button>
        <el-button @click="priceDialogVisible = true">价格展示设置</el-button>
        <el-button @click="openLimitDialog()">库存上下限设置</el-button>
        <el-button @click="handlePrint">打 印</el-button>
      </div>
    </div>
    <div class="page-body">
      <aside class="category-panel">
        <SideTreePanel
          ref="sideTreeRef"
          title="商品分类"
          :data="categoryTree"
          edit-to="/categories"
          placeholder="请输入分类名称"
          @node-click="handleCategoryClick"
          @title-click="handleCategoryRootClick"
        />
      </aside>
      <main class="list-panel">
        <div class="toolbar-row">
          <el-radio-group v-model="mode" @change="handleModeChange">
            <el-radio-button value="product">按商品</el-radio-button>
            <el-radio-button value="category">按分类</el-radio-button>
          </el-radio-group>
          <el-checkbox v-model="filters.onlyWithStock" @change="handleSearch">仅显示有账面库存商品</el-checkbox>
        </div>
        <div class="filter-row">
          <el-input
            v-model="filters.keyword"
            placeholder="商品名称/编号"
            clearable
            class="filter-keyword"
            @keyup.enter="handleSearch"
          >
            <template #prefix><el-icon><Search /></el-icon></template>
          </el-input>
          <RemoteSelect v-model="filters.warehouseId" api-url="/api/v1/warehouses" placeholder="仓库" class="filter-item" />
          <el-select v-model="filters.stockStatus" placeholder="库存状态" clearable class="filter-item">
            <el-option label="全部" value="" />
            <el-option label="有库存" value="in" />
            <el-option label="无库存" value="out" />
            <el-option label="低于下限" value="belowMin" />
            <el-option label="高于上限" value="aboveMax" />
          </el-select>
          <el-select v-model="filters.productStatus" placeholder="商品状态" clearable class="filter-item">
            <el-option label="全部" value="" />
            <el-option label="已启用" value="1" />
            <el-option label="已停用" value="0" />
          </el-select>
          <el-button type="primary" :icon="Search" @click="handleSearch">查 询</el-button>
        </div>

        <el-table
          v-if="mode === 'product'"
          :data="list"
          v-loading="loading"
          border
          stripe
          show-summary
          :summary-method="productSummary"
          @selection-change="(rows: StockStatusItem[]) => (selectedRows = rows)"
        >
          <el-table-column type="selection" width="45" />
          <el-table-column type="index" label="#" width="55" :index="indexMethod" />
          <el-table-column label="操作" width="70" align="center">
            <template #default="{ row }">
              <el-dropdown trigger="click" @command="(cmd: string) => handleRowCommand(cmd, row)">
                <el-button link type="primary" class="row-more-btn">
                  <el-icon><More /></el-icon>
                </el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item command="flows">明细</el-dropdown-item>
                    <el-dropdown-item command="distribution">库存分布</el-dropdown-item>
                    <el-dropdown-item command="limits">上下限设置</el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </template>
          </el-table-column>
          <el-table-column label="图片" width="70" align="center">
            <template #default="{ row }">
              <el-image v-if="row.image" :src="row.image" fit="cover" class="product-img" :preview-src-list="[row.image]" preview-teleported />
              <el-icon v-else class="product-img-placeholder"><Picture /></el-icon>
            </template>
          </el-table-column>
          <el-table-column label="商品名称" min-width="160" show-overflow-tooltip>
            <template #default="{ row }">
              <el-link type="primary" :underline="false" @click="goProduct(row)">{{ row.productName }}</el-link>
            </template>
          </el-table-column>
          <el-table-column prop="specification" label="规格" min-width="110" show-overflow-tooltip />
          <el-table-column prop="quantity" label="库存总量" width="100" align="right" sortable>
            <template #default="{ row }">
              <span :class="{ 'qty-zero': row.quantity <= 0 }">{{ formatQty(row.quantity) }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="unit" label="单位" width="70" />
          <el-table-column v-if="priceCols.costPrice" label="成本均价" width="100" align="right">
            <template #default="{ row }">{{ formatMoney(row.costPrice) }}</template>
          </el-table-column>
          <el-table-column v-if="priceCols.costAmount" prop="costAmount" label="库存总额" width="120" align="right">
            <template #default="{ row }">{{ formatMoney(row.costAmount) }}</template>
          </el-table-column>
          <el-table-column prop="reservedQty" label="预订量" width="90" align="right" />
          <el-table-column prop="pendingOutQty" label="待出库数量" width="100" align="right" />
          <el-table-column prop="availableQty" label="可用库存量" width="100" align="right">
            <template #default="{ row }">{{ formatQty(row.availableQty) }}</template>
          </el-table-column>
          <el-table-column prop="maxStock" label="库存上限" width="90" align="right" />
          <el-table-column prop="minStock" label="库存下限" width="90" align="right" />
          <el-table-column label="上下架状态" width="95" align="center">
            <template #default="{ row }">
              <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">{{ row.status === 1 ? '上架' : '下架' }}</el-tag>
            </template>
          </el-table-column>
        </el-table>

        <el-table
          v-else
          :data="categoryList"
          v-loading="loading"
          border
          stripe
          show-summary
          :summary-method="categorySummary"
        >
          <el-table-column type="index" label="#" width="55" />
          <el-table-column prop="categoryName" label="分类名称" min-width="180" />
          <el-table-column prop="productCount" label="商品数" width="100" align="right" />
          <el-table-column prop="quantity" label="库存总量" min-width="120" align="right">
            <template #default="{ row }">{{ formatQty(row.quantity) }}</template>
          </el-table-column>
          <el-table-column v-if="priceCols.costAmount" prop="costAmount" label="库存总额" min-width="140" align="right">
            <template #default="{ row }">{{ formatMoney(row.costAmount) }}</template>
          </el-table-column>
        </el-table>

        <div v-if="mode === 'product'" class="pagination-row">
          <span class="total-text">总{{ total }}条，每页显示{{ pagination.pageSize }}条</span>
          <el-pagination
            v-model:current-page="pagination.page"
            v-model:page-size="pagination.pageSize"
            :total="total"
            :page-sizes="[30, 50, 100, 200]"
            layout="prev, pager, next, sizes, jumper"
            background
            @current-change="fetchList"
            @size-change="handleSizeChange"
          />
        </div>
      </main>
    </div>

    <!-- 库存状况导入 -->
    <el-dialog v-model="importDialogVisible" title="库存状况导入" width="480px" @closed="resetImportDialog">
      <el-form label-width="90px">
        <el-form-item label="目标仓库">
          <RemoteSelect v-model="importWarehouseId" api-url="/api/v1/warehouses" placeholder="为空则取公司第一个仓库" class="import-warehouse" />
        </el-form-item>
        <el-form-item label="导出文件">
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
              <div class="el-upload__tip">支持老系统“库存状况列表”导出文件（.xls/.xlsx，≤10MB）</div>
            </template>
          </el-upload>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="importDialogVisible = false">取 消</el-button>
        <el-button type="primary" :loading="importing" @click="submitImport">开始导入</el-button>
      </template>
    </el-dialog>

    <!-- 价格展示设置 -->
    <el-dialog v-model="priceDialogVisible" title="价格展示设置" width="380px">
      <el-checkbox v-model="priceCols.costPrice">成本均价</el-checkbox>
      <el-checkbox v-model="priceCols.costAmount">库存总额（成本金额）</el-checkbox>
      <template #footer>
        <el-button type="primary" @click="savePriceCols">确 定</el-button>
      </template>
    </el-dialog>

    <!-- 库存上下限设置 -->
    <el-dialog v-model="limitDialogVisible" title="库存上下限设置" width="420px">
      <el-alert
        v-if="limitTargetIds.length > 1"
        type="info"
        :closable="false"
        :title="`将对选中的 ${limitTargetIds.length} 个商品批量设置`"
        class="limit-alert"
      />
      <el-form label-width="90px">
        <el-form-item label="库存上限">
          <el-input-number v-model="limitForm.maxStock" :min="0" :precision="2" class="limit-input" />
        </el-form-item>
        <el-form-item label="库存下限">
          <el-input-number v-model="limitForm.minStock" :min="0" :precision="2" class="limit-input" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="limitDialogVisible = false">取 消</el-button>
        <el-button type="primary" :loading="limitSaving" @click="submitLimits">确 定</el-button>
      </template>
    </el-dialog>

    <!-- 明细（出入库流水） -->
    <el-dialog v-model="flowDialogVisible" :title="`出入库明细 - ${currentProduct?.productName ?? ''}`" width="820px">
      <el-table :data="flowList" v-loading="flowLoading" border stripe size="small" max-height="420">
        <el-table-column prop="billDate" label="单据日期" width="110" />
        <el-table-column prop="billType" label="业务类型" width="100">
          <template #default="{ row }">
            <el-tag :type="row.inQty > 0 ? 'success' : 'warning'" size="small">{{ row.billType }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="billNo" label="单据编号" min-width="150" show-overflow-tooltip />
        <el-table-column prop="warehouseName" label="仓库" min-width="110" show-overflow-tooltip />
        <el-table-column label="入库量" width="90" align="right">
          <template #default="{ row }">{{ row.inQty > 0 ? formatQty(row.inQty) : '' }}</template>
        </el-table-column>
        <el-table-column label="出库量" width="90" align="right">
          <template #default="{ row }">{{ row.outQty > 0 ? formatQty(row.outQty) : '' }}</template>
        </el-table-column>
        <el-table-column label="单价" width="100" align="right">
          <template #default="{ row }">{{ row.price > 0 ? formatMoney(row.price) : '' }}</template>
        </el-table-column>
      </el-table>
      <div class="dialog-pagination">
        <el-pagination
          v-model:current-page="flowPagination.page"
          :page-size="flowPagination.pageSize"
          :total="flowTotal"
          layout="prev, pager, next, total"
          small
          @current-change="fetchFlows"
        />
      </div>
    </el-dialog>

    <!-- 库存分布 -->
    <el-dialog v-model="distDialogVisible" :title="`库存分布 - ${currentProduct?.productName ?? ''}`" width="560px">
      <el-table :data="distList" v-loading="distLoading" border stripe size="small" max-height="420">
        <el-table-column prop="warehouseName" label="仓库" min-width="140" />
        <el-table-column label="库存数量" width="110" align="right">
          <template #default="{ row }">{{ formatQty(row.quantity) }}</template>
        </el-table-column>
        <el-table-column label="占比" min-width="150">
          <template #default="{ row }">
            <el-progress
              :percentage="distTotalQty > 0 ? Math.round((row.quantity / distTotalQty) * 100) : 0"
              :stroke-width="10"
            />
          </template>
        </el-table-column>
        <el-table-column v-if="priceCols.costAmount" label="库存金额" width="120" align="right">
          <template #default="{ row }">{{ formatMoney(row.costAmount) }}</template>
        </el-table-column>
      </el-table>
      <div class="dist-total">合计库存：{{ formatQty(distTotalQty) }}</div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { More, Picture, Search, UploadFilled } from '@element-plus/icons-vue'
import SideTreePanel from '@/components/SideTreePanel.vue'
import RemoteSelect from '@/components/RemoteSelect.vue'
import { fetchCategoryTree, type CategoryNode } from '@/api/category'
import {
  fetchStockDistribution,
  fetchStockFlows,
  fetchStockStatus,
  importStockStatus,
  updateStockLimits,
  type StockDistributionItem,
  type StockFlowItem,
  type StockStatusCategoryItem,
  type StockStatusItem,
  type StockStatusSummary,
} from '@/api/stockStatus'

const router = useRouter()

const mode = ref<'product' | 'category'>('product')
const list = ref<StockStatusItem[]>([])
const categoryList = ref<StockStatusCategoryItem[]>([])
const total = ref(0)
const summary = ref<StockStatusSummary>({ totalQuantity: 0, totalAmount: 0 })
const loading = ref(false)
const selectedRows = ref<StockStatusItem[]>([])

const pagination = reactive({ page: 1, pageSize: 30 })
const filters = reactive({
  keyword: '',
  warehouseId: undefined as number | undefined,
  categoryId: undefined as number | undefined,
  stockStatus: '',
  productStatus: '',
  onlyWithStock: false,
})

const categoryTree = ref<CategoryNode[]>([])
const sideTreeRef = ref()

const PRICE_COLS_KEY = 'stock-status-price-cols'
const priceCols = reactive({ costPrice: true, costAmount: true })
const priceDialogVisible = ref(false)

const limitDialogVisible = ref(false)
const limitSaving = ref(false)
const limitTargetIds = ref<number[]>([])
const limitForm = reactive({ minStock: 0, maxStock: 0 })

function normalizeCategoryNodes(nodes: CategoryNode[]): CategoryNode[] {
  return nodes.map((n) => ({ ...n, children: n.children ? normalizeCategoryNodes(n.children) : [] }))
}

async function loadCategoryTree() {
  try {
    const res = await fetchCategoryTree()
    if (res.data.code === 0 || res.data.code === 200) {
      categoryTree.value = normalizeCategoryNodes(res.data.data || [])
    }
  } catch {
    // 拦截器已提示
  }
}

async function fetchList() {
  loading.value = true
  try {
    const params: Record<string, any> = {
      page: pagination.page,
      pageSize: pagination.pageSize,
      groupBy: mode.value,
      onlyWithStock: filters.onlyWithStock,
    }
    if (filters.keyword) params.keyword = filters.keyword
    if (filters.warehouseId) params.warehouseId = filters.warehouseId
    if (filters.categoryId) params.categoryId = filters.categoryId
    if (filters.stockStatus) params.stockStatus = filters.stockStatus
    if (filters.productStatus !== '') params.productStatus = filters.productStatus

    const res = await fetchStockStatus(params)
    if (res.data.code === 0 || res.data.code === 200) {
      const data = res.data.data
      if (mode.value === 'product') {
        list.value = data?.list ?? []
      } else {
        categoryList.value = data?.list ?? []
      }
      total.value = data?.total ?? 0
      summary.value = data?.summary ?? { totalQuantity: 0, totalAmount: 0 }
    }
  } catch {
    // 拦截器已提示
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  pagination.page = 1
  fetchList()
}

function handleModeChange() {
  pagination.page = 1
  fetchList()
}

function handleSizeChange() {
  pagination.page = 1
  fetchList()
}

function handleCategoryClick(data: CategoryNode) {
  filters.categoryId = data.id
  pagination.page = 1
  fetchList()
}

function handleCategoryRootClick() {
  filters.categoryId = undefined
  sideTreeRef.value?.clearCurrentKey()
  pagination.page = 1
  fetchList()
}

function indexMethod(index: number) {
  return (pagination.page - 1) * pagination.pageSize + index + 1
}

function formatQty(val: number) {
  return Number.isInteger(val) ? String(val) : val.toFixed(2)
}

function formatMoney(val: number) {
  return `¥ ${val.toFixed(2)}`
}

function productSummary({ columns }: { columns: { property?: string }[] }) {
  return columns.map((col, i) => {
    if (i === 0) return '合计'
    if (col.property === 'quantity') return formatQty(summary.value.totalQuantity)
    if (col.property === 'costAmount') return formatMoney(summary.value.totalAmount)
    return ''
  })
}

function categorySummary({ columns }: { columns: { property?: string }[] }) {
  return columns.map((col, i) => {
    if (i === 0) return '合计'
    if (col.property === 'quantity') return formatQty(summary.value.totalQuantity)
    if (col.property === 'costAmount') return formatMoney(summary.value.totalAmount)
    return ''
  })
}

function goProduct(row: StockStatusItem) {
  router.push(`/products/${row.productId}`)
}

// ==================== 行操作：明细 / 库存分布 ====================

const currentProduct = ref<StockStatusItem | null>(null)

const flowDialogVisible = ref(false)
const flowLoading = ref(false)
const flowList = ref<StockFlowItem[]>([])
const flowTotal = ref(0)
const flowPagination = reactive({ page: 1, pageSize: 10 })

const distDialogVisible = ref(false)
const distLoading = ref(false)
const distList = ref<StockDistributionItem[]>([])
const distTotalQty = ref(0)

function handleRowCommand(cmd: string, row: StockStatusItem) {
  currentProduct.value = row
  if (cmd === 'flows') {
    flowPagination.page = 1
    flowDialogVisible.value = true
    fetchFlows()
  } else if (cmd === 'distribution') {
    distDialogVisible.value = true
    fetchDistribution()
  } else if (cmd === 'limits') {
    openLimitDialog(row)
  }
}

async function fetchFlows() {
  if (!currentProduct.value) return
  flowLoading.value = true
  try {
    const res = await fetchStockFlows(currentProduct.value.productId, {
      page: flowPagination.page,
      pageSize: flowPagination.pageSize,
    })
    if (res.data.code === 0 || res.data.code === 200) {
      flowList.value = res.data.data?.list ?? []
      flowTotal.value = res.data.data?.total ?? 0
    }
  } catch {
    // 拦截器已提示
  } finally {
    flowLoading.value = false
  }
}

async function fetchDistribution() {
  if (!currentProduct.value) return
  distLoading.value = true
  try {
    const res = await fetchStockDistribution(currentProduct.value.productId)
    if (res.data.code === 0 || res.data.code === 200) {
      distList.value = res.data.data?.list ?? []
      distTotalQty.value = res.data.data?.totalQuantity ?? 0
    }
  } catch {
    // 拦截器已提示
  } finally {
    distLoading.value = false
  }
}

function openLimitDialog(row?: StockStatusItem) {
  if (row) {
    limitTargetIds.value = [row.productId]
    limitForm.minStock = row.minStock
    limitForm.maxStock = row.maxStock
  } else {
    if (selectedRows.value.length === 0) {
      ElMessage.warning('请先勾选要设置的商品')
      return
    }
    limitTargetIds.value = selectedRows.value.map((r) => r.productId)
    limitForm.minStock = 0
    limitForm.maxStock = 0
  }
  limitDialogVisible.value = true
}

async function submitLimits() {
  limitSaving.value = true
  try {
    await updateStockLimits({
      productIds: limitTargetIds.value,
      minStock: limitForm.minStock,
      maxStock: limitForm.maxStock,
    })
    ElMessage.success('设置成功')
    limitDialogVisible.value = false
    fetchList()
  } catch {
    // 拦截器已提示
  } finally {
    limitSaving.value = false
  }
}

function savePriceCols() {
  localStorage.setItem(PRICE_COLS_KEY, JSON.stringify(priceCols))
  priceDialogVisible.value = false
}

function loadPriceCols() {
  try {
    const saved = localStorage.getItem(PRICE_COLS_KEY)
    if (saved) Object.assign(priceCols, JSON.parse(saved))
  } catch {
    // 忽略损坏的本地配置
  }
}

function handlePrint() {
  window.print()
}

// ==================== 库存状况导入 ====================

const importDialogVisible = ref(false)
const importWarehouseId = ref<number | undefined>(undefined)
const importFileList = ref<any[]>([])
const importFile = ref<File | null>(null)
const importing = ref(false)

function openImportDialog() {
  importDialogVisible.value = true
}

function resetImportDialog() {
  importWarehouseId.value = undefined
  importFileList.value = []
  importFile.value = null
}

function handleImportFileChange(uploadFile: any) {
  importFileList.value = [uploadFile]
  importFile.value = uploadFile.raw ?? null
}

function handleImportFileRemove() {
  importFileList.value = []
  importFile.value = null
}

async function submitImport() {
  if (!importFile.value) {
    ElMessage.warning('请先选择要导入的文件')
    return
  }
  importing.value = true
  try {
    const res = await importStockStatus(importFile.value, importWarehouseId.value)
    if (res.data.code === 0 || res.data.code === 200) {
      const data = res.data.data
      ElMessage.success(
        `导入完成（${data.warehouse}）：更新 ${data.updated}，新增 ${data.created}，库存写入 ${data.stocked}` +
          (data.skipped ? `，跳过 ${data.skipped} 行` : ''),
      )
      importDialogVisible.value = false
      fetchList()
    }
  } catch {
    // 拦截器已提示
  } finally {
    importing.value = false
  }
}

onMounted(() => {
  loadPriceCols()
  loadCategoryTree()
  fetchList()
})
</script>

<style scoped>
.page { display: flex; flex-direction: column; height: 100%; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; }
.page-title { font-size: 18px; font-weight: 600; margin: 0; }
.page-actions { display: flex; gap: 8px; }
.page-body { display: flex; gap: 12px; flex: 1; min-height: 0; }
.category-panel {
  width: 220px;
  flex-shrink: 0;
  background: var(--el-bg-color);
  border-radius: 8px;
  padding: 12px;
  display: flex;
  overflow: auto;
}
.list-panel {
  flex: 1;
  min-width: 0;
  background: var(--el-bg-color);
  border-radius: 8px;
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.toolbar-row { display: flex; justify-content: space-between; align-items: center; }
.filter-row { display: flex; gap: 8px; align-items: center; flex-wrap: wrap; }
.filter-keyword { width: 220px; }
.filter-item { width: 160px; }
.product-img { width: 36px; height: 36px; border-radius: 4px; display: block; margin: 0 auto; }
.product-img-placeholder { font-size: 22px; color: var(--el-text-color-placeholder); }
.qty-zero { color: var(--el-color-danger); font-weight: 600; }
.row-more-btn { font-size: 16px; }
.dialog-pagination { display: flex; justify-content: flex-end; margin-top: 12px; }
.dist-total { margin-top: 12px; text-align: right; font-weight: 600; }
.pagination-row { display: flex; justify-content: space-between; align-items: center; }
.total-text { color: var(--el-text-color-secondary); font-size: 13px; }
.limit-alert { margin-bottom: 12px; }
.limit-input { width: 100%; }
.import-warehouse { width: 100%; }

@media print {
  .page-header .page-actions, .category-panel, .filter-row, .toolbar-row, .pagination-row { display: none; }
}
</style>
