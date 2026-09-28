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

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: localStorage.getItem('token') || '',
    user: JSON.parse(localStorage.getItem('user') || 'null') as UserInfo | null,
  }),
  actions: {
    async login(username: string, password: string) {
      const res = await api.post('/api/v1/auth/login', { username, password })
      this.setLogin(res.data.data)
    },
    setLogin(data: { accessToken: string; refreshToken?: string; user: UserInfo }) {
      this.token = data.accessToken
      this.user = data.user
      localStorage.setItem('token', data.accessToken)
      localStorage.setItem('user', JSON.stringify(data.user))
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
      localStorage.removeItem('user')
    },
  },
})
