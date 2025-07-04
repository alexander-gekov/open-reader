from sqlalchemy import Column, Integer, String, Boolean, DateTime, Text, ForeignKey
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import relationship
from sqlalchemy.dialects.postgresql import UUID
from datetime import datetime
import uuid

Base = declarative_base()

class PDFDocument(Base):
    __tablename__ = "pdfs"
    
    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4, index=True)
    title = Column(String, nullable=False)
    url = Column(String, nullable=False)
    coverUrl = Column(String, nullable=True)
    totalPages = Column(Integer, nullable=False)
    isArchived = Column(Boolean, default=False)
    createdAt = Column(DateTime, default=datetime.utcnow)
    updatedAt = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)
    
    chunks = relationship("Chunk", back_populates="pdf", cascade="all, delete-orphan")

class Chunk(Base):
    __tablename__ = "chunks"
    
    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4, index=True)
    pdfId = Column(UUID(as_uuid=True), ForeignKey("pdfs.id"), nullable=False)
    index = Column(Integer, nullable=False)
    text = Column(Text, nullable=False)
    audioUrl = Column(String, nullable=True)
    status = Column(String, default="pending")  # pending, processing, completed, error
    error_message = Column(Text, nullable=True)
    createdAt = Column(DateTime, default=datetime.utcnow)
    updatedAt = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)
    
    pdf = relationship("PDFDocument", back_populates="chunks")

class TTSSettings(Base):
    __tablename__ = "tts_settings"
    
    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4, index=True)
    userId = Column(String, unique=True, nullable=False)
    provider = Column(String, default="elevenlabs")
    apiKey = Column(String, nullable=False)
    model = Column(String, nullable=True)
    voice = Column(String, nullable=True)
    createdAt = Column(DateTime, default=datetime.utcnow)
    updatedAt = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

class WaitlistEntry(Base):
    __tablename__ = "waitlist_entries"
    
    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4, index=True)
    email = Column(String, unique=True, nullable=False)
    name = Column(String, nullable=True)
    createdAt = Column(DateTime, default=datetime.utcnow)