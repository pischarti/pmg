<script setup lang="ts">

const navigation = ref([
  {
    title: 'Guide',
    icon: 'i-lucide-book-open',
    path: '#getting-started',
    children: [
      {
        title: 'Introduction',
        path: '/wcf/introduction',
        active: true
      },
      {
        title: 'Installation',
        path: '/wcf/installation'
      }
    ]
  },
  {
    title: 'Unit 2',
    icon: 'i-lucide-album',
    path: '/wcf/Unit 2',
    children: [
      {
        title: 'chapter 2'
        path: '/wcf/chapter 2'
        
      },
      {
     
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
