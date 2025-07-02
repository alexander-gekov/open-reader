<template>
  <div class="bg-card border rounded-lg p-4 shadow-sm">
    <div class="flex items-center gap-4">
      <Button
        :variant="isPlaying ? 'secondary' : 'default'"
        size="sm"
        @click="togglePlayback"
        :disabled="!currentAudioUrl || isLoading"
      >
        <Icon v-if="isLoading" name="lucide:loader-2" class="w-4 h-4 animate-spin" />
        <Icon v-else-if="isPlaying" name="lucide:pause" class="w-4 h-4" />
        <Icon v-else name="lucide:play" class="w-4 h-4" />
      </Button>
      
      <div class="flex-1 space-y-2">
        <div class="flex items-center justify-between text-sm text-muted-foreground">
          <span>Chunk {{ currentChunkIndex + 1 }} of {{ totalChunks }}</span>
          <span v-if="currentAudioUrl">{{ formatTime(currentTime) }} / {{ formatTime(duration) }}</span>
        </div>
        
        <div class="w-full bg-secondary rounded-full h-2">
          <div 
            class="bg-primary h-2 rounded-full transition-all duration-200"
            :style="{ width: `${progress}%` }"
          />
        </div>
      </div>
      
      <div class="flex gap-2">
        <Button
          variant="outline"
          size="sm"
          @click="previousChunk"
          :disabled="currentChunkIndex === 0 || isLoading"
        >
          <Icon name="lucide:skip-back" class="w-4 h-4" />
        </Button>
        
        <Button
          variant="outline"
          size="sm"
          @click="nextChunk"
          :disabled="currentChunkIndex >= totalChunks - 1 || isLoading"
        >
          <Icon name="lucide:skip-forward" class="w-4 h-4" />
        </Button>
      </div>
    </div>
    
    <audio
      ref="audioElement"
      @loadedmetadata="onLoadedMetadata"
      @timeupdate="onTimeUpdate"
      @ended="onAudioEnded"
      @canplaythrough="onCanPlayThrough"
      preload="metadata"
    />
  </div>
</template>

<script setup lang="ts">
interface Props {
  chunks: Array<{
    id: number
    chunk_index: number
    content: string
    has_audio: boolean
  }>
  currentChunkIndex: number
}

interface Emits {
  (e: 'update:currentChunkIndex', index: number): void
  (e: 'chunkChanged', chunkId: number): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const { generateTts, getTtsStatus } = useApi()

const audioElement = ref<HTMLAudioElement>()
const isPlaying = ref(false)
const isLoading = ref(false)
const currentTime = ref(0)
const duration = ref(0)
const currentAudioUrl = ref<string | null>(null)

const totalChunks = computed(() => props.chunks.length)
const currentChunk = computed(() => props.chunks[props.currentChunkIndex])
const progress = computed(() => duration.value ? (currentTime.value / duration.value) * 100 : 0)

const formatTime = (seconds: number) => {
  const mins = Math.floor(seconds / 60)
  const secs = Math.floor(seconds % 60)
  return `${mins}:${secs.toString().padStart(2, '0')}`
}

const loadAudioForChunk = async (chunkId: number) => {
  isLoading.value = true
  
  try {
    const { data: statusData, error } = await getTtsStatus(chunkId)
    
    if (error) {
      console.error('Error getting TTS status:', error)
      return
    }
    
    if (statusData?.status === 'completed' && statusData.audio_url) {
      currentAudioUrl.value = statusData.audio_url
      if (audioElement.value) {
        audioElement.value.src = statusData.audio_url
        audioElement.value.load()
      }
    } else if (statusData?.status === 'not_generated') {
      await generateTts(chunkId)
      
      const checkStatus = async () => {
        const { data: newStatusData } = await getTtsStatus(chunkId)
        
        if (newStatusData?.status === 'completed' && newStatusData.audio_url) {
          currentAudioUrl.value = newStatusData.audio_url
          if (audioElement.value) {
            audioElement.value.src = newStatusData.audio_url
            audioElement.value.load()
          }
        } else if (newStatusData?.status === 'processing') {
          setTimeout(checkStatus, 2000)
        }
      }
      
      setTimeout(checkStatus, 1000)
    }
  } finally {
    isLoading.value = false
  }
}

const togglePlayback = async () => {
  if (!audioElement.value || !currentAudioUrl.value) return
  
  if (isPlaying.value) {
    audioElement.value.pause()
  } else {
    try {
      await audioElement.value.play()
    } catch (error) {
      console.error('Error playing audio:', error)
    }
  }
}

const previousChunk = () => {
  if (props.currentChunkIndex > 0) {
    const newIndex = props.currentChunkIndex - 1
    emit('update:currentChunkIndex', newIndex)
    emit('chunkChanged', props.chunks[newIndex].id)
  }
}

const nextChunk = () => {
  if (props.currentChunkIndex < totalChunks.value - 1) {
    const newIndex = props.currentChunkIndex + 1
    emit('update:currentChunkIndex', newIndex)
    emit('chunkChanged', props.chunks[newIndex].id)
  }
}

const onLoadedMetadata = () => {
  if (audioElement.value) {
    duration.value = audioElement.value.duration
  }
}

const onTimeUpdate = () => {
  if (audioElement.value) {
    currentTime.value = audioElement.value.currentTime
  }
}

const onAudioEnded = () => {
  isPlaying.value = false
  nextChunk()
}

const onCanPlayThrough = () => {
  isLoading.value = false
}

watch(() => props.currentChunkIndex, (newIndex) => {
  if (props.chunks[newIndex]) {
    isPlaying.value = false
    currentTime.value = 0
    loadAudioForChunk(props.chunks[newIndex].id)
  }
}, { immediate: true })

watch(audioElement, (newElement) => {
  if (newElement) {
    newElement.addEventListener('play', () => {
      isPlaying.value = true
    })
    
    newElement.addEventListener('pause', () => {
      isPlaying.value = false
    })
  }
})
</script>