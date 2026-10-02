<template>
  <div>
    <div class="toolbar">
      <el-select
        v-model="productId"
        filterable
        remote
        :remote-method="searchProducts"
        placeholder="搜索并选择要设置关联的商品"
        style="width: 320px"
        @change="loadRelations"
      >
        <el-option v-for="p in productOptions" :key="p.id" :label="`${p.name}（${p.code || '-'}）`" :value="p.id" />
      </el-select>
      <el-button type="primary" :disabled="!productId" @click="openAdd">设置关联</el-button>
    </div>

    <el-table :data="relations" v-loading="loading" border stripe>
      <el-table-column type="index" label="序" width="60" />
      <el-table-column label="图片" width="80" align="center">
        <template #default="{ row }">
          <el-image v-if="row.image" :src="row.image" style="width: 40px; height: 40px" fit="cover" />
          <span v-else>-</span>
        </template>
      </el-table-column>
      <el-table-column prop="name" label="商品名称" min-width="180" show-overflow-tooltip />
      <el-table-column prop="category" label="分类" width="140" />
      <el-table-column prop="brand" label="品牌" width="120" />
      <el-table-column label="关联状态" width="100" align="center">
        <template #default><el-tag type="success" size="small">已关联</el-tag></template>
      </el-table-column>
      <el-table-column label="操作" width="100" fixed="right">
        <template #default="{ row }">
          <el-button link type="danger" @click="removeRelation(row)">移除</el-button>
        </template>
      </el-table-column>
      <template #empty>{{ productId ? '暂无关联商品' : '请先选择商品' }}</template>
    </el-table>

    <el-dialog v-model="addVisible" title="设置关联商品" width="560px">
      <el-select
        v-model="selectedIds"
        multiple
        filterable
        remote
        :remote-method="searchAddProducts"
        placeholder="搜索商品并多选"
        style="width: 100%"
      >
        <el-option v-for="p in addOptions" :key="p.id" :label="`${p.name}（${p.code || '-'}）`" :value="p.id" />
      </el-select>
      <template #footer>
        <el-button @click="addVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="save">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import api from '@/api/client'

interface Option { id: number; name: string; code?: string }
interface RelationRow { id: number; relatedProductId: number; name: string; image: string; category: string; brand: string }

const productId = ref<number | null>(null)
const productOptions = ref<Option[]>([])
const addOptions = ref<Option[]>([])
const relations = ref<RelationRow[]>([])
const loading = ref(false)
const addVisible = ref(false)
const saving = ref(false)
const selectedIds = ref<number[]>([])

async function searchProducts(kw: string) {
  const res = await api.get('/api/v1/products', { params: { page: 1, pageSize: 20, keyword: kw } })
  if (res.data.code === 0 || res.data.code === 200) productOptions.value = res.data.data?.list ?? []
}

async function searchAddProducts(kw: string) {
  const res = await api.get('/api/v1/products', { params: { page: 1, pageSize: 20, keyword: kw } })
  if (res.data.code === 0 || res.data.code === 200) addOptions.value = res.data.data?.list ?? []
}

async function loadRelations() {
  if (!productId.value) return
  loading.value = true
  try {
    const res = await api.get('/api/v1/product-relations', { params: { productId: productId.value } })
    if (res.data.code === 0 || res.data.code === 200) relations.value = res.data.data?.list ?? []
  } finally {
    loading.value = false
  }
}

function openAdd() {
  selectedIds.value = relations.value.map(r => r.relatedProductId)
  addOptions.value = []
  addVisible.value = true
}

async function save() {
  if (!productId.value) return
  saving.value = true
  try {
    const res = await api.put('/api/v1/product-relations', {
      productId: productId.value,
      relatedIds: selectedIds.value,
    })
    if (res.data.code === 0 || res.data.code === 200) {
      ElMessage.success('已保存')
      addVisible.value = false
      loadRelations()
    } else {
      ElMessage.error(res.data.message || '保存失败')
    }
  } finally {
    saving.value = false
  }
}

async function removeRelation(row: RelationRow) {
  const ids = relations.value.filter(r => r.relatedProductId !== row.relatedProductId).map(r => r.relatedProductId)
  const res = await api.put('/api/v1/product-relations', { productId: productId.value, relatedIds: ids })
  if (res.data.code === 0 || res.data.code === 200) {
    ElMessage.success('已移除')
    loadRelations()
  }
}
</script>

<style scoped>
.toolbar { display: flex; gap: 12px; margin-bottom: 12px; }
</style>
