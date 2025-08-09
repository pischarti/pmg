<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  index: number
  section: {
    Content: string
    ContentWithProofs?: string
  }
}>()

const sectionId = computed(() => `section${props.index + 1}`)
</script>

<template>
  <div
    :id="sectionId"
    class="scroll-mt-[calc(48px+var(--ui-header-height))]"
  >
    <h2 class="text-2xl font-bold text-highlighted mt-8 mb-4">
      Section {{ index + 1 }}
    </h2>
    {{ section.Content }}
  </div>
</template>
