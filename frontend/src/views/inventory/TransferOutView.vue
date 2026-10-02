<template>
  <TransferBillList
    title="调拨出库单"
    api-url="/api/v1/transfer-outs"
    create-route-name="InventoryTransferOutCreate"
    detail-route-name="InventoryTransferOutDetail"
  />
</template>

<script setup lang="ts">
import TransferBillList from './TransferBillList.vue'
</script>
