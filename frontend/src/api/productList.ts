import api from './client'
import { ensureXlsxFile } from '@/utils/spreadsheet'

export interface ProductListItem {
  id: number
  code: string
  barcode: string
  name: string
  image?: string
  specCount?: number
  tags?: string[]
  categoryName: string
  brandName: string
  unit: string
  status: number
  mallName?: string
  topicCategory?: string
  pinyinCode?: string
  purchasePrice?: number
  retailPrice?: number
  defaultPrice?: number
  description?: string
  totalStock?: number
  createdAt?: string
  mallSortWeight?: number
  searchKeywords?: string
  salesQty?: number
  weight?: number
  volume?: number
  warehouseName?: string
  supplierName?: string
}

export interface SpecListItem {
  id: number
  productId: number
  specValue: string
  code: string
  barcode: string
  onShelf: number
  productName: string
  productImage: string
  brandName: string
  categoryName: string
  unit: string
}

export interface ProductQuery {
  keyword?: string
  name?: string
  categoryId?: number
  brandId?: number
  tagId?: number
  topicCategory?: string
  stockStatus?: string
  hasImage?: number
  status?: number | ''
  orderBy?: string
  order?: 'asc' | 'desc'
  page: number
  pageSize: number
}

export interface ImportErrorItem {
  row: number
  code: string
  reason: string
}

export interface ImportResult {
  created: number
  updated?: number
  failed: number
  errors: ImportErrorItem[]
}

export function fetchSettings() {
  return api.get('/api/v1/settings')
}

export function saveSettings(data: Record<string, string>) {
  return api.put('/api/v1/settings', data)
}

export function fetchProducts(params: ProductQuery) {
  return api.get('/api/v1/products', { params })
}

export function fetchSpecItems(params: ProductQuery) {
  return api.get('/api/v1/products/spec-items', { params })
}

export function deleteProduct(id: number) {
  return api.delete(`/api/v1/products/${id}`)
}

export async function importProductsSystem(file: File) {
  const fd = new FormData()
  fd.append('file', await ensureXlsxFile(file))
  return api.post('/api/v1/products/import/system', fd)
}

export async function importProductsCustom(file: File, overwriteEmpty: boolean) {
  const fd = new FormData()
  fd.append('file', await ensureXlsxFile(file))
  fd.append('overwriteEmpty', overwriteEmpty ? '1' : '0')
  return api.post('/api/v1/products/import/custom', fd)
}

/** 批量更新商品排序权重 / 搜索关键词（商品排序设置、完善搜索关键词） */
export function batchUpdateProductFields(items: { id: number; mallSortWeight?: number; searchKeywords?: string }[]) {
  return api.put('/api/v1/products/batch-fields', { items })
}
