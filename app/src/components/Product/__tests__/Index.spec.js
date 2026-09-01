import { describe, expect, it, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import Index from '../Index.vue'
import { useNotificationStore } from '@/stores/notifications'

// Mock router
const mockPush = vi.fn()
vi.mock('vue-router', () => ({
  useRouter: () => ({
    push: mockPush,
  }),
}))

// Mock network
vi.mock('@/network/request.js', () => ({
  getProducts: vi.fn(),
  getMyProducts: vi.fn(),
  deleteProduct: vi.fn(),
  getCategories: vi.fn().mockResolvedValue({ ok: true, data: { categories: [] } }),
  searchCategory: vi.fn().mockResolvedValue({ ok: true, data: { categories: [] } }),
}))

vi.mock('@/utils/logger', () => ({
  default: { Debug: vi.fn(), Info: vi.fn(), Warn: vi.fn(), Error: vi.fn() },
}))

import {
  getProducts,
  getMyProducts,
  deleteProduct,
  getCategories,
  searchCategory,
} from '@/network/request.js'

describe('Product/Index.vue', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(getCategories).mockResolvedValue({
      ok: true,
      data: {
        categories: [
          { id: 'cat-1', name: 'Tech' },
          { id: 'cat-2', name: 'Apparel' },
        ],
      },
    })
    vi.mocked(searchCategory).mockResolvedValue({
      ok: true,
      data: { categories: [] },
    })
    const pinia = createPinia()
    setActivePinia(pinia)
  })

  const mockProductData = {
    total: 2,
    limit: 20,
    offset: 0,
    products: [
      {
        id: 'p-1',
        name: 'Product A',
        code: 'PRD-A',
        price_cents: 1000,
        categories: [{ id: 'cat-1', name: 'Tech' }],
      },
      { id: 'p-2', name: 'Product B', code: 'PRD-B', price_cents: 2000, categories: [] },
    ],
  }

  it('fetches all products on mount when myProducts is false', async () => {
    vi.mocked(getProducts).mockResolvedValue({ ok: true, data: mockProductData })

    const wrapper = mount(Index, {
      props: { myProducts: false, showControls: false },
    })
    await flushPromises()

    expect(getProducts).toHaveBeenCalledWith(0, 20)
    expect(getMyProducts).not.toHaveBeenCalled()

    // Table should render scalar cells + categories
    const tableCells = wrapper.findAll('td')
    expect(tableCells[0]?.text()).toBe('p-1')
    expect(tableCells[1]?.text()).toBe('Product A')
    expect(tableCells[2]?.text()).toBe('PRD-A')
    expect(tableCells[3]?.text()).toBe('1000') // price_cents
    expect(tableCells[4]?.text()).toBe('Tech') // category name
  })

  it('fetches only my products on mount when myProducts is true', async () => {
    vi.mocked(getMyProducts).mockResolvedValue({ ok: true, data: mockProductData })

    mount(Index, {
      props: { myProducts: true, showControls: false },
    })
    await flushPromises()

    expect(getMyProducts).toHaveBeenCalledWith(0, 20)
    expect(getProducts).not.toHaveBeenCalled()
  })

  it('triggers router push when clicking New Product button', async () => {
    vi.mocked(getProducts).mockResolvedValue({ ok: true, data: mockProductData })

    const wrapper = mount(Index, {
      props: { myProducts: false, showControls: false },
    })
    await flushPromises()

    const button = wrapper.findAll('button').find((b) => b.text().includes('New Product'))
    expect(button).toBeDefined()
    await button.trigger('click')

    expect(mockPush).toHaveBeenCalledWith({ name: 'products-create' })
  })

  it('shows and handles action dropdown menu when showControls is true', async () => {
    vi.mocked(getProducts).mockResolvedValue({ ok: true, data: mockProductData })

    const wrapper = mount(Index, {
      props: { myProducts: false, showControls: true },
    })
    await flushPromises()

    // Toggle menu
    const actionBtns = wrapper.findAll('button[aria-label="Product actions"]')
    expect(actionBtns).toHaveLength(2)

    await actionBtns[0]?.trigger('click')
    await flushPromises()

    // Check menu links
    const menuButtons = wrapper.findAll('.absolute button')
    expect(menuButtons).toHaveLength(3)

    // Edit categories
    await menuButtons[0]?.trigger('click')
    expect(mockPush).toHaveBeenCalledWith({
      name: 'products-modify',
      params: { id: 'p-1', action: 'edit' },
      hash: '#category',
    })

    // Edit product
    await menuButtons[1]?.trigger('click')
    expect(mockPush).toHaveBeenCalledWith({
      name: 'products-modify',
      params: { id: 'p-1', action: 'edit' },
    })

    // Delete product
    vi.mocked(deleteProduct).mockResolvedValue({ ok: true, data: 'Product deleted successfully' })
    await menuButtons[2]?.trigger('click')
    await flushPromises()

    expect(deleteProduct).toHaveBeenCalledWith('p-1')
    expect(getProducts).toHaveBeenCalledTimes(2) // Refreshed
  })

  it('handles pagination next and previous', async () => {
    vi.mocked(getProducts)
      .mockResolvedValueOnce({
        ok: true,
        data: {
          total: 25,
          limit: 20,
          offset: 0,
          products: Array.from({ length: 20 }, (_, i) => ({
            id: `p-${i}`,
            name: `Product ${i}`,
            code: `PRD-${i}`,
            price_cents: 100,
            categories: [],
          })),
        },
      })
      .mockResolvedValueOnce({
        ok: true,
        data: {
          total: 25,
          limit: 20,
          offset: 1,
          products: [],
        },
      })
      .mockResolvedValueOnce({
        ok: true,
        data: {
          total: 25,
          limit: 20,
          offset: 0,
          products: [],
        },
      })

    const wrapper = mount(Index, {
      props: { myProducts: false, showControls: false },
    })
    await flushPromises()

    const buttons = wrapper.findAll('nav button')
    const nextButton = buttons.find((b) => b.text() === 'Next')
    const prevButton = buttons.find((b) => b.text() === 'Previous')

    await nextButton?.trigger('click')
    await flushPromises()

    expect(getProducts).toHaveBeenLastCalledWith(1, 20)

    await prevButton?.trigger('click')
    await flushPromises()

    expect(getProducts).toHaveBeenLastCalledWith(0, 20)
  })

  it('displays error if pagination next is clicked on last page', async () => {
    const errorStore = useNotificationStore()
    vi.spyOn(errorStore, 'show')

    vi.mocked(getProducts).mockResolvedValue({
      ok: true,
      data: {
        total: 10,
        limit: 20,
        offset: 0,
        products: [],
      },
    })

    const wrapper = mount(Index, {
      props: { myProducts: false, showControls: false },
    })
    await flushPromises()

    const nextButton = wrapper.findAll('nav button').find((b) => b.text() === 'Next')
    await nextButton?.trigger('click')

    expect(errorStore.show).toHaveBeenCalledWith('No records on next page.')
  })

  it('displays error if pagination previous is clicked on first page', async () => {
    const errorStore = useNotificationStore()
    vi.spyOn(errorStore, 'show')

    vi.mocked(getProducts).mockResolvedValue({
      ok: true,
      data: {
        total: 10,
        limit: 20,
        offset: 0,
        products: [],
      },
    })

    const wrapper = mount(Index, {
      props: { myProducts: false, showControls: false },
    })
    await flushPromises()

    const prevButton = wrapper.findAll('nav button').find((b) => b.text() === 'Previous')
    await prevButton?.trigger('click')

    expect(errorStore.show).toHaveBeenCalledWith('No records on previous page.')
  })

  it('next displays error if products ref is falsy', async () => {
    const errorStore = useNotificationStore()
    vi.spyOn(errorStore, 'show')
    const wrapper = mount(Index, {
      props: { myProducts: false, showControls: false },
    })
    await flushPromises()

    wrapper.vm.products = null
    wrapper.vm.next()

    expect(errorStore.show).toHaveBeenCalledWith('Something went wrong')
  })

  it('previous displays error if products ref is falsy', async () => {
    const errorStore = useNotificationStore()
    vi.spyOn(errorStore, 'show')
    const wrapper = mount(Index, {
      props: { myProducts: false, showControls: false },
    })
    await flushPromises()

    wrapper.vm.products = null
    wrapper.vm.previous()

    expect(errorStore.show).toHaveBeenCalledWith('Something went wrong')
  })

  describe('Search and Category filtering', () => {
    it('searches products by name with debouncing', async () => {
      vi.useFakeTimers()
      vi.mocked(getProducts).mockResolvedValue({ ok: true, data: mockProductData })

      const wrapper = mount(Index, {
        props: { myProducts: false, showControls: false },
      })
      await flushPromises()

      const searchInput = wrapper.find('input[type="search"]')
      expect(searchInput.exists()).toBe(true)

      await searchInput.setValue('headphone')
      // Fast forward debounce timer
      vi.advanceTimersByTime(350)
      await flushPromises()

      expect(getProducts).toHaveBeenLastCalledWith(0, 20, { name: 'headphone' })
      vi.useRealTimers()
    })

    it('clears search when clear button is clicked', async () => {
      vi.mocked(getProducts).mockResolvedValue({ ok: true, data: mockProductData })

      const wrapper = mount(Index, {
        props: { myProducts: false, showControls: false },
      })
      await flushPromises()

      await wrapper.find('input[type="search"]').setValue('test')
      await flushPromises()

      const clearBtn = wrapper.find('button[aria-label="Clear search"]')
      expect(clearBtn.exists()).toBe(true)
      await clearBtn.trigger('click')
      await flushPromises()

      expect(wrapper.find('input[type="search"]').element.value).toBe('')
      expect(getProducts).toHaveBeenLastCalledWith(0, 20)
    })

    it('filters by multiple categories and displays category chips', async () => {
      vi.mocked(getProducts).mockResolvedValue({ ok: true, data: mockProductData })

      const wrapper = mount(Index, {
        props: { myProducts: false, showControls: false },
      })
      await flushPromises()

      // Open category dropdown
      const filterBtn = wrapper.find('button[aria-label="Filter by categories"]')
      await filterBtn.trigger('click')
      await flushPromises()

      // Checkboxes for categories
      const checkboxes = wrapper.findAll('input[type="checkbox"]')
      expect(checkboxes.length).toBe(2)

      // Select first category (Tech)
      await checkboxes[0]?.setValue(true)
      await flushPromises()

      expect(getProducts).toHaveBeenLastCalledWith(0, 20, { categoryIds: ['cat-1'] })

      // Select second category (Apparel)
      await checkboxes[1]?.setValue(true)
      await flushPromises()

      expect(getProducts).toHaveBeenLastCalledWith(0, 20, { categoryIds: ['cat-1', 'cat-2'] })

      // Badge count should be 2
      expect(filterBtn.text()).toContain('2')

      // Category chips should be displayed
      const chips = wrapper.findAll('button[aria-label^="Remove "]')
      expect(chips.length).toBe(2)

      // Remove first category chip
      await chips[0]?.trigger('click')
      await flushPromises()

      expect(getProducts).toHaveBeenLastCalledWith(0, 20, { categoryIds: ['cat-2'] })
    })

    it('resets all filters on clicking reset button', async () => {
      vi.mocked(getProducts).mockResolvedValue({ ok: true, data: mockProductData })

      const wrapper = mount(Index, {
        props: { myProducts: false, showControls: false },
      })
      await flushPromises()

      await wrapper.find('input[type="search"]').setValue('query')
      await flushPromises()

      // Reset button should appear
      const resetBtn = wrapper.findAll('button').find((b) => b.text() === 'Reset filters')
      expect(resetBtn).toBeDefined()

      await resetBtn?.trigger('click')
      await flushPromises()

      expect(wrapper.find('input[type="search"]').element.value).toBe('')
      expect(getProducts).toHaveBeenLastCalledWith(0, 20)
    })

    it('passes search and category filters to getMyProducts when myProducts is true', async () => {
      vi.useFakeTimers()
      vi.mocked(getMyProducts).mockResolvedValue({ ok: true, data: mockProductData })

      const wrapper = mount(Index, {
        props: { myProducts: true, showControls: true },
      })
      await flushPromises()

      await wrapper.find('input[type="search"]').setValue('my-item')
      vi.advanceTimersByTime(350)
      await flushPromises()

      expect(getMyProducts).toHaveBeenLastCalledWith(0, 20, { name: 'my-item' })
      vi.useRealTimers()
    })

    it('maintains category selection during pagination next and previous', async () => {
      vi.mocked(getProducts)
        .mockResolvedValueOnce({ ok: true, data: mockProductData }) // initial mount
        .mockResolvedValueOnce({
          ok: true,
          data: {
            total: 35,
            limit: 20,
            offset: 0,
            products: [{ id: 'p-1', name: 'Product 1' }],
          },
        }) // after category select
        .mockResolvedValueOnce({
          ok: true,
          data: {
            total: 35,
            limit: 20,
            offset: 20,
            products: [{ id: 'p-21', name: 'Product 21' }],
          },
        }) // after next page
        .mockResolvedValueOnce({
          ok: true,
          data: {
            total: 35,
            limit: 20,
            offset: 0,
            products: [{ id: 'p-1', name: 'Product 1' }],
          },
        }) // after prev page

      const wrapper = mount(Index, {
        props: { myProducts: false, showControls: false },
      })
      await flushPromises()

      // Select category
      const filterBtn = wrapper.find('button[aria-label="Filter by categories"]')
      await filterBtn.trigger('click')
      await flushPromises()

      const checkbox = wrapper.find('input[type="checkbox"]')
      await checkbox.setValue(true)
      await flushPromises()

      expect(getProducts).toHaveBeenLastCalledWith(0, 20, { categoryIds: ['cat-1'] })

      // Click next page
      const nextBtn = wrapper.findAll('nav button').find((b) => b.text() === 'Next')
      await nextBtn?.trigger('click')
      await flushPromises()

      // Expect page 1 with same category selection
      expect(getProducts).toHaveBeenLastCalledWith(1, 20, { categoryIds: ['cat-1'] })

      // Click previous page
      const prevBtn = wrapper.findAll('nav button').find((b) => b.text() === 'Previous')
      await prevBtn?.trigger('click')
      await flushPromises()

      // Expect back on page 0 with same category selection
      expect(getProducts).toHaveBeenLastCalledWith(0, 20, { categoryIds: ['cat-1'] })
    })
  })
})
