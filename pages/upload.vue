<template>
  <div class="container mx-auto p-6 max-w-4xl">
    <div class="space-y-8">
      <div>
        <h1 class="text-3xl font-bold mb-2">PDF to Audio Converter</h1>
        <p class="text-muted-foreground">
          Upload PDF documents and convert them to audio for easy listening
        </p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Upload PDF</CardTitle>
          <CardDescription>
            Select PDF files to upload and convert to audio
          </CardDescription>
        </CardHeader>
        <CardContent>
          <FileUpload 
            v-model="selectedFiles" 
            @upload="handleUpload"
            class="mb-4" 
          />
          
          <div v-if="uploadProgress.length > 0" class="space-y-4">
            <h3 class="font-semibold">Upload Progress</h3>
            <div v-for="progress in uploadProgress" :key="progress.fileName" class="space-y-2">
              <div class="flex justify-between text-sm">
                <span>{{ progress.fileName }}</span>
                <span>{{ progress.status }}</span>
              </div>
              <Progress :value="progress.progress" class="w-full" />
            </div>
          </div>
        </CardContent>
      </Card>

      <Card v-if="uploadedDocuments.length > 0">
        <CardHeader>
          <CardTitle>Uploaded Documents</CardTitle>
          <CardDescription>
            Your processed PDF documents
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div class="space-y-4">
            <div 
              v-for="doc in uploadedDocuments" 
              :key="doc.id"
              class="border rounded-lg p-4 space-y-3">
              <div class="flex justify-between items-start">
                <div>
                  <h4 class="font-medium">{{ doc.originalName }}</h4>
                  <p class="text-sm text-muted-foreground">
                    {{ doc.totalPages }} pages • {{ doc.chunksCount }} chunks
                  </p>
                </div>
                <Badge :variant="doc.status === 'COMPLETED' ? 'default' : 'secondary'">
                  {{ doc.status }}
                </Badge>
              </div>
              
              <div v-if="doc.status === 'COMPLETED'" class="space-y-2">
                <h5 class="text-sm font-medium">Audio Chunks</h5>
                <div class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-2">
                  <Button
                    v-for="chunkIndex in doc.chunksCount"
                    :key="chunkIndex"
                    variant="outline"
                    size="sm"
                    @click="playChunk(doc.id, chunkIndex - 1)"
                    :disabled="isLoadingAudio(doc.id, chunkIndex - 1)"
                    class="flex items-center gap-2">
                    <LucidePlay v-if="!isLoadingAudio(doc.id, chunkIndex - 1)" class="w-3 h-3" />
                    <LucideLoader2 v-else class="w-3 h-3 animate-spin" />
                    Chunk {{ chunkIndex }}
                  </Button>
                </div>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      <Card v-if="currentAudio">
        <CardHeader>
          <CardTitle>Audio Player</CardTitle>
        </CardHeader>
        <CardContent>
          <div class="space-y-4">
            <p class="text-sm text-muted-foreground">
              Playing: {{ currentAudio.documentName }} - Chunk {{ currentAudio.chunkIndex + 1 }}
            </p>
            <audio 
              ref="audioPlayer"
              controls 
              class="w-full"
              @ended="onAudioEnded">
              <source :src="currentAudio.audioUrl" type="audio/mpeg">
              Your browser does not support audio playback.
            </audio>
          </div>
        </CardContent>
      </Card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { LucidePlay, LucideLoader2 } from 'lucide-vue-next'

interface UploadProgress {
  fileName: string
  status: string
  progress: number
}

interface UploadedDocument {
  id: string
  fileName: string
  originalName: string
  status: string
  totalPages: number
  chunksCount: number
}

interface CurrentAudio {
  documentName: string
  chunkIndex: number
  audioUrl: string
}

const selectedFiles = ref<File[]>([])
const uploadProgress = ref<UploadProgress[]>([])
const uploadedDocuments = ref<UploadedDocument[]>([])
const currentAudio = ref<CurrentAudio | null>(null)
const audioPlayer = ref<HTMLAudioElement>()
const loadingAudio = ref<Set<string>>(new Set())

function isLoadingAudio(docId: string, chunkIndex: number): boolean {
  return loadingAudio.value.has(`${docId}-${chunkIndex}`)
}

async function handleUpload(files: File[]) {
  uploadProgress.value = files.map(file => ({
    fileName: file.name,
    status: 'Uploading...',
    progress: 0
  }))

  for (let i = 0; i < files.length; i++) {
    const file = files[i]
    const progressItem = uploadProgress.value[i]
    
    try {
      progressItem.status = 'Processing...'
      progressItem.progress = 50

      const formData = new FormData()
      formData.append('file', file)

      const response = await $fetch('/api/upload', {
        method: 'POST',
        body: formData
      })

      if (response.success) {
        progressItem.status = 'Completed'
        progressItem.progress = 100
        
        uploadedDocuments.value.push(response.document)
      } else {
        throw new Error('Upload failed')
      }
    } catch (error) {
      console.error('Upload error:', error)
      progressItem.status = 'Failed'
      progressItem.progress = 0
    }
  }

  // Clear files after upload
  selectedFiles.value = []
}

async function playChunk(docId: string, chunkIndex: number) {
  const loadingKey = `${docId}-${chunkIndex}`
  loadingAudio.value.add(loadingKey)

  try {
    const response = await $fetch(`/api/audio/${docId}/${chunkIndex}`)
    
    if (response.status === 'completed' && response.audio_url) {
      const doc = uploadedDocuments.value.find(d => d.id === docId)
      currentAudio.value = {
        documentName: doc?.originalName || 'Unknown',
        chunkIndex,
        audioUrl: response.audio_url
      }
      
      // Wait for next tick to ensure audio element is updated
      await nextTick()
      audioPlayer.value?.play()
    } else if (response.status === 'processing') {
      // TODO: Implement polling or WebSocket for real-time updates
      alert('Audio is being generated. Please try again in a moment.')
    }
  } catch (error) {
    console.error('Error playing chunk:', error)
    alert('Failed to load audio. Please try again.')
  } finally {
    loadingAudio.value.delete(loadingKey)
  }
}

function onAudioEnded() {
  // Could implement auto-play next chunk here
  console.log('Audio playback ended')
}

// Set page metadata
definePageMeta({
  title: 'PDF Upload',
  layout: 'default'
})
</script>