<template>
  <div class="page" v-loading="loading">
    <el-page-header @back="goBack" :title="pageTitle" />
    <el-card class="detail-card">
      <el-form :model="form" label-width="100px" :disabled="!isEditable">
        <el-row :gutter="20">
          <el-col :span="8"><el-form-item label="单据编号"><el-input v-model="form.billNo" disabled /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="出库日期" required><el-date-picker v-model="form.billDate" type="date" value-format="YYYY-MM-DD" style="width:100%" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="销售订单"><RemoteSelect v-model="form.orderId" api-url="/api/v1/sales-orders" /></el-form-item></el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="8"><el-form-item label="客户" required><RemoteSelect v-model="form.customerId" api-url="/api/v1/customers" :params="{ type: 'customer' }" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="仓库" required><RemoteSelect v-model="form.warehouseId" api-url="/api/v1/warehouses" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="折扣"><el-input-number v-model="form.discount" :min="0" :precision="2" style="width:100%" /></el-form-item></el-col>
        </el-row>
        <el-form-item label="备注"><el-input v-model="form.remark" type="textarea" /></el-form-item>
      </el-form>
    </el-card>
    <el-card class="detail-card"><BillItemTable :items="items" :editable="isEditable" @add="addItem" @remove="removeItem" /></el-card>
    <el-card class="detail-card" v-if="!isEditable && form.status === 'draft'">
      <div class="status-actions">
        <el-button type="success" @click="doStatusAction({ api: '/api/v1/sales-outstock/:id/complete' })">完成出库</el-button>
      </div>
    </el-card>
    <div class="footer-actions">
      <template v-if="isEditable"><el-button type="primary" @click="save" :loading="saving">保存</el-button></template>
      <el-button @click="goBack">返回</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useBillDetail } from '@/composables/useBillDetail'
import RemoteSelect from '@/components/RemoteSelect.vue'
import BillItemTable from '@/components/BillItemTable.vue'

const route = useRoute()
const id = route.params.id as string
const mode = (route.query.mode as string) || 'view'

const { form, items, loading, saving, isEditable, initCreate, loadDetail, addItem, removeItem, save, doStatusAction, goBack, setMode } = useBillDetail({
  baseUrl: '/api/v1/sales-outstock',
  defaultForm: () => ({ billDate: '', customerId: undefined, warehouseId: undefined, orderId: undefined, discount: 0, remark: '' }),
  defaultItem: () => ({ productId: undefined, quantity: 0, price: 0, remark: '' }),
})

const pageTitle = computed(() => id === 'new' ? '新建销售出库' : mode === 'edit' ? '编辑销售出库' : '销售出库详情')

if (id === 'new') { initCreate() } else { loadDetail(id); if (mode === 'edit') setMode('edit') }
</script>

<style scoped>
.page { padding: 20px; }
.detail-card { margin-top: 16px; }
.footer-actions { margin-top: 16px; text-align: center; }
.status-actions { text-align: center; }
</style>
