<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">条码打印</h2>
      <div class="actions">
        <el-button @click="clearAll">清空</el-button>
        <el-button type="primary" :disabled="!printItems.length" @click="doPrint">打印</el-button>
      </div>
    </div>

    <el-card>
      <el-form inline @submit.prevent>
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
        <template #empty>请搜索并添加要打印条码的商品</template>
      </el-table>
    </el-card>

    <!-- 打印预览区（打印时仅输出此区域） -->
    <div v-if="printItems.length" class="print-area" id="barcode-print-area">
      <div v-for="(label, i) in labels" :key="i" class="label">
        <div v-if="labelOpts.showName" class="label-name">{{ label.name }}</div>
        <div v-if="labelOpts.showSpec" class="label-spec">{{ label.spec }}</div>
        <svg :ref="(el) => setSvgRef(el, label.code)" class="label-barcode"></svg>
        <div v-if="labelOpts.showPrice" class="label-price">￥{{ label.retailPrice.toFixed(2) }}</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, reactive, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import JsBarcode from 'jsbarcode'
import api from '@/api/client'

interface Product { id: number; name: string; code: string; barcode: string; spec?: string; specification?: string; retailPrice?: number; retail_price?: number }
interface PrintItem { id: number; name: string; spec: string; barcode: string; retailPrice: number; qty: number }

const productOptions = ref<Product[]>([])
const pickId = ref<number>()
const printItems = ref<PrintItem[]>([])
const labelOpts = reactive({ showName: true, showSpec: true, showPrice: true })
const svgRefs = new Map<string, SVGElement[]>()

const labels = computed(() => {
  const arr: { name: string; spec: string; code: string; retailPrice: number }[] = []
  for (const item of printItems.value) {
    for (let i = 0; i < item.qty; i++) {
      arr.push({ name: item.name, spec: item.spec, code: item.barcode, retailPrice: item.retailPrice })
    }
  }
  return arr
})

watch(labels, async () => {
  await nextTick()
  renderBarcodes()
}, { deep: true })

function setSvgRef(el: Element | any, code: string) {
  if (!el) return
  const list = svgRefs.get(code) ?? []
  if (!list.includes(el as SVGElement)) {
    list.push(el as SVGElement)
    svgRefs.set(code, list)
  }
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
  svgRefs.forEach((els, code) => els.forEach((el) => renderOne(el, code)))
}

async function searchProducts(kw: string) {
  const res = await api.get('/api/v1/products', { params: { page: 1, pageSize: 50, keyword: kw } })
  if (res.data.code === 0 || res.data.code === 200) productOptions.value = res.data.data?.list ?? []
}

function addProduct(id: number | undefined) {
  if (!id) return
  const p = productOptions.value.find((x) => x.id === id)
  pickId.value = undefined
  if (!p) return
  const exist = printItems.value.find((x) => x.id === p.id)
  if (exist) {
    exist.qty += 1
    return
  }
  printItems.value.push({
    id: p.id,
    name: p.name,
    spec: p.spec ?? p.specification ?? '',
    barcode: p.barcode || p.code || String(p.id),
    retailPrice: p.retailPrice ?? p.retail_price ?? 0,
    qty: 1,
  })
}

function clearAll() {
  printItems.value = []
}

function doPrint() {
  if (!printItems.value.length) {
    ElMessage.warning('请先添加商品')
    return
  }
  renderBarcodes()
  window.print()
}

searchProducts('')
</script>

<style scoped>
.page { display: flex; flex-direction: column; gap: 12px; }
.page-header { display: flex; justify-content: space-between; align-items: center; }
.page-title { font-size: 18px; font-weight: 600; margin: 0; }
.actions { display: flex; gap: 8px; }
.empty { color: #909399; }

.print-area {
  display: flex; flex-wrap: wrap; gap: 8px; margin-top: 12px;
}
.label {
  width: 180px; border: 1px dashed #dcdfe6; border-radius: 4px;
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
