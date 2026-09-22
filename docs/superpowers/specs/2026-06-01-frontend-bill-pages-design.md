# 前端单据页面功能开发设计文档

## 背景

后端已完整实现 15 个业务模块的 REST API（采购、销售、库存、审批、CRM、商城等），但前端视图大多为空壳（25-59 行），仅有基础表格和简单弹窗，缺少搜索筛选、分页、详情查看、从表明细编辑、状态流转等核心功能。

## 目标

按模块优先级逐个深度开发完整单据页面，使前端功能与后端 API 对齐，达到可投入使用的 ERP-lite 系统标准。

## 设计决策

### 开发顺序

按业务主线优先级，逐个模块深度开发：

1. **采购模块**（订单 / 入库 / 退货 / 付款）
2. **销售模块**（订单 / 出库 / 退货 / 收款）
3. **库存模块**（盘点 / 调拨 / 预警）
4. **审批模块**（流程 / 记录）

### 页面模板

两种标准页面模板：

#### 模板 A：单据列表页

```
搜索筛选区（可展开/折叠）
  ├── 单据编号输入框
  ├── 日期范围选择器
  ├── 状态下拉框
  ├── 关联对象选择器（供应商/客户/仓库等）
  └── 搜索/重置按钮

操作区
  └── 新增按钮

分页表格
  ├── 列：ID、单据编号、日期、关联对象、总金额、状态
  ├── 状态列：el-tag 彩色标签（草稿=info、已确认=success、已取消=danger、已完成=primary）
  └── 操作列：根据状态动态显示按钮
      ├── 草稿：查看 / 编辑 / 确认 / 取消 / 删除
      ├── 已确认：查看 / 完成
      ├── 已完成：查看
      └── 已取消：查看

底部分页器（el-pagination）
```

#### 模板 B：单据详情页

路由格式：`/模块名/单据类型/:id?mode=edit`

```
页面头部
  ├── 标题（新建/编辑/查看 采购订单）
  └── 返回按钮

主表信息区（el-card）
  ├── 单据编号（系统自动生成，只读）
  ├── 日期选择器
  ├── 关联对象选择器（供应商/客户/仓库，带搜索下拉）
  ├── 备注文本框
  └── 其他主表字段...

从表明细区（el-card）
  ├── 标题：商品明细 + 添加行按钮
  └── el-table 行内编辑
      ├── 商品选择（下拉搜索，显示商品名称+编码）
      ├── 数量输入
      ├── 单价输入
      ├── 金额（自动计算：数量×单价）
      ├── 仓库选择（如有）
      └── 删除行按钮
      └── 空行提示

汇总区
  └── 合计数量 / 合计金额

底部操作区
  ├── 保存按钮（草稿状态）
  ├── 提交/确认按钮（草稿状态）
  ├── 取消按钮（草稿状态）
  ├── 完成按钮（已确认状态）
  └── 返回按钮
```

**新增/编辑/查看统一走详情页**：
- 点击"新增" → `/purchase-orders/new`
- 点击"编辑" → `/purchase-orders/:id?mode=edit`
- 点击"查看" → `/purchase-orders/:id`
- 共用同一页面组件，`mode` 参数控制可编辑状态（`create`/`edit`/`view`）
- `mode=edit` 时表单可编辑；无 `mode` 参数时为查看模式（只读）；`new` 路由时为创建模式

### 状态流转规则

| 模块 | 单据类型 | 状态 | 可执行操作 |
|------|---------|------|-----------|
| 采购 | 订单 | draft → confirmed → cancelled | 确认、取消 |
| 采购 | 入库 | draft → completed | 完成 |
| 采购 | 退货 | draft → completed | 完成 |
| 采购 | 付款 | draft → completed | 完成 |
| 销售 | 订单 | draft → confirmed → cancelled | 确认、取消 |
| 销售 | 出库 | draft → completed | 完成 |
| 销售 | 退货 | draft → completed | 完成 |
| 销售 | 收款 | draft → completed | 完成 |
| 库存 | 盘点 | draft → completed | 完成 |
| 库存 | 调拨 | draft → completed | 完成 |
| 审批 | 记录 | pending → approved / rejected | 通过、驳回 |

### 关联数据选择器

详情页中需要下拉选择的关联数据，通过对应列表接口获取：

| 字段 | 数据源接口 | 显示字段 |
|------|-----------|---------|
| 供应商 | `GET /api/v1/customers?type=supplier` | name |
| 客户 | `GET /api/v1/customers?type=customer` | name |
| 仓库 | `GET /api/v1/warehouses` | name |
| 商品 | `GET /api/v1/products` | name + code |
| 员工 | `GET /api/v1/employees` | name |
| 部门 | `GET /api/v1/departments` | name |

所有选择器支持远程搜索（`el-select` + `filterable` + `remote`）。

### 数据流

```
列表页
  → 进入页面调用 fetchList()
  → 搜索条件变化 → 触发 handleSearch() → 重置 page=1 → fetchList()
  → 分页变化 → 触发 handleCurrentChange/handleSizeChange → fetchList()
  → 点击状态流转按钮 → 调用 API → 成功后刷新列表

详情页
  → 进入页面
    ├── create 模式：初始化空表单和空明细表
    └── edit/view 模式：调用 GET /api/v1/xxx/:id 加载数据
  → 添加明细行 → 推入空对象到 items 数组
  → 删除明细行 → splice 数组
  → 保存/提交 → POST/PUT 数据 → 成功后跳回列表页
  → 状态流转 → 调用对应 API → 成功后刷新数据
```

### 响应式状态管理

- 继续使用 Pinia `auth` store 管理登录态
- 单据页面状态使用组件内 `ref`/`reactive`，不单独建 store
- 列表页仍可使用 `useCrud` composable，但需扩展支持分页和搜索
- 详情页单独用 `ref` 管理表单状态和明细数组

### 组件复用策略

| 组件 | 用途 | 位置 |
|------|------|------|
| `BillListTemplate` | 单据列表页通用模板（搜索+表格+分页） | `components/` |
| `BillDetailTemplate` | 单据详情页通用模板（表单+明细+操作） | `components/` |
| `RemoteSelect` | 远程搜索下拉选择器 | `components/` |
| `BillItemTable` | 从表明细编辑表格 | `components/` |

### API 对接规范

- 列表接口：`GET /api/v1/xxx?page=1&pageSize=20&keyword=...`
- 详情接口：`GET /api/v1/xxx/:id`
- 创建接口：`POST /api/v1/xxx`
- 更新接口：`PUT /api/v1/xxx/:id`
- 删除接口：`DELETE /api/v1/xxx/:id`
- 状态流转：`POST /api/v1/xxx/:id/confirm`（确认）等
- 响应统一格式：`{ code, message, data }`

## 非功能要求

- 页面加载时显示 `el-loading` 或 `v-loading`
- 操作成功后显示 `ElMessage.success`
- 操作失败显示 `ElMessage.error`（取后端 message）
- 删除和关键操作前显示 `ElMessageBox.confirm` 确认
- 表单提交时禁用按钮防止重复提交

## 开发顺序

### Phase 1：基础组件升级（1 个模块）

先完成采购模块全部 4 个子模块，同时沉淀出可复用组件：

1. 升级 `useCrud` 支持分页和搜索
2. 创建 `BillListTemplate`、`BillDetailTemplate`、`RemoteSelect`、`BillItemTable` 组件
3. 开发采购订单完整页面（列表 + 详情）
4. 开发采购入库完整页面
5. 开发采购退货完整页面
6. 开发采购付款完整页面

### Phase 2：横向复制（3 个模块）

用 Phase 1 沉淀的模板快速开发：

1. 销售模块（订单 / 出库 / 退货 / 收款）
2. 库存模块（盘点 / 调拨 / 预警）
3. 审批模块（流程 / 记录）

## 风险与注意事项

- 后端 API 中的从表明细字段名可能不一致（如 `items`、`details`、`orderItems`），需逐个模块核对
- 状态流转 API 路径可能不一致（如 `confirm`、`complete`、`approve`），需按模块核对
- 关联对象接口可能需要新增查询参数（如 `?type=supplier`），如果后端不支持需同步调整
- 单据编号由后端自动生成，前端创建时不需要填写
