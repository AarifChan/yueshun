<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">打印设置</h2>
    </div>
    <el-card>
      <el-tabs v-model="tab">
        <el-tab-pane label="打印设置" name="setting">
          <SettingsForm v-if="tab === 'setting'" :fields="printFields" />
        </el-tab-pane>
        <el-tab-pane label="打印机设置" name="printer">
          <div v-if="tab === 'printer'">
            <div class="toolbar">
              <el-button type="primary" icon="Plus" @click="openForm()">添加打印机</el-button>
              <span class="hint">支持远程打印、云盒子打印、云打印机三种接入方式。</span>
            </div>
            <el-table :data="printers" v-loading="loading" border stripe>
              <el-table-column type="index" label="序" width="60" />
              <el-table-column prop="name" label="打印机名称" min-width="150" />
              <el-table-column label="类型" width="120" align="center">
                <template #default="{ row }">{{ typeLabel(row.type) }}</template>
              </el-table-column>
              <el-table-column prop="deviceNo" label="设备编号" min-width="140">
                <template #default="{ row }">{{ row.deviceNo || '-' }}</template>
              </el-table-column>
              <el-table-column label="默认" width="80" align="center">
                <template #default="{ row }">
                  <el-tag v-if="row.isDefault" type="success" size="small">默认</el-tag>
                  <span v-else>-</span>
                </template>
              </el-table-column>
              <el-table-column label="状态" width="90" align="center">
                <template #default="{ row }">
                  <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">{{ row.status === 1 ? '启用' : '停用' }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="remark" label="备注" min-width="140">
                <template #default="{ row }">{{ row.remark || '-' }}</template>
              </el-table-column>
              <el-table-column label="操作" width="130" fixed="right">
                <template #default="{ row }">
                  <el-button link type="primary" @click="openForm(row)">编辑</el-button>
                  <el-button link type="danger" @click="remove(row)">删除</el-button>
                </template>
              </el-table-column>
              <template #empty>暂无打印机</template>
            </el-table>
          </div>
        </el-tab-pane>
        <el-tab-pane label="单据模板" name="billTpl">
          <el-alert v-if="tab === 'billTpl'" type="info" :closable="false"
            title="单据模板设置与模板中心：销售订单、销售出库单、采购订单、采购入库单等模板在线编辑能力建设中。" />
        </el-tab-pane>
        <el-tab-pane label="其他模板" name="otherTpl">
          <div v-if="tab === 'otherTpl'" class="tpl-list">
            <el-tag v-for="t in otherTemplates" :key="t" size="large" class="tpl-tag">{{ t }}</el-tag>
            <el-alert type="info" :closable="false" title="条码模板可在「商品 → 条码打印」中直接调整标签内容与尺寸。" style="margin-top: 12px" />
          </div>
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <el-dialog v-model="formVisible" :title="form.id ? '编辑打印机' : '添加打印机'" width="480px">
      <el-form label-width="100px">
        <el-form-item label="名称" required>
          <el-input v-model="form.name" placeholder="如 前台小票机" />
        </el-form-item>
        <el-form-item label="类型">
          <el-radio-group v-model="form.type">
            <el-radio value="remote">远程打印</el-radio>
            <el-radio value="cloud_box">云盒子</el-radio>
            <el-radio value="cloud">云打印机</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="设备编号">
          <el-input v-model="form.deviceNo" placeholder="打印机终端号/设备号" />
        </el-form-item>
        <el-form-item label="设为默认">
          <el-switch v-model="form.isDefault" />
        </el-form-item>
        <el-form-item label="启用">
          <el-switch v-model="formEnabled" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.remark" type="textarea" :rows="2" />
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
import SettingsForm, { type SettingField } from './SettingsForm.vue'

const tab = ref('setting')

const printFields: SettingField[] = [
  { key: 'print.mallOrderAutoPrint', label: '商城订单支付成功后自动打印', type: 'switch' },
  { key: 'print.maxPrintCount', label: '单据打印次数控制', type: 'number', min: 0, max: 99, default: '0', hint: '0 表示不限制' },
  { key: 'print.footerSeparatePage', label: '分页模式时单据表尾单独一页时需要打印', type: 'switch' },
  { key: 'print.afterPost', label: '过账后打印', type: 'switch', hint: '仅过账单据允许打印' },
  { key: 'print.hideDecimals', label: '金额为整数时隐藏小数位数', type: 'switch', default: '1' },
]

const otherTemplates = ['商品条码模板', '批次号模板', '贴包单模板', '拣货箱签模板', '销售订单设置', '销售出库单设置']

// ---- 打印机 CRUD ----
interface Printer { id: number; name: string; type: string; deviceNo: string; isDefault: boolean; status: number; remark: string }

const printers = ref<Printer[]>([])
const loading = ref(false)
const formVisible = ref(false)
const saving = ref(false)
const form = ref({ id: 0, name: '', type: 'remote', deviceNo: '', isDefault: false, remark: '' })
const formEnabled = ref(true)

function typeLabel(t: string) {
  return t === 'cloud_box' ? '云盒子打印' : t === 'cloud' ? '云打印机' : '远程打印'
}

async function load() {
  loading.value = true
  try {
    const res = await api.get('/api/v1/printers')
    if (res.data.code === 0 || res.data.code === 200) printers.value = res.data.data?.list ?? []
  } finally {
    loading.value = false
  }
}

function openForm(row?: Printer) {
  if (row) {
    form.value = { id: row.id, name: row.name, type: row.type, deviceNo: row.deviceNo, isDefault: row.isDefault, remark: row.remark }
    formEnabled.value = row.status === 1
  } else {
    form.value = { id: 0, name: '', type: 'remote', deviceNo: '', isDefault: false, remark: '' }
    formEnabled.value = true
  }
  formVisible.value = true
}

async function save() {
  if (!form.value.name.trim()) {
    ElMessage.warning('请填写打印机名称')
    return
  }
  saving.value = true
  try {
    const payload = { ...form.value, status: formEnabled.value ? 1 : 0 }
    const res = form.value.id
      ? await api.put(`/api/v1/printers/${form.value.id}`, payload)
      : await api.post('/api/v1/printers', payload)
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

async function remove(row: Printer) {
  await ElMessageBox.confirm(`确定删除打印机「${row.name}」？`, '提示', { type: 'warning' })
  const res = await api.delete(`/api/v1/printers/${row.id}`)
  if (res.data.code === 0 || res.data.code === 200) {
    ElMessage.success('已删除')
    load()
  }
}

onMounted(load)
</script>

<style scoped>
.page { display: flex; flex-direction: column; gap: 12px; }
.page-header { display: flex; justify-content: space-between; align-items: center; }
.page-title { font-size: 18px; font-weight: 600; margin: 0; }
.toolbar { display: flex; align-items: center; gap: 12px; margin-bottom: 12px; }
.hint { color: #909399; font-size: 12px; }
.tpl-list { padding: 4px 0; }
.tpl-tag { margin: 0 8px 8px 0; }
</style>
