<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">商品近效期预警</h2>
      <div>
        <el-button @click="openMsgSetting">消息设置</el-button>
        <el-button @click="doExport">导出</el-button>
        <el-button type="primary" @click="search">查询</el-button>
      </div>
    </div>
    <el-card>
      <el-form inline class="search-form">
        <el-form-item label="商品/批号">
          <el-input v-model="filters.keyword" placeholder="批号/商品名称/编号/条码/规格" clearable style="width: 220px" />
        </el-form-item>
        <el-form-item label="仓库">
          <RemoteSelect v-model="filters.warehouseId" api-url="/api/v1/warehouses" placeholder="全部" />
        </el-form-item>
        <el-form-item label="预警天数">
          <el-input-number v-model="filters.days" :min="1" :max="365" style="width: 110px" />
        </el-form-item>
      </el-form>
      <el-table :data="list" v-loading="loading" border stripe>
        <el-table-column prop="productName" label="商品名称" min-width="150" fixed="left" />
        <el-table-column prop="specification" label="规格" width="100" />
        <el-table-column prop="batchNo" label="批号" width="110" />
        <el-table-column prop="produceDate" label="生产日期" width="100" />
        <el-table-column prop="shelfLifeDays" label="保质期(天)" width="90" align="right" />
        <el-table-column prop="expiryDate" label="到期日期" width="100" />
        <el-table-column label="剩余有效天数" width="110" align="right">
          <template #default="{ row }">
            <span :class="{ 'warn-text': row.remainDays <= 30 }">{{ row.remainDays }}</span>
          </template>
        </el-table-column>
        <el-table-column label="过期天数" width="90" align="right">
          <template #default="{ row }">
            <span v-if="row.expiredDays > 0" class="expired-text">{{ row.expiredDays }}</span>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="unit" label="单位" width="64" />
        <el-table-column prop="warehouseName" label="仓库" width="100" />
        <el-table-column prop="quantity" label="数量" width="90" align="right" />
        <el-table-column label="成本价" width="90" align="right">
          <template #default="{ row }">{{ row.costPrice?.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="成本金额" width="110" align="right">
          <template #default="{ row }">{{ row.costAmount?.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column prop="supplierName" label="批次供应商" width="110" />
        <el-table-column prop="code" label="商品编号" width="110" />
        <template #empty>暂无数据</template>
      </el-table>
      <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" :page-sizes="[30,50,100]" layout="total, sizes, prev, pager, next" @size-change="() => load(1)" @current-change="load" class="pagination" />
    </el-card>

    <el-dialog v-model="msgSettingVisible" title="消息设置" width="560px">
      <el-alert type="info" :closable="false" title="管理员和系统管理员默认将收到消息" style="margin-bottom: 16px" />
      <el-form v-loading="msgSettingLoading" label-width="90px">
        <el-form-item label="临近效期">
          <div class="notify-row">
            <el-switch v-model="msgSetting.warnEnabled" />
            <span class="notify-desc">商品过期前，接近效期报警天数发送预警通知，可添加接收人</span>
          </div>
          <RemoteSelect
            v-model="msgSetting.warnReceivers"
            api-url="/api/v1/employees"
            placeholder="添加接收人"
            multiple
            class="notify-receivers"
          />
        </el-form-item>
        <el-form-item label="过期通知">
          <div class="notify-row">
            <el-switch v-model="msgSetting.expiredEnabled" />
            <span class="notify-desc">商品过期后发送过期通知，可添加接收人</span>
          </div>
          <RemoteSelect
            v-model="msgSetting.expiredReceivers"
            api-url="/api/v1/employees"
            placeholder="添加接收人"
            multiple
            class="notify-receivers"
          />
        </el-form-item>
        <el-form-item label="发送时间">
          <el-select v-model="msgSetting.sendHour" style="width: 120px">
            <el-option v-for="h in 24" :key="h - 1" :label="`${h - 1} 时`" :value="h - 1" />
          </el-select>
          <span class="notify-desc" style="margin-left: 8px">每天该小时开始发送消息</span>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="msgSettingVisible = false">取消</el-button>
        <el-button type="primary" :loading="msgSettingSaving" @click="saveMsgSetting">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import RemoteSelect from '@/components/RemoteSelect.vue'
import { useReport } from '@/composables/useReport'
import api from '@/api/client'

interface Row {
  id: number; productName: string; specification: string; batchNo: string
  produceDate?: string; shelfLifeDays: number; expiryDate?: string
  remainDays: number; expiredDays: number
  unit: string; warehouseName: string; quantity: number; costPrice: number; costAmount: number
  supplierName: string; code: string
}

const { list, total, loading, page, pageSize, filters, load, search, exportCsv } = useReport<Row>(
  '/api/v1/inventory-reports/expiry-warning',
  { keyword: '', warehouseId: undefined, days: 90 },
)

function doExport() {
  exportCsv('商品近效期预警', [
    { key: 'productName', label: '商品名称' }, { key: 'batchNo', label: '批号' },
    { key: 'produceDate', label: '生产日期' }, { key: 'shelfLifeDays', label: '保质期' },
    { key: 'expiryDate', label: '到期日期' }, { key: 'remainDays', label: '剩余有效天数' },
    { key: 'expiredDays', label: '过期天数' }, { key: 'warehouseName', label: '仓库' },
    { key: 'quantity', label: '数量' }, { key: 'costAmount', label: '成本金额' },
  ])
}

onMounted(() => load(1))

// ---- 消息设置（管理员和系统管理员默认将收到消息）----
const msgSettingVisible = ref(false)
const msgSettingLoading = ref(false)
const msgSettingSaving = ref(false)
const msgSetting = ref({
  warnEnabled: false,
  warnReceivers: [] as number[],
  expiredEnabled: false,
  expiredReceivers: [] as number[],
  sendHour: 8,
})

async function openMsgSetting() {
  msgSettingVisible.value = true
  msgSettingLoading.value = true
  try {
    const res = await api.get('/api/v1/message-settings/expiry-warning')
    if (res.data.code === 0 || res.data.code === 200) {
      const d = res.data.data || {}
      msgSetting.value = {
        warnEnabled: d.warnEnabled ?? false,
        warnReceivers: d.warnReceivers ?? [],
        expiredEnabled: d.expiredEnabled ?? false,
        expiredReceivers: d.expiredReceivers ?? [],
        sendHour: d.sendHour ?? 8,
      }
    }
  } finally {
    msgSettingLoading.value = false
  }
}

async function saveMsgSetting() {
  msgSettingSaving.value = true
  try {
    const res = await api.put('/api/v1/message-settings/expiry-warning', msgSetting.value)
    if (res.data.code === 0 || res.data.code === 200) {
      ElMessage.success('已保存')
      msgSettingVisible.value = false
    } else {
      ElMessage.error(res.data.message || '保存失败')
    }
  } finally {
    msgSettingSaving.value = false
  }
}
</script>

<style scoped>
.page { display: flex; flex-direction: column; gap: 12px; }
.page-header { display: flex; justify-content: space-between; align-items: center; }
.page-title { font-size: 18px; font-weight: 600; margin: 0; }
.warn-text { color: var(--el-color-warning); font-weight: 600; }
.expired-text { color: var(--el-color-danger); font-weight: 600; }
.pagination { margin-top: 16px; justify-content: flex-end; }
.notify-row { display: flex; align-items: center; gap: 10px; }
.notify-desc { font-size: 12px; color: #909399; }
.notify-receivers { margin-top: 8px; width: 100%; }
</style>
