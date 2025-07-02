<template>
  <div class="min-h-screen bg-background">
    <div class="container mx-auto px-4 py-8">
      <div class="max-w-4xl mx-auto space-y-8">
        <div class="text-center space-y-4">
          <h1 class="text-4xl font-bold tracking-tight">Open Reader</h1>
          <p class="text-xl text-muted-foreground">
            Upload PDFs and listen to them with AI-generated audio
          </p>
        </div>

        <Card class="p-6">
          <CardHeader>
            <CardTitle>Upload PDF Document</CardTitle>
            <CardDescription>
              Upload a PDF file to convert it to audio using AI text-to-speech
            </CardDescription>
          </CardHeader>
          
          <CardContent class="space-y-4">
            <div
              @drop="onDrop"
              @dragover.prevent
              @dragenter.prevent
              class="border-2 border-dashed border-muted-foreground/25 rounded-lg p-8 text-center hover:border-muted-foreground/50 transition-colors"
              :class="{ 'border-primary bg-primary/5': isDragging }"
            >
              <Icon name="lucide:upload-cloud" class="w-12 h-12 mx-auto mb-4 text-muted-foreground" />
              
              <div class="space-y-2">
                <p class="text-lg font-medium">
                  Drop your PDF here or 
                  <Button
                    variant="link" 
                    class="p-0 h-auto text-lg"
                    @click="triggerFileInput"
                  >
                    browse files
                  </Button>
                </p>
                <p class="text-sm text-muted-foreground">
                  Supports PDF files up to 50MB
                </p>
              </div>
              
              <input
                ref="fileInput"
                type="file"
                accept=".pdf"
                @change="onFileSelect"
                class="hidden"
              />
            </div>
            
            <div v-if="uploadError" class="mt-4">
              <Alert variant="destructive">
                <Icon name="lucide:alert-circle" class="h-4 w-4" />
                <AlertTitle>Upload Error</AlertTitle>
                <AlertDescription>{{ uploadError }}</AlertDescription>
              </Alert>
            </div>
            
            <div v-if="isUploading" class="mt-4">
              <div class="flex items-center gap-2 text-sm text-muted-foreground">
                <Icon name="lucide:loader-2" class="w-4 h-4 animate-spin" />
                Uploading {{ selectedFile?.name }}...
              </div>
              <div class="w-full bg-secondary rounded-full h-2 mt-2">
                <div 
                  class="bg-primary h-2 rounded-full transition-all duration-200"
                  :style="{ width: `${uploadProgress}%` }"
                />
              </div>
            </div>
          </CardContent>
        </Card>

        <div v-if="documents.length > 0" class="space-y-4">
          <div class="flex items-center justify-between">
            <h2 class="text-2xl font-semibold">Your Documents</h2>
            <Button @click="loadDocuments" variant="outline" size="sm">
              <Icon name="lucide:refresh-cw" class="w-4 h-4 mr-2" />
              Refresh
            </Button>
          </div>
          
          <div class="grid gap-4">
            <Card
              v-for="document in documents"
              :key="document.id"
              class="hover:shadow-md transition-shadow cursor-pointer"
              @click="openDocument(document)"
            >
              <CardContent class="p-4">
                <div class="flex items-center justify-between">
                  <div class="space-y-1">
                    <h3 class="font-medium">{{ document.filename }}</h3>
                    <div class="flex items-center gap-4 text-sm text-muted-foreground">
                      <span>{{ formatFileSize(document.file_size) }}</span>
                      <span>{{ formatDate(document.created_at) }}</span>
                      <Badge :variant="getStatusVariant(document.status)">
                        {{ document.status }}
                      </Badge>
                    </div>
                  </div>
                  
                  <Icon name="lucide:chevron-right" class="w-4 h-4" />
                </div>
              </CardContent>
            </Card>
          </div>
        </div>
        
        <div v-else-if="!isLoadingDocuments" class="text-center py-12">
          <Icon name="lucide:file-text" class="w-16 h-16 mx-auto mb-4 text-muted-foreground" />
          <h3 class="text-lg font-medium mb-2">No documents yet</h3>
          <p class="text-muted-foreground">Upload your first PDF to get started</p>
        </div>
        
        <div v-if="isLoadingDocuments" class="flex items-center justify-center py-12">
          <Icon name="lucide:loader-2" class="w-8 h-8 animate-spin" />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const router = useRouter()
const { uploadPdf, getDocuments } = useApi()

const fileInput = ref<HTMLInputElement>()
const selectedFile = ref<File | null>(null)
const isUploading = ref(false)
const uploadProgress = ref(0)
const uploadError = ref<string | null>(null)
const isDragging = ref(false)

const documents = ref<any[]>([])
const isLoadingDocuments = ref(false)

const formatFileSize = (bytes: number) => {
  const sizes = ['Bytes', 'KB', 'MB', 'GB']
  if (bytes === 0) return '0 Bytes'
  const i = Math.floor(Math.log(bytes) / Math.log(1024))
  return Math.round(bytes / Math.pow(1024, i) * 100) / 100 + ' ' + sizes[i]
}

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleDateString()
}

const getStatusVariant = (status: string) => {
  switch (status) {
    case 'completed': return 'default'
    case 'processing': return 'secondary'
    case 'error': return 'destructive'
    case 'uploaded': return 'outline'
    default: return 'outline'
  }
}

const triggerFileInput = () => {
  fileInput.value?.click()
}

const onFileSelect = (event: Event) => {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (file) {
    handleFileUpload(file)
  }
}

const onDrop = (event: DragEvent) => {
  event.preventDefault()
  isDragging.value = false
  
  const files = event.dataTransfer?.files
  if (files && files.length > 0) {
    const file = files[0]
    if (file.type === 'application/pdf') {
      handleFileUpload(file)
    } else {
      uploadError.value = 'Please select a PDF file'
    }
  }
}

const handleFileUpload = async (file: File) => {
  selectedFile.value = file
  isUploading.value = true
  uploadError.value = null
  uploadProgress.value = 0
  
  try {
    const progressInterval = setInterval(() => {
      if (uploadProgress.value < 90) {
        uploadProgress.value += 10
      }
    }, 200)
    
    const { data, error } = await uploadPdf(file)
    
    clearInterval(progressInterval)
    uploadProgress.value = 100
    
    if (error) {
      uploadError.value = error
    } else if (data) {
      await loadDocuments()
      selectedFile.value = null
      uploadProgress.value = 0
    }
    
  } catch (error: any) {
    uploadError.value = error.message || 'Upload failed'
  } finally {
    isUploading.value = false
  }
}

const loadDocuments = async () => {
  try {
    isLoadingDocuments.value = true
    const { data, error } = await getDocuments()
    
    if (error) {
      console.error('Error loading documents:', error)
    } else {
      documents.value = data || []
    }
  } catch (error) {
    console.error('Error loading documents:', error)
  } finally {
    isLoadingDocuments.value = false
  }
}

const openDocument = (document: any) => {
  router.push(`/read/${document.id}`)
}

onMounted(() => {
  loadDocuments()
})

useSeoMeta({
  title: 'Open Reader - AI-Powered PDF to Audio',
  description: 'Upload PDFs and listen to them with AI-generated text-to-speech',
})
</script>
