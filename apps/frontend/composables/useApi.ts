export const useApi = () => {
  const config = useRuntimeConfig()
  const backendUrl = config.public.backendUrl || 'http://localhost:8000'

  const apiCall = async (endpoint: string, options: any = {}) => {
    try {
      const response = await $fetch(`${backendUrl}${endpoint}`, {
        ...options,
        headers: {
          ...options.headers,
        },
      })
      return { data: response, error: null }
    } catch (error: any) {
      console.error('API Error:', error)
      return { data: null, error: error.data?.detail || error.message || 'An error occurred' }
    }
  }

  const uploadPdf = async (file: File) => {
    const formData = new FormData()
    formData.append('file', file)
    
    return await apiCall('/api/upload', {
      method: 'POST',
      body: formData,
    })
  }

  const processPdf = async (documentId: number) => {
    return await apiCall(`/api/pdf/process/${documentId}`, {
      method: 'POST',
    })
  }

  const getDocumentChunks = async (documentId: number) => {
    return await apiCall(`/api/pdf/${documentId}/chunks`)
  }

  const getProcessingStatus = async (documentId: number) => {
    return await apiCall(`/api/pdf/${documentId}/status`)
  }

  const generateTts = async (chunkId: number) => {
    return await apiCall('/api/tts/generate', {
      method: 'POST',
      body: { chunk_id: chunkId },
    })
  }

  const getTtsStatus = async (chunkId: number) => {
    return await apiCall(`/api/tts/status/${chunkId}`)
  }

  const generateBatchTts = async (chunkIds: number[]) => {
    return await apiCall('/api/tts/generate-batch', {
      method: 'POST',
      body: chunkIds,
    })
  }

  const getDocuments = async () => {
    return await apiCall('/api/documents')
  }

  return {
    uploadPdf,
    processPdf,
    getDocumentChunks,
    getProcessingStatus,
    generateTts,
    getTtsStatus,
    generateBatchTts,
    getDocuments,
  }
}