import { describe, expect, it, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import ModifyProduct from '../ModifyProduct.vue'

let mockRouteParams = { id: 'p-123' }
const mockPush = vi.fn()
vi.mock('vue-router', () => ({
  useRoute: () => ({
    get params() {
      return mockRouteParams
    },
  }),
  useRouter: () => ({
    push: mockPush,
  }),
}))

import { useNotificationStore } from '@/stores/notifications'

describe('ModifyProduct.vue', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockRouteParams = { id: 'p-123' }
    setActivePinia(createPinia())
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

  it('redirects to hash category and notifies when saved event is emitted', async () => {
    const notificationStore = useNotificationStore()
    vi.spyOn(notificationStore, 'success')

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

    expect(notificationStore.success).toHaveBeenCalledWith('Product successfully edited.')
    expect(mockPush).toHaveBeenCalledWith({
      name: 'products-modify',
      params: { id: 'p-new' },
      hash: '#category',
    })
  })

  it('triggers product added notification when adding a new product (no id in route)', async () => {
    mockRouteParams = {}
    const notificationStore = useNotificationStore()
    vi.spyOn(notificationStore, 'success')

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

    expect(notificationStore.success).toHaveBeenCalledWith('Product successfully added.')
  })

  it('triggers notification error if children emit error event', async () => {
    const notificationStore = useNotificationStore()
    vi.spyOn(notificationStore, 'error')

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

    expect(notificationStore.error).toHaveBeenCalledWith('Something failed!')
  })
})
