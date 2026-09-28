<template>
  <div class="page">
    <el-card>
      <template #header>
        <div class="card-header"><span>规格单位条码</span></div>
      </template>
      <el-form :model="searchForm" inline class="search-form">
        <el-form-item label="关键词"><el-input v-model="searchForm.keyword" placeholder="名称/编码/条码" clearable /></el-form-item>
        <el-form-item><el-button type="primary" @click="handleSearch">搜索</el-button><el-button @click="handleReset">重置</el-button></el-form-item>
      </el-form>
      <el-table :data="list" v-loading="loading" border stripe>
        <el-table-column prop="name" label="商品名称" min-width="160" />
        <el-table-column prop="code" label="编码" min-width="110" />
        <el-table-column prop="specification" label="规格" min-width="110">
          <template #default="{ row }">{{ row.specification || '-' }}</template>
        </el-table-column>
        <el-table-column prop="unit" label="主单位" width="90" />
        <el-table-column prop="barcode" label="主条码" min-width="130">
          <template #default="{ row }">{{ row.barcode || '-' }}</template>
        </el-table-column>
        <el-table-column label="操作" width="170" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="openMaintain(row)">维护</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-pagination v-model:current-page="pagination.page" v-model:page-size="pagination.pageSize" :total="total" :page-sizes="[10,20,50]" layout="total, sizes, prev, pager, next" @size-change="handleSizeChange" @current-change="handleCurrentChange" class="pagination" />
    </el-card>

    <el-dialog v-model="dialogVisible" :title="`维护规格单位条码 - ${currentProduct?.name || ''}`" width="760px" top="6vh">
      <el-form label-width="90px" class="basic-form">
        <el-row :gutter="12">
          <el-col :span="8">
            <el-form-item label="规格"><el-input v-model="basic.specification" placeholder="如：500ml/箱" /></el-form-item>
          </el-col>
          <el-col :span="6">
            <el-form-item label="主单位"><el-input v-model="basic.unit" /></el-form-item>
          </el-col>
          <el-col :span="10">
            <el-form-item label="主条码"><el-input v-model="basic.barcode" /></el-form-item>
          </el-col>
        </el-row>
        <el-form-item>
          <el-button size="small" type="primary" plain :loading="savingBasic" @click="saveBasic">保存基本信息</el-button>
        </el-form-item>
      </el-form>

      <el-divider content-position="left">辅助单位</el-divider>
      <el-table :data="unitRows" size="small" border>
        <el-table-column label="单位名称" min-width="140">
          <template #default="{ row }"><el-input v-model="row.name" size="small" placeholder="如：箱" /></template>
        </el-table-column>
        <el-table-column label="换算系数" width="140">
          <template #default="{ row }"><el-input-number v-model="row.conversion" size="small" :min="0.0001" :controls="false" style="width: 100%" /></template>
        </el-table-column>
        <el-table-column label="单位条码" min-width="160">
          <template #default="{ row }"><el-input v-model="row.barcode" size="small" /></template>
        </el-table-column>
        <el-table-column label="默认" width="70" align="center">
          <template #default="{ row }"><el-checkbox v-model="row.isDefault" /></template>
        </el-table-column>
        <el-table-column label="" width="70" align="center">
          <template #default="{ $index }"><el-button link type="danger" size="small" @click="unitRows.splice($index, 1)">删除</el-button></template>
        </el-table-column>
      </el-table>
      <div class="row-actions">
        <el-button size="small" @click="unitRows.push({ id: 0, name: '', conversion: 1, barcode: '', isDefault: false })">+ 添加单位</el-button>
        <el-button size="small" type="primary" plain :loading="savingUnits" @click="saveUnits">保存辅助单位</el-button>
      </div>

      <el-divider content-position="left">多单位条码</el-divider>
      <el-table :data="barcodeRows" size="small" border>
        <el-table-column label="所属单位" min-width="150">
          <template #default="{ row }">
            <el-select v-model="row.unitId" size="small" style="width: 100%">
              <el-option label="主单位" :value="0" />
              <el-option v-for="u in unitRows.filter((u) => u.id)" :key="u.id" :label="u.name" :value="u.id" />
            </el-select>
          </template>
        </el-table-column>
        <el-table-column label="条码" min-width="200">
          <template #default="{ row }"><el-input v-model="row.barcode" size="small" /></template>
        </el-table-column>
        <el-table-column label="" width="70" align="center">
          <template #default="{ $index }"><el-button link type="danger" size="small" @click="barcodeRows.splice($index, 1)">删除</el-button></template>
        </el-table-column>
      </el-table>
      <div class="row-actions">
        <el-button size="small" @click="barcodeRows.push({ unitId: 0, barcode: '' })">+ 添加条码</el-button>
        <el-button size="small" type="primary" plain :loading="savingBarcodes" @click="saveBarcodes">保存条码</el-button>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import api from '@/api/client'
import { ElMessage } from 'element-plus'
import { useCrud } from '@/composables/useCrud'

interface Product { id: number; name: string; code: string; specification: string; unit: string; barcode: string }

const crud = useCrud<Product>({ baseUrl: '/api/v1/products', listUrl: '/api/v1/products' })
const { list, total, loading, searchForm, pagination, fetchList, handleSearch, handleReset, handleSizeChange, handleCurrentChange } = crud
fetchList()

const dialogVisible = ref(false)
const currentProduct = ref<Product | null>(null)
const basic = reactive({ specification: '', unit: '', barcode: '' })
const unitRows = ref<{ id: number; name: string; conversion: number; barcode: string; isDefault: boolean }[]>([])
const barcodeRows = ref<{ unitId: number; barcode: string }[]>([])
const savingBasic = ref(false)
const savingUnits = ref(false)
const savingBarcodes = ref(false)

async function loadDetail() {
  if (!currentProduct.value) return
  const res = await api.get(`/api/v1/products/${currentProduct.value.id}`)
  const d = res.data.data
  basic.specification = d.product?.specification || ''
  basic.unit = d.product?.unit || ''
  basic.barcode = d.product?.barcode || ''
  unitRows.value = (d.units || []).map((u: any) => ({
    id: u.id,
    name: u.name,
    conversion: Number(u.conversion),
    barcode: '',
    isDefault: !!u.isDefault,
  }))
  barcodeRows.value = (d.barcodes || []).map((b: any) => ({ unitId: b.unitId, barcode: b.barcode }))
}

async function openMaintain(row: Product) {
  currentProduct.value = row
  dialogVisible.value = true
  await loadDetail()
}

async function saveBasic() {
  if (!currentProduct.value) return
  savingBasic.value = true
  try {
    const res = await api.put(`/api/v1/products/${currentProduct.value.id}`, { ...basic })
    if (res.data.code === 200) {
      ElMessage.success('保存成功')
      fetchList()
    }
  } finally {
    savingBasic.value = false
  }
}

async function saveUnits() {
  if (!currentProduct.value) return
  const units = unitRows.value.filter((u) => u.name)
  if (units.some((u) => !u.conversion || u.conversion <= 0)) {
    ElMessage.warning('换算系数必须大于 0')
    return
  }
  savingUnits.value = true
  try {
    const res = await api.put(`/api/v1/products/${currentProduct.value.id}/units`, { units })
    if (res.data.code === 200) {
      ElMessage.success('辅助单位已保存')
      await loadDetail()
    }
  } finally {
    savingUnits.value = false
  }
}

async function saveBarcodes() {
  if (!currentProduct.value) return
  const barcodes = barcodeRows.value.filter((b) => b.barcode.trim())
  savingBarcodes.value = true
  try {
    const res = await api.put(`/api/v1/products/${currentProduct.value.id}/barcodes`, { barcodes })
    if (res.data.code === 200) ElMessage.success('条码已保存')
  } finally {
    savingBarcodes.value = false
  }
}
</script>

<style scoped>
.page { padding: 20px; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
.search-form { margin-bottom: 16px; }
.pagination { margin-top: 16px; justify-content: flex-end; }
.basic-form { margin-bottom: -18px; }
.row-actions {
  display: flex;
  gap: 8px;
  margin: 8px 0 4px;
}
</style>
