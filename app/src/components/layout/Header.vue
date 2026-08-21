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
          to="/products/"
          active-class="text-white underline underline-offset-4"
          class="text-sm font-medium text-emerald-50 transition hover:text-white"
        >
          Products
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
          to="/products"
          active-class="bg-emerald-600 text-white"
          class="block rounded-lg px-3 py-2.5 text-sm font-medium hover:bg-emerald-600"
          @click="mobileMenuOpen = false"
        >
          Products
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

      </div>
    </div>
  </header>
</template>

<script setup>
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