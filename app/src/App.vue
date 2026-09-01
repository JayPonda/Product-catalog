<script setup>
import { defineAsyncComponent } from 'vue'
import { RouterView } from 'vue-router'
import { Notivue, Notification } from 'notivue'

// A single layout is used for every route; the Header itself decides what to
// render based on the live authentication state (see stores/auth.ts).
const AppLayout = defineAsyncComponent(() => import('@/layouts/Base.vue'))
</script>

<template>
  <Notivue v-slot="item">
    <Notification :item="item" />
  </Notivue>
  <AppLayout>
    <RouterView />
  </AppLayout>
</template>
