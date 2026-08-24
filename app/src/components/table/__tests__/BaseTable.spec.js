import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import { h } from 'vue'
import BaseTable from '../BaseTable.vue'

describe('BaseTable.vue', () => {
  const columns = [
    { key: 'name', label: 'Name' },
    { key: 'role', label: 'Role' },
  ]
  const items = [
    { id: '1', name: 'Alice', role: 'Admin' },
    { id: '2', name: 'Bob', role: 'User' },
  ]

  it('renders table headers and row items correctly', () => {
    const wrapper = mount(BaseTable, {
      props: { columns, items },
    })

    // Verify headers
    const headers = wrapper.findAll('th')
    expect(headers).toHaveLength(2)
    expect(headers[0]?.text()).toBe('Name')
    expect(headers[1]?.text()).toBe('Role')

    // Verify rows
    const rows = wrapper.findAll('tbody tr')
    expect(rows).toHaveLength(2)

    // Verify cell contents
    const row1Cells = rows[0]?.findAll('td')
    const row2Cells = rows[1]?.findAll('td')
    expect(row1Cells?.[0]?.text()).toBe('Alice')
    expect(row1Cells?.[1]?.text()).toBe('Admin')
    expect(row2Cells?.[0]?.text()).toBe('Bob')
    expect(row2Cells?.[1]?.text()).toBe('User')
  })

  it('renders empty state when no items are provided', () => {
    const wrapper = mount(BaseTable, {
      props: { columns, items: [] },
    })

    const rows = wrapper.findAll('tbody tr')
    expect(rows).toHaveLength(1)
    expect(rows[0]?.find('td').text()).toBe('No records found.')
    expect(rows[0]?.find('td').attributes('colspan')).toBe('2')
  })

  it('supports custom slots for headers', () => {
    const wrapper = mount(BaseTable, {
      props: { columns, items },
      slots: {
        'header-name': '<span class="custom-header">Custom Name</span>',
      },
    })

    expect(wrapper.find('.custom-header').exists()).toBe(true)
    expect(wrapper.find('.custom-header').text()).toBe('Custom Name')
  })

  it('supports custom slots for cells', () => {
    const wrapper = mount(BaseTable, {
      props: { columns, items },
      slots: {
        'cell-role': (slotProps) =>
          h('span', { class: 'badge' }, slotProps.value.toUpperCase()),
      },
    })

    const badge = wrapper.find('.badge')
    expect(badge.exists()).toBe(true)
    expect(badge.text()).toBe('ADMIN')
  })
})
