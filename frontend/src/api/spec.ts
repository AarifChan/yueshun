import api from './client'

export interface ProductSpec {
  id: number
  name: string
  values: string
  remark: string
  sort: number
  status: number
}

export interface SpecPayload {
  name: string
  values: string
  sort: number
  status: number
}

export function fetchSpecs(params: { keyword?: string }) {
  return api.get('/api/v1/product-settings/specs', {
    params: { ...params, page: 1, pageSize: 500 },
  })
}

export function createSpec(data: SpecPayload) {
  return api.post('/api/v1/product-settings/specs', data)
}

export function updateSpec(id: number, data: SpecPayload) {
  return api.put(`/api/v1/product-settings/specs/${id}`, data)
}

export function deleteSpec(id: number) {
  return api.delete(`/api/v1/product-settings/specs/${id}`)
}

export function sortSpecs(items: { id: number; sort: number }[]) {
  return api.post('/api/v1/product-settings/specs/sort', { items })
}
