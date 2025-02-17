<script setup>
defineProps({
  title: String,
  value: [String, Number],
});
</script>

<template>
  <div class="flex justify-between items-center bg-gray-100 p-4 rounded-lg shadow-md">
    <p class="text-sm font-medium text-gray-600">{{ title }}: {{ value }}</p>

  </div>
</template>