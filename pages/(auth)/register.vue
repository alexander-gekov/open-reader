<script setup lang="ts">
definePageMeta({
  layout: "blank",
});
</script>

<template>
  <LayoutAuth>
    <div class="grid mx-auto max-w-sm gap-6">
      <AuthSignUp />

      <p class="px-8 text-center text-sm text-muted-foreground">
        By clicking continue, you agree to our
        <a
          href="/terms"
          class="underline underline-offset-4 hover:text-primary">
          Terms of Service
        </a>
        and
        <a
          href="/privacy"
          class="underline underline-offset-4 hover:text-primary">
          Privacy Policy
        </a>
        .
      </p>
    </div>
  </LayoutAuth>
</template>

<style scoped></style>
