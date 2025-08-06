<script setup lang="ts">
definePageMeta({
  layout: 'wcf'
})

const route = useRoute()

const { data: wcf } = await useAsyncData('wcf', () => {
  return queryCollection('wcf')
    .first()
})

</script>

<template>
  <div>
    {{ route.path.includes('cat') ? 'Meow' : route.path }}
  </div>
</template>
