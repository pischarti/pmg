<script setup lang="ts">
// Find a single author
const { data: author } = await useAsyncData('larbish', () => {
  return queryCollection('authors')
    .where('name', '=', 'Baptiste Leproux')
    .first()
})

// Get all authors
const { data: authors } = await useAsyncData('authors', () => {
  return queryCollection('authors')
    .order('name', 'DESC')
    .all()
})
</script>

<template>
  <div>
    author: {{ author }}
    <p />
    authors: {{ authors }}
  </div>
</template>
