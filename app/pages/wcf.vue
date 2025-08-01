<script setup lang="ts">

const { data: wcf } = await useAsyncData('wcf', () => {
  return queryCollection('wcf')
    .first()
})

</script>

<template>
  <div>
    wcf: {{ JSON.stringify(wcf?.Data, null, 2) }}
  </div>
</template>
