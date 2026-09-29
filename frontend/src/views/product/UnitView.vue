<template>
  <div class="page">
    <div class="page-header">
      <span class="page-title">商品单位</span>
      <el-button type="primary" :icon="Plus" @click="openCreate">新增</el-button>
    </div>
    <el-card>
      <div class="filter-bar">
        <el-select
          v-model="storageType"
          placeholder="默认存放类型"
          clearable
          class="filter-select"
          @change="fetchList"
        >
          <el-option label="散货" :value="1" />
          <el-option label="整件" :value="2" />
        </el-select>
      </div>
      <el-table ref="tableRef" :data="list" v-loading="loading" row-key="id" border>
        <el-table-column label="排序" width="60" align="center">
          <template #default>
            <el-icon class="drag-handle" :class="{ disabled: isFiltered }"><Rank /></el-icon>
          </template>
        </el-table-column>
        <el-table-column type="selection" width="46" align="center" />
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
        <el-table-column prop="name" label="商品单位" min-width="180" />
        <el-table-column label="默认存放类型" min-width="140">
          <template #default="{ row }">{{ row.storageType === 2 ? '整件' : '散货' }}</template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑单位' : '新增单位'" width="480px" destroy-on-close>
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="100px">
        <el-form-item label="单位名称" prop="name">
          <el-input v-model="form.name" placeholder="单位名称" maxlength="4" show-word-limit />
        </el-form-item>
        <el-form-item label="默认存放类型" prop="storageType">
          <el-radio-group v-model="form.storageType">
            <el-radio :value="1">散货</el-radio>
            <el-radio :value="2">整件</el-radio>
          </el-radio-group>
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
import { More, Plus, Rank } from '@element-plus/icons-vue'
import Sortable from 'sortablejs'
import { createUnit, deleteUnit, fetchUnits, sortUnits, updateUnit, type GoodsUnit } from '@/api/unit'

const list = ref<GoodsUnit[]>([])
const loading = ref(false)
const tableRef = ref()
const storageType = ref<number | undefined>()
let sortable: Sortable | null = null

const isFiltered = computed(() => storageType.value !== undefined && storageType.value !== null)

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
    const params: { storageType?: number } = {}
    if (isFiltered.value) params.storageType = storageType.value
    const res = await fetchUnits(params)
    if (ok(res)) {
      list.value = res.data.data?.list || []
      await nextTick()
      initSortable()
    } else {
      ElMessage.error(res.data.message || '获取单位失败')
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
    const res = await sortUnits(items)
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
const form = reactive({ name: '', storageType: 1 })
const formRules = {
  name: [{ required: true, message: '请输入单位名称', trigger: 'blur' }],
  storageType: [{ required: true, message: '请选择默认存放类型', trigger: 'change' }],
}

function openCreate() {
  Object.assign(form, { name: '', storageType: 1 })
  isEdit.value = false
  currentId.value = null
  dialogVisible.value = true
}

function openEdit(row: GoodsUnit) {
  Object.assign(form, { name: row.name, storageType: row.storageType || 1 })
  isEdit.value = true
  currentId.value = row.id
  dialogVisible.value = true
}

function handleCommand(cmd: string, row: GoodsUnit) {
  if (cmd === 'edit') openEdit(row)
  else if (cmd === 'delete') handleDelete(row)
}

async function handleSubmit() {
  await formRef.value?.validate()
  submitting.value = true
  try {
    const payload = { name: form.name, sort: 0, status: 1, storageType: form.storageType }
    const res = isEdit.value && currentId.value !== null
      ? await updateUnit(currentId.value, payload)
      : await createUnit(payload)
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

async function handleDelete(row: GoodsUnit) {
  try {
    await ElMessageBox.confirm(`确认删除单位“${row.name}”？`, '提示', { type: 'warning' })
  } catch {
    return
  }
  const res = await deleteUnit(row.id)
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
.filter-bar { margin-bottom: 16px; }
.filter-select { width: 160px; }
.drag-handle { cursor: move; color: #909399; }
.drag-handle.disabled { cursor: not-allowed; opacity: 0.4; }
.danger-text { color: var(--el-color-danger); }
</style>
