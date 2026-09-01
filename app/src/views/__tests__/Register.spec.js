import { describe, expect, it, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import Register from '../Register.vue'
import { useAuthStore } from '@/stores/auth'
import { useNotificationStore } from '@/stores/notifications'

const mockPush = vi.fn()
vi.mock('vue-router', () => ({
  useRouter: () => ({
    push: mockPush,
  }),
  RouterLink: {
    template: '<a><slot /></a>',
  },
}))

describe('Register.vue', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    setActivePinia(createPinia())
  })

  it('renders all form inputs and button', () => {
    const wrapper = mount(Register)
    expect(wrapper.find('input#first_name').exists()).toBe(true)
    expect(wrapper.find('input#last_name').exists()).toBe(true)
    expect(wrapper.find('input#email').exists()).toBe(true)
    expect(wrapper.find('input#password').exists()).toBe(true)
    expect(wrapper.find('input#confirm_password').exists()).toBe(true)
    expect(wrapper.find('button[type="submit"]').text()).toBe('Create account')
  })

  it('toggles password and confirm visibility when eye icons are clicked', async () => {
    const wrapper = mount(Register)
    const passwordInput = wrapper.find('input#password')
    const confirmInput = wrapper.find('input#confirm_password')

    expect(passwordInput.attributes('type')).toBe('password')
    expect(confirmInput.attributes('type')).toBe('password')

    const toggles = wrapper.findAll('button[type="button"]')
    expect(toggles).toHaveLength(2)

    await toggles[0]?.trigger('click')
    await toggles[1]?.trigger('click')
    await flushPromises()

    expect(passwordInput.attributes('type')).toBe('text')
    expect(confirmInput.attributes('type')).toBe('text')
  })

  it('shows validation errors when fields are invalid or empty', async () => {
    const wrapper = mount(Register)

    // Submit empty
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.text()).toContain('First name is required.')
    expect(wrapper.text()).toContain('Last name is required.')
    expect(wrapper.text()).toContain('A valid email is required.')
    expect(wrapper.text()).toContain('Password must be at least 8 characters.')

    // Submit mismatched passwords
    await wrapper.find('input#password').setValue('password123')
    await wrapper.find('input#confirm_password').setValue('mismatch')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.text()).toContain('Passwords do not match.')
  })

  it('calls auth register and redirects to /login on success', async () => {
    const authStore = useAuthStore()
    vi.spyOn(authStore, 'register').mockResolvedValue({
      ok: true,
      data: { email: 'john.doe@example.com' },
    })

    const wrapper = mount(Register)
    await wrapper.find('input#first_name').setValue('John')
    await wrapper.find('input#last_name').setValue('Doe')
    await wrapper.find('input#email').setValue('john.doe@example.com')
    await wrapper.find('input#password').setValue('password123')
    await wrapper.find('input#confirm_password').setValue('password123')

    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(authStore.register).toHaveBeenCalledWith({
      first_name: 'John',
      last_name: 'Doe',
      email: 'john.doe@example.com',
      password: 'password123',
    })
    expect(mockPush).toHaveBeenCalledWith('/login')
  })

  it('triggers notification error message on server failure', async () => {
    const authStore = useAuthStore()
    const notificationStore = useNotificationStore()
    vi.spyOn(notificationStore, 'error')
    vi.spyOn(authStore, 'register').mockResolvedValue({
      ok: false,
      error: 400,
      message: 'Email already exists',
    })

    const wrapper = mount(Register)
    await wrapper.find('input#first_name').setValue('John')
    await wrapper.find('input#last_name').setValue('Doe')
    await wrapper.find('input#email').setValue('john.doe@example.com')
    await wrapper.find('input#password').setValue('password123')
    await wrapper.find('input#confirm_password').setValue('password123')

    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(notificationStore.error).toHaveBeenCalledWith('Email already exists')
    expect(mockPush).not.toHaveBeenCalled()
  })
})
