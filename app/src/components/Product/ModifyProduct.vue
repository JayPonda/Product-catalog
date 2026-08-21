<template>
    <div class="space-y-6">

        <!-- error alert -->
        <div v-if="error" role="alert"
            class="flex items-center justify-between rounded-md border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
            <span>{{ error }}</span>
            <button @click="error = ''" aria-label="Dismiss"
                class="ml-4 rounded p-1 leading-none text-red-500 transition-colors hover:bg-red-100 hover:text-red-700 focus:outline-none">
                &#10005;
            </button>
        </div>

        <!-- product details form -->
        <form class="space-y-4 rounded-lg border border-gray-200 bg-white p-6" @submit.prevent="saveProduct">
            <div>
                <label for="product-name" class="block text-sm font-medium text-gray-700">Product name <span
                        class="text-red-500">*</span></label>
                <input id="product-name" type="text" v-model="formData.name" @input="fieldErrors.name = ''"
                    :class="['mt-1 block w-full rounded-md px-3 py-2 text-sm focus:outline-none focus-visible:ring-2', fieldErrors.name ? 'border border-red-300 focus-visible:ring-red-500' : 'border border-gray-300 focus-visible:ring-emerald-500']" />
                <p v-if="fieldErrors.name" class="mt-1 text-sm text-red-600">{{ fieldErrors.name }}</p>
            </div>

            <div>
                <label for="product-description" class="block text-sm font-medium text-gray-700">Description <span
                        class="text-red-500">*</span></label>
                <textarea id="product-description" v-model="formData.description" rows="3"
                    class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-emerald-500 focus:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500"></textarea>
                <p v-if="fieldErrors.description" class="mt-1 text-sm text-red-600">{{ fieldErrors.description }}</p>
            </div>

            <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
                <div>
                    <label for="product-stock" class="block text-sm font-medium text-gray-700">Stock quantity <span
                            class="text-red-500">*</span></label>
                    <input id="product-stock" type="number" min="0" v-model="formData.stock_quantity"
                        @input="fieldErrors.stock_quantity = ''"
                        :class="['mt-1 block w-full rounded-md px-3 py-2 text-sm focus:outline-none focus-visible:ring-2', fieldErrors.stock_quantity ? 'border border-red-300 focus-visible:ring-red-500' : 'border border-gray-300 focus-visible:ring-emerald-500']" />
                    <p v-if="fieldErrors.stock_quantity" class="mt-1 text-sm text-red-600">{{ fieldErrors.stock_quantity
                        }}</p>
                </div>

                <div>
                    <label for="product-price" class="block text-sm font-medium text-gray-700">Price <span
                            class="text-red-500">*</span></label>
                    <input id="product-price" type="number" min="0" step="0.01" v-model="formData.price"
                        @input="fieldErrors.price = ''"
                        :class="['mt-1 block w-full rounded-md px-3 py-2 text-sm focus:outline-none focus-visible:ring-2', fieldErrors.price ? 'border border-red-300 focus-visible:ring-red-500' : 'border border-gray-300 focus-visible:ring-emerald-500']" />
                    <p v-if="fieldErrors.price" class="mt-1 text-sm text-red-600">{{ fieldErrors.price }}</p>
                </div>
            </div>

            <div class="flex items-center justify-end">
                <button type="submit"
                    class="rounded-md bg-emerald-700 px-4 py-2 font-bold text-white transition-colors hover:bg-emerald-800 focus:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500">
                    Save Product
                </button>
            </div>
        </form>

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
                    <span v-for="category in visibleLinkedCategories" :key="category.id"
                        class="inline-flex items-center gap-1 rounded-full bg-emerald-50 px-3 py-1 text-sm font-medium text-emerald-700">
                        {{ category.name }}
                        <button type="button" @click="unlinkProductCategory(category)"
                            :disabled="busyCategoryId === category.id"
                            class="rounded-full text-emerald-500 transition-colors hover:bg-emerald-100 hover:text-emerald-700 focus:outline-none disabled:opacity-50"
                            aria-label="Unlink category">&#10005;</button>
                    </span>
                </div>
                <button v-if="linkedCategories.length > linkedVisibleCount" type="button"
                    @click="linkedVisibleCount += 20"
                    class="mt-3 rounded-md border border-gray-300 px-3 py-1.5 text-sm font-medium text-gray-700 transition-colors hover:bg-gray-50 focus:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500">
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
                        <input type="search" v-model="categorySearch"
                            class="w-48 rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-emerald-500 focus:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500" />
                        <!-- dropdown with matching categories -->
                        <ul v-if="categoryDropdownOpen"
                            class="absolute right-0 z-10 mt-1 max-h-60 w-56 overflow-y-auto rounded-md border border-gray-200 bg-white py-1 shadow-lg">
                            <li v-if="visibleCategoryResults.length === 0" class="px-3 py-2 text-sm text-gray-400">
                                No matching categories
                            </li>
                            <li v-else v-for="category in visibleCategoryResults" :key="category.id">
                                <button type="button" @mousedown.prevent="pickCategory(category)"
                                    class="block w-full px-3 py-2 text-left text-sm hover:bg-gray-50">
                                    {{ category.name }}
                                </button>
                            </li>
                        </ul>
                    </div>
                    <button type="button" @click="addCategories" :disabled="addingCategories"
                        class="rounded-md bg-emerald-700 px-4 py-2 font-bold text-white transition-colors hover:bg-emerald-800 focus:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500 disabled:opacity-50">
                        Link category
                    </button>
                </div>

                <div v-if="pendingCategories.length === 0" class="text-sm text-gray-400">
                    Search and select categories to link.
                </div>
                <div v-else class="flex flex-wrap gap-2">
                    <span v-for="category in pendingCategories" :key="category.id"
                        class="inline-flex items-center gap-1 rounded-full bg-gray-100 px-3 py-1 text-sm text-gray-700">
                        {{ category.name }}
                        <button type="button" @click="removePending(category)"
                            class="rounded-full text-gray-500 transition-colors hover:bg-gray-200 hover:text-gray-700 focus:outline-none"
                            aria-label="Remove from selection">&#10005;</button>
                    </span>
                </div>
            </div>
        </div>

    </div>
</template>


<script setup>
import { computed, onMounted, reactive, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router'
import {
    getProduct,
    createProduct,
    updateProduct,
    searchCategory,
    linkCategory as linkCategoryRequest,
    unlinkCategory as unlinkCategoryRequest
} from '@/network/request'
const route = useRoute()
const router = useRouter()

const error = ref('')
const previousData = ref({})
const formData = reactive({})
const fieldErrors = reactive({
    name: '',
    description: '',
    stock_quantity: '',
    price: ''
})
const categorySearch = ref('')
const categoryResults = ref([])
const categoryDropdownOpen = ref(false)
let categorySearchTimer = null
const productCategories = ref([])
// ids of categories already persisted for this product (everything else is pending)
const savedCategoryIds = ref(new Set())

// split the combined list into linked vs pending for the two sections
const linkedCategories = computed(() => productCategories.value.filter((c) => savedCategoryIds.value.has(c.id)))
const pendingCategories = computed(() => productCategories.value.filter((c) => !savedCategoryIds.value.has(c.id)))
// linked categories are shown as tags, up to 20 at a time
const linkedVisibleCount = ref(20)
const visibleLinkedCategories = computed(() => linkedCategories.value.slice(0, linkedVisibleCount.value))

// drop a pending selection without unlinking (it was never linked)
function removePending(category) {
    productCategories.value = productCategories.value.filter((item) => item.id !== category.id)
}

// guards against double-submitting a link/unlink for the same row
const busyCategoryId = ref(null)

async function unlinkProductCategory(category) {
    if (!route.params.id) {
        error.value = 'Save the product before unlinking categories.'
        return
    }
    busyCategoryId.value = category.id
    const data = await unlinkCategoryRequest(route.params.id, category.id)
    busyCategoryId.value = null
    if (data.ok) {
        savedCategoryIds.value.delete(category.id)
        // only linked categories are listed, so the row disappears on unlink
        productCategories.value = productCategories.value.filter((item) => item.id !== category.id)
    } else {
        error.value = data.error
    }
}

// hide categories already opted in by the product
const visibleCategoryResults = computed(() => {
    console.log(categoryResults.value)
    return categoryResults.value.filter((category) => !productCategories.value.some((item) => item.id === category.id))
}
)

// search-as-you-type: debounce 250ms, then hit /categories/match?name=<prefix>
watch(categorySearch, (value) => {
    clearTimeout(categorySearchTimer)
    const query = value.trim()
    if (!query) {
        categoryResults.value = []
        categoryDropdownOpen.value = false
        return
    }
    categorySearchTimer = setTimeout(async () => {
        const data = await searchCategory(query)
        categoryResults.value = data.ok ? data.data.categories : []
        console.log(categoryResults.value, data)
        categoryDropdownOpen.value = true
    }, 250)
})

// pick a category from the dropdown -> queue it in the table (link happens on "Add category")
function pickCategory(category) {
    if (!productCategories.value.some((item) => item.id === category.id)) {
        productCategories.value.push({ ...category, addedAt: new Date().toLocaleString() })
        console.log('picked category', category)
    }
    categorySearch.value = ''
    categoryResults.value = []
    categoryDropdownOpen.value = false
}

// link every queued (not yet linked) category at once
const addingCategories = ref(false)

async function addCategories() {
    if (!route.params.id) {
        error.value = 'Save the product before linking categories.'
        return
    }
    const pending = productCategories.value.filter((category) => !savedCategoryIds.value.has(category.id))
    if (pending.length === 0) return
    addingCategories.value = true
    const failures = []
    for (const category of pending) {
        const data = await linkCategoryRequest(route.params.id, category.id)
        if (data.ok) {
            savedCategoryIds.value.add(category.id)
        } else {
            failures.push(category.name)
        }
    }
    addingCategories.value = false
    if (failures.length) error.value = `Could not add: ${failures.join(', ')}`
}

async function getProductInformation() {
    if (route.params.id) {
        const data = await getProduct(route.params.id)
        if (!data.ok) error.value = data.error;
        previousData.value = data.data

        // fetched product becomes the form defaults (price stored in cents -> show dollars)
        formData.name = data.data.name ?? ''
        formData.description = data.data.description ?? ''
        formData.stock_quantity = data.data.stock_quantity ?? ''
        formData.price = data.data.price != null ? (data.data.price / 100).toFixed(2) : ''

        // categories already linked on the backend -> show as Saved
        for (const category of data.data.categories ?? []) {
            if (!productCategories.value.some((item) => item.id === category.id)) {
                productCategories.value.push(category)
            }
            savedCategoryIds.value.add(category.id)
        }
    }
}

// invisible / non-renderable characters: control chars, BOM, zero-width, line separators
function hasInvisibleChars(value) {
    return /[\u0000-\u001F\u007F-\u009F\u200B-\u200F\u2028\u2029\uFEFF]/.test(value)
}

// required, max length, not only numeric/symbol+numeric (needs a letter), no invisible chars
function validateText(value, max, label) {
    const text = value?.trim() ?? ''
    if (!text) return `${label} is required.`
    if (text.length > max) return `${label} must be ${max} characters or fewer.`
    if (!/\p{L}/u.test(text)) return `${label} cannot be only numbers or symbols.`
    if (hasInvisibleChars(text)) return `${label} contains invisible or non-renderable characters.`
    return ''
}

// dollars string -> whole cents; only first 2 decimal digits count, rest ignored
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

    return !fieldErrors.name && !fieldErrors.description && !fieldErrors.stock_quantity && !fieldErrors.price
}

async function saveProduct() {
    if (!validateForm()) return
    const payload = { ...formData, price: priceToCents(formData.price) }
    console.log('save product', payload)

    let data;
    if (route.params.id) {
        data = await updateProduct(route.params.id, payload)
    } else {
        data = await createProduct(payload)
    }

    if (!data.ok) error.value = data.error;
    
    console.log(JSON.stringify(data.data))

    router.push({
        name: 'products-modify',
        params: { id: data.data.id },
        hash: '#category'
    })

}

onMounted(getProductInformation)

</script>
