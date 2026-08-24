import { ref } from 'vue'
import { defineStore } from 'pinia'

const DEFAULT_MESSAGE = 'Something went wrong.'

export const useErrorStore = defineStore('errors', () => {
  const message = ref('')

  function show(raw) {
    if (typeof raw === 'string') {
      message.value = raw.trim() ? raw : ''
      return
    }
    if (raw === null || raw === undefined || raw === '') {
      message.value = ''
      return
    }
    message.value = DEFAULT_MESSAGE
  }

  function clear() {
    message.value = ''
  }

  return { message, show, clear }
})
