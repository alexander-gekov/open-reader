import { PrismaClient } from '@prisma/client'
import { ChunkCache } from '../../../../lib/redis'

const prisma = new PrismaClient()
const chunkCache = new ChunkCache()

export default defineEventHandler(async (event) => {
  try {
    const docId = getRouterParam(event, 'docId')
    const chunkIndex = parseInt(getRouterParam(event, 'chunkIndex') || '0')

    if (!docId || isNaN(chunkIndex)) {
      throw createError({
        statusCode: 400,
        statusMessage: 'Invalid document ID or chunk index'
      })
    }

    // Check Redis cache first
    const cached = await chunkCache.get(docId, chunkIndex)
    
    if (cached) {
      if (cached.status === 'completed' && cached.audio_url) {
        return {
          status: 'completed',
          audio_url: cached.audio_url,
          cached: true
        }
      }
      
      if (cached.status === 'processing') {
        return {
          status: 'processing',
          cached: true
        }
      }
    }

    // Verify chunk exists in database
    const chunk = await prisma.chunk.findUnique({
      where: {
        documentId_index: {
          documentId: docId,
          index: chunkIndex
        }
      },
      include: {
        document: true,
        audioFiles: {
          where: { status: 'COMPLETED' },
          orderBy: { createdAt: 'desc' },
          take: 1
        }
      }
    })

    if (!chunk) {
      throw createError({
        statusCode: 404,
        statusMessage: 'Chunk not found'
      })
    }

    // Check if we have a completed audio file
    const existingAudio = chunk.audioFiles[0]
    if (existingAudio) {
      // Cache the result
      await chunkCache.setCompleted(docId, chunkIndex, existingAudio.s3Url)
      
      return {
        status: 'completed',
        audio_url: existingAudio.s3Url,
        cached: false
      }
    }

    // Set processing status in cache
    await chunkCache.setProcessing(docId, chunkIndex)

    // TODO: Trigger background TTS generation job
    // This would typically use Trigger.dev or a similar service
    // triggerTTSGeneration(docId, chunkIndex, chunk.text)

    return {
      status: 'processing',
      message: 'Audio generation started',
      cached: false
    }

  } catch (error) {
    console.error('Audio request error:', error)
    
    if (error.statusCode) {
      throw error
    }

    throw createError({
      statusCode: 500,
      statusMessage: 'Internal server error'
    })
  }
})