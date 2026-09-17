import { beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('@/api/login', () => ({
  login: vi.fn(),
  logout: vi.fn(),
  refreshToken: vi.fn(),
  wxLogin: vi.fn(),
  getWxCode: vi.fn(),
  getUserInfo: vi.fn(),
  wecomMpLogin: vi.fn(),
}))

import { getUserInfo, wecomMpLogin } from '@/api/login'
import { useTokenStore } from './token'

const tokenRes = {
  accessToken: 'at',
  refreshToken: 'rt',
  accessExpiresIn: 7200,
  refreshExpiresIn: 604800,
}

describe('wecomMpLogin', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('登录成功：写入 tokenInfo 并拉取用户信息', async () => {
    vi.mocked(wecomMpLogin).mockResolvedValue(tokenRes)
    vi.mocked(getUserInfo).mockResolvedValue({ userId: 1, username: 'zhangsan', nickname: '张三' } as any)

    const store = useTokenStore()
    await store.wecomMpLogin('phone-code')

    expect(wecomMpLogin).toHaveBeenCalledWith('phone-code')
    expect(store.tokenInfo).toMatchObject({ accessToken: 'at', refreshToken: 'rt' })
    expect(uni.showToast).toHaveBeenCalledWith(expect.objectContaining({ title: '登录成功' }))
  })

  it('非企业成员 4102：原样抛出由页面处理引导，不弹成功 toast', async () => {
    vi.mocked(wecomMpLogin).mockRejectedValue({ code: 4102, message: '请先使用企业微信扫码加入企业', data: { inviteQrUrl: 'https://x/qr.png' } })

    const store = useTokenStore()
    await expect(store.wecomMpLogin('phone-code')).rejects.toMatchObject({ code: 4102 })
    expect(store.tokenInfo).toMatchObject({ accessToken: '' })
    expect(uni.showToast).not.toHaveBeenCalledWith(expect.objectContaining({ title: '登录成功' }))
  })
})
