import { ref } from 'vue'
import { defineStore } from 'pinia'
import { registerUser, loginUser, logoutUser, getCurrentUser } from '@/network/request.js'
import logger from '@/utils/logger'

export const useAuthStore = defineStore('auth', () => {
  const isAuthenticated = ref(false)
  const user = ref(null)

  async function register(payload) {
    logger.Debug('auth.js', 'register', 'registering user', { email: payload.email })
    const res = await registerUser(payload)
    if (!res.ok) {
      logger.Warn('auth.js', 'register', 'registration failed', { error: res.message || res.error })
      return res
    }
    user.value = res.data.user
    isAuthenticated.value = true
    logger.Info('auth.js', 'register', 'user registered', { user_id: res.data.user.id })
    return res
  }

  async function login(payload) {
    logger.Debug('auth.js', 'login', 'logging in', { email: payload.email })
    const res = await loginUser(payload)
    if (!res.ok) {
      logger.Warn('auth.js', 'login', 'login failed', { error: res.message || res.error })
      return res
    }
    user.value = res.data.user
    isAuthenticated.value = true
    logger.Info('auth.js', 'login', 'login successful', { user_id: res.data.user.id })
    return res
  }

  async function logout() {
    logger.Debug('auth.js', 'logout', 'logging out')
    await logoutUser()
    user.value = null
    isAuthenticated.value = false
    logger.Info('auth.js', 'logout', 'logout successful')
  }

  async function fetchMe() {
    logger.Debug('auth.js', 'fetchMe', 'fetching current user')
    const res = await getCurrentUser()
    if (res.ok) {
      user.value = res.data.user
      isAuthenticated.value = true
      logger.Debug('auth.js', 'fetchMe', 'user restored', { user_id: res.data.user.id })
    } else {
      user.value = null
      isAuthenticated.value = false
      logger.Debug('auth.js', 'fetchMe', 'no active session')
    }
    return res
  }

  return { isAuthenticated, user, register, login, logout, fetchMe }
})
