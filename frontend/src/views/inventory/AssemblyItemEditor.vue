<template>
  <div class="item-editor">
    <el-table :data="modelValue" border size="small">
      <el-table-column type="index" width="46" />
      <el-table-column label="商品" min-width="220">
        <template #default="{ row }">
          <el-select
            v-model="row.productId" filterable remote :remote-method="searchProducts" :loading="loading"
            placeholder="商品名称/编号" style="width: 100%" size="small"
            @focus="searchProducts('')" @change="(val: number) => onChange(row, val)"
          >
            <el-option v-for="p in options" :key="p.id" :label="`${p.name}${p.specification ? ' / ' + p.specification : ''}`" :value="p.id" />
          </el-select>
        </template>
      </el-table-column>
      <el-table-column label="数量" width="140">
        <template #default="{ row }">
          <el-input-number v-model="row.quantity" :min="0" :precision="2" controls-position="right" style="width: 100%" size="small" />
        </template>
      </el-table-column>
      <el-table-column label="操作" width="64" align="center">
        <template #default="{ $index }">
          <el-button type="danger" link size="small" @click="removeRow($index)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-button size="small" style="margin-top: 6px" @click="addRow">添加行</el-button>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import api from '@/api/client'

interface ItemEntry { productId?: number; productName?: string; quantity: number }

const modelValue = defineModel<ItemEntry[]>({ default: () => [] })

const options = ref<any[]>([])
const loading = ref(false)

async function searchProducts(keyword: string) {
  loading.value = true
  try {
    const res = await api.get('/api/v1/products', { params: { keyword, pageSize: 50 } })
    if (res.data.code === 0 || res.data.code === 200) options.value = res.data.data?.list ?? res.data.data?.items ?? []
  } finally { loading.value = false }
}

function onChange(row: ItemEntry, productId: number) {
  const p = options.value.find((x) => x.id === productId)
  if (p) row.productName = p.name
}

function addRow() { modelValue.value = [...modelValue.value, { quantity: 1 }] }
function removeRow(index: number) { modelValue.value = modelValue.value.filter((_, i) => i !== index) }
</script>

<style scoped>
.item-editor { width: 100%; }
</style>
