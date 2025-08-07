<script setup lang="ts">
import { computed } from 'vue'
import { useWcf } from '~/composables/useWcf'

definePageMeta({
  layout: "wcf",
});

const route = useRoute()
const { wcf } = await useWcf()
const chapter = computed(() => {
  if (!wcf.value?.Data) return null
  return wcf.value.Data.find(ch => ch.Chapter === '1')
})
const title = computed(() => chapter.value ? `Chapter ${chapter.value.Chapter}` : 'Loading...')
const description = 'Of the Holy Scripture'
const content: Record<string, string> = { blahblah: 'blah' }
</script>

<template>
  <div>
    <UPageHero :title="title" :description="description">
      <template #top>
        <div
          class="absolute rounded-full dark:bg-primary blur-[300px] size-60 sm:size-80 transform -translate-x-1/2 left-1/2 -translate-y-80"
        />
        <LazyStarsBg />
      </template>
    </UPageHero>
    <USeparator />
    <UPageBody v-if="chapter">
      <template v-for="(section, index) in chapter.Sections" :key="index">
        <div :id="`section${index + 1}`">
          {{ section.Content }}
        </div>
        <USeparator v-if="index < chapter.Sections.length - 1" />
      </template>
    </UPageBody>
  </div>
</template>

<!-- <UPageBody v-if="route.path.includes('ch1')">
  {{ wcf.Data[0].Sections }}
  <USeparator />
</UPageBody> -->
