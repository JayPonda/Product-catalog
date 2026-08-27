import { ref } from 'vue'
import { defineStore } from 'pinia'
import logger from '@/utils/logger'

const DEFAULT_MESSAGE = 'Something went wrong.'

export const useErrorStore = defineStore('errors', () => {
  const message = ref('')

  function show(raw) {
    if (typeof raw === 'string') {
      message.value = raw.trim() ? raw : ''
      if (message.value) {
        logger.Warn('errors.js', 'show', message.value)
      } else {
        logger.Debug('errors.js', 'show', 'error cleared')
      }
      return
    }
    if (raw === null || raw === undefined || raw === '') {
      message.value = ''
      logger.Debug('errors.js', 'show', 'error cleared')
      return
    }
    message.value = DEFAULT_MESSAGE
    logger.Warn('errors.js', 'show', DEFAULT_MESSAGE)
  }

  function clear() {
    message.value = ''
    logger.Debug('errors.js', 'clear', 'error cleared')
  }

  return { message, show, clear }
})
