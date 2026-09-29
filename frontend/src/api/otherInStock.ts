import api from './client'

export interface OtherInStockItem {
  id?: number
  inStockId?: number
  productId: number | undefined
  productName?: string
  productCode?: string
  barcode?: string
  specification?: string
  unit?: string
  image?: string
  stockQty?: number
  quantity: number
  price: number
  amount?: number
  remark?: string
  product?: {
    id: number
    name: string
    code: string
    barcode: string
    specification: string
    unit: string
    image: string
  }
}

export interface OtherInStock {
  id: number
  billNo: string
  billDate: string
  createdAt: string
  warehouseId: number
  warehouseName?: string
  inType: string
  counterpart: string
  settleUnit: string
  handlerId: number
  handlerName?: string
  deptId: number
  deptName?: string
  totalQty: number
  amount: number
  status: string
  operatorId: number
  operatorName?: string
  remark: string
  items?: OtherInStockItem[]
}

export interface OtherInStockListQuery {
  page?: number
  pageSize?: number
  keyword?: string
  productKw?: string
  warehouseId?: number
  status?: string
  inType?: string
  startDate?: string
  endDate?: string
}

export interface OtherInStockSavePayload {
  warehouseId: number
  billDate: string
  inType: string
  counterpart?: string
  settleUnit?: string
  handlerId?: number
  deptId?: number
  remark?: string
  items: { productId: number; quantity: number; price: number; remark?: string }[]
  complete: boolean
}

export function fetchOtherInStocks(params: OtherInStockListQuery) {
  return api.get('/api/v1/other-in-stocks', { params })
}

export function fetchOtherInStock(id: number | string) {
  return api.get(`/api/v1/other-in-stocks/${id}`)
}

export function createOtherInStock(data: OtherInStockSavePayload) {
  return api.post('/api/v1/other-in-stocks', data)
}

export function updateOtherInStock(id: number | string, data: OtherInStockSavePayload) {
  return api.put(`/api/v1/other-in-stocks/${id}`, data)
}

export function deleteOtherInStock(id: number) {
  return api.delete(`/api/v1/other-in-stocks/${id}`)
}

export function completeOtherInStock(id: number) {
  return api.put(`/api/v1/other-in-stocks/${id}/complete`)
}

export const OTHER_IN_TYPES = ['报溢入库', '盘盈入库', '期初入库', '其他入库']
