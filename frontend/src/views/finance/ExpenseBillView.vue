<template>
  <FinanceBillView title="费用单" api-url="/api/v1/expense-bills" item-type="expense" item-label="费用项目" />
</template>

<script setup lang="ts">
import FinanceBillView from './FinanceBillView.vue'
</script>
