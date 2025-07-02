<template>
  <div class="min-h-screen bg-background">
    <div class="container mx-auto px-4 py-8">
      <div v-if="loading" class="flex items-center justify-center h-64">
        <div class="text-center space-y-4">
          <Icon name="lucide:loader-2" class="w-8 h-8 animate-spin mx-auto" />
          <p class="text-muted-foreground">Loading document...</p>
        </div>
      </div>
      
      <div v-else-if="error" class="text-center py-16">
        <Alert variant="destructive" class="max-w-md mx-auto">
          <Icon name="lucide:alert-circle" class="h-4 w-4" />
          <AlertTitle>Error</AlertTitle>
          <AlertDescription>{{ error }}</AlertDescription>
        </Alert>
        
        <Button @click="$router.back()" variant="outline" class="mt-4">
          <Icon name="lucide:arrow-left" class="w-4 h-4 mr-2" />
          Go Back
        </Button>
      </div>
      
      <div v-else-if="document" class="space-y-6">
        <div class="flex items-center justify-between">
          <div class="space-y-1">
            <h1 class="text-2xl font-bold">{{ document.filename }}</h1>
            <p class="text-sm text-muted-foreground">
              {{ chunks.length }} text chunks
              <span v-if="document.status === 'processing'" class="text-yellow-600">
                • Processing...
              </span>
              <span v-else-if="document.status === 'completed'" class="text-green-600">
                • Ready
              </span>
              <span v-else-if="document.status === 'error'" class="text-red-600">
                • Error
              </span>
            </p>
          </div>
          
          <Button @click="$router.back()" variant="outline">
            <Icon name="lucide:arrow-left" class="w-4 h-4 mr-2" />
            Back
          </Button>
        </div>
        
        <div v-if="document.status === 'uploaded'" class="text-center py-8">
          <div class="space-y-4">
            <p class="text-muted-foreground">This document hasn't been processed yet.</p>
            <Button @click="processPdfDocument" :disabled="isProcessing">
              <Icon v-if="isProcessing" name="lucide:loader-2" class="w-4 h-4 mr-2 animate-spin" />
              <Icon v-else name="lucide:play" class="w-4 h-4 mr-2" />
              {{ isProcessing ? 'Processing...' : 'Process Document' }}
            </Button>
          </div>
        </div>
        
        <div v-else-if="document.status === 'processing'" class="text-center py-8">
          <div class="space-y-4">
            <Icon name="lucide:loader-2" class="w-8 h-8 animate-spin mx-auto" />
            <p class="text-muted-foreground">Processing document and extracting text...</p>
            <Button @click="checkStatus" variant="outline" size="sm">
              <Icon name="lucide:refresh-cw" class="w-4 h-4 mr-2" />
              Refresh Status
            </Button>
          </div>
        </div>
        
        <div v-else-if="chunks.length > 0" class="grid gap-6">
          <ScrollingText 
            :chunks="chunks"
            :current-chunk-index="currentChunkIndex"
            :is-playing="false"
            :generating-audio-ids="generatingAudioIds"
          />
          
          <AudioPlayer
            :chunks="chunks"
            v-model:current-chunk-index="currentChunkIndex"
            @chunk-changed="onChunkChanged"
          />
          
          <div class="flex items-center justify-between text-sm text-muted-foreground">
            <span>{{ audioChunksCount }} of {{ chunks.length }} chunks have audio</span>
            <Button 
              @click="generateAllAudio" 
              variant="outline" 
              size="sm"
              :disabled="isGeneratingBatch || audioChunksCount === chunks.length"
            >
              <Icon v-if="isGeneratingBatch" name="lucide:loader-2" class="w-4 h-4 mr-2 animate-spin" />
              <Icon v-else name="lucide:volume-2" class="w-4 h-4 mr-2" />
              Generate All Audio
            </Button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const route = useRoute()
const router = useRouter()
const { 
  getDocuments, 
  getDocumentChunks, 
  getProcessingStatus, 
  processPdf, 
  generateBatchTts 
} = useApi()

const documentId = computed(() => parseInt(route.params.id as string))

const loading = ref(true)
const error = ref<string | null>(null)
const document = ref<any>(null)
const chunks = ref<any[]>([])
const currentChunkIndex = ref(0)
const isProcessing = ref(false)
const isGeneratingBatch = ref(false)
const generatingAudioIds = ref<number[]>([])

const audioChunksCount = computed(() => 
  chunks.value.filter(chunk => chunk.has_audio).length
)

const loadDocument = async () => {
  try {
    loading.value = true
    error.value = null
    
    const { data: documents, error: docsError } = await getDocuments()
    
    if (docsError) {
      error.value = docsError
      return
    }
    
    document.value = documents?.find((doc: any) => doc.id === documentId.value)
    
    if (!document.value) {
      error.value = 'Document not found'
      return
    }
    
    if (document.value.status === 'completed') {
      await loadChunks()
    }
    
  } catch (err: any) {
    error.value = err.message || 'Failed to load document'
  } finally {
    loading.value = false
  }
}

const loadChunks = async () => {
  try {
    const { data: chunksData, error: chunksError } = await getDocumentChunks(documentId.value)
    
    if (chunksError) {
      error.value = chunksError
      return
    }
    
    chunks.value = chunksData || []
  } catch (err: any) {
    error.value = err.message || 'Failed to load text chunks'
  }
}

const processPdfDocument = async () => {
  try {
    isProcessing.value = true
    
    const { data, error: processError } = await processPdf(documentId.value)
    
    if (processError) {
      error.value = processError
      return
    }
    
    document.value.status = 'processing'
    
    const checkProcessingStatus = async () => {
      const { data: statusData } = await getProcessingStatus(documentId.value)
      
      if (statusData?.status === 'completed') {
        document.value.status = 'completed'
        await loadChunks()
        isProcessing.value = false
      } else if (statusData?.status === 'error') {
        document.value.status = 'error'
        error.value = 'Processing failed'
        isProcessing.value = false
      } else {
        setTimeout(checkProcessingStatus, 2000)
      }
    }
    
    setTimeout(checkProcessingStatus, 1000)
    
  } catch (err: any) {
    error.value = err.message || 'Failed to process document'
    isProcessing.value = false
  }
}

const checkStatus = async () => {
  try {
    const { data: statusData } = await getProcessingStatus(documentId.value)
    
    if (statusData) {
      document.value.status = statusData.status
      
      if (statusData.status === 'completed') {
        await loadChunks()
      }
    }
  } catch (err: any) {
    console.error('Error checking status:', err)
  }
}

const generateAllAudio = async () => {
  try {
    isGeneratingBatch.value = true
    
    const chunkIds = chunks.value
      .filter(chunk => !chunk.has_audio)
      .map(chunk => chunk.id)
    
    if (chunkIds.length === 0) return
    
    generatingAudioIds.value = [...chunkIds]
    
    const { error: batchError } = await generateBatchTts(chunkIds)
    
    if (batchError) {
      error.value = batchError
      return
    }
    
    const checkBatchStatus = async () => {
      let allCompleted = true
      
      for (const chunk of chunks.value) {
        if (generatingAudioIds.value.includes(chunk.id)) {
          chunk.has_audio = true
        }
      }
      
      if (allCompleted) {
        generatingAudioIds.value = []
        isGeneratingBatch.value = false
      } else {
        setTimeout(checkBatchStatus, 3000)
      }
    }
    
    setTimeout(checkBatchStatus, 2000)
    
  } catch (err: any) {
    error.value = err.message || 'Failed to generate audio'
    isGeneratingBatch.value = false
    generatingAudioIds.value = []
  }
}

const onChunkChanged = (chunkId: number) => {
  const index = chunks.value.findIndex(chunk => chunk.id === chunkId)
  if (index !== -1) {
    currentChunkIndex.value = index
  }
}

onMounted(() => {
  loadDocument()
})

watch(() => route.params.id, () => {
  if (route.params.id) {
    loadDocument()
  }
})
</script>