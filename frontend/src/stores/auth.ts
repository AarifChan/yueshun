import { defineStore } from 'pinia'
import api from '@/api/client'

export interface UserInfo {
  id: number
  username: string
  name: string
  phone?: string
  roleId?: number
  deptId?: number
}

interface LoginData {
  accessToken: string
  refreshToken?: string
  user: UserInfo
}

function loginError(message: string, res: { data: { code?: number; data?: unknown } }) {
  const err = new Error(message) as Error & { code?: number; data?: unknown }
  err.code = res.data.code
  err.data = res.data.data
  return err
}

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: localStorage.getItem('token') || '',
    user: JSON.parse(localStorage.getItem('user') || 'null') as UserInfo | null,
  }),
  getters: {
    isLoggedIn: (state) => !!state.token,
  },
  actions: {
    async login(username: string, password: string) {
      const res = await api.post('/api/v1/auth/login', { username, password })
      if (res.data.code === 0 || res.data.code === 200) {
        this.setLogin(res.data.data)
        return res.data.data
      }
      throw loginError(res.data.message || '登录失败', res)
    },
    async wecomLogin(code: string) {
      const res = await api.post('/api/v1/auth/wecom/web', { code })
      if (res.data.code === 0 || res.data.code === 200) {
        this.setLogin(res.data.data)
        return res.data.data
      }
      throw loginError(res.data.message || '企业微信登录失败', res)
    },
    setLogin(data: LoginData) {
      this.token = data.accessToken
      this.user = data.user
      localStorage.setItem('token', data.accessToken)
      if (data.refreshToken) {
        localStorage.setItem('refreshToken', data.refreshToken)
      }
      localStorage.setItem('user', JSON.stringify(data.user))
    },
    async fetchUser() {
      const res = await api.get('/api/v1/auth/me')
      if (res.data.code === 0 || res.data.code === 200) {
        this.user = res.data.data
        localStorage.setItem('user', JSON.stringify(res.data.data))
        return res.data.data
      }
      throw loginError(res.data.message || '获取用户信息失败', res)
    },
    async logout() {
      try {
        await api.post('/api/v1/auth/logout')
      } catch {
        // 忽略登出接口错误
      }
      this.token = ''
      this.user = null
      localStorage.removeItem('token')
      localStorage.removeItem('refreshToken')
      localStorage.removeItem('user')
    },
  },
})
