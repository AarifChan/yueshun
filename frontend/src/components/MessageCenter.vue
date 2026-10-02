<template>
  <el-popover placement="bottom-end" :width="660" trigger="click" @show="onShow">
    <template #reference>
      <span class="msg-entry">
        <el-badge :value="unreadTotal" :hidden="unreadTotal === 0" :max="99">
          <el-icon :size="18"><Bell /></el-icon>
        </el-badge>
      </span>
    </template>

    <div class="msg-panel">
      <div class="msg-side">
        <div
          v-for="cat in categories"
          :key="cat.key"
          class="msg-cat"
          :class="{ active: cat.key === activeCategory }"
          @click="switchCategory(cat.key)"
        >
          <span>{{ cat.label }}</span>
          <el-badge
            :value="unreadByCategory[cat.key] || 0"
            :hidden="!unreadByCategory[cat.key]"
            :max="99"
            class="cat-badge"
          />
        </div>
      </div>

      <div class="msg-main">
        <div class="msg-toolbar">
          <el-radio-group v-model="activeType" size="small" @change="loadList(1)">
            <el-radio-button v-for="f in currentFilters" :key="f.value" :value="f.value">{{ f.label }}</el-radio-button>
          </el-radio-group>
        </div>
        <div class="msg-actions">
          <el-button link type="primary" size="small" @click="readAllCategory">全部已读</el-button>
          <el-button link type="primary" size="small" @click="readAll">一键已读</el-button>
          <el-dropdown trigger="click" placement="bottom-end">
            <el-button link size="small" class="voice-entry">
              语音提醒设置<el-icon><ArrowDown /></el-icon>
            </el-button>
            <template #dropdown>
              <div class="voice-panel">
                <div class="voice-row">
                  <span>新订单语音提醒</span>
                  <el-switch v-model="voice.newOrderVoice" @change="saveVoice" />
                </div>
                <div class="voice-row">
                  <span>新审批语音提醒</span>
                  <el-switch v-model="voice.newApprovalVoice" @change="saveVoice" />
                </div>
              </div>
            </template>
          </el-dropdown>
        </div>

        <div v-loading="loading" class="msg-list">
          <template v-if="list.length">
            <div
              v-for="m in list"
              :key="m.id"
              class="msg-item"
              :class="{ unread: !m.read }"
              @click="readOne(m)"
            >
              <span class="msg-dot" />
              <span class="msg-title">{{ m.title }}</span>
              <span class="msg-time">{{ formatTime(m.createdAt) }}</span>
            </div>
          </template>
          <el-empty v-else description="暂无数据" :image-size="60" />
        </div>

        <el-pagination
          v-if="total > pageSize"
          v-model:current-page="page"
          :page-size="pageSize"
          :total="total"
          layout="prev, pager, next"
          small
          class="msg-pagination"
          @current-change="loadList"
        />
      </div>
    </div>
  </el-popover>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Bell, ArrowDown } from '@element-plus/icons-vue'
import api from '@/api/client'

// 分类与筛选项：docs/功能文档/13-企业微信通知.md 实际看到的结构
interface Filter { value: string; label: string }
interface Category { key: string; label: string; filters: Filter[] }

const ALL: Filter = { value: '', label: '全部' }
const categories: Category[] = [
  { key: 'bill', label: '单据消息', filters: [{ value: 'order', label: '全部订单' }, { value: 'return_payment', label: '退单收付款单' }] },
  { key: 'customer', label: '客户消息', filters: [ALL, { value: 'new_audit', label: '新客户审核' }, { value: 'contact_overdue', label: '联系超期预警' }, { value: 'trade_overdue', label: '交易超期预警' }, { value: 'orderer_audit', label: '订货人审核' }] },
  { key: 'approval', label: '审批消息', filters: [ALL, { value: 'mine', label: '我提交的' }, { value: 'pending', label: '待处理' }, { value: 'cc', label: '抄送我的' }, { value: 'mention', label: '@我的' }] },
  { key: 'report', label: '汇报消息', filters: [ALL, { value: 'cc', label: '抄送' }, { value: 'remind', label: '提醒' }, { value: 'comment', label: '评论' }] },
  { key: 'system', label: '系统消息', filters: [ALL, { value: 'notice', label: '系统提示' }, { value: 'announcement', label: '系统公告' }] },
  { key: 'eco', label: '生态互联', filters: [ALL, { value: 'connection', label: '连接' }, { value: 'mall', label: '商城' }, { value: 'product', label: '商品' }, { value: 'bill', label: '单据' }] },
  { key: 'biz', label: '业务提醒', filters: [ALL, { value: 'biz_report', label: '经营报表' }, { value: 'stock_alarm', label: '库存报警' }, { value: 'fund', label: '资金管理' }] },
  { key: 'print', label: '打印消息', filters: [ALL, { value: 'print', label: '打印' }] },
  { key: 'warehouse', label: '仓配消息', filters: [ALL, { value: 'shelving', label: '上架' }] },
]

interface MessageItem {
  id: number; category: string; type: string; title: string
  content: string; read: boolean; createdAt: string
}

const activeCategory = ref('bill')
const activeType = ref(categories[0].filters[0].value)
const list = ref<MessageItem[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 20
const loading = ref(false)
const unreadTotal = ref(0)
const unreadByCategory = ref<Record<string, number>>({})

const currentFilters = computed(() => categories.find((c) => c.key === activeCategory.value)?.filters ?? [])

async function loadUnread() {
  try {
    const res = await api.get('/api/v1/messages/unread-count')
    if (res.data.code === 0 || res.data.code === 200) {
      unreadTotal.value = res.data.data?.total ?? 0
      unreadByCategory.value = res.data.data?.byCategory ?? {}
    }
  } catch { /* 未读数失败不打扰 */ }
}

async function loadList(p = 1) {
  page.value = p
  loading.value = true
  try {
    const res = await api.get('/api/v1/messages', {
      params: {
        category: activeCategory.value,
        type: activeType.value || undefined,
        page: page.value,
        pageSize,
      },
    })
    if (res.data.code === 0 || res.data.code === 200) {
      list.value = res.data.data?.list ?? []
      total.value = res.data.data?.total ?? 0
    }
  } finally {
    loading.value = false
  }
}

function switchCategory(key: string) {
  activeCategory.value = key
  activeType.value = currentFilters.value[0]?.value ?? ''
  loadList(1)
}

async function readOne(m: MessageItem) {
  if (m.read) return
  m.read = true
  try {
    await api.put(`/api/v1/messages/${m.id}/read`)
    loadUnread()
  } catch { /* 已读失败保持本地状态 */ }
}

async function readAllCategory() {
  await api.put('/api/v1/messages/read-all', null, { params: { category: activeCategory.value } })
  ElMessage.success('已全部已读')
  loadList(1)
  loadUnread()
}

async function readAll() {
  await api.put('/api/v1/messages/read-all')
  ElMessage.success('已一键已读')
  loadList(1)
  loadUnread()
}

// ---- 语音提醒设置（文档默认值：新订单开、新审批关）----
const voice = reactive({ newOrderVoice: true, newApprovalVoice: false })
let voiceLoaded = false

async function loadVoice() {
  try {
    const res = await api.get('/api/v1/message-settings/voice')
    if (res.data.code === 0 || res.data.code === 200) {
      voice.newOrderVoice = res.data.data?.newOrderVoice ?? true
      voice.newApprovalVoice = res.data.data?.newApprovalVoice ?? false
      voiceLoaded = true
    }
  } catch { /* 保留默认值 */ }
}

async function saveVoice() {
  if (!voiceLoaded) return
  try {
    await api.put('/api/v1/message-settings/voice', {
      newOrderVoice: voice.newOrderVoice,
      newApprovalVoice: voice.newApprovalVoice,
    })
    ElMessage.success('已保存')
  } catch {
    ElMessage.error('保存失败')
  }
}

function onShow() {
  loadUnread()
  loadList(1)
  if (!voiceLoaded) loadVoice()
}

function formatTime(t: string) {
  return t ? new Date(t).toLocaleString('zh-CN', { hour12: false }) : ''
}

onMounted(loadUnread)
</script>

<style scoped>
.msg-entry {
  display: flex;
  align-items: center;
  padding: 6px 10px;
  border-radius: 6px;
  cursor: pointer;
  color: #fff;
  transition: background-color 0.2s;
}
.msg-entry:hover {
  background-color: rgba(255, 255, 255, 0.15);
}
.msg-panel {
  display: flex;
  height: 420px;
}
.msg-side {
  width: 120px;
  flex-shrink: 0;
  border-right: 1px solid #ebeef5;
  overflow-y: auto;
}
.msg-cat {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 12px;
  font-size: 13px;
  color: #606266;
  cursor: pointer;
}
.msg-cat:hover {
  background: #f5f7fa;
}
.msg-cat.active {
  color: #2f54eb;
  background: #f0f5ff;
  font-weight: 500;
}
.msg-main {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  padding-left: 12px;
}
.msg-toolbar {
  overflow-x: auto;
  padding-bottom: 4px;
}
.msg-actions {
  display: flex;
  align-items: center;
  gap: 4px;
  border-bottom: 1px solid #f0f0f0;
  padding-bottom: 4px;
}
.voice-entry {
  margin-left: auto;
  color: #606266;
}
.voice-panel {
  padding: 8px 16px;
}
.voice-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 8px 0;
  font-size: 13px;
  color: #303133;
  white-space: nowrap;
}
.msg-list {
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
}
.msg-list .el-empty {
  flex: 1;
}
.msg-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 4px;
  border-bottom: 1px solid #f5f5f5;
  cursor: pointer;
  font-size: 13px;
  color: #909399;
}
.msg-item.unread {
  color: #303133;
}
.msg-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: transparent;
  flex-shrink: 0;
}
.msg-item.unread .msg-dot {
  background: #f56c6c;
}
.msg-title {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.msg-item.unread .msg-title {
  font-weight: 500;
}
.msg-time {
  flex-shrink: 0;
  font-size: 12px;
  color: #c0c4cc;
}
.msg-pagination {
  justify-content: flex-end;
  padding-top: 8px;
}
.cat-badge :deep(.el-badge__content) {
  position: static;
  transform: none;
}
</style>
