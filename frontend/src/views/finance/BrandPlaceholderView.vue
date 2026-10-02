<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">{{ title }}</h2>
    </div>
    <el-card>
      <el-form inline>
        <el-form-item label="客户名称"><el-input placeholder="客户名称" clearable disabled /></el-form-item>
        <el-form-item label="品牌"><el-input placeholder="品牌" clearable disabled /></el-form-item>
        <el-form-item>
          <el-button type="primary" disabled>查询</el-button>
        </el-form-item>
      </el-form>
      <el-empty description="暂无数据（品牌级预收/待结算业务未启用）" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
defineProps<{ title: string }>()
</script>

<style scoped>
.page { display: flex; flex-direction: column; gap: 12px; }
.page-header { display: flex; justify-content: space-between; align-items: center; }
.page-title { font-size: 18px; font-weight: 600; margin: 0; }
</style>
