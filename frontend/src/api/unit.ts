import api from './client'

export interface GoodsUnit {
  id: number
  name: string
  code: string
  sort: number
  status: number
  storageType: number
}

export interface UnitPayload {
  name: string
  sort: number
  status: number
  storageType: number
}

export function fetchUnits(params: { keyword?: string; storageType?: number }) {
  return api.get('/api/v1/product-settings/units', {
    params: { ...params, page: 1, pageSize: 500 },
  })
}

export function createUnit(data: UnitPayload) {
  return api.post('/api/v1/product-settings/units', data)
}

export function updateUnit(id: number, data: UnitPayload) {
  return api.put(`/api/v1/product-settings/units/${id}`, data)
}

export function deleteUnit(id: number) {
  return api.delete(`/api/v1/product-settings/units/${id}`)
}

export function sortUnits(items: { id: number; sort: number }[]) {
  return api.post('/api/v1/product-settings/units/sort', { items })
}
