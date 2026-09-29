import api from './client'

export interface ProductUnitItem {
  id?: number
  name: string
  conversion: number
  isDefault: boolean
  barcode: string
  allowSale: boolean
  defaultSale: boolean
  defaultPurchase: boolean
  storageType: number
  refPrice: number
  wholesalePrice?: number
  retailPrice?: number
  minSalePrice?: number
  defaultPrice?: number
}

export interface ProductSpecItem {
  specValue: string
  code: string
  barcode: string
  weight: number
  volume: number
  remark1: string
  remark2: string
  onShelf: boolean
  sort: number
}

export interface OpeningStockItem {
  warehouseId: number
  quantity: number
}

export interface ProductPayload {
  name: string
  mallName: string
  categoryId: number
  brandId?: number
  topicCategory: string
  pinyinCode: string
  code: string
  barcode: string
  unit: string
  specification: string
  purchasePrice: number
  wholesalePrice: number
  retailPrice: number
  minStock: number
  maxStock: number
  description: string
  image: string
  videoUrl: string
  detailContent: string
  minOrderQty: number
  maxOrderQty: number
  orderByMultiple: boolean
  noAvailableStockControl: boolean
  noBookStockControl: boolean
  manualLevelPriceDefault: boolean
  auxPriceReverseCalc: boolean
  minSalePrice?: number
  mallSortWeight?: number
  searchKeywords?: string
  warehouseId?: number
  supplierId?: number
  forbidPurchaseUnits?: string
  status: number
}

export interface Warehouse {
  id: number
  name: string
  code: string
  status: number
}

export interface GoodsUnitOption {
  id: number
  name: string
  storageType: number
}

export function fetchProduct(id: number | string) {
  return api.get(`/api/v1/products/${id}`)
}

export function createProduct(data: ProductPayload) {
  return api.post('/api/v1/products', data)
}

export function updateProduct(id: number | string, data: ProductPayload) {
  return api.put(`/api/v1/products/${id}`, data)
}

export function replaceProductUnits(id: number | string, units: ProductUnitItem[]) {
  return api.put(`/api/v1/products/${id}/units`, { units })
}

export function replaceProductSpecs(id: number | string, items: ProductSpecItem[]) {
  return api.put(`/api/v1/products/${id}/specs`, { items })
}

export function saveOpeningStock(productId: number | string, items: OpeningStockItem[]) {
  return api.post('/api/v1/products/opening-stock', { productId, items })
}

export function fetchTopicCategories() {
  return api.get('/api/v1/products/topic-categories')
}

export function fetchBrandOptions() {
  return api.get('/api/v1/products/brands', { params: { page: 1, pageSize: 500 } })
}

export function fetchGoodsUnitOptions() {
  return api.get('/api/v1/product-settings/units', { params: { page: 1, pageSize: 500 } })
}

export function fetchWarehouseOptions() {
  return api.get('/api/v1/warehouses', { params: { page: 1, pageSize: 500 } })
}

export function uploadFile(file: File) {
  const fd = new FormData()
  fd.append('file', file)
  return api.post('/api/v1/upload', fd)
}
