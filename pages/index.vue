<script setup lang="ts">
definePageMeta({
  title: "Home",
  layout: "default",
});

const files = ref<File[]>([]);
</script>

<template>
  <div class="container mx-auto">
    <div class="space-y-6 p-8 dark:bg-black">
      <FileUpload
        v-model="files"
        class="rounded-lg border border-dashed border-neutral-200 dark:border-neutral-800">
        <template #default="{ openFileDialog }">
          <FileUploadGrid :open-file-dialog="openFileDialog" />
        </template>
      </FileUpload>
    </div>
  </div>
</template>
