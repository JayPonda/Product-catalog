import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import Footer from '../Footer.vue'

describe('Footer', () => {
  it('shows the product name and the current year', () => {
    const wrapper = mount(Footer)

    expect(wrapper.text()).toContain(
      `© ${new Date().getFullYear()} Product Catalog`,
    )
  })
})
