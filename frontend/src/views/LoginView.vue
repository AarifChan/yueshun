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
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-position="top"
        @keyup.enter="handleLogin"
      >
        <el-form-item label="用户名" prop="username">
          <el-input
            v-model="form.username"
            placeholder="请输入用户名"
            prefix-icon="User"
            size="large"
          />
        </el-form-item>
        <el-form-item label="密码" prop="password">
          <el-input
            v-model="form.password"
            type="password"
            placeholder="请输入密码"
            prefix-icon="Lock"
            size="large"
            show-password
          />
        </el-form-item>
        <el-form-item>
          <el-button
            type="primary"
            size="large"
            style="width: 100%"
            :loading="loading"
            @click="handleLogin"
          >
            登录
          </el-button>
        </el-form-item>
      </el-form>

      <el-divider class="login-divider"><span class="divider-text">其他登录方式</span></el-divider>
      <div class="third-party">
        <div class="wecom-btn" @click="openWeCom">
          <el-icon :size="22"><ChatDotSquare /></el-icon>
          <span>企业微信登录</span>
        </div>
      </div>

      <el-dialog
        v-model="wecomDialogVisible"
        title="企业微信扫码登录"
        width="420px"
        align-center
        @close="wecomQrUrl = ''"
      >
        <div v-if="wecomQrUrl" class="wecom-qr-wrap">
          <iframe :src="wecomQrUrl" class="wecom-qr-frame" frameborder="0" />
        </div>
        <div v-else class="wecom-loading">
          <el-icon class="is-loading" :size="24"><Loading /></el-icon>
          <p>正在加载二维码…</p>
        </div>
      </el-dialog>

      <div class="login-tips">
        <p>默认账号: admin / admin123</p>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import api from '@/api/client'
import { ElMessage } from 'element-plus'

const router = useRouter()
const authStore = useAuthStore()
const formRef = ref()
const loading = ref(false)

const wecomDialogVisible = ref(false)
const wecomQrUrl = ref('')

const form = reactive({
  username: '',
  password: '',
})

const rules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
}

async function handleLogin() {
  try {
    await formRef.value.validate()
    loading.value = true
    await authStore.login(form.username, form.password)
    ElMessage.success('登录成功')
    router.push('/dashboard')
  } catch (error: any) {
    ElMessage.error(error.message || '登录失败')
  } finally {
    loading.value = false
  }
}

async function openWeCom() {
  try {
    const status = await api.get('/api/v1/auth/wecom/status')
    if (!status.data.data?.enabled) {
      ElMessage.warning('企业微信登录未配置，请联系管理员')
      return
    }
    wecomDialogVisible.value = true
    wecomQrUrl.value = ''
    const res = await api.get('/api/v1/auth/wecom/qrcode-url')
    wecomQrUrl.value = res.data.data?.url || ''
  } catch (error: any) {
    ElMessage.error(error.response?.data?.message || '企业微信登录暂不可用')
  }
}

async function handleWeComMessage(event: MessageEvent) {
  if (event.origin !== window.location.origin) return
  const data = event.data
  if (!data || data.type !== 'wecom-login') return
  if (!data.ok) {
    ElMessage.error(data.message || '企业微信登录失败')
    wecomDialogVisible.value = false
    return
  }
  authStore.setLogin({ accessToken: data.accessToken, user: { id: 0, username: '', name: data.user || '' } })
  ElMessage.success('企业微信登录成功')
  wecomDialogVisible.value = false
  try {
    const me = await api.get('/api/v1/auth/me')
    if (me.data.data) authStore.setLogin({ accessToken: data.accessToken, user: me.data.data })
  } catch {
    // 拉取用户信息失败时保留已有信息
  }
  router.push('/dashboard')
}

onMounted(() => {
  window.addEventListener('message', handleWeComMessage)
})

onUnmounted(() => {
  window.removeEventListener('message', handleWeComMessage)
})
</script>

<style scoped>
.login-container {
  min-height: 100vh;
  display: flex;
  justify-content: center;
  align-items: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}
.login-card {
  width: 420px;
}
.login-header {
  text-align: center;
}
.login-header h2 {
  margin: 10px 0 5px;
  font-size: 24px;
  color: #303133;
}
.login-header p {
  margin: 0;
  font-size: 14px;
  color: #909399;
}
.login-tips {
  text-align: center;
  margin-top: 16px;
}
.login-tips p {
  font-size: 12px;
  color: #909399;
}
.login-divider {
  margin: 8px 0 16px;
}
.divider-text {
  font-size: 12px;
  color: #909399;
}
.third-party {
  display: flex;
  justify-content: center;
}
.wecom-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 20px;
  border: 1px solid #07c160;
  border-radius: 6px;
  color: #07c160;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s;
}
.wecom-btn:hover {
  background: #07c160;
  color: #fff;
}
.wecom-qr-wrap {
  display: flex;
  justify-content: center;
}
.wecom-qr-frame {
  width: 350px;
  height: 400px;
}
.wecom-loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 60px 0;
  color: #909399;
}
</style>
