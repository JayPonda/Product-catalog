<template>
  <div class="flex min-h-[70vh] items-center justify-center px-4">
    <div class="w-full max-w-md space-y-6 rounded-lg border border-gray-200 bg-white p-8 shadow-sm">
      <div class="text-center">
        <h1 class="text-2xl font-bold text-gray-900">Welcome back</h1>
        <p class="mt-1 text-sm text-gray-500">Sign in to your account</p>
      </div>

      <form class="space-y-4" @submit.prevent="submit">
        <div>
          <label for="email" class="block text-sm font-medium text-gray-700"
            >Email <span class="text-red-500">*</span></label
          >
          <input
            id="email"
            type="email"
            v-model="form.email"
            autocomplete="email"
            :class="[
              'mt-1 block w-full rounded-md px-3 py-2 text-sm focus:outline-none focus-visible:ring-2',
              fieldErrors.email
                ? 'border border-red-300 focus-visible:ring-red-500'
                : 'border border-gray-300 focus-visible:ring-emerald-500',
            ]"
          />
          <p v-if="fieldErrors.email" class="mt-1 text-sm text-red-600">{{ fieldErrors.email }}</p>
        </div>

        <div>
          <label for="password" class="block text-sm font-medium text-gray-700"
            >Password <span class="text-red-500">*</span></label
          >
          <div class="relative">
            <input
              id="password"
              :type="showPassword ? 'text' : 'password'"
              v-model="form.password"
              autocomplete="current-password"
              :class="[
                'mt-1 block w-full rounded-md px-3 py-2 pr-10 text-sm focus:outline-none focus-visible:ring-2',
                fieldErrors.password
                  ? 'border border-red-300 focus-visible:ring-red-500'
                  : 'border border-gray-300 focus-visible:ring-emerald-500',
              ]"
            />
            <button
              type="button"
              @click="showPassword = !showPassword"
              class="absolute inset-y-0 right-0 flex items-center pr-3 text-gray-400 transition-colors hover:text-gray-600"
              :aria-label="showPassword ? 'Hide password' : 'Show password'"
            >
              <EyeOff v-if="showPassword" class="h-5 w-5" />
              <Eye v-else class="h-5 w-5" />
            </button>
          </div>
          <p v-if="fieldErrors.password" class="mt-1 text-sm text-red-600">
            {{ fieldErrors.password }}
          </p>
        </div>

        <button
          type="submit"
          :disabled="loading"
          class="w-full rounded-md bg-emerald-700 px-4 py-2 font-bold text-white transition-colors hover:bg-emerald-800 focus:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500 disabled:opacity-50"
        >
          {{ loading ? 'Signing in…' : 'Sign in' }}
        </button>
      </form>

      <p class="text-center text-sm text-gray-500">
        No account?
        <RouterLink to="/register" class="font-medium text-emerald-700 hover:underline"
          >Create one</RouterLink
        >
      </p>
    </div>
  </div>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useNotificationStore } from '@/stores/notifications'
import { Eye, EyeOff } from '@lucide/vue'
import logger from '@/utils/logger'

const router = useRouter()
const auth = useAuthStore()
const notifications = useNotificationStore()

const loading = ref(false)
const showPassword = ref(false)
const fieldErrors = reactive({ email: '', password: '' })
const form = reactive({ email: '', password: '' })

function validate() {
  fieldErrors.email = form.email.trim() ? '' : 'Email is required.'
  fieldErrors.password = form.password ? '' : 'Password is required.'
  return !fieldErrors.email && !fieldErrors.password
}

async function submit() {
  if (!validate()) return

  logger.Debug('Login.vue', 'submit', 'login attempt', { email: form.email.trim() })
  loading.value = true
  const res = await auth.login({ email: form.email.trim(), password: form.password })
  loading.value = false

  if (!res.ok) {
    logger.Warn('Login.vue', 'submit', 'login failed', { error: res.message || 'Login failed.' })
    notifications.error(res.message || 'Login failed.')
    return
  }

  logger.Info('Login.vue', 'submit', 'login successful')
  notifications.success('Welcome back!')
  router.push('/products')
}
</script>
