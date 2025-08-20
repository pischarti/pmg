<script setup lang="ts">
import { computed } from 'vue'
import type { AccordionItem } from '@nuxt/ui'
import WcfProofsAccordion from '~/components/WcfProofsAccordion.vue'

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
const hasProofs = computed(() => (props.proofs?.length ?? 0) > 0)
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
          <WcfProofsAccordion
            v-if="hasProofs"
            :proofs="proofs"
          />
          <p
            v-else
            class="pb-3.5 text-sm text-muted"
          >
            No proofs provided.
          </p>
        </div>
        <div v-if="item.label === 'Discussion'">
          <p
            class="pb-3.5 text-sm text-muted"
          >
            No discussion provided.
          </p>
        </div>
      </template>
    </UAccordion>
  </div>
</template>
