<template>
  <el-container class="layout-container">
    <el-header class="top-header" height="50px">
      <div class="logo">
        <el-icon :size="26" color="#ffffff"><Box /></el-icon>
        <span>智账系统</span>
      </div>
      <nav class="top-nav">
        <div
          v-for="item in menuTree"
          :key="item.key"
          class="nav-item"
          :class="{ 'is-active': isNavActive(item) }"
          @mouseenter="openNav(item, $event)"
          @mouseleave="scheduleCloseNav"
          @click="onNavClick(item)"
        >
          <el-icon :size="16"><component :is="item.icon || 'Menu'" /></el-icon>
          <span>{{ item.title }}</span>
          <el-icon v-if="hasPanel(item)" class="nav-arrow" :size="12"><ArrowDown /></el-icon>
        </div>
      </nav>
      <div class="header-right">
        <MessageCenter />
        <el-dropdown @command="handleCommand">
          <span class="user-info">
            <el-icon><UserFilled /></el-icon>
            {{ authStore.user?.name || authStore.user?.username || '用户' }}
            <el-icon><ArrowDown /></el-icon>
          </span>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="logout">退出登录</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </el-header>

    <div class="tab-bar">
      <div class="tabs">
        <div
          v-for="tab in tabs"
          :key="tab.path"
          class="tab-item"
          :class="{ active: tab.path === route.path }"
          @click="goTab(tab.path)"
        >
          {{ tab.title }}
          <el-icon v-if="tab.path !== '/dashboard'" class="tab-close" @click.stop="closeTab(tab.path)"><Close /></el-icon>
        </div>
      </div>
      <div class="tab-actions">
        <el-tooltip content="刷新当前页"><el-icon @click="reload"><Refresh /></el-icon></el-tooltip>
        <el-tooltip content="关闭其他页签"><el-icon @click="closeOthers"><CircleClose /></el-icon></el-tooltip>
      </div>
    </div>

    <el-main class="layout-main">
      <div class="page-body">
        <router-view v-slot="{ Component }">
          <component :is="Component" :key="route.fullPath" />
        </router-view>
      </div>
    </el-main>

    <teleport to="body">
      <div
        v-if="navOpen"
        class="nav-panel"
        :style="{ left: navPos.left + 'px', top: navPos.top + 'px' }"
        @mouseenter="cancelCloseNav"
        @mouseleave="scheduleCloseNav"
      >
        <template v-if="navOpen.mega">
          <div v-for="col in navOpen.mega" :key="col.title" class="mega-col">
            <div class="mega-col-title">
              <el-icon :size="17"><component :is="col.icon || 'Menu'" /></el-icon>
              <span>{{ col.title }}</span>
            </div>
            <div
              v-for="entry in col.items"
              :key="entry.path + entry.title"
              class="mega-item"
              :class="{ active: route.path === entry.path }"
              @click="goNav(entry.path)"
            >
              {{ entry.title }}
            </div>
          </div>
        </template>
        <div v-else class="nav-panel-simple">
          <div
            v-for="child in navOpen.children"
            :key="child.path"
            class="mega-item"
            :class="{ active: route.path === child.path }"
            @click="goNav(child.path)"
          >
            <el-icon :size="15"><component :is="child.icon || 'Document'" /></el-icon>
            <span>{{ child.title }}</span>
          </div>
        </div>
      </div>
    </teleport>
  </el-container>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessageBox } from 'element-plus'
import { useAuthStore } from '@/stores/auth'
import MessageCenter from '@/components/MessageCenter.vue'
import router from '@/router'

const route = useRoute()
const vueRouter = useRouter()
const authStore = useAuthStore()

interface MenuItem {
  key: string
  title: string
  icon?: string
  path?: string
  mega?: { title: string; icon?: string; items: { title: string; path: string }[] }[]
  children: { path: string; title: string; icon?: string }[]
}

const menuTree = computed<MenuItem[]>(() => {
  const layout = router.options.routes.find((r) => r.path === '/')
  const visible = (r: any) => r.meta?.title && !r.meta?.hidden
  return (layout?.children || [])
    .filter(visible)
    .map((r: any) => {
      const children = (r.children || []).filter(visible).map((c: any) => ({
        path: `/${c.path}`,
        title: c.meta?.title,
        icon: c.meta?.icon,
      }))
      return {
        key: r.name ? String(r.name) : `/${r.path}`,
        title: r.meta?.title,
        icon: r.meta?.icon,
        path: r.path ? `/${r.path}` : undefined,
        mega: r.meta?.mega,
        children,
      }
    })
})

// 页签
interface Tab { path: string; fullPath: string; title: string }
const tabs = ref<Tab[]>([{ path: '/dashboard', fullPath: '/dashboard', title: '首页' }])

watch(() => route.fullPath, () => {
  if (!route.meta?.title) return
  if (!tabs.value.some((t) => t.fullPath === route.fullPath)) {
    tabs.value.push({ path: route.path, fullPath: route.fullPath, title: String(route.meta.title) })
  }
}, { immediate: true })

// 导航下拉面板
const navOpen = ref<MenuItem | null>(null)
const navPos = ref({ left: 0, top: 0 })
let navTimer: ReturnType<typeof setTimeout> | undefined

function hasPanel(item: MenuItem) {
  return !!item.mega || item.children.length > 0
}

// 当前路由所属 mega 分组
const activeMegaKey = computed(() => {
  const item = menuTree.value.find((m) =>
    m.mega?.some((col) => col.items.some((entry) => entry.path === route.path))
  )
  return item?.key || ''
})

function isNavActive(item: MenuItem) {
  const path = route.meta?.hidden && typeof route.path === 'string'
    ? route.path.replace(/\/\d+$/, '')
    : route.path
  if (item.path) return path === item.path
  if (item.mega) return item.key === activeMegaKey.value
  return item.children.some((c) => c.path === path)
}

function openNav(item: MenuItem, event: MouseEvent) {
  if (!hasPanel(item)) return
  cancelCloseNav()
  const rect = (event.currentTarget as HTMLElement).getBoundingClientRect()
  const cols = item.mega ? item.mega.length : 1
  const width = item.mega ? Math.max(420, cols * 220) : 220
  navPos.value = {
    left: Math.max(8, Math.min(rect.left, window.innerWidth - width - 8)),
    top: rect.bottom,
  }
  navOpen.value = item
}

function onNavClick(item: MenuItem) {
  if (item.path) goNav(item.path)
}

function scheduleCloseNav() {
  cancelCloseNav()
  navTimer = setTimeout(() => { navOpen.value = null }, 150)
}

function cancelCloseNav() {
  if (navTimer) {
    clearTimeout(navTimer)
    navTimer = undefined
  }
}

function goNav(path: string) {
  navOpen.value = null
  if (path !== route.path) vueRouter.push(path)
}

watch(() => route.fullPath, () => {
  navOpen.value = null
})

function goTab(path: string) {
  if (path !== route.path) vueRouter.push(path)
}

function closeTab(path: string) {
  const idx = tabs.value.findIndex((t) => t.path === path)
  if (idx === -1) return
  tabs.value.splice(idx, 1)
  if (route.path === path) {
    const next = tabs.value[idx - 1] || tabs.value[0]
    vueRouter.push(next.fullPath)
  }
}

function closeOthers() {
  tabs.value = tabs.value.filter((t) => t.path === '/dashboard' || t.fullPath === route.fullPath)
}

function reload() {
  vueRouter.replace(route.fullPath)
}

async function handleCommand(command: string) {
  if (command === 'logout') {
    try {
      await ElMessageBox.confirm('确认退出登录？', '提示', { type: 'warning' })
    } catch {
      return
    }
    await authStore.logout()
    vueRouter.push('/login')
  }
}
</script>

<style scoped>
.layout-container {
  height: 100vh;
  display: flex;
  flex-direction: column;
}
.top-header {
  display: flex;
  align-items: center;
  background-color: #2f54eb;
  padding: 0 16px;
}
.logo {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #fff;
  font-size: 17px;
  font-weight: 600;
  margin-right: 24px;
  white-space: nowrap;
}
.top-nav {
  flex: 1;
  min-width: 0;
  display: flex;
  align-items: center;
  overflow-x: auto;
  overflow-y: hidden;
  scrollbar-width: none;
}
.top-nav::-webkit-scrollbar {
  display: none;
}
.nav-item {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 7px 14px;
  border-radius: 4px;
  font-size: 14px;
  color: rgba(255, 255, 255, 0.85);
  cursor: pointer;
  white-space: nowrap;
  flex-shrink: 0;
  transition: background-color 0.2s;
}
.nav-item:hover {
  background-color: rgba(255, 255, 255, 0.1);
}
.nav-item.is-active {
  background-color: #1d39c4;
  color: #ffffff;
}
.nav-item .nav-arrow {
  margin-left: -2px;
}
.header-right {
  margin-left: 16px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: 8px;
}
.user-info {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 10px;
  border-radius: 6px;
  cursor: pointer;
  color: #fff;
  transition: background-color 0.2s;
}
.user-info:hover {
  background-color: rgba(255, 255, 255, 0.15);
}
.tab-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 40px;
  background: #f5f7fa;
  border-bottom: 1px solid #e4e7ed;
  padding: 0 12px;
}
.tabs {
  display: flex;
  align-items: center;
  gap: 2px;
  overflow-x: auto;
}
.tab-item {
  display: flex;
  align-items: center;
  gap: 6px;
  height: 28px;
  line-height: 28px;
  padding: 0 14px;
  font-size: 13px;
  color: #606266;
  border-radius: 4px;
  cursor: pointer;
  white-space: nowrap;
  transition: all 0.2s;
  position: relative;
}
.tab-item:hover {
  color: #2f54eb;
  background: #e8edf5;
}
.tab-item.active {
  color: #2f54eb;
  background: #fff;
  font-weight: 500;
  box-shadow: 0 1px 4px rgba(47, 84, 235, 0.15);
}
.tab-close {
  font-size: 12px;
  border-radius: 50%;
  padding: 1px;
}
.tab-close:hover {
  background: rgba(47, 84, 235, 0.15);
}
.tab-actions {
  display: flex;
  align-items: center;
  gap: 14px;
  color: #909399;
  cursor: pointer;
}
.tab-actions .el-icon:hover {
  color: #409eff;
}
.layout-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: #f0f2f5;
  padding: 0;
  overflow: hidden;
}
.page-body {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
}
.nav-panel {
  position: fixed;
  z-index: 3000;
  display: flex;
  gap: 24px;
  padding: 20px 28px;
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 6px 24px rgba(0, 0, 0, 0.18);
  border: 1px solid #ebeef5;
}
.nav-panel-simple {
  display: flex;
  flex-direction: column;
  min-width: 170px;
}
.nav-panel-simple .mega-item {
  display: flex;
  align-items: center;
  gap: 8px;
}
.mega-col {
  min-width: 150px;
}
.mega-col-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 15px;
  font-weight: 600;
  color: #303133;
  padding-bottom: 12px;
  margin-bottom: 6px;
  border-bottom: 1px solid #f0f0f0;
}
.mega-col-title .el-icon {
  color: #2f54eb;
}
.mega-item {
  padding: 9px 8px;
  font-size: 14px;
  color: #606266;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.15s;
}
.mega-item:hover {
  color: #2f54eb;
  background: #f0f5ff;
}
.mega-item.active {
  color: #2f54eb;
  font-weight: 500;
}
</style>
