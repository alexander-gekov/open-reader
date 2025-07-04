from pydantic import BaseModel
from typing import Optional, List
from datetime import datetime

class TTSSettings(BaseModel):
    provider: str
    apiKey: str
    model: Optional[str] = None
    voice: Optional[str] = None

class UploadResponse(BaseModel):
    success: bool
    message: str
    chunks: List[str]
    audioId: str
    totalChunks: int

class AudioStatusResponse(BaseModel):
    status: str
    url: Optional[str] = None
    hasNext: Optional[bool] = None
    nextReady: Optional[bool] = None
    error: Optional[str] = None

class AudioGenerationRequest(BaseModel):
    text: str
    settings: TTSSettings
    filename: str
    chunk: int

class PDFDocumentResponse(BaseModel):
    id: str
    title: str
    url: str
    coverUrl: Optional[str] = None
    totalPages: int
    isArchived: bool
    createdAt: datetime
    updatedAt: datetime

class ChunkResponse(BaseModel):
    id: str
    pdfId: str
    index: int
    text: str
    audioUrl: Optional[str] = None
    status: str
    error_message: Optional[str] = None
    createdAt: datetime
    updatedAt: datetime