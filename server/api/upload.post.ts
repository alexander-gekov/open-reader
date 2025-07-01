import { PrismaClient } from '@prisma/client'
import { S3Client, PutObjectCommand } from '@aws-sdk/client-s3'
import { PDFProcessor } from '../../lib/pdf-processor'
import { randomUUID } from 'crypto'

const prisma = new PrismaClient()

const s3Client = new S3Client({
  region: process.env.AWS_REGION || 'auto',
  endpoint: process.env.AWS_ENDPOINT,
  credentials: {
    accessKeyId: process.env.AWS_ACCESS_KEY_ID || '',
    secretAccessKey: process.env.AWS_SECRET_ACCESS_KEY || ''
  }
})

export default defineEventHandler(async (event) => {
  try {
    const formData = await readMultipartFormData(event)
    
    if (!formData || formData.length === 0) {
      throw createError({
        statusCode: 400,
        statusMessage: 'No files uploaded'
      })
    }

    // For now, process the first file. You can modify to handle multiple files
    const file = formData[0]
    
    if (!file.data || !file.filename) {
      throw createError({
        statusCode: 400,
        statusMessage: 'Invalid file data'
      })
    }

    // Validate file type
    if (!file.filename.toLowerCase().endsWith('.pdf')) {
      throw createError({
        statusCode: 400,
        statusMessage: 'Only PDF files are allowed'
      })
    }

    // Get user ID from authentication (you'll need to implement this based on your auth setup)
    // For now, using a placeholder - replace with actual user ID from Clerk
    const userId = 'temp-user-id' // TODO: Get from authentication

    // Generate unique filename
    const fileId = randomUUID()
    const fileName = `${fileId}.pdf`
    const s3Key = `pdfs/${fileName}`

    // Upload to S3/R2
    const uploadCommand = new PutObjectCommand({
      Bucket: process.env.AWS_S3_BUCKET || 'openreader-pdfs',
      Key: s3Key,
      Body: file.data,
      ContentType: 'application/pdf'
    })

    await s3Client.send(uploadCommand)

    // Generate S3 URL
    const s3Url = `https://${process.env.AWS_S3_BUCKET || 'openreader-pdfs'}.s3.${process.env.AWS_REGION || 'auto'}.amazonaws.com/${s3Key}`

    // Process PDF
    const processed = await PDFProcessor.processPDF(Buffer.from(file.data))

    // Save document to database
    const document = await prisma.document.create({
      data: {
        fileName,
        originalName: file.filename,
        fileSize: file.data.length,
        mimeType: 'application/pdf',
        s3Key,
        s3Url,
        textContent: processed.text,
        totalPages: processed.totalPages,
        status: 'PROCESSING',
        userId
      }
    })

    // Save chunks to database
    const chunks = await Promise.all(
      processed.chunks.map(chunk =>
        prisma.chunk.create({
          data: {
            documentId: document.id,
            index: chunk.index,
            text: chunk.text,
            wordCount: chunk.wordCount,
            startPage: chunk.startPage,
            endPage: chunk.endPage
          }
        })
      )
    )

    // Update document status to completed
    await prisma.document.update({
      where: { id: document.id },
      data: { status: 'COMPLETED' }
    })

    return {
      success: true,
      document: {
        id: document.id,
        fileName: document.fileName,
        originalName: document.originalName,
        status: document.status,
        totalPages: document.totalPages,
        chunksCount: chunks.length
      }
    }

  } catch (error) {
    console.error('Upload error:', error)
    
    if (error.statusCode) {
      throw error
    }

    throw createError({
      statusCode: 500,
      statusMessage: 'Internal server error during upload'
    })
  }
})