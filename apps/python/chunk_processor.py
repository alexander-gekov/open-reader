import logging
import asyncio
from typing import Dict, Any
from sqlalchemy.orm import Session
from models import Chunk
from schemas import TTSSettings
from tts_providers import TTSProviderFactory
from s3_storage import S3Storage
import uuid

logger = logging.getLogger(__name__)

class ChunkProcessor:
    """Handles asynchronous processing of text chunks into audio."""
    
    def __init__(self):
        self.processing_chunks: Dict[str, bool] = {}
    
    async def process_chunk(
        self, 
        chunk: Chunk, 
        tts_settings: TTSSettings, 
        s3_storage: S3Storage,
        db: Session
    ) -> bool:
        """
        Process a single chunk by generating audio and uploading to S3.
        """
        chunk_id = chunk.id
        
        # Prevent duplicate processing
        if chunk_id in self.processing_chunks:
            logger.info(f"Chunk {chunk_id} already being processed")
            return False
        
        self.processing_chunks[chunk_id] = True
        
        try:
            # Update chunk status to processing
            chunk.status = "processing"
            db.commit()
            
            logger.info(f"Starting processing of chunk {chunk.index}: {chunk.text[:50]}...")
            
            # Create TTS provider
            tts_provider = TTSProviderFactory.create_provider(
                tts_settings.provider,
                tts_settings.apiKey
            )
            
            # Generate audio
            options = {
                "model": tts_settings.model,
                "voice": tts_settings.voice,
                "filename": f"chunk_{chunk.pdfId}",
                "chunk": str(chunk.index)
            }
            
            audio_data = await tts_provider.generate_audio(chunk.text, options)
            
            # Upload audio to S3
            audio_url = await s3_storage.upload_audio(
                audio_data,
                f"pdf_{chunk.pdfId}",
                chunk.index
            )
            
            # Update chunk with audio URL
            chunk.audioUrl = audio_url
            chunk.status = "completed"
            db.commit()
            
            logger.info(f"Successfully processed chunk {chunk.index}, audio URL: {audio_url}")
            return True
            
        except Exception as e:
            logger.error(f"Error processing chunk {chunk_id}: {str(e)}")
            
            # Update chunk status to error
            chunk.status = "error"
            chunk.error_message = str(e)
            db.commit()
            
            return False
            
        finally:
            # Remove from processing set
            self.processing_chunks.pop(chunk_id, None)
    
    async def process_chunks_batch(
        self,
        chunks: list[Chunk],
        tts_settings: TTSSettings,
        s3_storage: S3Storage,
        db: Session,
        batch_size: int = 3
    ) -> int:
        """
        Process multiple chunks in batches to avoid overwhelming the TTS service.
        """
        successful_count = 0
        
        for i in range(0, len(chunks), batch_size):
            batch = chunks[i:i + batch_size]
            
            # Process batch concurrently
            tasks = [
                self.process_chunk(chunk, tts_settings, s3_storage, db)
                for chunk in batch
            ]
            
            results = await asyncio.gather(*tasks, return_exceptions=True)
            
            # Count successful processings
            for result in results:
                if result is True:
                    successful_count += 1
            
            # Wait a bit between batches to be respectful to the TTS service
            if i + batch_size < len(chunks):
                await asyncio.sleep(1)
        
        return successful_count
    
    def is_processing(self, chunk_id: str) -> bool:
        """Check if a chunk is currently being processed."""
        return chunk_id in self.processing_chunks
    
    def get_processing_status(self) -> Dict[str, bool]:
        """Get the current processing status of all chunks."""
        return self.processing_chunks.copy()