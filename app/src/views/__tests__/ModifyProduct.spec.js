import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import ModifyProduct from '../ModifyProduct.vue'

const IndexStub = {
  name: 'Index',
  template: '<div data-test="modify-product-index" />',
}

describe('ModifyProduct view', () => {
  it('renders the product form inside a padded container', () => {
    const wrapper = mount(ModifyProduct, {
      global: { stubs: { Index: IndexStub } },
    })

    expect(wrapper.find('div.p-8').exists()).toBe(true)
    expect(wrapper.find('[data-test="modify-product-index"]').exists()).toBe(true)
  })
})
