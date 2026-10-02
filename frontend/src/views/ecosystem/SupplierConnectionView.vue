<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">供应商连接</h2>
      <div class="actions">
        <el-button @click="router.push('/ecosystem/guide')">生态互联使用介绍</el-button>
        <el-button type="primary" @click="openDialog()">新增连接</el-button>
      </div>
    </div>
    <el-card>
      <el-form inline @submit.prevent>
        <el-form-item label="搜索">
          <el-input v-model="filters.keyword" placeholder="企业名称/企业编号" clearable style="width: 200px" />
        </el-form-item>
        <el-form-item label="连接状态">
          <el-select v-model="filters.status" style="width: 130px">
            <el-option label="全部" value="" />
            <el-option label="待连接" value="pending" />
            <el-option label="已连接" value="connected" />
            <el-option label="已断开" value="disconnected" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="search">搜索</el-button>
          <el-button @click="reset">重置</el-button>
        </el-form-item>
      </el-form>
      <el-table :data="list" v-loading="loading" border stripe>
        <el-table-column type="index" label="序" width="60" />
        <el-table-column prop="corpName" label="上游企业微信名称" min-width="180" show-overflow-tooltip />
        <el-table-column prop="corpCode" label="企业编号" width="140" />
        <el-table-column label="关联供应商" min-width="150">
          <template #default="{ row }">{{ row.supplier || '-' }}</template>
        </el-table-column>
        <el-table-column label="连接状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 'connected' ? 'success' : row.status === 'pending' ? 'warning' : 'info'">
              {{ row.status === 'connected' ? '已连接' : row.status === 'pending' ? '待连接' : '已断开' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button v-if="row.status !== 'connected'" link type="success" @click="openConnect(row)">连接</el-button>
            <el-button v-else link type="warning" @click="disconnect(row)">断开</el-button>
            <el-button link type="primary" @click="openDialog(row)">编辑</el-button>
            <el-button link type="danger" @click="remove(row)">删除</el-button>
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

    <!-- 新增/编辑连接 -->
    <el-dialog v-model="dialogVisible" :title="form.id ? '编辑连接' : '新增连接'" width="480px">
      <el-form :model="form" label-width="110px">
        <el-form-item label="上游企业名称" required><el-input v-model="form.corpName" placeholder="上游企业微信名称" /></el-form-item>
        <el-form-item label="企业编号"><el-input v-model="form.corpCode" /></el-form-item>
        <el-form-item label="备注"><el-input v-model="form.remark" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="save">保存</el-button>
      </template>
    </el-dialog>

    <!-- 连接（选择关联供应商） -->
    <el-dialog v-model="connectVisible" :title="`连接：${connectRow?.corpName ?? ''}`" width="440px">
      <el-form label-width="90px">
        <el-form-item label="关联供应商" required>
          <el-select v-model="connectSupplierId" filterable style="width: 100%">
            <el-option v-for="s in suppliers" :key="s.id" :label="s.name" :value="s.id" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="connectVisible = false">取消</el-button>
        <el-button type="primary" @click="doConnect">确认连接</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '@/api/client'
import { useReport } from '@/composables/useReport'

interface Row {
  id: number; corpName: string; corpCode: string; supplierId: number; supplier: string
  status: string; remark: string; createdAt: string
}
interface Option { id: number; name: string }

const router = useRouter()
const { list, total, loading, page, pageSize, filters, load, search, reset } =
  useReport<Row>('/api/v1/supplier-connections', { keyword: '', status: '' })

const dialogVisible = ref(false)
const form = ref({ id: 0, corpName: '', corpCode: '', supplierId: 0, remark: '' })

const connectVisible = ref(false)
const connectRow = ref<Row | null>(null)
const connectSupplierId = ref<number>()
const suppliers = ref<Option[]>([])

async function loadSuppliers() {
  const res = await api.get('/api/v1/suppliers', { params: { page: 1, pageSize: 500 } })
  if (res.data.code === 0 || res.data.code === 200) suppliers.value = res.data.data?.list ?? []
}

function openDialog(row?: Row) {
  form.value = row
    ? { id: row.id, corpName: row.corpName, corpCode: row.corpCode, supplierId: row.supplierId, remark: row.remark }
    : { id: 0, corpName: '', corpCode: '', supplierId: 0, remark: '' }
  dialogVisible.value = true
}

async function save() {
  if (!form.value.corpName) { ElMessage.warning('请填写上游企业名称'); return }
  const res = form.value.id
    ? await api.put(`/api/v1/supplier-connections/${form.value.id}`, form.value)
    : await api.post('/api/v1/supplier-connections', form.value)
  if (res.data.code === 0 || res.data.code === 200) {
    ElMessage.success('保存成功')
    dialogVisible.value = false
    load()
  } else {
    ElMessage.error(res.data.message || '保存失败')
  }
}

function openConnect(row: Row) {
  connectRow.value = row
  connectSupplierId.value = row.supplierId || undefined
  connectVisible.value = true
}

async function doConnect() {
  if (!connectSupplierId.value) { ElMessage.warning('请选择关联供应商'); return }
  const res = await api.put(`/api/v1/supplier-connections/${connectRow.value?.id}/connect`, { supplierId: connectSupplierId.value })
  if (res.data.code === 0 || res.data.code === 200) {
    ElMessage.success('已连接')
    connectVisible.value = false
    load()
  } else {
    ElMessage.error(res.data.message || '连接失败')
  }
}

async function disconnect(row: Row) {
  await ElMessageBox.confirm(`确定断开与「${row.corpName}」的连接？`, '提示', { type: 'warning' })
  const res = await api.put(`/api/v1/supplier-connections/${row.id}/disconnect`)
  if (res.data.code === 0 || res.data.code === 200) {
    ElMessage.success('已断开')
    load()
  }
}

async function remove(row: Row) {
  await ElMessageBox.confirm(`确定删除连接「${row.corpName}」？`, '提示', { type: 'warning' })
  const res = await api.delete(`/api/v1/supplier-connections/${row.id}`)
  if (res.data.code === 0 || res.data.code === 200) {
    ElMessage.success('已删除')
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
.actions { display: flex; gap: 8px; }
</style>
