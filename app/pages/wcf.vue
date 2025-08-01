<script setup lang="ts">

const { data: wcf } = await useAsyncData('wcf', () => {
  return queryCollection('wcf')
    .first()
})

</script>

<template>
  <div>
    {{ wcf?.Data.map(d => d.Chapter) }}
  </div>
</template>
