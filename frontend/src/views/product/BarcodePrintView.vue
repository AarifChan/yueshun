<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">条码打印</h2>
      <div class="actions">
        <el-button @click="billDialogVisible = true">引入单据打印</el-button>
        <el-button @click="templateDialogVisible = true">设置模板</el-button>
        <el-button @click="printByMode('unit')">打印单位条码</el-button>
        <el-button @click="printByMode('specUnit')">打印规格单位条码</el-button>
        <el-button @click="printByMode('spec')">打印规格条码（Ctrl+P）</el-button>
        <el-button @click="clearAll">清空</el-button>
        <el-button type="primary" :disabled="!printItems.length" @click="doPrint">打印</el-button>
      </div>
    </div>

    <el-card>
      <el-form inline @submit.prevent>
        <el-form-item label="扫码快速录入">
          <el-input
            ref="scanInputRef"
            v-model="scanCode"
            placeholder="扫码或输入条码/编号后回车（Ctrl+F1 聚焦）"
            clearable
            style="width: 280px"
            @keyup.enter="quickAdd"
          />
        </el-form-item>
        <el-form-item label="商品">
          <el-select v-model="pickId" filterable remote :remote-method="searchProducts" placeholder="输入商品名称/编号搜索" style="width: 280px" @change="addProduct">
            <el-option v-for="p in productOptions" :key="p.id" :label="`${p.name}（${p.code || '无编号'}）`" :value="p.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="标签内容">
          <el-checkbox v-model="labelOpts.showName">名称</el-checkbox>
          <el-checkbox v-model="labelOpts.showSpec">规格</el-checkbox>
          <el-checkbox v-model="labelOpts.showPrice">零售价</el-checkbox>
        </el-form-item>
      </el-form>

      <el-table :data="printItems" border stripe>
        <el-table-column type="index" width="50" />
        <el-table-column prop="name" label="商品名称" min-width="180" show-overflow-tooltip />
        <el-table-column prop="spec" label="规格" width="120" />
        <el-table-column prop="barcode" label="条码" width="150">
          <template #default="{ row }">
            <span :class="{ empty: !row.barcode }">{{ row.barcode || '未设置（将使用编号）' }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="retailPrice" label="零售价" width="100" align="right">
          <template #default="{ row }">{{ row.retailPrice?.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="打印数量" width="150">
          <template #default="{ row }">
            <el-input-number v-model="row.qty" :min="1" :max="999" size="small" />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="80" fixed="right">
          <template #default="{ $index }">
            <el-button link type="danger" @click="printItems.splice($index, 1)">移除</el-button>
          </template>
        </el-table-column>
        <template #empty>暂无数据</template>
      </el-table>
    </el-card>

    <!-- 设置模板 -->
    <el-dialog v-model="templateDialogVisible" title="设置模板" width="420px">
      <el-form label-width="90px">
        <el-form-item label="标签内容">
          <el-checkbox v-model="labelOpts.showName">名称</el-checkbox>
          <el-checkbox v-model="labelOpts.showSpec">规格</el-checkbox>
          <el-checkbox v-model="labelOpts.showPrice">零售价</el-checkbox>
        </el-form-item>
        <el-form-item label="标签宽度">
          <el-input-number v-model="labelOpts.width" :min="120" :max="400" :step="10" /> px
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button type="primary" @click="templateDialogVisible = false">确定</el-button>
      </template>
    </el-dialog>

    <!-- 引入单据打印 -->
    <el-dialog v-model="billDialogVisible" title="引入单据打印" width="720px" :close-on-click-modal="false">
      <el-form inline @submit.prevent>
        <el-form-item label="单据类型">
          <el-radio-group v-model="billType" @change="loadBills">
            <el-radio-button value="purchase-in">采购入库单</el-radio-button>
            <el-radio-button value="other-in">其他入库单</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item>
          <el-input v-model="billKeyword" placeholder="单号" clearable style="width: 180px" @keyup.enter="loadBills" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadBills">查询</el-button>
        </el-form-item>
      </el-form>
      <el-table :data="billList" v-loading="billLoading" border size="small" height="320" highlight-current-row @current-change="(row: BillRow | undefined) => (selectedBill = row)">
        <el-table-column prop="billNo" label="单号" min-width="150" />
        <el-table-column prop="billDate" label="日期" width="110" />
        <el-table-column prop="party" label="往来单位" min-width="140"><template #default="{ row }">{{ row.party || '-' }}</template></el-table-column>
        <el-table-column prop="itemCount" label="明细数" width="80" align="right" />
        <el-table-column prop="amount" label="金额" width="100" align="right"><template #default="{ row }">{{ row.amount?.toFixed(2) }}</template></el-table-column>
        <template #empty>暂无数据</template>
      </el-table>
      <template #footer>
        <el-button @click="billDialogVisible = false">取消</el-button>
        <el-button type="primary" :disabled="!selectedBill" :loading="importingBill" @click="importBill">引入</el-button>
      </template>
    </el-dialog>

    <!-- 打印预览区（打印时仅输出此区域） -->
    <div v-if="labels.length" class="print-area" id="barcode-print-area">
      <div v-for="(label, i) in labels" :key="i" class="label" :style="{ width: labelOpts.width + 'px' }">
        <div v-if="labelOpts.showName" class="label-name">{{ label.name }}</div>
        <div v-if="labelOpts.showSpec" class="label-spec">{{ label.spec }}</div>
        <svg :ref="(el) => setSvgRef(el, label.code + '-' + i)" class="label-barcode" :data-code="label.code"></svg>
        <div v-if="labelOpts.showPrice" class="label-price">￥{{ label.retailPrice.toFixed(2) }}</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import JsBarcode from 'jsbarcode'
import api from '@/api/client'

interface Product { id: number; name: string; code: string; barcode: string; spec?: string; specification?: string; retailPrice?: number; retail_price?: number }
interface PrintItem { id: number; name: string; spec: string; barcode: string; retailPrice: number; qty: number }
interface SpecRow { id: number; specValue: string; code: string; barcode: string }
interface BillRow { id: number; billNo: string; billDate: string; party: string; itemCount: number; amount: number }

type PrintMode = 'unit' | 'specUnit' | 'spec'

const productOptions = ref<Product[]>([])
const pickId = ref<number>()
const printItems = ref<PrintItem[]>([])
const labelOpts = reactive({ showName: true, showSpec: true, showPrice: true, width: 180 })
const printMode = ref<PrintMode>('unit')
const specCache = ref<Record<number, SpecRow[]>>({})
const svgRefs = new Map<string, SVGElement>()
const scanInputRef = ref()
const scanCode = ref('')

// ==================== 标签生成 ====================

const labels = computed(() => {
  const arr: { name: string; spec: string; code: string; retailPrice: number }[] = []
  for (const item of printItems.value) {
    if (printMode.value === 'unit') {
      for (let i = 0; i < item.qty; i++) {
        arr.push({ name: item.name, spec: item.spec, code: item.barcode, retailPrice: item.retailPrice })
      }
      continue
    }
    const specs = specCache.value[item.id] ?? []
    for (const s of specs) {
      const specText = printMode.value === 'specUnit' && item.spec ? `${s.specValue} / ${item.spec}` : s.specValue
      const code = s.barcode || s.code || item.barcode
      for (let i = 0; i < item.qty; i++) {
        arr.push({ name: item.name, spec: specText, code, retailPrice: item.retailPrice })
      }
    }
  }
  return arr
})

watch(labels, async () => {
  await nextTick()
  renderBarcodes()
}, { deep: true })

function setSvgRef(el: Element | any, key: string) {
  if (!el) return
  svgRefs.set(key, el as SVGElement)
  const code = (el as SVGElement).dataset?.code ?? ''
  renderOne(el as SVGElement, code)
}

function renderOne(el: SVGElement, code: string) {
  try {
    JsBarcode(el, code || 'EMPTY', { format: 'CODE128', width: 1.4, height: 40, fontSize: 12, margin: 2 })
  } catch {
    // 非法条码字符时忽略
  }
}

function renderBarcodes() {
  svgRefs.forEach((el) => renderOne(el, el.dataset?.code ?? ''))
}

async function ensureSpecs(productId: number) {
  if (specCache.value[productId]) return
  const res = await api.get(`/api/v1/products/${productId}`)
  if (res.data.code === 0 || res.data.code === 200) {
    specCache.value = { ...specCache.value, [productId]: res.data.data?.specItems ?? [] }
  }
}

// ==================== 打印 ====================

async function printByMode(mode: PrintMode) {
  if (!printItems.value.length) {
    ElMessage.warning('请先添加商品')
    return
  }
  if (mode !== 'unit') {
    for (const item of printItems.value) {
      await ensureSpecs(item.id)
    }
    const hasSpec = printItems.value.some((item) => (specCache.value[item.id] ?? []).length > 0)
    if (!hasSpec) {
      ElMessage.warning('所选商品没有规格行，无法打印规格条码')
      return
    }
  }
  printMode.value = mode
  await nextTick()
  renderBarcodes()
  window.print()
}

function doPrint() {
  printByMode('unit')
}

function clearAll() {
  printItems.value = []
}

// ==================== 商品选择 / 扫码录入 ====================

async function searchProducts(kw: string) {
  const res = await api.get('/api/v1/products', { params: { page: 1, pageSize: 50, keyword: kw } })
  if (res.data.code === 0 || res.data.code === 200) productOptions.value = res.data.data?.list ?? []
}

function pushItem(p: { id: number; name: string; spec?: string; barcode?: string; code?: string; retailPrice?: number }, qty = 1) {
  const exist = printItems.value.find((x) => x.id === p.id)
  if (exist) {
    exist.qty += qty
    return
  }
  printItems.value.push({
    id: p.id,
    name: p.name,
    spec: p.spec ?? '',
    barcode: p.barcode || p.code || String(p.id),
    retailPrice: p.retailPrice ?? 0,
    qty,
  })
}

function addProduct(id: number | undefined) {
  if (!id) return
  const p = productOptions.value.find((x) => x.id === id)
  pickId.value = undefined
  if (!p) return
  pushItem({ id: p.id, name: p.name, spec: p.spec ?? p.specification ?? '', barcode: p.barcode, code: p.code, retailPrice: p.retailPrice ?? p.retail_price ?? 0 })
}

async function quickAdd() {
  const kw = scanCode.value.trim()
  if (!kw) return
  const res = await api.get('/api/v1/products', { params: { page: 1, pageSize: 50, keyword: kw } })
  if (res.data.code === 0 || res.data.code === 200) {
    const list: Product[] = res.data.data?.list ?? []
    const hit = list.find((p) => p.barcode === kw || p.code === kw) ?? list[0]
    if (hit) {
      pushItem({ id: hit.id, name: hit.name, spec: hit.spec ?? hit.specification ?? '', barcode: hit.barcode, code: hit.code, retailPrice: hit.retailPrice ?? hit.retail_price ?? 0 })
      scanCode.value = ''
    } else {
      ElMessage.warning(`未找到条码/编号为「${kw}」的商品`)
    }
  }
}

function onGlobalKeydown(e: KeyboardEvent) {
  if (e.ctrlKey && e.key === 'F1') {
    e.preventDefault()
    scanInputRef.value?.focus()
  } else if (e.ctrlKey && (e.key === 'p' || e.key === 'P')) {
    e.preventDefault()
    printByMode('spec')
  }
}

// ==================== 引入单据打印 ====================

const billDialogVisible = ref(false)
const templateDialogVisible = ref(false)
const billType = ref<'purchase-in' | 'other-in'>('purchase-in')
const billKeyword = ref('')
const billList = ref<BillRow[]>([])
const billLoading = ref(false)
const selectedBill = ref<BillRow>()
const importingBill = ref(false)

async function loadBills() {
  billLoading.value = true
  try {
    const res = await api.get('/api/v1/product-barcodes/print-bills', { params: { type: billType.value, keyword: billKeyword.value } })
    if (res.data.code === 0 || res.data.code === 200) {
      billList.value = res.data.data?.list ?? []
      selectedBill.value = undefined
    }
  } finally {
    billLoading.value = false
  }
}

async function importBill() {
  if (!selectedBill.value) return
  importingBill.value = true
  try {
    const res = await api.get('/api/v1/product-barcodes/print-bill-items', { params: { type: billType.value, id: selectedBill.value.id } })
    if (res.data.code === 0 || res.data.code === 200) {
      const items = res.data.data?.list ?? []
      if (!items.length) {
        ElMessage.warning('该单据没有明细')
        return
      }
      for (const it of items) {
        pushItem({ id: it.productId, name: it.name, spec: it.spec, barcode: it.barcode, code: it.code, retailPrice: it.retailPrice }, Math.max(1, Math.round(it.quantity)))
      }
      ElMessage.success(`已引入 ${items.length} 条明细`)
      billDialogVisible.value = false
    }
  } finally {
    importingBill.value = false
  }
}

watch(billDialogVisible, (v) => {
  if (v) loadBills()
})

onMounted(() => {
  searchProducts('')
  window.addEventListener('keydown', onGlobalKeydown)
})

onBeforeUnmount(() => {
  window.removeEventListener('keydown', onGlobalKeydown)
})
</script>

<style scoped>
.page { display: flex; flex-direction: column; gap: 12px; }
.page-header { display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 8px; }
.page-title { font-size: 18px; font-weight: 600; margin: 0; }
.actions { display: flex; gap: 8px; flex-wrap: wrap; }
.empty { color: #909399; }

.print-area {
  display: flex; flex-wrap: wrap; gap: 8px; margin-top: 12px;
}
.label {
  border: 1px dashed #dcdfe6; border-radius: 4px;
  padding: 8px; text-align: center; page-break-inside: avoid;
}
.label-name { font-size: 13px; font-weight: 600; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.label-spec { font-size: 11px; color: #606266; }
.label-price { font-size: 13px; font-weight: 600; color: #f56c6c; }

@media print {
  .page-header, .el-card { display: none !important; }
  .print-area { margin: 0; }
  .label { border: none; }
}
</style>
