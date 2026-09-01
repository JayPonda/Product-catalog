import { describe, expect, it, vi, beforeEach, afterEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import CategoryLinker from '../CategoryLinker.vue'

vi.mock('@/network/request', () => ({
  getProduct: vi.fn(),
  searchCategory: vi.fn(),
  linkCategory: vi.fn(),
  unlinkCategory: vi.fn(),
}))

import { getProduct, searchCategory, linkCategory, unlinkCategory } from '@/network/request'
import { useNotificationStore } from '@/stores/notifications'

describe('CategoryLinker.vue', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.useFakeTimers()
    setActivePinia(createPinia())
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  const getMockProduct = () => ({
    id: 'p-1',
    name: 'Product A',
    categories: [{ id: 'cat-1', name: 'Electronics' }],
  })

  it('fetches and displays linked categories on init', async () => {
    vi.mocked(getProduct).mockResolvedValue({ ok: true, data: getMockProduct() })

    const wrapper = mount(CategoryLinker, {
      props: { productId: 'p-1' },
    })
    await flushPromises()

    expect(getProduct).toHaveBeenCalledWith('p-1')
    expect(wrapper.text()).toContain('Electronics')
  })

  it('unlinks linked category when click cross button', async () => {
    vi.mocked(getProduct).mockResolvedValue({ ok: true, data: getMockProduct() })
    vi.mocked(unlinkCategory).mockResolvedValue({
      ok: true,
      data: 'Category unlinked successfully',
    })

    const wrapper = mount(CategoryLinker, {
      props: { productId: 'p-1' },
    })
    await flushPromises()

    const unlinkBtn = wrapper.find('button[aria-label="Unlink category"]')
    expect(unlinkBtn.exists()).toBe(true)

    await unlinkBtn.trigger('click')
    await flushPromises()

    expect(unlinkCategory).toHaveBeenCalledWith('p-1', 'cat-1')
    expect(wrapper.text()).not.toContain('Electronics')
  })

  it('searches for category and handles pending selection', async () => {
    vi.mocked(getProduct).mockResolvedValue({ ok: true, data: getMockProduct() })
    vi.mocked(searchCategory).mockResolvedValue({
      ok: true,
      data: {
        categories: [{ id: 'cat-2', name: 'Software' }],
      },
    })

    const wrapper = mount(CategoryLinker, {
      props: { productId: 'p-1' },
    })
    await flushPromises()

    const searchInput = wrapper.find('input[type="search"]')
    await searchInput.setValue('Soft')

    // Fast-forward debounce timer
    vi.advanceTimersByTime(250)
    await flushPromises()

    expect(searchCategory).toHaveBeenCalledWith('Soft')
    expect(wrapper.find('ul').exists()).toBe(true)

    const listBtn = wrapper.find('ul button')
    expect(listBtn.text()).toBe('Software')

    await listBtn.trigger('mousedown')
    await flushPromises()

    expect(wrapper.text()).toContain('Software') // Now in pending list
  })

  it('allows removing a pending category tag before linking', async () => {
    vi.mocked(getProduct).mockResolvedValue({ ok: true, data: getMockProduct() })
    vi.mocked(searchCategory).mockResolvedValue({
      ok: true,
      data: {
        categories: [{ id: 'cat-2', name: 'Software' }],
      },
    })

    const wrapper = mount(CategoryLinker, {
      props: { productId: 'p-1' },
    })
    await flushPromises()

    const searchInput = wrapper.find('input[type="search"]')
    await searchInput.setValue('Soft')
    vi.advanceTimersByTime(250)
    await flushPromises()

    await wrapper.find('ul button').trigger('mousedown')
    await flushPromises()

    expect(wrapper.text()).toContain('Software')

    // Find and click the remove pending category button
    const removeBtn = wrapper.find('button[aria-label="Remove from selection"]')
    expect(removeBtn.exists()).toBe(true)
    await removeBtn.trigger('click')
    await flushPromises()

    expect(wrapper.text()).not.toContain('Software')
  })

  it('submits pending category link calls when button is clicked', async () => {
    vi.mocked(getProduct).mockResolvedValue({ ok: true, data: getMockProduct() })
    vi.mocked(searchCategory).mockResolvedValue({
      ok: true,
      data: {
        categories: [{ id: 'cat-2', name: 'Software' }],
      },
    })
    vi.mocked(linkCategory).mockResolvedValue({ ok: true, data: {} })

    const wrapper = mount(CategoryLinker, {
      props: { productId: 'p-1' },
    })
    await flushPromises()

    const searchInput = wrapper.find('input[type="search"]')
    await searchInput.setValue('Soft')
    vi.advanceTimersByTime(250)
    await flushPromises()

    await wrapper.find('ul button').trigger('mousedown')
    await flushPromises()

    const notificationStore = useNotificationStore()
    vi.spyOn(notificationStore, 'success')

    const linkBtn = wrapper.find('button.bg-emerald-700')
    expect(linkBtn.exists()).toBe(true)
    await linkBtn.trigger('click')
    await flushPromises()

    expect(linkCategory).toHaveBeenCalledWith('p-1', 'cat-2')
    expect(notificationStore.success).toHaveBeenCalledWith('Category linked successfully.')
  })

  it('emits error when getProduct API fails on init', async () => {
    vi.mocked(getProduct).mockResolvedValue({ ok: false, error: 500 })

    const wrapper = mount(CategoryLinker, {
      props: { productId: 'p-1' },
    })
    await flushPromises()

    expect(wrapper.emitted('error')).toBeTruthy()
    const errs = wrapper.emitted('error')
    expect(errs[errs.length - 1]).toEqual(['500'])
  })

  it('emits error when unlinkCategory API fails', async () => {
    vi.mocked(getProduct).mockResolvedValue({ ok: true, data: getMockProduct() })
    vi.mocked(unlinkCategory).mockResolvedValue({ ok: false, error: 403 })

    const wrapper = mount(CategoryLinker, {
      props: { productId: 'p-1' },
    })
    await flushPromises()

    await wrapper.find('button[aria-label="Unlink category"]').trigger('click')
    await flushPromises()

    expect(wrapper.emitted('error')).toBeTruthy()
    const errs = wrapper.emitted('error')
    expect(errs[errs.length - 1]).toEqual(['403'])
  })

  it('emits error when linkCategory API fails', async () => {
    vi.mocked(getProduct).mockResolvedValue({ ok: true, data: getMockProduct() })
    vi.mocked(searchCategory).mockResolvedValue({
      ok: true,
      data: {
        categories: [{ id: 'cat-2', name: 'Software' }],
      },
    })
    vi.mocked(linkCategory).mockResolvedValue({ ok: false, error: 400 })

    const wrapper = mount(CategoryLinker, {
      props: { productId: 'p-1' },
    })
    await flushPromises()

    await wrapper.find('input[type="search"]').setValue('Soft')
    vi.advanceTimersByTime(250)
    await flushPromises()

    await wrapper.find('ul button').trigger('mousedown')
    await flushPromises()

    await wrapper.find('button.bg-emerald-700').trigger('click')
    await flushPromises()

    expect(wrapper.emitted('error')).toBeTruthy()
    const errs = wrapper.emitted('error')
    expect(errs[errs.length - 1]).toEqual(['Could not add: Software'])
  })
})
