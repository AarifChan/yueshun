<template>
  <div class="callback-container">
    <el-card class="callback-card">
      <el-result v-if="status === 'loading'" icon="info" title="正在登录…" sub-title="企业微信授权校验中，请稍候" />
      <el-result v-else icon="error" title="登录失败" :sub-title="errorMsg">
        <template #extra>
          <el-image v-if="inviteQrUrl" :src="inviteQrUrl" fit="contain" style="width: 200px; height: 200px" />
          <p v-if="inviteQrUrl" class="qr-tip">请使用企业微信扫码加入企业后重试</p>
          <el-button type="primary" @click="$router.replace('/login')">返回登录页</el-button>
        </template>
      </el-result>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const status = ref<'loading' | 'error'>('loading')
const errorMsg = ref('')
const inviteQrUrl = ref('')

onMounted(async () => {
  const { code, state } = route.query as Record<string, string>
  if (!code) {
    status.value = 'error'
    errorMsg.value = '缺少授权码，请重新扫码'
    return
  }
  // state 防 CSRF：与登录页生成时存入的值比对
  const savedState = sessionStorage.getItem('wecom_login_state')
  sessionStorage.removeItem('wecom_login_state')
  if (!savedState || state !== savedState) {
    status.value = 'error'
    errorMsg.value = '登录状态校验失败，请重新扫码'
    return
  }
  try {
    await authStore.wecomLogin(code)
    router.replace('/dashboard')
  } catch (e: any) {
    status.value = 'error'
    errorMsg.value = e.message || '企业微信登录失败'
    if (e.code === 4102) {
      inviteQrUrl.value = (e.data as { inviteQrUrl?: string })?.inviteQrUrl || ''
    }
  }
})
</script>

<style scoped>
.callback-container { min-height: 100vh; display: flex; justify-content: center; align-items: center; background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); }
.callback-card { width: 460px; }
.qr-tip { margin: 12px 0; font-size: 14px; color: #606266; }
</style>
