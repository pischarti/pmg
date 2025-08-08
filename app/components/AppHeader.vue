<script setup lang="ts">
import { useWcf } from '~/composables/useWcf'
import { ref } from 'vue'

const open = ref(false)
const route = useRoute()
// const router = useRouter()
const { wcf } = await useWcf()

const items = computed(() => [{
  label: 'WCF',
  to: '/wcf',
  active: route.path.startsWith('/docs')
}, {
  label: 'Subscribe',
  to: '/pricing'
}])

const chapterItems = computed(() => (wcf.value?.Data || []).map(ch => ({
  label: `Chapter ${ch.Chapter}`,
  to: `/wcf/chapter${ch.Chapter}`,
  onSelect: () => open.value = false
})))

// function go(to: string) {
//   router.push(to)
// }
</script>

<template>
  <UHeader>
    <template #left>
      <NuxtLink
        to="/"
        class="font-semibold hidden lg:inline"
      >
        WSC
      </NuxtLink>

      <UDrawer
        v-model:open="open"
        direction="left"
        class="lg:hidden"
        title="Chapters"
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
