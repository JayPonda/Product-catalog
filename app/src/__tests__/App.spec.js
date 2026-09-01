import { describe, expect, it, beforeAll, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia } from 'pinia'
import { createNotivue } from 'notivue'
import App from '../App.vue'

describe('App.vue', () => {
  beforeAll(() => {
    Object.defineProperty(window, 'matchMedia', {
      writable: true,
      value: vi.fn().mockImplementation((query) => ({
        matches: false,
        media: query,
        onchange: null,
        addListener: vi.fn(),
        removeListener: vi.fn(),
        addEventListener: vi.fn(),
        removeEventListener: vi.fn(),
        dispatchEvent: vi.fn(),
      })),
    })

    global.ResizeObserver = class ResizeObserver {
      observe() {}
      unobserve() {}
      disconnect() {}
    }
  })

  it('mounts App with Notivue container without errors', () => {
    const wrapper = mount(App, {
      global: {
        plugins: [createPinia(), createNotivue()],
        stubs: {
          RouterView: { template: '<div data-test="router-view" />' },
          AppLayout: { template: '<div><slot /></div>' },
        },
      },
    })
    expect(wrapper.exists()).toBe(true)
  })
})
