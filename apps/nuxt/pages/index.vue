<script setup lang="ts">
import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent,
  CardFooter,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Progress } from "@/components/ui/progress";
import { ScrollArea } from "@/components/ui/scroll-area";
import {
  LucideCirclePause,
  LucideCirclePlay,
  LucideLoader2,
  LucidePlayCircle,
  LucideUpload,
} from "lucide-vue-next";
import { Slider } from "@/components/ui/slider";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

definePageMeta({
  title: "Home",
  layout: "default",
});

interface UploadResponse {
  success: boolean;
  message: string;
  chunks: string[];
  audioId: string;
  totalChunks: number;
}

interface CurrentDoc {
  audioId: string;
  chunks: string[];
  totalChunks: number;
  currentChunk: number;
}

interface AudioResponse {
  waiting?: boolean;
  message?: string;
  retry?: number;
  completed?: boolean;
}

const files = ref<File[]>([]);
const isUploading = ref(false);
const currentDoc = ref<CurrentDoc | null>(null);
const isPlayingAudio = ref(false);
const audioElement = ref<HTMLAudioElement | null>(null);

const audioPlayer = ref<HTMLAudioElement | null>(null);
const nextAudioPlayer = ref<HTMLAudioElement | null>(null);
const isPlaying = ref(false);
const volume = ref([1]);
const playbackRate = ref(1);
const currentTime = ref(0);
const duration = ref(0);
const progressBarRef = ref<HTMLElement>();
const previousVolume = ref([1]);
const error = ref<string | null>(null);
const isPreloading = ref(false);

const formatTime = (time: number) => {
  const minutes = Math.floor(time / 60);
  const seconds = Math.floor(time % 60);
  return `${minutes}:${seconds.toString().padStart(2, "0")}`;
};

const handleProgressClick = (event: MouseEvent) => {
  if (!progressBarRef.value || !audioPlayer.value) return;

  const rect = progressBarRef.value.getBoundingClientRect();
  const x = event.clientX - rect.left;
  const percentage = x / rect.width;
  const newTime = percentage * duration.value;

  audioPlayer.value.currentTime = newTime;
};

const togglePlay = async () => {
  if (!audioPlayer.value) return;

  if (isPlaying.value) {
    audioPlayer.value.pause();
    isPlaying.value = false;
  } else {
    isPlaying.value = true;
    await playNextChunk();
  }
};

const toggleMute = () => {
  if (!audioPlayer.value) return;

  if (volume.value[0] > 0) {
    previousVolume.value = [...volume.value];
    volume.value = [0];
  } else {
    volume.value = [...previousVolume.value];
  }
};

const seek = (seconds: number) => {
  if (!audioPlayer.value) return;

  audioPlayer.value.currentTime = Math.max(
    0,
    Math.min(audioPlayer.value.currentTime + seconds, duration.value)
  );
};

watch(volume, (newVolume) => {
  if (audioPlayer.value) {
    audioPlayer.value.volume = newVolume[0];
  }
  if (nextAudioPlayer.value) {
    nextAudioPlayer.value.volume = newVolume[0];
  }
});

watch(playbackRate, (newRate) => {
  if (audioPlayer.value) {
    audioPlayer.value.playbackRate = Number(newRate);
  }
  if (nextAudioPlayer.value) {
    nextAudioPlayer.value.playbackRate = Number(newRate);
  }
});

const preloadNextChunk = async () => {
  if (
    !currentDoc.value ||
    currentDoc.value.currentChunk >= currentDoc.value.totalChunks - 1
  )
    return;

  try {
    isPreloading.value = true;
    const nextChunkIndex = currentDoc.value.currentChunk + 1;

    // Trigger the processing of next chunk immediately
    try {
      await fetch(
        `http://localhost:8080/start-next/${currentDoc.value.currentChunk}`
      );
    } catch (err) {
      console.error("Failed to trigger next chunk processing:", err);
    }

    // Start polling for the next chunk
    while (isPreloading.value) {
      const response = await fetch(
        `http://localhost:8080/audio/status/${nextChunkIndex}`
      );
      const data = await response.json();

      if (data.status === "ready" && data.url) {
        const audioUrl = `http://localhost:8080${data.url}`;

        // Create and setup the next audio player
        nextAudioPlayer.value = new Audio();

        // Set up event listeners before setting src to catch loading events
        nextAudioPlayer.value.preload = "auto"; // Force preloading

        // Create a promise that resolves when enough data is loaded
        const canPlayPromise = new Promise((resolve) => {
          nextAudioPlayer.value?.addEventListener("canplaythrough", resolve, {
            once: true,
          });
        });

        // Set properties and start loading
        nextAudioPlayer.value.playbackRate = playbackRate.value;
        nextAudioPlayer.value.volume = volume.value[0];
        nextAudioPlayer.value.src = audioUrl;

        // Wait for enough data to be loaded
        await canPlayPromise;

        isPreloading.value = false;
        return;
      }

      if (data.status === "error") {
        console.error("Error preloading next chunk:", data.error);
        isPreloading.value = false;
        return;
      }

      // Wait before polling again
      await new Promise((resolve) => setTimeout(resolve, 500)); // Reduced polling interval
    }
  } catch (err) {
    console.error("Error preloading next chunk:", err);
    isPreloading.value = false;
  }
};

const setupAudioPlayer = () => {
  if (!audioPlayer.value) return;

  audioPlayer.value.addEventListener("ended", async () => {
    console.log("Audio ended");

    // Move to next chunk if available
    if (
      currentDoc.value &&
      currentDoc.value.currentChunk < currentDoc.value.totalChunks - 1 &&
      nextAudioPlayer.value
    ) {
      currentDoc.value.currentChunk++;
      isPlaying.value = true;

      // Swap players immediately
      if (audioPlayer.value) {
        const oldPlayer = audioPlayer.value;
        audioPlayer.value = nextAudioPlayer.value;
        nextAudioPlayer.value = null;

        // Start playing immediately - the audio should be preloaded
        try {
          const playPromise = audioPlayer.value.play();
          if (playPromise) {
            playPromise.catch((error) => {
              console.error("Error during playback:", error);
              isPlaying.value = false;
            });
          }
        } catch (error) {
          console.error("Error starting playback:", error);
          isPlaying.value = false;
        }

        // Clean up old player after starting new playback
        oldPlayer.remove();
        setupAudioPlayer();

        // Start preloading next chunk immediately
        preloadNextChunk();
      }
    } else {
      stopPlayback();
    }
  });

  audioPlayer.value.addEventListener("error", (e) => {
    console.error("Audio error:", e);
    error.value = "Error playing audio";
    isPlaying.value = false;
  });

  audioPlayer.value.addEventListener("playing", () => {
    isPlaying.value = true;
    // Start preloading the next chunk when current starts playing
    if (
      currentDoc.value &&
      currentDoc.value.currentChunk < currentDoc.value.totalChunks - 1
    ) {
      preloadNextChunk();
    }
  });

  audioPlayer.value.addEventListener("pause", () => {
    isPlaying.value = false;
  });

  audioPlayer.value.addEventListener("timeupdate", () => {
    currentTime.value = audioPlayer.value?.currentTime || 0;
  });

  audioPlayer.value.addEventListener("loadedmetadata", () => {
    duration.value = audioPlayer.value?.duration || 0;
  });
};

const playNextChunk = async () => {
  if (!currentDoc.value) return;

  try {
    console.log("Checking audio status...");
    const response = await fetch(
      `http://localhost:8080/audio/status/${currentDoc.value.currentChunk}`
    );
    const data = await response.json();
    console.log("Audio status:", data);

    if (data.status === "ready" && data.url) {
      const audioUrl = `http://localhost:8080${data.url}`;
      console.log("Playing audio from URL:", audioUrl);

      if (!audioPlayer.value) {
        audioPlayer.value = new Audio();
        setupAudioPlayer();
      }

      // Set up event listeners before setting src
      audioPlayer.value.preload = "auto";
      const canPlayPromise = new Promise((resolve) => {
        audioPlayer.value?.addEventListener("canplaythrough", resolve, {
          once: true,
        });
      });

      // If this is a new chunk, set it up for playback
      if (audioPlayer.value.src !== audioUrl) {
        audioPlayer.value.src = audioUrl;
        audioPlayer.value.playbackRate = playbackRate.value;
        audioPlayer.value.volume = volume.value[0];

        // Wait for enough data to be loaded
        await canPlayPromise;
      }

      // Start playing
      if (!audioPlayer.value.src.includes(data.url) || isPlaying.value) {
        try {
          const playPromise = audioPlayer.value.play();
          if (playPromise) {
            playPromise.catch((error) => {
              console.error("Error during playback:", error);
              isPlaying.value = false;
            });
          }
        } catch (error) {
          console.error("Error starting playback:", error);
          isPlaying.value = false;
        }
        console.log("Audio started playing");
      }
      return;
    }

    if (data.status === "error") {
      console.error("Error generating audio:", data.error);
      error.value = data.error || "Failed to generate audio";
      isPlaying.value = false;
      return;
    }

    // If still processing, wait and try again
    await new Promise((resolve) => setTimeout(resolve, 500)); // Reduced polling interval
    await playNextChunk();
  } catch (err) {
    console.error("Error playing audio:", err);
    error.value = "Failed to play audio";
    isPlaying.value = false;
  }
};

const startAudioPlayback = async () => {
  if (!currentDoc.value) return;

  currentDoc.value.currentChunk = 0;
  isPlaying.value = true;
  await playNextChunk();
};

// Handles a single PDF file upload and sets the current document once processed
const uploadFile = async (file: File) => {
  if (file.type !== "application/pdf") {
    alert("Please upload a PDF file");
    return;
  }

  if (file.size > 20 * 1024 * 1024) {
    // 20 MB limit
    alert("File size should be less than 20MB");
    return;
  }

  try {
    isUploading.value = true;

    const formData = new FormData();
    formData.append("file", file);

    const response = await $fetch<UploadResponse>("/api/upload", {
      method: "POST",
      body: formData,
    });

    currentDoc.value = {
      audioId: response.audioId,
      chunks: response.chunks,
      totalChunks: response.totalChunks,
      currentChunk: 0,
    };
  } catch (error) {
    console.error("Upload failed:", error);
    alert("Failed to upload and process the PDF");
  } finally {
    isUploading.value = false;
  }
};

// Triggered by the Upload button
const uploadToR2 = async () => {
  if (!files.value.length) return;

  // Currently we only support uploading the first selected PDF
  await uploadFile(files.value[0]);

  // Clear the list after successful upload so the UI resets
  files.value = [];
};

const playAudio = async (audioData: Uint8Array) => {
  const blob = new Blob([audioData], { type: "audio/mpeg" });
  const url = URL.createObjectURL(blob);
  const audio = new Audio(url);

  try {
    await new Promise<void>((resolve, reject) => {
      audio.onended = () => resolve();
      audio.onerror = reject;
      audio.play();
    });
  } finally {
    URL.revokeObjectURL(url);
  }
};

const resetUpload = () => {
  currentDoc.value = null;
  files.value = [];
  isPlayingAudio.value = false;
  if (audioElement.value) {
    audioElement.value.pause();
    audioElement.value = null;
  }
};

const stopPlayback = () => {
  isPlaying.value = false;
  if (audioPlayer.value) {
    audioPlayer.value.pause();
  }
};

const selectAndPlayChunk = async (chunkIndex: number) => {
  if (
    !currentDoc.value ||
    chunkIndex < 0 ||
    chunkIndex >= currentDoc.value.totalChunks
  )
    return;

  // Stop current playback and clean up
  if (audioPlayer.value) {
    audioPlayer.value.pause();
    audioPlayer.value.remove();
  }
  if (nextAudioPlayer.value) {
    nextAudioPlayer.value.remove();
    nextAudioPlayer.value = null;
  }
  isPreloading.value = false;

  // Set new chunk and start playing
  currentDoc.value.currentChunk = chunkIndex;
  isPlaying.value = true;
  await playNextChunk();
};
</script>

<template>
  <div class="container mx-auto">
    <div v-if="!currentDoc" class="space-y-6 p-8 dark:bg-black">
      <FileUpload
        v-model="files"
        class="rounded-lg border border-dashed border-neutral-200 dark:border-neutral-800">
        <FileUploadGrid />
      </FileUpload>

      <div class="flex justify-end">
        <Button
          :disabled="!files.length || isUploading"
          @click="uploadToR2"
          :loading="isUploading">
          Upload
        </Button>
      </div>
    </div>

    <div v-else class="space-y-6 p-8">
      <Card>
        <CardHeader>
          <CardTitle>PDF Processed Successfully</CardTitle>
          <CardDescription>
            {{ currentDoc.totalChunks }} text chunks ready for audio generation
          </CardDescription>
        </CardHeader>

        <CardContent class="space-y-6">
          <div class="space-y-4">
            <div class="flex justify-between items-center">
              <span class="text-sm font-medium text-foreground">Progress</span>
              <span class="text-sm text-muted-foreground">
                {{ currentDoc.currentChunk }} / {{ currentDoc.totalChunks }}
              </span>
            </div>

            <Progress
              :model-value="
                (currentDoc.currentChunk / currentDoc.totalChunks) * 100
              " />
          </div>

          <div class="space-y-2">
            <h3 class="font-medium text-foreground">Text Chunks:</h3>
            <ScrollArea class="h-[240px]">
              <div class="space-y-2 pr-4">
                <div
                  v-for="(chunk, index) in currentDoc.chunks"
                  :key="index"
                  class="p-3 rounded-md border text-sm"
                  :class="[
                    index === currentDoc.currentChunk
                      ? 'bg-primary/10 border-primary/20'
                      : 'bg-muted/50 border-border',
                  ]">
                  <span class="font-medium text-xs text-muted-foreground"
                    >Chunk {{ index + 1 }}:</span
                  >
                  <span class="text-foreground">{{ chunk }}</span>
                </div>
              </div>
            </ScrollArea>
          </div>

          <!-- Chunk Selection -->
          <!-- <div class="flex flex-wrap gap-2">
            <Button
              v-for="index in currentDoc.totalChunks"
              :key="index"
              :variant="
                currentDoc.currentChunk === index - 1 ? 'default' : 'outline'
              "
              @click="selectAndPlayChunk(index - 1)"
              class="text-sm">
              Chunk {{ index }}
            </Button>
          </div> -->

          <!-- Audio Player -->
          <div v-if="audioPlayer" class="space-y-4">
            <!-- Time Progress -->
            <div class="flex justify-between text-sm text-muted-foreground">
              <span>{{ formatTime(currentTime) }}</span>
              <span>{{ formatTime(duration) }}</span>
            </div>

            <!-- Progress Bar -->
            <div
              class="w-full h-2 bg-muted rounded-full cursor-pointer"
              @click="handleProgressClick"
              ref="progressBarRef">
              <div
                class="h-full bg-primary rounded-full transition-all"
                :style="{ width: `${(currentTime / duration) * 100}%` }" />
            </div>

            <!-- Playback Controls -->
            <div class="flex items-center justify-between">
              <!-- Volume Control -->
              <div class="flex items-center gap-2">
                <Button
                  variant="ghost"
                  size="icon"
                  @click="toggleMute"
                  class="h-8 w-8">
                  <LucideVolume2 v-if="volume[0] > 0.5" class="mr-2 h-4 w-4" />
                  <LucideVolume1
                    v-else-if="volume[0] > 0"
                    class="mr-2 h-4 w-4" />
                  <LucideVolumeX v-else class="mr-2 h-4 w-4" />
                </Button>
                <Slider
                  v-model="volume"
                  :min="0"
                  :max="1"
                  :step="0.1"
                  class="w-24" />
              </div>

              <!-- Play/Pause -->
              <div class="flex items-center gap-2">
                <Button
                  variant="outline"
                  size="icon"
                  @click="seek(-10)"
                  class="h-8 w-8">
                  <LucideRewind class="mr-2 h-4 w-4" />
                </Button>

                <Button variant="default" size="icon" @click="togglePlay">
                  <LucideCirclePlay v-if="!isPlaying" class="mr-2 h-4 w-4" />
                  <LucideCirclePause v-else class="mr-2 h-4 w-4" />
                </Button>

                <Button
                  variant="outline"
                  size="icon"
                  @click="seek(10)"
                  class="h-8 w-8">
                  <LucideFastForward class="mr-2 h-4 w-4" />
                </Button>
              </div>

              <!-- Playback Speed -->
              <Select v-model="playbackRate">
                <SelectTrigger class="w-24">
                  <SelectValue>{{ playbackRate }}x</SelectValue>
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="0.5">0.5x</SelectItem>
                  <SelectItem value="0.75">0.75x</SelectItem>
                  <SelectItem value="1">1x</SelectItem>
                  <SelectItem value="1.25">1.25x</SelectItem>
                  <SelectItem value="1.5">1.5x</SelectItem>
                  <SelectItem value="2">2x</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
        </CardContent>

        <CardFooter class="flex gap-3">
          <Button @click="startAudioPlayback" :disabled="isPlaying">
            <LucidePlayCircle v-if="!isPlaying" class="mr-2 h-4 w-4" />
            <LucideLoader2 v-else class="mr-2 h-4 w-4 animate-spin" />
            {{ isPlaying ? "Playing..." : "Start Audio Playback" }}
          </Button>
          <Button variant="outline" @click="resetUpload">
            <LucideUpload class="mr-2 h-4 w-4" />
            Upload Another PDF
          </Button>
        </CardFooter>
      </Card>
    </div>
  </div>
</template>
