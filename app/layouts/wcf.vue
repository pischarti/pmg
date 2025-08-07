<script setup lang="ts">
const route = useRoute()

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
      }
    ]
  },
  {
    title: 'Epistemology',
    icon: 'i-lucide-brain',
    path: '/wcf/epistemology',
    children: [
      {
        title: 'Chapter 1',
        path: '/wcf/chapter1'
      }
    ]
  },
  {
    title: 'Theology Proper',
    icon: 'i-lucide-database',
    path: '/wcf/theologyproper',
    children: [
      {
        title: 'Chapter 2',
        path: '/wcf/chapter2'
      },
      {
        title: 'Chapter 3',
        path: '/wcf/chapter3'
      },
      {
        title: 'Chapter 4',
        path: '/wcf/chapter4'
      }
    ]
      {  
        title: 'Chapter 5'   
        path:  '/wcf/chapter5'
      }
  }     
])

const sections = [
  {
    id: 'usage',
    depth: 2,
    text: 'Usage',
    children: [
      {
        id: 'title',
        depth: 3,
        text: 'Title'
      },
      {
        id: 'color',
        depth: 3,
        text: 'Color'
      },
      {
        id: 'blah',
        depth: 3,
        text: 'route.path'
      }
    ]
  }
]

const links = ref(sections)
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

              <UContentNavigation :navigation="navigation" highlight />
            </UPageAside>
          </template>

          <template #right>
            <UPageAside>
              <UContentToc :links="links" />
            </UPageAside>
          </template>

          <slot />
          <!-- <div id="blah">
            {{ route.path.includes("cat") ? "Meow" : "route.path" }}
          </div> -->
        </UPage>
      </UContainer>
    </UMain>

    <AppFooter />
  </div>
</template>
