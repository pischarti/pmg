<script setup lang="ts">
// Find a single author
const { data: Baptiste } = await useAsyncData('larbish', () => {
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
    author: {{ Baptiste?.meta.body.name }}
    <p />
    authors: {{ authors.map(a => a.meta.body.name) }}
  </div>
</template>
