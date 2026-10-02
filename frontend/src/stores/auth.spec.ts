import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'

vi.mock('@/api/client', () => ({
  default: { post: vi.fn(), get: vi.fn() },
}))

import api from '@/api/client'
import { useAuthStore } from './auth'

describe('wecomLogin', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()
    vi.clearAllMocks()
  })

  it('扫码登录成功：存储 token 与用户信息', async () => {
    vi.mocked(api.post).mockResolvedValue({
      data: {
        code: 200,
        message: 'success',
        data: {
          accessToken: 'at',
          refreshToken: 'rt',
          accessExpiresIn: 7200,
          refreshExpiresIn: 604800,
          user: { id: 1, username: 'zhangsan', name: '张三', phone: '13800000001', roleId: 2, deptId: 1 },
        },
      },
    })
    const store = useAuthStore()
    await store.wecomLogin('auth-code')
    expect(api.post).toHaveBeenCalledWith('/api/v1/auth/wecom/web', { code: 'auth-code' })
    expect(store.token).toBe('at')
    expect(localStorage.getItem('token')).toBe('at')
    expect(store.user?.name).toBe('张三')
  })

  it('未建档：抛出带 4103 的错误', async () => {
    vi.mocked(api.post).mockResolvedValue({
      data: { code: 4103, message: '请联系管理员在后台添加员工档案' },
    })
    const store = useAuthStore()
    await expect(store.wecomLogin('auth-code')).rejects.toMatchObject({ code: 4103 })
    expect(store.token).toBe('')
  })
})

describe('login', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()
    vi.clearAllMocks()
  })

  it('非超管密码登录被拒：抛出带 4101 的错误', async () => {
    vi.mocked(api.post).mockResolvedValue({
      data: { code: 4101, message: '当前账号请使用企业微信登录' },
    })
    const store = useAuthStore()
    await expect(store.login('staff', '123456')).rejects.toMatchObject({ code: 4101 })
    expect(store.token).toBe('')
  })
})
