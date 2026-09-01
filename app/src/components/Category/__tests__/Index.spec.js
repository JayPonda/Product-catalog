import { describe, expect, it, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import Index from '../Index.vue'
import { useAuthStore } from '@/stores/auth'
import { useNotificationStore } from '@/stores/notifications'

vi.mock('@/network/request.js', () => ({
  getCategories: vi.fn(),
  createCategory: vi.fn(),
}))

vi.mock('@/utils/logger', () => ({
  default: { Debug: vi.fn(), Info: vi.fn(), Warn: vi.fn(), Error: vi.fn() },
}))

import { getCategories, createCategory } from '@/network/request.js'

describe('Category/Index.vue', () => {
  const mockCategoryData = {
    total: 2,
    limit: 20,
    offset: 0,
    categories: [
      { id: 'cat-1', name: 'Electronics' },
      { id: 'cat-2', name: 'Clothing' },
    ],
  }

  beforeEach(() => {
    vi.clearAllMocks()
    const pinia = createPinia()
    setActivePinia(pinia)
  })

  it('fetches categories on mount and displays them in the table', async () => {
    vi.mocked(getCategories).mockResolvedValue({ ok: true, data: mockCategoryData })

    const wrapper = mount(Index)
    await flushPromises()

    expect(getCategories).toHaveBeenCalledWith(0, 20)
    const cells = wrapper.findAll('td')
    expect(cells[0]?.text()).toBe('cat-1')
    expect(cells[1]?.text()).toBe('Electronics')
    expect(cells[2]?.text()).toBe('cat-2')
    expect(cells[3]?.text()).toBe('Clothing')
  })

  it('searches categories by name with debouncing', async () => {
    vi.useFakeTimers()
    vi.mocked(getCategories).mockResolvedValue({ ok: true, data: mockCategoryData })

    const wrapper = mount(Index)
    await flushPromises()

    const searchInput = wrapper.find('input[type="search"]')
    expect(searchInput.exists()).toBe(true)

    await searchInput.setValue('elect')
    vi.advanceTimersByTime(350)
    await flushPromises()

    expect(getCategories).toHaveBeenLastCalledWith(0, 20, { name: 'elect' })
    vi.useRealTimers()
  })

  it('clears search when clear button is clicked', async () => {
    vi.mocked(getCategories).mockResolvedValue({ ok: true, data: mockCategoryData })

    const wrapper = mount(Index)
    await flushPromises()

    await wrapper.find('input[type="search"]').setValue('tools')
    await flushPromises()

    const clearBtn = wrapper.find('button[aria-label="Clear search"]')
    expect(clearBtn.exists()).toBe(true)
    await clearBtn.trigger('click')
    await flushPromises()

    expect(wrapper.find('input[type="search"]').element.value).toBe('')
    expect(getCategories).toHaveBeenLastCalledWith(0, 20)
  })

  it('resets search when reset button is clicked', async () => {
    vi.mocked(getCategories).mockResolvedValue({ ok: true, data: mockCategoryData })

    const wrapper = mount(Index)
    await flushPromises()

    await wrapper.find('input[type="search"]').setValue('gadget')
    await flushPromises()

    const resetBtn = wrapper.findAll('button').find((b) => b.text() === 'Reset')
    expect(resetBtn).toBeDefined()
    await resetBtn?.trigger('click')
    await flushPromises()

    expect(wrapper.find('input[type="search"]').element.value).toBe('')
    expect(getCategories).toHaveBeenLastCalledWith(0, 20)
  })

  it('allows authenticated users to create a new category', async () => {
    vi.mocked(getCategories).mockResolvedValue({ ok: true, data: mockCategoryData })
    vi.mocked(createCategory).mockResolvedValue({
      ok: true,
      data: { id: 'cat-3', name: 'Home' },
    })

    const auth = useAuthStore()
    auth.isAuthenticated = true

    const wrapper = mount(Index)
    await flushPromises()

    const input = wrapper.find('#new-category')
    expect(input.exists()).toBe(true)

    await input.setValue('Home')
    const addBtn = wrapper.findAll('button').find((b) => b.text() === 'Add')
    await addBtn?.trigger('click')
    await flushPromises()

    expect(createCategory).toHaveBeenCalledWith('Home')
  })

  it('shows error notification when fetch fails', async () => {
    const errorStore = useNotificationStore()
    vi.spyOn(errorStore, 'show')
    vi.mocked(getCategories).mockResolvedValue({ ok: false, error: 500, message: 'Server error' })

    mount(Index)
    await flushPromises()

    expect(errorStore.show).toHaveBeenCalledWith('Server error')
  })
})
