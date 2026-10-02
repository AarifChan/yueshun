<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">组装拆装单</h2>
      <el-tag v-if="form.status === 'completed'" type="success">已过账</el-tag>
      <el-tag v-else-if="form.id" type="info">草稿</el-tag>
    </div>

    <div class="form-panel" v-loading="loading">
      <el-form :model="form" label-width="80px" :disabled="!editable" class="bill-form">
        <el-row :gutter="16">
          <el-col :span="6">
            <el-form-item label="出库仓库" required>
              <RemoteSelect v-model="form.outWarehouseId" api-url="/api/v1/warehouses" placeholder="请选择" />
            </el-form-item>
          </el-col>
          <el-col :span="6">
            <el-form-item label="入库仓库" required>
              <RemoteSelect v-model="form.inWarehouseId" api-url="/api/v1/warehouses" placeholder="请选择" />
            </el-form-item>
          </el-col>
          <el-col :span="6">
            <el-form-item label="拆装模板">
              <el-select v-model="form.templateId" placeholder="可不选" clearable style="width: 100%" @change="applyTemplate">
                <el-option v-for="t in templates" :key="t.id" :label="t.name" :value="t.id" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="6">
            <el-form-item label="加工费用">
              <el-input-number v-model="form.fee" :min="0" :precision="2" controls-position="right" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="6">
            <el-form-item label="录单时间">
              <el-date-picker v-model="form.billDate" type="date" value-format="YYYY-MM-DD" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="6">
            <el-form-item label="经手人">
              <RemoteSelect v-model="form.handlerId" api-url="/api/v1/employees" placeholder="请选择" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="单据备注"><el-input v-model="form.remark" placeholder="请输入" /></el-form-item>
          </el-col>
        </el-row>
      </el-form>
    </div>

    <div class="items-panel">
      <div class="items-toolbar">
        <span class="items-title">出库明细（原材料/被拆商品）</span>
        <el-button v-if="editable" type="primary" size="small" @click="addRow('out')">添加行</el-button>
      </div>
      <el-table :data="outItems" border size="small">
        <el-table-column type="index" width="50" />
        <el-table-column label="商品名称" min-width="220">
          <template #default="{ row }">
            <ProductCell v-if="editable" :row="row" :options="productOptions" :loading="productLoading" @search="searchProducts" @change="handleProductChange" />
            <span v-else>{{ row.productName }}</span>
          </template>
        </el-table-column>
        <el-table-column label="规格" width="110"><template #default="{ row }">{{ row.specification || '-' }}</template></el-table-column>
        <el-table-column label="单位" width="70"><template #default="{ row }">{{ row.unit || '-' }}</template></el-table-column>
        <el-table-column label="数量" width="130">
          <template #default="{ row }">
            <el-input-number v-if="editable" v-model="row.quantity" :min="0" :precision="2" controls-position="right" style="width: 100%" />
            <span v-else>{{ row.quantity }}</span>
          </template>
        </el-table-column>
        <el-table-column label="单价" width="130">
          <template #default="{ row }">
            <el-input-number v-if="editable" v-model="row.price" :min="0" :precision="4" controls-position="right" style="width: 100%" />
            <span v-else>{{ row.price }}</span>
          </template>
        </el-table-column>
        <el-table-column label="金额" width="110" align="right">
          <template #default="{ row }">{{ ((row.quantity || 0) * (row.price || 0)).toFixed(2) }}</template>
        </el-table-column>
        <el-table-column v-if="editable" label="操作" width="64">
          <template #default="{ $index }">
            <el-button type="danger" link size="small" @click="removeRow('out', $index)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="items-toolbar" style="margin-top: 16px">
        <span class="items-title">入库明细（成品/拆出商品）</span>
        <el-button v-if="editable" type="primary" size="small" @click="addRow('in')">添加行</el-button>
      </div>
      <el-table :data="inItems" border size="small">
        <el-table-column type="index" width="50" />
        <el-table-column label="商品名称" min-width="220">
          <template #default="{ row }">
            <ProductCell v-if="editable" :row="row" :options="productOptions" :loading="productLoading" @search="searchProducts" @change="handleProductChange" />
            <span v-else>{{ row.productName }}</span>
          </template>
        </el-table-column>
        <el-table-column label="规格" width="110"><template #default="{ row }">{{ row.specification || '-' }}</template></el-table-column>
        <el-table-column label="单位" width="70"><template #default="{ row }">{{ row.unit || '-' }}</template></el-table-column>
        <el-table-column label="数量" width="130">
          <template #default="{ row }">
            <el-input-number v-if="editable" v-model="row.quantity" :min="0" :precision="2" controls-position="right" style="width: 100%" />
            <span v-else>{{ row.quantity }}</span>
          </template>
        </el-table-column>
        <el-table-column label="单价" width="130">
          <template #default="{ row }">
            <el-input-number v-if="editable" v-model="row.price" :min="0" :precision="4" controls-position="right" style="width: 100%" />
            <span v-else>{{ row.price }}</span>
          </template>
        </el-table-column>
        <el-table-column label="金额" width="110" align="right">
          <template #default="{ row }">{{ ((row.quantity || 0) * (row.price || 0)).toFixed(2) }}</template>
        </el-table-column>
        <el-table-column v-if="editable" label="操作" width="64">
          <template #default="{ $index }">
            <el-button type="danger" link size="small" @click="removeRow('in', $index)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="summary">
        <span>出库金额: ¥ {{ outAmount }}</span>
        <span>入库金额: ¥ {{ inAmount }}</span>
      </div>
    </div>

    <div class="footer-bar">
      <span class="operator">制单人：{{ operatorName }}</span>
      <div class="footer-actions">
        <el-button @click="goBack">返 回</el-button>
        <template v-if="editable">
          <el-button :loading="saving" @click="save(false)">存入草稿</el-button>
          <el-button type="primary" :loading="saving" @click="save(true)">保存并过账</el-button>
        </template>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import dayjs from 'dayjs'
import { ElMessage, ElSelect, ElOption } from 'element-plus'
import RemoteSelect from '@/components/RemoteSelect.vue'
import api from '@/api/client'
import { useAuthStore } from '@/stores/auth'

interface ItemRow {
  direction: 'out' | 'in'
  productId?: number; productName?: string; specification?: string; unit?: string
  quantity: number; price: number
}

// 内联商品选择单元格组件
const ProductCell = defineComponent({
  props: { row: { type: Object, required: true }, options: { type: Array, default: () => [] }, loading: Boolean },
  emits: ['search', 'change'],
  setup(props, { emit }) {
    return () => h(ElSelect, {
      modelValue: (props.row as ItemRow).productId,
      'onUpdate:modelValue': (v: number) => { (props.row as ItemRow).productId = v },
      filterable: true, remote: true, loading: props.loading,
      remoteMethod: (kw: string) => emit('search', kw),
      placeholder: '商品名称/编码/规格', style: 'width: 100%',
      onFocus: () => emit('search', ''),
      onChange: (v: number) => emit('change', props.row, v),
    }, () => (props.options as any[]).map((p) => h(ElOption, {
      key: p.id, value: p.id,
      label: `${p.name}${p.specification ? ' / ' + p.specification : ''}`,
    })))
  },
})

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const billId = computed(() => (route.params.id && route.params.id !== 'new' ? Number(route.params.id) : null))
const editable = ref(true)
const loading = ref(false)
const saving = ref(false)

const form = reactive({
  id: 0,
  templateId: undefined as number | undefined,
  outWarehouseId: undefined as number | undefined,
  inWarehouseId: undefined as number | undefined,
  billDate: dayjs().format('YYYY-MM-DD'),
  fee: 0,
  handlerId: authStore.user?.id as number | undefined,
  remark: '',
  status: 'draft',
})
const items = ref<ItemRow[]>([])
const templates = ref<any[]>([])
const operatorName = computed(() => authStore.user?.name || authStore.user?.username || '-')
const productOptions = ref<any[]>([])
const productLoading = ref(false)

const outItems = computed(() => items.value.filter((i) => i.direction === 'out'))
const inItems = computed(() => items.value.filter((i) => i.direction === 'in'))
const outAmount = computed(() => outItems.value.reduce((s, i) => s + (i.quantity || 0) * (i.price || 0), 0).toFixed(2))
const inAmount = computed(() => inItems.value.reduce((s, i) => s + (i.quantity || 0) * (i.price || 0), 0).toFixed(2))

async function searchProducts(keyword: string) {
  productLoading.value = true
  try {
    const res = await api.get('/api/v1/products', { params: { keyword, pageSize: 50 } })
    if (res.data.code === 0 || res.data.code === 200) productOptions.value = res.data.data?.list ?? res.data.data?.items ?? []
  } finally { productLoading.value = false }
}

function handleProductChange(row: ItemRow, productId: number) {
  const p = productOptions.value.find((x) => x.id === productId)
  if (p) {
    row.productName = p.name; row.specification = p.specification; row.unit = p.unit
    if (!row.price) row.price = p.purchasePrice || 0
  }
}

function addRow(direction: 'out' | 'in') { items.value.push({ direction, quantity: 1, price: 0 }) }
function removeRow(direction: 'out' | 'in', index: number) {
  const targets = items.value.map((v, i) => ({ v, i })).filter(({ v }) => v.direction === direction)
  const target = targets[index]
  if (target) items.value.splice(target.i, 1)
}

async function loadTemplates() {
  const res = await api.get('/api/v1/assembly-templates', { params: { pageSize: 100 } })
  if (res.data.code === 0 || res.data.code === 200) templates.value = res.data.data?.list ?? []
}

function applyTemplate(templateId: number) {
  const tpl = templates.value.find((t) => t.id === templateId)
  if (!tpl) return
  try {
    const outs = JSON.parse(tpl.outItems || '[]') as any[]
    const ins = JSON.parse(tpl.inItems || '[]') as any[]
    items.value = [
      ...outs.map((i) => ({ direction: 'out' as const, productId: i.productId, productName: i.productName, quantity: i.quantity || 1, price: 0 })),
      ...ins.map((i) => ({ direction: 'in' as const, productId: i.productId, productName: i.productName, quantity: i.quantity || 1, price: 0 })),
    ]
  } catch { /* 忽略 */ }
}

async function loadDetail(id: number) {
  loading.value = true
  try {
    const res = await api.get(`/api/v1/assembly-orders/${id}`)
    if (res.data.code === 0 || res.data.code === 200) {
      const data = res.data.data
      Object.assign(form, {
        id: data.id, templateId: data.templateId || undefined,
        outWarehouseId: data.outWarehouseId, inWarehouseId: data.inWarehouseId,
        billDate: data.billDate ? dayjs(data.billDate).format('YYYY-MM-DD') : dayjs().format('YYYY-MM-DD'),
        fee: data.fee, handlerId: data.handlerId || undefined, remark: data.remark, status: data.status,
      })
      items.value = (data.items ?? []).map((it: any) => ({
        direction: it.direction, productId: it.productId, productName: it.product?.name,
        specification: it.product?.specification, unit: it.product?.unit, quantity: it.quantity, price: it.price,
      }))
      if (data.status !== 'draft') editable.value = false
    }
  } finally { loading.value = false }
}

async function save(complete: boolean) {
  if (!form.outWarehouseId || !form.inWarehouseId) { ElMessage.warning('请选择出库/入库仓库'); return }
  const validItems = items.value.filter((i) => i.productId)
  if (!validItems.length) { ElMessage.warning('请至少添加一行商品'); return }
  if (!validItems.some((i) => i.direction === 'out') || !validItems.some((i) => i.direction === 'in')) {
    ElMessage.warning('出库明细和入库明细都至少需要一行')
    return
  }
  saving.value = true
  try {
    const payload = {
      templateId: form.templateId || 0,
      outWarehouseId: form.outWarehouseId, inWarehouseId: form.inWarehouseId,
      billDate: form.billDate, fee: form.fee, handlerId: form.handlerId, remark: form.remark,
      items: validItems.map((i) => ({ direction: i.direction, productId: i.productId!, quantity: i.quantity, price: i.price || 0 })),
    }
    let id = billId.value
    const res = id ? await api.put(`/api/v1/assembly-orders/${id}`, payload) : await api.post('/api/v1/assembly-orders', payload)
    if (res.data.code === 0 || res.data.code === 200) {
      if (!id) id = res.data.data?.id
      if (complete && id) await api.put(`/api/v1/assembly-orders/${id}/complete`)
      ElMessage.success(complete ? '已过账' : '已存入草稿')
      router.back()
    }
  } finally { saving.value = false }
}

function goBack() { router.back() }

onMounted(() => {
  loadTemplates()
  if (billId.value) loadDetail(billId.value)
  else { addRow('out'); addRow('in') }
})
</script>

<style scoped>
.page { display: flex; flex-direction: column; height: 100%; gap: 12px; }
.page-header { display: flex; align-items: center; gap: 12px; }
.page-title { font-size: 18px; font-weight: 600; margin: 0; }
.form-panel { background: var(--el-bg-color); border-radius: 8px; padding: 16px 16px 0; }
.items-panel { background: var(--el-bg-color); border-radius: 8px; padding: 16px; flex: 1; min-height: 0; overflow: auto; }
.items-toolbar { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px; }
.items-title { font-weight: 600; }
.summary { margin-top: 8px; text-align: right; font-weight: bold; }
.summary span { margin-left: 24px; }
.footer-bar { display: flex; justify-content: space-between; align-items: center; background: var(--el-bg-color); border-radius: 8px; padding: 12px 16px; }
.operator { color: var(--el-text-color-secondary); }
.footer-actions { display: flex; gap: 8px; }
</style>
