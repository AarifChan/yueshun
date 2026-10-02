<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">调拨差异处理</h2>
      <el-button :icon="Refresh" @click="load()">刷新</el-button>
    </div>
    <el-card>
      <el-form inline class="search-form">
        <el-form-item label="调拨出库单号"><el-input v-model="search.keyword" placeholder="请输入调拨出库单号" clearable /></el-form-item>
        <el-form-item label="调拨类型">
          <el-select v-model="search.transferType" placeholder="请选择" clearable style="width: 130px">
            <el-option label="同价调拨" value="same" /><el-option label="变价调拨" value="diff" />
          </el-select>
        </el-form-item>
        <el-form-item label="商品"><el-input v-model="search.productKw" placeholder="商品名称/编号" clearable /></el-form-item>
        <el-form-item label="调出仓库">
          <RemoteSelect v-model="search.fromWarehouseId" api-url="/api/v1/warehouses" placeholder="请选择" />
        </el-form-item>
        <el-form-item label="录单时间">
          <el-date-picker v-model="dateRange" type="daterange" value-format="YYYY-MM-DD" start-placeholder="开始时间" end-placeholder="结束时间" style="width: 260px" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="load(1)">查询</el-button>
        </el-form-item>
      </el-form>
      <div class="check-row">
        <el-checkbox v-model="search.includeNoDiff" @change="load(1)">包含没有差异商品行</el-checkbox>
        <el-checkbox v-model="search.includeNotReceived" @change="load(1)">包含未入库商品行</el-checkbox>
      </div>
      <el-table :data="list" v-loading="loading" border stripe>
        <el-table-column label="操作" width="100" fixed="left">
          <template #default="{ row }">
            <el-button v-if="row.diffStatus === 'pending'" link type="primary" @click="openHandle(row)">处理</el-button>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="录单时间" width="170">
          <template #default="{ row }">{{ fmtTime(row.createdAt) }}</template>
        </el-table-column>
        <el-table-column prop="outBillNo" label="调拨出库单单号" width="170" />
        <el-table-column label="调拨类型" width="90" align="center">
          <template #default="{ row }">{{ row.transferType === 'same' ? '同价' : '变价' }}</template>
        </el-table-column>
        <el-table-column prop="fromWarehouseName" label="调出仓库" width="110" />
        <el-table-column prop="toWarehouseName" label="调入仓库" width="110" />
        <el-table-column prop="productName" label="商品名称" min-width="150" />
        <el-table-column prop="specification" label="商品规格" width="110" />
        <el-table-column prop="unit" label="单位" width="70" />
        <el-table-column prop="outQty" label="出库数量" width="90" align="right" />
        <el-table-column prop="inQty" label="入库数量" width="90" align="right" />
        <el-table-column label="差异处理状态" width="110" align="center">
          <template #default="{ row }">
            <el-tag :type="row.diffStatus === 'handled' ? 'success' : row.diffStatus === 'pending' ? 'warning' : 'info'">
              {{ row.diffStatus === 'handled' ? '已处理' : row.diffStatus === 'pending' ? '待处理' : '无差异' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="diffResult" label="差异处理结果" width="120" />
        <el-table-column prop="diffQty" label="差异数量" width="90" align="right" />
        <template #empty>暂无搜索结果</template>
      </el-table>
      <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" :page-sizes="[20,30,50]" layout="total, sizes, prev, pager, next" @size-change="() => load(1)" @current-change="load" class="pagination" />
    </el-card>

    <el-dialog v-model="handleVisible" title="差异处理" width="420px">
      <el-form label-width="90px">
        <el-form-item label="商品"><span>{{ currentRow?.productName }}</span></el-form-item>
        <el-form-item label="差异数量"><span>{{ currentRow?.diffQty }}</span></el-form-item>
        <el-form-item label="处理结果" required>
          <el-select v-model="handleResult" placeholder="请选择" style="width: 100%">
            <el-option label="补发" value="补发" />
            <el-option label="退回调出仓" value="退回调出仓" />
            <el-option label="报损" value="报损" />
            <el-option label="其他" value="其他" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="handleVisible = false">取消</el-button>
        <el-button type="primary" @click="submitHandle">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { Refresh } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import RemoteSelect from '@/components/RemoteSelect.vue'
import api from '@/api/client'

interface Row {
  itemId: number; outId: number; outBillNo: string; transferType: string
  fromWarehouseName: string; toWarehouseName: string; productName: string; specification: string; unit: string
  outQty: number; inQty: number; diffQty: number; diffStatus: string; diffResult: string; createdAt: string
}

const list = ref<Row[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(30)
const loading = ref(false)
const dateRange = ref<[string, string] | null>(null)
const search = reactive({
  keyword: '', transferType: '', productKw: '', fromWarehouseId: undefined as number | undefined,
  includeNoDiff: false, includeNotReceived: true,
})
const handleVisible = ref(false)
const handleResult = ref('')
const currentRow = ref<Row | null>(null)

function fmtTime(t: string) { return t ? t.replace('T', ' ').slice(0, 19) : '-' }

async function load(p = page.value) {
  page.value = p
  loading.value = true
  try {
    const res = await api.get('/api/v1/transfer-diffs', {
      params: { page: page.value, pageSize: pageSize.value, ...search, startDate: dateRange.value?.[0] || '', endDate: dateRange.value?.[1] || '' },
    })
    if (res.data.code === 0 || res.data.code === 200) {
      list.value = res.data.data?.list ?? []
      total.value = res.data.data?.total ?? 0
    }
  } finally { loading.value = false }
}

function openHandle(row: Row) {
  currentRow.value = row
  handleResult.value = ''
  handleVisible.value = true
}

async function submitHandle() {
  if (!handleResult.value) { ElMessage.warning('请选择处理结果'); return }
  const res = await api.put(`/api/v1/transfer-diffs/${currentRow.value!.itemId}/handle`, { result: handleResult.value })
  if (res.data.code === 0 || res.data.code === 200) {
    ElMessage.success('处理成功')
    handleVisible.value = false
    load()
  }
}

onMounted(() => load(1))
</script>

<style scoped>
.page { display: flex; flex-direction: column; gap: 12px; }
.page-header { display: flex; justify-content: space-between; align-items: center; }
.page-title { font-size: 18px; font-weight: 600; margin: 0; }
.check-row { margin-bottom: 12px; }
.pagination { margin-top: 16px; justify-content: flex-end; }
</style>
