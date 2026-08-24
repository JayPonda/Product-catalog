import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import ModifyCategory from '../ModifyCategory.vue'

describe('ModifyCategory.vue', () => {
  it('renders correctly', () => {
    const wrapper = mount(ModifyCategory)
    expect(wrapper.find('div.p-4').exists()).toBe(true)
  })
})
