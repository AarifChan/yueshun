<template>
  <TransferBillList
    title="调拨入库单"
    api-url="/api/v1/transfer-ins"
    create-route-name="InventoryTransferInCreate"
    detail-route-name="InventoryTransferInDetail"
  />
</template>

<script setup lang="ts">
import TransferBillList from './TransferBillList.vue'
</script>
