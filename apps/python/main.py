from fastapi import FastAPI, HTTPException, Depends, UploadFile, File, Header
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import StreamingResponse
from sqlalchemy.orm import Session
from typing import Optional, List
import os
import asyncio
import logging
import io
from dotenv import load_dotenv

from database import get_db, engine, Base
from models import PDFDocument, Chunk, TTSSettings as TTSSettingsModel
from schemas import UploadResponse, AudioStatusResponse, TTSSettings, AudioGenerationRequest
from pdf_processor import PDFProcessor
from tts_providers import TTSProviderFactory
from s3_storage import S3Storage
from chunk_processor import ChunkProcessor

load_dotenv()

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

app = FastAPI(title="Open Reader Python Backend", version="1.0.0")

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

Base.metadata.create_all(bind=engine)

pdf_processor = PDFProcessor()
s3_storage = S3Storage()
chunk_processor = ChunkProcessor()

@app.post("/upload", response_model=UploadResponse)
async def upload_pdf(
    file: UploadFile = File(...),
    x_tts_provider: Optional[str] = Header(None, alias="X-TTS-Provider"),
    x_tts_api_key: Optional[str] = Header(None, alias="X-TTS-API-Key"),
    x_tts_model: Optional[str] = Header(None, alias="X-TTS-Model"),
    x_tts_voice: Optional[str] = Header(None, alias="X-TTS-Voice"),
    db: Session = Depends(get_db)
):
    """
    Upload and process a PDF file.
    """
    try:
        if not file.filename.lower().endswith('.pdf'):
            raise HTTPException(status_code=400, detail="Only PDF files are allowed")
        
        if file.size > 20 * 1024 * 1024:  # 20MB limit
            raise HTTPException(status_code=400, detail="File size should be less than 20MB")
        
        # Create TTS settings
        tts_settings = TTSSettings(
            provider=x_tts_provider or "elevenlabs",
            apiKey=x_tts_api_key or "",
            model=x_tts_model or "",
            voice=x_tts_voice or ""
        )
        
        if tts_settings.provider != "fallback" and not tts_settings.apiKey:
            raise HTTPException(status_code=400, detail="API key is required for non-fallback providers")
        
        # Read file content
        content = await file.read()
        
        # Extract text from PDF
        text = pdf_processor.extract_text_from_pdf(content)
        if not text.strip():
            raise HTTPException(status_code=400, detail="No text found in PDF")
        
        # Chunk the text
        chunks = pdf_processor.chunk_text(text)
        if not chunks:
            raise HTTPException(status_code=400, detail="No text chunks created")
        
        # Get total pages
        total_pages = pdf_processor.get_page_count(content)
        
        # Upload PDF to S3
        pdf_key = f"pdfs/{file.filename}"
        pdf_url = await s3_storage.upload_file(content, pdf_key, "application/pdf")
        
        # Create PDF record in database
        pdf_doc = PDFDocument(
            title=file.filename,
            url=pdf_url,
            totalPages=total_pages,
            isArchived=False
        )
        db.add(pdf_doc)
        db.commit()
        db.refresh(pdf_doc)
        
        # Create chunk records in database
        chunk_records = []
        for i, chunk_text in enumerate(chunks):
            chunk_record = Chunk(
                pdfId=pdf_doc.id,
                index=i,
                text=chunk_text,
                audioUrl=None,
                status="pending"
            )
            chunk_records.append(chunk_record)
            db.add(chunk_record)
        
        db.commit()
        
        # Start processing first chunk
        audio_id = pdf_doc.id
        asyncio.create_task(
            chunk_processor.process_chunk(
                chunk_records[0], tts_settings, s3_storage, db
            )
        )
        
        return UploadResponse(
            success=True,
            message="PDF processed successfully",
            chunks=chunks,
            audioId=audio_id,
            totalChunks=len(chunks)
        )
        
    except Exception as e:
        logger.error(f"Error processing PDF: {str(e)}")
        raise HTTPException(status_code=500, detail=str(e))

@app.get("/audio/status/{chunk_index}", response_model=AudioStatusResponse)
async def get_audio_status(
    chunk_index: int,
    db: Session = Depends(get_db)
):
    """
    Get the status of audio generation for a specific chunk.
    """
    try:
        chunk = db.query(Chunk).filter(Chunk.index == chunk_index).first()
        if not chunk:
            raise HTTPException(status_code=404, detail="Chunk not found")
        
        if chunk.status == "completed" and chunk.audioUrl:
            next_chunk = db.query(Chunk).filter(Chunk.index == chunk_index + 1).first()
            return AudioStatusResponse(
                status="ready",
                url=chunk.audioUrl,
                hasNext=next_chunk is not None,
                nextReady=next_chunk.status == "completed" if next_chunk else False
            )
        
        if chunk.status == "processing":
            return AudioStatusResponse(status="processing")
        
        if chunk.status == "error":
            return AudioStatusResponse(status="error", error=chunk.error_message)
        
        return AudioStatusResponse(status="pending")
        
    except Exception as e:
        logger.error(f"Error getting audio status: {str(e)}")
        raise HTTPException(status_code=500, detail=str(e))

@app.get("/start-next/{current_chunk}")
async def start_next_chunk(
    current_chunk: int,
    db: Session = Depends(get_db)
):
    """
    Trigger processing of the next chunk.
    """
    try:
        next_chunk_index = current_chunk + 1
        next_chunk = db.query(Chunk).filter(Chunk.index == next_chunk_index).first()
        
        if not next_chunk:
            return {"message": "No more chunks to process"}
        
        if next_chunk.status == "completed":
            return {"message": "Next chunk already processed"}
        
        if next_chunk.status == "processing":
            return {"message": "Next chunk already being processed"}
        
        # Get TTS settings (this should be stored per user/session)
        tts_settings = TTSSettings(
            provider="elevenlabs",
            apiKey=os.getenv("ELEVENLABS_API_KEY", ""),
            model="eleven_flash_v2_5",
            voice="cgSgspJ2msm6clMCkdW9"
        )
        
        # Start processing next chunk
        asyncio.create_task(
            chunk_processor.process_chunk(next_chunk, tts_settings, s3_storage, db)
        )
        
        # Also start processing the chunk after next if it exists
        after_next_chunk = db.query(Chunk).filter(Chunk.index == next_chunk_index + 1).first()
        if after_next_chunk and after_next_chunk.status == "pending":
            asyncio.create_task(
                chunk_processor.process_chunk(after_next_chunk, tts_settings, s3_storage, db)
            )
        
        return {"message": "Started processing chunks"}
        
    except Exception as e:
        logger.error(f"Error starting next chunk: {str(e)}")
        raise HTTPException(status_code=500, detail=str(e))

@app.get("/health")
async def health_check():
    """
    Health check endpoint.
    """
    return {"status": "healthy"}

@app.get("/status")
async def get_status():
    """
    Get overall processing status.
    """
    return {
        "status": "ready",
        "message": "Python backend is running"
    }

@app.post("/generate-audio")
async def generate_audio(
    request: AudioGenerationRequest,
    db: Session = Depends(get_db)
):
    """
    Generate audio for a given text and settings.
    """
    try:
        tts_provider = TTSProviderFactory.create_provider(
            request.settings.provider,
            request.settings.apiKey
        )
        
        options = {
            "model": request.settings.model,
            "voice": request.settings.voice,
            "filename": request.filename,
            "chunk": str(request.chunk)
        }
        
        audio_data = await tts_provider.generate_audio(request.text, options)
        
        return StreamingResponse(
            io.BytesIO(audio_data),
            media_type="audio/mpeg",
            headers={
                "Content-Length": str(len(audio_data)),
                "Cache-Control": "no-cache"
            }
        )
        
    except Exception as e:
        logger.error(f"Error generating audio: {str(e)}")
        raise HTTPException(status_code=500, detail=str(e))

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)