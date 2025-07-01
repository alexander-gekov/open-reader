<script setup lang="ts">
import { LucideMoon, LucideSun } from "lucide-vue-next";

const colorMode = useColorMode();
</script>

<template>
  <DropdownMenu>
    <DropdownMenuTrigger as-child>
      <Button size="icon" variant="ghost">
        <LucideSun
          class="w-4 h-4 rotate-0 scale-100 transition-all dark:-rotate-90 dark:scale-0" />
        <LucideMoon
          class="w-4 h-4 absolute rotate-90 scale-0 transition-all dark:rotate-0 dark:scale-100" />
        <span class="sr-only">Toggle theme</span>
      </Button>
    </DropdownMenuTrigger>
    <DropdownMenuContent align="end">
      <DropdownMenuItem @click="colorMode.preference = 'light'">
        Light
      </DropdownMenuItem>
      <DropdownMenuItem @click="colorMode.preference = 'dark'">
        Dark
      </DropdownMenuItem>
      <DropdownMenuItem @click="colorMode.preference = 'system'">
        System
      </DropdownMenuItem>
    </DropdownMenuContent>
  </DropdownMenu>
</template>
