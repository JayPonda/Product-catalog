import { describe, expect, it, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import Index from '../Index.vue'
import { useErrorStore } from '@/stores/errors'

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
}))

vi.mock('@/utils/logger', () => ({
  default: { Debug: vi.fn(), Info: vi.fn(), Warn: vi.fn(), Error: vi.fn() },
}))

import { getProducts, getMyProducts, deleteProduct } from '@/network/request.js'

describe('Product/Index.vue', () => {
  beforeEach(() => {
    vi.clearAllMocks()
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

    const button = wrapper.find('button')
    expect(button.text()).toBe('New Product')
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
    const errorStore = useErrorStore()
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
    const errorStore = useErrorStore()
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
    const errorStore = useErrorStore()
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
    const errorStore = useErrorStore()
    vi.spyOn(errorStore, 'show')
    const wrapper = mount(Index, {
      props: { myProducts: false, showControls: false },
    })
    await flushPromises()

    wrapper.vm.products = null
    wrapper.vm.previous()

    expect(errorStore.show).toHaveBeenCalledWith('Something went wrong')
  })
})
