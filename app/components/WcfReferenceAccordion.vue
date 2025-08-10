<script setup lang="ts">
import type { AccordionItem } from '@nuxt/ui'

const props = defineProps<{ reference: string }>()

const refItems: AccordionItem[] = [
  {
    label: props.reference,
    content: props.reference
  }
]
</script>

<template>
  <UAccordion :items="refItems" />
</template>
