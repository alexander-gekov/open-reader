# PDF to Audio Converter - Implementation Guide

## Overview

This system converts PDF documents to audio by:
1. Uploading PDFs to cloud storage (S3/R2)
2. Parsing text and splitting into manageable chunks
3. Converting chunks to audio via TTS
4. Caching audio URLs in Redis
5. Providing a web interface for playback

## ✅ Completed Tasks

### 1. File Upload Component Enhancement
- ✅ Modified `components/ui/file-upload/FileUpload.vue` to only accept PDF files
- ✅ Added proper upload button with PDF validation
- ✅ Added file removal and clear functionality
- ✅ Enhanced UI with better file management

### 2. PDF Processing Library
- ✅ Installed `pdf-parse` and `@types/pdf-parse`
- ✅ Created `lib/pdf-processor.ts` with text extraction and chunking logic
- ✅ Implemented sentence splitting and smart chunking with overlap

### 3. Database Schema
- ✅ Updated `prisma/schema.prisma` with comprehensive models:
  - `User` - User management
  - `Document` - PDF metadata
  - `Chunk` - Text chunks from PDFs
  - `AudioFile` - Generated audio files
  - Status enums for processing states

### 4. Redis Integration
- ✅ Installed `ioredis` and `@upstash/redis`
- ✅ Created `lib/redis.ts` with caching utilities
- ✅ Implemented `ChunkCache` class for audio URL caching

### 5. Backend API Endpoints
- ✅ Created `server/api/upload.post.ts` for PDF upload and processing
- ✅ Created `server/api/audio/[docId]/[chunkIndex].get.ts` for audio requests
- ✅ Implemented Redis-first caching strategy

### 6. Frontend Implementation
- ✅ Created `pages/upload.vue` with complete workflow
- ✅ Upload progress tracking
- ✅ Document management interface
- ✅ Audio player with chunk selection

### 7. Development Environment
- ✅ Created `docker-compose.yml` for PostgreSQL and Redis
- ✅ Set up `.env` template with all required variables

## 🔧 Setup Instructions

### 1. Environment Setup

```bash
# Copy and configure environment variables
cp .env.example .env

# Required variables:
DATABASE_URL="postgresql://postgres:password@localhost:5432/openreader?schema=public"
REDIS_URL="redis://localhost:6379"
AWS_ACCESS_KEY_ID="your-access-key"
AWS_SECRET_ACCESS_KEY="your-secret-key"
AWS_REGION="your-region"
AWS_S3_BUCKET="your-bucket-name"
TOGETHER_API_KEY="your-together-api-key"
```

### 2. Database and Redis Setup

```bash
# Start PostgreSQL and Redis
docker-compose up -d

# Install dependencies
pnpm install

# Generate Prisma client and run migrations
pnpm prisma generate
pnpm prisma db push
```

### 3. Install Additional Dependencies

```bash
# Background job processing
pnpm add @trigger.dev/sdk

# Audio processing (if needed)
pnpm add @aws-sdk/client-polly  # For AWS Polly TTS
```

## 🚧 Remaining Tasks

### 1. TTS Integration
- [ ] Set up Together AI TTS or AWS Polly
- [ ] Create TTS generation service
- [ ] Implement audio file upload to S3/R2

### 2. Background Job Processing
- [ ] Set up Trigger.dev
- [ ] Create TTS generation jobs
- [ ] Implement job queue for audio processing

### 3. Authentication Integration
- [ ] Integrate Clerk authentication
- [ ] Add user-specific document access
- [ ] Implement proper user sessions

### 4. Production Setup
- [ ] Configure production database
- [ ] Set up Redis in production
- [ ] Configure S3/R2 bucket with proper CORS
- [ ] Add proper error handling and logging

## 📁 File Structure

```
open-reader/
├── components/ui/file-upload/
│   └── FileUpload.vue              # Enhanced PDF upload component
├── lib/
│   ├── pdf-processor.ts            # PDF parsing and chunking
│   ├── redis.ts                    # Redis caching utilities
│   └── s3.ts                       # S3 utilities (existing)
├── pages/
│   └── upload.vue                  # Main upload interface
├── server/api/
│   ├── upload.post.ts              # PDF upload endpoint
│   └── audio/[docId]/[chunkIndex].get.ts  # Audio request endpoint
├── prisma/
│   └── schema.prisma               # Database schema
├── docker-compose.yml              # Development services
└── .env                           # Environment variables
```

## 🔄 Workflow

1. **Upload**: User uploads PDF files via the web interface
2. **Processing**: PDF is parsed, text extracted, and split into chunks
3. **Storage**: PDF stored in S3/R2, metadata in PostgreSQL
4. **Audio Request**: User clicks to play a chunk
5. **Cache Check**: System checks Redis for existing audio
6. **TTS Generation**: If not cached, triggers background TTS job
7. **Playback**: Returns audio URL for immediate playback

## 🎯 Next Steps

1. **Set up TTS Service**:
   ```typescript
   // Example TTS integration
   import Together from "together-ai"
   
   const together = new Together({
     apiKey: process.env.TOGETHER_API_KEY
   })
   
   async function generateTTS(text: string): Promise<Buffer> {
     // Implement TTS generation
   }
   ```

2. **Configure Trigger.dev**:
   ```bash
   npx trigger.dev init
   ```

3. **Set up production environment variables**

4. **Test end-to-end workflow**

## 🐛 Known Issues

- TypeScript configuration may need adjustment for Prisma imports
- Server-side functions may need proper Nuxt 3 configuration
- Redis connection needs production-ready configuration
- S3/R2 CORS needs proper setup for audio playback

## 🔧 Troubleshooting

### Prisma Issues
```bash
pnpm prisma generate
pnpm prisma db push
```

### TypeScript Errors
Check `tsconfig.json` and ensure proper module resolution.

### Docker Issues
```bash
docker-compose down -v
docker-compose up -d
```

This implementation provides a solid foundation for the PDF to audio conversion system with proper caching, chunking, and a modern web interface.