<script setup lang="ts">
import { ref, computed } from 'vue'
import { useWcf } from '~/composables/useWcf'

const open = ref(false)
const route = useRoute()
// const router = useRouter()
const { wcf } = await useWcf()

const items = computed(() => [
  {
    label: 'Progress',
    to: '/wcf',
    active: route.path.startsWith('/wcf')
  },
  {
    label: 'The Team',
    to: '/pricing'
  },
  {
    label: 'WOLFGANG',
    to: '/blog'
  }
])

const chapterItems = computed(() =>
  (wcf.value?.Data || []).map(ch => ({
    label: `Chapter ${ch.Chapter}`,
    to: `/wcf/chapter${ch.Chapter}`,
    onSelect: () => (open.value = false)
  }))
)

const chapters = computed(() =>
  (wcf.value?.Data || []).map(ch => ch.Chapter)
)

const currentChapterNum = computed(() => {
  const match = route.path.match(/chapter(\d+)/)
  return match ? match[1] : null
})

const currentIndex = computed(() => {
  if (!currentChapterNum.value) return -1
  return chapters.value.findIndex(c => c === currentChapterNum.value)
})

const prevChapter = computed(() =>
  currentIndex.value > 0 ? chapters.value[currentIndex.value - 1] : null
)

const nextChapter = computed(() =>
  currentIndex.value >= 0 && currentIndex.value < chapters.value.length - 1
    ? chapters.value[currentIndex.value + 1]
    : null
)

// function go (to: string) {
//   router.push(to)
// }
</script>

<template>
  <UHeader>
    <template #left>
      <NuxtLink
        to="/"
        class="font-semibold hidden lg:inline"
      > WSC </NuxtLink>

      <UDrawer
        v-model:open="open"
        direction="left"
        class="lg:hidden"
        title="Chapters"
        description="Select a chapter"
      >
        <UButton
          icon="i-lucide-book-open"
          class="lg:hidden"
        />
        <template #content>
          <UNavigationMenu
            class="overflow-y-auto"
            :items="chapterItems"
            orientation="vertical"
            variant="link"
          />
        </template>
      </UDrawer>

      <UButton
        v-if="route.path.startsWith('/wcf')"
        icon="i-lucide-chevron-left"
        :to="prevChapter ? `/wcf/chapter${prevChapter}` : undefined"
        :disabled="!prevChapter"
        color="neutral"
        variant="ghost"
        aria-label="Previous chapter"
      />

      <UButton
        v-if="route.path.startsWith('/wcf')"
        icon="i-lucide-chevron-right"
        :to="nextChapter ? `/wcf/chapter${nextChapter}` : undefined"
        :disabled="!nextChapter"
        color="neutral"
        variant="ghost"
        aria-label="Next chapter"
      />
    </template>

    <UNavigationMenu
      :items="items"
      variant="link"
    />

    <template #right>
      <UColorModeButton />

      <UButton
        icon="i-lucide-log-in"
        color="neutral"
        variant="ghost"
        to="/login"
        class="lg:hidden"
      />

      <UButton
        label="Sign in"
        color="neutral"
        variant="outline"
        to="/login"
        class="hidden lg:inline-flex"
      />

      <UButton
        label="Sign up"
        color="neutral"
        trailing-icon="i-lucide-arrow-right"
        class="hidden lg:inline-flex"
        to="/signup"
      />
    </template>

    <template #body>
      <UNavigationMenu
        :items="items"
        orientation="vertical"
        class="-mx-2.5"
      />

      <USeparator class="my-6" />

      <UButton
        label="Sign in"
        color="neutral"
        variant="subtle"
        to="/login"
        block
        class="mb-3"
      />

      <UButton
        label="Sign up"
        color="neutral"
        to="/signup"
        block
      />
    </template>
  </UHeader>
</template>
