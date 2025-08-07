<script setup lang="ts">
import { computed } from 'vue'
import { useWcf } from '~/composables/useWcf'

const route = useRoute()
const { wcf } = await useWcf()

const currentChapter = computed(() => {
  if (!wcf.value?.Data) return null
  const chapterMatch = route.path.match(/chapter(\d+)/)
  if (!chapterMatch) return null
  return wcf.value.Data.find(ch => ch.Chapter === chapterMatch[1])
})

const navigation = computed(() => {
  if (!wcf.value?.Data) return []

  // Group chapters into categories
  const categories = {
    'Introduction': {
      icon: 'i-lucide-book-open',
      chapters: ['1']
    },
    'Theology Proper': {
      icon: 'i-lucide-database',
      chapters: ['2', '3', '4', '5']
    },
    'Anthropology': {
      icon: 'i-lucide-user',
      chapters: ['6', '7', '8', '9']
    }
  }

  return Object.entries(categories).map(([title, { icon, chapters }]) => ({
    title,
    icon,
    path: `/wcf/${title.toLowerCase().replace(/\s+/g, '')}`,
    children: chapters.map(chapterNum => ({
      title: `Chapter ${chapterNum}`,
      path: `/wcf/chapter${chapterNum}`,
      active: route.path.includes(`chapter${chapterNum}`)
    }))
  }))
})

const links = computed(() => {
  if (!currentChapter.value) return []

  return currentChapter.value.Sections.map((section, index) => {
    // Take the first sentence or first 50 characters as the section title
    const text = section.Content.split('.')[0] + '.'
    const sectionTitle = text.length > 50 ? text.substring(0, 47) + '...' : text

    return {
      id: `section${index + 1}`,
      depth: 2,
      text: sectionTitle
    }
  })
})
</script>

<template>
  <div>
    <AppHeader />

    <UMain>
      <UContainer>
        <UPage>
          <template #left>
            <UPageAside>
              <template #top>
                <UContentSearchButton :collapsed="false" />
              </template>

              <UContentNavigation
                v-if="navigation.length"
                :navigation="navigation"
                highlight
              />
            </UPageAside>
          </template>

          <template #right>
            <UPageAside>
              <UContentToc
                v-if="links.length"
                :links="links"
              />
            </UPageAside>
          </template>

          <slot />
        </UPage>
      </UContainer>
    </UMain>

    <AppFooter />
  </div>
</template>
