<template>
  <div class="login-container">
    <el-card class="login-card" shadow="always">
      <template #header>
        <div class="login-header">
          <el-icon :size="40" color="#409EFF"><Box /></el-icon>
          <h2>智账系统</h2>
          <p>企业进销存管理平台</p>
        </div>
      </template>

      <el-tabs v-model="activeTab" stretch>
        <el-tab-pane label="企业微信扫码" name="wecom">
          <div v-if="wecomLoading" class="qr-loading" v-loading="true" style="height: 300px" />
          <div v-show="!wecomLoading && wecomConfigured" id="wecom-qr-container" class="qr-box" />
          <el-result v-if="!wecomLoading && !wecomConfigured" icon="warning" title="企业微信登录未配置" sub-title="请联系管理员配置，或切换账号密码登录（仅超级管理员）" />
          <p v-if="wecomConfigured" class="qr-tip">使用企业微信扫码登录，未加入企业请先联系管理员</p>
        </el-tab-pane>

        <el-tab-pane label="账号密码（仅超级管理员）" name="password">
          <el-form
            ref="formRef"
            :model="form"
            :rules="rules"
            label-position="top"
            @keyup.enter="handleLogin"
          >
            <el-form-item label="用户名" prop="username">
              <el-input v-model="form.username" placeholder="请输入用户名" prefix-icon="User" size="large" />
            </el-form-item>
            <el-form-item label="密码" prop="password">
              <el-input v-model="form.password" type="password" placeholder="请输入密码" prefix-icon="Lock" size="large" show-password />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" size="large" style="width: 100%" :loading="loading" @click="handleLogin">
                登录
              </el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import api from '@/api/client'
import { ElMessage } from 'element-plus'

declare global {
  interface Window { WwLogin?: new (options: Record<string, string>) => unknown }
}

const WWLOGIN_SDK_URL = 'https://wwcdn.weixin.qq.com/node/wework/wwopen/js/wwLogin-1.2.7.js'

const router = useRouter()
const authStore = useAuthStore()
const formRef = ref()
const loading = ref(false)
const activeTab = ref<'wecom' | 'password'>('wecom')
const wecomLoading = ref(true)
const wecomConfigured = ref(false)

const form = reactive({ username: '', password: '' })

const rules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
}

function loadScript(src: string) {
  return new Promise<void>((resolve, reject) => {
    const el = document.createElement('script')
    el.src = src
    el.onload = () => resolve()
    el.onerror = () => reject(new Error('企业微信组件加载失败'))
    document.head.appendChild(el)
  })
}

async function mountWecomQr() {
  try {
    const res = await api.get('/api/v1/auth/wecom/config')
    if (res.data.code !== 200 && res.data.code !== 0) {
      return // 未配置：保持 wecomConfigured = false
    }
    const { corpid, agentId, redirectHost } = res.data.data
    const state = crypto.randomUUID()
    sessionStorage.setItem('wecom_login_state', state)
    await loadScript(WWLOGIN_SDK_URL)
    const origin = redirectHost || window.location.origin
    new window.WwLogin!({
      id: 'wecom-qr-container',
      appid: corpid,
      agentid: String(agentId),
      redirect_uri: encodeURIComponent(`${origin}/login/wecom/callback`),
      state,
      href: '',
    })
    wecomConfigured.value = true
  } catch {
    // 网络异常等按未配置处理，用户可用密码通道
  } finally {
    wecomLoading.value = false
  }
}

onMounted(mountWecomQr)

async function handleLogin() {
  try {
    await formRef.value.validate()
    loading.value = true
    await authStore.login(form.username, form.password)
    ElMessage.success('登录成功')
    router.push('/dashboard')
  } catch (error: any) {
    // 4101：非超管密码登录被拒，引导扫码
    if (error?.code === 4101) {
      ElMessage.warning('请使用企业微信扫码登录')
      activeTab.value = 'wecom'
    } else {
      ElMessage.error(error.message || '登录失败')
    }
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-container { min-height: 100vh; display: flex; justify-content: center; align-items: center; background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); }
.login-card { width: 460px; }
.login-header { text-align: center; }
.login-header h2 { margin: 10px 0 5px; font-size: 24px; color: #303133; }
.login-header p { margin: 0; font-size: 14px; color: #909399; }
.qr-box { display: flex; justify-content: center; min-height: 300px; }
.qr-tip { text-align: center; font-size: 12px; color: #909399; margin: 12px 0 0; }
</style>
