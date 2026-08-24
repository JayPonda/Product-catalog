import { describe, expect, it, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import ErrorBanner from '../ErrorBanner.vue'
import { useErrorStore } from '@/stores/errors'

describe('ErrorBanner.vue', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('renders nothing when there is no error message', () => {
    const wrapper = mount(ErrorBanner)
    expect(wrapper.find('[role="alert"]').exists()).toBe(false)
  })

  it('renders error message and clears it on click', async () => {
    const errorStore = useErrorStore()
    errorStore.show('Test Error message')

    const wrapper = mount(ErrorBanner)
    expect(wrapper.find('[role="alert"]').exists()).toBe(true)
    expect(wrapper.text()).toContain('Test Error message')

    const button = wrapper.find('button[aria-label="Dismiss"]')
    expect(button.exists()).toBe(true)
    await button.trigger('click')

    expect(errorStore.message).toBe('')
  })
})
