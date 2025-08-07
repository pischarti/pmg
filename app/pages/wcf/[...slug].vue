<script setup lang="ts">
definePageMeta({
  layout: "wcf",
});

const route = useRoute()

const { data: wcf } = await useAsyncData("wcf", () => {
  return queryCollection("wcf").first()
})

const title = 'Chapter I'
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
    <UPageBody
      v-for="(ch, index) in wcf.Data"
      :key="index"
    >
      <div v-if="route.path.includes('ch1') && index == 0">
        <div
          v-for="(ss, si) in ch.Sections"
          :key="si"
        >
          <div :id="`ss${si}`">
            {{ ss.Content }}
          </div>
          <USeparator />
        </div>
      </div>
      <div v-if="route.path.includes('ch2') && index == 1">
        {{ ch }}
      </div>
    </UPageBody>
  </div>
</template>

<!-- <UPageBody v-if="route.path.includes('ch1')">
  {{ wcf.Data[0].Sections }}
  <USeparator />
</UPageBody> -->
