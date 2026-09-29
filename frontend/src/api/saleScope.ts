import api from './client'

export type ProductDim = 'category' | 'brand' | 'supplier'
export type CustomerDim = 'levelPrice' | 'customerCategory' | 'region' | 'customerTag'
export type SaleRuleTargetType = ProductDim | 'product'

export interface SaleScopeDimension {
  productDim: ProductDim
  customerDim: CustomerDim
}

export interface SaleRuleContent {
  levelPriceIds: number[]
  customerCategoryIds: number[]
  regionIds: number[]
  customerTagIds: number[]
  customerIds: number[]
}

export interface SaleRuleSummary {
  levelPrices: string
  customerCategories: string
  regions: string
  customerTags: string
  customers: string
}

export interface SaleRuleRow {
  targetType: SaleRuleTargetType
  targetId: number
  name: string
  hasChildren: boolean
  rule: SaleRuleContent | null
  ruleSummary: SaleRuleSummary | null
  children?: SaleRuleRow[]
}

export function emptyRule(): SaleRuleContent {
  return { levelPriceIds: [], customerCategoryIds: [], regionIds: [], customerTagIds: [], customerIds: [] }
}

export function getSaleScopeDimension() {
  return api.get('/api/v1/product-settings/sale-rules/dimension')
}

export function setSaleScopeDimension(data: SaleScopeDimension) {
  return api.post('/api/v1/product-settings/sale-rules/dimension', data)
}

export function fetchSaleRuleTree(params: { keyword?: string; parentId?: number } = {}) {
  return api.get('/api/v1/product-settings/sale-rules/tree', { params })
}

export function fetchSaleRuleProducts(params: { categoryId?: number; keyword?: string } = {}) {
  return api.get('/api/v1/product-settings/sale-rules/products', { params })
}

export function saveSaleRule(data: { targetType: SaleRuleTargetType; targetId: number; rule: SaleRuleContent }) {
  return api.put('/api/v1/product-settings/sale-rules', data)
}
