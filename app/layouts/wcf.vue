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
      chapters: []
    },
    'Epistemology': {
      icon: 'i-lucide-book-open',
      chapters: ['1']
    },
    'Theology Proper': {
      icon: 'i-lucide-database',
      chapters: ['2', '3', '4', '5']
    },
    'Anthropology': {
      icon: 'i-lucide-user',
      chapters: ['6']
    },
    'Soteriology': {
    icon: 'i-lucide-cross',
    chapters: ['7', '8', '9', '10', '11', '12', '13', '14', '15', '16', '17', '18']
    },
     'Theonomy': {
    icon: 'i-lucide-gavel',
    chapters: ['19', '20', '21', '22', '23', '24']
     },
      'Ecclesiology': {
    icon: 'i-lucide-church',
    chapters: ['25', '26', '27', '28', '29', '30', '30']
     },
       'Eschatology': {
    icon: 'i-lucide-chevron-end',
    chapters: ['31', '32']
       },
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

  return currentChapter.value.Sections.map((_, index) => ({
    id: `section${index + 1}`,
    depth: 2,
    text: `Section ${index + 1}`,
    number: index + 1
  }))
})

function onTocMove(id: string) {
  if (!id) return
  const targetId = id.startsWith('#') ? id.slice(1) : id
  const element = import.meta.client ? document.getElementById(targetId) : null
  if (element) {
    element.scrollIntoView({ behavior: 'smooth', block: 'start' })
  }
}
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
                @move="onTocMove"
              >
                <template #link="{ link }">
                  Section {{ link.number }}
                </template>
              </UContentToc>
            </UPageAside>
          </template>

          <slot />
        </UPage>
      </UContainer>
    </UMain>

    <AppFooter />
  </div>
</template>
