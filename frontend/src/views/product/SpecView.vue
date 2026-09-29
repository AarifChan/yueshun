<template>
  <div class="page">
    <div class="page-header">
      <span class="page-title">商品规格</span>
      <el-button type="primary" :icon="Plus" @click="openCreate">新增</el-button>
    </div>
    <el-card>
      <div class="filter-bar">
        <el-input
          v-model="keyword"
          placeholder="请输入规格组"
          clearable
          class="filter-input"
          @keyup.enter="fetchList"
          @clear="fetchList"
        />
        <el-button type="primary" @click="fetchList">搜 索</el-button>
      </div>
      <el-table ref="tableRef" :data="list" v-loading="loading" row-key="id" border>
        <el-table-column label="排序" width="60" align="center">
          <template #default>
            <el-icon class="drag-handle" :class="{ disabled: isFiltered }"><Rank /></el-icon>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="70" align="center">
          <template #default="{ row }">
            <el-dropdown trigger="click" @command="(cmd: string) => handleCommand(cmd, row)">
              <el-button text :icon="More" />
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="edit">编辑</el-dropdown-item>
                  <el-dropdown-item command="delete"><span class="danger-text">删除</span></el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </template>
        </el-table-column>
        <el-table-column prop="name" label="规格组" min-width="160" />
        <el-table-column prop="values" label="规格值" min-width="260" show-overflow-tooltip />
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑规格' : '新增规格'" width="520px" destroy-on-close>
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="80px">
        <el-form-item label="规格组" prop="name">
          <el-input v-model="form.name" placeholder="请输入规格组" maxlength="10" show-word-limit />
        </el-form-item>
        <el-form-item label="规格值">
          <div class="value-list">
            <div v-for="(_, index) in form.values" :key="index" class="value-row">
              <el-input v-model="form.values[index]" placeholder="请输入" maxlength="50" show-word-limit />
              <el-icon v-if="form.values.length > 1" class="value-remove" @click="removeValue(index)"><Delete /></el-icon>
            </div>
            <el-button class="value-add" @click="addValue">＋ 添加规格值</el-button>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance } from 'element-plus'
import { Delete, More, Plus, Rank } from '@element-plus/icons-vue'
import Sortable from 'sortablejs'
import { createSpec, deleteSpec, fetchSpecs, sortSpecs, updateSpec, type ProductSpec } from '@/api/spec'

const list = ref<ProductSpec[]>([])
const loading = ref(false)
const tableRef = ref()
const keyword = ref('')
let sortable: Sortable | null = null

const isFiltered = computed(() => keyword.value.trim() !== '')

function ok(res: any) {
  return res.data.code === 0 || res.data.code === 200
}

function destroySortable() {
  sortable?.destroy()
  sortable = null
}

function initSortable() {
  destroySortable()
  if (isFiltered.value) return
  const tbody = tableRef.value?.$el?.querySelector('.el-table__body tbody')
  if (!tbody) return
  sortable = Sortable.create(tbody, {
    handle: '.drag-handle',
    animation: 150,
    onEnd: onDragEnd,
  })
}

async function fetchList() {
  loading.value = true
  try {
    const params: { keyword?: string } = {}
    if (isFiltered.value) params.keyword = keyword.value.trim()
    const res = await fetchSpecs(params)
    if (ok(res)) {
      list.value = res.data.data?.list || []
      await nextTick()
      initSortable()
    } else {
      ElMessage.error(res.data.message || '获取规格失败')
    }
  } finally {
    loading.value = false
  }
}

async function onDragEnd(evt: Sortable.SortableEvent) {
  const { oldIndex, newIndex } = evt
  if (oldIndex === undefined || newIndex === undefined || oldIndex === newIndex) return
  const rows = [...list.value]
  const [moved] = rows.splice(oldIndex, 1)
  if (!moved) {
    fetchList()
    return
  }
  rows.splice(newIndex, 0, moved)
  const items = rows.map((r, i) => ({ id: r.id, sort: i }))
  try {
    const res = await sortSpecs(items)
    if (ok(res)) {
      ElMessage.success('排序成功')
    } else {
      ElMessage.error(res.data.message || '排序失败')
    }
  } finally {
    fetchList()
  }
}

const dialogVisible = ref(false)
const isEdit = ref(false)
const currentId = ref<number | null>(null)
const submitting = ref(false)
const formRef = ref<FormInstance>()
const form = reactive({ name: '', values: [''] as string[] })
const formRules = {
  name: [{ required: true, message: '请输入规格组', trigger: 'blur' }],
}

function addValue() {
  form.values.push('')
}

function removeValue(index: number) {
  if (form.values.length <= 1) return
  form.values.splice(index, 1)
}

function openCreate() {
  form.name = ''
  form.values = ['']
  isEdit.value = false
  currentId.value = null
  dialogVisible.value = true
}

function openEdit(row: ProductSpec) {
  form.name = row.name
  const values = (row.values || '').split(',').map((v) => v.trim()).filter((v) => v !== '')
  form.values = values.length > 0 ? values : ['']
  isEdit.value = true
  currentId.value = row.id
  dialogVisible.value = true
}

function handleCommand(cmd: string, row: ProductSpec) {
  if (cmd === 'edit') openEdit(row)
  else if (cmd === 'delete') handleDelete(row)
}

async function handleSubmit() {
  await formRef.value?.validate()
  submitting.value = true
  try {
    const values = form.values.map((v) => v.trim()).filter((v) => v !== '').join(',')
    const payload = { name: form.name, values, sort: 0, status: 1 }
    const res = isEdit.value && currentId.value !== null
      ? await updateSpec(currentId.value, payload)
      : await createSpec(payload)
    if (ok(res)) {
      ElMessage.success(isEdit.value ? '编辑成功' : '新增成功')
      dialogVisible.value = false
      fetchList()
    } else {
      ElMessage.error(res.data.message || '操作失败')
    }
  } finally {
    submitting.value = false
  }
}

async function handleDelete(row: ProductSpec) {
  try {
    await ElMessageBox.confirm(`确认删除规格“${row.name}”？`, '提示', { type: 'warning' })
  } catch {
    return
  }
  const res = await deleteSpec(row.id)
  if (ok(res)) {
    ElMessage.success('删除成功')
    fetchList()
  } else {
    ElMessage.error(res.data.message || '删除失败')
  }
}

onMounted(fetchList)
onBeforeUnmount(destroySortable)
</script>

<style scoped>
.page { padding: 20px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.page-title { font-size: 18px; font-weight: 600; }
.filter-bar { display: flex; gap: 12px; margin-bottom: 16px; }
.filter-input { width: 240px; }
.drag-handle { cursor: move; color: #909399; }
.drag-handle.disabled { cursor: not-allowed; opacity: 0.4; }
.danger-text { color: var(--el-color-danger); }
.value-list { width: 100%; display: flex; flex-direction: column; gap: 10px; }
.value-row { display: flex; align-items: center; gap: 8px; }
.value-remove { cursor: pointer; color: #909399; }
.value-remove:hover { color: var(--el-color-danger); }
.value-add { width: 100%; border-style: dashed; }
</style>
