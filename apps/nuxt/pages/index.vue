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
  LucideFastForward,
  LucideLoader2,
  LucidePauseCircle,
  LucidePlayCircle,
  LucideRewind,
  LucideUpload,
  LucideUploadCloud,
  LucideVolume1,
  LucideVolume2,
  LucideVolumeX,
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

const { data: ttsSettings } = await useFetch("/api/settings");

const showSettingsWarning = computed(() => {
  return (
    !ttsSettings.value?.provider ||
    (ttsSettings.value?.provider !== "fallback" && !ttsSettings.value?.apiKey)
  );
});

const calculateCredits = (text: string) => {
  return text.length;
};

const calculateCost = (credits: number) => {
  return (credits * 0.001).toFixed(2);
};

const totalCredits = computed(() => {
  if (!currentDoc.value?.chunks) return 0;
  return currentDoc.value.chunks.reduce(
    (acc, chunk) => acc + calculateCredits(chunk),
    0
  );
});

const scrollToCurrentChunk = () => {
  const currentChunkElement = document.querySelector(
    `[data-chunk-index="${currentDoc.value?.currentChunk}"]`
  );
  if (currentChunkElement) {
    currentChunkElement.scrollIntoView({ behavior: "smooth", block: "center" });
  }
};

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
    currentDoc.value.currentChunk >= currentDoc.value.totalChunks - 1 ||
    isPreloading.value
  )
    return;

  try {
    isPreloading.value = true;
    const nextChunkIndex = currentDoc.value.currentChunk + 1;

    // Trigger the processing of next chunk immediately
    try {
      await $fetch(`/api/audio/start-next/${currentDoc.value.currentChunk}`);
    } catch (err) {
      console.error("Failed to trigger next chunk processing:", err);
    }

    // Start polling for the next chunk with longer intervals
    while (isPreloading.value) {
      const response = await $fetch(`/api/audio/status/${nextChunkIndex}`);

      if (response.status === "ready" && response.url) {
        const audioUrl = response.url;

        // Create and setup the next audio player
        nextAudioPlayer.value = new Audio();
        nextAudioPlayer.value.preload = "auto";

        // Create a promise that resolves when enough data is loaded
        const canPlayPromise = new Promise((resolve, reject) => {
          nextAudioPlayer.value?.addEventListener("canplaythrough", resolve, {
            once: true,
          });
          nextAudioPlayer.value?.addEventListener(
            "error",
            (e) => {
              console.error("Error loading audio:", e);
              reject(new Error("Failed to load audio"));
            },
            { once: true }
          );
        });

        // Set properties and start loading
        nextAudioPlayer.value.src = audioUrl;
        nextAudioPlayer.value.playbackRate = playbackRate.value;
        nextAudioPlayer.value.volume = volume.value[0];

        try {
          // Wait for enough data to be loaded
          await canPlayPromise;
          isPreloading.value = false;
          return;
        } catch (err) {
          console.error("Error during preload:", err);
          nextAudioPlayer.value = null;
          // Wait before trying again
          await new Promise((resolve) => setTimeout(resolve, 2000));
          continue;
        }
      }

      if (response.status === "error") {
        console.error("Error preloading next chunk:", response.error);
        isPreloading.value = false;
        return;
      }

      // Wait longer before polling again
      await new Promise((resolve) => setTimeout(resolve, 2000));
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
      currentDoc.value.currentChunk < currentDoc.value.totalChunks - 1
    ) {
      currentDoc.value.currentChunk++;
      isPlaying.value = true;
      scrollToCurrentChunk();
      await playNextChunk();
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
    // Start processing the next chunk immediately when current starts playing
    if (
      currentDoc.value &&
      currentDoc.value.currentChunk < currentDoc.value.totalChunks - 1
    ) {
      console.log("Triggering next chunk processing");
      $fetch(`/api/audio/start-next/${currentDoc.value.currentChunk}`).catch(
        (err) => {
          console.error("Failed to trigger next chunk processing:", err);
        }
      );
    }
  });

  audioPlayer.value.addEventListener("pause", () => {
    isPlaying.value = false;
  });

  audioPlayer.value.addEventListener("timeupdate", () => {
    currentTime.value = audioPlayer.value?.currentTime || 0;

    // When we're 50% through the current chunk, ensure next chunks are ready
    if (
      audioPlayer.value &&
      currentDoc.value &&
      currentDoc.value.currentChunk < currentDoc.value.totalChunks - 1 &&
      currentTime.value / duration.value > 0.5
    ) {
      preloadNextChunk();
    }
  });

  audioPlayer.value.addEventListener("loadedmetadata", () => {
    duration.value = audioPlayer.value?.duration || 0;
  });
};

const playNextChunk = async () => {
  if (!currentDoc.value) return;

  try {
    console.log("Checking audio status...");
    const response = await $fetch(
      `/api/audio/status/${currentDoc.value.currentChunk}`
    );
    console.log("Audio status:", response);

    if (response.status === "ready" && response.url) {
      const audioUrl = response.url;
      console.log("Playing audio from URL:", audioUrl);

      if (!audioPlayer.value) {
        audioPlayer.value = new Audio();
        setupAudioPlayer();
      }

      // Set up event listeners before setting src
      audioPlayer.value.preload = "auto";
      const canPlayPromise = new Promise((resolve, reject) => {
        audioPlayer.value?.addEventListener("canplaythrough", resolve, {
          once: true,
        });
        audioPlayer.value?.addEventListener(
          "error",
          (e) => {
            console.error("Error loading audio:", e);
            reject(new Error("Failed to load audio"));
          },
          { once: true }
        );
      });

      // If this is a new chunk, set it up for playback
      if (audioPlayer.value.src !== audioUrl) {
        audioPlayer.value.src = audioUrl;
        audioPlayer.value.playbackRate = playbackRate.value;
        audioPlayer.value.volume = volume.value[0];

        try {
          // Wait for enough data to be loaded
          await canPlayPromise;
        } catch (err) {
          console.error("Error loading audio:", err);
          error.value = "Failed to load audio";
          isPlaying.value = false;
          return;
        }
      }

      // Start playing
      try {
        const playPromise = audioPlayer.value.play();
        if (playPromise) {
          await playPromise;
        }
      } catch (error) {
        console.error("Error starting playback:", error);
        isPlaying.value = false;
      }
      return;
    }

    if (response.status === "error") {
      console.error("Error generating audio:", response.error);
      error.value = response.error || "Failed to generate audio";
      isPlaying.value = false;
      return;
    }

    // If still processing, wait longer before checking again
    await new Promise((resolve) => setTimeout(resolve, 2000));
    if (isPlaying.value) {
      // Only continue polling if we're still supposed to be playing
      await playNextChunk();
    }
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
const handleFileUpload = async (file: File) => {
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

    console.log("TTS Settings:", ttsSettings.value); // Debug log

    const headers = {
      "X-TTS-Provider": ttsSettings.value?.provider || "elevenlabs",
      "X-TTS-API-Key": ttsSettings.value?.apiKey || "",
      "X-TTS-Model": ttsSettings.value?.model || "",
      "X-TTS-Voice": ttsSettings.value?.voice || "",
    };

    console.log("Request Headers:", headers); // Debug log

    const response = await $fetch<UploadResponse>("/api/upload", {
      method: "POST",
      body: formData,
      headers,
    });

    currentDoc.value = {
      audioId: response.audioId,
      chunks: response.chunks,
      totalChunks: response.totalChunks,
      currentChunk: 0,
    };
  } catch (error: any) {
    console.error("Error uploading file:", error);
    const errorMessage =
      error.data?.message ||
      error.message ||
      "Failed to upload and process the PDF";
    alert(errorMessage);
  } finally {
    isUploading.value = false;
  }
};

// Triggered by the Upload button
const uploadToR2 = async () => {
  if (!files.value.length) return;

  // Currently we only support uploading the first selected PDF
  await handleFileUpload(files.value[0]);

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

  // Set new chunk
  currentDoc.value.currentChunk = chunkIndex;
  scrollToCurrentChunk();

  // Trigger generation of this chunk and the next one
  try {
    await $fetch(`/api/audio/start-next/${chunkIndex - 1}`);
  } catch (err) {
    console.error("Failed to trigger chunk processing:", err);
  }

  // Start playing
  isPlaying.value = true;
  await playNextChunk();
};
</script>

<template>
  <div class="container mx-auto">
    <div v-if="showSettingsWarning" class="p-8">
      <Card>
        <CardHeader>
          <CardTitle>TTS Settings Required</CardTitle>
          <CardDescription>
            Please configure your Text-to-Speech settings before uploading files
          </CardDescription>
        </CardHeader>
        <CardContent>
          <p class="text-muted-foreground">
            You need to set up your preferred TTS provider and API key to use
            this feature.
          </p>
        </CardContent>
        <CardFooter>
          <NuxtLink to="/settings">
            <Button>Configure Settings</Button>
          </NuxtLink>
        </CardFooter>
      </Card>
    </div>

    <div v-else-if="!currentDoc" class="space-y-6 p-8 dark:bg-black">
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
            <div class="flex justify-between items-center">
              <h3 class="font-medium text-foreground">Text Chunks:</h3>
              <div class="text-sm text-muted-foreground">
                Total Credits: {{ totalCredits }} (~{{
                  calculateCost(totalCredits)
                }}¢)
              </div>
            </div>
            <ScrollArea class="h-[240px]">
              <div class="space-y-2 pr-4">
                <div
                  v-for="(chunk, index) in currentDoc.chunks"
                  :key="index"
                  :data-chunk-index="index"
                  class="p-3 rounded-md border text-sm cursor-pointer hover:bg-primary/5 transition-colors"
                  :class="[
                    index === currentDoc.currentChunk
                      ? 'bg-primary/10 border-primary/20'
                      : 'bg-muted/50 border-border',
                  ]"
                  @click="selectAndPlayChunk(index)">
                  <span class="font-medium text-xs text-muted-foreground"
                    >Chunk {{ index + 1 }}:</span
                  >
                  <span
                    class="text-foreground"
                    :class="{
                      'text-primary font-medium':
                        index === currentDoc.currentChunk,
                    }"
                    >{{ chunk }}</span
                  >
                  <div class="text-xs text-muted-foreground mt-1">
                    Credits: {{ calculateCredits(chunk) }} (~{{
                      calculateCost(calculateCredits(chunk))
                    }}¢)
                  </div>
                </div>
              </div>
            </ScrollArea>
          </div>

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
                  <LucideVolume2 v-if="volume[0] > 0.5" class="h-4 w-4" />
                  <LucideVolume1 v-else-if="volume[0] > 0" class="h-4 w-4" />
                  <LucideVolumeX v-else class="h-4 w-4" />
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
                  <LucideRewind class="h-4 w-4" />
                </Button>

                <Button variant="default" size="icon" @click="togglePlay">
                  <LucidePlayCircle
                    v-if="!isPlaying"
                    class="h-4 w-4 text-background" />
                  <LucidePauseCircle v-else class="h-4 w-4" />
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
            <LucidePlayCircle
              v-if="!isPlaying"
              class="h-4 w-4 text-background" />
            <LucideLoader2 v-else class="h-4 w-4 animate-spin" />
            {{ isPlaying ? "Playing..." : "Start Audio Playback" }}
          </Button>
          <Button variant="outline" @click="resetUpload">
            <LucideUploadCloud class="h-4 w-4" />
            Upload Another PDF
          </Button>
        </CardFooter>
      </Card>
    </div>
  </div>
</template>
