from fastapi import APIRouter, HTTPException, Depends, BackgroundTasks
from sqlalchemy.orm import Session
import uuid
from typing import List

from app.database import get_db
from app.models.document import TextChunk
from app.models.schemas import TTSRequest, TTSResponse
from app.services.tts_service import tts_service
from app.services.s3_service import s3_service
from app.services.redis_service import redis_service

router = APIRouter()


@router.post("/tts/generate", response_model=TTSResponse)
async def generate_tts(
    request: TTSRequest,
    background_tasks: BackgroundTasks,
    db: Session = Depends(get_db)
):
    """Generate TTS for a text chunk"""
    
    chunk = db.query(TextChunk).filter(TextChunk.id == request.chunk_id).first()
    if not chunk:
        raise HTTPException(status_code=404, detail="Text chunk not found")
    
    cached_audio_key = redis_service.get_audio_cache(request.chunk_id)
    if cached_audio_key and s3_service.file_exists(cached_audio_key):
        audio_url = s3_service.get_presigned_url(cached_audio_key)
        if audio_url:
            return TTSResponse(
                chunk_id=request.chunk_id,
                audio_url=audio_url,
                cached=True
            )
    
    processing_status = redis_service.get_processing_status(request.chunk_id)
    if processing_status == "processing":
        raise HTTPException(status_code=409, detail="TTS generation already in progress")
    
    redis_service.set_processing_status(request.chunk_id, "processing")
    background_tasks.add_task(generate_tts_background, request.chunk_id, chunk.content, db)
    
    buffer_queue = redis_service.get_buffer_queue(chunk.document_id)
    if request.chunk_id not in buffer_queue:
        next_chunks = db.query(TextChunk).filter(
            TextChunk.document_id == chunk.document_id,
            TextChunk.chunk_index > chunk.chunk_index,
            TextChunk.chunk_index <= chunk.chunk_index + 2
        ).all()
        
        for next_chunk in next_chunks:
            if not redis_service.get_audio_cache(next_chunk.id):
                background_tasks.add_task(generate_tts_background, next_chunk.id, next_chunk.content, db)
    
    raise HTTPException(status_code=202, detail="TTS generation started. Please check back shortly.")


async def generate_tts_background(chunk_id: int, text: str, db: Session):
    """Background task to generate TTS"""
    try:
        audio_data = await tts_service.generate_speech(text)
        
        if audio_data:
            audio_key = f"audio/{uuid.uuid4()}.mp3"
            
            success = s3_service.upload_file(
                file_content=audio_data,
                key=audio_key,
                content_type="audio/mpeg"
            )
            
            if success:
                chunk = db.query(TextChunk).filter(TextChunk.id == chunk_id).first()
                if chunk:
                    chunk.has_audio = True
                    chunk.audio_file_key = audio_key
                    db.commit()
                
                redis_service.set_audio_cache(chunk_id, audio_key)
        
        redis_service.delete_processing_status(chunk_id)
        
    except Exception as e:
        print(f"Error generating TTS for chunk {chunk_id}: {e}")
        redis_service.delete_processing_status(chunk_id)


@router.get("/tts/status/{chunk_id}")
async def get_tts_status(chunk_id: int, db: Session = Depends(get_db)):
    """Get TTS generation status for a chunk"""
    
    chunk = db.query(TextChunk).filter(TextChunk.id == chunk_id).first()
    if not chunk:
        raise HTTPException(status_code=404, detail="Text chunk not found")
    
    cached_audio_key = redis_service.get_audio_cache(chunk_id)
    processing_status = redis_service.get_processing_status(chunk_id)
    
    if cached_audio_key and s3_service.file_exists(cached_audio_key):
        audio_url = s3_service.get_presigned_url(cached_audio_key)
        return {
            "chunk_id": chunk_id,
            "status": "completed",
            "audio_url": audio_url,
            "has_audio": True
        }
    elif processing_status == "processing":
        return {
            "chunk_id": chunk_id,
            "status": "processing",
            "audio_url": None,
            "has_audio": False
        }
    else:
        return {
            "chunk_id": chunk_id,
            "status": "not_generated",
            "audio_url": None,
            "has_audio": chunk.has_audio
        }


@router.post("/tts/generate-batch")
async def generate_batch_tts(
    chunk_ids: List[int],
    background_tasks: BackgroundTasks,
    db: Session = Depends(get_db)
):
    """Generate TTS for multiple chunks"""
    
    chunks = db.query(TextChunk).filter(TextChunk.id.in_(chunk_ids)).all()
    if not chunks:
        raise HTTPException(status_code=404, detail="No valid chunks found")
    
    for chunk in chunks:
        cached_audio_key = redis_service.get_audio_cache(chunk.id)
        if not cached_audio_key or not s3_service.file_exists(cached_audio_key):
            processing_status = redis_service.get_processing_status(chunk.id)
            if processing_status != "processing":
                redis_service.set_processing_status(chunk.id, "processing")
                background_tasks.add_task(generate_tts_background, chunk.id, chunk.content, db)
    
    return {"message": f"TTS generation started for {len(chunks)} chunks"}