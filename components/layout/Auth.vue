<script setup lang="ts"></script>

<template>
  <div
    class="relative flex items-center justify-center px-4 h-dvh lg:max-w-none lg:px-0">
    <div
      class="relative hidden h-full flex-1 flex-col bg-muted p-10 text-white lg:flex dark:border-r">
      <div class="absolute inset-0 bg-zinc-900" />
      <div
        class="relative z-20 flex items-center font-mono text-lg font-medium">
        OpenReader
      </div>
      <div class="relative z-20 mt-auto">
        <blockquote class="space-y-2">
          <p class="text-lg">
            &ldquo;OpenReader is the best way to listen to your PDFs.&rdquo;
          </p>
          <footer class="text-sm">Alexander Gekov (me)</footer>
        </blockquote>
      </div>
    </div>
    <div class="mx-auto flex-1 lg:p-8">
      <slot />
    </div>
  </div>
</template>

<style scoped></style>
