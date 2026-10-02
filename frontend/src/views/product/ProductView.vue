<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">商品信息列表</h2>
      <div class="page-actions">
        <el-button @click="openImport">商品导入</el-button>
        <el-button type="primary" @click="goCreate"><el-icon><Plus /></el-icon>新增商品</el-button>
      </div>
    </div>
    <div class="page-body">
      <aside class="category-panel" :class="{ collapsed: categoryPanelCollapsed }">
        <div v-show="!categoryPanelCollapsed" class="category-panel-content">
          <SideTreePanel
            ref="sideTreeRef"
            title="商品分类"
            :data="categoryTree"
            edit-to="/categories"
            placeholder="请输入分类名称"
            @node-click="handleCategoryClick"
            @title-click="handleCategoryRootClick"
          />
        </div>
        <el-tooltip v-if="!categoryPanelCollapsed" content="收起" placement="right">
          <div class="panel-collapse-handle" @click="categoryPanelCollapsed = true">
            <el-icon><DArrowLeft /></el-icon>
          </div>
        </el-tooltip>
        <el-tooltip v-else content="展开" placement="right">
          <div class="panel-expand-strip" @click="categoryPanelCollapsed = false">
            <el-icon><DArrowRight /></el-icon>
            <span class="panel-expand-text">展开</span>
          </div>
        </el-tooltip>
      </aside>
      <main class="list-panel">
        <el-radio-group v-model="mode" class="mode-tabs" @change="handleModeChange">
          <el-radio-button value="product">按商品</el-radio-button>
          <el-radio-button value="spec">按规格</el-radio-button>
        </el-radio-group>
        <div class="filter-row">
          <div class="filter-cell" :style="filterCellStyle(keywordFilterConf)">
            <el-input v-model="filters.keyword" placeholder="商品名称/编号/条码" clearable @keyup.enter="handleSearch" />
          </div>
          <div v-for="conf in shownFilterConfs" :key="conf.key" class="filter-cell" :style="filterCellStyle(conf)">
            <el-input v-if="conf.key === 'name'" v-model="filters.name" placeholder="请输入商品名称" clearable @keyup.enter="handleSearch" />
            <el-tree-select
              v-else-if="conf.key === 'category'"
              v-model="filters.categoryId"
              :data="categoryTree"
              :props="{ label: 'name', children: 'children' }"
              node-key="id"
              check-strictly
              clearable
              placeholder="商品分类"
              class="filter-control"
              @change="handleFilterCategoryChange"
            />
            <el-select v-else-if="conf.key === 'tag'" v-model="filters.tagId" placeholder="商品标签" clearable class="filter-control">
              <el-option v-for="tag in tagOptions" :key="tag.id" :label="tag.name" :value="tag.id" />
            </el-select>
            <el-select v-else-if="conf.key === 'topic'" v-model="filters.topicCategory" placeholder="专题分类" filterable allow-create clearable class="filter-control">
              <el-option v-for="topic in topicOptions" :key="topic" :label="topic" :value="topic" />
            </el-select>
            <el-select v-else-if="conf.key === 'brand'" v-model="filters.brandId" placeholder="商品品牌" clearable class="filter-control">
              <el-option v-for="brand in brandOptions" :key="brand.id" :label="brand.name" :value="brand.id" />
            </el-select>
            <el-select v-else-if="conf.key === 'image'" v-model="filters.hasImage" placeholder="商品图片" class="filter-control">
              <el-option label="全部" value="" />
              <el-option label="有图" :value="1" />
              <el-option label="无图" :value="0" />
            </el-select>
            <el-select v-else-if="conf.key === 'status'" v-model="filters.status" placeholder="商品状态" class="filter-control">
              <el-option label="全部" value="" />
              <el-option label="已启用" :value="1" />
              <el-option label="已禁用" :value="0" />
            </el-select>
            <el-select v-else-if="conf.key === 'stock'" v-model="filters.stockStatus" placeholder="库存状态" class="filter-control">
              <el-option label="全部" value="" />
              <el-option label="有库存" value="in" />
              <el-option label="无库存" value="out" />
            </el-select>
          </div>
          <div class="filter-actions">
            <el-tooltip content="筛选项设置" placement="top">
              <el-button :icon="Setting" @click="openFilterDialog" />
            </el-tooltip>
            <el-button type="primary" @click="handleSearch">查 询</el-button>
          </div>
        </div>
        <div class="table-wrap">
        <el-table
          v-if="mode === 'product'"
          :data="productList"
          v-loading="loading"
          border
          stripe
          height="100%"
          row-key="id"
          :expand-row-keys="expandedKeys"
          @sort-change="handleSortChange"
        >
          <el-table-column type="expand" width="1" class-name="expand-content-col">
            <template #default="{ row }">
              <div class="spec-detail" v-loading="!!specDetails[row.id]?.loading">
                <el-table v-if="specDetails[row.id]?.items?.length" :data="specDetails[row.id].items" size="small" border>
                  <el-table-column prop="specValue" label="规格值" min-width="120" />
                  <el-table-column prop="code" label="商品编号" min-width="110" />
                  <el-table-column label="规格条码" min-width="110"><template #default="{ row: spec }">{{ spec.barcode || '-' }}</template></el-table-column>
                  <el-table-column label="上架" width="70"><template #default="{ row: spec }">{{ spec.onShelf === 1 ? '是' : '否' }}</template></el-table-column>
                  <el-table-column prop="weight" label="重量" width="90" />
                  <el-table-column prop="volume" label="体积" width="90" />
                </el-table>
                <div v-else-if="!specDetails[row.id]?.loading" class="spec-empty">暂无规格</div>
              </div>
            </template>
          </el-table-column>
          <el-table-column width="60" :fixed="freezeCount > 0 ? 'left' : undefined">
            <template #header>
              <el-tooltip content="表头字段设置" placement="top">
                <el-icon class="col-setting-icon" @click="openColumnDialog"><Setting /></el-icon>
              </el-tooltip>
            </template>
            <template #default="{ row, $index }">
              <div class="index-cell" @click.stop="toggleExpand(row)">
                <span class="index-num">{{ (pagination.page - 1) * pagination.pageSize + $index + 1 }}</span>
                <el-icon class="expand-arrow"><ArrowDown v-if="expandedKeys.includes(row.id)" /><ArrowRight v-else /></el-icon>
              </div>
            </template>
          </el-table-column>
          <el-table-column type="selection" width="45" :fixed="freezeCount > 0 ? 'left' : undefined" />
          <el-table-column
            v-for="(col, idx) in visibleColumns"
            :key="col.key"
            :prop="col.key"
            :label="col.label"
            :width="col.width"
            :min-width="col.minWidth"
            :sortable="col.sortable ? 'custom' : false"
            :fixed="idx < freezeCount ? 'left' : undefined"
          >
            <template v-if="col.key === 'brand'" #header>
              <span class="brand-header">
                <span>商品品牌</span>
                <el-dropdown trigger="click" popper-class="brand-filter-popper" @command="handleBrandFilter">
                  <el-icon class="brand-filter-icon" :class="{ active: !!filters.brandId }"><CaretBottom /></el-icon>
                  <template #dropdown>
                    <el-dropdown-menu>
                      <el-dropdown-item :command="0" :class="{ active: !filters.brandId }">全部</el-dropdown-item>
                      <el-dropdown-item
                        v-for="brand in brandOptions"
                        :key="brand.id"
                        :command="brand.id"
                        :class="{ active: filters.brandId === brand.id }"
                      >
                        {{ brand.name }}
                      </el-dropdown-item>
                    </el-dropdown-menu>
                  </template>
                </el-dropdown>
              </span>
            </template>
            <template #default="{ row }">
              <template v-if="col.key === 'image'">
                <el-image v-if="row.image" :src="row.image" fit="cover" class="thumb" :preview-src-list="[row.image]" preview-teleported />
                <div v-else class="thumb thumb-placeholder"><el-icon><Picture /></el-icon></div>
              </template>
              <el-link v-else-if="col.key === 'name'" type="primary" :underline="false" @click="goDetail(row.id)">{{ row.name }}</el-link>
              <template v-else-if="col.key === 'spec'">{{ row.specCount ?? 0 }}种</template>
              <template v-else>{{ cellText(col.key, row) }}</template>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="70">
            <template #default="{ row }">
              <el-dropdown trigger="click" @command="(cmd: string) => handleCommand(cmd, row)">
                <el-button text><el-icon><MoreFilled /></el-icon></el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item command="view">查看</el-dropdown-item>
                    <el-dropdown-item command="edit">编辑</el-dropdown-item>
                    <el-dropdown-item command="delete">删除</el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </template>
          </el-table-column>
        </el-table>
        <el-table v-else :data="specList" v-loading="loading" border stripe height="100%">
          <el-table-column type="selection" width="45" />
          <el-table-column label="商品图片" width="80">
            <template #default="{ row }">
              <el-image v-if="row.productImage" :src="row.productImage" fit="cover" class="thumb" :preview-src-list="[row.productImage]" preview-teleported />
              <div v-else class="thumb thumb-placeholder"><el-icon><Picture /></el-icon></div>
            </template>
          </el-table-column>
          <el-table-column label="商品名称" min-width="160">
            <template #default="{ row }">
              <el-link type="primary" :underline="false" @click="goDetail(row.productId)">{{ row.productName }}</el-link>
            </template>
          </el-table-column>
          <el-table-column prop="specValue" label="规格值" min-width="120" />
          <el-table-column prop="code" label="商品编号" min-width="110" />
          <el-table-column prop="barcode" label="规格条码" min-width="110"><template #default="{ row }">{{ row.barcode || '-' }}</template></el-table-column>
          <el-table-column prop="brandName" label="商品品牌" min-width="100"><template #default="{ row }">{{ row.brandName || '-' }}</template></el-table-column>
          <el-table-column label="上架" width="80"><template #default="{ row }">{{ row.onShelf === 1 ? '是' : '否' }}</template></el-table-column>
          <el-table-column label="操作" width="110">
            <template #default="{ row }">
              <el-link type="primary" :underline="false" @click="goDetail(row.productId)">查看</el-link>
              <el-divider direction="vertical" />
              <el-link type="primary" :underline="false" @click="goEdit(row.productId)">编辑</el-link>
            </template>
          </el-table-column>
        </el-table>
        </div>
        <div class="pagination-bar">
          <span class="pagination-info">总{{ total }}条，每页显示{{ pagination.pageSize }}条</span>
          <div class="pagination-controls">
            <el-button :icon="Refresh" @click="fetchList">刷新</el-button>
            <el-pagination
              v-model:current-page="pagination.page"
              v-model:page-size="pagination.pageSize"
              :total="total"
              :page-sizes="[30, 50, 100]"
              layout="prev, pager, next, sizes, jumper"
              background
              @size-change="handleSizeChange"
              @current-change="handleCurrentChange"
            />
          </div>
        </div>
      </main>
    </div>

    <el-dialog v-model="columnDialogVisible" title="编辑显示字段" width="80%" :close-on-click-modal="false" @closed="destroyColumnSortable">
      <div class="column-dialog-body">
        <div class="column-options">
          <div class="column-panel-title">可选字段</div>
          <div class="column-checkbox-grid">
            <el-checkbox
              v-for="field in ALL_COLUMNS"
              :key="field.key"
              :model-value="draftKeys.includes(field.key)"
              :disabled="field.required"
              class="column-checkbox"
              @change="(val: boolean | string | number) => toggleDraftColumn(field.key, !!val)"
            >
              {{ field.label }}
            </el-checkbox>
          </div>
        </div>
        <div class="column-selected">
          <div class="column-selected-header">
            <span class="column-panel-title">当前选中字段排序</span>
            <span class="freeze-setting">
              冻结前
              <el-select v-model="draftFreeze" class="freeze-select">
                <el-option v-for="n in 6" :key="n - 1" :label="`${n - 1}列`" :value="n - 1" />
              </el-select>
            </span>
          </div>
          <div ref="sortListRef" class="column-sort-list">
            <div v-for="key in draftKeys" :key="key" class="column-sort-item">
              <el-icon class="drag-handle"><Rank /></el-icon>
              <span class="column-sort-label">{{ columnLabel(key) }}</span>
              <el-icon v-if="!isRequiredColumn(key)" class="column-remove" @click="removeDraftColumn(key)"><Close /></el-icon>
            </div>
          </div>
        </div>
      </div>
      <template #footer>
        <div class="column-dialog-footer">
          <el-link type="primary" :underline="false" @click="resetColumns">重置</el-link>
          <span>
            <el-button @click="columnDialogVisible = false">取消</el-button>
            <el-button type="primary" @click="applyColumns">确定</el-button>
          </span>
        </div>
      </template>
    </el-dialog>

    <el-dialog v-model="filterDialogVisible" title="筛选项设置" width="65%" @closed="destroyFilterSortable">
      <el-input v-model="filterDialogKeyword" placeholder="请输入关键词" clearable :prefix-icon="Search" class="filter-dialog-search" />
      <el-table ref="filterTableRef" :data="filteredDraftFilterConf" row-key="key" border size="small">
        <el-table-column label="序" width="50"><template #default="{ $index }">{{ $index + 1 }}</template></el-table-column>
        <el-table-column label="排序" width="60" align="center">
          <template #default><el-icon class="drag-handle"><Rank /></el-icon></template>
        </el-table-column>
        <el-table-column label="名称" min-width="120"><template #default="{ row }">{{ filterLabel(row.key) }}</template></el-table-column>
        <el-table-column label="是否显示" width="110" align="center">
          <template #header>
            <el-checkbox :model-value="draftAllShown" @change="(val: boolean | string | number) => toggleDraftShowAll(!!val)">是否显示</el-checkbox>
          </template>
          <template #default="{ row }">
            <el-checkbox :model-value="row.show" :disabled="row.key === 'keyword'" @change="(val: boolean | string | number) => (row.show = !!val)" />
          </template>
        </el-table-column>
        <el-table-column label="列宽" width="110">
          <template #default="{ row }">
            <el-select v-model="row.width" size="small">
              <el-option v-for="n in [1, 2, 3]" :key="n" :label="String(n)" :value="n" />
            </el-select>
          </template>
        </el-table-column>
      </el-table>
      <template #footer>
        <div class="column-dialog-footer">
          <el-link type="primary" :underline="false" @click="resetFilterConfig">恢复默认</el-link>
          <span>
            <el-button @click="filterDialogVisible = false">取消</el-button>
            <el-button type="primary" @click="applyFilterConfig">确定</el-button>
          </span>
        </div>
      </template>
    </el-dialog>

    <el-dialog v-model="importVisible" title="商品导入" width="560px" :close-on-click-modal="false" @closed="resetImport">
      <div v-if="importStep === 1" class="import-modes">
        <div
          v-for="opt in importModeOptions"
          :key="opt.value"
          class="import-mode-card"
          :class="{ active: importMode === opt.value }"
          @click="importMode = opt.value"
        >
          <el-radio v-model="importMode" :value="opt.value" class="import-mode-radio">{{ opt.label }}</el-radio>
          <div class="import-mode-desc">
            <p v-for="(line, i) in opt.desc" :key="i">{{ line }}</p>
          </div>
        </div>
        <p class="import-notice">注意：两种方式的导入模板不能通用，请核对！</p>
      </div>
      <div v-else>
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
              <div class="el-upload__tip">支持 .xlsx / .xls 格式文件</div>
            </template>
          </el-upload>
          <el-checkbox v-if="importMode === 'custom'" v-model="overwriteEmpty" class="overwrite-checkbox">空信息覆盖</el-checkbox>
        </template>
        <template v-else>
          <div class="import-summary">
            <el-tag type="success">成功新增 {{ importResult.created }}</el-tag>
            <el-tag v-if="importMode === 'custom'" type="warning">更新 {{ importResult.updated ?? 0 }}</el-tag>
            <el-tag type="danger">失败 {{ importResult.failed }}</el-tag>
          </div>
          <el-table v-if="importResult.errors && importResult.errors.length" :data="importResult.errors.slice(0, 20)" border size="small" max-height="260">
            <el-table-column prop="row" label="行号" width="80" />
            <el-table-column prop="code" label="编号" width="140" />
            <el-table-column prop="reason" label="原因" min-width="200" />
          </el-table>
        </template>
      </div>
      <template #footer>
        <template v-if="importStep === 1">
          <el-button @click="importVisible = false">取消</el-button>
          <el-button type="primary" @click="importStep = 2">下一步</el-button>
        </template>
        <template v-else-if="!importResult">
          <el-button @click="importStep = 1">上一步</el-button>
          <el-button type="primary" :loading="importing" :disabled="!importFile" @click="startImport">开始导入</el-button>
        </template>
        <template v-else>
          <el-button type="primary" @click="finishImport">完成</el-button>
        </template>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowDown, ArrowRight, CaretBottom, Close, DArrowLeft, DArrowRight, MoreFilled, Picture, Plus, Rank, Refresh, Search, Setting, UploadFilled } from '@element-plus/icons-vue'
import type { UploadFile } from 'element-plus'
import Sortable from 'sortablejs'
import {
  fetchProducts,
  fetchSpecItems,
  deleteProduct,
  importProductsSystem,
  importProductsCustom,
  fetchSettings,
  saveSettings,
  type ProductListItem,
  type SpecListItem,
  type ImportResult,
} from '@/api/productList'
import { fetchCategoryTree, type CategoryNode } from '@/api/category'
import { fetchTags, type ProductTag } from '@/api/tag'
import { fetchBrandOptions, fetchProduct, fetchTopicCategories } from '@/api/productDetail'
import SideTreePanel from '@/components/SideTreePanel.vue'

const router = useRouter()
const route = useRoute()

const mode = ref<'product' | 'spec'>('product')
const loading = ref(false)
const productList = ref<ProductListItem[]>([])
const specList = ref<SpecListItem[]>([])
const total = ref(0)
const pagination = reactive({ page: 1, pageSize: 100 })
const filters = reactive<{ keyword: string; name: string; categoryId?: number; brandId?: number; tagId?: number; topicCategory: string; stockStatus: string; hasImage: number | ''; status: number | '' }>({
  keyword: '',
  name: '',
  categoryId: undefined,
  brandId: undefined,
  tagId: undefined,
  topicCategory: '',
  stockStatus: '',
  hasImage: '',
  status: '',
})

const categoryTree = ref<CategoryNode[]>([])
const sideTreeRef = ref()
const tagOptions = ref<ProductTag[]>([])
const brandOptions = ref<{ id: number; name: string }[]>([])
const topicOptions = ref<string[]>([])
const categoryPanelCollapsed = ref(false)

function normalizeCategoryNodes(nodes: CategoryNode[]): CategoryNode[] {
  return nodes.map((n) => ({ ...n, children: n.children ? normalizeCategoryNodes(n.children) : [] }))
}

function handleCategoryRootClick() {
  filters.categoryId = undefined
  sideTreeRef.value?.clearCurrentKey()
  pagination.page = 1
  fetchList()
}

const sortState = reactive<{ orderBy?: string; order?: 'asc' | 'desc' }>({})
const SORTABLE_KEYS = new Set(['code', 'barcode', 'name'])

function handleSortChange({ prop, order }: { prop: string; order: 'ascending' | 'descending' | null }) {
  if (!order || !SORTABLE_KEYS.has(prop)) {
    sortState.orderBy = undefined
    sortState.order = undefined
  } else {
    sortState.orderBy = prop
    sortState.order = order === 'ascending' ? 'asc' : 'desc'
  }
  pagination.page = 1
  fetchList()
}

function handleBrandFilter(command: number) {
  filters.brandId = command > 0 ? command : undefined
  pagination.page = 1
  fetchList()
}

interface SpecDetailRow {
  specValue: string
  code: string
  barcode: string
  onShelf: number
  weight: number
  volume: number
}

const expandedKeys = ref<number[]>([])
const specDetails = reactive<Record<number, { loading: boolean; items: SpecDetailRow[] }>>({})

function toggleExpand(row: ProductListItem) {
  const idx = expandedKeys.value.indexOf(row.id)
  if (idx >= 0) {
    expandedKeys.value.splice(idx, 1)
    return
  }
  expandedKeys.value.push(row.id)
  loadSpecDetails(row.id)
}

async function loadSpecDetails(id: number) {
  if (specDetails[id]) return
  specDetails[id] = { loading: true, items: [] }
  try {
    const res = await fetchProduct(id)
    if (res.data.code === 0 || res.data.code === 200) {
      specDetails[id] = { loading: false, items: res.data.data?.specItems ?? [] }
    } else {
      specDetails[id] = { loading: false, items: [] }
    }
  } catch {
    specDetails[id] = { loading: false, items: [] }
  }
}

interface ColumnDef {
  key: string
  label: string
  width?: number
  minWidth?: number
  required?: boolean
  sortable?: boolean
}

const ALL_COLUMNS: ColumnDef[] = [
  { key: 'image', label: '图片', width: 70 },
  { key: 'code', label: '商品编号', minWidth: 110, sortable: true },
  { key: 'barcode', label: '商品条码', minWidth: 110, sortable: true },
  { key: 'name', label: '商品名称', minWidth: 160, required: true, sortable: true },
  { key: 'mallName', label: '商城展示名称', minWidth: 130 },
  { key: 'spec', label: '规格', width: 80, required: true },
  { key: 'category', label: '商品分类', minWidth: 100 },
  { key: 'brand', label: '商品品牌', minWidth: 100 },
  { key: 'tag', label: '商品标签', minWidth: 140 },
  { key: 'topic', label: '专题分类', minWidth: 100 },
  { key: 'pinyin', label: '拼音码', minWidth: 100 },
  { key: 'unit', label: '基本单位', width: 90 },
  { key: 'conversion', label: '换算关系', minWidth: 100 },
  { key: 'purchasePrice', label: '参考采购价', minWidth: 100 },
  { key: 'latestPurchase', label: '最近进价', minWidth: 100 },
  { key: 'defaultPrice', label: '默认订货价', minWidth: 100 },
  { key: 'retailPrice', label: '零售价', minWidth: 90 },
  { key: 'warehouse', label: '出库仓库', minWidth: 100 },
  { key: 'salesQty', label: '销量', width: 80 },
  { key: 'remark', label: '备注', minWidth: 140 },
  { key: 'totalStock', label: '库存总量', minWidth: 90 },
  { key: 'availableStock', label: '可用库存量', minWidth: 100 },
  { key: 'createdAt', label: '创建时间', minWidth: 110 },
  { key: 'shelfStatus', label: '上下架状态', minWidth: 100 },
  { key: 'sortWeight', label: '排序权重', width: 90 },
  { key: 'status', label: '商品状态', width: 90 },
  { key: 'weight', label: '商品重量(kg)', minWidth: 110 },
  { key: 'volume', label: '商品体积(m³)', minWidth: 110 },
]

const COLUMN_SETTING_KEY = 'productList.columnConfig'
const DEFAULT_COLUMNS = ['image', 'code', 'barcode', 'name', 'spec', 'brand', 'tag']
const DEFAULT_FREEZE = 2

const columnKeys = ref<string[]>([...DEFAULT_COLUMNS])
const freezeCount = ref(DEFAULT_FREEZE)
const visibleColumns = computed(() =>
  columnKeys.value.map((key) => ALL_COLUMNS.find((c) => c.key === key)).filter((c): c is ColumnDef => !!c)
)

const columnDialogVisible = ref(false)
const draftKeys = ref<string[]>([])
const draftFreeze = ref(0)
const sortListRef = ref<HTMLElement>()
let columnSortable: Sortable | null = null

function columnLabel(key: string) {
  return ALL_COLUMNS.find((c) => c.key === key)?.label ?? key
}

function isRequiredColumn(key: string) {
  return !!ALL_COLUMNS.find((c) => c.key === key)?.required
}

function fmtPrice(v?: number) {
  return v === undefined || v === null ? '-' : v.toFixed(2)
}

function cellText(key: string, row: ProductListItem): string {
  switch (key) {
    case 'code': return row.code || '-'
    case 'barcode': return row.barcode || '-'
    case 'mallName': return row.mallName || '-'
    case 'category': return row.categoryName || '-'
    case 'brand': return row.brandName || '-'
    case 'tag': return row.tags && row.tags.length ? row.tags.join('，') : '-'
    case 'topic': return row.topicCategory || '-'
    case 'pinyin': return row.pinyinCode || '-'
    case 'unit': return row.unit || '-'
    case 'purchasePrice': return fmtPrice(row.purchasePrice)
    case 'defaultPrice': return fmtPrice(row.defaultPrice)
    case 'retailPrice': return fmtPrice(row.retailPrice)
    case 'warehouse': return row.warehouseName || '-'
    case 'sortWeight': return row.mallSortWeight === undefined || row.mallSortWeight === null ? '-' : String(row.mallSortWeight)
    case 'remark': return row.description || '-'
    case 'totalStock': return row.totalStock === undefined || row.totalStock === null ? '-' : String(row.totalStock)
    case 'createdAt': return row.createdAt ? row.createdAt.slice(0, 10) : '-'
    case 'shelfStatus': return row.status === 1 ? '已启用' : '已禁用'
    case 'status': return row.status === 1 ? '已启用' : '已禁用'
    default: return '-'
  }
}

function destroyColumnSortable() {
  columnSortable?.destroy()
  columnSortable = null
}

function initColumnSortable() {
  destroyColumnSortable()
  if (!sortListRef.value) return
  columnSortable = Sortable.create(sortListRef.value, {
    handle: '.drag-handle',
    animation: 150,
    onEnd: (evt) => {
      const { oldIndex, newIndex } = evt
      if (oldIndex === undefined || newIndex === undefined || oldIndex === newIndex) return
      const rows = [...draftKeys.value]
      const [moved] = rows.splice(oldIndex, 1)
      if (moved === undefined) return
      rows.splice(newIndex, 0, moved)
      draftKeys.value = rows
    },
  })
}

function openColumnDialog() {
  draftKeys.value = [...columnKeys.value]
  draftFreeze.value = freezeCount.value
  columnDialogVisible.value = true
  nextTick(initColumnSortable)
}

function toggleDraftColumn(key: string, checked: boolean) {
  if (checked) {
    if (!draftKeys.value.includes(key)) draftKeys.value.push(key)
  } else {
    draftKeys.value = draftKeys.value.filter((k) => k !== key)
  }
}

function removeDraftColumn(key: string) {
  if (isRequiredColumn(key)) return
  draftKeys.value = draftKeys.value.filter((k) => k !== key)
}

async function saveColumnConfig() {
  try {
    await saveSettings({
      [COLUMN_SETTING_KEY]: JSON.stringify({ columns: columnKeys.value, freeze: freezeCount.value }),
    })
  } catch {
    // 拦截器已提示
  }
}

async function applyColumns() {
  columnKeys.value = [...draftKeys.value]
  freezeCount.value = draftFreeze.value
  columnDialogVisible.value = false
  await saveColumnConfig()
}

async function resetColumns() {
  columnKeys.value = [...DEFAULT_COLUMNS]
  freezeCount.value = DEFAULT_FREEZE
  draftKeys.value = [...DEFAULT_COLUMNS]
  draftFreeze.value = DEFAULT_FREEZE
  columnDialogVisible.value = false
  await saveColumnConfig()
}

async function loadColumnConfig() {
  try {
    const res = await fetchSettings()
    if (res.data.code === 0 || res.data.code === 200) {
      const raw = res.data.data?.[COLUMN_SETTING_KEY]
      if (!raw) return
      const parsed = JSON.parse(raw)
      const validKeys = new Set(ALL_COLUMNS.map((c) => c.key))
      const keys = Array.isArray(parsed.columns)
        ? parsed.columns.filter((k: unknown): k is string => typeof k === 'string' && validKeys.has(k))
        : []
      for (const c of ALL_COLUMNS) {
        if (c.required && !keys.includes(c.key)) keys.push(c.key)
      }
      if (keys.length) columnKeys.value = keys
      const freeze = Number(parsed.freeze)
      if (Number.isFinite(freeze)) freezeCount.value = Math.min(5, Math.max(0, Math.floor(freeze)))
    }
  } catch {
    // 配置缺失或解析失败时使用默认
  }
}

interface FilterItemConf {
  key: string
  show: boolean
  width: 1 | 2 | 3
}

const ALL_FILTERS = [
  { key: 'keyword', label: '搜索' },
  { key: 'name', label: '商品名称' },
  { key: 'category', label: '商品分类' },
  { key: 'tag', label: '商品标签' },
  { key: 'topic', label: '专题分类' },
  { key: 'brand', label: '商品品牌' },
  { key: 'image', label: '商品图片' },
  { key: 'status', label: '商品状态' },
  { key: 'stock', label: '库存状态' },
]

const FILTER_SETTING_KEY = 'productList.filterConfig'

function defaultFilterConf(): FilterItemConf[] {
  const shown = new Set(['keyword', 'tag', 'image', 'status'])
  return ALL_FILTERS.map((f) => ({ key: f.key, show: shown.has(f.key), width: 1 }))
}

const filterConf = ref<FilterItemConf[]>(defaultFilterConf())
const filterDialogVisible = ref(false)
const draftFilterConf = ref<FilterItemConf[]>([])
const filterDialogKeyword = ref('')
const filterTableRef = ref()
let filterSortable: Sortable | null = null

const filteredDraftFilterConf = computed(() => {
  const kw = filterDialogKeyword.value.trim().toLowerCase()
  if (!kw) return draftFilterConf.value
  return draftFilterConf.value.filter((c) => filterLabel(c.key).toLowerCase().includes(kw))
})

const draftAllShown = computed(() => draftFilterConf.value.length > 0 && draftFilterConf.value.every((c) => c.show))

const shownFilterConfs = computed(() => filterConf.value.filter((c) => c.show && c.key !== 'keyword'))
const keywordFilterConf = computed<FilterItemConf>(() => filterConf.value.find((c) => c.key === 'keyword') ?? { key: 'keyword', show: true, width: 1 })

function filterLabel(key: string) {
  return ALL_FILTERS.find((f) => f.key === key)?.label ?? key
}

function filterCellStyle(conf: FilterItemConf) {
  return { flex: `${conf.width} 1 0`, minWidth: `${140 * conf.width}px` }
}

function toggleDraftShowAll(checked: boolean) {
  draftFilterConf.value.forEach((c) => {
    c.show = c.key === 'keyword' ? true : checked
  })
}

function destroyFilterSortable() {
  filterSortable?.destroy()
  filterSortable = null
}

function initFilterSortable() {
  destroyFilterSortable()
  if (filterDialogKeyword.value.trim()) return
  const tbody = filterTableRef.value?.$el?.querySelector('.el-table__body tbody')
  if (!tbody) return
  filterSortable = Sortable.create(tbody, {
    handle: '.drag-handle',
    animation: 150,
    onEnd: (evt) => {
      const { oldIndex, newIndex } = evt
      if (oldIndex === undefined || newIndex === undefined || oldIndex === newIndex) return
      const rows = [...draftFilterConf.value]
      const [moved] = rows.splice(oldIndex, 1)
      if (moved === undefined) return
      rows.splice(newIndex, 0, moved)
      draftFilterConf.value = rows
    },
  })
}

watch(filterDialogKeyword, () => {
  nextTick(initFilterSortable)
})

function openFilterDialog() {
  draftFilterConf.value = filterConf.value.map((c) => ({ ...c }))
  filterDialogKeyword.value = ''
  filterDialogVisible.value = true
  nextTick(initFilterSortable)
}

function clearDisabledFilterValues() {
  const hidden = new Set(filterConf.value.filter((c) => !c.show).map((c) => c.key))
  if (hidden.has('name')) filters.name = ''
  if (hidden.has('category')) {
    filters.categoryId = undefined
    sideTreeRef.value?.clearCurrentKey()
  }
  if (hidden.has('tag')) filters.tagId = undefined
  if (hidden.has('topic')) filters.topicCategory = ''
  if (hidden.has('brand')) filters.brandId = undefined
  if (hidden.has('image')) filters.hasImage = ''
  if (hidden.has('status')) filters.status = ''
  if (hidden.has('stock')) filters.stockStatus = ''
}

async function saveFilterConfig() {
  try {
    await saveSettings({ [FILTER_SETTING_KEY]: JSON.stringify(filterConf.value) })
  } catch {
    // 拦截器已提示
  }
}

async function applyFilterConfig() {
  const conf = draftFilterConf.value.map((c) => ({ ...c }))
  const keyword = conf.find((c) => c.key === 'keyword')
  if (keyword) keyword.show = true
  filterConf.value = conf
  filterDialogVisible.value = false
  clearDisabledFilterValues()
  pagination.page = 1
  fetchList()
  await saveFilterConfig()
}

function resetFilterConfig() {
  draftFilterConf.value = defaultFilterConf()
  nextTick(initFilterSortable)
}

function normalizeFilterConf(parsed: unknown): FilterItemConf[] | null {
  if (!Array.isArray(parsed) || parsed.length === 0) return null
  if (typeof parsed[0] !== 'object' || parsed[0] === null) return null
  const valid = new Set(ALL_FILTERS.map((f) => f.key))
  const result: FilterItemConf[] = []
  for (const item of parsed as Record<string, unknown>[]) {
    if (!item || typeof item !== 'object') continue
    const key = item.key
    if (typeof key !== 'string' || !valid.has(key) || result.some((r) => r.key === key)) continue
    const width = Number(item.width)
    result.push({
      key,
      show: key === 'keyword' ? true : !!item.show,
      width: ([1, 2, 3].includes(width) ? width : 1) as 1 | 2 | 3,
    })
  }
  for (const f of ALL_FILTERS) {
    if (!result.some((r) => r.key === f.key)) result.push({ key: f.key, show: f.key === 'keyword', width: 1 })
  }
  return result
}

async function loadFilterConfig() {
  try {
    const res = await fetchSettings()
    if (res.data.code === 0 || res.data.code === 200) {
      const raw = res.data.data?.[FILTER_SETTING_KEY]
      if (!raw) return
      const conf = normalizeFilterConf(JSON.parse(raw))
      if (!conf) return
      filterConf.value = conf
      clearDisabledFilterValues()
    }
  } catch {
    // 配置缺失或解析失败时使用默认
  }
}

function handleFilterCategoryChange(val?: number) {
  if (!val) {
    filters.categoryId = undefined
    sideTreeRef.value?.clearCurrentKey()
  } else {
    sideTreeRef.value?.setCurrentKey(val)
  }
  pagination.page = 1
  fetchList()
}

function buildParams() {
  return {
    keyword: filters.keyword || undefined,
    name: filters.name || undefined,
    categoryId: filters.categoryId,
    brandId: filters.brandId,
    tagId: filters.tagId,
    topicCategory: filters.topicCategory || undefined,
    stockStatus: filters.stockStatus || undefined,
    hasImage: filters.hasImage === '' ? undefined : filters.hasImage,
    status: filters.status === '' ? undefined : filters.status,
    orderBy: sortState.orderBy,
    order: sortState.order,
    page: pagination.page,
    pageSize: pagination.pageSize,
  }
}

async function fetchList() {
  loading.value = true
  try {
    const params = buildParams()
    const res = mode.value === 'product' ? await fetchProducts(params) : await fetchSpecItems(params)
    if (res.data.code === 0 || res.data.code === 200) {
      const data = res.data.data
      const rows = Array.isArray(data) ? data : (data?.list ?? [])
      total.value = data?.total ?? rows.length
      if (mode.value === 'product') {
        productList.value = rows
      } else {
        specList.value = rows
      }
    }
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

function handleSizeChange(size: number) {
  pagination.pageSize = size
  pagination.page = 1
  fetchList()
}

function handleCurrentChange(page: number) {
  pagination.page = page
  fetchList()
}

function handleCategoryClick(node: CategoryNode) {
  filters.categoryId = node.id
  pagination.page = 1
  fetchList()
}

function handleCommand(cmd: string, row: ProductListItem) {
  if (cmd === 'view') {
    goDetail(row.id)
  } else if (cmd === 'edit') {
    goEdit(row.id)
  } else if (cmd === 'delete') {
    handleDelete(row)
  }
}

async function handleDelete(row: ProductListItem) {
  try {
    await ElMessageBox.confirm(`确认删除商品「${row.name}」？`, '提示', { type: 'warning' })
  } catch {
    return
  }
  try {
    const res = await deleteProduct(row.id)
    if (res.data.code === 0 || res.data.code === 200) {
      ElMessage.success('删除成功')
      fetchList()
    } else {
      ElMessage.error(res.data.message || '删除失败')
    }
  } catch {
    // 拦截器已提示
  }
}

const importVisible = ref(false)
const importStep = ref(1)
const importMode = ref<'system' | 'custom'>('system')
const importFile = ref<File | null>(null)
const importFileList = ref<UploadFile[]>([])
const overwriteEmpty = ref(false)
const importing = ref(false)
const importResult = ref<ImportResult | null>(null)

const importModeOptions = [
  {
    value: 'system' as const,
    label: '系统模板导入',
    desc: ['①该方式只能处理新商品导入。', '②新商品导入时，商品编号等必填字段不允许为空。'],
  },
  {
    value: 'custom' as const,
    label: '自定义模板导入',
    desc: [
      '①该方式可以同时处理新商品导入和商品资料更新。商品编号已存在时更新该商品；编号不存在或为空时将自动新增商品（商品名称、基本单位为必填）。',
      '②系统会根据字段名称逐一匹配后导入信息。没有匹配的字段不导入。',
      '③商品更新时，仅更新已匹配的字段，没有匹配的字段将不更新。更新字段时，原内容将全部被覆盖（空信息覆盖请勾选配置项）。请谨慎操作！',
    ],
  },
]

function openImport() {
  importVisible.value = true
}

function resetImport() {
  importStep.value = 1
  importMode.value = 'system'
  importFile.value = null
  importFileList.value = []
  overwriteEmpty.value = false
  importing.value = false
  importResult.value = null
}

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
    const res = importMode.value === 'system'
      ? await importProductsSystem(importFile.value)
      : await importProductsCustom(importFile.value, overwriteEmpty.value)
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
  fetchList()
}

function goCreate() { router.push('/products/new') }
function goDetail(id: number) { router.push(`/products/${id}`) }
function goEdit(id: number) { router.push(`/products/${id}?mode=edit`) }

onMounted(async () => {
  loadColumnConfig()
  loadFilterConfig()
  const categoryId = Number(route.query.categoryId)
  if (categoryId) {
    filters.categoryId = categoryId
  }
  const brandId = Number(route.query.brandId)
  if (brandId) {
    filters.brandId = brandId
  }
  const tagId = Number(route.query.tagId)
  if (tagId) {
    filters.tagId = tagId
  }
  fetchList()
  try {
    const res = await fetchCategoryTree()
    if (res.data.code === 0 || res.data.code === 200) {
      const nodes = Array.isArray(res.data.data) ? res.data.data : []
      categoryTree.value = normalizeCategoryNodes(nodes)
      if (filters.categoryId) {
        sideTreeRef.value?.setCurrentKey(filters.categoryId)
      }
    }
  } catch {
    // 拦截器已提示
  }
  try {
    const res = await fetchTags()
    if (res.data.code === 0 || res.data.code === 200) {
      const data = res.data.data
      tagOptions.value = Array.isArray(data) ? data : (data?.list ?? [])
    }
  } catch {
    // 拦截器已提示
  }
  try {
    const res = await fetchBrandOptions()
    if (res.data.code === 0 || res.data.code === 200) {
      const data = res.data.data
      brandOptions.value = Array.isArray(data) ? data : (data?.list ?? [])
    }
  } catch {
    // 拦截器已提示
  }
  try {
    const res = await fetchTopicCategories()
    if (res.data.code === 0 || res.data.code === 200) {
      topicOptions.value = Array.isArray(res.data.data) ? res.data.data : []
    }
  } catch {
    // 拦截器已提示
  }
})

onBeforeUnmount(() => {
  destroyColumnSortable()
  destroyFilterSortable()
})
</script>

<style scoped>
.page { padding: 20px; height: 100%; box-sizing: border-box; display: flex; flex-direction: column; overflow: hidden; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; flex-shrink: 0; }
.page-title { margin: 0; font-size: 20px; }
.page-actions { display: flex; gap: 8px; }
.page-body { flex: 1; min-height: 0; display: flex; gap: 16px; align-items: stretch; }
.category-panel {
  width: 200px;
  flex-shrink: 0;
  border: 1px solid var(--el-border-color);
  border-radius: 4px;
  padding: 12px;
  background: var(--el-bg-color);
  position: relative;
  transition: width 0.2s ease, padding 0.2s ease;
  display: flex;
  flex-direction: column;
}
.category-panel-content { flex: 1; min-height: 0; display: flex; flex-direction: column; }
.category-panel-content :deep(.el-tree) { flex: 1; min-height: 0; overflow-y: auto; }
.category-panel.collapsed {
  width: 30px;
  padding: 0;
  border-color: transparent;
  overflow: visible;
  align-self: stretch;
}
.panel-collapse-handle {
  position: absolute;
  right: -14px;
  top: 50%;
  transform: translateY(-50%);
  width: 14px;
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--el-color-primary-light-8);
  color: var(--el-color-primary);
  cursor: pointer;
  border-radius: 0 4px 4px 0;
  z-index: 2;
}
.panel-collapse-handle:hover { background: var(--el-color-primary-light-7); }
.panel-expand-strip {
  width: 30px;
  height: 100%;
  min-height: 200px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  background: var(--el-color-primary-light-8);
  color: var(--el-color-primary);
  cursor: pointer;
  border-radius: 4px;
}
.panel-expand-strip:hover { background: var(--el-color-primary-light-7); }
.panel-expand-text { writing-mode: vertical-lr; letter-spacing: 4px; font-size: 13px; }
.list-panel { flex: 1; min-width: 0; min-height: 0; display: flex; flex-direction: column; }
.mode-tabs { margin-bottom: 12px; flex-shrink: 0; }
.table-wrap { flex: 1; min-height: 0; }
.filter-row { display: flex; gap: 8px; margin-bottom: 12px; flex-wrap: wrap; flex-shrink: 0; }
.filter-cell { min-width: 140px; }
.filter-control { width: 100%; }
.filter-actions { margin-left: auto; display: flex; gap: 8px; }
.filter-dialog-search { width: 280px; margin-bottom: 12px; }
.thumb { width: 40px; height: 40px; border-radius: 4px; display: flex; align-items: center; justify-content: center; }
.thumb-placeholder { background: var(--el-fill-color-light); color: var(--el-text-color-secondary); }
.pagination-bar { display: flex; justify-content: space-between; align-items: center; margin-top: 16px; flex-shrink: 0; }
.pagination-info { color: var(--el-text-color-secondary); font-size: 13px; }
.pagination-controls { display: flex; align-items: center; gap: 12px; }
.import-modes { display: flex; flex-direction: column; gap: 12px; }
.import-mode-card {
  border: 1px solid var(--el-border-color);
  border-radius: 6px;
  padding: 12px 16px;
  cursor: pointer;
}
.import-mode-card.active { border-color: var(--el-color-primary); background: var(--el-color-primary-light-9); }
.import-mode-radio { font-weight: 600; }
.import-mode-desc { margin: 4px 0 0 24px; color: var(--el-text-color-secondary); font-size: 13px; }
.import-mode-desc p { margin: 4px 0; }
.import-notice { color: var(--el-color-danger); background: var(--el-color-danger-light-9); padding: 8px 12px; border-radius: 4px; margin: 4px 0 0; }
.overwrite-checkbox { margin-top: 12px; }
.import-summary { display: flex; gap: 12px; margin-bottom: 12px; }
.col-setting-icon { cursor: pointer; color: var(--el-text-color-secondary); vertical-align: middle; }
.col-setting-icon:hover { color: var(--el-color-primary); }
:deep(.expand-content-col) { padding: 0 !important; }
:deep(.expand-content-col .el-table__expand-icon) { display: none; }
.index-cell { display: flex; align-items: center; justify-content: center; cursor: pointer; min-height: 23px; }
.index-cell .expand-arrow { display: none; color: var(--el-color-primary); }
.index-cell:hover .index-num { display: none; }
.index-cell:hover .expand-arrow { display: inline-flex; }
.spec-detail { padding: 8px 24px; min-height: 40px; }
.spec-empty { color: var(--el-text-color-secondary); padding: 8px 0; }
.brand-header { display: inline-flex; align-items: center; gap: 2px; }
.brand-filter-icon { cursor: pointer; color: var(--el-text-color-secondary); vertical-align: middle; }
.brand-filter-icon.active { color: var(--el-color-primary); }
.column-dialog-body { display: flex; gap: 24px; }
.column-options { flex: 3; min-width: 0; }
.column-selected { flex: 2; min-width: 0; display: flex; flex-direction: column; }
.column-panel-title { font-weight: 600; margin-bottom: 12px; }
.column-checkbox-grid { display: flex; flex-wrap: wrap; }
.column-checkbox { width: 25%; margin-right: 0; }
.column-selected-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; }
.column-selected-header .column-panel-title { margin-bottom: 0; }
.freeze-setting { display: inline-flex; align-items: center; gap: 8px; font-size: 13px; color: var(--el-text-color-regular); }
.freeze-select { width: 80px; }
.column-sort-list {
  border: 1px solid var(--el-border-color);
  border-radius: 4px;
  max-height: 420px;
  overflow-y: auto;
}
.column-sort-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border-bottom: 1px solid var(--el-border-color-lighter);
}
.column-sort-item:last-child { border-bottom: none; }
.column-sort-item .drag-handle { cursor: move; color: var(--el-text-color-secondary); }
.column-sort-label { flex: 1; }
.column-remove { cursor: pointer; color: var(--el-text-color-secondary); }
.column-remove:hover { color: var(--el-color-danger); }
.column-dialog-footer { display: flex; justify-content: space-between; align-items: center; width: 100%; }
</style>

<style>
.brand-filter-popper .el-dropdown-menu { max-height: 320px; overflow-y: auto; }
.brand-filter-popper .el-dropdown-menu__item.active { color: var(--el-color-primary); font-weight: 600; }
</style>
