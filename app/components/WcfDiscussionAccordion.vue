<script setup lang="ts">
import { computed } from 'vue'
import type { AccordionItem } from '@nuxt/ui'

const props = defineProps<{
  discussion?: Array<{
    Id: number
    References: string[]
  }>
}>()

const sortedDiscussion = computed(() => {
  const list = props.discussion ?? []
  return [...list].sort((a, b) => a.Id - b.Id)
})

const proofItems = computed<AccordionItem[]>(() =>
  sortedProofs.value.map(p => ({
    label: `[${p.Id}] ${p.References.join(', ')}`
  }))
)
</script>

<template>
  <div>
    <UAccordion
      v-if="sortedProofs.length"
      :items="proofItems"
      class="pl-6 sm:pl-8"
    >
      <template #content="{ index }">
        <ul class="list-disc pl-6 space-y-1 text-sm text-muted">
          <li
            v-for="ref in (sortedProofs[index]?.References || [])"
            :key="ref"
          >
            <WcfReferenceAccordion :reference="ref" />
          </li>
        </ul>
      </template>
    </UAccordion>

    <p
      v-else
      class="pb-3.5 text-sm text-muted"
    >
      No proofs provided.
    </p>
  </div>
</template>
