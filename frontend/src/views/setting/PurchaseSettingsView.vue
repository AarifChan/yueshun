<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">采购设置</h2>
    </div>
    <el-card>
      <SettingsForm :fields="fields" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import SettingsForm, { type SettingField } from './SettingsForm.vue'

const fields: SettingField[] = [
  { key: 'purchase.autoInStock', label: '提交订单自动生成采购入库单并过账', type: 'switch', hint: '开启后采购订单提交即入库' },
  { key: 'purchase.useOutWarehouse', label: '采购时以商品出库仓库为采购入库仓库', type: 'switch' },
  { key: 'purchase.priceTrack', label: '采购价格跟踪', type: 'switch', default: '1', hint: '开单时自动带入最近采购价' },
  { key: 'purchase.prepayAccount', label: '启用预付款账户支付', type: 'switch', hint: '允许使用供应商预付款余额支付采购单' },
]
</script>

<style scoped>
.page { display: flex; flex-direction: column; gap: 12px; }
.page-header { display: flex; justify-content: space-between; align-items: center; }
.page-title { font-size: 18px; font-weight: 600; margin: 0; }
</style>
