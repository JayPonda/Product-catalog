<template>
  <div class="space-y-6">
    <!-- section 2: already linked categories as tags -->
    <div id="category" class="rounded-lg border border-gray-200">
      <div class="border-b border-gray-200 p-4">
        <h2 class="text-sm font-semibold text-gray-700">Linked categories</h2>
      </div>
      <div class="p-4">
        <div v-if="linkedCategories.length === 0" class="text-sm text-gray-400">
          No categories linked yet.
        </div>
        <div v-else class="flex flex-wrap gap-2">
          <span
            v-for="category in visibleLinkedCategories"
            :key="category.id"
            class="inline-flex items-center gap-1 rounded-full bg-emerald-50 px-3 py-1 text-sm font-medium text-emerald-700"
          >
            {{ category.name }}
            <button
              type="button"
              @click="unlinkProductCategory(category)"
              :disabled="busyCategoryId === category.id"
              class="rounded-full text-emerald-500 transition-colors hover:bg-emerald-100 hover:text-emerald-700 focus:outline-none disabled:opacity-50"
              aria-label="Unlink category"
            >
              &#10005;
            </button>
          </span>
        </div>
        <button
          v-if="linkedCategories.length > linkedVisibleCount"
          type="button"
          @click="linkedVisibleCount += 20"
          class="mt-3 rounded-md border border-gray-300 px-3 py-1.5 text-sm font-medium text-gray-700 transition-colors hover:bg-gray-50 focus:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500"
        >
          Load more
        </button>
      </div>
    </div>

    <!-- section 3: search + link, pending categories as tags -->
    <div class="rounded-lg border border-gray-200">
      <div class="border-b border-gray-200 p-4">
        <h2 class="text-sm font-semibold text-gray-700">Add categories</h2>
      </div>
      <div class="space-y-4 p-4">
        <div class="flex items-center gap-2">
          <div class="relative">
            <input
              type="search"
              v-model="categorySearch"
              placeholder="Search category..."
              class="w-48 rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-emerald-500 focus:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500"
            />
            <!-- dropdown with matching categories -->
            <ul
              v-if="categoryDropdownOpen"
              class="absolute right-0 z-10 mt-1 max-h-60 w-56 overflow-y-auto rounded-md border border-gray-200 bg-white py-1 shadow-lg"
            >
              <li
                v-if="visibleCategoryResults.length === 0"
                class="px-3 py-2 text-sm text-gray-400"
              >
                No matching categories
              </li>
              <li v-else v-for="category in visibleCategoryResults" :key="category.id">
                <button
                  type="button"
                  @mousedown.prevent="pickCategory(category)"
                  class="block w-full px-3 py-2 text-left text-sm hover:bg-gray-50"
                >
                  {{ category.name }}
                </button>
              </li>
            </ul>
          </div>
          <button
            type="button"
            @click="addCategories"
            :disabled="addingCategories"
            class="rounded-md bg-emerald-700 px-4 py-2 font-bold text-white transition-colors hover:bg-emerald-800 focus:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500 disabled:opacity-50"
          >
            Link category
          </button>
        </div>

        <div v-if="pendingCategories.length === 0" class="text-sm text-gray-400">
          Search and select categories to link.
        </div>
        <div v-else class="flex flex-wrap gap-2">
          <span
            v-for="category in pendingCategories"
            :key="category.id"
            class="inline-flex items-center gap-1 rounded-full bg-gray-100 px-3 py-1 text-sm text-gray-700"
          >
            {{ category.name }}
            <button
              type="button"
              @click="removePending(category)"
              class="rounded-full text-gray-500 transition-colors hover:bg-gray-200 hover:text-gray-700 focus:outline-none"
              aria-label="Remove from selection"
            >
              &#10005;
            </button>
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { getProduct, searchCategory, linkCategory, unlinkCategory } from '@/network/request'
import { useNotificationStore } from '@/stores/notifications'
import logger from '@/utils/logger'

const props = defineProps({
  productId: {
    type: String,
    required: true,
  },
})

const emit = defineEmits(['error'])
const notifications = useNotificationStore()

const categorySearch = ref('')
const categoryResults = ref([])
const categoryDropdownOpen = ref(false)
let categorySearchTimer = undefined

const productCategories = ref([])
const savedCategoryIds = ref(new Set())

const linkedCategories = computed(() =>
  productCategories.value.filter((c) => savedCategoryIds.value.has(c.id)),
)
const pendingCategories = computed(() =>
  productCategories.value.filter((c) => !savedCategoryIds.value.has(c.id)),
)

const linkedVisibleCount = ref(20)
const visibleLinkedCategories = computed(() =>
  linkedCategories.value.slice(0, linkedVisibleCount.value),
)

const visibleCategoryResults = computed(() => {
  return categoryResults.value.filter(
    (category) => !productCategories.value.some((item) => item.id === category.id),
  )
})

async function fetchLinkedCategories() {
  if (props.productId) {
    emit('error', '')
    logger.Debug('CategoryLinker.vue', 'fetchLinkedCategories', 'fetching linked categories', {
      productId: props.productId,
    })
    const response = await getProduct(props.productId)
    if (response.ok) {
      productCategories.value = response.data.categories ?? []
      savedCategoryIds.value = new Set(productCategories.value.map((c) => c.id))
      logger.Debug('CategoryLinker.vue', 'fetchLinkedCategories', 'linked categories loaded', {
        count: productCategories.value.length,
      })
    } else {
      logger.Warn(
        'CategoryLinker.vue',
        'fetchLinkedCategories',
        'failed to load linked categories',
        { error: String(response.error ?? '') },
      )
      emit('error', String(response.error ?? ''))
    }
  }
}

function removePending(category) {
  productCategories.value = productCategories.value.filter((item) => item.id !== category.id)
}

const busyCategoryId = ref(null)

async function unlinkProductCategory(category) {
  if (!props.productId) return
  emit('error', '')
  busyCategoryId.value = category.id
  const data = await unlinkCategory(props.productId, category.id)
  busyCategoryId.value = null
  if (data.ok) {
    logger.Debug('CategoryLinker.vue', 'unlinkProductCategory', 'category unlinked', {
      categoryId: category.id,
    })
    savedCategoryIds.value.delete(category.id)
    productCategories.value = productCategories.value.filter((item) => item.id !== category.id)
    notifications.success(`Category "${category.name}" unlinked.`)
  } else {
    logger.Warn('CategoryLinker.vue', 'unlinkProductCategory', 'failed to unlink category', {
      categoryId: category.id,
      error: String(data.error ?? ''),
    })
    emit('error', String(data.error ?? ''))
  }
}

watch(categorySearch, (value) => {
  clearTimeout(categorySearchTimer)
  const query = value.trim()
  if (!query) {
    categoryResults.value = []
    categoryDropdownOpen.value = false
    return
  }
  categorySearchTimer = window.setTimeout(async () => {
    const data = await searchCategory(query)
    categoryResults.value = data.ok ? data.data.categories : []
    categoryDropdownOpen.value = true
    logger.Debug('CategoryLinker.vue', 'searchCategory', 'category search completed', {
      query,
      results: categoryResults.value.length,
    })
  }, 250)
})

function pickCategory(category) {
  if (!productCategories.value.some((item) => item.id === category.id)) {
    productCategories.value.push({ ...category, addedAt: new Date().toLocaleString() })
  }
  categorySearch.value = ''
  categoryResults.value = []
  categoryDropdownOpen.value = false
}

const addingCategories = ref(false)

async function addCategories() {
  if (!props.productId) return
  const pending = pendingCategories.value
  if (pending.length === 0) return
  emit('error', '')
  addingCategories.value = true
  const failures = []
  for (const category of pending) {
    const data = await linkCategory(props.productId, category.id)
    if (data.ok) {
      savedCategoryIds.value.add(category.id)
      logger.Debug('CategoryLinker.vue', 'addCategories', 'category linked', {
        categoryId: category.id,
      })
    } else {
      failures.push(category.name)
      logger.Warn('CategoryLinker.vue', 'addCategories', 'failed to link category', {
        categoryId: category.id,
        error: String(data.error ?? ''),
      })
    }
  }
  addingCategories.value = false
  if (failures.length) {
    emit('error', `Could not add: ${failures.join(', ')}`)
  } else {
    notifications.success(
      pending.length === 1 ? 'Category linked successfully.' : 'Categories linked successfully.',
    )
  }
}

onMounted(fetchLinkedCategories)
watch(() => props.productId, fetchLinkedCategories)
</script>
