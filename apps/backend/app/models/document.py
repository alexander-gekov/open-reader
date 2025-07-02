from sqlalchemy import Column, Integer, String, Text, DateTime, ForeignKey, Boolean
from sqlalchemy.orm import relationship
from sqlalchemy.sql import func
from app.database import Base


class Document(Base):
    __tablename__ = "documents"
    
    id = Column(Integer, primary_key=True, index=True)
    filename = Column(String, nullable=False)
    file_key = Column(String, unique=True, nullable=False)
    content_type = Column(String, nullable=False)
    file_size = Column(Integer, nullable=False)
    status = Column(String, default="uploaded")  # uploaded, processing, completed, error
    created_at = Column(DateTime, server_default=func.now())
    updated_at = Column(DateTime, server_default=func.now(), onupdate=func.now())
    
    chunks = relationship("TextChunk", back_populates="document", cascade="all, delete-orphan")


class TextChunk(Base):
    __tablename__ = "text_chunks"
    
    id = Column(Integer, primary_key=True, index=True)
    document_id = Column(Integer, ForeignKey("documents.id"), nullable=False)
    chunk_index = Column(Integer, nullable=False)
    content = Column(Text, nullable=False)
    has_audio = Column(Boolean, default=False)
    audio_file_key = Column(String, nullable=True)
    created_at = Column(DateTime, server_default=func.now())
    
    document = relationship("Document", back_populates="chunks")