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
  LucideArrowRight,
  LucideCheck,
  LucideGithub,
  LucideTwitter,
} from "lucide-vue-next";
import { Slider } from "@/components/ui/slider";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Input } from "@/components/ui/input";
import { FlipWords } from "@/components/ui/flip-words";

definePageMeta({
  title: "Open Reader - Open Source PDF & Audiobook Reader",
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

const email = ref("");
const name = ref("");
const isSubmitting = ref(false);
const isSuccess = ref(false);

type TTSProvider = {
  provider: string;
  apiKey: string;
  model: string;
  voice: string;
};

const { data: ttsSettings } = await useFetch<TTSProvider>("/api/settings");

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

const handleSubmit = async () => {
  if (!email.value) return;

  try {
    isSubmitting.value = true;
    await $fetch("/api/waitlist", {
      method: "POST",
      body: {
        email: email.value,
        name: name.value,
      },
    });
    isSuccess.value = true;
  } catch (error: any) {
    console.error("Error joining waitlist:", error);
    alert(error.data?.message || "Failed to join waitlist");
  } finally {
    isSubmitting.value = false;
  }
};
</script>

<template>
  <div class="container mx-auto px-4 py-24 space-y-24">
    <!-- Hero Section -->
    <div class="text-center space-y-6">
      <div
        class="flex items-center justify-center gap-2 text-sm text-muted-foreground mb-8">
        <span class="px-3 py-1 rounded-full bg-muted">Open Source</span>
        <span class="px-3 py-1 rounded-full bg-muted">Self-Hostable</span>
        <span class="px-3 py-1 rounded-full bg-muted">Privacy-First</span>
      </div>
      <h1 class="text-4xl sm:text-5xl md:text-6xl font-bold tracking-loose">
        Your
        <span class="text-primary inline-flex">
          <FlipWords
            :words="[
              'Open Source',
              'Privacy-First',
              'Self-Hostable',
              'Multi-Provider',
              'Budget-Friendly',
            ]"
            :duration="2000" />
        </span>
        Alternative to ElevenReader
      </h1>
      <p class="text-xl text-muted-foreground max-w-2xl mx-auto">
        Convert PDFs and books into natural-sounding audio using your preferred
        AI voice provider. Self-host for privacy or use our cloud solution -
        you're in control.
      </p>
      <div class="flex justify-center gap-4">
        <Button
          size="lg"
          @click="
            $el
              .querySelector('#waitlist-form')
              .scrollIntoView({ behavior: 'smooth' })
          ">
          Join Waitlist
          <LucideArrowRight class="ml-2 h-4 w-4" />
        </Button>
        <a
          href="https://github.com/alexander-gekov/open-reader"
          target="_blank"
          rel="noopener">
          <Button size="lg" variant="outline">
            <LucideGithub class="mr-2 h-4 w-4" />
            Star on GitHub
          </Button>
        </a>
      </div>
    </div>

    <!-- Features Section -->
    <div class="grid md:grid-cols-3 gap-8">
      <Card>
        <CardContent class="pt-6">
          <div class="space-y-2">
            <h3 class="text-xl font-semibold">Multiple Voice Providers</h3>
            <p class="text-muted-foreground">
              Use ElevenLabs, AWS Polly, Cartesia/Sonic, or other providers.
              Bring your own API key or use our cloud service.
            </p>
          </div>
        </CardContent>
      </Card>
      <Card>
        <CardContent class="pt-6">
          <div class="space-y-2">
            <h3 class="text-xl font-semibold">PDF & eBook Support</h3>
            <p class="text-muted-foreground">
              Convert PDFs and eBooks with smart text extraction and formatting
              preservation.
            </p>
          </div>
        </CardContent>
      </Card>
      <Card>
        <CardContent class="pt-6">
          <div class="space-y-2">
            <h3 class="text-xl font-semibold">Self-Hostable</h3>
            <p class="text-muted-foreground">
              Deploy on your own infrastructure for complete privacy and
              control.
            </p>
          </div>
        </CardContent>
      </Card>
    </div>

    <!-- Waitlist Section -->
    <div id="waitlist-form" class="max-w-xl mx-auto text-center space-y-8">
      <div class="space-y-4">
        <h2 class="text-3xl font-bold">Get Early Access</h2>
        <p class="text-muted-foreground">
          Join the waitlist to be among the first to experience Open Reader and
          receive:
        </p>
        <ul class="text-left space-y-2 mt-4">
          <li class="flex items-center gap-2">
            <LucideCheck class="h-4 w-4 text-primary" />
            <span>Early access to the platform</span>
          </li>
          <li class="flex items-center gap-2">
            <LucideCheck class="h-4 w-4 text-primary" />
            <span>Special launch pricing</span>
          </li>
          <li class="flex items-center gap-2">
            <LucideCheck class="h-4 w-4 text-primary" />
            <span>Priority support and feature requests</span>
          </li>
        </ul>
      </div>

      <Card v-if="!isSuccess" class="border-primary/50">
        <CardContent class="pt-6">
          <form @submit.prevent="handleSubmit" class="space-y-4">
            <div class="space-y-2">
              <Input
                v-model="email"
                type="email"
                placeholder="Enter your email"
                required />
            </div>
            <div class="space-y-2">
              <Input
                v-model="name"
                type="text"
                placeholder="Your name (optional)" />
            </div>
            <Button type="submit" class="w-full" :loading="isSubmitting">
              Join Waitlist
            </Button>
          </form>
        </CardContent>
      </Card>

      <Card v-else>
        <CardContent class="pt-6">
          <div class="flex items-center justify-center gap-2 text-primary">
            <LucideCheck class="h-5 w-5" />
            <span class="font-medium">You're on the list!</span>
          </div>
          <p class="mt-2 text-muted-foreground">
            We'll notify you when Open Reader launches. Thank you for your
            interest!
          </p>
        </CardContent>
      </Card>
    </div>

    <!-- Footer -->
    <footer class="border-t pt-8">
      <div class="flex justify-between items-center">
        <p class="text-sm text-muted-foreground">
          © 2025 Open Reader. All rights reserved.
        </p>
        <div class="flex items-center gap-4">
          <a
            href="https://x.com/AlexanderGekov"
            target="_blank"
            rel="noopener"
            class="text-muted-foreground hover:text-primary transition-colors">
            <LucideTwitter class="h-5 w-5" />
          </a>
          <a
            href="https://github.com/alexander-gekov"
            target="_blank"
            rel="noopener"
            class="text-muted-foreground hover:text-primary transition-colors">
            <LucideGithub class="h-5 w-5" />
          </a>
        </div>
      </div>
    </footer>
  </div>
</template>
