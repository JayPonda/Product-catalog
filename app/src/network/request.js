function getBaseUrl(prefix, url, queryparams) {
  const finalUrl = new URL(import.meta.env.VITE_BACKEND_URL + prefix + url)
  const params = new URLSearchParams(queryparams)
  finalUrl.search = params.toString()
  return finalUrl.toString()
}

// Cookie-based auth requires credentials to be sent on every request.
const DEFAULT_OPTS = {
  credentials: 'include',
  headers: {
    'Content-Type': 'application/json',
  },
}

const V1 = '/api/v1'

export async function getCategories(page, limit) {
  const offset = page * limit

  try {
    const response = await fetch(getBaseUrl(V1, '/categories', { limit, offset }), DEFAULT_OPTS)

    if (!response.ok) {
      return {
        ok: false,
        error: response.status,
      }
    }

    const data = await response.json()

    return {
      ok: true,
      data: data,
    }
  } catch (error) {
    console.error(error)
    return {
      ok: false,
      error: error,
    }
  }
}

export async function getProducts(page, limit) {
  const offset = page * limit

  try {
    const response = await fetch(getBaseUrl(V1, '/products', { limit, offset }), DEFAULT_OPTS)

    if (!response.ok) {
      return {
        ok: false,
        error: response.status,
      }
    }

    const data = await response.json()

    return {
      ok: true,
      data: data,
    }
  } catch (error) {
    console.error(error)
    return {
      ok: false,
      error: error,
    }
  }
}

export async function createCategory(categoryName) {
  try {
    const response = await fetch(
      getBaseUrl(V1, '/categories'),
      {
        ...DEFAULT_OPTS,
        method: 'POST',
        body: JSON.stringify({ name: categoryName }),
      },
    )

    if (!response.ok) {
      return {
        ok: false,
        error: response.status,
      }
    }

    const data = await response.json()

    return {
      ok: true,
      data: data,
    }
  } catch (error) {
    console.error(error)
    return {
      ok: false,
      error: error,
    }
  }
}

export async function deleteProduct(productId) {
  try {
    const response = await fetch(getBaseUrl(V1, `/products/${productId}`), {
      ...DEFAULT_OPTS,
      method: 'DELETE',
    })

    if (!response.ok) {
      return {
        ok: false,
        error: response.status,
      }
    }

    return {
      ok: true,
      data: 'Product deleted successfully',
    }
  } catch (error) {
    console.error(error)
    return {
      ok: false,
      error: error,
    }
  }
}

export async function getProduct(productId) {
  try {
    const response = await fetch(getBaseUrl(V1, `/products/${productId}`), DEFAULT_OPTS)

    if (!response.ok) {
      return {
        ok: false,
        error: response.status,
      }
    }

    const data = await response.json()

    return {
      ok: true,
      data: data,
    }
  } catch (error) {
    console.error(error)
    return {
      ok: false,
      error: error,
    }
  }
}

export async function createProduct(productInfo) {
  try {
    const response = await fetch(getBaseUrl(V1, `/products/`), {
      ...DEFAULT_OPTS,
      method: 'POST',
      body: JSON.stringify(productInfo),
    })

    if (!response.ok) {
      return {
        ok: false,
        error: response.status,
      }
    }

    const data = await response.json()

    return {
      ok: true,
      data: data,
    }
  } catch (error) {
    console.error(error)
    return {
      ok: false,
      error: error,
    }
  }
}

export async function updateProduct(productId, productInfo) {
  try {
    const response = await fetch(getBaseUrl(V1, `/products/${productId}`), {
      ...DEFAULT_OPTS,
      method: 'PUT',
      body: JSON.stringify(productInfo),
    })

    if (!response.ok) {
      return {
        ok: false,
        error: response.status,
      }
    }

    const data = await response.json()

    return {
      ok: true,
      data: data,
    }
  } catch (error) {
    console.error(error)
    return {
      ok: false,
      error: error,
    }
  }
}

export async function linkCategory(productId, categoryId) {
  try {
    const response = await fetch(getBaseUrl(V1, `/products/${productId}/categories/link`), {
      ...DEFAULT_OPTS,
      method: 'POST',
      body: JSON.stringify({ category_id: categoryId }),
    })

    if (!response.ok) {
      return {
        ok: false,
        error: response.status,
      }
    }

    return {
      ok: true,
      data: 'Category linked successfully',
    }
  } catch (error) {
    console.error(error)
    return {
      ok: false,
      error: error,
    }
  }
}

export async function unlinkCategory(productId, categoryId) {
  try {
    const response = await fetch(getBaseUrl(V1, `/products/${productId}/categories/unlink`), {
      ...DEFAULT_OPTS,
      method: 'POST',
      body: JSON.stringify({ category_id: categoryId }),
    })

    if (!response.ok) {
      return {
        ok: false,
        error: response.status,
      }
    }

    return {
      ok: true,
      data: 'Category unlinked successfully',
    }
  } catch (error) {
    console.error(error)
    return {
      ok: false,
      error: error,
    }
  }
}

export async function searchCategory(categoryletter) {
  try {
    const response = await fetch(
      getBaseUrl(V1, `/categories/match`, { name: categoryletter }),
      DEFAULT_OPTS,
    )

    if (!response.ok) {
      return {
        ok: false,
        error: response.status,
      }
    }

    const data = await response.json()

    return {
      ok: true,
      data: data,
    }
  } catch (error) {
    console.error(error)
    return {
      ok: false,
      error: error,
    }
  }
}

// ----- Auth -----

export async function registerUser(payload) {
  try {
    const response = await fetch(getBaseUrl(V1, '/auth/register'), {
      ...DEFAULT_OPTS,
      method: 'POST',
      body: JSON.stringify(payload),
    })

    if (!response.ok) {
      const data = await response.json().catch(() => ({}))
      return {
        ok: false,
        error: response.status,
        message: data.error ?? 'Registration failed',
      }
    }

    const data = await response.json()
    return {
      ok: true,
      data: data,
    }
  } catch (error) {
    console.error(error)
    return {
      ok: false,
      error: error,
    }
  }
}

export async function loginUser(payload) {
  try {
    const response = await fetch(getBaseUrl(V1, '/auth/login'), {
      ...DEFAULT_OPTS,
      method: 'POST',
      body: JSON.stringify(payload),
    })

    if (!response.ok) {
      const data = await response.json().catch(() => ({}))
      return {
        ok: false,
        error: response.status,
        message: data.error ?? 'Login failed',
      }
    }

    const data = await response.json()
    return {
      ok: true,
      data: data,
    }
  } catch (error) {
    console.error(error)
    return {
      ok: false,
      error: error,
    }
  }
}

export async function logoutUser() {
  try {
    const response = await fetch(getBaseUrl(V1, '/auth/logout'), {
      ...DEFAULT_OPTS,
      method: 'POST',
    })

    if (!response.ok) {
      return {
        ok: false,
        error: response.status,
      }
    }

    return {
      ok: true,
    }
  } catch (error) {
    console.error(error)
    return {
      ok: false,
      error: error,
    }
  }
}

export async function getCurrentUser() {
  try {
    const response = await fetch(getBaseUrl(V1, '/auth/me'), DEFAULT_OPTS)

    if (!response.ok) {
      return {
        ok: false,
        error: response.status,
      }
    }

    const data = await response.json()
    return {
      ok: true,
      data: data,
    }
  } catch (error) {
    console.error(error)
    return {
      ok: false,
      error: error,
    }
  }
}
