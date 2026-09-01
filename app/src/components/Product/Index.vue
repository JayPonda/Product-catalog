<template>
  <div class="space-y-6">
    <!-- toolbar: search by name, multi-category filter, reset, and new product button -->
    <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
      <div class="flex flex-1 flex-wrap items-center gap-3">
        <!-- search by name -->
        <div class="relative w-full sm:w-64">
          <Search
            class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400"
          />
          <input
            v-model="searchQuery"
            type="search"
            placeholder="Search by name..."
            aria-label="Search products by name"
            class="w-full rounded-md border border-gray-300 py-2 pl-9 pr-8 text-sm focus:border-emerald-500 focus:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500"
          />
          <button
            v-if="searchQuery"
            @click="clearSearch"
            type="button"
            class="absolute right-2.5 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600 focus:outline-none"
            aria-label="Clear search"
          >
            <X class="h-4 w-4" />
          </button>
        </div>

        <!-- multi-category filter dropdown -->
        <div class="relative" ref="categoryDropdownRef">
          <button
            @click="toggleCategoryDropdown"
            type="button"
            aria-label="Filter by categories"
            class="inline-flex items-center gap-2 rounded-md border border-gray-300 bg-white px-3 py-2 text-sm font-medium text-gray-700 shadow-xs hover:bg-gray-50 focus:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500"
          >
            <Filter class="h-4 w-4 text-gray-500" />
            <span>Categories</span>
            <span
              v-if="selectedCategories.length > 0"
              class="rounded-full bg-emerald-100 px-2 py-0.5 text-xs font-semibold text-emerald-800"
            >
              {{ selectedCategories.length }}
            </span>
            <ChevronDown class="h-4 w-4 text-gray-400" />
          </button>

          <!-- dropdown menu -->
          <div
            v-if="categoryDropdownOpen"
            class="absolute left-0 z-50 mt-1 w-64 rounded-md border border-gray-200 bg-white p-2 shadow-lg"
          >
            <div class="mb-2">
              <input
                v-model="categoryFilterSearch"
                type="text"
                placeholder="Filter categories..."
                class="w-full rounded border border-gray-300 px-2 py-1 text-xs focus:border-emerald-500 focus:outline-none"
              />
            </div>

            <div class="max-h-48 overflow-y-auto space-y-1">
              <div
                v-if="filteredCategories.length === 0"
                class="px-2 py-2 text-center text-xs text-gray-400"
              >
                No categories found
              </div>
              <label
                v-for="category in filteredCategories"
                :key="category.id"
                class="flex cursor-pointer items-center gap-2 rounded px-2 py-1.5 text-sm hover:bg-gray-50"
              >
                <input
                  type="checkbox"
                  :value="category.id"
                  :checked="isCategorySelected(category.id)"
                  @change="toggleCategorySelection(category)"
                  class="rounded border-gray-300 text-emerald-600 focus:ring-emerald-500"
                />
                <span class="truncate text-gray-700">{{ category.name }}</span>
              </label>
            </div>

            <div
              v-if="selectedCategories.length > 0"
              class="mt-2 flex items-center justify-between border-t border-gray-100 pt-2"
            >
              <button
                @click="clearSelectedCategories"
                type="button"
                class="text-xs font-medium text-red-600 hover:text-red-700"
              >
                Clear
              </button>
              <span class="text-xs text-gray-400">{{ selectedCategories.length }} selected</span>
            </div>
          </div>
        </div>

        <!-- reset all filters button -->
        <button
          v-if="searchQuery || selectedCategories.length > 0"
          @click="clearAllFilters"
          type="button"
          class="text-xs font-medium text-gray-500 underline hover:text-gray-800"
        >
          Reset filters
        </button>
      </div>

      <!-- button for create a new product (visible on all product pages) -->
      <div class="flex items-center justify-end">
        <button
          @click="addProduct"
          class="rounded-md bg-emerald-700 px-4 py-2 font-bold text-white transition-colors hover:bg-emerald-800 focus:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500"
        >
          New Product
        </button>
      </div>
    </div>

    <!-- active category chips -->
    <div v-if="selectedCategories.length > 0" class="-mt-2 flex flex-wrap items-center gap-1.5">
      <span class="mr-1 text-xs text-gray-500">Categories:</span>
      <span
        v-for="cat in selectedCategories"
        :key="cat.id"
        class="inline-flex items-center gap-1 rounded-full bg-emerald-50 px-2.5 py-1 text-xs font-medium text-emerald-700"
      >
        {{ cat.name }}
        <button
          type="button"
          @click="removeCategory(cat.id)"
          class="rounded-full text-emerald-500 hover:bg-emerald-100 hover:text-emerald-700 focus:outline-none"
          :aria-label="`Remove ${cat.name}`"
        >
          <X class="h-3 w-3" />
        </button>
      </span>
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
import { computed, ref, onMounted, onBeforeUnmount, watch } from 'vue'
import { useRouter } from 'vue-router'
import {
  getProducts,
  getMyProducts,
  deleteProduct,
  getCategories,
  searchCategory,
} from '@/network/request.js'
import { useNotificationStore } from '@/stores/notifications'
import { EllipsisVertical, Search, X, Filter, ChevronDown } from '@lucide/vue'
import BaseTable from '@/components/table/BaseTable.vue'
import logger from '@/utils/logger'

const props = defineProps({
  showControls: { type: Boolean, default: false },
  myProducts: { type: Boolean, default: false },
})

const router = useRouter()
const notifications = useNotificationStore()

// Search & filter state
const searchQuery = ref('')
const selectedCategories = ref([])
const categoryDropdownOpen = ref(false)
const categoryFilterSearch = ref('')
const availableCategories = ref([])
const categoryDropdownRef = ref(null)
let searchDebounceTimer = undefined
let categorySearchDebounceTimer = undefined

const filteredCategories = computed(() => {
  const query = categoryFilterSearch.value.trim().toLowerCase()
  if (!query) {
    return availableCategories.value
  }
  return availableCategories.value.filter((cat) => cat.name.toLowerCase().includes(query))
})

function isCategorySelected(categoryId) {
  return selectedCategories.value.some((c) => c.id === categoryId)
}

function toggleCategorySelection(category) {
  const idx = selectedCategories.value.findIndex((c) => c.id === category.id)
  if (idx >= 0) {
    selectedCategories.value.splice(idx, 1)
  } else {
    selectedCategories.value.push(category)
  }
  curruntPage.value = 0
  fetchProducts()
}

function removeCategory(categoryId) {
  selectedCategories.value = selectedCategories.value.filter((c) => c.id !== categoryId)
  curruntPage.value = 0
  fetchProducts()
}

function clearSearch() {
  searchQuery.value = ''
  clearTimeout(searchDebounceTimer)
  curruntPage.value = 0
  fetchProducts()
}

function clearSelectedCategories() {
  selectedCategories.value = []
  curruntPage.value = 0
  fetchProducts()
}

function clearAllFilters() {
  searchQuery.value = ''
  clearTimeout(searchDebounceTimer)
  selectedCategories.value = []
  curruntPage.value = 0
  fetchProducts()
}

async function loadAvailableCategories() {
  if (availableCategories.value.length > 0) return
  const res = await getCategories(0, 100)
  if (res.ok && res.data?.categories) {
    availableCategories.value = res.data.categories
  }
}

function toggleCategoryDropdown() {
  categoryDropdownOpen.value = !categoryDropdownOpen.value
  if (categoryDropdownOpen.value) {
    loadAvailableCategories()
  }
}

function handleClickOutside(e) {
  if (categoryDropdownRef.value && !categoryDropdownRef.value.contains(e.target)) {
    categoryDropdownOpen.value = false
  }
}

watch(searchQuery, () => {
  clearTimeout(searchDebounceTimer)
  searchDebounceTimer = window.setTimeout(() => {
    curruntPage.value = 0
    fetchProducts()
  }, 300)
})

watch(categoryFilterSearch, (val) => {
  const term = val.trim()
  if (!term) return
  clearTimeout(categorySearchDebounceTimer)
  categorySearchDebounceTimer = window.setTimeout(async () => {
    const res = await searchCategory(term)
    if (res.ok && res.data?.categories) {
      const existingMap = new Map(availableCategories.value.map((c) => [c.id, c]))
      for (const cat of res.data.categories) {
        if (!existingMap.has(cat.id)) {
          availableCategories.value.push(cat)
        }
      }
    }
  }, 250)
})

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
const productTableTitle = ref(['id', 'name', 'description', 'price', 'stock_quantity'])
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
  const filterOptions = {}
  if (searchQuery.value.trim()) {
    filterOptions.name = searchQuery.value.trim()
  }
  if (selectedCategories.value.length > 0) {
    filterOptions.categoryIds = selectedCategories.value.map((c) => c.id)
  }

  const response =
    Object.keys(filterOptions).length > 0
      ? await fetchFn(curruntPage.value, curruntLimit, filterOptions)
      : await fetchFn(curruntPage.value, curruntLimit)

  logger.Debug('Product/Index.vue', 'fetchProducts', 'fetch response', {
    ok: response.ok,
    data: response.data,
  })
  if (response.ok) {
    if (response.data?.products?.length > 0) {
      const keys = Object.keys(response.data.products[0])
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
  const total = products.value.total ?? 0

  if ((curruntPage.value + 1) * curruntLimit < total) {
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
  if (curruntPage.value > 0) {
    curruntPage.value = curruntPage.value - 1
    fetchProducts()
  } else {
    notifications.show('No records on previous page.')
  }
}

onMounted(() => {
  fetchProducts()
  loadAvailableCategories()
  if (typeof document !== 'undefined') {
    document.addEventListener('click', handleClickOutside)
  }
})

onBeforeUnmount(() => {
  clearTimeout(searchDebounceTimer)
  clearTimeout(categorySearchDebounceTimer)
  if (typeof document !== 'undefined') {
    document.removeEventListener('click', handleClickOutside)
  }
})
</script>
