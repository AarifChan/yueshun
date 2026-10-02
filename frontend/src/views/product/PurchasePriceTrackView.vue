<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">采购价格跟踪</h2>
      <div class="actions">
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
        <template #empty>暂无数据</template>
      </el-table>
      <el-pagination
        v-model:current-page="page" v-model:page-size="pageSize"
        :total="total" :page-sizes="[30, 50, 100]"
        layout="total, sizes, prev, pager, next" style="margin-top: 12px; justify-content: flex-end"
        @current-change="load()" @size-change="load(1)" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useReport } from '@/composables/useReport'
import { fetchCategoryTree, type CategoryNode } from '@/api/category'
import { fetchBrandOptions } from '@/api/productDetail'
import { fetchTags, type ProductTag } from '@/api/tag'
import { fetchUnits, type GoodsUnit } from '@/api/unit'

interface Row {
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
</style>
