<template>
    <div class="space-y-6">

        <!-- button for create a new product (authenticated users only) -->
        <div class="flex items-center justify-end" v-if="auth.isAuthenticated">
            <button @click="addProduct"
                class="rounded-md bg-emerald-700 px-4 py-2 font-bold text-white transition-colors hover:bg-emerald-800 focus:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500">
                New Product
            </button>
        </div>

        <!-- table showing the results -->
        <div class="overflow-x-auto rounded-lg border border-gray-200">


            <table class="min-w-full divide-y divide-gray-200 text-left text-sm"
                v-if="Object.keys(products).length >= 0">
                <thead class="bg-gray-50">
                    <tr>
                        <th class="px-4 py-3 font-semibold text-gray-700" v-for="title in scalarTitles"
                            :key="title">{{ title.replace(/_/g, ' ').replace(/^\w/, (c) => c.toUpperCase()) }}</th>
                        <th class="px-4 py-3 font-semibold text-gray-700">Categories</th>
                        <th class="px-4 py-3 font-semibold text-gray-700" v-if="auth.isAuthenticated"></th>
                    </tr>
                </thead>
                <tbody class="divide-y divide-gray-200 bg-white">
                    <tr class="hover:bg-gray-50" v-for="(productData, index) in products.products"
                        :key="productData.id">
                        <td class="px-4 py-3" v-for="title in scalarTitles" :key="title"> {{
                            productData[title] }}</td>
                        <td class="px-4 py-3">
                            <div class="flex flex-wrap gap-1">
                                <span v-for="category in productData.categories" :key="category.id"
                                    class="inline-flex items-center rounded-full bg-emerald-50 px-3 py-1 text-sm font-medium text-emerald-700">
                                    {{ category.name }}
                                </span>
                                <span v-if="!productData.categories || productData.categories.length === 0"
                                    class="text-sm text-gray-400">&mdash;</span>
                            </div>
                        </td>
                        <td class="relative px-4 py-3 text-right" v-if="auth.isAuthenticated">
                            <button @click="toggleMenu(productData.id)" aria-label="Product actions"
                                class="rounded-md p-1.5 text-gray-500 transition-colors hover:bg-gray-100 hover:text-gray-700 focus:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500">
                                <EllipsisVertical class="h-5 w-5" />
                            </button>

                            <!-- click-away layer -->
                            <div v-if="openMenuId === productData.id" class="fixed inset-0 z-10"
                                @click="openMenuId = null"></div>

                            <!-- dropdown menu -->
                            <div v-if="openMenuId === productData.id"
                                class="absolute right-4 z-20 mt-1 w-44 overflow-hidden rounded-md border border-gray-200 bg-white py-1 shadow-lg">
                                <button @click="addCategories(productData)"
                                    class="block w-full px-4 py-2 text-left text-sm text-gray-700 hover:bg-gray-50">
                                    Add Categories
                                </button>
                                <button @click="editProduct(productData)"
                                    class="block w-full px-4 py-2 text-left text-sm text-gray-700 hover:bg-gray-50">
                                    Edit Product
                                </button>
                                <button @click="removeProduct(productData)"
                                    class="block w-full px-4 py-2 text-left text-sm text-red-600 hover:bg-red-50">
                                    Delete Product
                                </button>
                            </div>
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>

        <!-- pagination component -->
        <nav class="flex items-center justify-end gap-2" aria-label="Pagination">
            <div v-if="error" role="alert"
                class="flex items-center justify-between rounded-md border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
                <span>{{ error }}</span>
                <button @click="error = ''" aria-label="Dismiss"
                    class="ml-4 rounded p-1 leading-none text-red-500 transition-colors hover:bg-red-100 hover:text-red-700 focus:outline-none">
                    &#10005;
                </button>
            </div>
            <button @click="previous"
                class="inline-flex items-center rounded-md border border-gray-300 bg-white px-3 py-2 text-sm font-medium leading-5 text-gray-700 shadow-xs transition-colors hover:bg-gray-50 hover:text-gray-900 focus:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500">
                Previous
            </button>
            <span class="px-2 text-sm font-medium text-gray-700">Page {{ curruntPage + 1 }}</span>
            <button @click="next"
                class="inline-flex items-center rounded-md border border-gray-300 bg-white px-3 py-2 text-sm font-medium leading-5 text-gray-700 shadow-xs transition-colors hover:bg-gray-50 hover:text-gray-900 focus:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500">
                Next
            </button>
        </nav>

    </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { getProducts, deleteProduct } from '@/network/request.js'
import { useAuthStore } from '@/stores/auth'
import { EllipsisVertical } from '@lucide/vue'

const router = useRouter()
const auth = useAuthStore()

// Scalar columns to render as plain cells; categories get their own tag column.
const scalarTitles = computed(() =>
  productTableTitle.value.filter((t) => t !== 'categories' && t !== 'deleted_at'),
)

const curruntPage = ref(0) // offset
const curruntLimit = 20
const products = ref({})
const productTableTitle = ref([])
const error = ref('')
const openMenuId = ref(null)

function toggleMenu(id) {
    openMenuId.value = openMenuId.value === id ? null : id
}

function addProduct(){
    router.push({
        name: 'products-create'
    })
}

function addCategories(product) {
    router.push({
        name: 'products-modify',
        params: { id: product.id, action: 'edit' },
        hash: '#category'
    })
}

function editProduct(product) {
    router.push({ name: 'products-modify', params: { id: product.id, action: 'edit' } })
}

async function removeProduct(product) {
    console.log('delete product', product.id)
    openMenuId.value = null

    const response = await deleteProduct(product.id)
    console.log(response.ok, response.data)
    if (response.ok) {
        fetchProducts()
    }
    error.value = response.error
}

async function fetchProducts() {
    const response = await getProducts(curruntPage.value, curruntLimit)
    console.log(response.ok, response.data)
    if (response.ok) {

        if (response.data?.products.length > 0) {
            const keys = Object.keys(response.data?.products[0])
            console.log(keys)
            productTableTitle.value = keys
        }

        products.value = response.data
    }
    error.value = response.error
}

function next() {
    if (!products.value) {
        error.value = 'Something went wrong';
        return
    }
    const product = products.value

    if (product.total - (product.limit * (product.offset + 1)) > 0) {
        curruntPage.value = curruntPage.value + 1
    }

    error.value = 'No records on next page.'
}


function previous() {

    if (!products.value) {
        error.value = 'Something went wrong';
        return
    }
    const product = products.value

    if (product.offset > 0) {
        curruntPage.value = curruntPage.value - 1
    }

    error.value = 'No records on previous page.'
}

onMounted(fetchProducts)


</script>
