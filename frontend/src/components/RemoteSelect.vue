<template>
  <el-select
    v-model="innerValue"
    :placeholder="placeholder"
    :disabled="disabled"
    :multiple="multiple"
    :collapse-tags="multiple"
    filterable
    remote
    :remote-method="handleSearch"
    :loading="loading"
    clearable
    style="width: 100%"
  >
    <el-option
      v-for="item in options"
      :key="item.value"
      :label="item.label"
      :value="item.value"
    />
  </el-select>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import api from '@/api/client'

type SelectValue = number | string

interface Props {
  modelValue: SelectValue | SelectValue[] | undefined
  apiUrl: string
  labelKey?: string
  valueKey?: string
  placeholder?: string
  disabled?: boolean
  multiple?: boolean
  params?: Record<string, any>
}

const props = withDefaults(defineProps<Props>(), {
  labelKey: 'name',
  valueKey: 'id',
  placeholder: '请选择',
  disabled: false,
  multiple: false,
})

const emit = defineEmits<{
  (e: 'update:modelValue', val: SelectValue | SelectValue[] | undefined): void
}>()

const innerValue = ref<SelectValue | SelectValue[] | undefined>(
  props.multiple ? (Array.isArray(props.modelValue) ? props.modelValue : []) : props.modelValue
)
const options = ref<{ label: string; value: SelectValue }[]>([])
const loading = ref(false)

watch(() => props.modelValue, (val) => {
  innerValue.value = val
})

watch(innerValue, (val) => {
  emit('update:modelValue', val)
})

async function handleSearch(query: string) {
  loading.value = true
  try {
    const res = await api.get(props.apiUrl, {
      params: { keyword: query, pageSize: 50, ...props.params },
    })
    if (res.data.code === 0 || res.data.code === 200) {
      const data = res.data.data
      const rows = Array.isArray(data) ? data : (data?.list ?? [])
      options.value = (Array.isArray(rows) ? rows : []).map((item: any) => ({
        label: item[props.labelKey],
        value: item[props.valueKey],
      }))
    }
  } catch {
    options.value = []
  } finally {
    loading.value = false
  }
}

// Initial load
handleSearch('')
</script>
