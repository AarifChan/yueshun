import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import api from '@/api/client'

export interface UserInfo {
  id: number
  username: string
  name: string
  phone: string
  roleId: number
  deptId: number
}

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string>(localStorage.getItem('token') || '')
  const user = ref<UserInfo | null>(null)
  const isLoggedIn = computed(() => !!token.value)

  async function login(username: string, password: string) {
    const res = await api.post('/api/v1/auth/login', { username, password })
    if (res.data.code === 0 || res.data.code === 200) {
      const data = res.data.data
      token.value = data.accessToken
      localStorage.setItem('token', data.accessToken)
      localStorage.setItem('refreshToken', data.refreshToken)
      user.value = data.user
      return data
    }
    const err = new Error(res.data.message || '登录失败') as Error & { code?: number }
    err.code = res.data.code
    throw err
  }

  async function wecomLogin(code: string) {
    const res = await api.post('/api/v1/auth/wecom/web', { code })
    if (res.data.code === 0 || res.data.code === 200) {
      const data = res.data.data
      token.value = data.accessToken
      localStorage.setItem('token', data.accessToken)
      localStorage.setItem('refreshToken', data.refreshToken)
      user.value = data.user
      return data
    }
    const err = new Error(res.data.message || '企业微信登录失败') as Error & { code?: number; data?: unknown }
    err.code = res.data.code
    err.data = res.data.data
    throw err
  }

  async function fetchUser() {
    const res = await api.get('/api/v1/auth/me')
    if (res.data.code === 0 || res.data.code === 200) {
      user.value = res.data.data
      return res.data.data
    }
    throw new Error(res.data.message || '获取用户信息失败')
  }

  function logout() {
    token.value = ''
    user.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('refreshToken')
  }

  return {
    token,
    user,
    isLoggedIn,
    login,
    wecomLogin,
    fetchUser,
    logout,
  }
})
