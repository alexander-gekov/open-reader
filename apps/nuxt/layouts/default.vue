<template>
  <div
    class="min-h-screen bg-background justify-between items-center mx-auto max-w-4xl">
    <SignedIn>
      <header
        class="sticky top-0 z-50 w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
        <nav class="container flex h-14 items-center">
          <NuxtLink to="/" class="mr-6 flex items-center space-x-2">
            <span class="font-bold">Open Reader</span>
          </NuxtLink>

          <div
            class="flex flex-1 items-center justify-between space-x-2 md:justify-end">
            <div class="w-full flex-1 md:w-auto md:flex-none">
              <!-- Add search or other controls here -->
            </div>
            <div class="flex items-center">
              <NuxtLink to="/settings" class="mr-4">
                <Button variant="ghost" size="icon">
                  <LucideSettings2 class="h-5 w-5" />
                  <span class="sr-only">Settings</span>
                </Button>
              </NuxtLink>
              <ColorModeSwitch />
            </div>
          </div>
        </nav>
      </header>

      <main class="flex-1">
        <slot />
      </main>
    </SignedIn>

    <SignedOut>
      <div class="flex min-h-screen items-center justify-center">
        <div class="mx-auto w-full max-w-sm space-y-6">
          <div class="space-y-2 text-center">
            <h1 class="text-2xl font-semibold tracking-tight">
              Welcome to Open Reader
            </h1>
            <p class="text-sm text-muted-foreground">
              Please sign in to continue
            </p>
          </div>
          <SignIn />
        </div>
      </div>
    </SignedOut>
  </div>
</template>

<script setup lang="ts">
import { Button } from "@/components/ui/button";
import { LucideSettings2 } from "lucide-vue-next";
</script>

<style scoped></style>
