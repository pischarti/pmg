<script setup lang="ts">

const navigation = ref([
  {
    title: 'Guide',
    icon: 'i-lucide-panda',
    path: '#getting-started',
    children: [
      {
        title: 'Introduction',
        path: ' I am Your Introduction. I am also a panda.',
        active: true
      }
    ]
  },
  {
    title: 'Unit 1',
    icon: 'i-lucide-fish',
    path: '/wcf/composable',
    children: [
      {
        title: 'Chapter 1',
        path: ' Hullo. I am Chapter one, the Fish, the first chapter of The Big Boring Book.'
      }
    ]
  },
  {
    title: 'Unit 2',
    icon: 'i-lucide-squirrel',
    path: '/wcf/composables',
    children: [
      {
        title: 'Chapter 2',
        path: ' I am chapter two of the Big Boring Book.'
      },
      {
        title: 'Chapter 3',
        path: ' I am chapter three of the Big Boring Book.'
      },
      {
        title: 'Chapter 4',
        path: ' I am chapter four of the Big Boring Book.'
      },
      {
        title: 'Chapter 5',
        path: ' I am chapter five of the Big Boring Book.'
      }
    ]
  }
])
</script>

<!-- <template>
  <UContentNavigation type="single" />
</template> -->


<!-- <script setup lang="ts">
import type { ContentNavigationItem } from '@nuxt/content'

const navigation = inject<Ref<ContentNavigationItem[]>>('navigation')
</script> -->

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
                :navigation="navigation"
                highlight
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
