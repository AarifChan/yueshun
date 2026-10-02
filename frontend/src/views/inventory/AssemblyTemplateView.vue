<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">拆装模板</h2>
      <el-button type="primary" :icon="Plus" @click="openCreate">新增拆装模板</el-button>
    </div>
    <el-card>
      <el-form inline class="search-form">
        <el-form-item label="模板名称"><el-input v-model="search.keyword" placeholder="模板名称" clearable /></el-form-item>
        <el-form-item>
          <el-button type="primary" @click="load(1)">查询</el-button>
          <el-button @click="reset">清空</el-button>
        </el-form-item>
      </el-form>
      <el-table :data="list" v-loading="loading" border stripe>
        <el-table-column label="操作" width="130" fixed="left">
          <template #default="{ row }">
            <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
            <el-button link type="danger" @click="remove(row)">删除</el-button>
          </template>
        </el-table-column>
        <el-table-column prop="name" label="模板名称" min-width="160" />
        <el-table-column label="出库商品" min-width="200">
          <template #default="{ row }">{{ summarize(row.outItems) }}</template>
        </el-table-column>
        <el-table-column label="入库商品" min-width="200">
          <template #default="{ row }">{{ summarize(row.inItems) }}</template>
        </el-table-column>
        <el-table-column prop="remark" label="备注" min-width="140" show-overflow-tooltip />
        <el-table-column label="状态" width="90" align="center">
          <template #default="{ row }"><el-tag :type="row.status === 1 ? 'success' : 'danger'">{{ row.status === 1 ? '启用' : '停用' }}</el-tag></template>
        </el-table-column>
        <template #empty>暂无搜索结果</template>
      </el-table>
      <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" :page-sizes="[20,30,50]" layout="total, sizes, prev, pager, next" @size-change="() => load(1)" @current-change="load" class="pagination" />
    </el-card>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑拆装模板' : '新增拆装模板'" width="760px">
      <el-form :model="form" label-width="90px">
        <el-form-item label="模板名称" required><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="备注"><el-input v-model="form.remark" /></el-form-item>
        <el-form-item label="状态"><el-switch v-model="form.status" :active-value="1" :inactive-value="0" /></el-form-item>
        <el-form-item label="出库商品">
          <ItemEditor v-model="form.outItemList" />
        </el-form-item>
        <el-form-item label="入库商品">
          <ItemEditor v-model="form.inItemList" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { Plus } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '@/api/client'
import ItemEditor from './AssemblyItemEditor.vue'

interface ItemEntry { productId?: number; productName?: string; quantity: number }
interface Template { id: number; name: string; outItems: string; inItems: string; remark: string; status: number }

const list = ref<Template[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(30)
const loading = ref(false)
const dialogVisible = ref(false)
const isEdit = ref(false)
const search = reactive({ keyword: '' })

const form = reactive({
  id: 0, name: '', remark: '', status: 1,
  outItemList: [] as ItemEntry[], inItemList: [] as ItemEntry[],
})

function parseItems(json: string): ItemEntry[] {
  try { return JSON.parse(json || '[]') } catch { return [] }
}

function summarize(json: string): string {
  const arr = parseItems(json)
  if (!arr.length) return '-'
  return arr.map((i) => `${i.productName || i.productId}×${i.quantity}`).join('、')
}

async function load(p = page.value) {
  page.value = p
  loading.value = true
  try {
    const res = await api.get('/api/v1/assembly-templates', { params: { page: page.value, pageSize: pageSize.value, ...search } })
    if (res.data.code === 0 || res.data.code === 200) {
      list.value = res.data.data?.list ?? []
      total.value = res.data.data?.total ?? 0
    }
  } finally { loading.value = false }
}

function reset() { search.keyword = ''; load(1) }

function openCreate() {
  Object.assign(form, { id: 0, name: '', remark: '', status: 1, outItemList: [], inItemList: [] })
  isEdit.value = false
  dialogVisible.value = true
}

function openEdit(row: Template) {
  Object.assign(form, {
    id: row.id, name: row.name, remark: row.remark, status: row.status,
    outItemList: parseItems(row.outItems), inItemList: parseItems(row.inItems),
  })
  isEdit.value = true
  dialogVisible.value = true
}

async function submit() {
  if (!form.name) { ElMessage.warning('请输入模板名称'); return }
  const payload = {
    name: form.name, remark: form.remark, status: form.status,
    outItems: JSON.stringify(form.outItemList.filter((i) => i.productId)),
    inItems: JSON.stringify(form.inItemList.filter((i) => i.productId)),
  }
  const res = isEdit.value
    ? await api.put(`/api/v1/assembly-templates/${form.id}`, payload)
    : await api.post('/api/v1/assembly-templates', payload)
  if (res.data.code === 0 || res.data.code === 200) {
    ElMessage.success('保存成功')
    dialogVisible.value = false
    load()
  }
}

async function remove(row: Template) {
  await ElMessageBox.confirm(`确认删除模板 ${row.name}？`, '提示', { type: 'warning' })
  const res = await api.delete(`/api/v1/assembly-templates/${row.id}`)
  if (res.data.code === 0 || res.data.code === 200) {
    ElMessage.success('删除成功')
    load()
  }
}

onMounted(() => load(1))
</script>

<style scoped>
.page { display: flex; flex-direction: column; gap: 12px; }
.page-header { display: flex; justify-content: space-between; align-items: center; }
.page-title { font-size: 18px; font-weight: 600; margin: 0; }
.pagination { margin-top: 16px; justify-content: flex-end; }
</style>
