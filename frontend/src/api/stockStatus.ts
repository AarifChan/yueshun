import api from './client'
import { ensureXlsxFile } from '@/utils/spreadsheet'

export interface StockStatusQuery {
  page?: number
  pageSize?: number
  keyword?: string
  warehouseId?: number
  categoryId?: number
  stockStatus?: string
  productStatus?: string
  onlyWithStock?: boolean
  groupBy?: 'product' | 'category'
}

export interface StockStatusItem {
  productId: number
  productName: string
  code: string
  specification: string
  unit: string
  image: string
  categoryId: number
  categoryName: string
  costPrice: number
  minStock: number
  maxStock: number
  status: number
  quantity: number
  costAmount: number
  reservedQty: number
  pendingOutQty: number
  availableQty: number
}

export interface StockStatusCategoryItem {
  categoryId: number
  categoryName: string
  productCount: number
  quantity: number
  costAmount: number
}

export interface StockStatusSummary {
  totalQuantity: number
  totalAmount: number
}

export function fetchStockStatus(params: StockStatusQuery) {
  return api.get('/api/v1/stocks/status', { params })
}

export function updateStockLimits(data: { productIds: number[]; minStock: number; maxStock: number }) {
  return api.put('/api/v1/stocks/status/limits', data)
}

export interface StockFlowItem {
  billDate: string
  billNo: string
  billType: string
  warehouseId: number
  warehouseName: string
  inQty: number
  outQty: number
  price: number
}

export interface StockDistributionItem {
  warehouseId: number
  warehouseName: string
  quantity: number
  costAmount: number
}

export function fetchStockFlows(productId: number, params: { page: number; pageSize: number }) {
  return api.get(`/api/v1/stocks/status/${productId}/flows`, { params })
}

export function fetchStockDistribution(productId: number) {
  return api.get(`/api/v1/stocks/status/${productId}/warehouses`)
}

export interface StockImportResult {
  warehouse: string
  updated: number
  created: number
  stocked: number
  skipped: number
}

export async function importStockStatus(file: File, warehouseId?: number) {
  const fd = new FormData()
  fd.append('file', await ensureXlsxFile(file))
  if (warehouseId) fd.append('warehouseId', String(warehouseId))
  return api.post('/api/v1/stocks/import', fd)
}
