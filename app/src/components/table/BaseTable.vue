<template>
  <div class="overflow-x-auto md:overflow-visible rounded-lg border border-gray-200">
    <table class="min-w-full divide-y divide-gray-200 text-left text-sm">
      <thead class="bg-gray-50">
        <tr>
          <th
            v-for="col in columns"
            :key="col.key"
            class="px-4 py-3 font-semibold text-gray-700"
          >
            <slot :name="`header-${col.key}`" :column="col">
              {{ col.label }}
            </slot>
          </th>
        </tr>
      </thead>
      <tbody class="divide-y divide-gray-200 bg-white">
        <tr
          v-for="(item, index) in items"
          :key="item.id || index"
          class="hover:bg-gray-50"
        >
          <td
            v-for="col in columns"
            :key="col.key"
            class="px-4 py-3 text-gray-900"
          >
            <slot :name="`cell-${col.key}`" :item="item" :value="item[col.key]" :index="index">
              {{ item[col.key] }}
            </slot>
          </td>
        </tr>
        <tr v-if="!items || items.length === 0">
          <td :colspan="columns.length" class="px-4 py-8 text-center text-gray-500">
            No records found.
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup>
defineProps({
  columns: {
    type: Array,
    required: true,
  },
  items: {
    type: Array,
    required: true,
  },
})
</script>
