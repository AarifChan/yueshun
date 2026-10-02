<template>
  <div>
    <div class="toolbar">
      <el-button type="primary" icon="Plus" @click="openForm()">新增专题分类</el-button>
      <span class="hint">专题分类可绑定商品并显示到小程序商城，支持设置可见客户范围。</span>
    </div>

    <el-table :data="list" v-loading="loading" border stripe>
      <el-table-column label="操作" width="130" fixed="left">
        <template #default="{ row }">
          <el-button link type="primary" @click="openForm(row)">编辑</el-button>
          <el-button link type="danger" @click="remove(row)">删除</el-button>
        </template>
      </el-table-column>
      <el-table-column label="分类图片" width="90" align="center">
        <template #default="{ row }">
          <el-image v-if="row.image" :src="row.image" style="width: 40px; height: 40px" fit="cover" />
          <span v-else>-</span>
        </template>
      </el-table-column>
      <el-table-column prop="name" label="专题分类" min-width="150" />
      <el-table-column label="可见客户范围" width="120" align="center">
        <template #default="{ row }">{{ row.visibleScope === 'part' ? '部分客户' : '全部客户' }}</template>
      </el-table-column>
      <el-table-column label="显示到小程序商城" width="130" align="center">
        <template #default="{ row }">
          <el-tag :type="row.showInMall ? 'success' : 'info'" size="small">{{ row.showInMall ? '显示' : '不显示' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="绑定商品" width="100" align="center">
        <template #default="{ row }">{{ row.productCount }} 个</template>
      </el-table-column>
      <el-table-column prop="sort" label="排序" width="80" align="center" />
      <template #empty>暂无数据</template>
    </el-table>

    <el-dialog v-model="formVisible" :title="form.id ? '编辑专题分类' : '新增专题分类'" width="560px">
      <el-form label-width="130px">
        <el-form-item label="专题分类名称" required>
          <el-input v-model="form.name" placeholder="如 新品专区" style="width: 320px" />
        </el-form-item>
        <el-form-item label="分类图片">
          <el-input v-model="form.image" placeholder="图片 URL" style="width: 320px" />
        </el-form-item>
        <el-form-item label="可见客户范围">
          <el-radio-group v-model="form.visibleScope">
            <el-radio value="all">全部客户</el-radio>
            <el-radio value="part">部分客户</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="显示到小程序商城">
          <el-switch v-model="form.showInMall" />
        </el-form-item>
        <el-form-item label="绑定商品">
          <el-select
            v-model="formProductIds"
            multiple
            filterable
            remote
            :remote-method="searchProducts"
            placeholder="搜索商品并多选"
            style="width: 100%"
          >
            <el-option v-for="p in productOptions" :key="p.id" :label="`${p.name}（${p.code || '-'}）`" :value="p.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="form.sort" :min="0" :max="9999" />
        </el-form-item>
        <el-form-item label="启用">
          <el-switch v-model="formEnabled" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="save">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '@/api/client'

interface Row {
  id: number; name: string; image: string; visibleScope: string
  showInMall: boolean; productIds: string; productCount: number; sort: number; status: number
}
interface Option { id: number; name: string; code?: string }

const list = ref<Row[]>([])
const loading = ref(false)
const formVisible = ref(false)
const saving = ref(false)
const form = ref({ id: 0, name: '', image: '', visibleScope: 'all', showInMall: true, sort: 0 })
const formEnabled = ref(true)
const formProductIds = ref<number[]>([])
const productOptions = ref<Option[]>([])

async function load() {
  loading.value = true
  try {
    const res = await api.get('/api/v1/subject-categories')
    if (res.data.code === 0 || res.data.code === 200) list.value = res.data.data?.list ?? []
  } finally {
    loading.value = false
  }
}

async function searchProducts(kw: string) {
  const res = await api.get('/api/v1/products', { params: { page: 1, pageSize: 20, keyword: kw } })
  if (res.data.code === 0 || res.data.code === 200) productOptions.value = res.data.data?.list ?? []
}

function openForm(row?: Row) {
  if (row) {
    form.value = { id: row.id, name: row.name, image: row.image, visibleScope: row.visibleScope, showInMall: row.showInMall, sort: row.sort }
    formEnabled.value = row.status === 1
    formProductIds.value = row.productIds ? row.productIds.split(',').filter(Boolean).map(Number) : []
  } else {
    form.value = { id: 0, name: '', image: '', visibleScope: 'all', showInMall: true, sort: 0 }
    formEnabled.value = true
    formProductIds.value = []
  }
  productOptions.value = []
  formVisible.value = true
}

async function save() {
  if (!form.value.name.trim()) {
    ElMessage.warning('请填写专题分类名称')
    return
  }
  saving.value = true
  try {
    const payload = {
      ...form.value,
      status: formEnabled.value ? 1 : 0,
      productIds: formProductIds.value.join(','),
    }
    const res = form.value.id
      ? await api.put(`/api/v1/subject-categories/${form.value.id}`, payload)
      : await api.post('/api/v1/subject-categories', payload)
    if (res.data.code === 0 || res.data.code === 200) {
      ElMessage.success('已保存')
      formVisible.value = false
      load()
    } else {
      ElMessage.error(res.data.message || '保存失败')
    }
  } finally {
    saving.value = false
  }
}

async function remove(row: Row) {
  await ElMessageBox.confirm(`确定删除专题分类「${row.name}」？`, '提示', { type: 'warning' })
  const res = await api.delete(`/api/v1/subject-categories/${row.id}`)
  if (res.data.code === 0 || res.data.code === 200) {
    ElMessage.success('已删除')
    load()
  }
}

onMounted(load)
</script>

<style scoped>
.toolbar { display: flex; align-items: center; gap: 12px; margin-bottom: 12px; }
.hint { color: #909399; font-size: 12px; }
</style>
