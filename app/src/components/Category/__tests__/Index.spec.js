import { describe, expect, it, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import Index from '../Index.vue'
import { useAuthStore } from '@/stores/auth'
import { useErrorStore } from '@/stores/errors'

vi.mock('@/network/request.js', () => ({
  getCategories: vi.fn(),
  createCategory: vi.fn(),
}))

import { getCategories, createCategory } from '@/network/request.js'

describe('Category/Index.vue', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    const pinia = createPinia()
    setActivePinia(pinia)
  })

  const mockCategories = {
    total: 2,
    limit: 20,
    offset: 0,
    categories: [
      { id: 'cat-1', name: 'Electronics' },
      { id: 'cat-2', name: 'Books' },
    ],
  }

  it('fetches categories on mount and renders them', async () => {
    vi.mocked(getCategories).mockResolvedValue({
      ok: true,
      data: mockCategories,
    })

    const wrapper = mount(Index)
    await flushPromises()

    expect(getCategories).toHaveBeenCalledWith(0, 20)
    // Check that BaseTable cell renders values
    const tableCells = wrapper.findAll('td')
    expect(tableCells[0]?.text()).toBe('cat-1')
    expect(tableCells[1]?.text()).toBe('Electronics')
    expect(tableCells[2]?.text()).toBe('cat-2')
    expect(tableCells[3]?.text()).toBe('Books')
  })

  it('renders add category controls only when authenticated', async () => {
    vi.mocked(getCategories).mockResolvedValue({
      ok: true,
      data: mockCategories,
    })

    const wrapper = mount(Index)
    const authStore = useAuthStore()

    // Logged out
    authStore.isAuthenticated = false
    await flushPromises()
    expect(wrapper.find('input#new-category').exists()).toBe(false)

    // Logged in
    authStore.isAuthenticated = true
    await flushPromises()
    expect(wrapper.find('input#new-category').exists()).toBe(true)
  })

  it('submits a new category and refreshes list on success', async () => {
    const authStore = useAuthStore()
    authStore.isAuthenticated = true

    vi.mocked(getCategories).mockResolvedValue({
      ok: true,
      data: mockCategories,
    })
    vi.mocked(createCategory).mockResolvedValue({
      ok: true,
      data: { id: 'cat-3', name: 'Clothing' },
    })

    const wrapper = mount(Index)
    await flushPromises()

    const input = wrapper.find('input#new-category')
    await input.setValue('Clothing')

    const button = wrapper.find('button[type="button"]')
    await button.trigger('click')
    await flushPromises()

    expect(createCategory).toHaveBeenCalledWith('Clothing')
    expect(getCategories).toHaveBeenCalledTimes(2)
  })

  it('displays error if new category name is empty', async () => {
    const authStore = useAuthStore()
    authStore.isAuthenticated = true
    const errorStore = useErrorStore()
    vi.spyOn(errorStore, 'show')

    vi.mocked(getCategories).mockResolvedValue({
      ok: true,
      data: mockCategories,
    })

    const wrapper = mount(Index)
    await flushPromises()

    const button = wrapper.find('button[type="button"]')
    await button.trigger('click')

    expect(errorStore.show).toHaveBeenCalledWith('Category name is required.')
    expect(createCategory).not.toHaveBeenCalled()
  })

  it('handles pagination next and previous', async () => {
    vi.mocked(getCategories).mockResolvedValue({
      ok: true,
      data: {
        total: 25,
        limit: 20,
        offset: 0,
        categories: Array.from({ length: 20 }, (_, i) => ({
          id: `cat-${i}`,
          name: `Cat ${i}`,
        })),
      },
    })

    const wrapper = mount(Index)
    await flushPromises()

    const buttons = wrapper.findAll('nav button')
    const nextButton = buttons.find((b) => b.text() === 'Next')
    const prevButton = buttons.find((b) => b.text() === 'Previous')

    await nextButton?.trigger('click')
    await flushPromises()

    expect(getCategories).toHaveBeenLastCalledWith(1, 20)

    await prevButton?.trigger('click')
    await flushPromises()

    expect(getCategories).toHaveBeenLastCalledWith(0, 20)
  })

  it('displays error if pagination next is clicked on last page', async () => {
    const errorStore = useErrorStore()
    vi.spyOn(errorStore, 'show')

    vi.mocked(getCategories).mockResolvedValue({
      ok: true,
      data: {
        total: 10,
        limit: 20,
        offset: 0,
        categories: [],
      },
    })

    const wrapper = mount(Index)
    await flushPromises()

    const nextButton = wrapper.findAll('nav button').find((b) => b.text() === 'Next')
    await nextButton?.trigger('click')

    expect(errorStore.show).toHaveBeenCalledWith('No records on next page.')
  })

  it('displays error if pagination previous is clicked on first page', async () => {
    const errorStore = useErrorStore()
    vi.spyOn(errorStore, 'show')

    vi.mocked(getCategories).mockResolvedValue({
      ok: true,
      data: {
        total: 10,
        limit: 20,
        offset: 0,
        categories: [],
      },
    })

    const wrapper = mount(Index)
    await flushPromises()

    const prevButton = wrapper.findAll('nav button').find((b) => b.text() === 'Previous')
    await prevButton?.trigger('click')

    expect(errorStore.show).toHaveBeenCalledWith('No records on previous page.')
  })

  it('next displays error if categories ref is falsy', async () => {
    const errorStore = useErrorStore()
    vi.spyOn(errorStore, 'show')
    const wrapper = mount(Index)
    await flushPromises()
    
    wrapper.vm.categories = null
    wrapper.vm.next()
    
    expect(errorStore.show).toHaveBeenCalledWith('Something went wrong')
  })

  it('previous displays error if categories ref is falsy', async () => {
    const errorStore = useErrorStore()
    vi.spyOn(errorStore, 'show')
    const wrapper = mount(Index)
    await flushPromises()
    
    wrapper.vm.categories = null
    wrapper.vm.previous()
    
    expect(errorStore.show).toHaveBeenCalledWith('Something went wrong')
  })
})
