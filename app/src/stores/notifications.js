import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import { push } from 'notivue'
import logger from '@/utils/logger'

const DEFAULT_ERROR_MESSAGE = 'Something went wrong.'
const DEFAULT_DURATION = 4000
const DEFAULT_ERROR_DURATION = 5000

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
    return add({ message: msg, type: 'error', ...opts })
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
    if (typeof raw === 'string') {
      if (!raw.trim()) {
        clear()
        return null
      }
      return notify(raw, type, opts)
    }
    if (raw === null || raw === undefined || raw === '') {
      clear()
      return null
    }
    return error(DEFAULT_ERROR_MESSAGE, opts)
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
