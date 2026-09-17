# 企业微信登录配置指南

本文档指导运维完成企业微信（WeCom）侧的全部配置，使智账系统的企业微信登录（Web 扫码 + 微信小程序手机号）可用。所有配置均通过 `ZHIZHANG_` 前缀环境变量注入，无需修改代码或配置文件。

## 1. 创建自建应用

1. 登录[企业微信管理后台](https://work.weixin.qq.com/wework_admin/frame)。
2. 进入 **应用管理 → 应用 → 自建**，点击 **创建应用**，填写名称（如「智账系统」）与 logo。
3. 创建后进入应用详情页，记录：
   - **AgentId**（如 `1000002`）
   - **Secret**：点击「查看 Secret」，会推送到管理员的企业微信会话中查看。
4. **可见范围**：选择全体成员，或按需勾选部门（不在可见范围内的成员无法扫码登录）。

## 2. 通讯录权限（手机号查 userid 必需）

小程序端登录会用成员手机号调用 `cgi-bin/user/getuserid` 反查企业成员 userid，该接口需要**通讯录级别权限**：

- 方式一：**管理工具 → 通讯录同步** → 开启「API 接口同步」，使用通讯录同步 Secret 作为 `ZHIZHANG_WECOM_SECRET`；
- 方式二：给自建应用配置「通讯录读取」权限后使用应用 Secret。

若用自建应用 Secret 调用报 `60011`（无权限）或 `48002`（API 接口被禁用），请改用通讯录同步 Secret 或给应用补充授权。

## 3. 可信域名与 WWLogin（Web 扫码必需）

1. 自建应用详情 → **开发者接口 → 企业微信授权登录** → **设置授权回调域**，填写 Web 管理端的域名（不含协议与路径，如 `erp.example.com`）。
2. 该域名必须与 `ZHIZHANG_WECOM_REDIRECT_HOST` 一致，否则扫码后回调会被企业微信拒绝。
3. 本地开发（localhost）无法完成扫码回调，请使用超级管理员密码通道登录。

## 4. 邀请二维码

非企业成员登录时会收到 `4102` 响应并展示引导二维码：

1. 管理后台 → **我的企业 → 邀请成员** → 下载邀请二维码图片；
2. 将图片上传至任意可公开访问的图床 / OSS；
3. 图片 URL 填入 `ZHIZHANG_WECOM_INVITE_QR_URL`。

## 5. 微信小程序 getPhoneNumber

小程序端一键登录依赖「手机号快速验证组件」：

1. 登录[小程序管理后台](https://mp.weixin.qq.com) → **开发管理 → 接口设置 → 手机号快速验证组件**，开通该能力（要求非个人主体且已微信认证）。
2. 小程序 **AppID / AppSecret** 分别填入 `ZHIZHANG_WECHAT_MP_APPID` / `ZHIZHANG_WECHAT_MP_SECRET`。
3. AppID 同步填入 `app/env/.env` 的 `VITE_WX_APPID`（小程序端构建用）。

## 6. 环境变量清单

对照仓库根目录 `.env.example`：

| 变量 | 说明 |
| --- | --- |
| `ZHIZHANG_WECOM_CORPID` | 企业 ID（我的企业 → 企业信息） |
| `ZHIZHANG_WECOM_AGENTID` | 自建应用 AgentId |
| `ZHIZHANG_WECOM_SECRET` | 自建应用 Secret（或通讯录同步 Secret，见第 2 节） |
| `ZHIZHANG_WECOM_INVITE_QR_URL` | 邀请成员二维码图片 URL |
| `ZHIZHANG_WECOM_REDIRECT_HOST` | Web 扫码回调域名（须为可信域名） |
| `ZHIZHANG_WECHAT_MP_APPID` | 微信小程序 AppID |
| `ZHIZHANG_WECHAT_MP_SECRET` | 微信小程序 AppSecret |

## 7. 验证步骤

1. 后端启动日志**无** `wecom login not configured` / `wechat mp not configured` 告警；
2. `curl http://<host>:8388/api/v1/auth/wecom/config` 返回 `code: 200` 且含 `corpid`；
3. Web 登录页切换到「企业微信登录」出现二维码，扫码后可进入系统；
4. 小程序端点击「企业微信一键登录」，授权手机号后登录成功；
5. 未加入企业的手机号登录返回引导页（4102），未建档成员提示联系管理员（4103）。
