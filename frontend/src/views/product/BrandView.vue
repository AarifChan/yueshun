<template>
  <div class="page">
    <div class="page-header">
      <span class="page-title">商品品牌</span>
      <el-button type="primary" :icon="Plus" @click="openCreate">新增品牌</el-button>
    </div>
    <div class="page-body">
      <el-card class="tree-panel" shadow="never">
        <SideTreePanel
          ref="treeRef"
          title="品牌分类"
          :data="sideTree"
          @node-click="onNodeClick"
          @edit-click="openCategoryManage"
        />
      </el-card>
      <el-card class="table-panel" shadow="never">
        <div class="search-bar">
          <el-input
            v-model="keyword"
            placeholder="请输入商品品牌"
            clearable
            class="search-input"
            @keyup.enter="handleSearch"
            @clear="handleSearch"
          />
          <el-button type="primary" @click="handleSearch">搜索</el-button>
        </div>
        <el-table ref="tableRef" :data="list" v-loading="loading" row-key="id" border>
          <el-table-column type="index" label="序" width="60" align="center" :index="indexMethod" />
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
          <el-table-column label="品牌图片" width="90" align="center">
            <template #default="{ row }">
              <el-image
                v-if="row.image"
                :src="row.image"
                :preview-src-list="[row.image]"
                preview-teleported
                fit="cover"
                class="brand-image"
              />
              <el-icon v-else class="image-placeholder"><Picture /></el-icon>
            </template>
          </el-table-column>
          <el-table-column prop="name" label="商品品牌" min-width="160" />
          <el-table-column label="品牌分类" min-width="120">
            <template #default="{ row }">{{ row.categoryName || '-' }}</template>
          </el-table-column>
          <el-table-column label="应用为商品展示图" width="140" align="center">
            <template #default="{ row }">
              <span class="dot-badge">
                <i class="dot" :class="{ on: row.applyImage === 1 }" />
                {{ row.applyImage === 1 ? '是' : '否' }}
              </span>
            </template>
          </el-table-column>
          <el-table-column label="绑定商品" width="100" align="center">
            <template #default="{ row }">
              <el-link type="primary" :underline="false" @click="goProducts(row)">{{ row.productCount }}</el-link>
            </template>
          </el-table-column>
        </el-table>
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[30, 50, 100]"
          layout="total, prev, pager, next, sizes, jumper"
          class="pagination"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </el-card>
    </div>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑品牌' : '新增品牌'" width="480px" destroy-on-close>
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="120px">
        <el-form-item label="品牌名称" prop="name">
          <el-input v-model="form.name" placeholder="品牌名称" maxlength="32" show-word-limit />
        </el-form-item>
        <el-form-item label="品牌分类" prop="categoryId">
          <el-tree-select
            v-model="form.categoryId"
            :data="dialogTree"
            node-key="id"
            :props="{ label: 'name', children: 'children' }"
            check-strictly
            class="full-width"
          />
        </el-form-item>
        <el-form-item label="品牌图片">
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
        <el-form-item label="应用为商品展示图">
          <el-switch v-model="form.applyImage" :active-value="1" :inactive-value="0" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="categoryDialogVisible" title="品牌分类管理" width="460px" destroy-on-close>
      <div class="category-manage-header">
        <el-button type="primary" size="small" :icon="Plus" @click="handleAddRootCategory">新增一级分类</el-button>
      </div>
      <el-tree
        v-if="categoryTree.length > 0"
        :data="categoryTree"
        node-key="id"
        :props="{ label: 'name', children: 'children' }"
        :expand-on-click-node="false"
        default-expand-all
      >
        <template #default="{ data }">
          <span class="manage-node">
            <span class="manage-node-name">{{ data.name }}</span>
            <span class="manage-actions">
              <el-link type="primary" :underline="false" @click.stop="handleAddChildCategory(data)">新增子级</el-link>
              <el-link type="primary" :underline="false" @click.stop="handleRenameCategory(data)">重命名</el-link>
              <el-link type="danger" :underline="false" @click.stop="handleDeleteCategory(data)">删除</el-link>
            </span>
          </span>
        </template>
      </el-tree>
      <el-empty v-else description="暂无分类" :image-size="60" />
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox, type FormInstance, type UploadFile, type UploadRequestOptions } from 'element-plus'
import { More, Picture, Plus, Rank } from '@element-plus/icons-vue'
import Sortable from 'sortablejs'
import SideTreePanel from '@/components/SideTreePanel.vue'
import {
  createBrand,
  createBrandCategory,
  deleteBrand,
  deleteBrandCategory,
  fetchBrandCategoryTree,
  fetchBrands,
  sortBrands,
  updateBrand,
  updateBrandCategory,
  uploadFile,
  type Brand,
  type BrandCategoryNode,
} from '@/api/brand'

interface SideNode {
  id: number | string
  name: string
  children?: SideNode[]
}

const router = useRouter()
const list = ref<Brand[]>([])
const loading = ref(false)
const tableRef = ref()
const treeRef = ref()
const categoryTree = ref<BrandCategoryNode[]>([])
const currentNodeKey = ref<number | string>('all')
const keyword = ref('')
const appliedKeyword = ref('')
const page = ref(1)
const pageSize = ref(30)
const total = ref(0)
let sortable: Sortable | null = null

const isFiltered = computed(() => appliedKeyword.value !== '' || currentNodeKey.value !== 'all')

const sideTree = computed<SideNode[]>(() => [
  { id: 'all', name: '全部分类' },
  { id: 0, name: '未分类' },
  ...(categoryTree.value as unknown as SideNode[]),
])

const dialogTree = computed<SideNode[]>(() => [
  { id: 0, name: '未分类' },
  ...(categoryTree.value as unknown as SideNode[]),
])

function ok(res: any) {
  return res.data.code === 0 || res.data.code === 200
}

function indexMethod(i: number) {
  return (page.value - 1) * pageSize.value + i + 1
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
    const params: { page: number; pageSize: number; keyword?: string; categoryId?: number } = {
      page: page.value,
      pageSize: pageSize.value,
    }
    if (appliedKeyword.value) params.keyword = appliedKeyword.value
    if (currentNodeKey.value !== 'all') params.categoryId = Number(currentNodeKey.value)
    const res = await fetchBrands(params)
    if (ok(res)) {
      list.value = res.data.data?.list || []
      total.value = res.data.data?.total || 0
      await nextTick()
      initSortable()
    } else {
      ElMessage.error(res.data.message || '获取品牌失败')
    }
  } finally {
    loading.value = false
  }
}

async function loadCategoryTree() {
  const res = await fetchBrandCategoryTree()
  if (ok(res)) {
    categoryTree.value = res.data.data || []
  }
}

function onNodeClick(data: SideNode) {
  currentNodeKey.value = data.id
  page.value = 1
  fetchList()
}

function handleSearch() {
  appliedKeyword.value = keyword.value.trim()
  page.value = 1
  fetchList()
}

function handleSizeChange(size: number) {
  pageSize.value = size
  page.value = 1
  fetchList()
}

function handleCurrentChange(p: number) {
  page.value = p
  fetchList()
}

function goProducts(row: Brand) {
  router.push(`/products?brandId=${row.id}`)
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
  const base = (page.value - 1) * pageSize.value
  const items = rows.map((r, i) => ({ id: r.id, sort: base + i }))
  try {
    const res = await sortBrands(items)
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
const fileList = ref<UploadFile[]>([])
const form = reactive({ name: '', categoryId: 0, image: '', applyImage: 0, sort: 0 })
const formRules = {
  name: [{ required: true, message: '请输入品牌名称', trigger: 'blur' }],
}

function openCreate() {
  Object.assign(form, { name: '', categoryId: 0, image: '', applyImage: 0, sort: 0 })
  fileList.value = []
  isEdit.value = false
  currentId.value = null
  dialogVisible.value = true
}

function openEdit(row: Brand) {
  Object.assign(form, {
    name: row.name,
    categoryId: row.categoryId || 0,
    image: row.image || '',
    applyImage: row.applyImage || 0,
    sort: row.sort || 0,
  })
  fileList.value = row.image ? [{ name: row.name, url: row.image, status: 'success', uid: Date.now() }] : []
  isEdit.value = true
  currentId.value = row.id
  dialogVisible.value = true
}

function handleCommand(cmd: string, row: Brand) {
  if (cmd === 'edit') openEdit(row)
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

async function handleSubmit() {
  await formRef.value?.validate()
  submitting.value = true
  try {
    const payload = {
      name: form.name,
      code: '',
      description: '',
      image: form.image,
      categoryId: form.categoryId,
      applyImage: form.applyImage,
      sort: form.sort,
      status: 1,
    }
    const res = isEdit.value && currentId.value !== null
      ? await updateBrand(currentId.value, payload)
      : await createBrand(payload)
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

async function handleDelete(row: Brand) {
  try {
    await ElMessageBox.confirm(`确认删除品牌“${row.name}”？`, '提示', { type: 'warning' })
  } catch {
    return
  }
  const res = await deleteBrand(row.id)
  if (ok(res)) {
    ElMessage.success('删除成功')
    fetchList()
  } else {
    ElMessage.error(res.data.message || '删除失败')
  }
}

const categoryDialogVisible = ref(false)

function openCategoryManage() {
  categoryDialogVisible.value = true
}

async function promptCategoryName(title: string, defaultValue = '') {
  try {
    const { value } = await ElMessageBox.prompt('请输入分类名称', title, {
      inputValue: defaultValue,
      inputPattern: /\S+/,
      inputErrorMessage: '请输入分类名称',
    })
    return value.trim()
  } catch {
    return null
  }
}

async function submitCategoryCreate(parentId: number, name: string) {
  const res = await createBrandCategory({ parentId, name, sort: 0, status: 1 })
  if (ok(res)) {
    ElMessage.success('新增成功')
    loadCategoryTree()
  } else {
    ElMessage.error(res.data.message || '新增失败')
  }
}

async function handleAddRootCategory() {
  const name = await promptCategoryName('新增一级分类')
  if (name === null) return
  await submitCategoryCreate(0, name)
}

async function handleAddChildCategory(data: BrandCategoryNode) {
  const name = await promptCategoryName(`新增子级分类（${data.name}）`)
  if (name === null) return
  await submitCategoryCreate(data.id, name)
}

async function handleRenameCategory(data: BrandCategoryNode) {
  const name = await promptCategoryName('重命名分类', data.name)
  if (name === null || name === data.name) return
  const res = await updateBrandCategory(data.id, {
    parentId: data.parentId,
    name,
    sort: data.sort,
    status: data.status,
  })
  if (ok(res)) {
    ElMessage.success('重命名成功')
    loadCategoryTree()
  } else {
    ElMessage.error(res.data.message || '重命名失败')
  }
}

async function handleDeleteCategory(data: BrandCategoryNode) {
  try {
    await ElMessageBox.confirm(`确认删除分类“${data.name}”？`, '提示', { type: 'warning' })
  } catch {
    return
  }
  const res = await deleteBrandCategory(data.id)
  if (ok(res)) {
    ElMessage.success('删除成功')
    if (currentNodeKey.value === data.id) {
      currentNodeKey.value = 'all'
      treeRef.value?.setCurrentKey('all')
      page.value = 1
      fetchList()
    }
    loadCategoryTree()
  } else {
    ElMessage.error(res.data.message || '删除失败')
  }
}

onMounted(() => {
  loadCategoryTree()
  fetchList()
  nextTick(() => treeRef.value?.setCurrentKey('all'))
})
onBeforeUnmount(destroySortable)
</script>

<style scoped>
.page { padding: 20px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.page-title { font-size: 18px; font-weight: 600; }
.page-body { display: flex; gap: 16px; align-items: flex-start; }
.tree-panel { width: 200px; flex-shrink: 0; }
.manage-node { flex: 1; display: flex; justify-content: space-between; align-items: center; padding-right: 8px; }
.manage-actions { display: none; gap: 10px; }
.manage-node:hover .manage-actions { display: inline-flex; }
.category-manage-header { display: flex; justify-content: flex-end; margin-bottom: 12px; }
.table-panel { flex: 1; min-width: 0; }
.search-bar { display: flex; gap: 12px; margin-bottom: 16px; }
.search-input { width: 240px; }
.brand-image { width: 40px; height: 40px; border-radius: 4px; vertical-align: middle; }
.image-placeholder { font-size: 22px; color: #c0c4cc; }
.dot-badge { display: inline-flex; align-items: center; gap: 6px; color: #606266; }
.dot-badge .dot { width: 8px; height: 8px; border-radius: 50%; background: #c0c4cc; }
.dot-badge .dot.on { background: #67c23a; }
.pagination { margin-top: 16px; justify-content: flex-end; }
.drag-handle { cursor: move; color: #909399; }
.drag-handle.disabled { cursor: not-allowed; opacity: 0.4; }
.danger-text { color: var(--el-color-danger); }
.full-width { width: 100%; }
</style>
