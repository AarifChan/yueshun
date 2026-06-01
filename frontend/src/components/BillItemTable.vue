<template>
  <div>
    <div class="table-header">
      <span>商品明细</span>
      <el-button type="primary" size="small" @click="$emit('add')" :disabled="!editable">添加行</el-button>
    </div>
    <el-table :data="items" border size="small">
      <el-table-column type="index" width="50" />
      <el-table-column label="商品" min-width="200">
        <template #default="{ row }">
          <RemoteSelect
            v-if="editable"
            v-model="row.productId"
            api-url="/api/v1/products"
            placeholder="选择商品"
          />
          <span v-else>{{ row.productName || row.productId }}</span>
        </template>
      </el-table-column>
      <el-table-column label="数量" width="120">
        <template #default="{ row }">
          <el-input-number v-if="editable" v-model="row.quantity" :min="0" :precision="2" controls-position="right" style="width: 100%" />
          <span v-else>{{ row.quantity }}</span>
        </template>
      </el-table-column>
      <el-table-column label="单价" width="120">
        <template #default="{ row }">
          <el-input-number v-if="editable" v-model="row.price" :min="0" :precision="2" controls-position="right" style="width: 100%" />
          <span v-else>{{ row.price }}</span>
        </template>
      </el-table-column>
      <el-table-column label="金额" width="120">
        <template #default="{ row }">
          {{ ((row.quantity || 0) * (row.price || 0)).toFixed(2) }}
        </template>
      </el-table-column>
      <el-table-column label="备注" min-width="150">
        <template #default="{ row }">
          <el-input v-if="editable" v-model="row.remark" size="small" />
          <span v-else>{{ row.remark }}</span>
        </template>
      </el-table-column>
      <el-table-column v-if="editable" label="操作" width="80" fixed="right">
        <template #default="{ $index }">
          <el-button type="danger" size="small" @click="$emit('remove', $index)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
    <div class="summary">
      <span>合计数量: {{ totalQuantity.toFixed(2) }}</span>
      <span>合计金额: {{ totalAmount.toFixed(2) }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import RemoteSelect from './RemoteSelect.vue'

interface Props {
  items: Record<string, any>[]
  editable: boolean
}

const props = defineProps<Props>()
defineEmits<{
  (e: 'add'): void
  (e: 'remove', index: number): void
}>()

const totalQuantity = computed(() =>
  props.items.reduce((sum, item) => sum + (item.quantity || 0), 0)
)

const totalAmount = computed(() =>
  props.items.reduce((sum, item) => sum + (item.quantity || 0) * (item.price || 0), 0)
)
</script>

<style scoped>
.table-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}
.summary {
  margin-top: 8px;
  text-align: right;
  font-weight: bold;
}
.summary span {
  margin-left: 24px;
}
</style>
