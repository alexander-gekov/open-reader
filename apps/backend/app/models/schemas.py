from pydantic import BaseModel
from typing import List, Optional
from datetime import datetime


class DocumentBase(BaseModel):
    filename: str
    file_key: str
    content_type: str
    file_size: int


class DocumentCreate(DocumentBase):
    pass


class TextChunkBase(BaseModel):
    chunk_index: int
    content: str


class TextChunk(TextChunkBase):
    id: int
    document_id: int
    has_audio: bool
    audio_file_key: Optional[str] = None
    created_at: datetime

    class Config:
        from_attributes = True


class Document(DocumentBase):
    id: int
    status: str
    created_at: datetime
    updated_at: datetime
    chunks: List[TextChunk] = []

    class Config:
        from_attributes = True


class ProcessPDFResponse(BaseModel):
    document_id: int
    total_chunks: int
    message: str


class TTSRequest(BaseModel):
    chunk_id: int


class TTSResponse(BaseModel):
    chunk_id: int
    audio_url: str
    cached: bool