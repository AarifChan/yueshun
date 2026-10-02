<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">商城设置</h2>
    </div>
    <el-card>
      <el-tabs v-model="tab">
        <el-tab-pane label="商城装修" name="decoration">
          <template v-if="tab === 'decoration'">
            <el-alert type="info" :closable="false" style="margin-bottom: 12px"
              title="装修方案可通过「生态互联 → 商城装修接收」从上游接收并应用，也可在小程序端可视化编辑。" />
            <el-descriptions :column="2" border>
              <el-descriptions label="当前模板">{{ decoration ? '已应用自定义模板' : '默认模板' }}</el-descriptions>
              <el-descriptions label="模板大小">{{ decorationSize }}</el-descriptions>
            </el-descriptions>
            <div style="margin-top: 12px">
              <el-button type="danger" plain :disabled="!decoration" @click="resetDecoration">恢复默认模板</el-button>
            </div>
          </template>
        </el-tab-pane>
        <el-tab-pane label="商品分类样式" name="category">
          <SettingsForm v-if="tab === 'category'" :fields="categoryFields" />
        </el-tab-pane>
        <el-tab-pane label="列表/详情样式" name="listStyle">
          <SettingsForm v-if="tab === 'listStyle'" :fields="listStyleFields" />
        </el-tab-pane>
        <el-tab-pane label="个人中心设置" name="personal">
          <SettingsForm v-if="tab === 'personal'" :fields="personalFields" />
        </el-tab-pane>
        <el-tab-pane label="全局风格" name="global">
          <SettingsForm v-if="tab === 'global'" :fields="globalFields" />
        </el-tab-pane>
        <el-tab-pane label="关联商品" name="relation">
          <RelationPanel v-if="tab === 'relation'" />
        </el-tab-pane>
        <el-tab-pane label="专题分类" name="subject">
          <SubjectPanel v-if="tab === 'subject'" />
        </el-tab-pane>
        <el-tab-pane label="订货商城设置" name="shop">
          <SettingsForm v-if="tab === 'shop'" :fields="shopFields" />
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '@/api/client'
import SettingsForm, { type SettingField } from './SettingsForm.vue'
import RelationPanel from './MallRelationPanel.vue'
import SubjectPanel from './MallSubjectPanel.vue'

const tab = ref('decoration')

// ---- 商城装修（mall.decoration）----
const decoration = ref('')
const decorationSize = computed(() => {
  if (!decoration.value) return '-'
  const kb = decoration.value.length / 1024
  return kb >= 1 ? `${kb.toFixed(1)} KB` : `${decoration.value.length} B`
})

async function loadDecoration() {
  const res = await api.get('/api/v1/settings')
  if (res.data.code === 0 || res.data.code === 200) {
    decoration.value = res.data.data?.['mall.decoration'] || ''
  }
}

async function resetDecoration() {
  await ElMessageBox.confirm('确定恢复默认模板？当前自定义装修内容将被清除。', '提示', { type: 'warning' })
  const res = await api.put('/api/v1/settings', { 'mall.decoration': '' })
  if (res.data.code === 0 || res.data.code === 200) {
    ElMessage.success('已恢复默认模板')
    loadDecoration()
  }
}

onMounted(loadDecoration)

// ---- 商品分类样式 ----
const categoryFields: SettingField[] = [
  { key: 'mall.cat.levelStyle', label: '分类层级样式', type: 'select', default: 'level1', options: [
    { label: '只有 1 级', value: 'level1' },
    { label: '只有 2 级', value: 'level2' },
    { label: '有 3/4 级分类', value: 'level34' },
  ] },
  { key: 'mall.cat.listStyle', label: '列表样式', type: 'select', default: 'small', options: [
    { label: '小图样式', value: 'small' },
    { label: '大图样式', value: 'big' },
    { label: '文字样式', value: 'text' },
  ] },
  { key: 'mall.cat.searchScope', label: '小图模式默认搜索范围', type: 'select', default: 'selected', options: [
    { label: '在已选中分类中搜索', value: 'selected' },
    { label: '在全部分类中搜索', value: 'all' },
  ] },
  { key: 'mall.cat.autoSwitch', label: '触底下滑自动切换分类', type: 'switch' },
  { key: 'mall.cat.subjectOrder', label: '展示顺序', type: 'select', default: 'top', options: [
    { label: '专题分类在商品分类上方', value: 'top' },
    { label: '专题分类在商品分类下方', value: 'bottom' },
    { label: '混排', value: 'mixed' },
  ] },
  { key: 'mall.cat.shareTitle', label: '分享标题', type: 'text', placeholder: '分类页分享标题' },
  { key: 'mall.cat.shareCover', label: '分享封面', type: 'text', placeholder: '图片 URL' },
]

// ---- 列表/详情样式 ----
const listStyleFields: SettingField[] = [
  { key: 'mall.list.style', label: '列表样式', type: 'select', default: 'two', options: [
    { label: '一行两个', value: 'two' },
    { label: '一行一个', value: 'one' },
    { label: '无图模式', value: 'noImage' },
  ] },
  { key: 'mall.list.specStyle', label: '规格样式', type: 'select', default: 'combo', options: [
    { label: '组合', value: 'combo' },
    { label: '平铺', value: 'flat' },
  ] },
  { key: 'mall.list.showSpecImage', label: '显示规格图片', type: 'switch' },
  { key: 'mall.list.showSpecRemark', label: '显示规格备注', type: 'switch' },
  { key: 'mall.list.specSearch', label: '规格搜索', type: 'switch' },
  { key: 'mall.list.detailAddCart', label: '详情加购', type: 'switch', default: '1' },
  { key: 'mall.list.listAddCart', label: '列表加购', type: 'switch', default: '1' },
  { key: 'mall.list.sort', label: '列表排序', type: 'select', default: 'default', options: [
    { label: '默认排序', value: 'default' },
    { label: '按销量', value: 'sales' },
    { label: '按上架时间', value: 'newest' },
    { label: '按价格', value: 'price' },
  ] },
  { key: 'mall.list.shareTitle', label: '分享标题', type: 'text' },
  { key: 'mall.list.shareCover', label: '分享封面', type: 'text', placeholder: '图片 URL' },
]

// ---- 个人中心设置 ----
const personalFields: SettingField[] = [
  { key: 'mall.my.showPayable', label: '我的资产-应付款', type: 'switch', default: '1' },
  { key: 'mall.my.showCredit', label: '我的资产-信用额度', type: 'switch', default: '1' },
  { key: 'mall.my.showPrepay', label: '我的资产-预付款', type: 'switch', default: '1' },
  { key: 'mall.my.showPoints', label: '我的资产-积分', type: 'switch', default: '1' },
  { key: 'mall.my.showCoupons', label: '我的资产-优惠券', type: 'switch', default: '1' },
  { key: 'mall.my.orderPendingPay', label: '我的订单-待付款', type: 'switch', default: '1' },
  { key: 'mall.my.orderPendingAudit', label: '我的订单-待审核', type: 'switch', default: '1' },
  { key: 'mall.my.orderPendingDelivery', label: '我的订单-待发货', type: 'switch', default: '1' },
  { key: 'mall.my.orderPendingReceive', label: '我的订单-待收货', type: 'switch', default: '1' },
  { key: 'mall.my.orderReturns', label: '我的订单-退货单', type: 'switch', default: '1' },
  { key: 'mall.my.svcPointMall', label: '我的服务-积分商城', type: 'switch', default: '1' },
  { key: 'mall.my.svcCouponCenter', label: '我的服务-领券中心', type: 'switch', default: '1' },
  { key: 'mall.my.svcStatement', label: '我的服务-往来对账', type: 'switch', default: '1' },
  { key: 'mall.my.svcFavorites', label: '我的服务-收藏商品', type: 'switch', default: '1' },
  { key: 'mall.my.svcInvoice', label: '我的服务-发票信息', type: 'switch' },
  { key: 'mall.my.svcAddress', label: '我的服务-收货信息', type: 'switch', default: '1' },
  { key: 'mall.my.svcKefu', label: '我的服务-在线客服', type: 'switch' },
  { key: 'mall.my.svcDownload', label: '我的服务-下载中心', type: 'switch' },
  { key: 'mall.my.allowPublicPrepay', label: '允许使用公共预收支付', type: 'switch' },
  { key: 'mall.my.allowBrandPrepay', label: '允许使用公共预收及品牌预收支付', type: 'switch' },
]

// ---- 全局风格 ----
const globalFields: SettingField[] = [
  { key: 'mall.style.themeColor', label: '主题颜色', type: 'text', placeholder: '#FF6600', default: '#FF6600' },
  { key: 'mall.style.headerBgColor', label: '顶部背景颜色', type: 'text', placeholder: '#FFFFFF', default: '#FFFFFF' },
  { key: 'mall.style.headerContentColor', label: '顶部内容颜色', type: 'text', placeholder: '#333333', default: '#333333' },
]

// ---- 订货商城设置 ----
const shopFields: SettingField[] = [
  { key: 'mall.shop.allowRetailRegister', label: '允许零售客户注册', type: 'switch' },
  { key: 'mall.shop.registerDefaultType', label: '客户注册默认类型', type: 'select', default: 'retail', options: [
    { label: '零售客户', value: 'retail' },
    { label: '批发客户', value: 'wholesale' },
  ] },
  { key: 'mall.shop.registerNeedAudit', label: '订货人初次登录需审核', type: 'switch' },
  { key: 'mall.shop.nonOrderCustomerView', label: '非订货客户的查看设置', type: 'select', default: 'all', options: [
    { label: '可查看全部商品', value: 'all' },
    { label: '不可查看商品', value: 'none' },
  ] },
  { key: 'mall.shop.hidePrice', label: '不允许查看价格', type: 'switch' },
  { key: 'mall.shop.businessHours', label: '商城营业时间', type: 'text', placeholder: '如 08:00-22:00，留空为全天' },
  { key: 'mall.shop.auxUnitMinMultiple', label: '辅助单位起订量倍数', type: 'number', min: 0, default: '0' },
  { key: 'mall.shop.stockFallbackQty', label: '库存不足起订量时按库存下单', type: 'switch' },
  { key: 'mall.shop.stockDisplay', label: '库存显示', type: 'select', default: 'text', options: [
    { label: '显示具体数量', value: 'exact' },
    { label: '仅显示有/无货', value: 'text' },
    { label: '不显示', value: 'none' },
  ] },
  { key: 'mall.shop.autoOnShelf', label: '可用库存大于 0 自动上架', type: 'switch' },
  { key: 'mall.shop.autoOffShelf', label: '可用库存小于等于 0 自动下架', type: 'switch' },
  { key: 'mall.shop.imageRatio', label: '商品图片比例', type: 'select', default: '1:1', options: [
    { label: '1:1', value: '1:1' },
    { label: '3:4', value: '3:4' },
    { label: '4:3', value: '4:3' },
  ] },
  { key: 'mall.shop.watermark', label: '商品图片水印', type: 'switch' },
  { key: 'mall.shop.showSales', label: '显示销量', type: 'switch', default: '1' },
  { key: 'mall.shop.showLinePrice', label: '显示划线价', type: 'switch' },
  { key: 'mall.shop.showRetailPrice', label: '显示建议零售价', type: 'switch' },
  { key: 'mall.shop.allowReturn', label: '售后-允许申请退货', type: 'switch', default: '1' },
  { key: 'mall.shop.allowCancel', label: '售后-允许取消订单', type: 'switch', default: '1' },
]
</script>

<style scoped>
.page { display: flex; flex-direction: column; gap: 12px; }
.page-header { display: flex; justify-content: space-between; align-items: center; }
.page-title { font-size: 18px; font-weight: 600; margin: 0; }
</style>
