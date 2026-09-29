import api from './client'

export interface Brand {
  id: number
  name: string
  code: string
  description: string
  image: string
  categoryId: number
  categoryName: string
  applyImage: number
  sort: number
  status: number
  productCount: number
}

export interface BrandPayload {
  name: string
  code: string
  description: string
  image: string
  categoryId: number
  applyImage: number
  sort: number
  status: number
}

export function fetchBrands(params: { page: number; pageSize: number; keyword?: string; categoryId?: number }) {
  return api.get('/api/v1/products/brands', { params })
}

export function createBrand(data: BrandPayload) {
  return api.post('/api/v1/products/brands', data)
}

export function updateBrand(id: number, data: BrandPayload) {
  return api.put(`/api/v1/products/brands/${id}`, data)
}

export function deleteBrand(id: number) {
  return api.delete(`/api/v1/products/brands/${id}`)
}

export function sortBrands(items: { id: number; sort: number }[]) {
  return api.post('/api/v1/products/brands/sort', { items })
}

export function uploadFile(file: File) {
  const fd = new FormData()
  fd.append('file', file)
  return api.post('/api/v1/upload', fd)
}
