import api from './client'

export interface ProductTag {
  id: number
  name: string
  color: string
  type: number
  rule: string
  filterable: number
  showInList: number
  showBadge: number
  sort: number
  status: number
  productCount?: number
}

export interface TagPayload {
  name: string
  color: string
  type: number
  rule: string
  filterable: number
  showInList: number
  showBadge: number
  sort: number
  status: number
}

export function fetchTags(params: { keyword?: string } = {}) {
  return api.get('/api/v1/product-settings/tags', {
    params: { ...params, page: 1, pageSize: 500 },
  })
}

export function createTag(data: TagPayload) {
  return api.post('/api/v1/product-settings/tags', data)
}

export function updateTag(id: number, data: TagPayload) {
  return api.put(`/api/v1/product-settings/tags/${id}`, data)
}

export function deleteTag(id: number) {
  return api.delete(`/api/v1/product-settings/tags/${id}`)
}

export function sortTags(items: { id: number; sort: number }[]) {
  return api.post('/api/v1/product-settings/tags/sort', { items })
}
