<template>
  <div class="relative h-96 overflow-hidden bg-background border rounded-lg">
    <div 
      ref="scrollContainer"
      class="absolute inset-0 p-8"
      :style="{ transform: `translateY(${scrollY}px)` }"
    >
      <div class="space-y-6 text-center">
        <div
          v-for="(chunk, index) in chunks"
          :key="chunk.id"
          :ref="el => chunkRefs[index] = el"
          class="transition-all duration-500 px-4"
          :class="[
            index === currentChunkIndex 
              ? 'text-primary text-lg font-medium scale-105' 
              : 'text-muted-foreground text-base',
            index < currentChunkIndex ? 'opacity-60' : '',
            index > currentChunkIndex + 2 ? 'opacity-40' : ''
          ]"
        >
          <p class="leading-relaxed max-w-2xl mx-auto">
            {{ chunk.content }}
          </p>
          
          <div 
            v-if="index === currentChunkIndex" 
            class="flex items-center justify-center gap-2 mt-3"
          >
            <div 
              v-if="chunk.has_audio"
              class="w-2 h-2 bg-green-500 rounded-full animate-pulse"
              title="Audio available"
            />
            <div 
              v-else-if="isGeneratingAudio(chunk.id)"
              class="w-2 h-2 bg-yellow-500 rounded-full animate-pulse"
              title="Generating audio..."
            />
            <div 
              v-else
              class="w-2 h-2 bg-gray-400 rounded-full"
              title="No audio"
            />
          </div>
        </div>
        
        <div class="h-32"></div>
      </div>
    </div>
    
    <div class="absolute inset-x-0 top-0 h-16 bg-gradient-to-b from-background to-transparent pointer-events-none" />
    <div class="absolute inset-x-0 bottom-0 h-16 bg-gradient-to-t from-background to-transparent pointer-events-none" />
    
    <div class="absolute top-4 right-4 text-sm text-muted-foreground bg-background/80 px-2 py-1 rounded">
      {{ currentChunkIndex + 1 }} / {{ chunks.length }}
    </div>
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
  isPlaying: boolean
  generatingAudioIds?: number[]
}

const props = withDefaults(defineProps<Props>(), {
  generatingAudioIds: () => []
})

const scrollContainer = ref<HTMLElement>()
const chunkRefs = ref<(HTMLElement | null)[]>([])
const scrollY = ref(0)

const isGeneratingAudio = (chunkId: number) => {
  return props.generatingAudioIds.includes(chunkId)
}

const scrollToChunk = (index: number) => {
  if (!scrollContainer.value || !chunkRefs.value[index]) return
  
  const containerHeight = scrollContainer.value.parentElement?.clientHeight || 0
  const chunkElement = chunkRefs.value[index]
  
  if (chunkElement) {
    const chunkTop = chunkElement.offsetTop
    const chunkHeight = chunkElement.offsetHeight
    
    const targetY = -(chunkTop - (containerHeight / 2) + (chunkHeight / 2))
    
    scrollY.value = targetY
  }
}

const autoScroll = () => {
  if (props.isPlaying && props.chunks.length > 0) {
    scrollToChunk(props.currentChunkIndex)
  }
}

watch(() => props.currentChunkIndex, (newIndex) => {
  scrollToChunk(newIndex)
}, { immediate: true })

watch(() => props.isPlaying, () => {
  autoScroll()
})

onMounted(() => {
  if (props.chunks.length > 0) {
    scrollToChunk(props.currentChunkIndex)
  }
})
</script>

<style scoped>
.scroll-container {
  scroll-behavior: smooth;
}
</style>