<script setup lang="ts">
import { defineAsyncComponent, computed, type DefineComponent } from 'vue';
import { RouterView, useRoute } from 'vue-router'

const layouts = {
  DefaultLayout: defineAsyncComponent(() => import('@/layouts/Base.vue')),
  LoginLayout: defineAsyncComponent(() => import('@/layouts/Auth.vue'))
}

const route = useRoute()
const layout = computed(() => {
  const layoutName = route.meta.layout ?? 'DefaultLayout'
  return layouts[layoutName]
})

</script>

<template>
  <component :is="layout">
    <RouterView />
  </component>
</template>