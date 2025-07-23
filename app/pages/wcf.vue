<script setup lang="ts">
const { data: page } = await useAsyncData("pricing", () =>
  queryCollection("pricing").first(),
);

const title = page.value?.seo?.title || page.value?.title;
const description = page.value?.seo?.description || page.value?.description;

useSeoMeta({
  title,
  ogTitle: title,
  description,
  ogDescription: description,
});

defineOgImageComponent("Saas");

const isYearly = ref("0");

const items = ref([
  {
    label: "Monthly",
    value: "0",
  },
  {
    label: "Yearly",
    value: "1",
  },
]);
</script>

<template>

</template>
