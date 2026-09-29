import api from './client'

export interface CategoryNode {
  id: number
  parentId: number
  name: string
  code: string
  sort: number
  status: number
  image: string
  productCount: number
  hasChildren?: boolean
  children: CategoryNode[] | null
}

export interface CategoryPayload {
  parentId: number
  name: string
  code: string
  sort: number
  status: number
  image: string
}

export function fetchCategoryTree(parentId?: number) {
  return api.get('/api/v1/products/categories/tree', {
    params: parentId === undefined ? {} : { parentId },
  })
}

export function createCategory(data: CategoryPayload) {
  return api.post('/api/v1/products/categories', data)
}

export function updateCategory(id: number, data: CategoryPayload) {
  return api.put(`/api/v1/products/categories/${id}`, data)
}

export function deleteCategory(id: number) {
  return api.delete(`/api/v1/products/categories/${id}`)
}

export function updateCategoryStatus(id: number, status: number) {
  return api.put(`/api/v1/products/categories/${id}/status`, { status })
}

export function sortCategories(items: { id: number; sort: number }[]) {
  return api.post('/api/v1/products/categories/sort', { items })
}

export function fillCategoryImages() {
  return api.post('/api/v1/products/categories/fill-images')
}

export function uploadFile(file: File) {
  const fd = new FormData()
  fd.append('file', file)
  return api.post('/api/v1/upload', fd)
}
