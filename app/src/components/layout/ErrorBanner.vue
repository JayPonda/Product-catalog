<script setup>
import { useErrorStore } from '@/stores/errors'

const errors = useErrorStore()
</script>

<template>
  <div
    v-if="errors.message"
    role="alert"
    class="fixed top-0 left-0 right-0 z-50 flex items-center justify-between rounded-md border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700 shadow-lg"
  >
    <span>{{ errors.message }}</span>
    <button
      @click="errors.clear()"
      aria-label="Dismiss"
      class="rounded p-1 leading-none text-red-500 transition-colors hover:bg-red-100 hover:text-red-700 focus:outline-none"
    >
      &#10005;
    </button>
  </div>
</template>
