# Open Reader

An AI-powered PDF to audio converter that extracts text from PDFs and generates speech using text-to-speech technology.

## Architecture

This is a pnpm monorepo with two main applications:

- **Frontend** (`apps/frontend`): Nuxt 3 application with Vue 3, TypeScript, and Tailwind CSS
- **Backend** (`apps/backend`): FastAPI Python application for PDF processing and TTS generation

## Features

- 📄 PDF upload and text extraction
- 🔤 Text chunking into sentence-based segments
- 🎵 AI-powered text-to-speech generation using Cartesia Sonic
- 📱 Modern scrolling text interface (like movie credits)
- 🎧 Audio player with chunk navigation
- 💾 Redis caching for audio files
- ☁️ R2 storage for file management
- 📊 Real-time processing status updates

## Setup

### Prerequisites

- Node.js 18+ and pnpm
- Python 3.9+
- PostgreSQL database
- Redis server
- Cloudflare R2 storage account
- Together AI API key
- Cartesia API key (optional, for better TTS)

### Installation

1. **Clone and install dependencies:**
   ```bash
   git clone <repository-url>
   cd open-reader
   pnpm install
   ```

2. **Backend setup:**
   ```bash
   cd apps/backend
   pip install -r requirements.txt
   
   # Copy environment file and configure
   cp .env.example .env
   # Edit .env with your database, Redis, and API keys
   ```

3. **Frontend setup:**
   ```bash
   cd apps/frontend
   
   # Copy environment file and configure
   cp .env.example .env
   # Edit .env with your configuration
   ```

### Environment Variables

#### Backend (.env)
```env
DATABASE_URL=postgresql://user:password@localhost:5432/openreader
REDIS_URL=redis://localhost:6379

R2_ACCOUNT_ID=your_r2_account_id
R2_ACCESS_KEY_ID=your_r2_access_key_id
R2_SECRET_ACCESS_KEY=your_r2_secret_access_key
R2_ENDPOINT=https://your_account_id.r2.cloudflarestorage.com
R2_BUCKET_NAME=open-reader

TOGETHER_API_KEY=your_together_api_key
CARTESIA_API_KEY=your_cartesia_api_key
```

#### Frontend (.env)
```env
BACKEND_URL=http://localhost:8000
```

### Running the Application

#### Development Mode

1. **Start the backend:**
   ```bash
   cd apps/backend
   pnpm dev
   # or
   python -m uvicorn main:app --reload --host 0.0.0.0 --port 8000
   ```

2. **Start the frontend:**
   ```bash
   cd apps/frontend
   pnpm dev
   ```

3. **Or run both concurrently from root:**
   ```bash
   pnpm dev
   ```

#### Production Mode

```bash
# Build both applications
pnpm build

# Start backend
cd apps/backend
pnpm start

# Start frontend
cd apps/frontend
pnpm preview
```

## Usage

1. **Upload PDF**: Go to the homepage and upload a PDF file
2. **Process Document**: Click "Process Document" to extract and chunk the text
3. **Read & Listen**: Navigate to the document page to see the scrolling text and audio player
4. **Generate Audio**: Click play to generate TTS for the current chunk, or "Generate All Audio" for batch processing

## Tech Stack

### Frontend
- **Nuxt 3** - Vue.js framework with auto-imports and file-based routing
- **Vue 3** - Progressive JavaScript framework with Composition API
- **TypeScript** - Type-safe JavaScript
- **Tailwind CSS** - Utility-first CSS framework
- **Shadcn Vue** - Beautiful and accessible UI components

### Backend
- **FastAPI** - Modern, fast Python web framework
- **SQLAlchemy** - SQL toolkit and ORM
- **PostgreSQL** - Relational database
- **Redis** - In-memory caching
- **PyPDF2** - PDF text extraction
- **NLTK** - Natural language processing for text chunking
- **Boto3** - AWS/R2 storage client
- **Together AI** - AI platform for models
- **Cartesia Sonic** - High-quality text-to-speech API

## API Endpoints

### Upload
- `POST /api/upload` - Upload PDF file

### PDF Processing
- `POST /api/pdf/process/{document_id}` - Process PDF and extract chunks
- `GET /api/pdf/{document_id}/chunks` - Get text chunks
- `GET /api/pdf/{document_id}/status` - Get processing status

### Text-to-Speech
- `POST /api/tts/generate` - Generate TTS for a chunk
- `GET /api/tts/status/{chunk_id}` - Get TTS status
- `POST /api/tts/generate-batch` - Generate TTS for multiple chunks

### Documents
- `GET /api/documents` - Get all documents
- `GET /api/documents/{document_id}` - Get specific document

## License

MIT License

## 🌟 Features

- 📚 **PDF to Audiobook Conversion**: Convert any PDF document into a high-quality audiobook
- 🎭 **Multiple AI Voices**: Choose from a variety of natural-sounding AI voices
- 📱 **Responsive Design**: Beautiful and intuitive interface that works on any device
- 🌓 **Dark Mode**: Easy on the eyes with automatic dark mode support
- 💾 **Local Processing**: Process your documents locally for enhanced privacy
- 🔄 **Progress Tracking**: Keep track of your reading progress across documents
- 📑 **Bookmarking**: Save your spot and continue listening later

## 🚀 Getting Started

### Prerequisites

- Node.js 18.x or higher
- pnpm (recommended) or npm

### Installation

1. Clone the repository:

```bash
git clone https://github.com/yourusername/open-reader.git
cd open-reader
```

2. Install dependencies:

```bash
pnpm install
```

3. Start the development server:

```bash
pnpm dev
```

The application will be available at `http://localhost:3000`.

## 🤝 Contributing

We welcome contributions! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [ElevenLabs](https://elevenlabs.io) - For inspiration and pushing the boundaries of AI voice technology
- [Nuxt](https://nuxt.com) - For the amazing framework
- [Together](https://together.ai) - For the AI inference endpoints
- The open-source community

## 📸 Screenshots

_Coming soon_

---

<div align="center">
  <p>Made with ❤️ by <a href="https://github.com/alexander-gekov">Alexander Gekov</a></p>
</div>
