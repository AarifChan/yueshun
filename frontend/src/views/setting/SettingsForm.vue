<template>
  <el-form label-width="240px" class="settings-form" v-loading="loading">
    <el-form-item v-for="f in fields" :key="f.key" :label="f.label">
      <div class="field-row">
        <el-switch v-if="f.type === 'switch'" v-model="boolValues[f.key]" />
        <el-input-number
          v-else-if="f.type === 'number'"
          v-model="numValues[f.key]"
          :min="f.min ?? 0"
          :max="f.max ?? 999999"
          :precision="f.precision ?? 0"
        />
        <el-select v-else-if="f.type === 'select'" v-model="strValues[f.key]" style="width: 280px">
          <el-option v-for="o in f.options || []" :key="o.value" :label="o.label" :value="o.value" />
        </el-select>
        <el-input
          v-else
          v-model="strValues[f.key]"
          style="width: 360px"
          :placeholder="f.placeholder"
          clearable
        />
        <span v-if="f.hint" class="hint">{{ f.hint }}</span>
      </div>
    </el-form-item>
    <el-form-item>
      <el-button type="primary" :loading="saving" @click="save">保存</el-button>
    </el-form-item>
  </el-form>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import api from '@/api/client'

export interface SettingField {
  key: string
  label: string
  type: 'switch' | 'number' | 'select' | 'text'
  options?: { label: string; value: string }[]
  hint?: string
  placeholder?: string
  min?: number
  max?: number
  precision?: number
  default?: string
}

const props = defineProps<{ fields: SettingField[] }>()

const loading = ref(false)
const saving = ref(false)
const boolValues = reactive<Record<string, boolean>>({})
const numValues = reactive<Record<string, number>>({})
const strValues = reactive<Record<string, string>>({})

async function load() {
  loading.value = true
  try {
    const res = await api.get('/api/v1/settings')
    const s: Record<string, string> = (res.data.code === 0 || res.data.code === 200) ? res.data.data || {} : {}
    for (const f of props.fields) {
      const v = s[f.key] ?? f.default ?? ''
      if (f.type === 'switch') boolValues[f.key] = v === '1'
      else if (f.type === 'number') numValues[f.key] = Number(v || 0)
      else strValues[f.key] = v
    }
  } finally {
    loading.value = false
  }
}

async function save() {
  saving.value = true
  try {
    const payload: Record<string, string> = {}
    for (const f of props.fields) {
      if (f.type === 'switch') payload[f.key] = boolValues[f.key] ? '1' : '0'
      else if (f.type === 'number') payload[f.key] = String(numValues[f.key] ?? 0)
      else payload[f.key] = strValues[f.key] ?? ''
    }
    const res = await api.put('/api/v1/settings', payload)
    if (res.data.code === 0 || res.data.code === 200) ElMessage.success('已保存')
    else ElMessage.error(res.data.message || '保存失败')
  } finally {
    saving.value = false
  }
}

onMounted(load)
</script>

<style scoped>
.field-row { display: flex; align-items: center; gap: 12px; flex-wrap: wrap; }
.hint { color: #909399; font-size: 12px; }
</style>
