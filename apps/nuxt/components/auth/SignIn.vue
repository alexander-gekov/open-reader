<script setup lang="ts"></script>

<template>
  <div class="grid gap-6">
    <div class="flex flex-col gap-4">
      <SignIn
        :appearance="{
          variables: {
            border: 0,
          },
        }" />
    </div>
  </div>

  <div class="mt-4 text-center text-sm text-muted-foreground">
    Don't have an account?
    <NuxtLink to="/register" class="underline"> Sign up </NuxtLink>
  </div>
</template>

<style scoped></style>
