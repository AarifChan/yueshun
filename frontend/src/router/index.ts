import { createRouter, createWebHistory, RouterView } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'Login',
      component: () => import('@/views/LoginView.vue'),
      meta: { public: true },
    },
    {
      path: '/login/wecom/callback',
      name: 'WecomCallback',
      component: () => import('@/views/WecomCallbackView.vue'),
      meta: { public: true },
    },
    {
      path: '/',
      name: 'Layout',
      component: () => import('@/layouts/MainLayout.vue'),
      redirect: '/dashboard',
      children: [
        {
          path: 'dashboard',
          name: 'Dashboard',
          component: () => import('@/views/DashboardView.vue'),
          meta: { title: '首页', icon: 'HomeFilled' },
        },
        {
          path: 'bill-center',
          name: 'BillCenter',
          component: () => import('@/views/bill/BillCenterView.vue'),
          meta: { title: '单据中心', icon: 'Tickets' },
        },
        // 基础资料
        {
          path: '',
          name: 'GroupBase',
          component: RouterView,
          meta: { title: '基础资料', icon: 'Setting' },
          children: [
            {
              path: 'dicts',
              name: 'Dicts',
              component: () => import('@/views/base/DictView.vue'),
              meta: { title: '字典管理', icon: 'Collection' },
            },
            {
              path: 'departments',
              name: 'Departments',
              component: () => import('@/views/base/DepartmentView.vue'),
              meta: { title: '部门管理', icon: 'OfficeBuilding' },
            },
            {
              path: 'employees',
              name: 'Employees',
              component: () => import('@/views/base/EmployeeView.vue'),
              meta: { title: '职员管理', icon: 'UserFilled' },
            },
            {
              path: 'roles',
              name: 'Roles',
              component: () => import('@/views/base/RoleView.vue'),
              meta: { title: '角色管理', icon: 'Key' },
            },
            {
              path: 'permissions',
              name: 'Permissions',
              component: () => import('@/views/base/PermissionView.vue'),
              meta: { title: '权限管理', icon: 'Lock' },
            },
          ],
        },
        // 资料
        {
          path: '',
          name: 'GroupBaseData',
          component: RouterView,
          meta: { title: '资料', icon: 'Folder' },
          children: [
            {
              path: 'base-data/stock-op-types',
              name: 'StockOpTypes',
              component: () => import('@/views/basedata/StockOpTypeView.vue'),
              meta: { title: '出入库类型', icon: 'Sort' },
            },
            {
              path: 'base-data/delivery-methods',
              name: 'DeliveryMethods',
              component: () => import('@/views/basedata/DeliveryMethodView.vue'),
              meta: { title: '发货方式', icon: 'Van' },
            },
            {
              path: 'base-data/logistics-companies',
              name: 'LogisticsCompanies',
              component: () => import('@/views/basedata/LogisticsCompanyView.vue'),
              meta: { title: '物流公司', icon: 'Location' },
            },
            {
              path: 'base-data/materials',
              name: 'Materials',
              component: () => import('@/views/basedata/MaterialView.vue'),
              meta: { title: '素材库', icon: 'Picture' },
            },
            {
              path: 'base-data/initial-stocks',
              name: 'InitialStocks',
              component: () => import('@/views/basedata/InitialStockView.vue'),
              meta: { title: '商品库存期初', icon: 'Box' },
            },
            {
              path: 'base-data/initial-balances',
              name: 'InitialBalances',
              component: () => import('@/views/basedata/InitialBalanceView.vue'),
              meta: { title: '往来期初', icon: 'ScaleToOriginal' },
            },
            {
              path: 'base-data/initial-accounts',
              name: 'InitialAccounts',
              component: () => import('@/views/basedata/InitialAccountView.vue'),
              meta: { title: '现金银行期初', icon: 'Wallet' },
            },
          ],
        },
        // 商品资料
        {
          path: '',
          name: 'GroupProduct',
          component: RouterView,
          meta: {
            title: '商品资料',
            icon: 'Goods',
            mega: [
              {
                title: '商品资料', icon: 'Goods',
                items: [
                  { title: '商品信息', path: '/products' },
                  { title: '条码打印', path: '/barcode-print' },
                  { title: '商品审核', path: '/product-audit' },
                  { title: '规格单位条码', path: '/sku-unit-codes' },
                ],
              },
              {
                title: '价格管理', icon: 'PriceTag',
                items: [
                  { title: '价格体系', path: '/price-levels' },
                  { title: '商品物价管理', path: '/price-manage' },
                  { title: '客户单独定价', path: '/customer-prices' },
                  { title: '销售价格跟踪', path: '/sale-price-track' },
                  { title: '采购价格跟踪', path: '/purchase-price-track' },
                ],
              },
              {
                title: '商品设置', icon: 'Setting',
                items: [
                  { title: '商品分类', path: '/categories' },
                  { title: '商品单位', path: '/product-units' },
                  { title: '商品规格', path: '/product-specs' },
                  { title: '商品品牌', path: '/brands' },
                  { title: '商品标签', path: '/product-tags' },
                  { title: '商品销售范围', path: '/product-scopes' },
                  { title: '商品自定义字段', path: '/product-custom-fields' },
                ],
              },
            ],
          },
          children: [
            {
              path: 'products',
              name: 'Products',
              component: () => import('@/views/product/ProductView.vue'),
              meta: { title: '商品管理', icon: 'Goods' },
            },
            {
              path: 'categories',
              name: 'Categories',
              component: () => import('@/views/product/CategoryView.vue'),
              meta: { title: '商品分类', icon: 'FolderOpened' },
            },
            {
              path: 'brands',
              name: 'Brands',
              component: () => import('@/views/product/BrandView.vue'),
              meta: { title: '品牌管理', icon: 'Flag' },
            },
            {
              path: 'price-levels',
              name: 'PriceLevels',
              component: () => import('@/views/price/PriceLevelView.vue'),
              meta: { title: '价格体系', icon: 'PriceTag' },
            },
            {
              path: 'price-manage',
              name: 'PriceManage',
              component: () => import('@/views/price/PriceManageView.vue'),
              meta: { title: '商品物价管理', icon: 'PriceTag' },
            },
            {
              path: 'product-units',
              name: 'ProductUnits',
              component: () => import('@/views/product/UnitView.vue'),
              meta: { title: '商品单位', icon: 'Odometer' },
            },
            {
              path: 'product-specs',
              name: 'ProductSpecs',
              component: () => import('@/views/product/SpecView.vue'),
              meta: { title: '商品规格', icon: 'Operation' },
            },
            {
              path: 'product-tags',
              name: 'ProductTags',
              component: () => import('@/views/product/TagView.vue'),
              meta: { title: '商品标签', icon: 'PriceTag' },
            },
            {
              path: 'product-scopes',
              name: 'ProductScopes',
              component: () => import('@/views/product/ScopeView.vue'),
              meta: { title: '商品销售范围', icon: 'Position' },
            },
            {
              path: 'barcode-print',
              name: 'BarcodePrint',
              component: () => import('@/views/product/BarcodePrintView.vue'),
              meta: { title: '条码打印', icon: 'Printer' },
            },
            {
              path: 'product-audit',
              name: 'ProductAudit',
              component: () => import('@/views/product/ProductAuditView.vue'),
              meta: { title: '商品审核', icon: 'CircleCheck' },
            },
            {
              path: 'sku-unit-codes',
              name: 'SkuUnitCodes',
              component: () => import('@/views/product/SkuUnitCodeView.vue'),
              meta: { title: '规格单位条码', icon: 'Postcard' },
            },
            {
              path: 'customer-prices',
              name: 'CustomerPrices',
              component: () => import('@/views/product/CustomerPriceView.vue'),
              meta: { title: '客户单独定价', icon: 'Money' },
            },
            {
              path: 'sale-price-track',
              name: 'SalePriceTrack',
              component: () => import('@/views/product/SalePriceTrackView.vue'),
              meta: { title: '销售价格跟踪', icon: 'TrendCharts' },
            },
            {
              path: 'purchase-price-track',
              name: 'PurchasePriceTrack',
              component: () => import('@/views/product/PurchasePriceTrackView.vue'),
              meta: { title: '采购价格跟踪', icon: 'TrendCharts' },
            },
            {
              path: 'product-custom-fields',
              name: 'ProductCustomFields',
              component: () => import('@/views/product/ProductCustomFieldView.vue'),
              meta: { title: '商品自定义字段', icon: 'EditPen' },
            },
          ],
        },
        // 客户管理
        {
          path: '',
          name: 'GroupCustomer',
          component: RouterView,
          meta: { title: '客户管理', icon: 'User' },
          children: [
            {
              path: 'customers',
              name: 'Customers',
              component: () => import('@/views/customer/CustomerView.vue'),
              meta: { title: '客户管理', icon: 'User' },
            },
            {
              path: 'customer-categories',
              name: 'CustomerCategories',
              component: () => import('@/views/customer/CustomerCategoryView.vue'),
              meta: { title: '客户分类', icon: 'Files' },
            },
            {
              path: 'regions',
              name: 'Regions',
              component: () => import('@/views/customer/RegionView.vue'),
              meta: { title: '区域管理', icon: 'MapLocation' },
            },
            {
              path: 'levels',
              name: 'Levels',
              component: () => import('@/views/customer/LevelView.vue'),
              meta: { title: '客户等级', icon: 'Medal' },
            },
          ],
        },
        // 库存管理
        {
          path: '',
          name: 'GroupInventory',
          component: RouterView,
          meta: {
            title: '库存管理',
            icon: 'Box',
            mega: [
              {
                title: '单据', icon: 'Tickets',
                items: [
                  { title: '其他入库单', path: '/inventory/other-in' },
                  { title: '其他出库单', path: '/inventory/other-out' },
                  { title: '成本调价单', path: '/inventory/cost-adjust' },
                  { title: '库存盘点单', path: '/inventory-checks' },
                ],
              },
              {
                title: '调拨业务', icon: 'Van',
                items: [
                  { title: '调拨申请单', path: '/inventory/transfer-apply' },
                  { title: '调拨出库单', path: '/inventory/transfer-out' },
                  { title: '调拨入库单', path: '/inventory/transfer-in' },
                  { title: '调拨差异处理', path: '/inventory/transfer-diff' },
                  { title: '同价调拨单', path: '/inventory-transfers' },
                ],
              },
              {
                title: '商品加工', icon: 'SetUp',
                items: [
                  { title: '拆装模板', path: '/inventory/assembly-templates' },
                  { title: '组装拆装单', path: '/inventory/assembly-orders' },
                ],
              },
              {
                title: '库存报表', icon: 'DataAnalysis',
                items: [
                  { title: '库存状况表', path: '/stock-status' },
                  { title: '库存分布表', path: '/inventory/stock-distribution' },
                  { title: '库存报警', path: '/inventory-warnings' },
                  { title: '商品进销存汇总', path: '/inventory/inout-summary' },
                  { title: '调拨汇总表', path: '/inventory/transfer-summary' },
                  { title: '其他出入库统计', path: '/inventory/other-inout-stats' },
                  { title: '库存批号统计', path: '/inventory/batch-stats' },
                  { title: '批号跟踪详情', path: '/inventory/batch-trace' },
                  { title: '商品近效期预警', path: '/inventory/expiry-warning' },
                  { title: '商品出入库流水', path: '/inventory/inout-flow' },
                ],
              },
              {
                title: '仓库设置', icon: 'House',
                items: [
                  { title: '仓库管理', path: '/warehouses' },
                  { title: '库存查询', path: '/stocks' },
                ],
              },
            ],
          },
          children: [
            {
              path: 'stocks',
              name: 'Stocks',
              component: () => import('@/views/inventory/StockView.vue'),
              meta: { title: '库存查询', icon: 'Coin' },
            },
            {
              path: 'stock-status',
              name: 'StockStatus',
              component: () => import('@/views/inventory/StockStatusView.vue'),
              meta: { title: '库存状况表', icon: 'DataAnalysis' },
            },
            {
              path: 'inventory-checks',
              name: 'InventoryChecks',
              component: () => import('@/views/inventory/CheckView.vue'),
              meta: { title: '库存盘点', icon: 'Check' },
            },
            {
              path: 'inventory-transfers',
              name: 'InventoryTransfers',
              component: () => import('@/views/inventory/TransferView.vue'),
              meta: { title: '库存调拨', icon: 'Switch' },
            },
            {
              path: 'inventory-warnings',
              name: 'InventoryWarnings',
              component: () => import('@/views/inventory/WarningView.vue'),
              meta: { title: '库存预警', icon: 'Warning' },
            },
            {
              path: 'warehouses',
              name: 'Warehouses',
              component: () => import('@/views/warehouse/WarehouseView.vue'),
              meta: { title: '仓库管理', icon: 'House' },
            },
            {
              path: 'inventory/other-in',
              name: 'InventoryOtherIn',
              component: () => import('@/views/inventory/OtherInStockView.vue'),
              meta: { title: '其他入库单', icon: 'Document' },
            },
            {
              path: 'inventory/other-in/new',
              name: 'InventoryOtherInCreate',
              component: () => import('@/views/inventory/OtherInStockDetailView.vue'),
              meta: { title: '新增其他入库单', hidden: true },
            },
            {
              path: 'inventory/other-in/:id',
              name: 'InventoryOtherInDetail',
              component: () => import('@/views/inventory/OtherInStockDetailView.vue'),
              meta: { title: '其他入库单详情', hidden: true },
            },
            {
              path: 'inventory/other-out',
              name: 'InventoryOtherOut',
              component: () => import('@/views/inventory/OtherOutStockView.vue'),
              meta: { title: '其他出库单', icon: 'Document' },
            },
            {
              path: 'inventory/other-out/new',
              name: 'InventoryOtherOutCreate',
              component: () => import('@/views/inventory/OtherOutStockDetailView.vue'),
              meta: { title: '新增其他出库单', hidden: true },
            },
            {
              path: 'inventory/other-out/:id',
              name: 'InventoryOtherOutDetail',
              component: () => import('@/views/inventory/OtherOutStockDetailView.vue'),
              meta: { title: '其他出库单详情', hidden: true },
            },
            {
              path: 'inventory/cost-adjust',
              name: 'InventoryCostAdjust',
              component: () => import('@/views/inventory/CostAdjustView.vue'),
              meta: { title: '成本调价单', icon: 'Document' },
            },
            {
              path: 'inventory/cost-adjust/new',
              name: 'InventoryCostAdjustCreate',
              component: () => import('@/views/inventory/CostAdjustDetailView.vue'),
              meta: { title: '新增成本调价单', hidden: true },
            },
            {
              path: 'inventory/cost-adjust/:id',
              name: 'InventoryCostAdjustDetail',
              component: () => import('@/views/inventory/CostAdjustDetailView.vue'),
              meta: { title: '成本调价单详情', hidden: true },
            },
            {
              path: 'inventory/transfer-apply',
              name: 'InventoryTransferApply',
              component: () => import('@/views/inventory/TransferApplyView.vue'),
              meta: { title: '调拨申请单', icon: 'Document' },
            },
            {
              path: 'inventory/transfer-apply/new',
              name: 'InventoryTransferApplyCreate',
              component: () => import('@/views/inventory/TransferApplyDetailView.vue'),
              meta: { title: '新增调拨申请单', hidden: true },
            },
            {
              path: 'inventory/transfer-apply/:id',
              name: 'InventoryTransferApplyDetail',
              component: () => import('@/views/inventory/TransferApplyDetailView.vue'),
              meta: { title: '调拨申请单详情', hidden: true },
            },
            {
              path: 'inventory/transfer-out',
              name: 'InventoryTransferOut',
              component: () => import('@/views/inventory/TransferOutView.vue'),
              meta: { title: '调拨出库单', icon: 'Document' },
            },
            {
              path: 'inventory/transfer-out/new',
              name: 'InventoryTransferOutCreate',
              component: () => import('@/views/inventory/TransferOutDetailView.vue'),
              meta: { title: '新增调拨出库单', hidden: true },
            },
            {
              path: 'inventory/transfer-out/:id',
              name: 'InventoryTransferOutDetail',
              component: () => import('@/views/inventory/TransferOutDetailView.vue'),
              meta: { title: '调拨出库单详情', hidden: true },
            },
            {
              path: 'inventory/transfer-in',
              name: 'InventoryTransferIn',
              component: () => import('@/views/inventory/TransferInView.vue'),
              meta: { title: '调拨入库单', icon: 'Document' },
            },
            {
              path: 'inventory/transfer-in/new',
              name: 'InventoryTransferInCreate',
              component: () => import('@/views/inventory/TransferInDetailView.vue'),
              meta: { title: '新增调拨入库单', hidden: true },
            },
            {
              path: 'inventory/transfer-in/:id',
              name: 'InventoryTransferInDetail',
              component: () => import('@/views/inventory/TransferInDetailView.vue'),
              meta: { title: '调拨入库单详情', hidden: true },
            },
            {
              path: 'inventory/transfer-diff',
              name: 'InventoryTransferDiff',
              component: () => import('@/views/inventory/TransferDiffView.vue'),
              meta: { title: '调拨差异处理', icon: 'Document' },
            },
            {
              path: 'inventory/assembly-templates',
              name: 'InventoryAssemblyTemplates',
              component: () => import('@/views/inventory/AssemblyTemplateView.vue'),
              meta: { title: '拆装模板', icon: 'Document' },
            },
            {
              path: 'inventory/assembly-orders',
              name: 'InventoryAssemblyOrders',
              component: () => import('@/views/inventory/AssemblyOrderView.vue'),
              meta: { title: '组装拆装单', icon: 'Document' },
            },
            {
              path: 'inventory/assembly-orders/new',
              name: 'InventoryAssemblyOrderCreate',
              component: () => import('@/views/inventory/AssemblyOrderDetailView.vue'),
              meta: { title: '新增组装拆装单', hidden: true },
            },
            {
              path: 'inventory/assembly-orders/:id',
              name: 'InventoryAssemblyOrderDetail',
              component: () => import('@/views/inventory/AssemblyOrderDetailView.vue'),
              meta: { title: '组装拆装单详情', hidden: true },
            },
            // 库存报表
            {
              path: 'inventory/stock-distribution',
              name: 'InventoryStockDistribution',
              component: () => import('@/views/inventory/StockDistributionView.vue'),
              meta: { title: '库存分布表', icon: 'Document' },
            },
            {
              path: 'inventory/inout-summary',
              name: 'InventoryInoutSummary',
              component: () => import('@/views/inventory/InoutSummaryView.vue'),
              meta: { title: '商品进销存汇总', icon: 'Document' },
            },
            {
              path: 'inventory/transfer-summary',
              name: 'InventoryTransferSummary',
              component: () => import('@/views/inventory/TransferSummaryView.vue'),
              meta: { title: '调拨汇总表', icon: 'Document' },
            },
            {
              path: 'inventory/other-inout-stats',
              name: 'InventoryOtherInoutStats',
              component: () => import('@/views/inventory/OtherInoutStatsView.vue'),
              meta: { title: '其他出入库统计', icon: 'Document' },
            },
            {
              path: 'inventory/batch-stats',
              name: 'InventoryBatchStats',
              component: () => import('@/views/inventory/BatchStatsView.vue'),
              meta: { title: '库存批号统计', icon: 'Document' },
            },
            {
              path: 'inventory/batch-trace',
              name: 'InventoryBatchTrace',
              component: () => import('@/views/inventory/BatchTraceView.vue'),
              meta: { title: '批号跟踪详情', icon: 'Document' },
            },
            {
              path: 'inventory/expiry-warning',
              name: 'InventoryExpiryWarning',
              component: () => import('@/views/inventory/ExpiryWarningView.vue'),
              meta: { title: '商品近效期预警', icon: 'Document' },
            },
            {
              path: 'inventory/inout-flow',
              name: 'InventoryInoutFlow',
              component: () => import('@/views/inventory/InoutFlowView.vue'),
              meta: { title: '商品出入库流水', icon: 'Document' },
            },
          ],
        },
        // 采购管理
        {
          path: '',
          name: 'GroupPurchase',
          component: RouterView,
          meta: { title: '采购管理', icon: 'ShoppingCart' },
          children: [
            {
              path: 'purchase-orders',
              name: 'PurchaseOrders',
              component: () => import('@/views/purchase/OrderView.vue'),
              meta: { title: '采购订单', icon: 'ShoppingCart' },
            },
            {
              path: 'purchase-instock',
              name: 'PurchaseInStock',
              component: () => import('@/views/purchase/InStockView.vue'),
              meta: { title: '采购入库', icon: 'Box' },
            },
            {
              path: 'purchase-returns',
              name: 'PurchaseReturns',
              component: () => import('@/views/purchase/ReturnView.vue'),
              meta: { title: '采购退货', icon: 'Back' },
            },
            {
              path: 'suppliers',
              name: 'Suppliers',
              component: () => import('@/views/purchase/SupplierView.vue'),
              meta: { title: '供应商管理', icon: 'OfficeBuilding' },
            },
            {
              path: 'purchase-payments',
              name: 'PurchasePayments',
              component: () => import('@/views/purchase/PaymentView.vue'),
              meta: { title: '采购付款', icon: 'CreditCard' },
            },
            // 采购策略
            {
              path: 'purchase-replenish',
              name: 'PurchaseReplenish',
              component: () => import('@/views/purchase/report/ReplenishView.vue'),
              meta: { title: '智能补货', icon: 'MagicStick' },
            },
            // 采购报表
            {
              path: 'purchase-reports/product-stats',
              name: 'PurchaseProductStats',
              component: () => import('@/views/purchase/report/ProductStatsView.vue'),
              meta: { title: '商品采购统计', icon: 'DataAnalysis' },
            },
            {
              path: 'purchase-reports/supplier-stats',
              name: 'PurchaseSupplierStats',
              component: () => import('@/views/purchase/report/SupplierStatsView.vue'),
              meta: { title: '供应商采购统计', icon: 'DataAnalysis' },
            },
            {
              path: 'purchase-reports/product-detail',
              name: 'PurchaseProductDetail',
              component: () => import('@/views/purchase/report/ProductDetailView.vue'),
              meta: { title: '商品采购明细统计', icon: 'DataAnalysis' },
            },
            {
              path: 'purchase-reports/order-execution',
              name: 'PurchaseOrderExecution',
              component: () => import('@/views/purchase/report/OrderExecutionView.vue'),
              meta: { title: '采购订单执行明细', icon: 'DataAnalysis' },
            },
          ],
        },
        // 销售管理
        {
          path: '',
          name: 'GroupSale',
          component: RouterView,
          meta: { title: '销售管理', icon: 'ShoppingBag' },
          children: [
            {
              path: 'sales-orders',
              name: 'SalesOrders',
              component: () => import('@/views/sale/OrderView.vue'),
              meta: { title: '销售订单', icon: 'ShoppingBag' },
            },
            {
              path: 'sales-outstock',
              name: 'SalesOutStock',
              component: () => import('@/views/sale/OutStockView.vue'),
              meta: { title: '销售出库', icon: 'Sell' },
            },
            {
              path: 'sales-returns',
              name: 'SalesReturns',
              component: () => import('@/views/sale/ReturnView.vue'),
              meta: { title: '销售退货', icon: 'RefreshLeft' },
            },
            {
              path: 'sales-receipts',
              name: 'SalesReceipts',
              component: () => import('@/views/sale/ReceiptView.vue'),
              meta: { title: '销售收款', icon: 'Money' },
            },
            // 销售对账
            {
              path: 'sale-reports/daily-summary',
              name: 'SaleDailySummary',
              component: () => import('@/views/sale/report/DailySummaryView.vue'),
              meta: { title: '日销售统计', icon: 'DataAnalysis' },
            },
            {
              path: 'sale-reports/daily-close',
              name: 'SaleDailyClose',
              component: () => import('@/views/sale/report/DailyCloseView.vue'),
              meta: { title: '日清日结', icon: 'DataAnalysis' },
            },
            // 销售报表
            {
              path: 'sale-reports/product-stats',
              name: 'SaleProductStats',
              component: () => import('@/views/sale/report/ProductStatsView.vue'),
              meta: { title: '商品销售统计', icon: 'DataAnalysis' },
            },
            {
              path: 'sale-reports/customer-stats',
              name: 'SaleCustomerStats',
              component: () => import('@/views/sale/report/CustomerStatsView.vue'),
              meta: { title: '客户销售统计', icon: 'DataAnalysis' },
            },
            {
              path: 'sale-reports/employee-stats',
              name: 'SaleEmployeeStats',
              component: () => import('@/views/sale/report/EmployeeStatsView.vue'),
              meta: { title: '职员销售业绩统计', icon: 'DataAnalysis' },
            },
            {
              path: 'sale-reports/product-detail',
              name: 'SaleProductDetail',
              component: () => import('@/views/sale/report/ProductDetailStatsView.vue'),
              meta: { title: '商品销售明细统计', icon: 'DataAnalysis' },
            },
            {
              path: 'sale-reports/order-execution',
              name: 'SaleOrderExecution',
              component: () => import('@/views/sale/report/OrderExecutionView.vue'),
              meta: { title: '销售订单执行明细', icon: 'DataAnalysis' },
            },
            {
              path: 'sale-reports/comprehensive',
              name: 'SaleComprehensive',
              component: () => import('@/views/sale/report/ComprehensiveView.vue'),
              meta: { title: '订销退综合分析', icon: 'DataAnalysis' },
            },
            // 销售提成
            {
              path: 'commission-plans',
              name: 'CommissionPlans',
              component: () => import('@/views/sale/report/CommissionPlanView.vue'),
              meta: { title: '提成方案列表', icon: 'SetUp' },
            },
            {
              path: 'commission-stats',
              name: 'CommissionStats',
              component: () => import('@/views/sale/report/CommissionStatsView.vue'),
              meta: { title: '职员提成统计', icon: 'DataAnalysis' },
            },
          ],
        },
        // 资金管理
        {
          path: '',
          name: 'GroupFinance',
          component: RouterView,
          meta: { title: '资金管理', icon: 'Wallet' },
          children: [
            {
              path: 'accounts',
              name: 'Accounts',
              component: () => import('@/views/warehouse/AccountView.vue'),
              meta: { title: '资金账户', icon: 'Wallet' },
            },
            {
              path: 'income-expense',
              name: 'IncomeExpense',
              component: () => import('@/views/warehouse/IncomeExpenseView.vue'),
              meta: { title: '收支项目', icon: 'Money' },
            },
            // 资金单据
            {
              path: 'expense-bills',
              name: 'ExpenseBills',
              component: () => import('@/views/finance/ExpenseBillView.vue'),
              meta: { title: '费用单', icon: 'Document' },
            },
            {
              path: 'other-income-bills',
              name: 'OtherIncomeBills',
              component: () => import('@/views/finance/OtherIncomeBillView.vue'),
              meta: { title: '其他收入单', icon: 'Document' },
            },
            // 往来报表
            {
              path: 'finance/receivable',
              name: 'FinanceReceivable',
              component: () => import('@/views/finance/ReceivableView.vue'),
              meta: { title: '应收查询', icon: 'DataAnalysis' },
            },
            {
              path: 'finance/payable',
              name: 'FinancePayable',
              component: () => import('@/views/finance/PayableView.vue'),
              meta: { title: '应付查询', icon: 'DataAnalysis' },
            },
            {
              path: 'finance/balance',
              name: 'FinanceBalance',
              component: () => import('@/views/finance/BalanceView.vue'),
              meta: { title: '往来余额查询', icon: 'DataAnalysis' },
            },
            {
              path: 'finance/customer-advance',
              name: 'FinanceCustomerAdvance',
              component: () => import('@/views/finance/CustomerAdvanceView.vue'),
              meta: { title: '客户预收查询', icon: 'DataAnalysis' },
            },
            {
              path: 'finance/brand-advance',
              name: 'FinanceBrandAdvance',
              component: () => import('@/views/finance/BrandPlaceholderView.vue'),
              props: { title: '品牌预收查询' },
              meta: { title: '品牌预收查询', icon: 'DataAnalysis' },
            },
            {
              path: 'finance/brand-unsettled',
              name: 'FinanceBrandUnsettled',
              component: () => import('@/views/finance/BrandPlaceholderView.vue'),
              props: { title: '品牌待结算查询' },
              meta: { title: '品牌待结算查询', icon: 'DataAnalysis' },
            },
            {
              path: 'finance/unsettled-bills',
              name: 'FinanceUnsettledBills',
              component: () => import('@/views/finance/UnsettledBillsView.vue'),
              meta: { title: '单据待结算查询', icon: 'DataAnalysis' },
            },
            // 统计报表
            {
              path: 'finance/cash-bank',
              name: 'FinanceCashBank',
              component: () => import('@/views/finance/CashBankView.vue'),
              meta: { title: '现金银行账户统计', icon: 'DataAnalysis' },
            },
            {
              path: 'finance/account-income',
              name: 'FinanceAccountIncome',
              component: () => import('@/views/finance/AccountIncomeView.vue'),
              meta: { title: '账户收支统计', icon: 'DataAnalysis' },
            },
            {
              path: 'finance/revenue-expense',
              name: 'FinanceRevenueExpense',
              component: () => import('@/views/finance/RevenueExpenseView.vue'),
              meta: { title: '收入费用统计', icon: 'DataAnalysis' },
            },
            {
              path: 'finance/expense-distribution',
              name: 'FinanceExpenseDistribution',
              component: () => import('@/views/finance/ExpenseDistributionView.vue'),
              meta: { title: '费用支出分布', icon: 'DataAnalysis' },
            },
            {
              path: 'finance/income-distribution',
              name: 'FinanceIncomeDistribution',
              component: () => import('@/views/finance/IncomeDistributionView.vue'),
              meta: { title: '其他收入分布', icon: 'DataAnalysis' },
            },
            {
              path: 'finance/operating-profit',
              name: 'FinanceOperatingProfit',
              component: () => import('@/views/finance/OperatingProfitView.vue'),
              meta: { title: '经营利润统计', icon: 'DataAnalysis' },
            },
          ],
        },
        // 商城管理
        {
          path: '',
          name: 'GroupMall',
          component: RouterView,
          meta: { title: '商城管理', icon: 'Shop' },
          children: [
            {
              path: 'mall-products',
              name: 'MallProducts',
              component: () => import('@/views/mall/ProductView.vue'),
              meta: { title: '商城商品', icon: 'Shop' },
            },
            {
              path: 'mall-orders',
              name: 'MallOrders',
              component: () => import('@/views/mall/OrderView.vue'),
              meta: { title: '商城订单', icon: 'Document' },
            },
            {
              path: 'mall-carts',
              name: 'MallCarts',
              component: () => import('@/views/mall/CartView.vue'),
              meta: { title: '购物车', icon: 'ShoppingCartFull' },
            },
          ],
        },
        // CRM管理
        {
          path: '',
          name: 'GroupCrm',
          component: RouterView,
          meta: { title: 'CRM管理', icon: 'PhoneFilled' },
          children: [
            {
              path: 'follow-ups',
              name: 'FollowUps',
              component: () => import('@/views/crm/FollowUpView.vue'),
              meta: { title: '跟进记录', icon: 'Phone' },
            },
            {
              path: 'opportunities',
              name: 'Opportunities',
              component: () => import('@/views/crm/OpportunityView.vue'),
              meta: { title: '商机管理', icon: 'TrendCharts' },
            },
            {
              path: 'contracts',
              name: 'Contracts',
              component: () => import('@/views/crm/ContractView.vue'),
              meta: { title: '合同管理', icon: 'DocumentChecked' },
            },
            // CRM 增强
            {
              path: 'customer-sea',
              name: 'CustomerSea',
              component: () => import('@/views/crm/CustomerSeaView.vue'),
              meta: { title: '客户公海', icon: 'Ship' },
            },
            {
              path: 'work-reports',
              name: 'WorkReports',
              component: () => import('@/views/crm/WorkReportView.vue'),
              meta: { title: '汇报', icon: 'EditPen' },
            },
            {
              path: 'customer-tags',
              name: 'CustomerTags',
              component: () => import('@/views/crm/CustomerTagView.vue'),
              meta: { title: '客户标签', icon: 'PriceTag' },
            },
            {
              path: 'supplier-categories',
              name: 'SupplierCategories',
              component: () => import('@/views/crm/SupplierCategoryView.vue'),
              meta: { title: '供应商分类', icon: 'Files' },
            },
            {
              path: 'custom-fields',
              name: 'CustomFields',
              component: () => import('@/views/crm/CustomFieldView.vue'),
              meta: { title: '客户自定义字段', icon: 'Operation' },
            },
            {
              path: 'report-templates',
              name: 'ReportTemplates',
              component: () => import('@/views/crm/ReportTemplateView.vue'),
              meta: { title: '汇报模板', icon: 'Document' },
            },
            {
              path: 'approval-templates',
              name: 'ApprovalTemplates',
              component: () => import('@/views/approval/ProcessView.vue'),
              meta: { title: '审批模板', icon: 'Stamp' },
            },
            {
              path: 'opportunity-settings',
              name: 'OpportunitySettings',
              component: () => import('@/views/crm/OpportunitySettingView.vue'),
              meta: { title: '商机设置', icon: 'SetUp' },
            },
            {
              path: 'sales-plans',
              name: 'SalesPlans',
              component: () => import('@/views/crm/SalesPlanView.vue'),
              meta: { title: '业务驾驶舱设置', icon: 'Odometer' },
            },
          ],
        },
        // 营销
        {
          path: '',
          name: 'GroupMarketing',
          component: RouterView,
          meta: {
            title: '营销',
            icon: 'Present',
            mega: [
              {
                title: '商品促销', icon: 'Present',
                items: [
                  { title: '限时特价', path: '/marketing/seckill' },
                  { title: '单品买赠', path: '/marketing/gift' },
                  { title: '阶梯价', path: '/marketing/tiered-price' },
                  { title: '组合促销', path: '/marketing/combo' },
                  { title: '优惠套餐', path: '/marketing/package' },
                ],
              },
              {
                title: '下单增长', icon: 'TrendCharts',
                items: [
                  { title: '整单优惠', path: '/marketing/whole-order' },
                  { title: '优惠券', path: '/marketing/coupons' },
                ],
              },
              {
                title: '拓客留客', icon: 'UserFilled',
                items: [
                  { title: '积分', path: '/marketing/points' },
                  { title: '分销', path: '/marketing/distribution' },
                  { title: '预收款储值', path: '/marketing/stored-value' },
                ],
              },
            ],
          },
          children: [
            {
              path: 'marketing/seckill',
              name: 'MarketingSeckill',
              component: () => import('@/views/marketing/SeckillView.vue'),
              meta: { title: '限时特价', icon: 'AlarmClock' },
            },
            {
              path: 'marketing/gift',
              name: 'MarketingGift',
              component: () => import('@/views/marketing/GiftView.vue'),
              meta: { title: '单品买赠', icon: 'Present' },
            },
            {
              path: 'marketing/tiered-price',
              name: 'MarketingTieredPrice',
              component: () => import('@/views/marketing/TieredPriceView.vue'),
              meta: { title: '阶梯价', icon: 'Sort' },
            },
            {
              path: 'marketing/combo',
              name: 'MarketingCombo',
              component: () => import('@/views/marketing/ComboPromoView.vue'),
              meta: { title: '组合促销', icon: 'Connection' },
            },
            {
              path: 'marketing/package',
              name: 'MarketingPackage',
              component: () => import('@/views/marketing/PackagePromoView.vue'),
              meta: { title: '优惠套餐', icon: 'Box' },
            },
            {
              path: 'marketing/whole-order',
              name: 'MarketingWholeOrder',
              component: () => import('@/views/marketing/WholeOrderView.vue'),
              meta: { title: '整单优惠', icon: 'Tickets' },
            },
            {
              path: 'marketing/coupons',
              name: 'MarketingCoupons',
              component: () => import('@/views/marketing/CouponView.vue'),
              meta: { title: '优惠券', icon: 'Ticket' },
            },
            {
              path: 'marketing/points',
              name: 'MarketingPoints',
              component: () => import('@/views/marketing/PointsView.vue'),
              meta: { title: '积分', icon: 'Star' },
            },
            {
              path: 'marketing/distribution',
              name: 'MarketingDistribution',
              component: () => import('@/views/marketing/DistributionView.vue'),
              meta: { title: '分销', icon: 'Share' },
            },
            {
              path: 'marketing/stored-value',
              name: 'MarketingStoredValue',
              component: () => import('@/views/marketing/StoredValueView.vue'),
              meta: { title: '预收款储值', icon: 'Wallet' },
            },
          ],
        },
        // BI
        {
          path: '',
          name: 'GroupBI',
          component: RouterView,
          meta: {
            title: 'BI',
            icon: 'DataAnalysis',
            mega: [
              {
                title: '商城经营分析', icon: 'Shop',
                items: [
                  { title: '客户商城上线', path: '/bi/mall-online' },
                ],
              },
              {
                title: '客户销售增长', icon: 'TrendCharts',
                items: [
                  { title: '新客跟踪', path: '/bi/new-customer-track' },
                  { title: '老客增品', path: '/bi/old-customer-increment' },
                  { title: '流失拉回', path: '/bi/lost-customer-back' },
                  { title: '商品铺市', path: '/bi/product-distribution' },
                ],
              },
            ],
          },
          children: [
            {
              path: 'bi/mall-online',
              name: 'BIMallOnline',
              component: () => import('@/views/bi/MallOnlineView.vue'),
              meta: { title: '客户商城上线', icon: 'Shop' },
            },
            {
              path: 'bi/new-customer-track',
              name: 'BINewCustomerTrack',
              component: () => import('@/views/bi/NewCustomerTrackView.vue'),
              meta: { title: '新客跟踪', icon: 'User' },
            },
            {
              path: 'bi/old-customer-increment',
              name: 'BIOldCustomerIncrement',
              component: () => import('@/views/bi/OldCustomerIncrementView.vue'),
              meta: { title: '老客增品', icon: 'CirclePlus' },
            },
            {
              path: 'bi/lost-customer-back',
              name: 'BILostCustomerBack',
              component: () => import('@/views/bi/LostCustomerBackView.vue'),
              meta: { title: '流失拉回', icon: 'RefreshLeft' },
            },
            {
              path: 'bi/product-distribution',
              name: 'BIProductDistribution',
              component: () => import('@/views/bi/ProductDistributionView.vue'),
              meta: { title: '商品铺市', icon: 'MapLocation' },
            },
          ],
        },
        // 生态互联
        {
          path: '',
          name: 'GroupEcosystem',
          component: RouterView,
          meta: { title: '生态互联', icon: 'Link' },
          children: [
            {
              path: 'ecosystem/guide',
              name: 'EcosystemGuide',
              component: () => import('@/views/ecosystem/UpstreamGuideView.vue'),
              meta: { title: '上游管理', icon: 'Reading' },
            },
            {
              path: 'ecosystem/supplier-links',
              name: 'EcosystemSupplierLinks',
              component: () => import('@/views/ecosystem/SupplierConnectionView.vue'),
              meta: { title: '供应商连接', icon: 'Connection' },
            },
            {
              path: 'ecosystem/receive-goods',
              name: 'EcosystemReceiveGoods',
              component: () => import('@/views/ecosystem/ReceiveGoodsView.vue'),
              meta: { title: '商品接收', icon: 'TakeawayBox' },
            },
            {
              path: 'ecosystem/receive-decoration',
              name: 'EcosystemReceiveDecoration',
              component: () => import('@/views/ecosystem/ReceiveDecorationView.vue'),
              meta: { title: '商城装修接收', icon: 'Brush' },
            },
          ],
        },
        // 审批管理
        {
          path: '',
          name: 'GroupApproval',
          component: RouterView,
          meta: { title: '审批管理', icon: 'Stamp' },
          children: [
            {
              path: 'approval-processes',
              name: 'ApprovalProcesses',
              component: () => import('@/views/approval/ProcessView.vue'),
              meta: { title: '审批流程', icon: 'SetUp' },
            },
            {
              path: 'approval-records',
              name: 'ApprovalRecords',
              component: () => import('@/views/approval/RecordView.vue'),
              meta: { title: '审批记录', icon: 'List' },
            },
          ],
        },
        // 详情页（不在菜单显示）
        {
          path: 'employees/:id',
          name: 'EmployeeDetail',
          component: () => import('@/views/base/EmployeeDetailView.vue'),
          meta: { title: '职员详情', hidden: true },
        },
        {
          path: 'products/:id',
          name: 'ProductDetail',
          component: () => import('@/views/product/ProductDetailView.vue'),
          meta: { title: '商品详情', hidden: true },
        },
        {
          path: 'customers/:id',
          name: 'CustomerDetail',
          component: () => import('@/views/customer/CustomerDetailView.vue'),
          meta: { title: '客户详情', hidden: true },
        },
        {
          path: 'inventory-checks/:id',
          name: 'InventoryCheckDetail',
          component: () => import('@/views/inventory/CheckDetailView.vue'),
          meta: { title: '库存盘点详情', hidden: true },
        },
        {
          path: 'inventory-transfers/:id',
          name: 'InventoryTransferDetail',
          component: () => import('@/views/inventory/TransferDetailView.vue'),
          meta: { title: '库存调拨详情', hidden: true },
        },
        {
          path: 'purchase-orders/:id',
          name: 'PurchaseOrderDetail',
          component: () => import('@/views/purchase/OrderDetailView.vue'),
          meta: { title: '采购订单详情', hidden: true },
        },
        {
          path: 'purchase-instock/:id',
          name: 'PurchaseInStockDetail',
          component: () => import('@/views/purchase/InStockDetailView.vue'),
          meta: { title: '采购入库详情', hidden: true },
        },
        {
          path: 'purchase-returns/:id',
          name: 'PurchaseReturnDetail',
          component: () => import('@/views/purchase/ReturnDetailView.vue'),
          meta: { title: '采购退货详情', hidden: true },
        },
        {
          path: 'purchase-payments/:id',
          name: 'PurchasePaymentDetail',
          component: () => import('@/views/purchase/PaymentDetailView.vue'),
          meta: { title: '采购付款详情', hidden: true },
        },
        {
          path: 'sales-orders/:id',
          name: 'SalesOrderDetail',
          component: () => import('@/views/sale/OrderDetailView.vue'),
          meta: { title: '销售订单详情', hidden: true },
        },
        {
          path: 'sales-outstock/:id',
          name: 'SalesOutStockDetail',
          component: () => import('@/views/sale/OutStockDetailView.vue'),
          meta: { title: '销售出库详情', hidden: true },
        },
        {
          path: 'sales-returns/:id',
          name: 'SalesReturnDetail',
          component: () => import('@/views/sale/ReturnDetailView.vue'),
          meta: { title: '销售退货详情', hidden: true },
        },
        {
          path: 'sales-receipts/:id',
          name: 'SalesReceiptDetail',
          component: () => import('@/views/sale/ReceiptDetailView.vue'),
          meta: { title: '销售收款详情', hidden: true },
        },
        {
          path: 'mall-orders/:id',
          name: 'MallOrderDetail',
          component: () => import('@/views/mall/OrderDetailView.vue'),
          meta: { title: '商城订单详情', hidden: true },
        },
        {
          path: 'contracts/:id',
          name: 'ContractDetail',
          component: () => import('@/views/crm/ContractDetailView.vue'),
          meta: { title: '合同详情', hidden: true },
        },
      ],
    },
  ],
})

router.beforeEach((to, _from, next) => {
  const authStore = useAuthStore()

  if (to.meta.public) {
    next()
    return
  }

  if (!authStore.token) {
    next('/login')
    return
  }

  next()
})

export default router
