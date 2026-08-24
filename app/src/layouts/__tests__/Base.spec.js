import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia } from 'pinia'
import Base from '../Base.vue'

function mountLayout() {
  return mount(Base, {
    slots: {
      default: '<p data-test="page">page content</p>',
    },
    global: {
      plugins: [createPinia()],
      stubs: {
        Header: { template: '<header data-test="header" />' },
        Footer: { template: '<footer data-test="footer" />' },
      },
    },
  })
}

describe('Base layout', () => {
  it('renders header, slotted page content and footer', () => {
    const wrapper = mountLayout()

    expect(wrapper.find('[data-test="header"]').exists()).toBe(true)
    expect(wrapper.find('main [data-test="page"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="footer"]').exists()).toBe(true)
  })

  it('orders header, main and footer vertically', () => {
    const html = mountLayout().html()
    const headerAt = html.indexOf('data-test="header"')
    const pageAt = html.indexOf('data-test="page"')
    const footerAt = html.indexOf('data-test="footer"')

    expect(headerAt).toBeGreaterThanOrEqual(0)
    expect(headerAt).toBeLessThan(pageAt)
    expect(pageAt).toBeLessThan(footerAt)
  })
})
