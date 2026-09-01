<template>
  <div class="space-y-6">
    <!-- Product Details Form -->
    <ProductForm :productId="productId" @saved="onProductSaved" @error="onError" />

    <!-- Category Manager (Only visible once product is saved) -->
    <CategoryLinker v-if="productId" :productId="productId" @error="onError" />
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useNotificationStore } from '@/stores/notifications'
import ProductForm from './ProductForm.vue'
import CategoryLinker from './CategoryLinker.vue'
import logger from '@/utils/logger'

const route = useRoute()
const router = useRouter()
const notifications = useNotificationStore()

const productId = computed(() => {
  const id = route.params.id
  return Array.isArray(id) ? id[0] : id
})

function onError(message) {
  if (message) {
    logger.Warn('ModifyProduct.vue', 'onError', String(message))
    notifications.error(String(message))
  }
}

function onProductSaved(savedProduct, isEditParam) {
  const isEdit = isEditParam !== undefined ? Boolean(isEditParam) : Boolean(productId.value)
  logger.Info('ModifyProduct.vue', 'onProductSaved', 'product saved', {
    id: savedProduct?.id,
    isEdit,
  })
  if (isEdit) {
    notifications.success('Product successfully edited.')
  } else {
    notifications.success('Product successfully added.')
  }
  router.push({
    name: 'products-modify',
    params: { id: savedProduct.id },
    hash: '#category',
  })
}
</script>
