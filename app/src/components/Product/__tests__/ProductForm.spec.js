import { describe, expect, it, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import ProductForm from '../ProductForm.vue'

vi.mock('@/network/request', () => ({
  getProduct: vi.fn(),
  createProduct: vi.fn(),
  updateProduct: vi.fn(),
}))

import { getProduct, createProduct, updateProduct } from '@/network/request'

describe('ProductForm.vue', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders form inputs correctly', () => {
    const wrapper = mount(ProductForm)
    expect(wrapper.find('input#product-name').exists()).toBe(true)
    expect(wrapper.find('textarea#product-description').exists()).toBe(true)
    expect(wrapper.find('input#product-stock').exists()).toBe(true)
    expect(wrapper.find('input#product-price').exists()).toBe(true)
  })

  it('fetches existing product details if productId is provided', async () => {
    vi.mocked(getProduct).mockResolvedValue({
      ok: true,
      data: {
        id: 'p-1',
        name: 'Existing Product',
        description: 'Existing Description',
        stock_quantity: 42,
        price: 9999,
      },
    })

    const wrapper = mount(ProductForm, {
      props: { productId: 'p-1' },
    })
    await flushPromises()

    expect(getProduct).toHaveBeenCalledWith('p-1')
    expect(wrapper.find('input#product-name').element.value).toBe('Existing Product')
    expect(wrapper.find('textarea#product-description').element.value).toBe('Existing Description')
    expect(wrapper.find('input#product-stock').element.value).toBe('42')
    expect(wrapper.find('input#product-price').element.value).toBe('99.99')
  })

  it('shows error messages for invalid input', async () => {
    const wrapper = mount(ProductForm)

    // Submit blank form
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.text()).toContain('Product name is required.')
    expect(wrapper.text()).toContain('Description is required.')
    expect(wrapper.text()).toContain('Stock quantity is required.')
    expect(wrapper.text()).toContain('Price is required.')
  })

  it('submits created product details when form is valid', async () => {
    vi.mocked(createProduct).mockResolvedValue({
      ok: true,
      data: { id: 'p-new' },
    })

    const wrapper = mount(ProductForm)

    await wrapper.find('input#product-name').setValue('New Product')
    await wrapper.find('textarea#product-description').setValue('A beautiful product description.')
    await wrapper.find('input#product-stock').setValue('10')
    await wrapper.find('input#product-price').setValue('19.99')

    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(createProduct).toHaveBeenCalledWith({
      name: 'New Product',
      description: 'A beautiful product description.',
      stock_quantity: 10,
      price: 1999,
    })

    expect(wrapper.emitted().saved).toBeTruthy()
    expect(wrapper.emitted().saved?.[0]).toEqual([{ id: 'p-new' }])
  })

  it('submits updated product details when editing', async () => {
    vi.mocked(getProduct).mockResolvedValue({
      ok: true,
      data: {
        id: 'p-1',
        name: 'Existing',
        description: 'Desc',
        stock_quantity: 5,
        price: 500,
      },
    })
    vi.mocked(updateProduct).mockResolvedValue({
      ok: true,
      data: { id: 'p-1' },
    })

    const wrapper = mount(ProductForm, {
      props: { productId: 'p-1' },
    })
    await flushPromises()

    await wrapper.find('input#product-name').setValue('Updated Name')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(updateProduct).toHaveBeenCalledWith('p-1', {
      name: 'Updated Name',
      description: 'Desc',
      stock_quantity: 5,
      price: 500,
    })
    expect(wrapper.emitted().saved).toBeTruthy()
  })

  it('shows error if price exceeds limit', async () => {
    const wrapper = mount(ProductForm)
    await wrapper.find('input#product-price').setValue('10000000.00') // $10,000,000.00 (limit is 9,999,999.99)
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.text()).toContain('Price cannot exceed $9,999,999.99.')
  })

  it('emits error if save request fails', async () => {
    vi.mocked(createProduct).mockResolvedValue({
      ok: false,
      error: 'Network Error',
    })

    const wrapper = mount(ProductForm)
    await wrapper.find('input#product-name').setValue('New Product')
    await wrapper.find('textarea#product-description').setValue('Desc')
    await wrapper.find('input#product-stock').setValue('1')
    await wrapper.find('input#product-price').setValue('1.00')

    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.emitted().error).toBeTruthy()
    expect(wrapper.emitted().error?.[1]).toEqual(['Network Error'])
  })

  it('emits error if fetching product information fails', async () => {
    vi.mocked(getProduct).mockResolvedValue({
      ok: false,
      error: 'Product Not Found',
    })

    const wrapper = mount(ProductForm, {
      props: { productId: 'p-1' },
    })
    await flushPromises()

    expect(wrapper.emitted().error).toBeTruthy()
    expect(wrapper.emitted().error?.[0]).toEqual(['Product Not Found'])
  })

  it('shows error if stock is less than 1', async () => {
    const wrapper = mount(ProductForm)
    await wrapper.find('input#product-stock').setValue('0')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.text()).toContain('Stock quantity must be at least 1.')
  })

  it('shows error if stock exceeds limit', async () => {
    const wrapper = mount(ProductForm)
    await wrapper.find('input#product-stock').setValue('2147483648')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.text()).toContain('Stock quantity cannot exceed 2147483647.')
  })

  it('shows error if price is less than or equal to 0', async () => {
    const wrapper = mount(ProductForm)
    await wrapper.find('input#product-price').setValue('-5.00')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.text()).toContain('Price must be greater than 0.')
  })
})
