<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">仓配设置</h2>
    </div>
    <el-card>
      <el-alert type="info" :closable="false" title="配送流程说明" style="margin-bottom: 16px" />
      <div class="flow">
        <div v-for="(s, i) in states" :key="s" class="flow-node">
          <div class="node">{{ s }}</div>
          <el-icon v-if="i < states.length - 1" class="arrow"><ArrowRight /></el-icon>
        </div>
      </div>
      <el-divider />
      <SettingsForm :fields="fields" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import SettingsForm, { type SettingField } from './SettingsForm.vue'

const states = [
  '订货', '部分出库', '待验货', '验货中', '验货完成',
  '待处理', '处理中', '待供应商发货', '待平台出库', '待平台收货',
  '待向客户发货', '客户待收货', '待客户收货', '配送完成', '收货完成', '已完成',
]

const fields: SettingField[] = [
  { key: 'dist.enableInspect', label: '启用验货环节', type: 'switch', hint: '开启后出库商品需验货才能发货' },
  { key: 'dist.platformDelivery', label: '启用平台配送', type: 'switch' },
  { key: 'dist.supplierDirectShip', label: '允许供应商直发', type: 'switch' },
]
</script>

<style scoped>
.page { display: flex; flex-direction: column; gap: 12px; }
.page-header { display: flex; justify-content: space-between; align-items: center; }
.page-title { font-size: 18px; font-weight: 600; margin: 0; }
.flow { display: flex; flex-wrap: wrap; gap: 8px 4px; padding: 8px 0; }
.flow-node { display: flex; align-items: center; gap: 4px; }
.node {
  padding: 6px 14px; border: 1px solid #dcdfe6; border-radius: 16px;
  font-size: 13px; color: #606266; background: #f5f7fa; white-space: nowrap;
}
.arrow { color: #c0c4cc; }
</style>
