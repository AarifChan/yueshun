import api from './client'

export interface PriceManageItem {
  id: number
  name: string
  code: string
  unit: string
  defaultPrice: number
  wholesalePrice: number
  retailPrice: number
  totalStock: number
  status: number
}

export interface PriceManageQuery {
  page: number
  pageSize: number
  keyword?: string
  categoryId?: number
  status?: number
  stockStatus?: string
  priceType?: string
  priceOp?: string
  priceValue?: string
}

export interface PriceImportError {
  row: number
  code: string
  reason: string
}

export interface PriceImportResult {
  updated: number
  failed: number
  errors: PriceImportError[]
}

export function fetchPriceManageList(params: PriceManageQuery) {
  return api.get('/api/v1/prices/manage', { params })
}

export function updatePriceManage(id: number, data: { wholesalePrice?: number; retailPrice?: number; defaultPrice?: number }) {
  return api.put(`/api/v1/prices/manage/${id}`, data)
}

export function importPriceManage(file: File) {
  const fd = new FormData()
  fd.append('file', file)
  return api.post('/api/v1/prices/manage/import', fd)
}

export function fetchSettings() {
  return api.get('/api/v1/settings')
}

export function saveSettings(data: Record<string, string>) {
  return api.put('/api/v1/settings', data)
}
