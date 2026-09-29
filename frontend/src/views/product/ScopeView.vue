<template>
  <div class="page">
    <div class="page-header">
      <span class="page-title">商品销售范围</span>
      <div class="header-right">
        <span class="dim-text">当前设置维度：{{ dimText }}</span>
        <el-button type="primary" @click="openDimensionDialog">设置维度</el-button>
      </div>
    </div>
    <el-card>
      <div class="search-row">
        <el-input
          v-model="keyword"
          placeholder="商品名称/编号/条码/关键字"
          clearable
          class="search-input"
          @keyup.enter="handleSearch"
        />
        <el-button type="primary" @click="handleSearch">搜索</el-button>
      </div>
      <el-table
        :key="tableKey"
        :data="tableData"
        v-loading="loading"
        row-key="key"
        lazy
        :load="loadChildren"
        :tree-props="{ children: 'children', hasChildren: 'hasChildren' }"
        border
      >
        <el-table-column type="selection" width="45" align="center" />
        <el-table-column label="操作" width="70" align="center">
          <template #default="{ row }">
            <el-button type="primary" link @click="openRuleDialog(row)">设置</el-button>
          </template>
        </el-table-column>
        <el-table-column label="商品分类/商品" min-width="240">
          <template #default="{ row }">
            <div class="target-name">
              <el-icon v-if="row.targetType !== 'product'" class="folder-icon"><Folder /></el-icon>
              <el-icon v-else class="product-icon"><Goods /></el-icon>
              <span>{{ row.name }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="可销售级别价" min-width="130" show-overflow-tooltip>
          <template #default="{ row }">{{ row.ruleSummary?.levelPrices || '不限' }}</template>
        </el-table-column>
        <el-table-column label="可销售客户分类" min-width="130" show-overflow-tooltip>
          <template #default="{ row }">{{ row.ruleSummary?.customerCategories || '不限' }}</template>
        </el-table-column>
        <el-table-column label="可销售客户销售区域" min-width="140" show-overflow-tooltip>
          <template #default="{ row }">{{ row.ruleSummary?.regions || '不限' }}</template>
        </el-table-column>
        <el-table-column label="可销售客户标签" min-width="130" show-overflow-tooltip>
          <template #default="{ row }">{{ row.ruleSummary?.customerTags || '不限' }}</template>
        </el-table-column>
        <el-table-column label="可销售独立客户" min-width="130" show-overflow-tooltip>
          <template #default="{ row }">{{ row.ruleSummary?.customers || '不限' }}</template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dimensionDialogVisible" title="设置" width="480px">
      <el-alert
        type="warning"
        show-icon
        closable
        title="切换后现有商品销售范围将失效，将按新的维度重新设置"
        class="dim-alert"
      />
      <div class="dim-tabs">
        <div
          class="dim-tab"
          :class="{ active: dimForm.tab === 'product' }"
          @click="dimForm.tab = 'product'"
        >按商品设置可销售客户</div>
        <div
          class="dim-tab"
          :class="{ active: dimForm.tab === 'customer' }"
          @click="dimForm.tab = 'customer'"
        >按客户设置可销售商品</div>
      </div>
      <div v-if="dimForm.tab === 'product'" class="dim-body">
        <el-radio-group v-model="dimForm.productDim">
          <el-radio value="category">按商品分类设置</el-radio>
          <el-radio value="brand">按商品品牌设置</el-radio>
          <el-radio value="supplier">按默认供应商设置</el-radio>
        </el-radio-group>
      </div>
      <div v-else class="dim-body">
        <el-radio-group v-model="dimForm.customerDim">
          <el-radio value="levelPrice">按级别价设置</el-radio>
          <el-radio value="customerCategory">按客户分类设置</el-radio>
          <el-radio value="region">按客户销售区域设置</el-radio>
          <el-radio value="customerTag">按客户标签设置</el-radio>
        </el-radio-group>
      </div>
      <template #footer>
        <el-button @click="dimensionDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="dimensionSaving" @click="handleDimensionSubmit">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="ruleDialogVisible" :title="`设置可销售范围 - ${editingRow?.name || ''}`" width="640px">
      <div class="rule-section">
        <div class="rule-title">可销售级别价<span class="rule-hint">不选即不限</span></div>
        <el-select v-model="ruleForm.levelPriceIds" multiple filterable placeholder="不限" class="rule-select">
          <el-option v-for="item in options.levelPrices" :key="item.id" :label="item.name" :value="item.id" />
        </el-select>
      </div>
      <div class="rule-section">
        <div class="rule-title">可销售客户分类<span class="rule-hint">不选即不限</span></div>
        <el-tree-select
          v-model="ruleForm.customerCategoryIds"
          :data="options.customerCategories"
          multiple
          show-checkbox
          check-strictly
          node-key="id"
          :props="{ label: 'name', children: 'children' }"
          placeholder="不限"
          class="rule-select"
        />
      </div>
      <div class="rule-section">
        <div class="rule-title">可销售客户销售区域<span class="rule-hint">不选即不限</span></div>
        <el-tree-select
          v-model="ruleForm.regionIds"
          :data="options.regions"
          multiple
          show-checkbox
          check-strictly
          node-key="id"
          :props="{ label: 'name', children: 'children' }"
          placeholder="不限"
          class="rule-select"
        />
      </div>
      <div class="rule-section">
        <div class="rule-title">可销售客户标签<span class="rule-hint">不选即不限</span></div>
        <el-select v-model="ruleForm.customerTagIds" multiple filterable placeholder="不限" class="rule-select">
          <el-option v-for="item in options.customerTags" :key="item.id" :label="item.name" :value="item.id" />
        </el-select>
      </div>
      <div class="rule-section">
        <div class="rule-title">可销售独立客户<span class="rule-hint">不选即不限</span></div>
        <el-select v-model="ruleForm.customerIds" multiple filterable placeholder="不限" class="rule-select">
          <el-option v-for="item in options.customers" :key="item.id" :label="item.name" :value="item.id" />
        </el-select>
      </div>
      <template #footer>
        <el-button @click="ruleDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="ruleSaving" @click="handleRuleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Folder, Goods } from '@element-plus/icons-vue'
import api from '@/api/client'
import {
  emptyRule,
  fetchSaleRuleProducts,
  fetchSaleRuleTree,
  getSaleScopeDimension,
  saveSaleRule,
  setSaleScopeDimension,
  type CustomerDim,
  type ProductDim,
  type SaleRuleContent,
  type SaleRuleRow,
  type SaleRuleSummary,
} from '@/api/saleScope'

interface TableRow extends SaleRuleRow {
  key: string
  children?: TableRow[]
}

const productDimLabels: Record<ProductDim, string> = {
  category: '按商品分类设置可销售客户',
  brand: '按商品品牌设置可销售客户',
  supplier: '按默认供应商设置可销售客户',
}

const dimension = reactive<{ productDim: ProductDim; customerDim: CustomerDim }>({
  productDim: 'category',
  customerDim: 'levelPrice',
})
const dimText = computed(() => productDimLabels[dimension.productDim])

const keyword = ref('')
const loading = ref(false)
const tableData = ref<TableRow[]>([])
const tableKey = ref(0)

function mapRow(row: SaleRuleRow): TableRow {
  return {
    ...row,
    key: `${row.targetType}-${row.targetId}`,
    children: row.children?.map(mapRow),
  }
}

async function loadDimension() {
  const res = await getSaleScopeDimension()
  const data = res.data.data
  dimension.productDim = data.productDim
  dimension.customerDim = data.customerDim
}

async function loadTree() {
  loading.value = true
  try {
    const res = await fetchSaleRuleTree()
    tableData.value = (res.data.data || []).map(mapRow)
    tableKey.value++
  } finally {
    loading.value = false
  }
}

async function handleSearch() {
  const kw = keyword.value.trim()
  if (!kw) {
    loadTree()
    return
  }
  loading.value = true
  try {
    const res = await fetchSaleRuleProducts({ keyword: kw })
    tableData.value = (res.data.data || []).map(mapRow)
    tableKey.value++
  } finally {
    loading.value = false
  }
}

async function loadChildren(row: TableRow, _treeNode: unknown, resolve: (rows: TableRow[]) => void) {
  if (row.targetType !== 'category') {
    resolve([])
    return
  }
  try {
    const [subRes, prodRes] = await Promise.all([
      fetchSaleRuleTree({ parentId: row.targetId }),
      fetchSaleRuleProducts({ categoryId: row.targetId }),
    ])
    const subs = (subRes.data.data || []).map(mapRow) as TableRow[]
    const products = (prodRes.data.data || []).map(mapRow) as TableRow[]
    resolve([...subs, ...products])
  } catch {
    resolve([])
  }
}

// ==================== 设置维度弹窗 ====================

const dimensionDialogVisible = ref(false)
const dimensionSaving = ref(false)
const dimForm = reactive<{ tab: 'product' | 'customer'; productDim: ProductDim; customerDim: CustomerDim }>({
  tab: 'product',
  productDim: 'category',
  customerDim: 'levelPrice',
})

function openDimensionDialog() {
  dimForm.tab = 'product'
  dimForm.productDim = dimension.productDim
  dimForm.customerDim = dimension.customerDim
  dimensionDialogVisible.value = true
}

async function handleDimensionSubmit() {
  const payload = {
    productDim: dimForm.tab === 'product' ? dimForm.productDim : dimension.productDim,
    customerDim: dimForm.tab === 'customer' ? dimForm.customerDim : dimension.customerDim,
  }
  if (payload.productDim === dimension.productDim && payload.customerDim === dimension.customerDim) {
    dimensionDialogVisible.value = false
    return
  }
  try {
    await ElMessageBox.confirm('切换后现有商品销售范围将失效，将按新的维度重新设置，确定切换？', '提示', {
      type: 'warning',
      confirmButtonText: '确定',
      cancelButtonText: '取消',
    })
  } catch {
    return
  }
  dimensionSaving.value = true
  try {
    await setSaleScopeDimension(payload)
    dimension.productDim = payload.productDim
    dimension.customerDim = payload.customerDim
    dimensionDialogVisible.value = false
    ElMessage.success('设置成功')
    keyword.value = ''
    loadTree()
  } finally {
    dimensionSaving.value = false
  }
}

// ==================== 行内设置弹窗 ====================

const ruleDialogVisible = ref(false)
const ruleSaving = ref(false)
const editingRow = ref<TableRow | null>(null)
const ruleForm = reactive<SaleRuleContent>(emptyRule())

interface TreeNode { id: number; name: string; parentId?: number; children?: TreeNode[] }

const options = reactive<{
  levelPrices: TreeNode[]
  customerCategories: TreeNode[]
  regions: TreeNode[]
  customerTags: TreeNode[]
  customers: TreeNode[]
}>({
  levelPrices: [],
  customerCategories: [],
  regions: [],
  customerTags: [],
  customers: [],
})

function buildTree(list: TreeNode[]): TreeNode[] {
  const map = new Map<number, TreeNode>()
  list.forEach((item) => map.set(item.id, { ...item, children: [] }))
  const roots: TreeNode[] = []
  map.forEach((node) => {
    if (node.parentId && map.has(node.parentId)) {
      map.get(node.parentId)!.children!.push(node)
    } else {
      roots.push(node)
    }
  })
  const prune = (nodes: TreeNode[]): TreeNode[] =>
    nodes.map((n) => (n.children && n.children.length ? { ...n, children: prune(n.children) } : { ...n, children: undefined }))
  return prune(roots)
}

let optionsLoaded = false
async function loadOptions() {
  if (optionsLoaded) return
  const [levels, categories, regions, tags, customers] = await Promise.all([
    api.get('/api/v1/prices/levels', { params: { page: 1, pageSize: 500 } }),
    api.get('/api/v1/customers/categories', { params: { page: 1, pageSize: 500 } }),
    api.get('/api/v1/customers/regions', { params: { page: 1, pageSize: 500 } }),
    api.get('/api/v1/customers/tags'),
    api.get('/api/v1/customers', { params: { page: 1, pageSize: 500 } }),
  ])
  options.levelPrices = levels.data.data.list
  options.customerCategories = buildTree(categories.data.data.list)
  options.regions = buildTree(regions.data.data.list)
  options.customerTags = tags.data.data
  options.customers = customers.data.data.list
  optionsLoaded = true
}

async function openRuleDialog(row: TableRow) {
  editingRow.value = row
  const rule = row.rule || emptyRule()
  ruleForm.levelPriceIds = [...rule.levelPriceIds]
  ruleForm.customerCategoryIds = [...rule.customerCategoryIds]
  ruleForm.regionIds = [...rule.regionIds]
  ruleForm.customerTagIds = [...rule.customerTagIds]
  ruleForm.customerIds = [...rule.customerIds]
  ruleDialogVisible.value = true
  loadOptions()
}

async function handleRuleSubmit() {
  const row = editingRow.value
  if (!row) return
  ruleSaving.value = true
  try {
    const rule: SaleRuleContent = {
      levelPriceIds: ruleForm.levelPriceIds,
      customerCategoryIds: ruleForm.customerCategoryIds,
      regionIds: ruleForm.regionIds,
      customerTagIds: ruleForm.customerTagIds,
      customerIds: ruleForm.customerIds,
    }
    const res = await saveSaleRule({ targetType: row.targetType, targetId: row.targetId, rule })
    const isEmpty = Object.values(rule).every((ids) => ids.length === 0)
    row.rule = isEmpty ? null : rule
    row.ruleSummary = isEmpty ? null : (res.data.data?.ruleSummary as SaleRuleSummary)
    ruleDialogVisible.value = false
    ElMessage.success('保存成功')
  } finally {
    ruleSaving.value = false
  }
}

onMounted(async () => {
  await loadDimension()
  loadTree()
})
</script>

<style scoped>
.page { padding: 20px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.page-title { font-size: 18px; font-weight: 600; }
.header-right { display: flex; align-items: center; gap: 12px; }
.dim-text { color: #606266; font-size: 14px; }
.search-row { display: flex; gap: 12px; margin-bottom: 16px; }
.search-input { width: 280px; }
.target-name { display: flex; align-items: center; gap: 6px; }
.folder-icon { color: #e6a23c; }
.product-icon { color: #909399; }
.dim-alert { margin-bottom: 16px; }
.dim-tabs { display: flex; gap: 12px; margin-bottom: 16px; }
.dim-tab {
  flex: 1;
  text-align: center;
  padding: 10px 0;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  cursor: pointer;
  color: #606266;
}
.dim-tab.active { border-color: var(--el-color-primary); color: var(--el-color-primary); background: var(--el-color-primary-light-9); }
.dim-body :deep(.el-radio-group) { display: flex; flex-direction: column; gap: 8px; }
.dim-body :deep(.el-radio) { margin-right: 0; }
.rule-section { margin-bottom: 16px; }
.rule-title { font-weight: 600; margin-bottom: 8px; }
.rule-hint { font-weight: 400; color: #909399; font-size: 12px; margin-left: 8px; }
.rule-select { width: 100%; }
</style>
