import type { IAuthLoginRes, ICaptcha, IDoubleTokenRes, IUpdateInfo, IUpdatePassword, IUserInfoRes } from './types/login'
import { http } from '@/http/http'

export { WECOM_CODE } from './types/login'

/**
 * 登录表单
 */
export interface ILoginForm {
  username: string
  password: string
}

/**
 * 获取验证码
 * @returns ICaptcha 验证码
 */
export function getCode() {
  return http.get<ICaptcha>('/user/getCode')
}

/**
 * 用户登录
 * @param loginForm 登录表单
 */
export function login(loginForm: ILoginForm) {
  return http.post<IAuthLoginRes>('/auth/login', loginForm)
}

/**
 * 刷新token
 * @param refreshToken 刷新token
 */
export function refreshToken(refreshToken: string) {
  return http.post<IDoubleTokenRes>('/auth/refresh', { refreshToken })
}

/**
 * 后端 /auth/me 返回的用户信息结构
 */
interface IBackendUserInfo {
  id: number
  username: string
  name: string
  phone: string
  roleId: number
  deptId: number
}

/**
 * 获取用户信息
 * 适配智账系统后端 /auth/me：id/name → userId/nickname
 */
export async function getUserInfo(): Promise<IUserInfoRes> {
  const res = await http.get<IBackendUserInfo>('/auth/me')
  return {
    ...res,
    userId: res.id,
    username: res.username,
    nickname: res.name,
  }
}

/**
 * 退出登录
 */
export function logout() {
  return http.post<void>('/auth/logout')
}

/**
 * 修改用户信息
 */
export function updateInfo(data: IUpdateInfo) {
  return http.post('/user/updateInfo', data)
}

/**
 * 修改用户密码
 */
export function updateUserPassword(data: IUpdatePassword) {
  return http.post('/user/updatePassword', data)
}

/**
 * 获取微信登录凭证
 * @returns Promise 包含微信登录凭证(code)
 */
export function getWxCode() {
  return new Promise<UniApp.LoginRes>((resolve, reject) => {
    uni.login({
      provider: 'weixin',
      success: res => resolve(res),
      fail: err => reject(new Error(err)),
    })
  })
}

/**
 * 微信登录
 * @param params 微信登录参数，包含code
 * @returns Promise 包含登录结果
 */
export function wxLogin(data: { code: string }) {
  return http.post<IAuthLoginRes>('/auth/wxLogin', data)
}

/**
 * 小程序企业微信登录（手机号授权码）
 * hideErrorToast: 4102/4103 等分支由登录页自行处理 UI（引导二维码等），不走全局 toast
 */
export function wecomMpLogin(code: string) {
  return http.post<IAuthLoginRes>('/auth/wecom/mp', { code }, undefined, undefined, { hideErrorToast: true })
}

export interface IWecomConfig {
  corpid: string
  agentId: number
  inviteQrUrl: string
  redirectHost: string
}

/**
 * 企业微信登录前端配置（corpid/agentId/inviteQrUrl）
 */
export function getWecomConfig() {
  return http.get<IWecomConfig>('/auth/wecom/config')
}
