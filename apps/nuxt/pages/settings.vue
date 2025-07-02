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
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

definePageMeta({
  title: "Settings",
  layout: "default",
  middleware: ["auth"],
});

import { toast } from "vue-sonner";
import {
  LucideCirclePlay,
  LucideFastForward,
  LucideRewind,
  LucideVolume1,
} from "lucide-vue-next";

interface Provider {
  id: string;
  name: string;
  description: string;
  models: Array<{ id: string; name: string }>;
  voices: Array<{ id: string; name: string }>;
}

interface TTSSettings {
  provider: string;
  apiKey: string;
  model?: string;
  voice?: string;
}

const providers = [
  {
    id: "elevenlabs",
    name: "ElevenLabs",
    description: "High-quality, natural-sounding voices with emotion control",
    models: [{ id: "eleven_flash_v2_5", name: "Flash V2.5" }],
    voices: [{ id: "cgSgspJ2msm6clMCkdW9", name: "Bella" }],
  },
  {
    id: "cartesia",
    name: "Cartesia",
    description: "High-quality dedicated TTS service with fast inference",
    models: [
      { id: "sonic-turbo", name: "Sonic Turbo" },
      { id: "sonic-2", name: "Sonic 2" },
    ],
    voices: [{ id: "694f9389-aac1-45b6-b726-9d9369183238", name: "Default" }],
  },
  {
    id: "together",
    name: "Together AI",
    description: "Fast and efficient TTS with Cartasia/sonic model",
    models: [{ id: "cartasia/sonic", name: "Sonic" }],
    voices: [
      { id: "v2/en_speaker_1", name: "Speaker 1" },
      { id: "v2/en_speaker_2", name: "Speaker 2" },
      { id: "v2/en_speaker_3", name: "Speaker 3" },
    ],
  },
  {
    id: "replicate",
    name: "Replicate",
    description: "Open-source TTS models with Kokoro-82m",
    models: [{ id: "jaaari/kokoro-82m", name: "Kokoro 82M" }],
    voices: [
      { id: "af_bella", name: "Bella" },
      { id: "af_daniel", name: "Daniel" },
    ],
  },
  {
    id: "fallback",
    name: "Fallback (No API Key Required)",
    description: "Local TTS provider that works without an API key",
    models: [],
    voices: [],
  },
];

const settings = ref({
  provider: "",
  apiKey: "",
  model: "",
  voice: "",
});

const isLoading = ref(false);

// Watch provider changes to clear fields when fallback is selected
watch(
  () => settings.value.provider,
  (newProvider) => {
    if (newProvider === "fallback") {
      settings.value.apiKey = "";
      settings.value.model = "";
      settings.value.voice = "";
    }
  }
);

const selectedProvider = computed(() => {
  return providers.find((p) => p.id === settings.value.provider);
});

// Load existing settings
const loadSettings = async () => {
  try {
    const response = await $fetch("/api/settings");
    if (response) {
      settings.value = response;
    }
  } catch (error) {
    console.error("Failed to load settings:", error);
  }
};

// Save settings
const saveSettings = async () => {
  try {
    isLoading.value = true;
    await $fetch("/api/settings", {
      method: "POST",
      body: settings.value,
    });
    await navigateTo("/");
  } catch (error) {
    console.error("Failed to save settings:", error);
  } finally {
    isLoading.value = false;
  }
};

onMounted(() => {
  loadSettings();
});
</script>

<template>
  <div class="container mx-auto py-16">
    <Card class="max-w-2xl mx-auto">
      <CardHeader class="space-y-2">
        <CardTitle>Text-to-Speech Settings</CardTitle>
        <CardDescription>
          Configure your preferred TTS provider and settings
        </CardDescription>
      </CardHeader>

      <CardContent class="p-6">
        <Form @submit="saveSettings" class="space-y-6">
          <FormField name="provider">
            <FormLabel>TTS Provider</FormLabel>
            <FormControl>
              <Select v-model="settings.provider">
                <SelectTrigger>
                  <SelectValue placeholder="Select a provider" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem
                    v-for="provider in providers"
                    :key="provider.id"
                    :value="provider.id">
                    {{ provider.name }}
                  </SelectItem>
                </SelectContent>
              </Select>
            </FormControl>
            <FormDescription v-if="selectedProvider">
              {{ selectedProvider.description }}
            </FormDescription>
          </FormField>

          <FormField name="apiKey" v-if="settings.provider !== 'fallback'">
            <FormLabel>API Key</FormLabel>
            <FormControl>
              <Input
                v-model="settings.apiKey"
                type="password"
                placeholder="Enter your API key" />
            </FormControl>
            <FormDescription>
              Your API key will be securely stored
            </FormDescription>
          </FormField>

          <FormField
            v-if="
              settings.provider !== 'fallback' &&
              (selectedProvider?.models?.length ?? 0) > 0
            "
            name="model">
            <FormLabel>Model</FormLabel>
            <FormControl>
              <Select v-model="settings.model">
                <SelectTrigger>
                  <SelectValue placeholder="Select a model" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem
                    v-for="model in selectedProvider?.models ?? []"
                    :key="model.id"
                    :value="model.id">
                    {{ model.name }}
                  </SelectItem>
                </SelectContent>
              </Select>
            </FormControl>
          </FormField>

          <FormField
            v-if="
              settings.provider !== 'fallback' &&
              (selectedProvider?.voices?.length ?? 0) > 0
            "
            name="voice">
            <FormLabel>Voice</FormLabel>
            <FormControl>
              <Select v-model="settings.voice">
                <SelectTrigger>
                  <SelectValue placeholder="Select a voice" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem
                    v-for="voice in selectedProvider?.voices ?? []"
                    :key="voice.id"
                    :value="voice.id">
                    {{ voice.name }}
                  </SelectItem>
                </SelectContent>
              </Select>
            </FormControl>
          </FormField>

          <Button type="submit" :loading="isLoading">Save Settings</Button>
        </Form>
      </CardContent>
    </Card>
  </div>
</template>
