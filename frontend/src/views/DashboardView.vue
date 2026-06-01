<template>
  <div class="dashboard">
    <h2>欢迎使用智账系统</h2>
    <el-row :gutter="20">
      <el-col :span="6" v-for="stat in stats" :key="stat.title">
        <el-card shadow="hover">
          <div class="stat-card">
            <el-icon :size="40" :color="stat.color">
              <component :is="stat.icon" />
            </el-icon>
            <div class="stat-info">
              <p class="stat-title">{{ stat.title }}</p>
              <p class="stat-value">{{ stat.value }}</p>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" style="margin-top: 20px">
      <el-col :span="12">
        <el-card title="快捷入口">
          <template #header>
            <span>快捷入口</span>
          </template>
          <div class="quick-links">
            <el-button type="primary" @click="$router.push('/products')">
              <el-icon><Goods /></el-icon> 商品管理
            </el-button>
            <el-button type="success" @click="$router.push('/customers')">
              <el-icon><User /></el-icon> 客户管理
            </el-button>
            <el-button type="warning" @click="$router.push('/purchase-orders')">
              <el-icon><ShoppingCart /></el-icon> 采购订单
            </el-button>
            <el-button type="danger" @click="$router.push('/sales-orders')">
              <el-icon><ShoppingBag /></el-icon> 销售订单
            </el-button>
          </div>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card>
          <template #header>
            <span>系统信息</span>
          </template>
          <el-descriptions :column="1" border>
            <el-descriptions-item label="系统名称">智账系统</el-descriptions-item>
            <el-descriptions-item label="版本">v1.0.0</el-descriptions-item>
            <el-descriptions-item label="当前用户">{{ authStore.user?.name || '-' }}</el-descriptions-item>
            <el-descriptions-item label="登录时间">{{ currentTime }}</el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const currentTime = ref(new Date().toLocaleString())

const stats = ref([
  { title: '商品数量', value: '0', icon: 'Goods', color: '#409EFF' },
  { title: '客户数量', value: '0', icon: 'User', color: '#67C23A' },
  { title: '待处理订单', value: '0', icon: 'ShoppingCart', color: '#E6A23C' },
  { title: '库存预警', value: '0', icon: 'Warning', color: '#F56C6C' },
])
</script>

<style scoped>
.dashboard h2 {
  margin-bottom: 20px;
  color: #303133;
}
.stat-card {
  display: flex;
  align-items: center;
  gap: 16px;
}
.stat-info {
  flex: 1;
}
.stat-title {
  font-size: 14px;
  color: #909399;
  margin: 0 0 8px;
}
.stat-value {
  font-size: 24px;
  font-weight: bold;
  color: #303133;
  margin: 0;
}
.quick-links {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}
.quick-links .el-button {
  display: flex;
  align-items: center;
  gap: 6px;
}
</style>
