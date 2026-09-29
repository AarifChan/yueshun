<template>
  <div class="side-tree-panel" v-loading="loading">
    <div class="panel-header">
      <span class="panel-title" @click="emit('title-click')">
        <el-icon class="panel-toggle" :class="{ collapsed: treeCollapsed }" @click.stop="treeCollapsed = !treeCollapsed"><ArrowDown /></el-icon>
        <span>{{ title }}</span>
      </span>
      <el-link v-if="editTo || hasEditClick" type="primary" :underline="false" @click="handleEditClick">编辑</el-link>
    </div>
    <el-input
      v-model="keyword"
      :placeholder="placeholder"
      clearable
      :prefix-icon="Search"
      class="panel-search"
      @keyup.enter="applyFilter"
      @clear="applyFilter"
    >
      <template #append>
        <el-button @click="applyFilter">搜 索</el-button>
      </template>
    </el-input>
    <el-tree
      v-show="!treeCollapsed"
      ref="treeRef"
      :data="data"
      :props="{ label: 'name', children: 'children' }"
      node-key="id"
      highlight-current
      :expand-on-click-node="false"
      :filter-node-method="filterNode"
      class="panel-tree"
      @node-click="(data: any, node: any) => emit('node-click', data, node)"
    >
      <template #default="slotProps">
        <slot v-bind="slotProps">
          <span class="tree-node">
            <el-icon><Folder /></el-icon>
            <span>{{ slotProps.data.name }}</span>
          </span>
        </slot>
      </template>
    </el-tree>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, useAttrs, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ArrowDown, Folder, Search } from '@element-plus/icons-vue'

const props = withDefaults(defineProps<{
  title: string
  data: any[]
  editTo?: string
  placeholder?: string
  loading?: boolean
}>(), {
  placeholder: '请输入分类…',
})

const emit = defineEmits<{
  (e: 'node-click', data: any, node: any): void
  (e: 'edit-click'): void
  (e: 'title-click'): void
}>()

const router = useRouter()
const attrs = useAttrs()
const hasEditClick = computed(() => !!attrs.onEditClick)

const treeRef = ref()
const keyword = ref('')
const treeCollapsed = ref(false)

watch(keyword, (val) => {
  treeRef.value?.filter(val)
})

function applyFilter() {
  treeRef.value?.filter(keyword.value)
}

function filterNode(value: string, data: any) {
  if (!value) return true
  return String(data.name || '').toLowerCase().includes(value.toLowerCase())
}

function handleEditClick() {
  if (props.editTo) {
    router.push(props.editTo)
  } else {
    emit('edit-click')
  }
}

defineExpose({
  filter: (val: string) => treeRef.value?.filter(val),
  setCurrentKey: (key: string | number | null) => treeRef.value?.setCurrentKey(key),
  clearCurrentKey: () => treeRef.value?.setCurrentKey(null),
})
</script>

<style scoped>
.side-tree-panel { display: flex; flex-direction: column; flex: 1; min-height: 0; width: 100%; }
.panel-header { display: flex; justify-content: space-between; align-items: center; font-weight: 600; margin-bottom: 8px; }
.panel-title { display: inline-flex; align-items: center; gap: 4px; cursor: pointer; }
.panel-toggle { transition: transform 0.2s ease; color: var(--el-text-color-secondary); }
.panel-toggle.collapsed { transform: rotate(-90deg); }
.panel-search { margin-bottom: 8px; }
.panel-tree { flex: 1; min-height: 0; overflow-y: auto; }
.tree-node { display: inline-flex; align-items: center; gap: 4px; }
</style>
