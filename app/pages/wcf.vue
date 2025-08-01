<script setup lang="ts">
// Find a single author
const { data: wcf } = await useAsyncData('wcf', () => {
  return queryCollection('wcf')
    .first()
})

</script>

<template>
  <div>
    wcf: {{ wcf?.meta.body.Metadata.AlternativeTitles }}
  </div>
</template>
