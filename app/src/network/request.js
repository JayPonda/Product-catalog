import logger from '@/utils/logger'

// Cookie-based auth requires credentials to be sent on every request.
const DEFAULT_OPTS = {
  credentials: 'include',
  headers: {
    'Content-Type': 'application/json',
  },
}

const V1 = '/api/v1'

function buildUrl(path, params) {
  const finalUrl = new URL(import.meta.env.VITE_BACKEND_URL + V1 + path, window.location.origin)
  finalUrl.search = new URLSearchParams(params).toString()
  return finalUrl.toString()
}

function getStatusMessage(status) {
  switch (status) {
    case 400:
      return 'Bad request.'
    case 401:
      return 'Unauthorized.'
    case 403:
      return 'Access forbidden.'
    case 404:
      return 'Resource not found.'
    case 409:
      return 'Conflict.'
    case 422:
      return 'Validation error.'
    case 429:
      return 'Too many requests. Please try again later.'
    case 500:
      return 'Internal server error. Please try again later.'
    case 502:
      return 'Bad gateway. Please try again later.'
    case 503:
      return 'Service unavailable. Please try again later.'
    case 504:
      return 'Gateway timeout. Please try again later.'
    default:
      return 'Something went wrong.'
  }
}

async function request(path, { params, rawData, defaultError, method, ...options } = {}) {
  const httpMethod = method || 'GET'
  const url = buildUrl(path, params)

  logger.Debug('request.js', 'request', `${httpMethod} ${path}`, { params })

  const fetchOpts = { ...DEFAULT_OPTS, ...options }
  if (method) {
    fetchOpts.method = method
  }

  try {
    const response = await fetch(url, fetchOpts)
    const correlationId = response.headers.get('X-Request-ID')

    if (!response.ok) {
      let body = null
      try {
        body = await response.json()
      } catch {
        try {
          const text = await response.text()
          if (text && !text.trim().startsWith('<')) {
            body = text.trim()
          }
        } catch {
          body = null
        }
      }

      let backendMessage = ''
      if (typeof body === 'object' && body !== null) {
        backendMessage =
          (typeof body.message === 'string' && body.message.trim()) ||
          (typeof body.error === 'string' && body.error.trim()) ||
          (typeof body.details === 'string' && body.details.trim()) ||
          ''
      } else if (typeof body === 'string') {
        backendMessage = body.trim()
      }

      const message = backendMessage || defaultError || getStatusMessage(response.status)

      logger.Warn(
        'request.js',
        'request',
        `${httpMethod} ${path} failed`,
        { status: response.status, message },
        correlationId,
      )
      return { ok: false, error: response.status, message }
    }

    if (rawData !== undefined) {
      logger.Debug('request.js', 'request', `${httpMethod} ${path} success`, { correlationId })
      return { ok: true, data: rawData }
    }

    const data = await response.json()
    logger.Debug('request.js', 'request', `${httpMethod} ${path} success`, { correlationId })
    return { ok: true, data }
  } catch (error) {
    logger.Error('request.js', 'request', `${httpMethod} ${path} error`, {
      path,
      error: error.message,
    })
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
