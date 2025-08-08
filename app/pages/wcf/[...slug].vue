<script setup lang="ts">
import { computed } from 'vue'
import { useWcf } from '~/composables/useWcf'

definePageMeta({
  layout: 'wcf'
})

const _route = useRoute()
const { wcf } = await useWcf()
const chapter = computed(() => {
  if (!wcf.value?.Data) return null
  const chapterMatch = _route.path.match(/chapter(\d+)/)
  if (!chapterMatch) return null
  return wcf.value.Data.find(ch => ch.Chapter === chapterMatch[1])
})

const heroTitle = computed(() => {
  if (chapter.value) return chapter.value.Title
  if (_route.path === '/wcf' || _route.path === '/wcf/introduction') {
    return wcf.value?.Metadata?.OriginStory || ''
  }
  return 'Loading...'
})

const title = 'Westminster Confession of Faith'
</script>

<template>
  <div>
    <UPageHero
      :headline="title"
      :description="heroTitle"
    >
      <template #top>
        <div
          class="absolute rounded-full dark:bg-primary blur-[300px] size-60 sm:size-80 transform -translate-x-1/2 left-1/2 -translate-y-80"
        />
        <LazyStarsBg />
      </template>
    </UPageHero>
    <USeparator />
    <UPageBody v-if="chapter">
      <template
        v-for="(section, index) in chapter.Sections"
        :key="index"
      >
        <div
          :id="`section${index + 1}`"
          class="scroll-mt-[calc(48px+var(--ui-header-height))]"
        >
          <h2 class="text-2xl font-bold text-highlighted mt-8 mb-4">
            Section {{ index + 1 }}
          </h2>
          {{ section.Content }}
        </div>
        <USeparator v-if="index < chapter.Sections.length - 1" />
      </template>
    </UPageBody>
  </div>
</template>
