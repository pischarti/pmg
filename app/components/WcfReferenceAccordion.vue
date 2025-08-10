<script setup lang="ts">
import { computed, onMounted } from 'vue'
import type { AccordionItem } from '@nuxt/ui'

const props = defineProps<{ reference: string }>()

// Use the Bible composable
const { getBibleVerses } = useBible()

// Use computed to handle async content loading
const refItems = computed<AccordionItem[]>(() => [
  {
    label: props.reference,
    content: 'Loading...' // Will be updated when async function completes
  }
])

// Load the content when component mounts
onMounted(async () => {
  const content = await getBibleVerses(props.reference)
  if (refItems.value[0]) {
    refItems.value[0].content = content
  }
})
</script>

<template>
  <UAccordion :items="refItems" />
</template>
