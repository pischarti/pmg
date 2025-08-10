<script setup lang="ts">
import { computed } from 'vue'
import type { AccordionItem } from '@nuxt/ui'

const items: AccordionItem[] = [
  {
    label: 'Proofs',
    icon: 'i-lucide-book-open'
  },
  {
    label: 'Discussion',
    icon: 'i-lucide-message-circle'
  }
]

const props = defineProps<{
  index: number
  section: {
    Content: string
    ContentWithProofs?: string
  }
  proofs?: Array<{
    Id: number
    References: string[]
  }>
}>()

const sectionId = computed(() => `section${props.index + 1}`)
const sortedProofs = computed(() => {
  const list = props.proofs ?? []
  return [...list].sort((a, b) => a.Id - b.Id)
})
</script>

<template>
  <div
    :id="sectionId"
    class="scroll-mt-[calc(48px+var(--ui-header-height))]"
  >
    <h2 class="text-2xl font-bold text-highlighted mt-8 mb-4">
      Section {{ index + 1 }}
    </h2>
    {{ section.ContentWithProofs }}
    <UAccordion :items="items">
      <template #content="{ item }">
        <div v-if="item.label === 'Proofs'">
          <div v-if="proofs?.length">
            <ul class="list-disc pl-6 space-y-1 text-sm text-muted">
              <li
                v-for="proof in sortedProofs"
                :key="proof.Id"
              >
                <span class="font-medium">[{{ proof.Id }}]: </span>
                <span>
                  {{ proof.References.join(', ') }}
                </span>
              </li>
            </ul>
          </div>
          <p
            v-else
            class="pb-3.5 text-sm text-muted"
          >
            No proofs provided.
          </p>
        </div>
      </template>
    </UAccordion>
  </div>
</template>
