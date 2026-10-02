<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">基础设置</h2>
    </div>
    <el-card>
      <el-tabs v-model="tab">
        <el-tab-pane label="功能配置" name="function">
          <SettingsForm v-if="tab === 'function'" :fields="functionFields" />
        </el-tab-pane>
        <el-tab-pane label="编号设置" name="number">
          <SettingsForm v-if="tab === 'number'" :fields="numberFields" />
        </el-tab-pane>
        <el-tab-pane label="开单设置" name="bill">
          <SettingsForm v-if="tab === 'bill'" :fields="billFields" />
        </el-tab-pane>
        <el-tab-pane label="聊天工具设置" name="chat">
          <el-alert type="info" :closable="false" title="聊天工具（侧边栏）使用说明" />
          <div class="chat-doc">
            <h3>场景说明</h3>
            <p>聊天工具侧边栏用于在企业微信/微信客服会话中快速查看客户信息、往来单据与商品，支持聊天中直接开单。</p>
            <h3>设置步骤</h3>
            <ol>
              <li>在企业微信管理后台「应用管理」中创建自建应用，获取 CorpID 与 Secret；</li>
              <li>在「企业设置 → 企业信息」中填写企业微信 CorpID；</li>
              <li>将侧边栏页面地址配置到企业微信聊天工具栏；</li>
              <li>职员在企业微信中登录后即可在会话侧边栏使用。</li>
            </ol>
          </div>
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import SettingsForm, { type SettingField } from './SettingsForm.vue'

const tab = ref('function')

const functionFields: SettingField[] = [
  { key: 'base.qtyDecimals', label: '数量小数位数', type: 'number', min: 0, max: 4, default: '2', hint: '单据中商品数量允许的小数位数' },
  { key: 'base.priceDecimals', label: '价格小数位数', type: 'number', min: 0, max: 4, default: '2' },
  { key: 'base.discountDecimals', label: '折扣小数位数', type: 'number', min: 0, max: 4, default: '2' },
  { key: 'base.thirdPartySettlement', label: '启用第三方结算', type: 'switch' },
  { key: 'base.accountPeriod', label: '启用账期', type: 'switch', hint: '开启后单据可设置账期与到期日' },
  { key: 'base.batchManage', label: '启用批次管理', type: 'switch', hint: '开启后商品支持批次号/生产日期管理' },
  { key: 'base.productRecommend', label: '商品推荐设置', type: 'select', default: 'none', options: [
    { label: '不启用', value: 'none' },
    { label: '按销量推荐', value: 'sales' },
    { label: '按关联商品推荐', value: 'relation' },
  ] },
]

const numberFields: SettingField[] = [
  { key: 'sn.productAuto', label: '商品自动编号', type: 'switch', default: '1' },
  { key: 'sn.productPrefix', label: '商品编号前缀', type: 'text', placeholder: '如 SP' },
  { key: 'sn.customerAuto', label: '客户自动编号', type: 'switch', default: '1' },
  { key: 'sn.customerPrefix', label: '客户编号前缀', type: 'text', placeholder: '如 KH' },
  { key: 'sn.supplierAuto', label: '供应商自动编号', type: 'switch', default: '1' },
  { key: 'sn.supplierPrefix', label: '供应商编号前缀', type: 'text', placeholder: '如 GYS' },
  { key: 'sn.employeeAuto', label: '职员自动编号', type: 'switch' },
  { key: 'sn.billRule', label: '单据编号生成规则', type: 'select', default: 'prefix_date_seq', options: [
    { label: '前缀 + 日期 + 流水号', value: 'prefix_date_seq' },
    { label: '前缀 + 流水号', value: 'prefix_seq' },
    { label: '日期 + 流水号', value: 'date_seq' },
  ] },
  { key: 'sn.approvalAuto', label: '审批自动编号', type: 'switch', default: '1' },
]

const billFields: SettingField[] = [
  { key: 'bill.enableDiscount', label: '启用折扣', type: 'switch', default: '1' },
  { key: 'bill.productSearchScope', label: '商品搜索设置', type: 'select', default: 'all', options: [
    { label: '搜索全部商品', value: 'all' },
    { label: '仅搜索有库存商品', value: 'inStock' },
    { label: '仅搜索上架商品', value: 'onShelf' },
  ] },
  { key: 'bill.continuousCreate', label: '启用单据连续新增', type: 'switch', hint: '保存单据后自动打开下一张' },
  { key: 'bill.draftUseSystemTime', label: '业务草稿录单时间取系统时间', type: 'switch' },
  { key: 'bill.datetimePrecision', label: '日期选择精确到时分', type: 'switch' },
  { key: 'bill.duplicateProduct', label: '开单重复选择商品设置', type: 'select', default: 'merge', options: [
    { label: '合并数量', value: 'merge' },
    { label: '新增一行', value: 'newRow' },
    { label: '禁止重复', value: 'forbid' },
  ] },
  { key: 'bill.listDefaultOrder', label: '单据列表默认显示顺序', type: 'select', default: 'date_desc', options: [
    { label: '按单据日期倒序', value: 'date_desc' },
    { label: '按创建时间倒序', value: 'created_desc' },
    { label: '按单据编号倒序', value: 'no_desc' },
  ] },
  { key: 'bill.defaultPayAmount', label: '开单页收/付款金额默认带入本单应收/应付金额', type: 'switch', default: '1' },
]
</script>

<style scoped>
.page { display: flex; flex-direction: column; gap: 12px; }
.page-header { display: flex; justify-content: space-between; align-items: center; }
.page-title { font-size: 18px; font-weight: 600; margin: 0; }
.chat-doc { padding: 16px 8px; line-height: 1.9; color: #606266; }
.chat-doc h3 { margin: 12px 0 4px; color: #303133; }
.chat-doc ol { padding-left: 20px; }
</style>
