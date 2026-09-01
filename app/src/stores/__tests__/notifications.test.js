import { describe, expect, it, beforeEach, afterEach, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useNotificationStore } from '../notifications'

describe('notifications store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.restoreAllMocks()
    vi.useRealTimers()
  })

  it('initializes with empty notifications and empty message', () => {
    const store = useNotificationStore()
    expect(store.notifications).toEqual([])
    expect(store.message).toBe('')
  })

  it('adds a notification with default properties', () => {
    const store = useNotificationStore()
    const id = store.add({ message: 'Hello world' })

    expect(id).toBeDefined()
    expect(store.notifications).toHaveLength(1)
    expect(store.notifications[0]).toMatchObject({
      id,
      message: 'Hello world',
      type: 'info',
      duration: 4000,
    })
    expect(store.message).toBe('Hello world')
  })

  it('ignores empty or whitespace-only messages', () => {
    const store = useNotificationStore()
    expect(store.add({ message: '' })).toBeNull()
    expect(store.add({ message: '   ' })).toBeNull()
    expect(store.add({ message: null })).toBeNull()
    expect(store.notifications).toHaveLength(0)
  })

  it('provides convenient type helpers', () => {
    const store = useNotificationStore()

    store.error('An error occurred')
    expect(store.notifications[0]).toMatchObject({
      message: 'An error occurred',
      type: 'error',
      duration: 5000,
    })

    store.success('Success message')
    expect(store.notifications[1]).toMatchObject({
      message: 'Success message',
      type: 'success',
      duration: 4000,
    })

    store.warning('Warning message')
    expect(store.notifications[2]).toMatchObject({
      message: 'Warning message',
      type: 'warning',
      duration: 4000,
    })

    store.info('Info message')
    expect(store.notifications[3]).toMatchObject({
      message: 'Info message',
      type: 'info',
      duration: 4000,
    })

    store.notify('General notice')
    expect(store.notifications[4]).toMatchObject({
      message: 'General notice',
      type: 'info',
      duration: 4000,
    })
  })

  it('auto-dismisses notification after duration expires', () => {
    const store = useNotificationStore()
    store.add({ message: 'Will expire', duration: 3000 })

    expect(store.notifications).toHaveLength(1)

    vi.advanceTimersByTime(2999)
    expect(store.notifications).toHaveLength(1)

    vi.advanceTimersByTime(1)
    expect(store.notifications).toHaveLength(0)
  })

  it('removes specific notification by id', () => {
    const store = useNotificationStore()
    const id1 = store.info('First')
    const id2 = store.info('Second')

    expect(store.notifications).toHaveLength(2)

    store.remove(id1)
    expect(store.notifications).toHaveLength(1)
    expect(store.notifications[0].id).toBe(id2)
  })

  it('clears all notifications and timers', () => {
    const store = useNotificationStore()
    store.info('First')
    store.info('Second')
    expect(store.notifications).toHaveLength(2)

    store.clear()
    expect(store.notifications).toHaveLength(0)
    expect(store.message).toBe('')
  })

  it('handles show() compatibility helper for various input types', () => {
    const store = useNotificationStore()

    // String input
    store.show('Legacy error message')
    expect(store.message).toBe('Legacy error message')
    expect(store.notifications[0].type).toBe('error')

    // Falsy input clears
    store.show(null)
    expect(store.notifications).toHaveLength(0)

    // Non-string truthy error object
    store.show(new Error('Boom'))
    expect(store.message).toBe('Something went wrong.')
    expect(store.notifications[0].type).toBe('error')

    // Empty string clears
    store.show('')
    expect(store.notifications).toHaveLength(0)

    // Status code translations (500 and 429)
    store.show(500)
    expect(store.message).toBe('Internal server error. Please try again later.')

    store.show('429')
    expect(store.message).toBe('Too many requests. Please try again later.')

    // Backend error object with message
    store.show({ message: 'Custom rate limit message' })
    expect(store.message).toBe('Custom rate limit message')

    // Backend error object with error field
    store.show({ error: 'Backend error string' })
    expect(store.message).toBe('Backend error string')
  })
})
