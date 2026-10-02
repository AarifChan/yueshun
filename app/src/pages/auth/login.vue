<script lang="ts" setup>
import { onLoad } from '@dcloudio/uni-app'
import { ref } from 'vue'
import { getWecomConfig, WECOM_CODE } from '@/api/login'
import { useTokenStore } from '@/store/token'
import { HOME_PAGE } from '@/utils'

definePage({
  style: {
    navigationBarTitleText: '登录',
  },
})

const tokenStore = useTokenStore()
const redirectUrl = ref('')

const username = ref('')
const password = ref('')
const loading = ref(false)

// 4102 引导态
const showInviteGuide = ref(false)
const inviteQrUrl = ref('')

onLoad((options) => {
  if (options?.redirect) {
    redirectUrl.value = decodeURIComponent(options.redirect)
  }
})

function navigateAfterLogin() {
  const target = redirectUrl.value || HOME_PAGE
  uni.redirectTo({ url: target })
}

async function doWecomLogin(e: any) {
  const code = e?.detail?.code
  if (!code) {
    uni.showToast({ title: '手机号授权失败，请重试', icon: 'none' })
    return
  }
  loading.value = true
  try {
    await tokenStore.wecomMpLogin(code)
    navigateAfterLogin()
  }
  catch (err: any) {
    if (err?.code === WECOM_CODE.NOT_MEMBER) {
      inviteQrUrl.value = err?.data?.inviteQrUrl || ''
      if (!inviteQrUrl.value) {
        // 后端未配邀请二维码时尝试拉一次公开配置
        try {
          inviteQrUrl.value = (await getWecomConfig()).inviteQrUrl
        }
        catch { /* 配置也没有就只显示文案 */ }
      }
      showInviteGuide.value = true
    }
    else if (err?.code === WECOM_CODE.NOT_REGISTERED) {
      uni.showToast({ title: '请联系管理员在后台添加员工档案', icon: 'none' })
    }
    else {
      uni.showToast({ title: err?.message || '登录失败，请重试', icon: 'none' })
    }
  }
  finally {
    loading.value = false
  }
}

async function doPasswordLogin() {
  if (!username.value || !password.value) {
    uni.showToast({ title: '请输入用户名和密码', icon: 'none' })
    return
  }
  loading.value = true
  try {
    await tokenStore.login({ username: username.value, password: password.value })
    navigateAfterLogin()
  }
  catch (err: any) {
    if (err?.code === WECOM_CODE.NEED_WECOM) {
      uni.showToast({ title: '请使用企业微信登录', icon: 'none' })
    }
    // 其余错误 http 层已弹 toast
  }
  finally {
    loading.value = false
  }
}
</script>

<template>
  <view class="login-page px-48rpx pt-100rpx">
    <view class="mb-80rpx text-center">
      <view class="text-44rpx font-bold">
        智账系统
      </view>
      <view class="mt-16rpx text-26rpx text-gray-400">
        企业进销存管理平台
      </view>
    </view>

    <!-- 企业微信一键登录：仅微信小程序支持 getPhoneNumber -->
    <!-- #ifdef MP-WEIXIN -->
    <button
      class="mb-32rpx h-88rpx w-full rounded-12rpx bg-green-600 text-32rpx text-white leading-88rpx"
      open-type="getPhoneNumber"
      :disabled="loading"
      @getphonenumber="doWecomLogin"
    >
      企业微信一键登录
    </button>
    <view class="mb-40rpx text-center text-24rpx text-gray-400">
      将校验您的企业微信成员身份（需已加入企业）
    </view>
    <!-- #endif -->

    <!-- 账号密码登录：仅超级管理员 -->
    <view class="mb-24rpx text-center text-24rpx text-gray-400">
      —— 账号密码登录（仅超级管理员） ——
    </view>
    <input
      v-model="username"
      class="mb-24rpx h-88rpx w-full rounded-12rpx bg-gray-100 px-24rpx text-30rpx"
      placeholder="用户名"
    >
    <input
      v-model="password"
      class="mb-32rpx h-88rpx w-full rounded-12rpx bg-gray-100 px-24rpx text-30rpx"
      password
      placeholder="密码"
    >
    <button
      class="h-88rpx w-full rounded-12rpx bg-blue-600 text-32rpx text-white leading-88rpx"
      :disabled="loading"
      @click="doPasswordLogin"
    >
      登录
    </button>

    <!-- 非企业成员引导：展示企业邀请二维码 -->
    <view v-if="showInviteGuide" class="mt-60rpx flex flex-col items-center">
      <view class="mb-16rpx text-28rpx text-gray-600">
        请先使用企业微信扫码加入企业
      </view>
      <image
        v-if="inviteQrUrl"
        :src="inviteQrUrl"
        class="h-400rpx w-400rpx"
        mode="aspectFit"
        show-menu-by-longpress
      />
      <view class="mt-16rpx text-24rpx text-gray-400">
        可长按保存二维码，在企业微信中扫码加入后重试
      </view>
    </view>
  </view>
</template>

<style lang="scss" scoped>
// UnoCSS 原子类已覆盖布局，此处保留块备用
</style>
