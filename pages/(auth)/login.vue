<script setup lang="ts">
definePageMeta({
  layout: "blank",
});
</script>

<template>
  <LayoutAuth reverse>
    <div class="grid mx-auto max-w-sm gap-6">
      <div class="grid gap-2 text-center">
        <h1 class="text-2xl font-semibold tracking-tight">Welcome back</h1>
      </div>
      <AuthSignIn />
    </div>
  </LayoutAuth>
</template>
