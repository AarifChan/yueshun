<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">企业设置</h2>
    </div>
    <el-card>
      <el-tabs v-model="tab">
        <el-tab-pane label="企业信息" name="info">
          <el-form v-if="tab === 'info'" label-width="140px" v-loading="infoLoading" style="max-width: 640px">
            <el-form-item label="企业名称" required>
              <el-input v-model="info.name" />
            </el-form-item>
            <el-form-item label="企业简称">
              <el-input v-model="info.shortName" placeholder="用于商城、单据显示" />
            </el-form-item>
            <el-form-item label="企业编码">
              <el-input v-model="info.code" disabled />
            </el-form-item>
            <el-form-item label="企业微信 CorpID">
              <el-input v-model="info.wecomCorpId" placeholder="企业微信管理后台可查" />
            </el-form-item>
            <el-form-item label="联系人">
              <el-input v-model="info.contact" />
            </el-form-item>
            <el-form-item label="联系电话">
              <el-input v-model="info.phone" />
            </el-form-item>
            <el-form-item label="地址">
              <el-input v-model="info.address" />
            </el-form-item>
            <el-form-item label="行业">
              <el-select v-model="info.industry" style="width: 280px" clearable>
                <el-option v-for="i in industries" :key="i" :label="i" :value="i" />
              </el-select>
            </el-form-item>
            <el-form-item label="企业纳税性质">
              <el-radio-group v-model="info.taxNature">
                <el-radio value="general">一般纳税人</el-radio>
                <el-radio value="small">小规模纳税人</el-radio>
              </el-radio-group>
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :loading="infoSaving" @click="saveInfo">保存</el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>
        <el-tab-pane label="企业短信" name="sms">
          <template v-if="tab === 'sms'">
            <el-row :gutter="12" style="margin-bottom: 16px">
              <el-col :span="6"><el-statistic title="今日发送" :value="0" /></el-col>
              <el-col :span="6"><el-statistic title="昨日发送" :value="0" /></el-col>
              <el-col :span="6"><el-statistic title="累计发送" :value="0" /></el-col>
              <el-col :span="6"><el-statistic title="剩余可发送" :value="0" /></el-col>
            </el-row>
            <el-alert type="info" :closable="false"
              title="短信签名设置：企业可在短信服务商处申请专属签名（需提供营业执照等资质），审核通过后短信将以企业签名发送。短信条数管理与购买请联系客服开通。" />
          </template>
        </el-tab-pane>
        <el-tab-pane label="企业云空间" name="storage">
          <template v-if="tab === 'storage'">
            <el-row :gutter="12">
              <el-col :span="8"><el-statistic title="图片占用" :value="storageText.images" /></el-col>
              <el-col :span="8"><el-statistic title="视频占用" :value="storageText.videos" /></el-col>
              <el-col :span="8"><el-statistic title="附件占用" :value="storageText.files" /></el-col>
            </el-row>
            <el-alert type="info" :closable="false" title="云空间按类型统计上传文件的占用情况。" style="margin-top: 16px" />
          </template>
        </el-tab-pane>
        <el-tab-pane label="操作日志" name="logs">
          <div v-if="tab === 'logs'">
            <div class="filter-bar">
              <el-input v-model="logFilters.keyword" placeholder="搜索操作详情" clearable style="width: 220px" @keyup.enter="loadLogs" />
              <el-date-picker v-model="logRange" type="daterange" start-placeholder="开始日期" end-placeholder="结束日期" value-format="YYYY-MM-DD" style="width: 260px" />
              <el-select v-model="logFilters.objectType" placeholder="操作对象" clearable style="width: 160px">
                <el-option v-for="o in objectTypes" :key="o" :label="o" :value="o" />
              </el-select>
              <el-button type="primary" icon="Search" @click="loadLogs">查询</el-button>
              <el-button icon="Refresh" @click="loadLogs">刷新</el-button>
            </div>
            <el-table :data="logs" v-loading="logsLoading" border stripe>
              <el-table-column type="index" label="序" width="60" />
              <el-table-column prop="objectType" label="操作对象" width="120" />
              <el-table-column label="操作时间" width="170">
                <template #default="{ row }">{{ formatTime(row.createdAt) }}</template>
              </el-table-column>
              <el-table-column prop="operator" label="操作人" width="110">
                <template #default="{ row }">{{ row.operator || '-' }}</template>
              </el-table-column>
              <el-table-column prop="ip" label="IP地址" width="130">
                <template #default="{ row }">{{ row.ip || '-' }}</template>
              </el-table-column>
              <el-table-column prop="source" label="操作来源" width="100" align="center">
                <template #default="{ row }">
                  <el-tag size="small" :type="row.source === 'web' ? 'primary' : 'success'">{{ row.source === 'web' ? '网页端' : row.source }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="detail" label="操作详情" min-width="280" show-overflow-tooltip />
              <template #empty>暂无数据</template>
            </el-table>
            <el-pagination
              v-model:current-page="logPage"
              v-model:page-size="logPageSize"
              :total="logTotal"
              layout="total, prev, pager, next, sizes"
              :page-sizes="[30, 50, 100]"
              style="margin-top: 12px; justify-content: flex-end"
              @current-change="loadLogs"
              @size-change="loadLogs"
            />
          </div>
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import api from '@/api/client'

const tab = ref('info')

// ---- 企业信息 ----
const info = ref({
  name: '', shortName: '', code: '', wecomCorpId: '',
  contact: '', phone: '', address: '', industry: '', taxNature: 'general',
})
const infoLoading = ref(false)
const infoSaving = ref(false)
const industries = ['批发零售', '食品饮料', '服装鞋帽', '日化用品', '五金建材', '数码家电', '医药健康', '母婴用品', '其他']

async function loadInfo() {
  infoLoading.value = true
  try {
    const res = await api.get('/api/v1/company-info')
    if (res.data.code === 0 || res.data.code === 200) {
      const d = res.data.data || {}
      info.value = {
        name: d.name || '', shortName: d.shortName || '', code: d.code || '',
        wecomCorpId: d.wecomCorpId || '', contact: d.contact || '', phone: d.phone || '',
        address: d.address || '', industry: d.industry || '', taxNature: d.taxNature || 'general',
      }
    }
  } finally {
    infoLoading.value = false
  }
}

async function saveInfo() {
  if (!info.value.name.trim()) {
    ElMessage.warning('请填写企业名称')
    return
  }
  infoSaving.value = true
  try {
    const res = await api.put('/api/v1/company-info', info.value)
    if (res.data.code === 0 || res.data.code === 200) ElMessage.success('已保存')
    else ElMessage.error(res.data.message || '保存失败')
  } finally {
    infoSaving.value = false
  }
}

// ---- 企业云空间（静态统计占位）----
const storageText = { images: 0, videos: 0, files: 0 }

// ---- 操作日志 ----
interface LogRow {
  id: number; objectType: string; action: string; userId: number; operator: string
  ip: string; source: string; detail: string; createdAt: string
}

const logs = ref<LogRow[]>([])
const logsLoading = ref(false)
const logPage = ref(1)
const logPageSize = ref(30)
const logTotal = ref(0)
const logRange = ref<[string, string] | null>(null)
const logFilters = ref({ keyword: '', objectType: '' })
const objectTypes = [
  '商品', '客户', '供应商', '仓库', '职员', '部门', '角色', '系统设置',
  '采购订单', '采购入库单', '采购退货单', '销售订单', '销售出库单', '销售退货单',
  '盘点单', '调拨单', '其他入库单', '其他出库单', '商城订单',
  '费用单', '其他收入单', '现金银行账户', '营销活动', '优惠券', '分销商',
]

async function loadLogs() {
  logsLoading.value = true
  try {
    const res = await api.get('/api/v1/operation-logs', {
      params: {
        page: logPage.value,
        pageSize: logPageSize.value,
        keyword: logFilters.value.keyword || undefined,
        objectType: logFilters.value.objectType || undefined,
        startDate: logRange.value?.[0],
        endDate: logRange.value?.[1],
      },
    })
    if (res.data.code === 0 || res.data.code === 200) {
      logs.value = res.data.data?.list ?? []
      logTotal.value = res.data.data?.total ?? 0
    }
  } finally {
    logsLoading.value = false
  }
}

function formatTime(t: string) {
  return t ? new Date(t).toLocaleString('zh-CN', { hour12: false }) : '-'
}

watch(tab, (v) => {
  if (v === 'logs' && !logs.value.length) loadLogs()
})

onMounted(loadInfo)
</script>

<style scoped>
.page { display: flex; flex-direction: column; gap: 12px; }
.page-header { display: flex; justify-content: space-between; align-items: center; }
.page-title { font-size: 18px; font-weight: 600; margin: 0; }
.filter-bar { display: flex; gap: 10px; margin-bottom: 12px; flex-wrap: wrap; }
</style>
