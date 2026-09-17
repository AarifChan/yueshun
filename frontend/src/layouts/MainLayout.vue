<template>
  <el-container class="main-layout">
    <el-aside width="220px" class="sidebar">
      <div class="logo">
        <el-icon :size="28" color="#409EFF"><Box /></el-icon>
        <span>智账系统</span>
      </div>
      <el-menu
        :default-active="activeMenu"
        router
        class="sidebar-menu"
        background-color="#304156"
        text-color="#bfcbd9"
        active-text-color="#409EFF"
        :collapse="isCollapse"
      >
        <el-menu-item index="/dashboard">
          <el-icon><HomeFilled /></el-icon>
          <span>首页</span>
        </el-menu-item>

        <el-sub-menu index="/base">
          <template #title>
            <el-icon><Setting /></el-icon>
            <span>基础资料</span>
          </template>
          <el-menu-item index="/dicts">字典管理</el-menu-item>
          <el-menu-item index="/departments">部门管理</el-menu-item>
          <el-menu-item index="/employees">职员管理</el-menu-item>
          <el-menu-item index="/roles">角色管理</el-menu-item>
          <el-menu-item index="/permissions">权限管理</el-menu-item>
        </el-sub-menu>

        <el-sub-menu index="/product">
          <template #title>
            <el-icon><Goods /></el-icon>
            <span>商品资料</span>
          </template>
          <el-menu-item index="/products">商品管理</el-menu-item>
          <el-menu-item index="/categories">商品分类</el-menu-item>
          <el-menu-item index="/brands">品牌管理</el-menu-item>
        </el-sub-menu>

        <el-sub-menu index="/customer">
          <template #title>
            <el-icon><User /></el-icon>
            <span>客户/供应商</span>
          </template>
          <el-menu-item index="/customers">客户管理</el-menu-item>
          <el-menu-item index="/customer-categories">客户分类</el-menu-item>
          <el-menu-item index="/regions">区域管理</el-menu-item>
          <el-menu-item index="/levels">客户等级</el-menu-item>
        </el-sub-menu>

        <el-sub-menu index="/warehouse">
          <template #title>
            <el-icon><House /></el-icon>
            <span>仓库/资金</span>
          </template>
          <el-menu-item index="/warehouses">仓库管理</el-menu-item>
          <el-menu-item index="/accounts">资金账户</el-menu-item>
          <el-menu-item index="/income-expense">收支项目</el-menu-item>
        </el-sub-menu>

        <el-menu-item index="/price-levels">
          <el-icon><PriceTag /></el-icon>
          <span>价格体系</span>
        </el-menu-item>

        <el-sub-menu index="/inventory">
          <template #title>
            <el-icon><Box /></el-icon>
            <span>库存模块</span>
          </template>
          <el-menu-item index="/inventory-checks">库存盘点</el-menu-item>
          <el-menu-item index="/inventory-transfers">库存调拨</el-menu-item>
          <el-menu-item index="/inventory-warnings">库存预警</el-menu-item>
        </el-sub-menu>

        <el-sub-menu index="/purchase">
          <template #title>
            <el-icon><ShoppingCart /></el-icon>
            <span>采购模块</span>
          </template>
          <el-menu-item index="/purchase-orders">采购订单</el-menu-item>
          <el-menu-item index="/purchase-instock">采购入库</el-menu-item>
          <el-menu-item index="/purchase-returns">采购退货</el-menu-item>
          <el-menu-item index="/purchase-payments">采购付款</el-menu-item>
        </el-sub-menu>

        <el-sub-menu index="/sales">
          <template #title>
            <el-icon><ShoppingBag /></el-icon>
            <span>销售模块</span>
          </template>
          <el-menu-item index="/sales-orders">销售订单</el-menu-item>
          <el-menu-item index="/sales-outstock">销售出库</el-menu-item>
          <el-menu-item index="/sales-returns">销售退货</el-menu-item>
          <el-menu-item index="/sales-receipts">销售收款</el-menu-item>
        </el-sub-menu>

        <el-sub-menu index="/mall">
          <template #title>
            <el-icon><Shop /></el-icon>
            <span>商城模块</span>
          </template>
          <el-menu-item index="/mall-products">商城商品</el-menu-item>
          <el-menu-item index="/mall-orders">商城订单</el-menu-item>
          <el-menu-item index="/mall-carts">购物车</el-menu-item>
        </el-sub-menu>

        <el-sub-menu index="/crm">
          <template #title>
            <el-icon><Phone /></el-icon>
            <span>CRM模块</span>
          </template>
          <el-menu-item index="/follow-ups">跟进记录</el-menu-item>
          <el-menu-item index="/opportunities">商机管理</el-menu-item>
          <el-menu-item index="/contracts">合同管理</el-menu-item>
        </el-sub-menu>

        <el-sub-menu index="/approval">
          <template #title>
            <el-icon><SetUp /></el-icon>
            <span>审批模块</span>
          </template>
          <el-menu-item index="/approval-processes">审批流程</el-menu-item>
          <el-menu-item index="/approval-records">审批记录</el-menu-item>
        </el-sub-menu>
      </el-menu>
    </el-aside>

    <el-container>
      <el-header class="header">
        <div class="header-left">
          <el-icon class="collapse-btn" @click="isCollapse = !isCollapse">
            <Fold v-if="!isCollapse" />
            <Expand v-else />
          </el-icon>
          <breadcrumb />
        </div>
        <div class="header-right">
          <el-dropdown @command="handleCommand">
            <span class="user-info">
              <el-icon><UserFilled /></el-icon>
              {{ authStore.user?.name || '用户' }}
              <el-icon><ArrowDown /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="profile">个人中心</el-dropdown-item>
                <el-dropdown-item divided command="logout">退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>
      <el-main class="main-content">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { ElMessage } from 'element-plus'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const isCollapse = ref(false)

const activeMenu = computed(() => route.path)

function handleCommand(command: string) {
  if (command === 'logout') {
    authStore.logout()
    ElMessage.success('已退出登录')
    router.push('/login')
  } else if (command === 'profile') {
    router.push('/profile')
  }
}
</script>

<style scoped>
.main-layout {
  min-height: 100vh;
}
.sidebar {
  background-color: #304156;
  transition: width 0.3s;
}
.logo {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  color: #fff;
  font-size: 18px;
  font-weight: bold;
  border-bottom: 1px solid #1f2d3d;
}
.sidebar-menu {
  border-right: none;
  height: calc(100vh - 60px);
}
.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background-color: #fff;
  box-shadow: 0 1px 4px rgba(0, 21, 41, 0.08);
}
.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}
.collapse-btn {
  font-size: 20px;
  cursor: pointer;
  color: #606266;
}
.header-right {
  display: flex;
  align-items: center;
}
.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  color: #606266;
}
.main-content {
  background-color: #f0f2f5;
  padding: 20px;
  min-height: calc(100vh - 60px);
}
</style>
