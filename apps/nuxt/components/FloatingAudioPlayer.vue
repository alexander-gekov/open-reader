<script setup lang="ts">
import { Button } from "@/components/ui/button";
import { Slider } from "@/components/ui/slider";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  LucideCirclePause,
  LucideCirclePlay,
  LucideFastForward,
  LucideLoader2,
  LucidePauseCircle,
  LucidePlayCircle,
  LucideRewind,
  LucideVolume1,
  LucideVolume2,
  LucideVolumeX,
} from "lucide-vue-next";
import type { AcceptableValue } from "reka-ui";

const props = defineProps<{
  volume: number[];
  playbackRate: number;
  isPlaying: boolean;
  currentTime: number;
  duration: number;
  isLoading: boolean;
}>();

const emit = defineEmits<{
  (e: "update:volume", value: number[]): void;
  (e: "update:playbackRate", value: number): void;
  (e: "toggle-play"): void;
  (e: "seek", seconds: number): void;
  (e: "toggle-mute"): void;
}>();

const progressBarRef = ref<HTMLElement>();

const formatTime = (time: number) => {
  const minutes = Math.floor(time / 60);
  const seconds = Math.floor(time % 60);
  return `${minutes}:${seconds.toString().padStart(2, "0")}`;
};

const handleProgressClick = (event: MouseEvent) => {
  if (!progressBarRef.value) return;

  const rect = progressBarRef.value.getBoundingClientRect();
  const x = event.clientX - rect.left;
  const percentage = x / rect.width;
  const newTime = percentage * props.duration;

  emit("seek", newTime - props.currentTime);
};

const handleVolumeChange = (value: number[] | undefined) => {
  if (value) {
    emit("update:volume", value);
  }
};

const handlePlaybackRateChange = (value: AcceptableValue) => {
  if (typeof value === "number") {
    emit("update:playbackRate", value);
  }
};
</script>

<template>
  <div
    class="fixed bottom-6 left-1/2 -translate-x-1/2 bg-background/80 backdrop-blur-lg border border-border rounded-lg p-4 z-50 shadow-lg w-[min(80%,600px)]">
    <div class="flex flex-col gap-2">
      <!-- Time Progress -->
      <div class="flex justify-between text-sm text-muted-foreground px-2">
        <span>{{ formatTime(currentTime) }}</span>
        <span>{{ formatTime(duration) }}</span>
      </div>

      <!-- Progress Bar -->
      <div
        class="w-full h-1.5 bg-muted/50 rounded-full cursor-pointer px-2"
        @click="handleProgressClick"
        ref="progressBarRef">
        <div
          class="h-full bg-primary rounded-full transition-all"
          :style="{ width: `${(currentTime / duration) * 100}%` }" />
      </div>

      <!-- Playback Controls -->
      <div class="flex items-center justify-between px-2">
        <!-- Volume Control -->
        <div class="flex items-center gap-2">
          <Button
            variant="ghost"
            size="icon"
            @click="$emit('toggle-mute')"
            class="h-8 w-8 hover:bg-background/50">
            <LucideVolume2 v-if="volume[0] > 0.5" class="h-4 w-4" />
            <LucideVolume1 v-else-if="volume[0] > 0" class="h-4 w-4" />
            <LucideVolumeX v-else class="h-4 w-4" />
          </Button>
          <Slider
            :model-value="volume"
            @update:model-value="handleVolumeChange"
            :min="0"
            :max="1"
            :step="0.1"
            class="w-24" />
        </div>

        <!-- Play/Pause -->
        <div class="flex items-center gap-2">
          <Button
            variant="ghost"
            size="icon"
            @click="$emit('seek', -10)"
            class="h-8 w-8 hover:bg-background/50">
            <LucideRewind class="h-4 w-4" />
          </Button>

          <Button
            variant="ghost"
            size="icon"
            @click="$emit('toggle-play')"
            :disabled="isLoading"
            class="h-10 w-10 hover:bg-background/50">
            <LucideLoader2 v-if="isLoading" class="h-5 w-5 animate-spin" />
            <LucidePlayCircle v-else-if="!isPlaying" class="h-5 w-5" />
            <LucidePauseCircle v-else class="h-5 w-5" />
          </Button>

          <Button
            variant="ghost"
            size="icon"
            @click="$emit('seek', 10)"
            class="h-8 w-8 hover:bg-background/50">
            <LucideFastForward class="h-4 w-4" />
          </Button>
        </div>

        <!-- Playback Speed -->
        <Select
          :model-value="playbackRate"
          @update:model-value="handlePlaybackRateChange">
          <SelectTrigger class="w-20 h-8">
            <SelectValue>{{ playbackRate }}x</SelectValue>
          </SelectTrigger>
          <SelectContent>
            <SelectItem :value="0.5">0.5x</SelectItem>
            <SelectItem :value="0.75">0.75x</SelectItem>
            <SelectItem :value="1">1x</SelectItem>
            <SelectItem :value="1.25">1.25x</SelectItem>
            <SelectItem :value="1.5">1.5x</SelectItem>
            <SelectItem :value="2">2x</SelectItem>
          </SelectContent>
        </Select>
      </div>
    </div>
  </div>
</template>

<style scoped>
.shadow-lg {
  box-shadow: 0 10px 15px -3px rgb(0 0 0 / 0.1), 0 4px 6px -4px rgb(0 0 0 / 0.1);
}
</style>
