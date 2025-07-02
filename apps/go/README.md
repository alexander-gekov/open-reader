# Go PDF TTS Backend

A Go backend service that processes PDF files, extracts text, chunks it into sentences, and generates TTS audio using Together.ai's Cartesia Sonic model with a rolling buffer system.

## Features

- PDF upload and text extraction
- Smart text chunking (1-2 sentences per chunk)
- TTS generation using Together.ai Cartesia Sonic
- Rolling buffer system for efficient audio streaming
- RESTful API with CORS support
- Automatic cleanup of uploaded files

## Setup

1. Install dependencies:

```bash
cd apps/go
go mod tidy
```

2. Set your Together.ai API key:

```bash
export TOGETHER_API_KEY="your_api_key_here"
```

3. Run the server:

```bash
go run main.go
```

The server will start on port 8080 (or the port specified in the `PORT` environment variable).

## API Endpoints

### POST /upload

Upload a PDF file for processing.

**Request:**

- Content-Type: multipart/form-data
- Body: PDF file with field name "file"

**Response:**

```json
{
  "message": "PDF processed successfully",
  "chunks": ["First sentence chunk.", "Second sentence chunk.", ...],
  "audio_id": "audio_1234567890"
}
```

### GET /audio/next

Get the next audio chunk in the sequence.

**Response:**

- Success (200): Audio data (audio/mpeg)
- Processing (202): `{"message": "Audio not ready yet", "retry": true}`
- Complete (204): `{"message": "No more audio chunks", "retry": false}`

### GET /status

Get the current processing status.

**Response:**

```json
{
  "status": "ready",
  "hasMore": true,
  "currentIdx": 2,
  "totalChunks": 10
}
```

### GET /health

Health check endpoint.

**Response:**

```json
{
  "status": "healthy"
}
```

## How It Works

1. **PDF Upload**: Client uploads a PDF to `/upload`
2. **Text Extraction**: Server extracts text from all pages
3. **Text Chunking**: Text is split into 1-2 sentence chunks (max 25 words)
4. **Rolling Buffer**:

   - Immediately starts generating TTS for chunk 0
   - Starts generating TTS for chunk 1 in parallel
   - As chunk 0 audio is consumed, starts generating chunk 2
   - Maintains a buffer of 1-2 pre-generated audio chunks

5. **Audio Streaming**: Client polls `/audio/next` to get sequential audio chunks

## Environment Variables

- `TOGETHER_API_KEY`: Required. Your Together.ai API key
- `PORT`: Optional. Server port (default: 8080)

## File Cleanup

Uploaded PDF files are automatically deleted after 24 hours to save disk space.
