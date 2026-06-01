<template>
  <div class="page" v-loading="loading">
    <el-page-header @back="goBack" :title="pageTitle" />
    <el-card class="detail-card">
      <el-form :model="form" label-width="100px" :disabled="!isEditable">
        <el-row :gutter="20">
          <el-col :span="8"><el-form-item label="单据编号"><el-input v-model="form.billNo" disabled /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="付款日期" required><el-date-picker v-model="form.billDate" type="date" value-format="YYYY-MM-DD" style="width:100%" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="供应商" required><RemoteSelect v-model="form.supplierId" api-url="/api/v1/customers" :params="{ type: 'supplier' }" /></el-form-item></el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="8"><el-form-item label="付款账户"><RemoteSelect v-model="form.accountId" api-url="/api/v1/accounts" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="总金额"><el-input-number v-model="form.totalAmount" :min="0" :precision="2" style="width:100%" /></el-form-item></el-col>
        </el-row>
        <el-form-item label="备注"><el-input v-model="form.remark" type="textarea" /></el-form-item>
      </el-form>
    </el-card>
    <el-card class="detail-card" v-if="!isEditable && form.status === 'draft'">
      <div class="status-actions"><el-button type="success" @click="doStatusAction({ api: '/api/v1/purchase-payments/:id/complete' })">完成付款</el-button></div>
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

const route = useRoute()
const id = route.params.id as string
const mode = (route.query.mode as string) || 'view'

const { form, loading, saving, isEditable, initCreate, loadDetail, save, doStatusAction, goBack, setMode } = useBillDetail({
  baseUrl: '/api/v1/purchase-payments',
  defaultForm: () => ({ billDate: '', supplierId: undefined, accountId: undefined, totalAmount: 0, remark: '' }),
  defaultItem: () => ({}),
})

const pageTitle = computed(() => id === 'new' ? '新建采购付款' : mode === 'edit' ? '编辑采购付款' : '采购付款详情')
if (id === 'new') { initCreate() } else { loadDetail(id); if (mode === 'edit') setMode('edit') }
</script>

<style scoped>
.page { padding: 20px; }
.detail-card { margin-top: 16px; }
.footer-actions { margin-top: 16px; text-align: center; }
.status-actions { text-align: center; }
</style>
