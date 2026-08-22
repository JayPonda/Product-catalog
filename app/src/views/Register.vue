<template>
  <div class="flex min-h-[70vh] items-center justify-center px-4">
    <div class="w-full max-w-md space-y-6 rounded-lg border border-gray-200 bg-white p-8 shadow-sm">
      <div class="text-center">
        <h1 class="text-2xl font-bold text-gray-900">Create your account</h1>
        <p class="mt-1 text-sm text-gray-500">Register to get started</p>
      </div>

      <div v-if="error" role="alert"
        class="flex items-center justify-between rounded-md border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
        <span>{{ error }}</span>
        <button @click="error = ''" aria-label="Dismiss"
          class="ml-4 rounded p-1 leading-none text-red-500 transition-colors hover:bg-red-100 hover:text-red-700 focus:outline-none">
          &#10005;
        </button>
      </div>

      <form class="space-y-4" @submit.prevent="submit">
        <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <div>
            <label for="first_name" class="block text-sm font-medium text-gray-700">First name <span class="text-red-500">*</span></label>
            <input id="first_name" type="text" v-model="form.first_name" autocomplete="given-name"
              :class="['mt-1 block w-full rounded-md px-3 py-2 text-sm focus:outline-none focus-visible:ring-2', fieldErrors.first_name ? 'border border-red-300 focus-visible:ring-red-500' : 'border border-gray-300 focus-visible:ring-emerald-500']" />
            <p v-if="fieldErrors.first_name" class="mt-1 text-sm text-red-600">{{ fieldErrors.first_name }}</p>
          </div>

          <div>
            <label for="last_name" class="block text-sm font-medium text-gray-700">Last name <span class="text-red-500">*</span></label>
            <input id="last_name" type="text" v-model="form.last_name" autocomplete="family-name"
              :class="['mt-1 block w-full rounded-md px-3 py-2 text-sm focus:outline-none focus-visible:ring-2', fieldErrors.last_name ? 'border border-red-300 focus-visible:ring-red-500' : 'border border-gray-300 focus-visible:ring-emerald-500']" />
            <p v-if="fieldErrors.last_name" class="mt-1 text-sm text-red-600">{{ fieldErrors.last_name }}</p>
          </div>
        </div>

        <div>
          <label for="email" class="block text-sm font-medium text-gray-700">Email <span class="text-red-500">*</span></label>
          <input id="email" type="email" v-model="form.email" autocomplete="email"
            :class="['mt-1 block w-full rounded-md px-3 py-2 text-sm focus:outline-none focus-visible:ring-2', fieldErrors.email ? 'border border-red-300 focus-visible:ring-red-500' : 'border border-gray-300 focus-visible:ring-emerald-500']" />
          <p v-if="fieldErrors.email" class="mt-1 text-sm text-red-600">{{ fieldErrors.email }}</p>
        </div>

        <div>
          <label for="password" class="block text-sm font-medium text-gray-700">Password <span class="text-red-500">*</span></label>
          <div class="relative">
            <input id="password" :type="showPassword ? 'text' : 'password'" v-model="form.password" autocomplete="new-password"
              :class="['mt-1 block w-full rounded-md px-3 py-2 pr-10 text-sm focus:outline-none focus-visible:ring-2', fieldErrors.password ? 'border border-red-300 focus-visible:ring-red-500' : 'border border-gray-300 focus-visible:ring-emerald-500']" />
            <button type="button" @click="showPassword = !showPassword"
              class="absolute inset-y-0 right-0 flex items-center pr-3 text-gray-400 transition-colors hover:text-gray-600"
              :aria-label="showPassword ? 'Hide password' : 'Show password'">
              <EyeOff v-if="showPassword" class="h-5 w-5" />
              <Eye v-else class="h-5 w-5" />
            </button>
          </div>
          <p v-if="fieldErrors.password" class="mt-1 text-sm text-red-600">{{ fieldErrors.password }}</p>
        </div>

        <div>
          <label for="confirm_password" class="block text-sm font-medium text-gray-700">Confirm password <span class="text-red-500">*</span></label>
          <div class="relative">
            <input id="confirm_password" :type="showConfirm ? 'text' : 'password'" v-model="form.confirm_password" autocomplete="new-password"
              :class="['mt-1 block w-full rounded-md px-3 py-2 pr-10 text-sm focus:outline-none focus-visible:ring-2', fieldErrors.confirm_password ? 'border border-red-300 focus-visible:ring-red-500' : 'border border-gray-300 focus-visible:ring-emerald-500']" />
            <button type="button" @click="showConfirm = !showConfirm"
              class="absolute inset-y-0 right-0 flex items-center pr-3 text-gray-400 transition-colors hover:text-gray-600"
              :aria-label="showConfirm ? 'Hide password' : 'Show password'">
              <EyeOff v-if="showConfirm" class="h-5 w-5" />
              <Eye v-else class="h-5 w-5" />
            </button>
          </div>
          <p v-if="fieldErrors.confirm_password" class="mt-1 text-sm text-red-600">{{ fieldErrors.confirm_password }}</p>
        </div>

        <button type="submit" :disabled="loading"
          class="w-full rounded-md bg-emerald-700 px-4 py-2 font-bold text-white transition-colors hover:bg-emerald-800 focus:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500 disabled:opacity-50">
          {{ loading ? 'Creating account…' : 'Create account' }}
        </button>
      </form>

      <p class="text-center text-sm text-gray-500">
        Already have an account?
        <RouterLink to="/login" class="font-medium text-emerald-700 hover:underline">Sign in</RouterLink>
      </p>
    </div>
  </div>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { Eye, EyeOff } from '@lucide/vue'

const router = useRouter()
const auth = useAuthStore()

const loading = ref(false)
const showPassword = ref(false)
const showConfirm = ref(false)
const error = ref('')
const fieldErrors = reactive({ first_name: '', last_name: '', email: '', password: '', confirm_password: '' })
const form = reactive({ first_name: '', last_name: '', email: '', password: '', confirm_password: '' })

function validate() {
  fieldErrors.first_name = form.first_name.trim() ? '' : 'First name is required.'
  fieldErrors.last_name = form.last_name.trim() ? '' : 'Last name is required.'
  fieldErrors.email = /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(form.email.trim()) ? '' : 'A valid email is required.'
  fieldErrors.password = form.password.length >= 8 ? '' : 'Password must be at least 8 characters.'
  if (!form.confirm_password) {
    fieldErrors.confirm_password = 'Please confirm your password.'
  } else if (form.confirm_password !== form.password) {
    fieldErrors.confirm_password = 'Passwords do not match.'
  } else {
    fieldErrors.confirm_password = ''
  }
  return !fieldErrors.first_name && !fieldErrors.last_name && !fieldErrors.email && !fieldErrors.password && !fieldErrors.confirm_password
}

async function submit() {
  error.value = ''
  if (!validate()) return

  loading.value = true
  const res = await auth.register({
    first_name: form.first_name.trim(),
    last_name: form.last_name.trim(),
    email: form.email.trim(),
    password: form.password,
  })
  loading.value = false

  if (!res.ok) {
    error.value = res.message || 'Registration failed.'
    return
  }

  router.push('/login')
}
</script>
