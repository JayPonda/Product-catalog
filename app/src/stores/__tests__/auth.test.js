import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'

const registerUser = vi.fn()
const loginUser = vi.fn()
const logoutUser = vi.fn()
const getCurrentUser = vi.fn()

vi.mock('@/network/request.js', () => ({
  registerUser: (...args) => registerUser(...args),
  loginUser: (...args) => loginUser(...args),
  logoutUser: (...args) => logoutUser(...args),
  getCurrentUser: (...args) => getCurrentUser(...args),
}))

import { useAuthStore } from '../auth'

describe('auth store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('login success sets user and isAuthenticated', async () => {
    loginUser.mockResolvedValue({ ok: true, data: { user: { email: 'a@b.c' } } })

    const auth = useAuthStore()
    const res = await auth.login({ email: 'a@b.c', password: 'pw' })

    expect(res.ok).toBe(true)
    expect(auth.isAuthenticated).toBe(true)
    expect(auth.user).toEqual({ email: 'a@b.c' })
  })

  it('login failure leaves state untouched and returns the result', async () => {
    const failure = { ok: false, error: 401, message: 'invalid' }
    loginUser.mockResolvedValue(failure)

    const auth = useAuthStore()
    const res = await auth.login({ email: 'a@b.c', password: 'bad' })

    expect(res).toEqual(failure)
    expect(auth.isAuthenticated).toBe(false)
    expect(auth.user).toBeNull()
  })

  it('register success marks the user authenticated', async () => {
    registerUser.mockResolvedValue({ ok: true, data: { user: { email: 'n@b.c' } } })

    const auth = useAuthStore()
    await auth.register({ email: 'n@b.c' })

    expect(auth.isAuthenticated).toBe(true)
    expect(auth.user?.email).toBe('n@b.c')
  })

  it('register failure does not authenticate', async () => {
    registerUser.mockResolvedValue({ ok: false, error: 409 })

    const auth = useAuthStore()
    await auth.register({ email: 'dup@b.c' })

    expect(auth.isAuthenticated).toBe(false)
  })

  it('logout clears state even when the API call fails', async () => {
    logoutUser.mockResolvedValue({ ok: false, error: 500 })
    getCurrentUser.mockResolvedValue({ ok: true, data: { user: { email: 'a@b.c' } } })

    const auth = useAuthStore()
    await auth.fetchMe()
    expect(auth.isAuthenticated).toBe(true)

    await auth.logout()

    expect(logoutUser).toHaveBeenCalledTimes(1)
    expect(auth.isAuthenticated).toBe(false)
    expect(auth.user).toBeNull()
  })

  it('fetchMe restores the session from the cookie endpoint', async () => {
    getCurrentUser.mockResolvedValue({ ok: true, data: { user: { email: 'me@b.c' } } })

    const auth = useAuthStore()
    const res = await auth.fetchMe()

    expect(res.ok).toBe(true)
    expect(auth.isAuthenticated).toBe(true)
    expect(auth.user?.email).toBe('me@b.c')
  })

  it('fetchMe resets state when the cookie is gone', async () => {
    getCurrentUser.mockResolvedValue({ ok: false, error: 401 })

    const auth = useAuthStore()
    auth.isAuthenticated = true
    auth.user = { email: 'stale@b.c' }

    await auth.fetchMe()

    expect(auth.isAuthenticated).toBe(false)
    expect(auth.user).toBeNull()
  })
})
