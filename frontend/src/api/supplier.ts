import api from './client'

export interface Supplier {
  id: number
  name: string
  code: string
  contact: string
  phone: string
  status: number
}

export interface SupplierPayload {
  name: string
  code: string
  contact: string
  phone: string
  status: number
}

export function fetchSuppliers(params: { keyword?: string; page: number; pageSize: number }) {
  return api.get('/api/v1/suppliers', { params })
}

export function createSupplier(data: SupplierPayload) {
  return api.post('/api/v1/suppliers', data)
}

export function updateSupplier(id: number, data: SupplierPayload) {
  return api.put(`/api/v1/suppliers/${id}`, data)
}

export function deleteSupplier(id: number) {
  return api.delete(`/api/v1/suppliers/${id}`)
}
