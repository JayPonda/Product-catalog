import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import Home from '../Home.vue'

const IndexStub = {
  name: 'Index',
  props: ['showControls', 'myProducts'],
  template:
    '<div data-test="product-index" :data-show-controls="String(showControls)" :data-my-products="String(myProducts)" />',
}

function mountView() {
  return mount(Home, {
    global: { stubs: { Index: IndexStub } },
  })
}

describe('Home view', () => {
  it('renders the public product list without controls', () => {
    const index = mountView().find('[data-test="product-index"]')

    expect(index.exists()).toBe(true)
    expect(index.attributes('data-show-controls')).toBe('false')
    expect(index.attributes('data-my-products')).toBe('false')
  })
})
