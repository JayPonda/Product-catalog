import { ref } from 'vue'
import { defineStore } from 'pinia'
import { registerUser, loginUser, logoutUser, getCurrentUser } from '@/network/request.js'

type AuthPayload = Record<string, string>

export const useAuthStore = defineStore('auth', () => {
  const isAuthenticated = ref(false)
  const user = ref<null | Record<string, string>>(null)

  async function register(payload: AuthPayload) {
    const res = await registerUser(payload)
    if (!res.ok) {
      return res
    }
    user.value = res.data.user
    isAuthenticated.value = true
    return res
  }

  async function login(payload: AuthPayload) {
    const res = await loginUser(payload)
    if (!res.ok) {
      return res
    }
    user.value = res.data.user
    isAuthenticated.value = true
    return res
  }

  async function logout() {
    await logoutUser()
    user.value = null
    isAuthenticated.value = false
  }

  // Restore auth state from the httpOnly cookie (called on app boot).
  async function fetchMe() {
    const res = await getCurrentUser()
    if (res.ok) {
      user.value = res.data.user
      isAuthenticated.value = true
    } else {
      user.value = null
      isAuthenticated.value = false
    }
    return res
  }

  return { isAuthenticated, user, register, login, logout, fetchMe }
})
