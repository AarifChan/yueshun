<template>
  <div class="coming-soon">
    <el-empty :description="`「${title}」功能开发中，敬请期待`" />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()
const title = computed(() => (route.meta.title as string) || '该页面')
</script>

<style scoped>
.coming-soon {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  background: var(--el-bg-color);
  border-radius: 8px;
}
</style>
