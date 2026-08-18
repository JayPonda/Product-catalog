<script setup lang="ts">
const currentYear = new Date().getFullYear()
</script>

<template>
  <footer class="border-t border-slate-800 bg-slate-900">
    <div
      class="mx-auto max-w-7xl px-6 py-6 text-center text-sm text-slate-400 lg:px-8"
    >
      © {{ currentYear }} Product Catalog. All rights reserved.
    </div>
  </footer>
</template>