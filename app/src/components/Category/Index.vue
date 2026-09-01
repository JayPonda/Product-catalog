<template>
  <div class="space-y-6">
    <!-- Toolbar: Search by category name, reset, and inline add category (for auth users) -->
    <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
      <div class="flex flex-1 items-center gap-3">
        <!-- Search by category name -->
        <div class="relative w-full sm:w-64">
          <Search
            class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400"
          />
          <input
            v-model="searchQuery"
            type="search"
            placeholder="Search categories..."
            aria-label="Search categories by name"
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

        <!-- Reset filter button -->
        <button
          v-if="searchQuery"
          @click="clearSearch"
          type="button"
          class="text-xs font-medium text-gray-500 underline hover:text-gray-800"
        >
          Reset
        </button>
      </div>

      <!-- inline add: type a name and add without leaving the page (authenticated users only) -->
      <div class="flex items-center gap-2" v-if="auth.isAuthenticated">
        <div>
          <label for="new-category" class="sr-only">New category</label>
          <input
            id="new-category"
            type="text"
            v-model="newCategory"
            @keyup.enter="addCategory"
            placeholder="Category name"
            class="block w-64 rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-emerald-500 focus:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500"
          />
        </div>
        <button
          type="button"
          @click="addCategory"
          :disabled="addingCategory"
          class="rounded-md bg-emerald-700 px-4 py-2 font-bold text-white transition-colors hover:bg-emerald-800 focus:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500 disabled:opacity-50"
        >
          Add
        </button>
      </div>
    </div>

    <!-- table showing the results -->
    <BaseTable
      v-if="categories && Object.keys(categories).length >= 0"
      :columns="columns"
      :items="categories.categories || []"
    />

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
import { computed, onMounted, onBeforeUnmount, ref, watch } from 'vue'
import { getCategories, createCategory } from '@/network/request.js'
import { useAuthStore } from '@/stores/auth'
import { useNotificationStore } from '@/stores/notifications'
import { Search, X } from '@lucide/vue'
import BaseTable from '@/components/table/BaseTable.vue'
import logger from '@/utils/logger'

const auth = useAuthStore()
const notifications = useNotificationStore()

const curruntPage = ref(0) // offset
const curruntLimit = 20
const categories = ref({})
const categoryTableTitle = ref(['id', 'name'])
const scalarCategoryTitles = computed(() =>
  categoryTableTitle.value.filter((t) => t !== 'deleted_at'),
)
const columns = computed(() =>
  scalarCategoryTitles.value.map((title) => ({
    key: title,
    label: title.replace(/_/g, ' ').replace(/^\w/, (c) => c.toUpperCase()),
  })),
)
const newCategory = ref('')
const addingCategory = ref(false)

// Search state
const searchQuery = ref('')
let searchDebounceTimer = undefined

function clearSearch() {
  searchQuery.value = ''
  clearTimeout(searchDebounceTimer)
  curruntPage.value = 0
  fetchCategories()
}

watch(searchQuery, () => {
  clearTimeout(searchDebounceTimer)
  searchDebounceTimer = window.setTimeout(() => {
    curruntPage.value = 0
    fetchCategories()
  }, 300)
})

async function addCategory() {
  const name = newCategory.value.trim()
  if (!name) {
    notifications.show('Category name is required.')
    return
  }
  addingCategory.value = true
  const response = await createCategory(name)
  addingCategory.value = false
  if (response.ok) {
    newCategory.value = ''
    await fetchCategories()
    notifications.success('Product category added successfully.')
  } else {
    notifications.show(response.message || response.error)
  }
}

async function fetchCategories() {
  const filterOptions = {}
  if (searchQuery.value.trim()) {
    filterOptions.name = searchQuery.value.trim()
  }

  const response =
    Object.keys(filterOptions).length > 0
      ? await getCategories(curruntPage.value, curruntLimit, filterOptions)
      : await getCategories(curruntPage.value, curruntLimit)

  logger.Debug('Category/Index.vue', 'fetchCategories', 'fetch response', {
    ok: response.ok,
    data: response.data,
  })
  if (response.ok) {
    if (response.data?.categories?.length > 0) {
      const keys = Object.keys(response.data.categories[0])
      logger.Debug('Category/Index.vue', 'fetchCategories', 'column keys', { keys })
      categoryTableTitle.value = keys
    }

    categories.value = response.data
  } else {
    notifications.show(response.message || response.error)
  }
}

function next() {
  if (!categories.value) {
    notifications.show('Something went wrong')
    return
  }
  if (
    categories.value.total !== undefined &&
    (curruntPage.value + 1) * curruntLimit < categories.value.total
  ) {
    curruntPage.value = curruntPage.value + 1
    fetchCategories()
  } else {
    notifications.show('No records on next page.')
  }
}

function previous() {
  if (!categories.value) {
    notifications.show('Something went wrong')
    return
  }
  if (curruntPage.value > 0) {
    curruntPage.value = curruntPage.value - 1
    fetchCategories()
  } else {
    notifications.show('No records on previous page.')
  }
}

onMounted(fetchCategories)

onBeforeUnmount(() => {
  clearTimeout(searchDebounceTimer)
})
</script>
