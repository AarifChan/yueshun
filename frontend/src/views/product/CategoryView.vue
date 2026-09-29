<template>
  <div class="page">
    <div class="page-header">
      <span class="page-title">商品分类</span>
      <div>
        <el-button @click="handleFillImages">快速补全分类图片</el-button>
        <el-button type="primary" :icon="Plus" @click="openCreate()">新增</el-button>
      </div>
    </div>
    <el-alert type="info" :closable="false" show-icon title="商品分类为商品的基本属性，每个商品只能对应一个商品分类" class="tip" />
    <el-card>
      <el-table
        :key="tableKey"
        ref="tableRef"
        :data="tree"
        v-loading="loading"
        row-key="id"
        lazy
        :load="loadChildren"
        :tree-props="{ children: 'children', hasChildren: 'hasChildren' }"
        border
        @expand-change="onExpandChange"
      >
        <el-table-column label="排序" width="60" align="center">
          <template #default>
            <el-icon class="drag-handle"><Rank /></el-icon>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="70" align="center">
          <template #default="{ row }">
            <el-dropdown trigger="click" @command="(cmd: string) => handleCommand(cmd, row)">
              <el-button text :icon="More" />
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="edit">编辑</el-dropdown-item>
                  <el-dropdown-item command="child">新增子分类</el-dropdown-item>
                  <el-dropdown-item command="delete"><span class="danger-text">删除</span></el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </template>
        </el-table-column>
        <el-table-column label="分类图片" width="100" align="center">
          <template #default="{ row }">
            <el-image
              v-if="row.image"
              :src="row.image"
              :preview-src-list="[row.image]"
              preview-teleported
              fit="cover"
              class="category-image"
            />
            <div v-else class="image-placeholder"><el-icon><Picture /></el-icon></div>
          </template>
        </el-table-column>
        <el-table-column label="商品分类" min-width="220">
          <template #default="{ row }">
            <div class="category-name" :style="{ paddingLeft: (levelMap.get(row.id) || 0) * 18 + 'px' }">
              <el-icon
                v-if="row.hasChildren"
                class="expand-icon"
                :class="{ expanded: isExpanded(row.id) }"
                @click.stop="toggleRow(row)"
              ><CaretRight /></el-icon>
              <span v-else class="expand-spacer"></span>
              <el-icon class="folder-icon"><Folder /></el-icon>
              <span>{{ row.name }}</span>
              <el-dropdown trigger="click" @command="(cmd: string) => handleAddCommand(cmd, row)">
                <el-icon class="row-add-trigger" @click.stop><CaretBottom /></el-icon>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item command="before">向前添加同级</el-dropdown-item>
                    <el-dropdown-item command="after">向后添加同级</el-dropdown-item>
                    <el-dropdown-item command="child">添加子分类</el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="显示到小程序商城" width="150" align="right">
          <template #default="{ row }">
            <el-switch
              :model-value="row.status"
              :active-value="1"
              :inactive-value="0"
              @change="(val: number) => onStatusChange(row, val)"
            />
          </template>
        </el-table-column>
        <el-table-column label="绑定商品" width="100" align="right">
          <template #default="{ row }">
            <el-link type="primary" @click="goProducts(row)">{{ row.productCount ?? 0 }}</el-link>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑分类' : '新增分类'" width="520px" destroy-on-close>
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="90px">
        <el-form-item label="父分类">
          <el-tree-select
            v-model="form.parentId"
            :data="parentOptions"
            :props="{ label: 'name', children: 'children' }"
            node-key="id"
            check-strictly
            class="full-width"
          />
        </el-form-item>
        <el-form-item label="名称" prop="name"><el-input v-model="form.name" placeholder="分类名称" /></el-form-item>
        <el-form-item label="编码"><el-input v-model="form.code" placeholder="分类编码" /></el-form-item>
        <el-form-item label="排序"><el-input-number v-model="form.sort" :min="0" /></el-form-item>
        <el-form-item label="状态"><el-switch v-model="form.status" :active-value="1" :inactive-value="0" /></el-form-item>
        <el-form-item label="分类图片">
          <el-upload
            v-model:file-list="fileList"
            list-type="picture-card"
            :limit="1"
            accept="image/*"
            :http-request="uploadImage"
            :on-remove="onImageRemove"
            :on-exceed="() => ElMessage.warning('仅支持上传一张图片')"
          >
            <el-icon><Plus /></el-icon>
          </el-upload>
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
import { computed, markRaw, nextTick, onBeforeUnmount, onMounted, reactive, ref, shallowRef } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox, type FormInstance, type UploadFile, type UploadRequestOptions } from 'element-plus'
import { CaretBottom, CaretRight, Folder, More, Picture, Plus, Rank } from '@element-plus/icons-vue'
import Sortable from 'sortablejs'
import {
  createCategory,
  deleteCategory,
  fetchCategoryTree,
  fillCategoryImages,
  sortCategories,
  updateCategory,
  updateCategoryStatus,
  uploadFile,
  type CategoryNode,
} from '@/api/category'

const router = useRouter()
const tree = shallowRef<CategoryNode[]>([])
const loading = ref(false)
const tableRef = ref()
const tableKey = ref(0)
const expandedKeys = new Set<number>()
const expandVersion = ref(0)
let levelMap = new Map<number, number>()
let parentMap = new Map<number, number>()
let sortable: Sortable | null = null

function ok(res: any) {
  return res.data.code === 0 || res.data.code === 200
}

function isExpanded(id: number) {
  expandVersion.value
  return expandedKeys.has(id)
}

function normalize(nodes: CategoryNode[], level: number, parentId: number) {
  for (const node of nodes) {
    levelMap.set(node.id, level)
    parentMap.set(node.id, parentId)
    if (node.children && node.children.length) {
      normalize(node.children, level + 1, node.id)
    } else {
      node.children = []
    }
  }
}

function expandAncestors(parentId: number) {
  let current = parentId
  while (current) {
    expandedKeys.add(current)
    current = parentMap.get(current) ?? 0
  }
}

async function loadNodeChildren(row: CategoryNode): Promise<CategoryNode[]> {
  try {
    const res = await fetchCategoryTree(row.id)
    if (ok(res)) {
      const children: CategoryNode[] = res.data.data || []
      normalize(children, (levelMap.get(row.id) ?? 0) + 1, row.id)
      row.children = children
      return children
    }
    ElMessage.error(res.data.message || '获取子分类失败')
  } catch {
  }
  row.children = []
  return []
}

async function loadChildren(row: CategoryNode, _treeNode: unknown, resolve: (data: CategoryNode[]) => void) {
  if (row.children?.length) {
    resolve(row.children)
    return
  }
  resolve(await loadNodeChildren(row))
}

async function fetchTree() {
  loading.value = true
  try {
    const res = await fetchCategoryTree(0)
    if (ok(res)) {
      const data: CategoryNode[] = res.data.data || []
      levelMap = new Map()
      parentMap = new Map()
      normalize(data, 0, 0)
      tree.value = markRaw(data)
      sortable?.destroy()
      sortable = null
      tableKey.value++
      await nextTick()
      await restoreExpanded(data)
      for (const id of [...expandedKeys]) {
        if (!levelMap.has(id)) expandedKeys.delete(id)
      }
      expandVersion.value++
      initSortable()
    } else {
      ElMessage.error(res.data.message || '获取分类失败')
    }
  } finally {
    loading.value = false
  }
}

async function restoreExpanded(nodes: CategoryNode[]) {
  for (const node of nodes) {
    if (expandedKeys.has(node.id) && node.hasChildren) {
      const children = await loadNodeChildren(node)
      tableRef.value?.toggleRowExpansion(node, true)
      if (children.length) await restoreExpanded(children)
    }
  }
}

function visibleRows(): CategoryNode[] {
  const rows: CategoryNode[] = []
  const walk = (nodes: CategoryNode[]) => {
    for (const node of nodes) {
      rows.push(node)
      if (expandedKeys.has(node.id) && node.children?.length) walk(node.children)
    }
  }
  walk(tree.value)
  return rows
}

function initSortable() {
  if (sortable) return
  const tbody = tableRef.value?.$el?.querySelector('.el-table__body tbody')
  if (!tbody) return
  sortable = Sortable.create(tbody, {
    handle: '.drag-handle',
    animation: 150,
    onEnd: onDragEnd,
  })
}

async function onDragEnd(evt: Sortable.SortableEvent) {
  const { oldIndex, newIndex } = evt
  if (oldIndex === undefined || newIndex === undefined || oldIndex === newIndex) return
  const rows = visibleRows()
  const moved = rows[oldIndex]
  if (!moved) {
    fetchTree()
    return
  }
  rows.splice(oldIndex, 1)
  rows.splice(newIndex, 0, moved)
  const prev = rows[newIndex - 1]
  const next = rows[newIndex + 1]
  const sameParent = (prev && prev.parentId === moved.parentId) || (next && next.parentId === moved.parentId)
  if (!sameParent) {
    ElMessage.warning('仅支持同级分类排序')
    fetchTree()
    return
  }
  const items = rows.filter((r) => r.parentId === moved.parentId).map((r, i) => ({ id: r.id, sort: i }))
  try {
    const res = await sortCategories(items)
    if (ok(res)) {
      ElMessage.success('排序成功')
      fetchTree()
    } else {
      ElMessage.error(res.data.message || '排序失败')
      fetchTree()
    }
  } catch {
    fetchTree()
  }
}

function onExpandChange(row: CategoryNode, expanded: boolean | CategoryNode[]) {
  const isExpandedVal = Array.isArray(expanded) ? expanded.some((r) => r.id === row.id) : expanded
  if (isExpandedVal) expandedKeys.add(row.id)
  else expandedKeys.delete(row.id)
  expandVersion.value++
}

function toggleRow(row: CategoryNode) {
  tableRef.value?.toggleRowExpansion(row)
}

async function onStatusChange(row: CategoryNode, val: number) {
  try {
    const res = await updateCategoryStatus(row.id, val)
    if (ok(res)) {
      row.status = val
      ElMessage.success('操作成功')
    } else {
      ElMessage.error(res.data.message || '操作失败')
    }
  } catch {
    // 失败时 model-value 未变更，开关自动回滚
  }
}

function goProducts(row: CategoryNode) {
  router.push({ path: '/products', query: { categoryId: row.id } })
}

const dialogVisible = ref(false)
const isEdit = ref(false)
const currentId = ref<number | null>(null)
const submitting = ref(false)
const formRef = ref<FormInstance>()
const fileList = ref<UploadFile[]>([])
const form = reactive({ parentId: 0, name: '', code: '', sort: 0, status: 1, image: '' })
const formRules = { name: [{ required: true, message: '请输入分类名称', trigger: 'blur' }] }

const dialogTree = shallowRef<CategoryNode[] | null>(null)
const parentOptions = computed(() => [{ id: 0, name: '顶级分类', code: '', sort: 0, status: 1, image: '', productCount: 0, children: dialogTree.value ?? [] } as CategoryNode])

async function ensureDialogTree() {
  if (dialogTree.value) return
  try {
    const res = await fetchCategoryTree()
    if (ok(res)) dialogTree.value = markRaw(res.data.data || [])
  } catch {
  }
}

function resetForm(parentId = 0) {
  Object.assign(form, { parentId, name: '', code: '', sort: 0, status: 1, image: '' })
  fileList.value = []
}

const pendingInsert = ref<{ refId: number; position: 'before' | 'after' } | null>(null)

function openCreate(parentId = 0) {
  resetForm(parentId)
  pendingInsert.value = null
  isEdit.value = false
  currentId.value = null
  ensureDialogTree()
  dialogVisible.value = true
}

function handleAddCommand(cmd: string, row: CategoryNode) {
  if (cmd === 'child') {
    openCreate(row.id)
  } else if (cmd === 'before' || cmd === 'after') {
    openCreate(row.parentId)
    pendingInsert.value = { refId: row.id, position: cmd }
  }
}

function openEdit(row: CategoryNode) {
  Object.assign(form, {
    parentId: row.parentId,
    name: row.name,
    code: row.code,
    sort: row.sort,
    status: row.status,
    image: row.image || '',
  })
  fileList.value = row.image ? [{ name: row.name, url: row.image, status: 'success', uid: Date.now() }] : []
  isEdit.value = true
  currentId.value = row.id
  ensureDialogTree()
  dialogVisible.value = true
}

function handleCommand(cmd: string, row: CategoryNode) {
  if (cmd === 'edit') openEdit(row)
  else if (cmd === 'child') openCreate(row.id)
  else if (cmd === 'delete') handleDelete(row)
}

async function uploadImage(options: UploadRequestOptions) {
  try {
    const res = await uploadFile(options.file)
    if (ok(res)) {
      form.image = res.data.data.url
      fileList.value = [{ name: options.file.name, url: form.image, status: 'success', uid: options.file.uid }]
      options.onSuccess(res.data.data)
    } else {
      options.onError(new Error(res.data.message || '上传失败') as any)
    }
  } catch (e: any) {
    options.onError(e as any)
  }
}

function onImageRemove() {
  form.image = ''
  fileList.value = []
}

function findSiblings(nodes: CategoryNode[], id: number): CategoryNode[] | null {
  if (nodes.some((n) => n.id === id)) return nodes
  for (const node of nodes) {
    if (node.children?.length) {
      const found = findSiblings(node.children, id)
      if (found) return found
    }
  }
  return null
}

async function applyPendingInsert(newId: number) {
  const pending = pendingInsert.value
  pendingInsert.value = null
  if (!pending) {
    fetchTree()
    return
  }
  const siblings = findSiblings(tree.value, pending.refId)
  if (siblings) {
    const idx = siblings.findIndex((s) => s.id === pending.refId)
    const ids = siblings.map((s) => s.id)
    ids.splice(pending.position === 'before' ? idx : idx + 1, 0, newId)
    try {
      const res = await sortCategories(ids.map((id, i) => ({ id, sort: i })))
      if (!ok(res)) ElMessage.error(res.data.message || '排序失败')
    } catch {
      fetchTree()
      return
    }
  }
  fetchTree()
}

async function handleSubmit() {
  await formRef.value?.validate()
  submitting.value = true
  try {
    const payload = { ...form }
    const res = isEdit.value && currentId.value !== null
      ? await updateCategory(currentId.value, payload)
      : await createCategory(payload)
    if (ok(res)) {
      ElMessage.success(isEdit.value ? '编辑成功' : '新增成功')
      dialogVisible.value = false
      dialogTree.value = null
      if (!isEdit.value) expandAncestors(form.parentId)
      const newId = res.data.data?.id
      if (!isEdit.value && pendingInsert.value && newId) {
        await applyPendingInsert(newId)
      } else {
        pendingInsert.value = null
        fetchTree()
      }
    } else {
      ElMessage.error(res.data.message || '操作失败')
    }
  } finally {
    submitting.value = false
  }
}

async function handleDelete(row: CategoryNode) {
  try {
    await ElMessageBox.confirm(`确认删除分类“${row.name}”？`, '提示', { type: 'warning' })
  } catch {
    return
  }
  const res = await deleteCategory(row.id)
  if (ok(res)) {
    ElMessage.success('删除成功')
    dialogTree.value = null
    fetchTree()
  } else {
    ElMessage.error(res.data.message || '删除失败')
  }
}

async function handleFillImages() {
  try {
    await ElMessageBox.confirm('将自动使用分类下第一个商城商品的图片补全缺失的分类图片，是否继续？', '提示', { type: 'warning' })
  } catch {
    return
  }
  const res = await fillCategoryImages()
  if (ok(res)) {
    ElMessage.success(`已补全 ${res.data.data.updated} 个分类图片`)
    fetchTree()
  } else {
    ElMessage.error(res.data.message || '操作失败')
  }
}

onMounted(fetchTree)
onBeforeUnmount(() => {
  sortable?.destroy()
  sortable = null
})
</script>

<style scoped>
.page { padding: 20px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.page-title { font-size: 18px; font-weight: 600; }
.tip { margin-bottom: 16px; }
.drag-handle { cursor: move; color: #909399; }
.danger-text { color: var(--el-color-danger); }
.category-image { width: 40px; height: 40px; border-radius: 4px; display: block; }
.image-placeholder {
  width: 40px; height: 40px; border-radius: 4px; margin: 0 auto;
  display: flex; align-items: center; justify-content: center;
  background: #f5f7fa; color: #c0c4cc;
}
.category-name { display: flex; align-items: center; gap: 6px; }
.expand-icon { cursor: pointer; transition: transform 0.2s; color: #909399; }
.expand-icon.expanded { transform: rotate(90deg); }
.expand-spacer { width: 16px; }
.folder-icon { color: #e6a23c; }
.full-width { width: 100%; }
.row-add-trigger { color: #c0c4cc; cursor: pointer; visibility: hidden; }
:deep(.el-table__row:hover .row-add-trigger) { visibility: visible; }
:deep(.el-table__body td:first-child .el-table__indent),
:deep(.el-table__body td:first-child .el-table__expand-icon),
:deep(.el-table__body td:first-child .el-table__placeholder) {
  display: none;
}
</style>
