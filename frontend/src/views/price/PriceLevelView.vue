<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">商品价格体系</h2>
      <div class="actions">
        <el-button type="primary" @click="openCreate">新增级别价</el-button>
      </div>
    </div>
    <el-card>
      <el-tabs model-value="level">
        <el-tab-pane label="级别价（必设）" name="level" />
      </el-tabs>
      <el-table :data="list" v-loading="loading" border stripe>
        <el-table-column prop="name" label="级别价名称" min-width="150" />
        <el-table-column prop="required" label="是否必填" width="100">
          <template #default="{ row }"><el-tag :type="row.required ? 'success' : 'info'">{{ row.required ? '是' : '否' }}</el-tag></template>
        </el-table-column>
        <el-table-column prop="rule" label="级别价规则" min-width="160"><template #default="{ row }">{{ row.rule || '-' }}</template></el-table-column>
        <el-table-column prop="customerScope" label="适用客户" min-width="160"><template #default="{ row }">{{ row.customerScope || '-' }}</template></el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="openEdit(row)">编辑</el-button>
            <el-button type="danger" size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
        <template #empty>暂无数据</template>
      </el-table>
      <el-pagination v-model:current-page="pagination.page" v-model:page-size="pagination.pageSize" :total="total" :page-sizes="[30,50,100]" layout="total, sizes, prev, pager, next" @size-change="handleSizeChange" @current-change="handleCurrentChange" class="pagination" />
    </el-card>
    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="500px">
      <el-form :model="form" label-width="90px">
        <el-form-item label="级别价名称" required><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="编码"><el-input v-model="form.code" /></el-form-item>
        <el-form-item label="是否必填"><el-switch v-model="form.required" /></el-form-item>
        <el-form-item label="级别价规则"><el-input v-model="form.rule" placeholder="如：零售价的 95%" /></el-form-item>
        <el-form-item label="适用客户"><el-input v-model="form.customerScope" placeholder="如：全部客户 / 批发客户" /></el-form-item>
        <el-form-item label="默认"><el-switch v-model="form.isDefault" /></el-form-item>
        <el-form-item label="状态"><el-switch v-model="form.status" :active-value="1" :inactive-value="0" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { useCrud } from '@/composables/useCrud'

interface PriceLevel {
  id: number
  name: string
  code: string
  required: boolean
  rule: string
  customerScope: string
  isDefault: boolean
  status: number
}

const crud = useCrud<PriceLevel>({ baseUrl: '/api/v1/prices/levels', defaultForm: () => ({ status: 1, isDefault: false, required: false, rule: '', customerScope: '' }) })
const { list, total, loading, dialogVisible, dialogTitle, form, pagination, fetchList, openCreate, openEdit, handleSubmit, handleDelete, handleSizeChange, handleCurrentChange } = crud
pagination.value.pageSize = 30
fetchList()
</script>

<style scoped>
.page { padding: 20px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.page-title { margin: 0; font-size: 18px; font-weight: 600; }
.actions { display: flex; gap: 8px; }
.pagination { margin-top: 16px; justify-content: flex-end; }
</style>
