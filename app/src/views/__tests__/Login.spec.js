import { describe, expect, it, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import Login from '../Login.vue'
import { useAuthStore } from '@/stores/auth'

const mockPush = vi.fn()
vi.mock('vue-router', () => ({
  useRouter: () => ({
    push: mockPush,
  }),
  RouterLink: {
    template: '<a><slot /></a>',
  },
}))

describe('Login.vue', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    setActivePinia(createPinia())
  })

  it('renders input fields and submit button', () => {
    const wrapper = mount(Login)
    expect(wrapper.find('input#email').exists()).toBe(true)
    expect(wrapper.find('input#password').exists()).toBe(true)
    expect(wrapper.find('button[type="submit"]').text()).toBe('Sign in')
  })

  it('toggles password visibility when eye icon is clicked', async () => {
    const wrapper = mount(Login)
    const passwordInput = wrapper.find('input#password')
    expect(passwordInput.attributes('type')).toBe('password')

    const toggleBtn = wrapper.find('button[type="button"]')
    await toggleBtn.trigger('click')
    await flushPromises()

    expect(passwordInput.attributes('type')).toBe('text')

    await toggleBtn.trigger('click')
    await flushPromises()

    expect(passwordInput.attributes('type')).toBe('password')
  })

  it('shows validation errors when fields are empty', async () => {
    const wrapper = mount(Login)
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.text()).toContain('Email is required.')
    expect(wrapper.text()).toContain('Password is required.')
  })

  it('calls auth login and redirects to /products on success', async () => {
    const authStore = useAuthStore()
    vi.spyOn(authStore, 'login').mockResolvedValue({
      ok: true,
      data: { email: 'user@example.com' },
    })

    const wrapper = mount(Login)
    await wrapper.find('input#email').setValue('user@example.com')
    await wrapper.find('input#password').setValue('password123')

    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(authStore.login).toHaveBeenCalledWith({
      email: 'user@example.com',
      password: 'password123',
    })
    expect(mockPush).toHaveBeenCalledWith('/products')
  })

  it('displays auth error message on failure', async () => {
    const authStore = useAuthStore()
    vi.spyOn(authStore, 'login').mockResolvedValue({
      ok: false,
      error: 401,
      message: 'Invalid credentials',
    })

    const wrapper = mount(Login)
    await wrapper.find('input#email').setValue('user@example.com')
    await wrapper.find('input#password').setValue('wrong')

    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.text()).toContain('Invalid credentials')
    expect(mockPush).not.toHaveBeenCalled()
  })

  it('dismisses auth error when dismiss button is clicked', async () => {
    const authStore = useAuthStore()
    vi.spyOn(authStore, 'login').mockResolvedValue({
      ok: false,
      error: 401,
      message: 'Invalid credentials',
    })

    const wrapper = mount(Login)
    await wrapper.find('input#email').setValue('user@example.com')
    await wrapper.find('input#password').setValue('wrong')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.find('[role="alert"]').exists()).toBe(true)

    const dismissBtn = wrapper.find('button[aria-label="Dismiss"]')
    expect(dismissBtn.exists()).toBe(true)
    await dismissBtn.trigger('click')
    await flushPromises()

    expect(wrapper.find('[role="alert"]').exists()).toBe(false)
  })
})
