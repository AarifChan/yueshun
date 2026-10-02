# AGENTS.md — 项目架构与开发指南

> 面向 AI 编码助手与新加入开发者的项目说明。阅读本文即可了解整体架构、目录约定与开发工作流。

## 一、项目概览

**yueshun（模块名 `zhizhang-server`）** 是一套进销存 + CRM 管理系统，前后端分离：

| 端 | 技术栈 | 目录 | 默认端口 |
|---|---|---|---|
| 后端 | Go 1.26 · Gin · GORM · PostgreSQL · Redis · JWT | `cmd/` `internal/` | 8388 |
| 前端 | Vue 3 · TypeScript · Vite · Element Plus · Pinia | `frontend/` | 3000（代理 `/api`、`/uploads` → 8388） |

业务域：基础资料（部门/职员/角色/权限/字典）、商品资料（分类/品牌/规格/单位/标签/销售范围）、价格体系、客户/供应商、采购（订单/入库/退货/付款）、销售（订单/出库/退货/收款）、库存（盘点/调拨/预警/库存状况表/其他入库单/其他出库单）、仓库资金账户、商城（商品/购物车/订单）、CRM（商机/跟进/合同）、审批流、统计报表、企业微信扫码登录。

## 二、目录结构

```
├── cmd/server/main.go        # 入口：加载配置 → 日志 → DB → 自动迁移 → 种子数据 → 启动路由
├── configs/config.yaml       # 应用配置（viper 加载：app/jwt/database/redis/wecom/upload/log）
├── internal/
│   ├── api/
│   │   ├── router.go         # 路由总装：实例化所有 Handler，注册中间件与 /api/v1 路由组
│   │   ├── handler/          # 业务处理器（~25 个文件，按领域划分，见下）
│   │   └── middleware/       # cors.go / jwt.go / log.go
│   ├── model/base.go         # 全部 GORM 模型（~69 个 struct，单文件集中定义）
│   └── pkg/
│       ├── database/         # GORM 初始化 + 库存相关共用逻辑（stock.go）
│       ├── redis/            # Redis 客户端初始化
│       ├── response/         # 统一响应结构（Response / PageResult / 错误码常量）
│       └── errors/           # 业务错误定义（BizError、哨兵错误）
├── docs/                     # swaggo 生成的 Swagger 文档 + superpowers/{plans,specs} 设计文档
├── frontend/src/
│   ├── api/                  # axios 封装（client.ts）+ 各模块 API 定义
│   ├── views/                # 页面，按业务域分目录（base/product/purchase/sale/inventory/...）
│   ├── layouts/MainLayout.vue
│   ├── router/index.ts       # 路由 + 登录守卫；菜单结构由路由 meta 驱动
│   ├── stores/auth.ts        # Pinia 认证状态
│   ├── composables/          # useCrud.ts（通用 CRUD 逻辑）、useBillDetail.ts（单据详情）
│   └── components/           # BillItemTable / RemoteSelect / SideTreePanel 等共享组件
├── scripts/start.sh          # 一键同时启动前后端
├── Makefile                  # build / dev / start / swagger / migrate / lint / test / pg-up
├── uploads/                  # 图片上传目录（后端静态服务 /uploads）
└── docs/superpowers/         # 需求规格与实施计划文档
```

## 三、后端架构约定

### 分层与调用链
```
main.go → api.SetupRouter(db, jwtCfg, wecomCfg, wxmpCfg, uploadDir)
        → handler.XxxHandler{db} → RegisterRoutes(authorized)  // 直接在 Handler 内用 GORM 操作 DB
```
无独立 service/repository 层 —— **业务逻辑写在 Handler 中**，GORM 直接查询 `internal/model` 的模型。新增功能请遵循同一模式，不要引入新分层。

### 关键约定
- **模型**：全部定义在 `internal/model/base.go`。`BaseModel`（ID/时间戳）与 `BaseModelWithCompany`（多公司隔离字段）两个基类。
- **统一响应**：必须经 `internal/pkg/response` 返回。成功 `Code=200`，业务错误码段 `4000-4999`；分页用 `response.PageResult(list, total, page, pageSize)`（空列表序列化为 `[]` 而非 `null`）。
- **认证**：JWT（access 2h / refresh 7d）。公开路由仅认证接口；其余全部挂在 `authorized` 组（`middleware.JWTMiddleware()`）。
- **路由注册**：每个 Handler 实现 `RegisterRoutes(r *gin.RouterGroup)`，在 `router.go` 中集中挂载。
- **数据库**：启动时 GORM AutoMigrate 自动迁移 + `handler.InitSeedData` 种子数据。改模型即迁移，无独立迁移文件。
- **API 文档**：swaggo 注解，`make swagger` 重新生成到 `docs/`，运行时访问 `/swagger/index.html`。
- **企业微信/小程序登录**：`handler/auth_wecom.go`（WecomAuthHandler）+ `internal/pkg/wecom`、`internal/pkg/wxmp` 客户端，由 `configs/config.yaml` 的 `wecom`/`wechat_mp` 段控制（留空即未配置，密码登录仅超级管理员可用，其余角色返回 4101 引导扫码）。

## 四、前端架构约定

- **HTTP 层**：`src/api/client.ts` 统一 axios 实例 —— 自动注入 `Bearer token`、401 时清 token 跳登录、统一 `ElMessage` 报错。**新 API 模块必须基于该 client**。
- **CRUD 页面**：优先复用 `src/composables/useCrud.ts`（列表/分页/搜索/弹窗表单/状态操作全套）；单据详情页用 `useBillDetail.ts`。响应兼容 `code === 0 || code === 200`，数据取 `data.list ?? data.items`。
- **组件**：Element Plus 按需自动注册（unplugin），`@element-plus/icons-vue` 图标；API 亦自动导入（vue/vue-router/pinia）。
- **路由即菜单**：`router/index.ts` 的 `meta.title/icon/mega` 驱动 `MainLayout` 导航（含 mega 菜单分组）。新增页面需在此注册。
- **路径别名**：`@` → `src/`。
- **测试**：vitest + @vue/test-utils（如 `useCrud.spec.ts`），`npm run test`。
- **构建**：`npm run build` 含 vue-tsc 类型检查前请确保通过。

## 五、开发工作流

```bash
# 环境准备（Docker 起 PG；Redis 需自备）
make pg-up

# 日常开发：一键同时启动前后端（推荐）
make start            # 等价 scripts/start.sh；make start-build 先编译后端
# 或分开：
make dev              # 后端（有 air 则热重载）
cd frontend && npm run dev

# 常用
make swagger          # 改 API 注解后重新生成文档
make lint / make test # golangci-lint / go test -race
cd frontend && npm run build && npm run test
```

默认账号等种子数据见 `handler.InitSeedData`。后端 `:8388`，前端 `:3000`，Swagger 在 `http://localhost:8388/swagger/index.html`。

## 六、修改代码时的注意事项

1. **保持单文件模型风格**：新表结构加到 `internal/model/base.go`，重启即自动迁移；注意 `BaseModelWithCompany` 的多公司字段。
2. **响应格式不要自创**：前后端已约定 `{code, message, data}` 及分页结构，前端 `useCrud` 依赖它。
3. **改完 API 记得 `make swagger`**。
4. **前端新页面**：注册路由（带 `meta.title/icon`）→ 视图放对应业务域目录 → API 函数放 `src/api/`。
5. **设计文档**：较大功能的需求/计划归档在 `docs/superpowers/{specs,plans}`（如企业微信登录方案）。
