<template>
    <div class="space-y-6">

        <!-- inline add: type a name and add without leaving the page (authenticated users only) -->
        <div class="flex items-end justify-end gap-2" v-if="auth.isAuthenticated">
            <div>
                <label for="new-category" class="block text-sm font-medium text-gray-700">New category</label>
                <input id="new-category" type="text" v-model="newCategory" @keyup.enter="addCategory"
                    placeholder="Category name"
                    class="mt-1 block w-64 rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-emerald-500 focus:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500" />
            </div>
            <button type="button" @click="addCategory" :disabled="addingCategory"
                class="rounded-md bg-emerald-700 px-4 py-2 font-bold text-white transition-colors hover:bg-emerald-800 focus:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500 disabled:opacity-50">
                Add
            </button>
        </div>

        <!-- table showing the results -->
        <div class="overflow-x-auto rounded-lg border border-gray-200">


            <table class="min-w-full divide-y divide-gray-200 text-left text-sm"
                v-if="Object.keys(categories).length >= 0">
                <thead class="bg-gray-50">
                    <tr>
                        <th class="px-4 py-3 font-semibold text-gray-700" v-for="categorieTitle in scalarCategoryTitles"
                            :key="categorieTitle">{{ categorieTitle.replace(/_/g, ' ').replace(/^\w/, (c) => c.toUpperCase()) }}</th>
                    </tr>
                </thead>
                <tbody class="divide-y divide-gray-200 bg-white">
                    <tr class="hover:bg-gray-50" v-for="(categoryData, index) in categories.categories"
                        :key="categoryData.id">
                        <td class="px-4 py-3" v-for="categorieTitle in scalarCategoryTitles" :key="categorieTitle"> {{
                            categoryData[categorieTitle] }}</td>
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

        <!-- error alert -->


    </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { getCategories, createCategory } from '@/network/request.js'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()

const curruntPage = ref(0) // offset
const curruntLimit = 20
const categories = ref({})
const categoryTableTitle = ref([])
const scalarCategoryTitles = computed(() =>
  categoryTableTitle.value.filter((t) => t !== 'deleted_at'),
)
const error = ref('')
const newCategory = ref('')
const addingCategory = ref(false)

async function addCategory() {
    const name = newCategory.value.trim()
    if (!name) {
        error.value = 'Category name is required.'
        return
    }
    addingCategory.value = true
    const response = await createCategory(name)
    addingCategory.value = false
    if (response.ok) {
        newCategory.value = ''
        await fetchCategories()
    }
    error.value = response.error
}

async function fetchCategories() {
    const response = await getCategories(curruntPage.value, curruntLimit)
    console.log(response.ok, response.data)
    if (response.ok) {

        if (response.data?.categories?.length > 0) {
            const keys = Object.keys(response.data?.categories[0])
            console.log(keys)
            categoryTableTitle.value = keys
        }

        categories.value = response.data
    }
    error.value = response.error
}

function next() {
    if (!categories.value) {
        error.value = 'Something went wrong';
        return
    }
    if ((curruntPage.value + 1) * curruntLimit < categories.value.total) {
        curruntPage.value = curruntPage.value + 1
        fetchCategories()
    } else {
        error.value = 'No records on next page.'
    }
}


function previous() {

    if (!categories.value) {
        error.value = 'Something went wrong';
        return
    }
    if (curruntPage.value > 0) {
        curruntPage.value = curruntPage.value - 1
        fetchCategories()
    } else {
        error.value = 'No records on previous page.'
    }
}

onMounted(fetchCategories)


</script>
