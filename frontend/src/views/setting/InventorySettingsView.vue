<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">库存设置</h2>
    </div>
    <el-card>
      <el-tabs v-model="tab">
        <el-tab-pane label="成本设置" name="cost">
          <SettingsForm v-if="tab === 'cost'" :fields="costFields" />
        </el-tab-pane>
        <el-tab-pane label="库存设置" name="stock">
          <SettingsForm v-if="tab === 'stock'" :fields="stockFields" />
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import SettingsForm, { type SettingField } from './SettingsForm.vue'

const tab = ref('cost')

const costFields: SettingField[] = [
  { key: 'stock.cost.allocatePurchaseExpense', label: '采购费用分摊至商品入库成本', type: 'switch', hint: '开启后采购费用按金额比例计入入库商品成本' },
  { key: 'stock.cost.method', label: '成本核算方式', type: 'select', default: 'moving_avg', options: [
    { label: '移动加权平均', value: 'moving_avg' },
    { label: '先进先出', value: 'fifo' },
  ] },
]

const stockFields: SettingField[] = [
  { key: 'stock.availableRule', label: '可用库存设置', type: 'select', default: 'book_minus_locked', options: [
    { label: '账面库存 - 锁定库存', value: 'book_minus_locked' },
    { label: '账面库存', value: 'book' },
    { label: '账面库存 - 待出库', value: 'book_minus_pending' },
  ] },
  { key: 'stock.availableControl', label: '可用库存管控设置', type: 'select', default: 'tip', options: [
    { label: '不足时仅提示', value: 'tip' },
    { label: '不足时禁止开单', value: 'forbid' },
    { label: '不控制', value: 'none' },
  ] },
  { key: 'stock.bookControl', label: '账面库存管控设置', type: 'select', default: 'none', options: [
    { label: '允许负库存', value: 'none' },
    { label: '不允许负库存', value: 'forbid' },
  ] },
  { key: 'stock.assemblyEqualAmount', label: '组装拆装单出入库金额相等时才允许过账', type: 'switch' },
  { key: 'stock.pdaCheckMode', label: 'PDA 盘点方式设置', type: 'select', default: 'blind', options: [
    { label: '盲盘（不显示账面数）', value: 'blind' },
    { label: '明盘（显示账面数）', value: 'open' },
  ] },
]
</script>

<style scoped>
.page { display: flex; flex-direction: column; gap: 12px; }
.page-header { display: flex; justify-content: space-between; align-items: center; }
.page-title { font-size: 18px; font-weight: 600; margin: 0; }
</style>
