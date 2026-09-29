<template>
  <div class="page" v-loading="loading">
    <el-tabs v-model="activeTab" class="center-tabs">
      <el-tab-pane label="基本信息" name="basic" />
      <el-tab-pane label="商品详情" name="detail" />
      <el-tab-pane label="开单控制" name="billing" />
    </el-tabs>

    <div v-show="activeTab === 'basic'" class="tab-body">
      <div class="quick-bar">
        <el-icon><Lightning /></el-icon>
        <span class="quick-title">快速录入</span>
        <span class="quick-desc">已连接条码库，可扫码快速录入</span>
      </div>

      <el-form ref="formRef" :model="form" :rules="formRules" label-width="110px">
        <div class="basic-columns">
          <div class="basic-left">
            <el-form-item label="商品名称" prop="name">
              <el-input v-model="form.name" maxlength="100" show-word-limit placeholder="请输入商品名称" />
            </el-form-item>
            <el-form-item label="商城展示名称" prop="mallName">
              <el-input v-model="form.mallName" maxlength="100" show-word-limit placeholder="请输入商城展示名称" />
            </el-form-item>
            <el-form-item label="商品分类" prop="categoryId">
              <el-tree-select
                v-model="form.categoryId"
                :data="categoryTree"
                :props="{ label: 'name', children: 'children' }"
                node-key="id"
                check-strictly
                placeholder="请选择商品分类"
                style="width: 100%"
              />
            </el-form-item>
            <el-form-item label="专题分类">
              <el-select v-model="form.topicCategory" filterable allow-create default-first-option placeholder="请选择或输入专题分类" style="width: 100%">
                <el-option v-for="t in topicCategories" :key="t" :label="t" :value="t" />
              </el-select>
            </el-form-item>
            <el-form-item label="拼音码">
              <el-input v-model="form.pinyinCode" maxlength="100" show-word-limit placeholder="请输入拼音码" />
            </el-form-item>
            <el-form-item label="商品品牌">
              <el-select v-model="form.brandId" filterable clearable placeholder="请选择商品品牌" style="width: 100%">
                <el-option v-for="b in brandOptions" :key="b.id" :label="b.name" :value="b.id" />
              </el-select>
            </el-form-item>

            <div class="collapse-toggle" @click="moreOpen = !moreOpen">
              展开
              <el-icon v-if="moreOpen"><ArrowUp /></el-icon>
              <el-icon v-else><ArrowDown /></el-icon>
            </div>
            <template v-if="moreOpen">
              <el-form-item label="商品编号">
                <el-input v-model="form.code" placeholder="留空自动生成" />
              </el-form-item>
              <el-form-item label="条码">
                <el-input v-model="form.barcode" placeholder="请输入条码" />
              </el-form-item>
              <el-form-item label="采购价">
                <el-input-number v-model="form.purchasePrice" :precision="2" :min="0" :controls="false" style="width: 100%" />
              </el-form-item>
              <el-form-item label="批发价">
                <el-input-number v-model="form.wholesalePrice" :precision="2" :min="0" :controls="false" style="width: 100%" />
              </el-form-item>
              <el-form-item label="零售价">
                <el-input-number v-model="form.retailPrice" :precision="2" :min="0" :controls="false" style="width: 100%" />
              </el-form-item>
              <el-form-item label="最低库存">
                <el-input-number v-model="form.minStock" :min="0" :controls="false" style="width: 100%" />
              </el-form-item>
              <el-form-item label="最高库存">
                <el-input-number v-model="form.maxStock" :min="0" :controls="false" style="width: 100%" />
              </el-form-item>
              <el-form-item label="最低售价">
                <el-input-number v-model="form.minSalePrice" :precision="2" :min="0" :controls="false" style="width: 100%" />
              </el-form-item>
              <el-form-item label="商城排序权重">
                <el-input-number v-model="form.mallSortWeight" :precision="0" :min="0" :controls="false" style="width: 100%" />
              </el-form-item>
              <el-form-item label="搜索关键词">
                <el-input v-model="form.searchKeywords" maxlength="200" show-word-limit placeholder="请输入搜索关键词" />
              </el-form-item>
              <el-form-item label="默认供应商">
                <RemoteSelect v-model="form.supplierId" api-url="/api/v1/suppliers" placeholder="请选择默认供应商" />
              </el-form-item>
              <el-form-item label="出库仓库">
                <RemoteSelect v-model="form.warehouseId" api-url="/api/v1/warehouses" placeholder="请选择出库仓库" />
              </el-form-item>
              <el-form-item label="禁购单位">
                <el-select v-model="form.forbidPurchaseUnits" multiple placeholder="请选择禁购单位" style="width: 100%">
                  <el-option v-for="name in forbidUnitOptions" :key="name" :label="name" :value="name" />
                </el-select>
              </el-form-item>
              <el-form-item label="商品描述">
                <el-input v-model="form.description" type="textarea" :rows="3" maxlength="500" show-word-limit />
              </el-form-item>
            </template>
          </div>

          <div class="basic-right">
            <el-upload
              class="media-box"
              :show-file-list="false"
              accept="image/*"
              :http-request="uploadMainImage"
            >
              <div class="media-box-inner">
                <img v-if="form.image" :src="form.image" class="media-preview" alt="" />
                <template v-else>
                  <el-icon class="media-plus"><Plus /></el-icon>
                  <div class="media-text">上传商品图</div>
                </template>
              </div>
            </el-upload>
            <div class="media-box media-disabled" title="暂不支持">
              <div class="media-box-inner">
                <el-icon class="media-plus"><Plus /></el-icon>
                <div class="media-text">上传视频</div>
              </div>
            </div>
          </div>
        </div>
      </el-form>

      <div class="section">
        <div class="section-header">
          <span class="section-title">商品单位</span>
          <div class="section-actions">
            <el-checkbox v-model="form.manualLevelPriceDefault">手动设置的级别价取默认订货价</el-checkbox>
            <el-checkbox v-model="form.auxPriceReverseCalc">辅助单位价格变更后反算基本单位价格</el-checkbox>
          </div>
        </div>
        <el-table :data="unitRows" border size="small">
          <el-table-column label="操作" width="60" align="center">
            <template #default="{ $index }">
              <el-button v-if="$index > 0" link type="danger" :icon="Delete" @click="removeUnit($index)" />
            </template>
          </el-table-column>
          <el-table-column label="单位类型" width="100" align="center">
            <template #default="{ $index }">{{ $index === 0 ? '基本单位' : '辅助单位' }}</template>
          </el-table-column>
          <el-table-column label="单位名称" min-width="130">
            <template #default="{ row }">
              <el-select v-model="row.name" filterable allow-create default-first-option placeholder="请选择" @change="(v: string) => onUnitNameChange(row, v)">
                <el-option v-for="u in goodsUnitOptions" :key="u.id" :label="u.name" :value="u.name" />
              </el-select>
            </template>
          </el-table-column>
          <el-table-column label="*换算率" width="120">
            <template #default="{ row, $index }">
              <el-input-number v-model="row.conversion" :min="0.0001" :controls="false" :disabled="$index === 0" style="width: 100%" />
            </template>
          </el-table-column>
          <el-table-column label="单位条码" min-width="120">
            <template #default="{ row }">
              <el-input v-model="row.barcode" placeholder="条码" />
            </template>
          </el-table-column>
          <el-table-column label="允许销售" width="90" align="center">
            <template #default="{ row }">
              <el-checkbox v-model="row.allowSale">允许</el-checkbox>
            </template>
          </el-table-column>
          <el-table-column label="默认销售单位" width="110" align="center">
            <template #default="{ row, $index }">
              <el-radio :model-value="row.defaultSale" :value="true" @change="setDefaultSale($index)">是</el-radio>
            </template>
          </el-table-column>
          <el-table-column label="默认采购单位" width="110" align="center">
            <template #default="{ row, $index }">
              <el-radio :model-value="row.defaultPurchase" :value="true" @change="setDefaultPurchase($index)">是</el-radio>
            </template>
          </el-table-column>
          <el-table-column label="存放类型" width="110">
            <template #default="{ row }">
              <el-select v-model="row.storageType">
                <el-option label="散货" :value="1" />
                <el-option label="整件" :value="2" />
              </el-select>
            </template>
          </el-table-column>
          <el-table-column label="参考价" width="130">
            <template #default="{ row }">
              <el-input-number v-model="row.refPrice" :precision="2" :min="0" :controls="false" style="width: 100%" />
            </template>
          </el-table-column>
          <el-table-column label="批发价" width="130">
            <template #default="{ row }">
              <el-input-number v-model="row.wholesalePrice" :precision="2" :min="0" :controls="false" style="width: 100%" />
            </template>
          </el-table-column>
          <el-table-column label="零售价" width="130">
            <template #default="{ row }">
              <el-input-number v-model="row.retailPrice" :precision="2" :min="0" :controls="false" style="width: 100%" />
            </template>
          </el-table-column>
        </el-table>
        <el-button link type="primary" :icon="Plus" @click="addUnit">新增单位</el-button>
      </div>

      <div class="section">
        <div class="section-header">
          <span class="section-title">商品规格</span>
          <div class="section-actions">
            <span class="switch-label">多规格</span>
            <el-switch v-model="multiSpec" @change="onMultiSpecChange" />
          </div>
        </div>
        <el-table :data="specRows" border size="small">
          <el-table-column label="规格" min-width="120">
            <template #default="{ row }">
              <el-input v-model="row.specValue" placeholder="如：红色/XL" />
            </template>
          </el-table-column>
          <el-table-column label="上架" width="70" align="center">
            <template #default="{ row }">
              <el-switch v-model="row.onShelf" />
            </template>
          </el-table-column>
          <el-table-column label="商品编号" min-width="110">
            <template #default="{ row }">
              <el-input v-model="row.code" />
            </template>
          </el-table-column>
          <el-table-column label="规格条码" min-width="110">
            <template #default="{ row }">
              <el-input v-model="row.barcode" />
            </template>
          </el-table-column>
          <el-table-column label="商品重量(kg)" width="120">
            <template #default="{ row }">
              <el-input-number v-model="row.weight" :min="0" :precision="3" :controls="false" style="width: 100%" />
            </template>
          </el-table-column>
          <el-table-column label="商品体积(m³)" width="120">
            <template #default="{ row }">
              <el-input-number v-model="row.volume" :min="0" :precision="4" :controls="false" style="width: 100%" />
            </template>
          </el-table-column>
          <el-table-column label="备注1" min-width="100">
            <template #default="{ row }">
              <el-input v-model="row.remark1" />
            </template>
          </el-table-column>
          <el-table-column label="备注2" min-width="100">
            <template #default="{ row }">
              <el-input v-model="row.remark2" />
            </template>
          </el-table-column>
          <el-table-column label="操作" width="60" align="center">
            <template #default="{ $index }">
              <el-button v-if="multiSpec && specRows.length > 1" link type="danger" :icon="Delete" @click="removeSpec($index)" />
            </template>
          </el-table-column>
        </el-table>
        <el-button v-if="multiSpec" link type="primary" :icon="Plus" @click="addSpec">添加规格</el-button>
      </div>

      <div class="section">
        <div class="section-header">
          <span class="section-title">期初库存</span>
        </div>
        <div v-if="openingStocks.length === 0" class="empty-text">暂无仓库</div>
        <div v-for="s in openingStocks" :key="s.warehouseId" class="stock-row">
          <span class="stock-name">{{ s.name }}</span>
          <el-input-number v-model="s.quantity" :min="0" :controls="false" style="width: 200px" />
        </div>
      </div>
    </div>

    <div v-show="activeTab === 'detail'" class="tab-body detail-tab">
      <div class="detail-add-bar">
        <el-upload :show-file-list="false" accept="image/*" :http-request="uploadDetailImage">
          <div class="add-box">
            <el-icon><Plus /></el-icon>
            <span>添加图片</span>
          </div>
        </el-upload>
        <div class="add-box add-disabled" title="暂不支持">
          <el-icon><Plus /></el-icon>
          <span>添加视频</span>
        </div>
        <div class="add-box" @click="addTextBlock">
          <el-icon><Plus /></el-icon>
          <span>添加文字</span>
        </div>
      </div>
      <div v-if="detailBlocks.length === 0" class="empty-text">暂无详情内容，点击上方按钮添加</div>
      <div v-for="(block, i) in detailBlocks" :key="i" class="detail-block">
        <img v-if="block.type === 'image'" :src="block.content" class="detail-image" alt="" />
        <el-input v-else-if="block.type === 'text'" v-model="block.content" type="textarea" :rows="3" placeholder="请输入文字" />
        <div v-else class="add-box add-disabled" title="暂不支持">
          <el-icon><Plus /></el-icon>
          <span>视频（暂不支持）</span>
        </div>
        <div class="block-actions">
          <el-button link :icon="ArrowUp" :disabled="i === 0" @click="moveBlock(i, -1)" />
          <el-button link :icon="ArrowDown" :disabled="i === detailBlocks.length - 1" @click="moveBlock(i, 1)" />
          <el-button link type="danger" :icon="Delete" @click="detailBlocks.splice(i, 1)" />
        </div>
      </div>
    </div>

    <div v-show="activeTab === 'billing'" class="tab-body billing-tab">
      <el-alert type="info" :closable="false" title="商品销售限制只影响代开单及小程序商城下单" class="billing-alert" />
      <div class="section">
        <div class="section-header">
          <span class="section-title">商品起订量/限订量</span>
        </div>
        <div class="billing-row">
          <span class="billing-label">起订量</span>
          <el-input-number v-model="form.minOrderQty" :min="0" :controls="false" style="width: 160px" />
          <span class="billing-unit">{{ baseUnitName }}</span>
          <span class="billing-label billing-gap">限订量</span>
          <el-input-number v-model="form.maxOrderQty" :min="0" :controls="false" style="width: 160px" />
          <span class="billing-unit">{{ baseUnitName }}</span>
        </div>
        <el-checkbox v-model="form.orderByMultiple">按起订量倍数订购</el-checkbox>
      </div>
      <div class="section">
        <div class="section-header">
          <span class="section-title">商品库存控制</span>
        </div>
        <el-checkbox v-model="form.noAvailableStockControl">无需管控可用库存</el-checkbox>
        <el-checkbox v-model="form.noBookStockControl">无需管控账面库存</el-checkbox>
      </div>
    </div>

    <div class="footer-bar">
      <el-button @click="save('new')" :loading="saving">保存并新增</el-button>
      <el-button @click="save('copy')" :loading="saving">保存并复制</el-button>
      <el-button type="primary" @click="save('back')" :loading="saving">保 存</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import type { UploadRequestOptions } from 'element-plus'
import { ArrowDown, ArrowUp, Delete, Lightning, Plus } from '@element-plus/icons-vue'
import { fetchCategoryTree } from '@/api/category'
import type { CategoryNode } from '@/api/category'
import RemoteSelect from '@/components/RemoteSelect.vue'
import {
  createProduct,
  fetchBrandOptions,
  fetchGoodsUnitOptions,
  fetchProduct,
  fetchTopicCategories,
  fetchWarehouseOptions,
  replaceProductSpecs,
  replaceProductUnits,
  saveOpeningStock,
  updateProduct,
  uploadFile,
} from '@/api/productDetail'
import type {
  GoodsUnitOption,
  OpeningStockItem,
  ProductPayload,
  ProductSpecItem,
  ProductUnitItem,
} from '@/api/productDetail'

interface DetailBlock {
  type: 'image' | 'video' | 'text'
  content: string
}

interface OpeningStockRow extends OpeningStockItem {
  name: string
}

const route = useRoute()
const router = useRouter()
const routeId = route.params.id as string
const productId = ref<number | undefined>(routeId === 'new' ? undefined : Number(routeId))
const isCreate = computed(() => productId.value === undefined)

const activeTab = ref('basic')
const loading = ref(false)
const saving = ref(false)
const moreOpen = ref(false)
const multiSpec = ref(false)

const formRef = ref<FormInstance>()
const form = reactive({
  name: '',
  mallName: '',
  categoryId: undefined as number | undefined,
  topicCategory: '',
  pinyinCode: '',
  brandId: undefined as number | undefined,
  code: '',
  barcode: '',
  specification: '',
  purchasePrice: 0,
  wholesalePrice: 0,
  retailPrice: 0,
  minStock: 0,
  maxStock: 0,
  description: '',
  image: '',
  videoUrl: '',
  minOrderQty: 0,
  maxOrderQty: 0,
  orderByMultiple: false,
  noAvailableStockControl: false,
  noBookStockControl: false,
  manualLevelPriceDefault: false,
  auxPriceReverseCalc: false,
  minSalePrice: 0,
  mallSortWeight: 0,
  searchKeywords: '',
  warehouseId: undefined as number | undefined,
  supplierId: undefined as number | undefined,
  forbidPurchaseUnits: [] as string[],
  status: 1,
})

const formRules: FormRules = {
  name: [{ required: true, message: '请输入商品名称', trigger: 'blur' }],
  mallName: [{ required: true, message: '请输入商城展示名称', trigger: 'blur' }],
  categoryId: [{ required: true, message: '请选择商品分类', trigger: 'change' }],
}

const categoryTree = ref<CategoryNode[]>([])
const topicCategories = ref<string[]>([])
const brandOptions = ref<{ id: number; name: string }[]>([])
const goodsUnitOptions = ref<GoodsUnitOption[]>([])
const unitRows = ref<ProductUnitItem[]>([])
const specRows = ref<ProductSpecItem[]>([])
const openingStocks = ref<OpeningStockRow[]>([])
const detailBlocks = ref<DetailBlock[]>([])

const baseUnitName = computed(() => unitRows.value[0]?.name || '单位')

const forbidUnitOptions = computed(() =>
  [...new Set(unitRows.value.map((u) => u.name).filter((n) => !!n))]
)

function ok(res: any) {
  return res.data.code === 0 || res.data.code === 200
}

function newBaseUnit(): ProductUnitItem {
  return {
    name: '套', conversion: 1, isDefault: true, barcode: '', allowSale: true,
    defaultSale: true, defaultPurchase: true, storageType: 1, refPrice: 0,
    wholesalePrice: 0, retailPrice: 0,
  }
}

function newAuxUnit(): ProductUnitItem {
  return {
    name: '', conversion: 1, isDefault: false, barcode: '', allowSale: true,
    defaultSale: false, defaultPurchase: false, storageType: 1, refPrice: 0,
    wholesalePrice: 0, retailPrice: 0,
  }
}

function newSpec(): ProductSpecItem {
  return { specValue: '', code: '', barcode: '', weight: 0, volume: 0, remark1: '', remark2: '', onShelf: true, sort: 0 }
}

function onUnitNameChange(row: ProductUnitItem, name: string) {
  const unit = goodsUnitOptions.value.find((u) => u.name === name)
  if (unit) row.storageType = unit.storageType
}

function setDefaultSale(index: number) {
  unitRows.value.forEach((r, i) => { r.defaultSale = i === index })
}

function setDefaultPurchase(index: number) {
  unitRows.value.forEach((r, i) => { r.defaultPurchase = i === index })
}

function addUnit() {
  unitRows.value.push(newAuxUnit())
}

function removeUnit(index: number) {
  const removed = unitRows.value[index]
  unitRows.value.splice(index, 1)
  if (removed.defaultSale) unitRows.value[0].defaultSale = true
  if (removed.defaultPurchase) unitRows.value[0].defaultPurchase = true
}

function onMultiSpecChange(val: string | number | boolean) {
  if (!val) specRows.value = [newSpec()]
  else if (specRows.value.length === 0) specRows.value = [newSpec()]
}

function addSpec() {
  specRows.value.push(newSpec())
}

function removeSpec(index: number) {
  specRows.value.splice(index, 1)
}

function addTextBlock() {
  detailBlocks.value.push({ type: 'text', content: '' })
}

function moveBlock(index: number, dir: number) {
  const target = index + dir
  if (target < 0 || target >= detailBlocks.value.length) return
  const [item] = detailBlocks.value.splice(index, 1)
  detailBlocks.value.splice(target, 0, item)
}

async function handleUpload(options: UploadRequestOptions, apply: (url: string) => void) {
  try {
    const res = await uploadFile(options.file as File)
    if (ok(res)) {
      apply(res.data.data.url)
      options.onSuccess(res.data.data)
    } else {
      options.onError(new Error(res.data.message || '上传失败') as any)
    }
  } catch (e: any) {
    options.onError(e as any)
  }
}

function uploadMainImage(options: UploadRequestOptions) {
  handleUpload(options, (url) => { form.image = url })
}

function uploadDetailImage(options: UploadRequestOptions) {
  handleUpload(options, (url) => { detailBlocks.value.push({ type: 'image', content: url }) })
}

async function loadOptions() {
  try {
    const [cats, topics, brands, units, warehouses] = await Promise.all([
      fetchCategoryTree(),
      fetchTopicCategories(),
      fetchBrandOptions(),
      fetchGoodsUnitOptions(),
      fetchWarehouseOptions(),
    ])
    if (ok(cats)) categoryTree.value = cats.data.data || []
    if (ok(topics)) topicCategories.value = topics.data.data || []
    if (ok(brands)) brandOptions.value = brands.data.data?.list || []
    if (ok(units)) goodsUnitOptions.value = units.data.data?.list || []
    if (ok(warehouses)) {
      openingStocks.value = (warehouses.data.data?.list || []).map((w: any) => ({
        warehouseId: w.id, name: w.name, quantity: 0,
      }))
    }
  } catch {
    // 拦截器已提示
  }
}

async function loadDetail(id: number) {
  loading.value = true
  try {
    const res = await fetchProduct(id)
    if (!ok(res)) return
    const d = res.data.data || {}
    const p = d.product ?? d
    Object.assign(form, {
      name: p.name || '',
      mallName: p.mallName || '',
      categoryId: p.categoryId || undefined,
      topicCategory: p.topicCategory || '',
      pinyinCode: p.pinyinCode || '',
      brandId: p.brandId || undefined,
      code: p.code || '',
      barcode: p.barcode || '',
      specification: p.specification || '',
      purchasePrice: p.purchasePrice || 0,
      wholesalePrice: p.wholesalePrice || 0,
      retailPrice: p.retailPrice || 0,
      minStock: p.minStock || 0,
      maxStock: p.maxStock || 0,
      description: p.description || '',
      image: p.image || '',
      videoUrl: p.videoUrl || '',
      minOrderQty: p.minOrderQty || 0,
      maxOrderQty: p.maxOrderQty || 0,
      orderByMultiple: !!p.orderByMultiple,
      noAvailableStockControl: !!p.noAvailableStockControl,
      noBookStockControl: !!p.noBookStockControl,
      manualLevelPriceDefault: !!p.manualLevelPriceDefault,
      auxPriceReverseCalc: !!p.auxPriceReverseCalc,
      minSalePrice: p.minSalePrice || 0,
      mallSortWeight: p.mallSortWeight || 0,
      searchKeywords: p.searchKeywords || '',
      warehouseId: p.warehouseId || undefined,
      supplierId: p.supplierId || undefined,
      forbidPurchaseUnits: p.forbidPurchaseUnits
        ? String(p.forbidPurchaseUnits).split(',').map((s: string) => s.trim()).filter(Boolean)
        : [],
      status: p.status ?? 1,
    })
    const units: any[] = d.units || []
    const base = units.find((u) => u.isDefault)
    const baseRow: ProductUnitItem = base
      ? { ...newBaseUnit(), ...base, isDefault: true, conversion: 1 }
      : { ...newBaseUnit(), name: p.unit || '套', barcode: p.barcode || '' }
    unitRows.value = [baseRow, ...units.filter((u) => !u.isDefault).map((u) => ({ ...newAuxUnit(), ...u }))]
    if (!unitRows.value.some((u) => u.defaultSale)) unitRows.value[0].defaultSale = true
    if (!unitRows.value.some((u) => u.defaultPurchase)) unitRows.value[0].defaultPurchase = true
    const specs: any[] = d.specItems || []
    specRows.value = specs.map((s, i) => ({ ...newSpec(), ...s, onShelf: !!s.onShelf, sort: i }))
    multiSpec.value = specRows.value.length > 1 || (specRows.value.length === 1 && !!specRows.value[0]?.specValue)
    if (specRows.value.length === 0) specRows.value = [newSpec()]
    try {
      const blocks = JSON.parse(p.detailContent || '[]')
      detailBlocks.value = Array.isArray(blocks) ? blocks : []
    } catch {
      detailBlocks.value = []
    }
  } catch {
    // 拦截器已提示
  } finally {
    loading.value = false
  }
}

function validateForm(): Promise<boolean> {
  return new Promise((resolve) => {
    formRef.value?.validate((valid) => {
      if (!valid) {
        activeTab.value = 'basic'
        resolve(false)
        return
      }
      for (const u of unitRows.value) {
        if (!u.name) {
          ElMessage.error('请填写单位名称')
          activeTab.value = 'basic'
          resolve(false)
          return
        }
        if (!u.conversion || u.conversion <= 0) {
          ElMessage.error('换算率必须大于 0')
          activeTab.value = 'basic'
          resolve(false)
          return
        }
      }
      if (unitRows.value.filter((u) => u.defaultSale).length !== 1) {
        ElMessage.error('默认销售单位必须唯一')
        activeTab.value = 'basic'
        resolve(false)
        return
      }
      if (unitRows.value.filter((u) => u.defaultPurchase).length !== 1) {
        ElMessage.error('默认采购单位必须唯一')
        activeTab.value = 'basic'
        resolve(false)
        return
      }
      resolve(true)
    })
  })
}

function buildPayload(): ProductPayload {
  return {
    name: form.name,
    mallName: form.mallName,
    categoryId: form.categoryId as number,
    brandId: form.brandId,
    topicCategory: form.topicCategory,
    pinyinCode: form.pinyinCode,
    code: form.code,
    barcode: form.barcode,
    unit: unitRows.value[0]?.name || '',
    specification: form.specification,
    purchasePrice: form.purchasePrice,
    wholesalePrice: form.wholesalePrice,
    retailPrice: form.retailPrice,
    minStock: form.minStock,
    maxStock: form.maxStock,
    description: form.description,
    image: form.image,
    videoUrl: form.videoUrl,
    detailContent: JSON.stringify(detailBlocks.value),
    minOrderQty: form.minOrderQty,
    maxOrderQty: form.maxOrderQty,
    orderByMultiple: form.orderByMultiple,
    noAvailableStockControl: form.noAvailableStockControl,
    noBookStockControl: form.noBookStockControl,
    manualLevelPriceDefault: form.manualLevelPriceDefault,
    auxPriceReverseCalc: form.auxPriceReverseCalc,
    minSalePrice: form.minSalePrice,
    mallSortWeight: form.mallSortWeight,
    searchKeywords: form.searchKeywords,
    warehouseId: form.warehouseId,
    supplierId: form.supplierId,
    forbidPurchaseUnits: form.forbidPurchaseUnits.join(','),
    status: form.status,
  }
}

function resetForm() {
  Object.assign(form, {
    name: '', mallName: '', categoryId: undefined, topicCategory: '', pinyinCode: '',
    brandId: undefined, code: '', barcode: '', specification: '',
    purchasePrice: 0, wholesalePrice: 0, retailPrice: 0, minStock: 0, maxStock: 0,
    description: '', image: '', videoUrl: '',
    minOrderQty: 0, maxOrderQty: 0, orderByMultiple: false,
    noAvailableStockControl: false, noBookStockControl: false,
    manualLevelPriceDefault: false, auxPriceReverseCalc: false,
    minSalePrice: 0, mallSortWeight: 0, searchKeywords: '',
    warehouseId: undefined, supplierId: undefined, forbidPurchaseUnits: [],
    status: 1,
  })
  unitRows.value = [newBaseUnit()]
  specRows.value = [newSpec()]
  multiSpec.value = false
  detailBlocks.value = []
  openingStocks.value.forEach((s) => { s.quantity = 0 })
  formRef.value?.clearValidate()
  activeTab.value = 'basic'
}

async function save(action: 'back' | 'new' | 'copy') {
  if (saving.value) return
  if (!(await validateForm())) return
  saving.value = true
  try {
    const wasCreate = isCreate.value
    const res = wasCreate
      ? await createProduct(buildPayload())
      : await updateProduct(productId.value as number, buildPayload())
    if (!ok(res)) return
    const id: number = wasCreate ? res.data.data?.id : (productId.value as number)
    if (wasCreate && res.data.data?.code) form.code = res.data.data.code

    const hasAux = unitRows.value.length > 1
    if (!wasCreate || hasAux) {
      const units = unitRows.value.map((u) => ({ ...u, conversion: u.isDefault ? 1 : u.conversion }))
      await replaceProductUnits(id, units)
    }
    const specItems = specRows.value
      .filter((s) => multiSpec.value || s.specValue || s.code || s.barcode)
      .map((s, i) => ({ ...s, sort: i }))
    if (!wasCreate || specItems.length > 0) {
      await replaceProductSpecs(id, specItems)
    }
    const stockItems = openingStocks.value
      .filter((s) => s.quantity > 0)
      .map((s) => ({ warehouseId: s.warehouseId, quantity: s.quantity }))
    if (stockItems.length > 0) {
      await saveOpeningStock(id, stockItems)
    }

    ElMessage.success('保存成功')
    if (action === 'back') {
      router.push('/products')
    } else if (action === 'new') {
      productId.value = undefined
      resetForm()
      if (routeId !== 'new') router.replace('/products/new')
    } else {
      productId.value = undefined
      form.code = ''
      unitRows.value.forEach((u) => { delete u.id })
      if (routeId !== 'new') router.replace('/products/new')
    }
  } catch {
    // 拦截器已提示
  } finally {
    saving.value = false
  }
}

unitRows.value = [newBaseUnit()]
specRows.value = [newSpec()]
loadOptions()
if (productId.value !== undefined) {
  loadDetail(productId.value)
}
</script>

<style scoped>
.page {
  padding: 0 20px 80px;
}
.center-tabs :deep(.el-tabs__nav-wrap::after) {
  display: none;
}
.center-tabs :deep(.el-tabs__nav) {
  float: none;
  display: flex;
  justify-content: center;
}
.tab-body {
  max-width: 1200px;
  margin: 0 auto;
}
.quick-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  background: #f5f7fa;
  border-radius: 4px;
  padding: 10px 16px;
  margin-bottom: 16px;
  color: #606266;
}
.quick-title {
  font-weight: 600;
  color: #303133;
}
.quick-desc {
  font-size: 13px;
}
.basic-columns {
  display: flex;
  gap: 32px;
}
.basic-left {
  flex: 0 0 60%;
}
.basic-right {
  flex: 0 0 30%;
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.collapse-toggle {
  color: var(--el-color-primary);
  cursor: pointer;
  margin: 0 0 16px 110px;
  display: flex;
  align-items: center;
  gap: 4px;
  user-select: none;
}
.media-box {
  border: 1px dashed #dcdfe6;
  border-radius: 6px;
  width: 100%;
  height: 180px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  overflow: hidden;
}
.media-box:hover {
  border-color: var(--el-color-primary);
}
.media-disabled {
  cursor: not-allowed;
  background: #fafafa;
  color: #c0c4cc;
}
.media-disabled:hover {
  border-color: #dcdfe6;
}
.media-box-inner {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  color: #909399;
  width: 100%;
  height: 100%;
  justify-content: center;
}
.media-plus {
  font-size: 28px;
}
.media-text {
  font-size: 14px;
}
.media-preview {
  width: 100%;
  height: 100%;
  object-fit: contain;
}
.section {
  margin-top: 24px;
}
.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}
.section-title {
  font-weight: 600;
  font-size: 15px;
  color: #303133;
}
.section-actions {
  display: flex;
  align-items: center;
  gap: 16px;
}
.switch-label {
  font-size: 14px;
  color: #606266;
}
.stock-row {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 8px 0;
}
.stock-name {
  width: 200px;
  color: #606266;
}
.empty-text {
  color: #909399;
  font-size: 13px;
  padding: 12px 0;
}
.detail-tab {
  max-width: 800px;
}
.detail-add-bar {
  display: flex;
  gap: 24px;
  justify-content: center;
  margin-bottom: 24px;
}
.add-box {
  border: 1px dashed #dcdfe6;
  border-radius: 6px;
  width: 160px;
  height: 100px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 6px;
  cursor: pointer;
  color: #909399;
}
.add-box:hover {
  border-color: var(--el-color-primary);
  color: var(--el-color-primary);
}
.add-disabled {
  cursor: not-allowed;
  background: #fafafa;
}
.add-disabled:hover {
  border-color: #dcdfe6;
  color: #909399;
}
.detail-block {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  margin-bottom: 16px;
}
.detail-image {
  max-width: 100%;
  border-radius: 4px;
}
.block-actions {
  display: flex;
  flex-direction: column;
}
.billing-tab {
  max-width: 800px;
}
.billing-alert {
  margin-bottom: 8px;
}
.billing-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
}
.billing-label {
  color: #606266;
}
.billing-gap {
  margin-left: 32px;
}
.billing-unit {
  color: #909399;
}
.footer-bar {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  background: #fff;
  border-top: 1px solid #e4e7ed;
  padding: 12px 0;
  display: flex;
  justify-content: center;
  gap: 16px;
  z-index: 10;
}
</style>
