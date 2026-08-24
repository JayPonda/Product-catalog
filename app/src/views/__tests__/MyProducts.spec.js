import { describe, expect, it } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia } from 'pinia'
import MyProducts from '../MyProducts.vue'
import { useAuthStore } from '@/stores/auth'

const IndexStub = {
  name: 'Index',
  props: ['showControls', 'myProducts'],
  template:
    '<div data-test="product-index" :data-show-controls="String(showControls)" :data-my-products="String(myProducts)" />',
}

function mountView() {
  const pinia = createPinia()
  const wrapper = mount(MyProducts, {
    global: {
      plugins: [pinia],
      stubs: { Index: IndexStub },
    },
  })
  return { wrapper, auth: useAuthStore(pinia) }
}

describe('MyProducts view', () => {
  it('always lists only the user products', () => {
    const { wrapper } = mountView()

    expect(
      wrapper.find('[data-test="product-index"]').attributes('data-my-products'),
    ).toBe('true')
  })

  it('shows controls only when authenticated', async () => {
    const { wrapper, auth } = mountView()
    const index = wrapper.find('[data-test="product-index"]')

    expect(index.attributes('data-show-controls')).toBe('false')

    auth.isAuthenticated = true
    await flushPromises()

    expect(index.attributes('data-show-controls')).toBe('true')
  })
})
