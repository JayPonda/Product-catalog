<template>
  <div class="space-y-6">
    <!-- button for create a new product (visible on all product pages) -->
    <div class="flex items-center justify-end">
      <button
        @click="addProduct"
        class="rounded-md bg-emerald-700 px-4 py-2 font-bold text-white transition-colors hover:bg-emerald-800 focus:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500"
      >
        New Product
      </button>
    </div>

    <!-- table showing the results -->
    <BaseTable
      v-if="products && Object.keys(products).length >= 0"
      :columns="columns"
      :items="products.products || []"
      class="min-h-[220px]"
    >
      <template #cell-categories="{ item }">
        <div class="flex flex-wrap gap-1">
          <span
            v-for="category in item.categories"
            :key="category.id"
            class="inline-flex items-center rounded-full bg-emerald-50 px-3 py-1 text-sm font-medium text-emerald-700"
          >
            {{ category.name }}
          </span>
          <span
            v-if="!item.categories || item.categories.length === 0"
            class="text-sm text-gray-400"
            >&mdash;</span
          >
        </div>
      </template>
      <template #cell-actions="{ item }">
        <div class="relative text-right">
          <button
            @click="toggleMenu(item.id)"
            aria-label="Product actions"
            class="rounded-md p-1.5 text-gray-500 transition-colors hover:bg-gray-100 hover:text-gray-700 focus:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500"
          >
            <EllipsisVertical class="h-5 w-5" />
          </button>

          <!-- click-away layer -->
          <div
            v-if="openMenuId === item.id"
            class="fixed inset-0 z-60"
            @click="openMenuId = null"
          ></div>

          <!-- dropdown menu -->
          <div
            v-if="openMenuId === item.id"
            class="absolute right-4 z-60 mt-1 w-44 overflow-hidden rounded-md border border-gray-200 bg-white py-1 shadow-lg text-left"
          >
            <button
              @click="addCategories(item)"
              class="block w-full px-4 py-2 text-left text-sm text-gray-700 hover:bg-gray-50"
            >
              Add Categories
            </button>
            <button
              @click="editProduct(item)"
              class="block w-full px-4 py-2 text-left text-sm text-gray-700 hover:bg-gray-50"
            >
              Edit Product
            </button>
            <button
              @click="removeProduct(item)"
              class="block w-full px-4 py-2 text-left text-sm text-red-600 hover:bg-red-50"
            >
              Delete Product
            </button>
          </div>
        </div>
      </template>
    </BaseTable>

    <!-- pagination component -->
    <nav class="flex items-center justify-end gap-2" aria-label="Pagination">
      <button
        @click="previous"
        class="inline-flex items-center rounded-md border border-gray-300 bg-white px-3 py-2 text-sm font-medium leading-5 text-gray-700 shadow-xs transition-colors hover:bg-gray-50 hover:text-gray-900 focus:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500"
      >
        Previous
      </button>
      <span class="px-2 text-sm font-medium text-gray-700">Page {{ curruntPage + 1 }}</span>
      <button
        @click="next"
        class="inline-flex items-center rounded-md border border-gray-300 bg-white px-3 py-2 text-sm font-medium leading-5 text-gray-700 shadow-xs transition-colors hover:bg-gray-50 hover:text-gray-900 focus:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500"
      >
        Next
      </button>
    </nav>
  </div>
</template>

<script setup>
import { computed, ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { getProducts, getMyProducts, deleteProduct } from '@/network/request.js'
import { useNotificationStore } from '@/stores/notifications'
import { EllipsisVertical } from '@lucide/vue'
import BaseTable from '@/components/table/BaseTable.vue'
import logger from '@/utils/logger'

const props = defineProps({
  showControls: { type: Boolean, default: false },
  myProducts: { type: Boolean, default: false },
})

const router = useRouter()
const notifications = useNotificationStore()

// Scalar columns to render as plain cells; categories and ownership get their own columns.
const scalarTitles = computed(() =>
  productTableTitle.value.filter(
    (t) => t !== 'categories' && t !== 'deleted_at' && t !== 'user_id',
  ),
)

const columns = computed(() => {
  const cols = scalarTitles.value.map((title) => ({
    key: title,
    label: title.replace(/_/g, ' ').replace(/^\w/, (c) => c.toUpperCase()),
  }))
  cols.push({ key: 'categories', label: 'Categories' })
  if (props.showControls) {
    cols.push({ key: 'actions', label: '' })
  }
  return cols
})

const curruntPage = ref(0) // offset
const curruntLimit = 20
const products = ref({})
const productTableTitle = ref([])
const openMenuId = ref(null)

function toggleMenu(id) {
  openMenuId.value = openMenuId.value === id ? null : id
}

function addProduct() {
  router.push({
    name: 'products-create',
  })
}

function addCategories(product) {
  router.push({
    name: 'products-modify',
    params: { id: product.id, action: 'edit' },
    hash: '#category',
  })
}

function editProduct(product) {
  router.push({ name: 'products-modify', params: { id: product.id, action: 'edit' } })
}

async function removeProduct(product) {
  logger.Debug('Product/Index.vue', 'removeProduct', 'deleting product', { id: product.id })
  openMenuId.value = null

  const response = await deleteProduct(product.id)
  logger.Debug('Product/Index.vue', 'removeProduct', 'delete response', {
    ok: response.ok,
    data: response.data,
  })
  if (response.ok) {
    fetchProducts()
    notifications.success('Product deleted successfully.')
  } else {
    notifications.show(response.message || response.error)
  }
}

async function fetchProducts() {
  const fetchFn = props.myProducts ? getMyProducts : getProducts
  const response = await fetchFn(curruntPage.value, curruntLimit)
  logger.Debug('Product/Index.vue', 'fetchProducts', 'fetch response', {
    ok: response.ok,
    data: response.data,
  })
  if (response.ok) {
    if (response.data?.products.length > 0) {
      const keys = Object.keys(response.data?.products[0])
      logger.Debug('Product/Index.vue', 'fetchProducts', 'column keys', { keys })
      productTableTitle.value = keys
    }

    products.value = response.data
  } else {
    notifications.show(response.message || response.error)
  }
}

function next() {
  if (!products.value) {
    notifications.show('Something went wrong')
    return
  }
  const product = products.value

  if (
    product.total !== undefined &&
    product.limit !== undefined &&
    product.offset !== undefined &&
    product.total - product.limit * (product.offset + 1) > 0
  ) {
    curruntPage.value = curruntPage.value + 1
    fetchProducts()
  } else {
    notifications.show('No records on next page.')
  }
}

function previous() {
  if (!products.value) {
    notifications.show('Something went wrong')
    return
  }
  const product = products.value

  if (product.offset !== undefined && product.offset > 0) {
    curruntPage.value = curruntPage.value - 1
    fetchProducts()
  } else {
    notifications.show('No records on previous page.')
  }
}

onMounted(fetchProducts)
</script>
