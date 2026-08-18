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
        v-if="props.isLoggedIn"
          to="/my-products"
          class="text-sm font-medium text-emerald-50 transition hover:text-white"
        >
          My Products
        </RouterLink>

        <RouterLink
          to="/categories"
          class="text-sm font-medium text-emerald-50 transition hover:text-white"
        >
          Categories
        </RouterLink>
      </div>

      <!-- Desktop actions -->
      <div class="hidden items-center gap-3 md:flex">
        <RouterLink
          v-if="!props.isLoggedIn"
          to="/login"
          class="rounded-lg px-4 py-2 text-sm font-semibold text-white transition hover:bg-emerald-600"
        >
          Log in
        </RouterLink>
        <RouterLink
          v-else
          to="/logout"
          class="rounded-lg px-4 py-2 text-sm font-semibold text-white transition hover:bg-emerald-600"
        >
          Log out
        </RouterLink>

        <RouterLink
          to="/"
          class="flex items-center gap-2 rounded-lg bg-white px-4 py-2 text-sm font-semibold text-emerald-700 shadow-sm transition hover:bg-emerald-50"
        >
          Browse products
          <ArrowRight class="size-4" />
        </RouterLink>
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
          v-if="props.isLoggedIn"
          to="/my-products"
          class="block rounded-lg px-3 py-2.5 text-sm font-medium hover:bg-emerald-600"
          @click="mobileMenuOpen = false"
        >
          My Products
        </RouterLink>

        <RouterLink
          to="/categories"
          class="block rounded-lg px-3 py-2.5 text-sm font-medium hover:bg-emerald-600"
          @click="mobileMenuOpen = false"
        >
          Categories
        </RouterLink>

        <div class="my-2 border-t border-emerald-600" />

        <RouterLink
          v-if="!props.isLoggedIn"
          to="/login"
          class="block rounded-lg px-3 py-2.5 text-sm font-medium hover:bg-emerald-600"
          @click="mobileMenuOpen = false"
        >
          Log in
        </RouterLink>
        <RouterLink
          v-else
          to="/logout"
          class="block rounded-lg px-3 py-2.5 text-sm font-medium hover:bg-emerald-600"
          @click="mobileMenuOpen = false"
        >
          Log out
        </RouterLink>

        <RouterLink
          to="/"
          class="mt-2 flex items-center justify-center gap-2 rounded-lg bg-white px-4 py-2.5 text-sm font-semibold text-emerald-700"
          @click="mobileMenuOpen = false"
        >
          Browse products
          <ArrowRight class="size-4" />
        </RouterLink>
      </div>
    </div>
  </header>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { RouterLink } from 'vue-router'
import {
  Package,
  Menu,
  X,
  ArrowRight,
} from '@lucide/vue'

const mobileMenuOpen = ref(false)

const props = defineProps({
  isLoggedIn: {
    type: Boolean, 
    default: false,
    required: true
  },
})

</script>