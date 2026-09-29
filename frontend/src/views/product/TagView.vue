<template>
  <div class="page">
    <div class="page-header">
      <span class="page-title">商品标签</span>
      <div class="header-right">
        <span class="stats-text">共 {{ list.length }} 个标签 · {{ badgeCount }} 个已配置角标</span>
        <el-dropdown trigger="click" @command="handleAddCommand">
          <el-button type="primary" :icon="Plus">新增</el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="manual">新增标签</el-dropdown-item>
              <el-dropdown-item command="smart">新增智能标签</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </div>
    <el-card>
      <el-table ref="tableRef" :data="list" v-loading="loading" row-key="id" border>
        <el-table-column label="排序" width="60" align="center">
          <template #default>
            <el-icon class="drag-handle"><Rank /></el-icon>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="70" align="center">
          <template #default="{ row }">
            <el-dropdown trigger="click" @command="(cmd: string) => handleRowCommand(cmd, row)">
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
        <el-table-column prop="name" label="商品标签" min-width="140">
          <template #default="{ row }">
            <el-tag :color="row.color" style="color: #fff; border: none">{{ row.name }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="标签类型" width="110" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.type === 2">智能标签</el-tag>
            <el-tag v-else type="primary">手动标签</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="标签规则" min-width="180">
          <template #default="{ row }">{{ row.type === 2 ? ruleSummary(row.rule) : '-' }}</template>
        </el-table-column>
        <el-table-column label="标签状态" width="100" align="center">
          <template #default="{ row }">{{ row.type === 2 ? '生效中' : '-' }}</template>
        </el-table-column>
        <el-table-column label="在小程序商城支持筛选" width="170" align="center">
          <template #default="{ row }">
            <el-switch :model-value="row.filterable === 1" @change="(v: boolean) => onSwitch(row, 'filterable', v)" />
          </template>
        </el-table-column>
        <el-table-column label="在小程序商城列表展示" width="170" align="center">
          <template #default="{ row }">
            <el-switch :model-value="row.showInList === 1" @change="(v: boolean) => onSwitch(row, 'showInList', v)" />
          </template>
        </el-table-column>
        <el-table-column label="商品角标" width="100" align="center">
          <template #default="{ row }">{{ row.showBadge === 1 ? '已配置' : '未配置' }}</template>
        </el-table-column>
        <el-table-column label="绑定商品" width="100" align="center">
          <template #default="{ row }">
            <el-link type="primary" :underline="false" @click="goProducts(row)">{{ row.productCount ?? 0 }}</el-link>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="manualDialogVisible" :title="isEdit ? '编辑标签' : '新增标签'" width="480px" destroy-on-close>
      <el-form ref="manualFormRef" :model="manualForm" :rules="manualRules" label-width="170px">
        <el-form-item label="标签名称" prop="name">
          <el-input v-model="manualForm.name" placeholder="标签名称" maxlength="20" show-word-limit />
        </el-form-item>
        <el-form-item label="在小程序商城支持筛选">
          <el-switch v-model="manualForm.filterable" :active-value="1" :inactive-value="0" />
        </el-form-item>
        <el-form-item label="在小程序商城列表展示">
          <el-switch v-model="manualForm.showInList" :active-value="1" :inactive-value="0" />
        </el-form-item>
        <el-form-item label="在商品卡片展示角标">
          <el-switch v-model="manualForm.showBadge" :active-value="1" :inactive-value="0" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="manualDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submitManual">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="smartDialogVisible" :title="isEdit ? '编辑智能标签' : '新增智能标签'" width="60%" destroy-on-close>
      <el-form ref="smartFormRef" :model="smartForm" :rules="smartRules" label-width="130px">
        <el-form-item label="智能标签名称" prop="name">
          <el-input v-model="smartForm.name" placeholder="智能标签名称" maxlength="20" show-word-limit style="max-width: 360px" />
        </el-form-item>
        <el-form-item label="标记商品范围">
          <div>
            <el-radio-group v-model="smartForm.scope">
              <el-radio value="all">全部商品</el-radio>
              <el-radio value="specific">指定商品</el-radio>
            </el-radio-group>
            <RemoteSelect
              v-if="smartForm.scope === 'specific'"
              v-model="smartForm.productIds"
              multiple
              api-url="/api/v1/products"
              placeholder="请选择商品"
              class="scope-select"
            />
            <div class="hint-text">{{ scopeHint }}</div>
          </div>
        </el-form-item>
        <el-form-item label="标签条件">
          <el-radio-group v-model="smartForm.match">
            <el-radio value="all">同时满足</el-radio>
            <el-radio value="any">满足其一</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label=" ">
          <div class="conditions-area">
            <div class="condition-tabs">
              <div
                v-for="(cond, idx) in smartForm.conditions"
                :key="idx"
                class="condition-tab"
                :class="{ active: activeCondition === idx }"
                @click="activeCondition = idx"
              >
                <el-icon class="check-icon"><CircleCheckFilled /></el-icon>
                <span>条件{{ idx + 1 }}</span>
                <el-dropdown
                  v-if="smartForm.conditions.length > 1"
                  trigger="click"
                  @command="(cmd: string) => cmd === 'delete' && removeCondition(idx)"
                >
                  <el-icon class="tab-more" @click.stop><MoreFilled /></el-icon>
                  <template #dropdown>
                    <el-dropdown-menu>
                      <el-dropdown-item command="delete"><span class="danger-text">删除</span></el-dropdown-item>
                    </el-dropdown-menu>
                  </template>
                </el-dropdown>
              </div>
              <el-link type="primary" :underline="false" class="add-condition" @click="addCondition">＋ 添加标签条件</el-link>
            </div>
            <div v-if="currentCondition" class="condition-card">
              <el-form-item label="数据类型" label-width="110px">
                <el-select v-model="currentCondition.dataType" style="width: 200px">
                  <el-option label="销售数据" value="sales" />
                </el-select>
              </el-form-item>
              <el-form-item label="销售统计时间范围" label-width="130px">
                <el-select v-model="currentCondition.timeRange" style="width: 200px">
                  <el-option label="本月" value="month" />
                  <el-option label="近三个月" value="quarter" />
                  <el-option label="近半年" value="halfYear" />
                  <el-option label="近一年" value="year" />
                </el-select>
              </el-form-item>
              <el-form-item label="销售统计客户范围" label-width="130px">
                <div>
                  <el-radio-group v-model="currentCondition.customerScope">
                    <el-radio value="all">全部客户</el-radio>
                    <el-radio value="specific">指定客户</el-radio>
                  </el-radio-group>
                  <RemoteSelect
                    v-if="currentCondition.customerScope === 'specific'"
                    v-model="currentCondition.customerIds"
                    multiple
                    api-url="/api/v1/customers"
                    placeholder="请选择客户"
                    class="scope-select"
                  />
                </div>
              </el-form-item>
              <el-form-item label="条件范围" label-width="110px">
                <div class="range-area">
                  <div class="range-tip">以下条件需同时满足</div>
                  <el-radio-group v-model="currentCondition.rangeMode">
                    <el-radio value="spec">按规格</el-radio>
                    <el-radio value="product">按商品</el-radio>
                  </el-radio-group>
                  <div v-for="(range, rIdx) in currentCondition.ranges" :key="rIdx" class="range-row">
                    <el-select
                      v-if="currentCondition.rangeMode === 'spec'"
                      v-model="range.target"
                      placeholder="请选择规格"
                      filterable
                      style="width: 200px"
                    >
                      <el-option v-for="opt in specOptions" :key="opt.value" :label="opt.label" :value="opt.value" />
                    </el-select>
                    <RemoteSelect
                      v-else
                      v-model="range.target"
                      api-url="/api/v1/products"
                      placeholder="请选择商品"
                      style="width: 200px"
                    />
                    <el-select v-model="range.op" style="width: 110px">
                      <el-option label="大于" value="gt" />
                      <el-option label="大于等于" value="gte" />
                      <el-option label="小于" value="lt" />
                      <el-option label="小于等于" value="lte" />
                    </el-select>
                    <el-input-number v-model="range.value" :min="0" :controls="false" style="width: 120px" placeholder="数值" />
                    <el-button text type="danger" :icon="Delete" @click="removeRange(rIdx)" />
                  </div>
                  <el-button class="add-range-btn" @click="addRange">＋ 添加条件范围</el-button>
                </div>
              </el-form-item>
            </div>
          </div>
        </el-form-item>
        <el-form-item label="在小程序商城支持筛选">
          <el-switch v-model="smartForm.filterable" :active-value="1" :inactive-value="0" />
        </el-form-item>
        <el-form-item label="在商品卡片展示角标">
          <el-switch v-model="smartForm.showBadge" :active-value="1" :inactive-value="0" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="smartDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submitSmart">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox, type FormInstance } from 'element-plus'
import { CircleCheckFilled, Delete, More, MoreFilled, Plus, Rank } from '@element-plus/icons-vue'
import Sortable from 'sortablejs'
import api from '@/api/client'
import RemoteSelect from '@/components/RemoteSelect.vue'
import { createTag, deleteTag, fetchTags, sortTags, updateTag, type ProductTag, type TagPayload } from '@/api/tag'

interface RuleRange {
  target: string | number | undefined
  op: 'gt' | 'gte' | 'lt' | 'lte'
  value: number
}

interface RuleCondition {
  dataType: 'sales'
  timeRange: 'month' | 'quarter' | 'halfYear' | 'year'
  customerScope: 'all' | 'specific'
  customerIds: number[]
  rangeMode: 'spec' | 'product'
  ranges: RuleRange[]
}

interface TagRule {
  scope: 'all' | 'specific'
  productIds: number[]
  match: 'all' | 'any'
  conditions: RuleCondition[]
}

const router = useRouter()
const list = ref<ProductTag[]>([])
const loading = ref(false)
const tableRef = ref()
let sortable: Sortable | null = null

const badgeCount = computed(() => list.value.filter((t) => t.showBadge === 1).length)

function ok(res: any) {
  return res.data.code === 0 || res.data.code === 200
}

function destroySortable() {
  sortable?.destroy()
  sortable = null
}

function initSortable() {
  destroySortable()
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
    const res = await fetchTags()
    if (ok(res)) {
      list.value = res.data.data?.list || []
      await nextTick()
      initSortable()
    } else {
      ElMessage.error(res.data.message || '获取标签失败')
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
    const res = await sortTags(items)
    if (ok(res)) {
      ElMessage.success('排序成功')
    } else {
      ElMessage.error(res.data.message || '排序失败')
    }
  } finally {
    fetchList()
  }
}

const timeRangeLabels: Record<string, string> = {
  month: '本月',
  quarter: '近三个月',
  halfYear: '近半年',
  year: '近一年',
}

function parseRule(rule: string): TagRule | null {
  if (!rule) return null
  try {
    const parsed = JSON.parse(rule)
    if (!parsed || !Array.isArray(parsed.conditions)) return null
    return parsed as TagRule
  } catch {
    return null
  }
}

function ruleSummary(rule: string): string {
  const parsed = parseRule(rule)
  if (!parsed || parsed.conditions.length === 0) return '-'
  const first = parsed.conditions[0]
  if (!first) return '-'
  const parts = ['销售数据', timeRangeLabels[first.timeRange] || '-', first.customerScope === 'specific' ? '指定客户' : '全部客户']
  let summary = parts.join(' · ')
  if (parsed.conditions.length > 1) summary += ` 等 ${parsed.conditions.length} 组条件`
  return summary
}

function toPayload(row: ProductTag): TagPayload {
  return {
    name: row.name,
    color: row.color || '#409EFF',
    type: row.type,
    rule: row.rule || '',
    filterable: row.filterable,
    showInList: row.showInList,
    showBadge: row.showBadge,
    sort: row.sort,
    status: row.status,
  }
}

async function onSwitch(row: ProductTag, field: 'filterable' | 'showInList', value: boolean) {
  const payload = toPayload(row)
  payload[field] = value ? 1 : 0
  const res = await updateTag(row.id, payload)
  if (ok(res)) {
    ElMessage.success('已更新')
    fetchList()
  } else {
    ElMessage.error(res.data.message || '更新失败')
    fetchList()
  }
}

function goProducts(row: ProductTag) {
  router.push(`/products?tagId=${row.id}`)
}

const isEdit = ref(false)
const currentId = ref<number | null>(null)
const submitting = ref(false)

const manualDialogVisible = ref(false)
const manualFormRef = ref<FormInstance>()
const manualForm = reactive({ name: '', filterable: 1, showInList: 0, showBadge: 0 })
const manualRules = {
  name: [{ required: true, message: '请输入标签名称', trigger: 'blur' }],
}

const smartDialogVisible = ref(false)
const smartFormRef = ref<FormInstance>()
const activeCondition = ref(0)
const smartForm = reactive({
  name: '',
  scope: 'all' as 'all' | 'specific',
  productIds: [] as number[],
  match: 'all' as 'all' | 'any',
  conditions: [] as RuleCondition[],
  filterable: 1,
  showBadge: 0,
})
const smartRules = {
  name: [{ required: true, message: '请输入智能标签名称', trigger: 'blur' }],
}

const scopeHint = computed(() =>
  smartForm.scope === 'specific'
    ? '本智能标签将针对指定商品中满足标签条件的商品进行标记'
    : '本智能标签将针对全部商品中满足标签条件的商品进行标记'
)

const currentCondition = computed(() => smartForm.conditions[activeCondition.value])

function defaultCondition(): RuleCondition {
  return { dataType: 'sales', timeRange: 'month', customerScope: 'all', customerIds: [], rangeMode: 'product', ranges: [] }
}

function addCondition() {
  smartForm.conditions.push(defaultCondition())
  activeCondition.value = smartForm.conditions.length - 1
}

function removeCondition(idx: number) {
  smartForm.conditions.splice(idx, 1)
  if (activeCondition.value >= smartForm.conditions.length) {
    activeCondition.value = Math.max(0, smartForm.conditions.length - 1)
  }
}

function addRange() {
  const cond = currentCondition.value
  if (!cond) return
  cond.ranges.push({ target: undefined, op: 'gte', value: 0 })
}

function removeRange(idx: number) {
  currentCondition.value?.ranges.splice(idx, 1)
}

const specOptions = ref<{ label: string; value: string }[]>([])

async function loadSpecOptions() {
  try {
    const res = await api.get('/api/v1/product-settings/specs', { params: { page: 1, pageSize: 500 } })
    if (ok(res)) {
      const specs = res.data.data?.list || []
      const opts: { label: string; value: string }[] = []
      for (const s of specs) {
        const values = String(s.values || '').split(',').map((v: string) => v.trim()).filter(Boolean)
        if (values.length === 0) {
          opts.push({ label: s.name, value: s.name })
        } else {
          for (const v of values) {
            opts.push({ label: `${s.name}：${v}`, value: `${s.name}:${v}` })
          }
        }
      }
      specOptions.value = opts
    }
  } catch {
    specOptions.value = []
  }
}

function handleAddCommand(cmd: string) {
  isEdit.value = false
  currentId.value = null
  if (cmd === 'smart') openSmartCreate()
  else openManualCreate()
}

function openManualCreate() {
  Object.assign(manualForm, { name: '', filterable: 1, showInList: 0, showBadge: 0 })
  manualDialogVisible.value = true
}

function openSmartCreate() {
  Object.assign(smartForm, {
    name: '',
    scope: 'all',
    productIds: [],
    match: 'all',
    conditions: [defaultCondition()],
    filterable: 1,
    showBadge: 0,
  })
  activeCondition.value = 0
  smartDialogVisible.value = true
  loadSpecOptions()
}

function openEdit(row: ProductTag) {
  isEdit.value = true
  currentId.value = row.id
  if (row.type === 2) {
    const rule = parseRule(row.rule)
    Object.assign(smartForm, {
      name: row.name,
      scope: rule?.scope || 'all',
      productIds: rule?.productIds || [],
      match: rule?.match || 'all',
      conditions: rule?.conditions?.length ? rule.conditions : [defaultCondition()],
      filterable: row.filterable,
      showBadge: row.showBadge,
    })
    activeCondition.value = 0
    smartDialogVisible.value = true
    loadSpecOptions()
  } else {
    Object.assign(manualForm, {
      name: row.name,
      filterable: row.filterable,
      showInList: row.showInList,
      showBadge: row.showBadge,
    })
    manualDialogVisible.value = true
  }
}

function handleRowCommand(cmd: string, row: ProductTag) {
  if (cmd === 'edit') openEdit(row)
  else if (cmd === 'delete') handleDelete(row)
}

async function submitManual() {
  await manualFormRef.value?.validate()
  submitting.value = true
  try {
    const base = isEdit.value && currentId.value !== null ? list.value.find((t) => t.id === currentId.value) : undefined
    const payload: TagPayload = {
      name: manualForm.name,
      color: base?.color || '#409EFF',
      type: 1,
      rule: '',
      filterable: manualForm.filterable,
      showInList: manualForm.showInList,
      showBadge: manualForm.showBadge,
      sort: base?.sort ?? 0,
      status: base?.status ?? 1,
    }
    const res = isEdit.value && currentId.value !== null
      ? await updateTag(currentId.value, payload)
      : await createTag(payload)
    if (ok(res)) {
      ElMessage.success(isEdit.value ? '编辑成功' : '新增成功')
      manualDialogVisible.value = false
      fetchList()
    } else {
      ElMessage.error(res.data.message || '操作失败')
    }
  } finally {
    submitting.value = false
  }
}

async function submitSmart() {
  await smartFormRef.value?.validate()
  submitting.value = true
  try {
    const rule: TagRule = {
      scope: smartForm.scope,
      productIds: smartForm.scope === 'specific' ? smartForm.productIds : [],
      match: smartForm.match,
      conditions: smartForm.conditions.map((cond) => ({
        ...cond,
        customerIds: cond.customerScope === 'specific' ? cond.customerIds : [],
      })),
    }
    const base = isEdit.value && currentId.value !== null ? list.value.find((t) => t.id === currentId.value) : undefined
    const payload: TagPayload = {
      name: smartForm.name,
      color: base?.color || '#409EFF',
      type: 2,
      rule: JSON.stringify(rule),
      filterable: smartForm.filterable,
      showInList: base?.showInList ?? 0,
      showBadge: smartForm.showBadge,
      sort: base?.sort ?? 0,
      status: base?.status ?? 1,
    }
    const res = isEdit.value && currentId.value !== null
      ? await updateTag(currentId.value, payload)
      : await createTag(payload)
    if (ok(res)) {
      ElMessage.success(isEdit.value ? '编辑成功' : '新增成功')
      smartDialogVisible.value = false
      fetchList()
    } else {
      ElMessage.error(res.data.message || '操作失败')
    }
  } finally {
    submitting.value = false
  }
}

async function handleDelete(row: ProductTag) {
  try {
    await ElMessageBox.confirm(`确认删除标签“${row.name}”？`, '提示', { type: 'warning' })
  } catch {
    return
  }
  const res = await deleteTag(row.id)
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
.header-right { display: flex; align-items: center; gap: 16px; }
.stats-text { color: #909399; font-size: 13px; }
.drag-handle { cursor: move; color: #909399; }
.danger-text { color: var(--el-color-danger); }
.hint-text { color: #909399; font-size: 12px; margin-top: 6px; }
.scope-select { margin-top: 8px; max-width: 420px; }
.conditions-area { width: 100%; }
.condition-tabs { display: flex; align-items: center; flex-wrap: wrap; gap: 8px; margin-bottom: 12px; }
.condition-tab {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 12px;
  border: 1px solid var(--el-border-color);
  border-radius: 4px;
  cursor: pointer;
  font-size: 13px;
}
.condition-tab.active { border-color: var(--el-color-primary); color: var(--el-color-primary); }
.check-icon { color: var(--el-color-success); }
.tab-more { cursor: pointer; }
.add-condition { margin-left: 8px; }
.condition-card {
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 6px;
  padding: 16px 16px 0;
  background: var(--el-fill-color-blank);
}
.range-area { width: 100%; }
.range-tip { color: #909399; font-size: 12px; margin-bottom: 8px; }
.range-row { display: flex; align-items: center; gap: 8px; margin-top: 10px; }
.add-range-btn {
  margin-top: 10px;
  width: 100%;
  border-style: dashed;
}
</style>
