import { describe, expect, it, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import ModifyProduct from '../ModifyProduct.vue'

const mockPush = vi.fn()
vi.mock('vue-router', () => ({
  useRoute: () => ({
    params: { id: 'p-123' },
  }),
  useRouter: () => ({
    push: mockPush,
  }),
}))

describe('ModifyProduct.vue', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders child sub-components properly', () => {
    const wrapper = mount(ModifyProduct, {
      global: {
        stubs: {
          ProductForm: { template: '<div class="stubbed-form">Form</div>' },
          CategoryLinker: {
            template: '<div class="stubbed-linker">Linker</div>',
          },
        },
      },
    })

    expect(wrapper.find('.stubbed-form').exists()).toBe(true)
    expect(wrapper.find('.stubbed-linker').exists()).toBe(true)
  })

  it('redirects to hash category when saved event is emitted', async () => {
    const wrapper = mount(ModifyProduct, {
      global: {
        stubs: {
          ProductForm: true,
          CategoryLinker: true,
        },
      },
    })

    const childForm = wrapper.findComponent({ name: 'ProductForm' })
    childForm.vm.$emit('saved', { id: 'p-new' })
    await flushPromises()

    expect(mockPush).toHaveBeenCalledWith({
      name: 'products-modify',
      params: { id: 'p-new' },
      hash: '#category',
    })
  })

  it('displays error banner if children emit error event', async () => {
    const wrapper = mount(ModifyProduct, {
      global: {
        stubs: {
          ProductForm: true,
          CategoryLinker: true,
        },
      },
    })

    const childForm = wrapper.findComponent({ name: 'ProductForm' })
    childForm.vm.$emit('error', 'Something failed!')
    await flushPromises()

    expect(wrapper.text()).toContain('Something failed!')
  })
})
