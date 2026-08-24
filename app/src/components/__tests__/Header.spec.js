import { describe, expect, it, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { createRouter, createMemoryHistory } from 'vue-router'

vi.mock('@/network/request.js', () => ({
  registerUser: vi.fn(),
  loginUser: vi.fn(),
  logoutUser: vi.fn(async () => ({ ok: true })),
  getCurrentUser: vi.fn(async () => ({ ok: false })),
}))

import Header from '../layout/Header.vue'
import { useAuthStore } from '@/stores/auth'

function makeRouter() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/', component: { template: '<div />' } },
      { path: '/login', component: { template: '<div />' } },
      { path: '/register', component: { template: '<div />' } },
      { path: '/categories', component: { template: '<div />' } },
      { path: '/my-products', component: { template: '<div />' } },
    ],
  })
}

async function mountHeader() {
  const pinia = setActivePinia(createPinia())
  const router = makeRouter()
  router.push('/')
  await router.isReady()

  const wrapper = mount(Header, {
    global: {
      plugins: [pinia, router],
    },
  })
  return { wrapper, auth: useAuthStore(), router }
}

describe('Header', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows login/register links and hides my-products when logged out', async () => {
    const { wrapper } = await mountHeader()

    expect(wrapper.text()).not.toContain('My Products')
    const links = wrapper.findAll('a').map((a) => a.text())
    expect(links).toContain('Login')
    expect(links).toContain('Register')
    expect(wrapper.find('button[aria-label="Open menu"]').exists()).toBe(true)
  })

  it('shows the user email and a logout button when authenticated', async () => {
    const { wrapper, auth } = await mountHeader()

    auth.isAuthenticated = true
    auth.user = { email: 'jane@b.c' }
    await flushPromises()

    expect(wrapper.text()).toContain('My Products')
    expect(wrapper.text()).toContain('jane@b.c')

    const buttons = wrapper.findAll('button').map((b) => b.text())
    expect(buttons).toContain('Logout')
    expect(wrapper.text()).not.toContain('Register')
  })

  it('logout clears the session and navigates to /login', async () => {
    const { wrapper, auth, router } = await mountHeader()
    auth.isAuthenticated = true
    auth.user = { email: 'jane@b.c' }
    await flushPromises()

    // The desktop Logout button is the last visible one outside the mobile panel.
    const logoutBtn = wrapper.findAll('button').find((b) => b.text() === 'Logout')
    await logoutBtn.trigger('click')
    await flushPromises()

    expect(auth.isAuthenticated).toBe(false)
    expect(auth.user).toBeNull()
    expect(router.currentRoute.value.path).toBe('/login')
  })

  it('toggles the mobile navigation panel', async () => {
    const { wrapper } = await mountHeader()

    const categoriesLinks = () =>
      wrapper.findAll('a').filter((a) => a.text() === 'Categories').length
    const toggle = wrapper.find('button[aria-label="Open menu"]')

    // Closed: only the desktop nav link exists.
    expect(categoriesLinks()).toBe(1)

    await toggle.trigger('click')
    expect(categoriesLinks()).toBe(2)

    await toggle.trigger('click')
    expect(categoriesLinks()).toBe(1)
  })

  it('closes mobile menu when clicking navigation links', async () => {
    const { wrapper } = await mountHeader()
    const toggle = wrapper.find('button[aria-label="Open menu"]')

    await toggle.trigger('click')

    const mobileLinks = wrapper.findAll('a').filter((a) => a.text() === 'Categories')
    expect(mobileLinks).toHaveLength(2)

    await mobileLinks[1].trigger('click')
    await flushPromises()

    const categoriesLinksCount = wrapper
      .findAll('a')
      .filter((a) => a.text() === 'Categories').length
    expect(categoriesLinksCount).toBe(1)
  })

  it('closes mobile menu when clicking login link in mobile menu', async () => {
    const { wrapper } = await mountHeader()
    const toggle = wrapper.find('button[aria-label="Open menu"]')

    await toggle.trigger('click')

    const loginLinks = wrapper.findAll('a').filter((a) => a.text() === 'Login')
    expect(loginLinks).toHaveLength(2)

    await loginLinks[1].trigger('click')
    await flushPromises()

    const loginLinksCount = wrapper.findAll('a').filter((a) => a.text() === 'Login').length
    expect(loginLinksCount).toBe(1)
  })
})
