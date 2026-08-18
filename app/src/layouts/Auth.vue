<template>
  <div class="flex min-h-screen flex-col">
    <Header class="sticky top-0 z-50" :is-logged-in="false" />

    <main class="flex-1">
      <slot />
    </main>

    <Footer class="" />
  </div>
</template>

<script setup lang="ts">
import Header from '@/components/layout/Header.vue'
import Footer from '@/components/layout/Footer.vue'
</script>