import { describe, it, expect, vi } from 'vitest'
import { useCrud } from './useCrud'

vi.mock('@/api/client', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
  },
}))

describe('useCrud', () => {
  it('returns reactive state', () => {
    const crud = useCrud({ baseUrl: '/test' })
    expect(crud.list.value).toEqual([])
    expect(crud.total.value).toBe(0)
    expect(crud.loading.value).toBe(false)
  })
})
