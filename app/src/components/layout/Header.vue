<template>
  <header
    class="border-b border-emerald-900/10 bg-emerald-700 text-white shadow-sm"
  >
    <nav
      class="mx-auto flex h-16 max-w-7xl items-center justify-between px-4 sm:px-6 lg:px-8"
      aria-label="Main navigation"
    >
      <!-- Logo -->
      <RouterLink
        to="/"
        class="flex items-center gap-2"
      >
        <div
          class="flex size-9 items-center justify-center rounded-lg bg-white text-emerald-700"
        >
          <Package class="size-5" />
        </div>

        <span class="text-lg font-bold tracking-tight">
          Product Catalog
        </span>
      </RouterLink>

      <!-- Desktop navigation -->
      <div class="hidden items-center gap-8 md:flex">
        <RouterLink
          to="/my-products"
          v-if="auth.isAuthenticated"
          active-class="text-white underline underline-offset-4"
          class="text-sm font-medium text-emerald-50 transition hover:text-white"
        >
          My Products
        </RouterLink>
        <RouterLink
          to="/categories"
          active-class="text-white underline underline-offset-4"
          class="text-sm font-medium text-emerald-50 transition hover:text-white"
        >
          Categories
        </RouterLink>
      </div>

      <!-- Desktop actions -->
      <div class="hidden items-center gap-3 md:flex">
        <template v-if="auth.isAuthenticated">
          <span class="text-sm font-medium text-emerald-50">{{ auth.user?.email }}</span>
          <button
            type="button"
            @click="handleLogout"
            class="rounded-lg bg-white px-3 py-2 text-sm font-bold text-emerald-700 transition hover:bg-emerald-50"
          >
            Logout
          </button>
        </template>
        <template v-else>
          <RouterLink
            to="/login"
            class="rounded-lg px-3 py-2 text-sm font-medium text-emerald-50 transition hover:bg-emerald-600 hover:text-white"
          >
            Login
          </RouterLink>
          <RouterLink
            to="/register"
            class="rounded-lg bg-white px-3 py-2 text-sm font-bold text-emerald-700 transition hover:bg-emerald-50"
          >
            Register
          </RouterLink>
        </template>
      </div>

      <!-- Mobile menu button -->
      <button
        type="button"
        class="rounded-lg p-2 text-white transition hover:bg-emerald-600 md:hidden"
        aria-label="Open menu"
        @click="mobileMenuOpen = !mobileMenuOpen"
      >
        <X
          v-if="mobileMenuOpen"
          class="size-6"
        />

        <Menu
          v-else
          class="size-6"
        />
      </button>
    </nav>

    <!-- Mobile navigation -->
    <div
      v-if="mobileMenuOpen"
      class="border-t border-emerald-600 bg-emerald-700 px-4 py-4 md:hidden"
    >
      <div class="mx-auto max-w-7xl space-y-1">
        <RouterLink
          to="/my-products"
          v-if="auth.isAuthenticated"
          active-class="bg-emerald-600 text-white"
          class="block rounded-lg px-3 py-2.5 text-sm font-medium hover:bg-emerald-600"
          @click="mobileMenuOpen = false"
        >
          My Products
        </RouterLink>

        <RouterLink
          to="/categories"
          active-class="bg-emerald-600 text-white"
          class="block rounded-lg px-3 py-2.5 text-sm font-medium hover:bg-emerald-600"
          @click="mobileMenuOpen = false"
        >
          Categories
        </RouterLink>

        <div class="my-2 border-t border-emerald-600" />

        <template v-if="auth.isAuthenticated">
          <p class="px-3 py-2 text-sm font-medium text-emerald-50">{{ auth.user?.email }}</p>
          <button
            type="button"
            @click="handleLogout"
            class="block w-full rounded-lg px-3 py-2.5 text-left text-sm font-medium hover:bg-emerald-600"
          >
            Logout
          </button>
        </template>
        <template v-else>
          <RouterLink
            to="/login"
            class="block rounded-lg px-3 py-2.5 text-sm font-medium hover:bg-emerald-600"
            @click="mobileMenuOpen = false"
          >
            Login
          </RouterLink>
          <RouterLink
            to="/register"
            class="block rounded-lg px-3 py-2.5 text-sm font-medium hover:bg-emerald-600"
            @click="mobileMenuOpen = false"
          >
            Register
          </RouterLink>
        </template>
      </div>
    </div>
  </header>
</template>

<script setup>
import { ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import {
  Package,
  Menu,
  X,
} from '@lucide/vue'

const mobileMenuOpen = ref(false)
const router = useRouter()
const auth = useAuthStore()

async function handleLogout() {
  mobileMenuOpen.value = false
  await auth.logout()
  router.push('/login')
}
</script>
