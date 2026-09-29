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
                ],
              },
              {
                title: '价格管理', icon: 'PriceTag',
                items: [
                  { title: '价格体系', path: '/price-levels' },
                  { title: '商品物价管理', path: '/price-manage' },
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
          meta: { title: '库存管理', icon: 'Box' },
          children: [
            {
              path: 'stocks',
              name: 'Stocks',
              component: () => import('@/views/inventory/StockView.vue'),
              meta: { title: '库存查询', icon: 'Coin' },
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
