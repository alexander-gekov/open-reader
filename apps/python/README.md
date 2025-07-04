# Open Reader Python Backend

A comprehensive Python backend for the Open Reader application, built with FastAPI, providing advanced PDF processing and TTS capabilities with multiple provider support.

## Features

### PDF Processing
- **Advanced Text Extraction**: Uses PyMuPDF (fitz) with PyPDF2 fallback for robust PDF text extraction
- **Smart Text Chunking**: NLTK-based sentence tokenization with intelligent chunking
- **Multi-page Support**: Handles PDFs of any size with proper page counting
- **Text Cleaning**: Sophisticated text normalization and cleanup

### TTS Providers
- **ElevenLabs**: High-quality voice synthesis with voice cloning support
- **AWS Polly**: Neural and standard voices with multiple languages
- **OpenAI TTS**: Latest AI-powered text-to-speech
- **Cartesia**: Fast, high-quality voice synthesis
- **Extensible**: Easy to add new TTS providers

### Storage & Database
- **S3 Integration**: Stores PDFs and audio files in AWS S3
- **PostgreSQL**: Robust database with proper relationships
- **Chunk Tracking**: Individual chunk status and audio URL tracking
- **Database Migrations**: Alembic for schema management

### API Features
- **FastAPI**: Modern, fast API framework with automatic OpenAPI docs
- **Async Processing**: Non-blocking chunk processing
- **Status Tracking**: Real-time chunk processing status
- **Error Handling**: Comprehensive error handling and logging
- **CORS Support**: Cross-origin resource sharing enabled

## Installation

### Prerequisites
- Python 3.11+
- PostgreSQL database
- AWS S3 bucket
- TTS provider API keys

### Setup

1. **Clone and navigate to the Python backend**:
```bash
cd apps/python
```

2. **Create virtual environment**:
```bash
python -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate
```

3. **Install dependencies**:
```bash
pip install -r requirements.txt
```

4. **Set up environment variables**:
```bash
cp .env.example .env
# Edit .env with your actual values
```

5. **Run database migrations**:
```bash
alembic upgrade head
```

6. **Start the server**:
```bash
uvicorn main:app --reload
```

The server will start on `http://localhost:8000`

## API Endpoints

### Core Endpoints

#### `POST /upload`
Upload and process a PDF file.

**Headers:**
- `X-TTS-Provider`: TTS provider (elevenlabs, polly, openai, cartesia)
- `X-TTS-API-Key`: API key for the chosen provider
- `X-TTS-Model`: Model ID (optional)
- `X-TTS-Voice`: Voice ID (optional)

**Request:** Multipart form data with PDF file

**Response:**
```json
{
  "success": true,
  "message": "PDF processed successfully",
  "chunks": ["First chunk text...", "Second chunk text..."],
  "audioId": "pdf-uuid",
  "totalChunks": 10
}
```

#### `GET /audio/status/{chunk_index}`
Get the status of audio generation for a specific chunk.

**Response:**
```json
{
  "status": "ready",
  "url": "https://bucket.s3.amazonaws.com/audio/file.mp3",
  "hasNext": true,
  "nextReady": false
}
```

#### `GET /start-next/{current_chunk}`
Trigger processing of the next chunk.

**Response:**
```json
{
  "message": "Started processing chunks"
}
```

#### `GET /health`
Health check endpoint.

#### `GET /status`
Get overall processing status.

#### `POST /generate-audio`
Generate audio for custom text and settings.

## Environment Variables

### Required
- `DATABASE_URL`: PostgreSQL connection string
- `AWS_ACCESS_KEY_ID`: AWS access key
- `AWS_SECRET_ACCESS_KEY`: AWS secret key
- `AWS_BUCKET_NAME`: S3 bucket name

### TTS Providers (at least one required)
- `ELEVENLABS_API_KEY`: ElevenLabs API key
- `OPENAI_API_KEY`: OpenAI API key
- `CARTESIA_API_KEY`: Cartesia API key

### Optional
- `AWS_REGION`: AWS region (default: us-east-1)
- `REDIS_URL`: Redis connection string
- `PORT`: Server port (default: 8000)
- `LOG_LEVEL`: Logging level (default: INFO)

## Database Schema

### Tables

#### `pdfs`
- `id`: UUID primary key
- `title`: PDF filename
- `url`: S3 URL of the PDF
- `coverUrl`: S3 URL of cover image (optional)
- `totalPages`: Number of pages
- `isArchived`: Archive flag
- `createdAt`, `updatedAt`: Timestamps

#### `chunks`
- `id`: UUID primary key
- `pdfId`: Foreign key to `pdfs.id`
- `index`: Chunk order number
- `text`: Chunk text content
- `audioUrl`: S3 URL of generated audio
- `status`: Processing status (pending, processing, completed, error)
- `error_message`: Error details if failed
- `createdAt`, `updatedAt`: Timestamps

## Docker Support

### Build Image
```bash
docker build -t open-reader-python .
```

### Run Container
```bash
docker run -p 8000:8000 --env-file .env open-reader-python
```

### Docker Compose
```yaml
version: '3.8'
services:
  python-backend:
    build: .
    ports:
      - "8000:8000"
    environment:
      - DATABASE_URL=postgresql://user:pass@db:5432/openreader
      - AWS_ACCESS_KEY_ID=your_key
      - AWS_SECRET_ACCESS_KEY=your_secret
      - AWS_BUCKET_NAME=your_bucket
      - ELEVENLABS_API_KEY=your_key
    depends_on:
      - db
  
  db:
    image: postgres:15
    environment:
      - POSTGRES_DB=openreader
      - POSTGRES_USER=user
      - POSTGRES_PASSWORD=pass
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:
```

## Development

### Running Tests
```bash
pytest
```

### Code Formatting
```bash
black .
isort .
```

### Type Checking
```bash
mypy .
```

### Database Migrations
```bash
# Create new migration
alembic revision --autogenerate -m "Description"

# Apply migrations
alembic upgrade head

# Rollback
alembic downgrade -1
```

## Architecture

### Components

1. **FastAPI Application** (`main.py`): Main API server
2. **PDF Processor** (`pdf_processor.py`): PDF text extraction and chunking
3. **TTS Providers** (`tts_providers.py`): Multiple TTS service integrations
4. **S3 Storage** (`s3_storage.py`): AWS S3 file management
5. **Chunk Processor** (`chunk_processor.py`): Async chunk processing
6. **Database Models** (`models.py`): SQLAlchemy models
7. **Schemas** (`schemas.py`): Pydantic request/response models

### Processing Flow

1. PDF uploaded via `/upload` endpoint
2. Text extracted using PyMuPDF/PyPDF2
3. Text chunked using NLTK sentence tokenization
4. PDF stored in S3, metadata in database
5. Chunks created in database with "pending" status
6. First chunk processing triggered asynchronously
7. TTS provider generates audio for chunk
8. Audio uploaded to S3, chunk status updated
9. Frontend polls `/audio/status/{chunk}` for completion
10. Frontend triggers `/start-next/{chunk}` for next chunk

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

This project is licensed under the MIT License.