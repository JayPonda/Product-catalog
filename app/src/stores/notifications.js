import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import { push } from 'notivue'
import logger from '@/utils/logger'

const DEFAULT_ERROR_MESSAGE = 'Something went wrong.'
const DEFAULT_DURATION = 4000
const DEFAULT_ERROR_DURATION = 5000

const STATUS_CODE_MESSAGES = {
  400: 'Bad request.',
  401: 'Unauthorized.',
  403: 'Access forbidden.',
  404: 'Resource not found.',
  409: 'Conflict.',
  422: 'Validation error.',
  429: 'Too many requests. Please try again later.',
  500: 'Internal server error. Please try again later.',
  502: 'Bad gateway. Please try again later.',
  503: 'Service unavailable. Please try again later.',
  504: 'Gateway timeout. Please try again later.',
}

function resolveErrorMessage(raw) {
  if (raw === null || raw === undefined || raw === '') {
    return ''
  }

  if (raw instanceof Error) {
    return DEFAULT_ERROR_MESSAGE
  }

  let text = ''
  if (typeof raw === 'object' && raw !== null) {
    text =
      (typeof raw.message === 'string' && raw.message) ||
      (typeof raw.error === 'string' && raw.error) ||
      ''
    if (!text && typeof raw.error === 'number') {
      text = STATUS_CODE_MESSAGES[raw.error] || DEFAULT_ERROR_MESSAGE
    }
  } else {
    text = String(raw)
  }

  const trimmed = text.trim()
  if (STATUS_CODE_MESSAGES[trimmed]) {
    return STATUS_CODE_MESSAGES[trimmed]
  }

  if (/^\d{3}$/.test(trimmed)) {
    return DEFAULT_ERROR_MESSAGE
  }

  return text
}

export const useNotificationStore = defineStore('notifications', () => {
  const currentMessage = ref('')
  const notifications = ref([])
  const timers = new Map()
  let counter = 0

  const message = computed(() => currentMessage.value)

  function remove(idOrObj) {
    const id = typeof idOrObj === 'object' && idOrObj !== null ? idOrObj.id : idOrObj
    if (timers.has(id)) {
      clearTimeout(timers.get(id))
      timers.delete(id)
    }
    const idx = notifications.value.findIndex((n) => n.id === id)
    if (idx !== -1) {
      const item = notifications.value[idx]
      item?.notivueItem?.destroy?.()
      notifications.value.splice(idx, 1)
      if (notifications.value.length === 0) {
        currentMessage.value = ''
      } else {
        currentMessage.value = notifications.value[notifications.value.length - 1].message
      }
      logger.Debug('notifications.js', 'remove', `notification removed id=${id}`)
    }
  }

  function clear() {
    for (const timer of timers.values()) {
      clearTimeout(timer)
    }
    timers.clear()
    push.clearAll()
    notifications.value = []
    currentMessage.value = ''
    logger.Debug('notifications.js', 'clear', 'all notifications cleared')
  }

  function add({ message: msg, type = 'info', title = '', duration } = {}) {
    if (!msg || (typeof msg === 'string' && !msg.trim())) {
      return null
    }

    const text = typeof msg === 'string' ? msg : DEFAULT_ERROR_MESSAGE
    const dur = duration ?? (type === 'error' ? DEFAULT_ERROR_DURATION : DEFAULT_DURATION)
    currentMessage.value = text

    const options = {
      message: text,
      duration: dur,
      ...(title ? { title } : {}),
    }

    let notivueItem
    if (type === 'error') {
      logger.Warn('notifications.js', 'add', text, { type })
      notivueItem = push.error(options)
    } else if (type === 'success') {
      logger.Info('notifications.js', 'add', text, { type })
      notivueItem = push.success(options)
    } else if (type === 'warning') {
      logger.Warn('notifications.js', 'add', text, { type })
      notivueItem = push.warning(options)
    } else {
      logger.Info('notifications.js', 'add', text, { type })
      notivueItem = push.info(options)
    }

    const id = notivueItem?.id ?? `${Date.now()}-${++counter}`

    const record = {
      id,
      message: text,
      type,
      duration: dur,
      notivueItem,
    }

    notifications.value.push(record)

    if (dur > 0 && dur !== Infinity) {
      const timer = setTimeout(() => {
        remove(id)
      }, dur)
      timers.set(id, timer)
    }

    return id
  }

  function error(msg, options = {}) {
    const opts = typeof options === 'number' ? { duration: options } : options
    const text = resolveErrorMessage(msg) || DEFAULT_ERROR_MESSAGE
    return add({ message: text, type: 'error', ...opts })
  }

  function success(msg, options = {}) {
    const opts = typeof options === 'number' ? { duration: options } : options
    return add({ message: msg, type: 'success', ...opts })
  }

  function warning(msg, options = {}) {
    const opts = typeof options === 'number' ? { duration: options } : options
    return add({ message: msg, type: 'warning', ...opts })
  }

  function info(msg, options = {}) {
    const opts = typeof options === 'number' ? { duration: options } : options
    return add({ message: msg, type: 'info', ...opts })
  }

  function notify(msg, type = 'info', options = {}) {
    const opts = typeof options === 'number' ? { duration: options } : options
    return add({ message: msg, type, ...opts })
  }

  function show(raw, type = 'error', options = {}) {
    const opts = typeof options === 'number' ? { duration: options } : options
    const text = resolveErrorMessage(raw)
    if (!text) {
      clear()
      return null
    }
    return notify(text, type, opts)
  }

  return {
    notifications,
    message,
    add,
    notify,
    error,
    success,
    warning,
    info,
    show,
    remove,
    clear,
  }
})
