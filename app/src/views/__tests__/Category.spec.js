import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import Category from '../Category.vue'

const IndexStub = {
  name: 'Index',
  template: '<div data-test="category-index" />',
}

describe('Category view', () => {
  it('renders the category index inside a padded container', () => {
    const wrapper = mount(Category, {
      global: { stubs: { Index: IndexStub } },
    })

    expect(wrapper.find('div.p-8').exists()).toBe(true)
    expect(wrapper.find('[data-test="category-index"]').exists()).toBe(true)
  })
})
