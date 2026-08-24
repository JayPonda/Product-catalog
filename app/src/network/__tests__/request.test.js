import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import {
  getCategories,
  getProducts,
  getMyProducts,
  createCategory,
  deleteProduct,
  getProduct,
  createProduct,
  updateProduct,
  linkCategory,
  unlinkCategory,
  searchCategory,
  registerUser,
  loginUser,
  logoutUser,
  getCurrentUser,
} from '../request.js'

const BACKEND = 'http://backend.test'

function jsonResponse(body, status = 200) {
  return {
    ok: status >= 200 && status < 300,
    status,
    json: async () => body,
  }
}

describe('network/request', () => {
  let fetchMock

  beforeEach(() => {
    vi.stubEnv('VITE_BACKEND_URL', BACKEND)
    fetchMock = vi.fn()
    vi.stubGlobal('fetch', fetchMock)
  })

  afterEach(() => {
    vi.unstubAllEnvs()
    vi.unstubAllGlobals()
  })

  describe('pagination endpoints', () => {
    it.each([
      ['getCategories', getCategories, '/categories'],
      ['getProducts', getProducts, '/products'],
      ['getMyProducts', getMyProducts, '/my-products'],
    ])('%s sends offset = page * limit and parses data', async (_name, fn, path) => {
      fetchMock.mockResolvedValue(jsonResponse({ total: 1 }))

      const res = await fn(2, 50)

      expect(res.ok).toBe(true)
      expect(res.data).toEqual({ total: 1 })

      const [url, opts] = fetchMock.mock.calls[0]
      expect(url).toBe(`${BACKEND}/api/v1${path}?limit=50&offset=100`)
      expect(opts.method).toBeUndefined()
      expect(opts.credentials).toBe('include')
      expect(opts.headers['Content-Type']).toBe('application/json')
    })

    it.each([
      ['getCategories', getCategories],
      ['getProducts', getProducts],
      ['getMyProducts', getMyProducts],
    ])('%s reports non-2xx as { ok:false, error }', async (_name, fn) => {
      fetchMock.mockResolvedValue(jsonResponse({}, 500))

      const res = await fn(0, 20)

      expect(res.ok).toBe(false)
      expect(res.error).toBe(500)
    })
  })

  describe('category writes', () => {
    it('createCategory posts a JSON body with the trimmed name left to backend', async () => {
      fetchMock.mockResolvedValue(jsonResponse({ id: 'c1', name: 'tools' }))

      const res = await createCategory('tools')

      expect(res.ok).toBe(true)
      const [url, opts] = fetchMock.mock.calls[0]
      expect(url).toBe(`${BACKEND}/api/v1/categories`)
      expect(opts.method).toBe('POST')
      expect(JSON.parse(opts.body)).toEqual({ name: 'tools' })
    })

    it('searchCategory queries by name', async () => {
      fetchMock.mockResolvedValue(jsonResponse({ categories: [] }))

      await searchCategory('to')

      const [url] = fetchMock.mock.calls[0]
      expect(url).toBe(`${BACKEND}/api/v1/categories/match?name=to`)
    })
  })

  describe('product writes', () => {
    it('createProduct POSTs product info', async () => {
      fetchMock.mockResolvedValue(jsonResponse({ id: 'p1' }))
      const payload = { name: 'Chair', description: 'x', price: 100, stock_quantity: 2 }

      const res = await createProduct(payload)

      expect(res.ok).toBe(true)
      const [url, opts] = fetchMock.mock.calls[0]
      expect(url).toBe(`${BACKEND}/api/v1/products/`)
      expect(opts.method).toBe('POST')
      expect(JSON.parse(opts.body)).toEqual(payload)
    })

    it('updateProduct PUTs to /products/:id', async () => {
      fetchMock.mockResolvedValue(jsonResponse({ id: 'p1' }))

      await updateProduct('p1', { name: 'New' })

      const [url, opts] = fetchMock.mock.calls[0]
      expect(url).toBe(`${BACKEND}/api/v1/products/p1`)
      expect(opts.method).toBe('PUT')
    })

    it('deleteProduct DELETEs and needs no JSON parse', async () => {
      fetchMock.mockResolvedValue({ ok: true, status: 204 })

      const res = await deleteProduct('p9')

      expect(res.ok).toBe(true)
      const [url, opts] = fetchMock.mock.calls[0]
      expect(url).toBe(`${BACKEND}/api/v1/products/p9`)
      expect(opts.method).toBe('DELETE')
    })

    it('getProduct GETs /products/:id', async () => {
      fetchMock.mockResolvedValue(jsonResponse({ id: 'p1', name: 'Chair' }))

      const res = await getProduct('p1')

      expect(res.ok).toBe(true)
      expect(res.data.name).toBe('Chair')
      const [url] = fetchMock.mock.calls[0]
      expect(url).toBe(`${BACKEND}/api/v1/products/p1`)
    })

    it.each([
      ['linkCategory', linkCategory, 'link', 'Category linked successfully'],
      ['unlinkCategory', unlinkCategory, 'unlink', 'Category unlinked successfully'],
    ])('%s POSTs category_id to the %s route', async (_name, fn, segment, msg) => {
      fetchMock.mockResolvedValue({ ok: true, status: 200 })

      const res = await fn('p1', 'c1')

      expect(res.ok).toBe(true)
      expect(res.data).toBe(msg)
      const [url, opts] = fetchMock.mock.calls[0]
      expect(url).toBe(`${BACKEND}/api/v1/products/p1/categories/${segment}`)
      expect(opts.method).toBe('POST')
      expect(JSON.parse(opts.body)).toEqual({ category_id: 'c1' })
    })
  })

  describe('auth endpoints', () => {
    it('registerUser resolves user data on success', async () => {
      fetchMock.mockResolvedValue(jsonResponse({ user: { email: 'a@b.c' } }, 201))

      const res = await registerUser({ email: 'a@b.c' })

      expect(res.ok).toBe(true)
      expect(res.data.user.email).toBe('a@b.c')
    })

    it('registerUser surfaces the backend error message on failure', async () => {
      fetchMock.mockResolvedValue(jsonResponse({ error: 'email already registered' }, 409))

      const res = await registerUser({})

      expect(res.ok).toBe(false)
      expect(res.error).toBe(409)
      expect(res.message).toBe('email already registered')
    })

    it('loginUser falls back to a default message when body is unreadable', async () => {
      fetchMock.mockResolvedValue({
        ok: false,
        status: 401,
        json: async () => {
          throw new Error('no json')
        },
      })

      const res = await loginUser({})

      expect(res.ok).toBe(false)
      expect(res.message).toBe('Login failed')
    })

    it('logoutUser POSTs without a body', async () => {
      fetchMock.mockResolvedValue({ ok: true, status: 204 })

      const res = await logoutUser()

      expect(res.ok).toBe(true)
      expect(fetchMock.mock.calls[0][1].method).toBe('POST')
    })

    it('getCurrentUser GETs /auth/me', async () => {
      fetchMock.mockResolvedValue(jsonResponse({ user: { email: 'a@b.c' } }))

      const res = await getCurrentUser()

      expect(res.ok).toBe(true)
      expect(fetchMock.mock.calls[0][0]).toBe(`${BACKEND}/api/v1/auth/me`)
    })
  })

  describe('transport failures', () => {
    it('returns { ok:false, error } when fetch rejects', async () => {
      fetchMock.mockRejectedValue(new TypeError('network down'))

      const res = await getCategories(0, 20)

      expect(res.ok).toBe(false)
      expect(res.error).toBeInstanceOf(TypeError)
    })
  })
})
