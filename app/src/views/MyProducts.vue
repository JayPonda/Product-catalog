<template>
  <div class="p-8">
    <Index :show-controls="auth.isAuthenticated" :my-products="true" />
  </div>
</template>

<script setup>
import Index from '@/components/Product/Index.vue'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
</script>
