<template>
  <div
    class="min-h-screen bg-background justify-between items-center mx-auto max-w-4xl">
    <header
      class="sticky top-0 z-50 w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <nav class="container flex px-4 lg:px-0 h-14 items-center">
        <NuxtLink to="/" class="mr-6 flex items-center space-x-2">
          <span class="font-bold">Open Reader</span>
        </NuxtLink>

        <div
          class="flex flex-1 items-center justify-between space-x-2 md:justify-end">
          <div class="w-full flex-1 md:w-auto md:flex-none">
            <!-- Add search or other controls here -->
          </div>
          <div class="flex items-center gap-2">
            <ColorModeSwitch />
            <div v-if="isSignedIn" class="flex items-center gap-2">
              <NuxtLink to="/settings">
                <Button variant="ghost" size="icon">
                  <LucideSettings2 class="h-5 w-5" />
                  <span class="sr-only">Settings</span>
                </Button>
              </NuxtLink>
              <NuxtLink to="/reader" class="mr-4">
                <Button variant="ghost" size="icon">
                  <LucideBookOpen class="h-5 w-5" />
                  <span class="sr-only">Library</span>
                </Button>
              </NuxtLink>
              <UserButton />
            </div>
          </div>
        </div>
      </nav>
    </header>

    <main class="flex-1">
      <slot />
    </main>
  </div>
</template>

<script setup lang="ts">
import { Button } from "@/components/ui/button";
import { LucideBookOpen, LucideSettings2 } from "lucide-vue-next";

const { isSignedIn } = useAuth();
</script>

<style scoped></style>
