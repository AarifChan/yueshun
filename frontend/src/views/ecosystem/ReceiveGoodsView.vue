<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">商品接收</h2>
    </div>
    <el-card>
      <el-tabs v-model="tab" @tab-change="onTab">
        <el-tab-pane label="自动接收" name="pending" />
        <el-tab-pane label="确认接收" name="received" />
        <el-tab-pane label="已关闭" name="closed" />
      </el-tabs>
      <el-form inline @submit.prevent>
        <el-form-item label="供应商">
          <el-select v-model="supplierId" filterable clearable style="width: 180px">
            <el-option v-for="s in suppliers" :key="s.id" :label="s.name" :value="s.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="商品品牌">
          <el-input v-model="brand" clearable style="width: 140px" />
        </el-form-item>
        <el-form-item label="搜索">
          <el-input v-model="keyword" placeholder="名称/编号/条码" clearable style="width: 170px" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="load(1)">搜索</el-button>
        </el-form-item>
      </el-form>
      <el-table :data="list" v-loading="loading" border stripe>
        <el-table-column label="图片" width="70" align="center">
          <template #default="{ row }">
            <el-image v-if="row.image" :src="row.image" style="width: 36px; height: 36px" fit="cover" />
            <span v-else class="no-img">无图</span>
          </template>
        </el-table-column>
        <el-table-column prop="code" label="商品编号" width="110" />
        <el-table-column prop="barcode" label="商品条码" width="130" />
        <el-table-column prop="name" label="商品名称" min-width="160" show-overflow-tooltip />
        <el-table-column prop="spec" label="规格" width="110" />
        <el-table-column prop="brand" label="商品品牌" width="110" />
        <el-table-column prop="supplier" label="供应商" min-width="130">
          <template #default="{ row }">{{ row.supplier || '-' }}</template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <template v-if="row.status === 'pending'">
              <el-button link type="success" @click="receive(row)">接收</el-button>
              <el-button link type="warning" @click="close(row)">关闭</el-button>
            </template>
            <el-button v-else-if="row.status === 'closed'" link type="primary" @click="reopen(row)">重新打开</el-button>
            <el-tag v-else type="success" size="small">已接收</el-tag>
          </template>
        </el-table-column>
        <template #empty>暂无搜索结果</template>
      </el-table>
      <el-pagination
        v-model:current-page="page" v-model:page-size="pageSize"
        :total="total" :page-sizes="[30, 50, 100]"
        layout="total, sizes, prev, pager, next" style="margin-top: 12px; justify-content: flex-end"
        @current-change="load()" @size-change="load(1)" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '@/api/client'
import { useReport } from '@/composables/useReport'

interface Row {
  id: number; image: string; code: string; barcode: string; name: string; spec: string
  brand: string; supplierId: number; supplier: string; status: string; productId: number
}
interface Option { id: number; name: string }

const tab = ref('pending')
const { list, total, loading, page, pageSize, filters, load } =
  useReport<Row>('/api/v1/received-goods', { status: 'pending', supplierId: '', brand: '', keyword: '' })

const suppliers = ref<Option[]>([])

const supplierId = computed({
  get: () => (filters.value.supplierId ? Number(filters.value.supplierId) : undefined),
  set: (v) => { filters.value.supplierId = v ?? '' },
})
const brand = computed({
  get: () => filters.value.brand,
  set: (v) => { filters.value.brand = v },
})
const keyword = computed({
  get: () => filters.value.keyword,
  set: (v) => { filters.value.keyword = v },
})

function onTab() {
  filters.value.status = tab.value
  load(1)
}

async function loadSuppliers() {
  const res = await api.get('/api/v1/suppliers', { params: { page: 1, pageSize: 500 } })
  if (res.data.code === 0 || res.data.code === 200) suppliers.value = res.data.data?.list ?? []
}

async function receive(row: Row) {
  await ElMessageBox.confirm(`确认接收商品「${row.name}」？接收后将自动创建商品档案。`, '确认接收', { type: 'warning' })
  const res = await api.put(`/api/v1/received-goods/${row.id}/receive`)
  if (res.data.code === 0 || res.data.code === 200) {
    ElMessage.success('已接收并创建商品档案')
    load()
  } else {
    ElMessage.error(res.data.message || '接收失败')
  }
}

async function close(row: Row) {
  await ElMessageBox.confirm(`确定关闭商品「${row.name}」？`, '提示', { type: 'warning' })
  const res = await api.put(`/api/v1/received-goods/${row.id}/close`)
  if (res.data.code === 0 || res.data.code === 200) {
    ElMessage.success('已关闭')
    load()
  }
}

async function reopen(row: Row) {
  const res = await api.put(`/api/v1/received-goods/${row.id}/reopen`)
  if (res.data.code === 0 || res.data.code === 200) {
    ElMessage.success('已重新打开')
    load()
  }
}

onMounted(() => {
  load(1)
  loadSuppliers()
})
</script>

<style scoped>
.page { display: flex; flex-direction: column; gap: 12px; }
.page-header { display: flex; justify-content: space-between; align-items: center; }
.page-title { font-size: 18px; font-weight: 600; margin: 0; }
.no-img { color: #c0c4cc; font-size: 12px; }
</style>
