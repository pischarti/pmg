<script setup lang="ts">
import { computed } from 'vue'
import type { AccordionItem } from '@nuxt/ui'
// import { useWcfDiscussion } from '#imports'

const props = defineProps<{
  discussion?: Array<{
    Id: number
    Contents: string[]
  }>
}>()

// const { getWcfStudy } = useWcfDiscussion()

const sortedDiscussion = computed(() => {
  const list = props.discussion ?? []
  return [...list].sort((a, b) => a.Id - b.Id)
})

const discussionItems = computed<AccordionItem[]>(() =>
  sortedDiscussion.value.map(p => ({
    label: `[${p.Id}] ${p.Contents.join(', ')}`
  }))
)
</script>

<template>
  <div>
    <UAccordion
      v-if="sortedDiscussion.length"
      :items="discussionItems"
      class="pl-6 sm:pl-8"
    >
      <template #content="{ index }">
        <ul class="list-disc pl-6 space-y-1 text-sm text-muted">
          <li
            v-for="con in (sortedDiscussion[index]?.Contents || [])"
            :key="con"
          />
        </ul>
      </template>
    </UAccordion>

    <p
      v-else
      class="pb-3.5 text-sm text-muted"
    >
      No 123 provided.
    </p>
  </div>
</template>
