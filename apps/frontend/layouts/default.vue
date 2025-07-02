<template>
  <ClerkLoaded>
    <nav class="flex justify-between items-center mx-auto max-w-6xl w-full p-6">
      <div class="flex flex-col">
        <NuxtLink to="/">OpenReader</NuxtLink>
        <p class="text-sm text-muted-foreground">
          Open Source ElevenReader Alternative.
        </p>
      </div>
      <ColorModeSwitch />
    </nav>
    <SignedIn>
      <div class="mx-auto max-w-6xl w-full">
        <NuxtPage />
      </div>
    </SignedIn>
    <SignedOut>
      <RedirectToSignIn />
    </SignedOut>
  </ClerkLoaded>
</template>

<script setup lang="ts">
const { userId } = useAuth();

watch(
  () => userId.value,
  async () => {
    if (userId.value === null) {
      navigateTo("/login");
    }
  }
);
</script>

<style scoped></style>
