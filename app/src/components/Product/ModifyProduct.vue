<template>
  <div class="space-y-6">
    <!-- error alert -->
    <div
      v-if="error"
      role="alert"
      class="flex items-center justify-between rounded-md border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700"
    >
      <span>{{ error }}</span>
      <button
        @click="error = ''"
        aria-label="Dismiss"
        class="ml-4 rounded p-1 leading-none text-red-500 transition-colors hover:bg-red-100 hover:text-red-700 focus:outline-none"
      >
        &#10005;
      </button>
    </div>

    <!-- Product Details Form -->
    <ProductForm :productId="productId" @saved="onProductSaved" @error="onError" />

    <!-- Category Manager (Only visible once product is saved) -->
    <CategoryLinker v-if="productId" :productId="productId" @error="onError" />
  </div>
</template>

<script setup>
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import ProductForm from './ProductForm.vue'
import CategoryLinker from './CategoryLinker.vue'
import logger from '@/utils/logger'

const route = useRoute()
const router = useRouter()

const error = ref('')

const productId = computed(() => {
  const id = route.params.id
  return Array.isArray(id) ? id[0] : id
})

function onError(message) {
  error.value = message
  if (message) {
    logger.Warn('ModifyProduct.vue', 'onError', String(message))
  }
}

function onProductSaved(savedProduct) {
  error.value = ''
  logger.Info('ModifyProduct.vue', 'onProductSaved', 'product saved', { id: savedProduct?.id })
  router.push({
    name: 'products-modify',
    params: { id: savedProduct.id },
    hash: '#category',
  })
}
</script>
