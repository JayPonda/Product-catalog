import { describe, expect, it, vi, beforeEach } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { flushPromises } from '@vue/test-utils'

const getCurrentUser = vi.fn()
const logoutUser = vi.fn()

vi.mock('@/network/request.js', () => ({
  getCurrentUser: (...args) => getCurrentUser(...args),
  logoutUser: (...args) => logoutUser(...args),
  loginUser: vi.fn(),
  registerUser: vi.fn(),
  getCategories: vi.fn(),
  getProducts: vi.fn(),
}))

async function makeRouter() {
  // Fresh import per call so the guard's module-level authInitialized flag resets.
  vi.resetModules()
  const { default: router } = await import('@/router/index')
  await router.push('/')
  await router.isReady()
  return router
}

describe('router auth guard', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    getCurrentUser.mockReset()
    logoutUser.mockReset()
  })

  it('restores the session once and allows protected routes when authenticated', async () => {
    getCurrentUser.mockResolvedValue({ ok: true, data: { user: { email: 'a@b.c' } } })

    const router = await makeRouter()
    await router.push('/my-products')

    expect(router.currentRoute.value.path).toBe('/my-products')
    expect(getCurrentUser).toHaveBeenCalledTimes(1)

    // Subsequent navigations must NOT re-fetch the session.
    await router.push('/categories')
    expect(getCurrentUser).toHaveBeenCalledTimes(1)
  })

  it('redirects unauthenticated users from protected routes to login with a redirect query', async () => {
    getCurrentUser.mockResolvedValue({ ok: false, error: 401 })

    const router = await makeRouter()
    await router.push('/my-products')

    expect(router.currentRoute.value.name).toBe('login')
    expect(router.currentRoute.value.query.redirect).toBe('/my-products')
  })

  it.each(['/products/add', '/products/some-id/edit'])('protects %s', async (path) => {
    getCurrentUser.mockResolvedValue({ ok: false, error: 401 })

    const router = await makeRouter()
    await router.push(path)

    expect(router.currentRoute.value.name).toBe('login')
  })

  it('leaves public routes open when unauthenticated', async () => {
    getCurrentUser.mockResolvedValue({ ok: false, error: 401 })

    const router = await makeRouter()
    await router.push('/products')

    expect(router.currentRoute.value.path).toBe('/products')
  })

  it('keeps the intended destination for edit-route redirects', async () => {
    getCurrentUser.mockResolvedValue({ ok: false, error: 401 })

    const router = await makeRouter()
    await router.push('/products/abc-123/edit')

    expect(router.currentRoute.value.query.redirect).toBe('/products/abc-123/edit')
  })

  it('scrolls to hash if element is found immediately', async () => {
    getCurrentUser.mockResolvedValue({ ok: true, data: { user: { email: 'a@b.c' } } })
    const router = await makeRouter()

    const fakeEl = { getBoundingClientRect: () => ({ top: 100 }) }
    const spyQuerySelector = vi.spyOn(document, 'querySelector').mockReturnValue(fakeEl)

    await router.push({ path: '/products', hash: '#anchor-1' })
    await flushPromises()

    expect(spyQuerySelector).toHaveBeenCalledWith('#anchor-1')
    spyQuerySelector.mockRestore()
  })

  it('scrolls to hash using waitForElement polling when element appears later', async () => {
    vi.useFakeTimers()
    getCurrentUser.mockResolvedValue({ ok: true, data: { user: { email: 'a@b.c' } } })
    const router = await makeRouter()

    let callCount = 0
    const fakeEl = { getBoundingClientRect: () => ({ top: 100, left: 0 }) }
    const spyQuerySelector = vi.spyOn(document, 'querySelector').mockImplementation(() => {
      callCount++
      return callCount > 2 ? fakeEl : null
    })

    router.push({ path: '/products', hash: '#delayed-anchor' })
    await vi.advanceTimersByTimeAsync(150)
    await flushPromises()

    expect(spyQuerySelector).toHaveBeenCalledWith('#delayed-anchor')
    expect(callCount).toBeGreaterThan(2)

    spyQuerySelector.mockRestore()
    vi.useRealTimers()
  })

  it('waitForElement resolves null when timeout expires', async () => {
    vi.useFakeTimers()
    getCurrentUser.mockResolvedValue({ ok: true, data: { user: { email: 'a@b.c' } } })
    const router = await makeRouter()

    const spyQuerySelector = vi.spyOn(document, 'querySelector').mockReturnValue(null)

    router.push({ path: '/products', hash: '#missing-anchor' })
    await vi.advanceTimersByTimeAsync(1100)
    await flushPromises()

    expect(spyQuerySelector).toHaveBeenCalledWith('#missing-anchor')

    spyQuerySelector.mockRestore()
    vi.useRealTimers()
  })
})
