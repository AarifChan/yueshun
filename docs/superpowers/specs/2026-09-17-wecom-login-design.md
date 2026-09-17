# 企业微信登录接入设计

- 日期：2026-09-17
- 状态：已评审（五节设计逐节确认）
- 范围：Go 后端、Web 管理端（frontend/）、小程序端（app/，uni-best）

## 1. 背景与目标

智账系统要求：除超级管理员外，所有内部用户必须通过企业微信登录；加入企业的方式是扫描企业邀请二维码（企业微信原生能力）。系统在用户登录时通过企业微信 API **强制验证成员身份**，未加入企业者拒绝登录。

### 已确认决策

| # | 决策点 | 结论 |
|---|--------|------|
| 1 | 强制范围 | 仅内部业务端（Web 管理端 + 小程序内部功能）；商城客户不受影响 |
| 2 | 小程序形态 | 纯微信小程序，员工授权手机号（getPhoneNumber），后端调企业微信 API 按手机号验证成员身份 |
| 3 | 员工建档 | 管理员后台预建档；首次登录按手机号自动绑定；未建档拒绝 |
| 4 | 超管例外 | `role.code = super_admin` 保留账号密码登录，其余角色密码登录一律拒绝 |
| 5 | 凭证现状 | 有企业（corpid），无自建应用 —— 交付物含应用创建指引文档 |

## 2. 总体架构（方案 A：服务端企业微信集成）

后端新增两个接口化小包，统一处理第三方 API 调用与 access_token 缓存；认证层新增三个端点；三端登录页改造。JWT 体系（双 token、`{code, message, data}` 信封、多租户 companyID 注入）完全不变。

```
┌─────────────┐  getPhoneNumber   ┌──────────────┐
│  微信小程序  │ ────────────────→ │              │
└─────────────┘   code            │              │
                                  │  zhizhang    │      ┌──────────────┐
┌─────────────┐  WWLogin 扫码     │   server     │      │  企业微信 API │
│ Web 管理端   │ ────────────────→ │              │ ───→ │  (qyapi)     │
└─────────────┘   auth code       │              │      └──────────────┘
                                  │              │      ┌──────────────┐
                                  │              │ ───→ │ 微信公众平台  │
                                  └──────────────┘      │  (换手机号)   │
                                                        └──────────────┘
```

## 3. 后端设计

### 3.1 新增包

```
internal/pkg/wecom/     # 企业微信 API 客户端
  client.go             # Client 接口 + HTTP 实现
  token.go              # access_token 缓存（Redis key: wecom:access_token，提前 300s 过期，singleflight 防击穿）
  types.go              # UserDetail、ErrNotMember / ErrInvalidCode 错误分类
internal/pkg/wxmp/      # 微信小程序 API 客户端
  client.go             # GetPhoneNumber(code) → purePhoneNumber
  token.go              # access_token 缓存（Redis key: wxmp:access_token，同样提前 300s）
```

**`wecom.Client` 接口**（handler 测试用 mock 注入）：

```go
type Client interface {
    // GetUserIDByMobile 按手机号查企业成员 userid
    // POST https://qyapi.weixin.qq.com/cgi-bin/user/getuserid
    GetUserIDByMobile(mobile string) (userID string, err error) // 非成员 → ErrNotMember
    // GetUserInfoByCode WWLogin 授权码换成员身份
    // GET https://qyapi.weixin.qq.com/cgi-bin/auth/getuserinfo
    GetUserInfoByCode(code string) (userID string, err error)   // 外部联系人 → ErrNotMember
    // GetUserDetail 按 userid 取成员详情（含 mobile、name）
    // GET https://qyapi.weixin.qq.com/cgi-bin/user/get
    GetUserDetail(userID string) (*UserDetail, error)
}
```

**`wxmp.Client` 接口**：

```go
type Client interface {
    // GetPhoneNumber 小程序手机号授权码换手机号
    // POST https://api.weixin.qq.com/wxa/business/getuserphonenumber
    GetPhoneNumber(code string) (purePhone string, err error)   // code 一次性、5 分钟有效
}
```

### 3.2 新增端点（`internal/api/handler/auth_wecom.go`）

| 端点 | 方法 | 鉴权 | 说明 |
|------|------|------|------|
| `/api/v1/auth/wecom/config` | GET | 公开 | 返回 `{corpid, agentId, inviteQrUrl, redirectHost}`，供登录页构造二维码与引导页 |
| `/api/v1/auth/wecom/mp` | POST | 公开 | body `{code}`：小程序登录 |
| `/api/v1/auth/wecom/web` | POST | 公开 | body `{code}`：Web 扫码登录 |

**小程序登录流程**：

```
code → wxmp.GetPhoneNumber → phone
     → wecom.GetUserIDByMobile(phone)
        ├─ ErrNotMember → 4102（data.inviteQrUrl）
        └─ userid
     → 绑定判定（见 §4.2）→ issueTokenPair(emp) → LoginResp
```

**Web 扫码登录流程**：

```
code → wecom.GetUserInfoByCode
        ├─ ErrNotMember（外部用户返回 openid 而非 userid）→ 4102
        └─ userid → wecom.GetUserDetail(userid) → mobile
     → 绑定判定（见 §4.2）→ issueTokenPair(emp) → LoginResp
```

### 3.3 现有代码改动

- `auth.go`：提取 `issueTokenPair(emp) (LoginResp, error)`，供密码登录 / 刷新 / 小程序 / Web 扫码四处复用（本仓库已实现 Refresh，改动是纯提取，不改行为）。
- `auth.go` Login 拦截：查到员工后联查 Role，`role.Code != "super_admin"` → 4101 拒绝。
- `employee.go`：create/update 增加手机号必填 + 公司内重复校验（`CodeDuplicate`）；update 支持 `wecomUserId` 显式传空串解绑。
- `main.go`：Config 增加 `Wecom`/`WechatMP` 两节；构造 wecom/wxmp 客户端并注入 auth handler；凭证为空时 warn 日志但不阻断启动。

## 4. 数据模型与绑定规则

### 4.1 Employee 模型改动

```go
Username    string  // 不变（必填唯一，仅作档案标识，企业微信登录不经过它）
Password    string  `gorm:"size:128"`                                     // 去掉 not null
Phone       string  `gorm:"size:20;uniqueIndex:idx_emp_phone_company"`   // 公司内唯一索引
WecomUserID string  `json:"wecomUserId" gorm:"size:64;index"`            // 新增
```

- AutoMigrate 自动处理：新增列、`password` DROP NOT NULL、唯一索引创建。现有数据仅 seed admin 一条，无冲突。
- 唯一索引前提：管理员建档时手机号必填（employee 接口同步加校验）。
- 实现注意：复合唯一索引 `(company_id, phone)` **不能用** GORM 同名 tag 方式实现——`CompanyID` 在共享的 `BaseModelWithCompany` 里，打 tag 会给所有租户模型都加索引。改为在 `autoMigrate()` 完成后执行原生 SQL：`CREATE UNIQUE INDEX IF NOT EXISTS idx_emp_phone_company ON employees (company_id, phone)`（PostgreSQL/SQLite 均兼容该语法，测试可用 SQLite 内存库）。

### 4.2 绑定判定（小程序与 Web 共用）

```
输入：wecom_userid + mobile
1. WHERE wecom_user_id = ? AND company_id = ? → 命中
   → 若 emp.Phone ≠ mobile，以企业微信为准更新 Phone → 判 status → 放行
2. 未命中 → WHERE phone = ? AND company_id = ? → 命中
   → 回写 WecomUserID（首次绑定）→ 判 status → 放行
3. 都未命中 → 4103「请联系管理员添加员工档案」
status != 1 → 403「账号已禁用，请联系管理员」
```

- **身份锚点是 `wecom_user_id`**，手机号仅作首次绑定桥梁；成员在企业微信改手机号不影响登录。
- **解绑**：管理员在员工编辑中将 `wecomUserId` 置空 → 员工下次登录重新走手机号绑定（离职重招/换号场景）。
- 查员工时先不带 status 条件，命中后单独判 status —— 保证「禁用」与「未建档」文案分开。

### 4.3 新增业务码（`internal/pkg/response`）

| 码 | 常量 | 场景 |
|---|------|------|
| 4101 | `CodeNeedWecom` | 非超管尝试密码登录 |
| 4102 | `CodeNotWecomMember` | 非企业成员（data 带 inviteQrUrl） |
| 4103 | `CodeEmpNotRegistered` | 是企业成员但系统未建档 |

## 5. 三端页面改造

### 5.1 Web 管理端（frontend/）

- `LoginView.vue` 改双 Tab：
  - Tab 1（默认）企业微信扫码：`GET /auth/wecom/config` 拿 corpid/agentId → 动态加载官方 WWLogin SDK（`wwLogin-1.2.7.js`）渲染二维码 iframe → 扫码授权后 redirect 到回调页。
  - Tab 2 账号密码登录：现有表单保留，标注「仅超级管理员」；收到 4101 展示对应文案。
- 新增 `WecomCallbackView.vue` + 路由 `/login/wecom/callback`：取 URL `code`/`state` → 校验 state（防 CSRF，见 §7）→ `POST /auth/wecom/web` → 存 token 跳 dashboard；4102/4103 分支展示对应文案 + 返回登录页链接。
- `stores/auth.ts` 加 `wecomLogin(code)` action，token 存储逻辑复用现有。
- 员工管理页：手机号必填（表单校验 + 说明文案「加入企业微信后首次登录自动绑定」）；列表加「企微绑定」列（已绑定显示成员 ID + 解绑按钮）。

### 5.2 小程序端（app/）

- `pages/auth/login.vue` 重写（替换 mock）：
  - 主按钮 `<button open-type="getPhoneNumber">` → code → `tokenStore.wecomMpLogin(code)` → 成功 navigateBack。
  - 次级入口：账号密码真实表单（超管/本地开发）。
  - 4102 → 引导页展示 `inviteQrUrl` 二维码图片，文案「请使用企业微信扫码加入企业后重试」；4103 → 提示联系管理员。
- `store/token.ts` 加 `wecomMpLogin(code)`，成功走现有 `_postLogin`（含 fetchUserInfo）。
- 登录拦截（`router/config.ts`）：商城相关页面路径加入 `EXCLUDE_LOGIN_PATH_LIST`（免登录白名单）；`LOGIN_PAGE_ENABLE_IN_MP` 改 `true` 启用小程序拦截。内部业务页保持强制登录。
- 商城客户登录（手机号/微信授权）不在本期范围。

## 6. 配置项

```yaml
# configs/config.yaml 新增
wecom:
  corpid: ""            # 企业 ID
  agentid: 0            # 自建应用 AgentID
  secret: ""            # 自建应用 Secret（仅服务端）
  invite_qr_url: ""     # 企业邀请二维码图片地址
  redirect_host: ""     # Web 扫码回调域名（须为应用可信域名）
wechat_mp:
  appid: ""             # 微信小程序 AppID
  secret: ""            # 微信小程序 Secret（仅服务端）
```

- 沿用 `ZHIZHANG_` 环境变量前缀覆盖（如 `ZHIZHANG_WECOM_CORPID`）。
- **配置以环境变量为准**：`config.yaml` 仅保留占位空值，`.env.example` 必须补齐全部新增变量及说明；部署时通过环境变量注入，不把真实凭证写进仓库。
- 启动校验：为空 → warn 日志，相关端点返回 4001「企业微信登录未配置」，不阻断启动。
- Secret 不出服务端；公开端点只暴露 corpid / agentId / inviteQrUrl。

## 7. 安全设计

- WWLogin 回调 `state` 防 CSRF：前端生成随机 state 存 sessionStorage，回调校验一致后才提交 code。
- access_token 缓存于 Redis（两个客户端各自提前 300s 刷新）；Redis 不可用时**跳过缓存直接调第三方 API**——`main.go` 里 Redis 本就是可选非阻塞依赖，登录链路不应比系统其他部分更脆弱；不引入进程内缓存复杂度。
- 小程序 code 一次性、5 分钟有效，后端原样透传微信侧错误为 400「授权已过期，请重试」。
- 服务端错误日志带上下文（zerolog），前端只展示业务文案，不泄露内部错误细节。
- 已知缺口（列为后续加固，不在本期）：登录端点频率限制（现有系统全端点均无限流器，需单独立项）。

## 8. 错误处理矩阵

| 场景 | code | 前端表现 |
|------|------|----------|
| 企微/微信 API 网络或 token 失效 | 500 | toast「服务异常，请稍后重试」 |
| 小程序 code 无效/过期 | 400 | toast「授权已过期，请重试」 |
| 非企业成员 | 4102 | 引导页 + inviteQrUrl |
| 企业成员但未建档 | 4103 | 「请联系管理员添加员工档案」 |
| 建档但禁用 | 403 | 「账号已禁用，请联系管理员」 |
| 非超管密码登录 | 4101 | 「请使用企业微信登录」 |
| 未配置企微/小程序凭证 | 4001 | 「企业微信登录未配置」 |

## 9. 测试策略

- **后端**（table-driven + `-race`，新增代码覆盖率 ≥ 80%）：
  - wecom/wxmp 客户端：`httptest.Server` 模拟第三方 API —— token 缓存命中/过期刷新、错误分类、超时。
  - handler 集成测试（mock Client 注入）：小程序登录六分支（已绑定/首绑/非成员/未建档/禁用/手机号变更更新）、Web 登录三分支（userid 命中 / fallback 手机号绑定 / 外部用户拒绝）、密码登录拦截（超管放行 / 普通角色 4101）、手机号重复建档 4002。
- **Web 管理端**：`wecomLogin` action 单测（Vitest mock axios）；回调页组件测试（code 缺失 / state 不匹配 / 4102 / 4103）。
- **小程序端**：`wecomMpLogin` 单测（mock http 层）；`router/interceptor.test.ts` 扩展（商城白名单放行 / 内部页强制登录）。
- **手动验收**：真机 `getPhoneNumber`、WWLogin 真实扫码（无法自动化，列入验收清单）。
- TDD：每个端点先写失败测试再实现（RED → GREEN → REFACTOR）。

## 10. 交付物清单

1. `internal/pkg/wecom/`、`internal/pkg/wxmp/` 两个新包（含测试）
2. `internal/api/handler/auth_wecom.go` + `auth.go`/`employee.go` 改动（含测试）
3. Employee 模型迁移（wecom_user_id 列、password 可空、phone 唯一索引）
4. Web 管理端：登录页双 Tab + 回调页 + auth store + 员工管理页改动
5. 小程序端：登录页重写 + token store + 拦截配置
6. `configs/config.yaml` 新增配置节 + `.env.example` 更新
7. `docs/wecom-setup.md`：自建应用创建、通讯录权限、可信域名、邀请二维码生成、getPhoneNumber 开通条件

## 11. 明确不做（本期）

- 商城客户登录体系
- 登录端点频率限制（后续加固项）
- 真实扫码的 E2E 自动化（只能手动验收）
- 企业微信通讯录双向同步（仅登录时按手机号查成员，不同步组织架构）
