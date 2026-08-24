import { describe, expect, it, beforeEach } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useErrorStore } from '../errors'

describe('errors store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('initializes with empty message', () => {
    const store = useErrorStore()
    expect(store.message).toBe('')
  })

  it('shows string error message correctly', () => {
    const store = useErrorStore()
    store.show('  An error occurred!  ')
    expect(store.message).toBe('  An error occurred!  ')
  })

  it('clears message when clear is called', () => {
    const store = useErrorStore()
    store.show('Error')
    store.clear()
    expect(store.message).toBe('')
  })

  it('shows default message for non-string truthy errors', () => {
    const store = useErrorStore()
    store.show(new Error('Raw error'))
    expect(store.message).toBe('Something went wrong.')
  })

  it('clears message for falsy/empty raw values', () => {
    const store = useErrorStore()
    store.show('Error')
    store.show(null)
    expect(store.message).toBe('')
  })
})
