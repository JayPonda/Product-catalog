// Cookie-based auth requires credentials to be sent on every request.
const DEFAULT_OPTS = {
  credentials: 'include',
  headers: {
    'Content-Type': 'application/json',
  },
}

const V1 = '/api/v1'

function buildUrl(path, params) {
  const finalUrl = new URL(import.meta.env.VITE_BACKEND_URL + V1 + path)
  finalUrl.search = new URLSearchParams(params).toString()
  return finalUrl.toString()
}

async function request(path, { params, rawData, defaultError, ...options } = {}) {
  try {
    const response = await fetch(buildUrl(path, params), { ...DEFAULT_OPTS, ...options })

    if (!response.ok) {
      if (defaultError === undefined) {
        return { ok: false, error: response.status }
      }
      const body = await response.json().catch(() => ({}))
      return { ok: false, error: response.status, message: body.error ?? defaultError }
    }

    if (rawData !== undefined) {
      return { ok: true, data: rawData }
    }

    return { ok: true, data: await response.json() }
  } catch (error) {
    console.error(error)
    return { ok: false, error }
  }
}

export function getCategories(page, limit) {
  const offset = page * limit
  return request('/categories', { params: { limit, offset } })
}

export function getProducts(page, limit) {
  const offset = page * limit
  return request('/products', { params: { limit, offset } })
}

export function getMyProducts(page, limit) {
  const offset = page * limit
  return request('/my-products', { params: { limit, offset } })
}

export function searchCategory(name) {
  return request('/categories/match', { params: { name } })
}

export function createCategory(categoryName) {
  return request('/categories', {
    method: 'POST',
    body: JSON.stringify({ name: categoryName }),
  })
}

export function getProduct(productId) {
  return request(`/products/${productId}`)
}

export function createProduct(productInfo) {
  return request('/products/', {
    method: 'POST',
    body: JSON.stringify(productInfo),
  })
}

export function updateProduct(productId, productInfo) {
  return request(`/products/${productId}`, {
    method: 'PUT',
    body: JSON.stringify(productInfo),
  })
}

export function deleteProduct(productId) {
  return request(`/products/${productId}`, {
    method: 'DELETE',
    rawData: 'Product deleted successfully',
  })
}

export function linkCategory(productId, categoryId) {
  return request(`/products/${productId}/categories/link`, {
    method: 'POST',
    body: JSON.stringify({ category_id: categoryId }),
    rawData: 'Category linked successfully',
  })
}

export function unlinkCategory(productId, categoryId) {
  return request(`/products/${productId}/categories/unlink`, {
    method: 'POST',
    body: JSON.stringify({ category_id: categoryId }),
    rawData: 'Category unlinked successfully',
  })
}

export function registerUser(payload) {
  return request('/auth/register', {
    method: 'POST',
    body: JSON.stringify(payload),
    defaultError: 'Registration failed',
  })
}

export function loginUser(payload) {
  return request('/auth/login', {
    method: 'POST',
    body: JSON.stringify(payload),
    defaultError: 'Login failed',
  })
}

export function logoutUser() {
  return request('/auth/logout', { method: 'POST', rawData: null })
}

export function getCurrentUser() {
  return request('/auth/me')
}
