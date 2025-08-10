<script setup lang="ts">
import { computed, onMounted } from 'vue'
import type { AccordionItem } from '@nuxt/ui'

const props = defineProps<{ reference: string }>()

// Function to get bible verses from the API
const getBibleVerses = async (reference: string): Promise<string> => {
  try {
    // Use the existing KJV API endpoint
    const response = await $fetch<{ verses?: Array<{ ref: string, text: string }>, error?: string }>('/api/kjv', {
      query: { q: reference }
    })

    if (response.verses && Array.isArray(response.verses)) {
      return response.verses
        .map((verse: { ref: string, text: string }) => `${verse.ref}: ${verse.text}`)
        .join('\n\n')
    }

    return reference
  } catch (error) {
    console.error('Error fetching bible verses:', error)
    return reference
  }
}

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
