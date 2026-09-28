<template>
  <div class="dashboard">
    <el-row :gutter="16">
      <el-col :span="18">
        <el-card shadow="never" class="dash-card">
          <template #header>
            <div class="card-header">
              <span class="card-title">常用功能</span>
            </div>
          </template>
          <div class="quick-grid">
            <div v-for="fn in quickFunctions" :key="fn.title" class="quick-item" @click="$router.push(fn.path)">
              <div class="quick-icon" :style="{ backgroundColor: fn.color }">
                <el-icon :size="24" color="#fff"><component :is="fn.icon" /></el-icon>
              </div>
              <span class="quick-label">{{ fn.title }}</span>
            </div>
          </div>
        </el-card>

        <el-card shadow="never" class="dash-card">
          <template #header>
            <div class="card-header">
              <span class="card-title">交易统计</span>
            </div>
          </template>
          <div class="stat-grid">
            <div v-for="stat in stats" :key="stat.title" class="stat-item">
              <p class="stat-value" :style="{ color: stat.color }">{{ stat.value }}</p>
              <p class="stat-title">{{ stat.title }}</p>
            </div>
          </div>
        </el-card>

        <el-card shadow="never" class="dash-card">
          <template #header>
            <div class="card-header">
              <span class="card-title">库存预警商品</span>
              <el-button link type="primary" size="small" @click="$router.push('/inventory-warnings')">更多</el-button>
            </div>
          </template>
          <el-table :data="warningList" size="small" :show-header="false">
            <template #empty><el-empty description="暂无库存预警" :image-size="60" /></template>
            <el-table-column prop="name" label="商品" />
            <el-table-column prop="warehouseName" label="仓库" width="140" />
            <el-table-column prop="quantity" label="当前库存" width="110" align="right" />
          </el-table>
        </el-card>
      </el-col>

      <el-col :span="6">
        <el-card shadow="never" class="dash-card">
          <template #header>
            <div class="card-header"><span class="card-title">系统信息</span></div>
          </template>
          <el-descriptions :column="1" border size="small">
            <el-descriptions-item label="系统名称">智账系统</el-descriptions-item>
            <el-descriptions-item label="版本">v1.0.0</el-descriptions-item>
            <el-descriptions-item label="当前用户">{{ authStore.user?.name || authStore.user?.username || '-' }}</el-descriptions-item>
            <el-descriptions-item label="登录时间">{{ currentTime }}</el-descriptions-item>
          </el-descriptions>
        </el-card>

        <el-card shadow="never" class="dash-card">
          <template #header>
            <div class="card-header"><span class="card-title">快捷入口</span></div>
          </template>
          <div class="side-links">
            <div v-for="link in sideLinks" :key="link.title" class="side-link" @click="$router.push(link.path)">
              <el-icon :color="link.color"><component :is="link.icon" /></el-icon>
              <span>{{ link.title }}</span>
              <el-icon class="arrow"><ArrowRight /></el-icon>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '@/stores/auth'
import api from '@/api/client'

const authStore = useAuthStore()
const currentTime = ref(new Date().toLocaleString())

const quickFunctions = [
  { title: '商品管理', icon: 'Goods', color: '#409EFF', path: '/products' },
  { title: '客户管理', icon: 'User', color: '#67C23A', path: '/customers' },
  { title: '库存查询', icon: 'Coin', color: '#E6A23C', path: '/stocks' },
  { title: '价格体系', icon: 'PriceTag', color: '#909399', path: '/price-levels' },
  { title: '采购订单', icon: 'ShoppingCart', color: '#F56C6C', path: '/purchase-orders' },
  { title: '销售订单', icon: 'ShoppingBag', color: '#2f54eb', path: '/sales-orders' },
]

const sideLinks = [
  { title: '商城商品', icon: 'Shop', color: '#409EFF', path: '/mall-products' },
  { title: '商城订单', icon: 'Document', color: '#67C23A', path: '/mall-orders' },
  { title: '审批流程', icon: 'SetUp', color: '#E6A23C', path: '/approval-processes' },
  { title: '审批记录', icon: 'List', color: '#909399', path: '/approval-records' },
]

const stats = ref([
  { title: '商品总数', value: '0', color: '#409EFF' },
  { title: '客户总数', value: '0', color: '#67C23A' },
  { title: '待处理单据', value: '0', color: '#E6A23C' },
  { title: '库存预警', value: '0', color: '#F56C6C' },
  { title: '今日销售额', value: '¥0.00', color: '#409EFF' },
  { title: '本月销售额', value: '¥0.00', color: '#67C23A' },
  { title: '累计应收欠款', value: '¥0.00', color: '#F56C6C' },
])

const warningList = ref<any[]>([])

onMounted(async () => {
  try {
    const res = await api.get('/api/v1/stats/dashboard')
    if (res.data.code === 0 || res.data.code === 200) {
      const d = res.data.data || {}
      const money = (v: unknown) => `¥${Number(v ?? 0).toFixed(2)}`
      stats.value = [
        { title: '商品总数', value: String(d.productCount ?? 0), color: '#409EFF' },
        { title: '客户总数', value: String(d.customerCount ?? 0), color: '#67C23A' },
        { title: '待处理单据', value: String(d.pendingBills ?? 0), color: '#E6A23C' },
        { title: '库存预警', value: String(d.warningCount ?? 0), color: '#F56C6C' },
        { title: '今日销售额', value: money(d.todaySales), color: '#409EFF' },
        { title: '本月销售额', value: money(d.monthSales), color: '#67C23A' },
        { title: '累计应收欠款', value: money(d.totalReceivable), color: '#F56C6C' },
      ]
    }
  } catch (error: any) {
    ElMessage.error(error?.message || '仪表盘数据加载失败')
  }

  try {
    const res = await api.get('/api/v1/inventory-warnings', { params: { page: 1, pageSize: 5 } })
    if (res.data.code === 0 || res.data.code === 200) {
      const data = res.data.data
      warningList.value = Array.isArray(data) ? data : (data?.list ?? [])
    }
  } catch {
    warningList.value = []
  }
})
</script>

<style scoped>
.dashboard {
  padding: 4px;
}
.dash-card {
  margin-bottom: 16px;
}
.dash-card :deep(.el-card__header) {
  padding: 12px 16px;
}
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.card-title {
  font-size: 15px;
  font-weight: 600;
  color: #303133;
  position: relative;
  padding-left: 10px;
}
.card-title::before {
  content: '';
  position: absolute;
  left: 0;
  top: 50%;
  transform: translateY(-50%);
  width: 3px;
  height: 14px;
  background: #2f54eb;
  border-radius: 2px;
}
.quick-grid {
  display: grid;
  grid-template-columns: repeat(6, 1fr);
  gap: 12px;
}
.quick-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 12px 0;
  border-radius: 6px;
  cursor: pointer;
  transition: background-color 0.2s;
}
.quick-item:hover {
  background: #f5f7fa;
}
.quick-icon {
  width: 48px;
  height: 48px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
}
.quick-label {
  font-size: 13px;
  color: #606266;
}
.stat-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 8px;
}
.stat-item {
  text-align: center;
  padding: 14px 0;
  border-radius: 6px;
  background: #f7f8fa;
}
.stat-value {
  font-size: 22px;
  font-weight: bold;
  margin: 0 0 6px;
}
.stat-title {
  font-size: 13px;
  color: #909399;
  margin: 0;
}
.side-links {
  display: flex;
  flex-direction: column;
}
.side-link {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 4px;
  font-size: 13px;
  color: #606266;
  cursor: pointer;
  border-bottom: 1px solid #f0f0f0;
}
.side-link:last-child {
  border-bottom: none;
}
.side-link:hover {
  color: #2f54eb;
}
.side-link .arrow {
  margin-left: auto;
  color: #c0c4cc;
}
</style>
