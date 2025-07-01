<script setup lang="ts">
import { cn } from "../../lib/utils";
</script>

<template>
  <div :class="cn('grid gap-6', $attrs.class ?? '')">
    <SignUp />
  </div>
</template>
