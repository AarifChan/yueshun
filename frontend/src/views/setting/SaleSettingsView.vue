<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">销售设置</h2>
    </div>
    <el-card>
      <el-tabs v-model="tab">
        <el-tab-pane label="流程配置" name="process">
          <SettingsForm v-if="tab === 'process'" :fields="processFields" />
        </el-tab-pane>
        <el-tab-pane label="订货流程设置" name="orderProcess">
          <el-alert type="info" :closable="false" title="订货流程：开始 → 提交订单 → 收货信息必填 → 交货时间 → 运费 → 审核 → 出库 → 发货 → 收货 → 订货完成 → 结束" style="margin-bottom: 16px" />
          <SettingsForm v-if="tab === 'orderProcess'" :fields="orderProcessFields" />
        </el-tab-pane>
        <el-tab-pane label="业务管控" name="control">
          <SettingsForm v-if="tab === 'control'" :fields="controlFields" />
        </el-tab-pane>
        <el-tab-pane label="价格设置" name="price">
          <SettingsForm v-if="tab === 'price'" :fields="priceFields" />
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import SettingsForm, { type SettingField } from './SettingsForm.vue'

const tab = ref('process')

const processFields: SettingField[] = [
  { key: 'sale.flow.submitOrder', label: '提交订单', type: 'switch', default: '1', hint: '订货流程起始环节' },
  { key: 'sale.flow.requireAddress', label: '收货信息必填', type: 'switch' },
  { key: 'sale.flow.deliveryTime', label: '订单/出库单交货时间设置', type: 'select', default: 'optional', options: [
    { label: '可选填', value: 'optional' },
    { label: '必填', value: 'required' },
    { label: '不显示', value: 'hidden' },
  ] },
  { key: 'sale.flow.freight', label: '运费设置', type: 'select', default: 'none', options: [
    { label: '不收运费', value: 'none' },
    { label: '固定运费', value: 'fixed' },
    { label: '按金额阶梯', value: 'tiered' },
  ] },
  { key: 'sale.flow.needAudit', label: '审核', type: 'switch', hint: '开启后订单需审核才能出库' },
  { key: 'sale.flow.enableOutStock', label: '出库', type: 'switch', default: '1' },
  { key: 'sale.flow.enableDelivery', label: '发货', type: 'switch', default: '1' },
  { key: 'sale.flow.enableReceive', label: '收货确认', type: 'switch', default: '1' },
  { key: 'sale.flow.orderComplete', label: '订货完成', type: 'switch', default: '1' },
]

const orderProcessFields: SettingField[] = [
  { key: 'sale.op.mallAutoAudit', label: '商城订单自动过审设置', type: 'select', default: 'manual', options: [
    { label: '人工审核', value: 'manual' },
    { label: '支付成功自动过审', value: 'paid' },
    { label: '提交即自动过审', value: 'submit' },
  ] },
  { key: 'sale.op.mallNeedAudit', label: '商城订单需要审核', type: 'switch', default: '1' },
  { key: 'sale.op.mallAutoOutStock', label: '过审后自动生成出库单', type: 'switch' },
]

const controlFields: SettingField[] = [
  { key: 'sale.ctl.minOrderQty', label: '商品起订量', type: 'number', min: 0, default: '0', hint: '0 表示不限制' },
  { key: 'sale.ctl.maxOrderQty', label: '商品限订量', type: 'number', min: 0, default: '0', hint: '0 表示不限制' },
  { key: 'sale.ctl.minOrderAmount', label: '订单起订金额', type: 'number', min: 0, precision: 2, default: '0' },
  { key: 'sale.ctl.returnOverTip', label: '销售退货超可退数量提示', type: 'switch', default: '1' },
  { key: 'sale.ctl.integerQty', label: '开单时商品数量只能整数', type: 'switch' },
  { key: 'sale.ctl.creditLimit', label: '客户超信用额度设置', type: 'select', default: 'tip', options: [
    { label: '仅提示', value: 'tip' },
    { label: '禁止开单', value: 'forbid' },
    { label: '不控制', value: 'none' },
  ] },
  { key: 'sale.ctl.expenseAllocate', label: '销售费用分摊至商品', type: 'switch', hint: '将销售费用按金额比例分摊到商品成本' },
  { key: 'sale.ctl.handlerRequired', label: '单据经手人必填', type: 'switch' },
  { key: 'sale.ctl.invoice', label: '发票设置', type: 'select', default: 'optional', options: [
    { label: '可选', value: 'optional' },
    { label: '必填', value: 'required' },
    { label: '不启用', value: 'none' },
  ] },
  { key: 'sale.ctl.saleScope', label: '商品销售范围设置', type: 'select', default: 'all', options: [
    { label: '全部商品可售', value: 'all' },
    { label: '仅销售范围内商品可售', value: 'scope' },
  ] },
  { key: 'sale.ctl.forbidUnitProxy', label: '代开销售订单/出库单时禁购单位可开单', type: 'switch' },
  { key: 'sale.ctl.importAutoUnit', label: '销售类单据批量导入时无单位自动匹配', type: 'switch', default: '1' },
]

const priceFields: SettingField[] = [
  { key: 'sale.price.belowCost', label: '商品售价低于成本价设置', type: 'select', default: 'tip', options: [
    { label: '仅提示', value: 'tip' },
    { label: '禁止开单', value: 'forbid' },
    { label: '不控制', value: 'none' },
  ] },
  { key: 'sale.price.belowMin', label: '商品售价低于最低售价设置', type: 'select', default: 'tip', options: [
    { label: '仅提示', value: 'tip' },
    { label: '禁止开单', value: 'forbid' },
    { label: '不控制', value: 'none' },
  ] },
  { key: 'sale.price.track', label: '销售价格跟踪', type: 'switch', default: '1', hint: '开单时自动带入该客户最近售价' },
]
</script>

<style scoped>
.page { display: flex; flex-direction: column; gap: 12px; }
.page-header { display: flex; justify-content: space-between; align-items: center; }
.page-title { font-size: 18px; font-weight: 600; margin: 0; }
</style>
