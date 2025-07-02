import redis
import json
from typing import Optional
from app.core.config import settings


class RedisService:
    def __init__(self):
        self.client = redis.from_url(settings.redis_url, decode_responses=True)
    
    def set_audio_cache(self, chunk_id: int, audio_file_key: str, expiry: int = 86400):
        """Cache audio file key for a chunk"""
        key = f"audio:chunk:{chunk_id}"
        self.client.setex(key, expiry, audio_file_key)
    
    def get_audio_cache(self, chunk_id: int) -> Optional[str]:
        """Get cached audio file key for a chunk"""
        key = f"audio:chunk:{chunk_id}"
        return self.client.get(key)
    
    def set_processing_status(self, chunk_id: int, status: str, expiry: int = 3600):
        """Set processing status for a chunk"""
        key = f"processing:chunk:{chunk_id}"
        self.client.setex(key, expiry, status)
    
    def get_processing_status(self, chunk_id: int) -> Optional[str]:
        """Get processing status for a chunk"""
        key = f"processing:chunk:{chunk_id}"
        return self.client.get(key)
    
    def delete_processing_status(self, chunk_id: int):
        """Delete processing status for a chunk"""
        key = f"processing:chunk:{chunk_id}"
        self.client.delete(key)
    
    def set_buffer_queue(self, document_id: int, chunk_ids: list, expiry: int = 3600):
        """Set buffer queue for a document"""
        key = f"buffer:document:{document_id}"
        self.client.setex(key, expiry, json.dumps(chunk_ids))
    
    def get_buffer_queue(self, document_id: int) -> list:
        """Get buffer queue for a document"""
        key = f"buffer:document:{document_id}"
        data = self.client.get(key)
        return json.loads(data) if data else []
    
    def update_buffer_queue(self, document_id: int, chunk_ids: list, expiry: int = 3600):
        """Update buffer queue for a document"""
        self.set_buffer_queue(document_id, chunk_ids, expiry)


redis_service = RedisService()