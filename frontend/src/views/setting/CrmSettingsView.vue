<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">CRM 设置</h2>
    </div>
    <el-card>
      <SettingsForm :fields="fields" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import SettingsForm, { type SettingField } from './SettingsForm.vue'

const fields: SettingField[] = [
  { key: 'crm.uncontactedDays', label: '未联系天数统计口径', type: 'number', min: 1, max: 365, default: '30', hint: '超过该天数未联系计入未联系客户' },
  { key: 'crm.reportAffectsUncontacted', label: '新增汇报影响未联系天数', type: 'switch', default: '1' },
  { key: 'crm.orderAffectsUncontacted', label: '新增订单影响未联系天数', type: 'switch', default: '1' },
  { key: 'crm.newCustomerNeedAudit', label: '手动新增客户需要审核', type: 'switch' },
  { key: 'crm.quickCreateOnBill', label: '开单未找到客户时快捷创建客户', type: 'switch', default: '1' },
  { key: 'crm.duplicateCheck', label: '客户/供应商查重', type: 'select', default: 'name', options: [
    { label: '按名称查重', value: 'name' },
    { label: '按手机号查重', value: 'phone' },
    { label: '名称 + 手机号查重', value: 'both' },
    { label: '不查重', value: 'none' },
  ] },
  { key: 'crm.oneCodeExpireHours', label: '一客一码时效（小时）', type: 'number', min: 1, max: 720, default: '72' },
  { key: 'crm.overdueTradeRemindDays', label: '超期未交易提醒（天）', type: 'number', min: 0, max: 365, default: '0', hint: '0 表示不提醒' },
  { key: 'crm.overdueContactRemindDays', label: '超期未联系提醒（天）', type: 'number', min: 0, max: 365, default: '0', hint: '0 表示不提醒' },
  { key: 'crm.poolAllowBilling', label: '客户公海是否允许开单', type: 'switch' },
  { key: 'crm.debtToPool', label: '欠款客户是否掉入公海', type: 'switch' },
  { key: 'crm.poolVisibleFields', label: '未领取公海客户可查看字段', type: 'select', default: 'name', options: [
    { label: '仅客户名称', value: 'name' },
    { label: '名称 + 公司地址', value: 'name_address' },
    { label: '名称 + 地址 + 联系人', value: 'name_address_contact' },
    { label: '名称 + 地址 + 联系人 + 收货信息', value: 'all' },
  ] },
]
</script>

<style scoped>
.page { display: flex; flex-direction: column; gap: 12px; }
.page-header { display: flex; justify-content: space-between; align-items: center; }
.page-title { font-size: 18px; font-weight: 600; margin: 0; }
</style>
