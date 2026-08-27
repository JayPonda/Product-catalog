<template>
  <div class="space-y-6">
    <!-- inline add: type a name and add without leaving the page (authenticated users only) -->
    <div class="flex items-end justify-end gap-2" v-if="auth.isAuthenticated">
      <div>
        <label for="new-category" class="block text-sm font-medium text-gray-700"
          >New category</label
        >
        <input
          id="new-category"
          type="text"
          v-model="newCategory"
          @keyup.enter="addCategory"
          placeholder="Category name"
          class="mt-1 block w-64 rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-emerald-500 focus:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500"
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

    <!-- error alert -->
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { getCategories, createCategory } from '@/network/request.js'
import { useAuthStore } from '@/stores/auth'
import { useErrorStore } from '@/stores/errors'
import BaseTable from '@/components/table/BaseTable.vue'
import logger from '@/utils/logger'

const auth = useAuthStore()
const error = useErrorStore()

const curruntPage = ref(0) // offset
const curruntLimit = 20
const categories = ref({})
const categoryTableTitle = ref([])
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

async function addCategory() {
  const name = newCategory.value.trim()
  if (!name) {
    error.show('Category name is required.')
    return
  }
  addingCategory.value = true
  const response = await createCategory(name)
  addingCategory.value = false
  if (response.ok) {
    newCategory.value = ''
    await fetchCategories()
  }
  error.show(String(response.error ?? ''))
}

async function fetchCategories() {
  const response = await getCategories(curruntPage.value, curruntLimit)
  logger.Debug('Category/Index.vue', 'fetchCategories', 'fetch response', { ok: response.ok, data: response.data })
  if (response.ok) {
    if (response.data?.categories?.length > 0) {
      const keys = Object.keys(response.data?.categories[0])
      logger.Debug('Category/Index.vue', 'fetchCategories', 'column keys', { keys })
      categoryTableTitle.value = keys
    }

    categories.value = response.data
  }
  error.show(String(response.error ?? ''))
}

function next() {
  if (!categories.value) {
    error.show('Something went wrong')
    return
  }
  if (
    categories.value.total !== undefined &&
    (curruntPage.value + 1) * curruntLimit < categories.value.total
  ) {
    curruntPage.value = curruntPage.value + 1
    fetchCategories()
  } else {
    error.show('No records on next page.')
  }
}

function previous() {
  if (!categories.value) {
    error.show('Something went wrong')
    return
  }
  if (curruntPage.value > 0) {
    curruntPage.value = curruntPage.value - 1
    fetchCategories()
  } else {
    error.show('No records on previous page.')
  }
}

onMounted(fetchCategories)
</script>
