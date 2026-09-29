import api from './client'

export interface OtherOutStockItem {
  id?: number
  outStockId?: number
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

export interface OtherOutStock {
  id: number
  billNo: string
  billDate: string
  createdAt: string
  warehouseId: number
  warehouseName?: string
  outType: string
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
  items?: OtherOutStockItem[]
}

export interface OtherOutStockListQuery {
  page?: number
  pageSize?: number
  keyword?: string
  productKw?: string
  warehouseId?: number
  status?: string
  outType?: string
  startDate?: string
  endDate?: string
}

export interface OtherOutStockSavePayload {
  warehouseId: number
  billDate: string
  outType: string
  counterpart?: string
  settleUnit?: string
  handlerId?: number
  deptId?: number
  remark?: string
  items: { productId: number; quantity: number; price: number; remark?: string }[]
  complete: boolean
}

export function fetchOtherOutStocks(params: OtherOutStockListQuery) {
  return api.get('/api/v1/other-out-stocks', { params })
}

export function fetchOtherOutStock(id: number | string) {
  return api.get(`/api/v1/other-out-stocks/${id}`)
}

export function createOtherOutStock(data: OtherOutStockSavePayload) {
  return api.post('/api/v1/other-out-stocks', data)
}

export function updateOtherOutStock(id: number | string, data: OtherOutStockSavePayload) {
  return api.put(`/api/v1/other-out-stocks/${id}`, data)
}

export function deleteOtherOutStock(id: number) {
  return api.delete(`/api/v1/other-out-stocks/${id}`)
}

export function completeOtherOutStock(id: number) {
  return api.put(`/api/v1/other-out-stocks/${id}/complete`)
}

export const OTHER_OUT_TYPES = ['报损出库', '盘亏出库', '其他出库']
