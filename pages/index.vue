<script setup lang="ts">
import { toast } from "vue-sonner";

definePageMeta({
  title: "Home",
  layout: "default",
});

interface UploadResponse {
  success: boolean;
  key: string;
}

const files = ref<File[]>([]);
const isUploading = ref(false);

async function uploadToR2() {
  if (!files.value.length) return;

  isUploading.value = true;

  try {
    const file = files.value[0];
    const formData = new FormData();
    formData.append("file", file);

    const response = await $fetch<UploadResponse>("/api/upload", {
      method: "POST",
      body: formData,
    });

    if (response.success) {
      files.value = [];
      toast.success("File uploaded successfully");
    }
  } catch (error) {
    console.error(error);
    toast.error("Failed to upload file", {
      description: error instanceof Error ? error.message : "Unknown error",
    });
  } finally {
    isUploading.value = false;
  }
}
</script>

<template>
  <div class="container mx-auto">
    <div class="space-y-6 p-8 dark:bg-black">
      <FileUpload
        v-model="files"
        class="rounded-lg border border-dashed border-neutral-200 dark:border-neutral-800">
        <FileUploadGrid />
      </FileUpload>

      <div class="flex justify-end">
        <Button
          :disabled="!files.length || isUploading"
          @click="uploadToR2"
          :loading="isUploading">
          Upload
        </Button>
      </div>
    </div>
  </div>
</template>
