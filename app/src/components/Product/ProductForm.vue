<template>
  <form
    class="space-y-4 rounded-lg border border-gray-200 bg-white p-6"
    @submit.prevent="saveProduct"
  >
    <div>
      <label for="product-name" class="block text-sm font-medium text-gray-700"
        >Product name <span class="text-red-500">*</span></label
      >
      <input
        id="product-name"
        type="text"
        v-model="formData.name"
        @input="fieldErrors.name = ''"
        :class="[
          'mt-1 block w-full rounded-md px-3 py-2 text-sm focus:outline-none focus-visible:ring-2',
          fieldErrors.name
            ? 'border border-red-300 focus-visible:ring-red-500'
            : 'border border-gray-300 focus-visible:ring-emerald-500',
        ]"
      />
      <p v-if="fieldErrors.name" class="mt-1 text-sm text-red-600">{{ fieldErrors.name }}</p>
    </div>

    <div>
      <label for="product-description" class="block text-sm font-medium text-gray-700"
        >Description <span class="text-red-500">*</span></label
      >
      <textarea
        id="product-description"
        v-model="formData.description"
        rows="3"
        class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-emerald-500 focus:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500"
      ></textarea>
      <p v-if="fieldErrors.description" class="mt-1 text-sm text-red-600">
        {{ fieldErrors.description }}
      </p>
    </div>

    <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
      <div>
        <label for="product-stock" class="block text-sm font-medium text-gray-700"
          >Stock quantity <span class="text-red-500">*</span></label
        >
        <input
          id="product-stock"
          type="number"
          min="0"
          v-model="formData.stock_quantity"
          @input="fieldErrors.stock_quantity = ''"
          :class="[
            'mt-1 block w-full rounded-md px-3 py-2 text-sm focus:outline-none focus-visible:ring-2',
            fieldErrors.stock_quantity
              ? 'border border-red-300 focus-visible:ring-red-500'
              : 'border border-gray-300 focus-visible:ring-emerald-500',
          ]"
        />
        <p v-if="fieldErrors.stock_quantity" class="mt-1 text-sm text-red-600">
          {{ fieldErrors.stock_quantity }}
        </p>
      </div>

      <div>
        <label for="product-price" class="block text-sm font-medium text-gray-700"
          >Price <span class="text-red-500">*</span></label
        >
        <input
          id="product-price"
          type="number"
          min="0"
          step="0.01"
          v-model="formData.price"
          @input="fieldErrors.price = ''"
          :class="[
            'mt-1 block w-full rounded-md px-3 py-2 text-sm focus:outline-none focus-visible:ring-2',
            fieldErrors.price
              ? 'border border-red-300 focus-visible:ring-red-500'
              : 'border border-gray-300 focus-visible:ring-emerald-500',
          ]"
        />
        <p v-if="fieldErrors.price" class="mt-1 text-sm text-red-600">{{ fieldErrors.price }}</p>
      </div>
    </div>

    <div class="flex items-center justify-end">
      <button
        type="submit"
        class="rounded-md bg-emerald-700 px-4 py-2 font-bold text-white transition-colors hover:bg-emerald-800 focus:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500"
      >
        Save Product
      </button>
    </div>
  </form>
</template>

<script setup>
import { onMounted, reactive, watch } from 'vue'
import { getProduct, createProduct, updateProduct } from '@/network/request'
import logger from '@/utils/logger'

const props = defineProps({
  productId: {
    type: String,
    required: false,
  },
})

const emit = defineEmits(['saved', 'error'])

const formData = reactive({
  name: '',
  description: '',
  stock_quantity: '',
  price: '',
})

const fieldErrors = reactive({
  name: '',
  description: '',
  stock_quantity: '',
  price: '',
})

async function fetchProductInformation() {
  if (props.productId) {
    logger.Debug('ProductForm.vue', 'fetchProductInformation', 'fetching product', {
      id: props.productId,
    })
    const data = await getProduct(props.productId)
    if (!data.ok) {
      logger.Warn('ProductForm.vue', 'fetchProductInformation', 'failed to load product', {
        error: String(data.error ?? ''),
      })
      emit('error', String(data.error ?? ''))
      return
    }
    formData.name = data.data.name ?? ''
    formData.description = data.data.description ?? ''
    formData.stock_quantity = data.data.stock_quantity ?? ''
    formData.price = data.data.price != null ? (data.data.price / 100).toFixed(2) : ''
    logger.Debug('ProductForm.vue', 'fetchProductInformation', 'product loaded', {
      name: formData.name,
    })
  }
}

const INVISIBLE_CHARS = new RegExp(
  '[\\u0000-\\u001F\\u007F-\\u009F\\u200B-\\u200F\\u2028\\u2029\\uFEFF]',
)
function hasInvisibleChars(value) {
  return INVISIBLE_CHARS.test(value)
}

function validateText(value, max, label) {
  const text = value?.trim() ?? ''
  if (!text) return `${label} is required.`
  if (text.length > max) return `${label} must be ${max} characters or fewer.`
  if (!/\p{L}/u.test(text)) return `${label} cannot be only numbers or symbols.`
  if (hasInvisibleChars(text)) return `${label} contains invisible or non-renderable characters.`
  return ''
}

function priceToCents(value) {
  const [dollarsPart, decimalPart = ''] = String(value).split('.')
  return Number(dollarsPart) * 100 + Number(decimalPart.slice(0, 2).padEnd(2, '0'))
}

function validateForm() {
  fieldErrors.name = validateText(formData.name, 50, 'Product name')
  fieldErrors.description = validateText(formData.description, 150, 'Description')

  const stock = Number(formData.stock_quantity)
  if (formData.stock_quantity === null || formData.stock_quantity === '') {
    fieldErrors.stock_quantity = 'Stock quantity is required.'
  } else if (!Number.isInteger(stock) || stock < 1) {
    fieldErrors.stock_quantity = 'Stock quantity must be at least 1.'
  } else if (stock > 2147483647) {
    fieldErrors.stock_quantity = 'Stock quantity cannot exceed 2147483647.'
  } else {
    fieldErrors.stock_quantity = ''
  }

  if (formData.price === null || formData.price === '') {
    fieldErrors.price = 'Price is required.'
  } else {
    const cents = priceToCents(formData.price)
    if (!Number.isFinite(cents) || cents <= 0) {
      fieldErrors.price = 'Price must be greater than 0.'
    } else if (cents > 999999999) {
      fieldErrors.price = 'Price cannot exceed $9,999,999.99.'
    } else {
      fieldErrors.price = ''
    }
  }

  return (
    !fieldErrors.name &&
    !fieldErrors.description &&
    !fieldErrors.stock_quantity &&
    !fieldErrors.price
  )
}

async function saveProduct() {
  if (!validateForm()) return
  const payload = { ...formData, price: priceToCents(formData.price) }
  emit('error', '')
  logger.Debug(
    'ProductForm.vue',
    'saveProduct',
    props.productId ? 'updating product' : 'creating product',
    { name: payload.name },
  )

  let response
  if (props.productId) {
    response = await updateProduct(props.productId, payload)
  } else {
    response = await createProduct(payload)
  }

  if (!response.ok) {
    logger.Warn('ProductForm.vue', 'saveProduct', 'failed to save product', {
      error: String(response.error ?? ''),
    })
    emit('error', String(response.error ?? ''))
    return
  }

  logger.Info('ProductForm.vue', 'saveProduct', 'product saved', { id: response.data?.id })
  emit('saved', response.data)
}

onMounted(fetchProductInformation)
watch(() => props.productId, fetchProductInformation)
</script>
