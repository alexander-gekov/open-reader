from fastapi import APIRouter, UploadFile, File, HTTPException, Depends
from sqlalchemy.orm import Session
import uuid
from datetime import datetime

from app.database import get_db
from app.models.document import Document
from app.models.schemas import Document as DocumentSchema
from app.services.s3_service import s3_service

router = APIRouter()


@router.post("/upload", response_model=DocumentSchema)
async def upload_file(
    file: UploadFile = File(...),
    db: Session = Depends(get_db)
):
    """Upload a PDF file to R2 storage and create database record"""
    
    if not file.content_type or not file.content_type.startswith("application/pdf"):
        raise HTTPException(status_code=400, detail="Only PDF files are allowed")
    
    if not file.size or file.size > 50 * 1024 * 1024:  # 50MB limit
        raise HTTPException(status_code=400, detail="File size must be less than 50MB")
    
    try:
        file_content = await file.read()
        
        file_key = f"pdfs/{uuid.uuid4()}-{file.filename}"
        
        success = s3_service.upload_file(
            file_content=file_content,
            key=file_key,
            content_type=file.content_type
        )
        
        if not success:
            raise HTTPException(status_code=500, detail="Failed to upload file to storage")
        
        document = Document(
            filename=file.filename,
            file_key=file_key,
            content_type=file.content_type,
            file_size=file.size,
            status="uploaded"
        )
        
        db.add(document)
        db.commit()
        db.refresh(document)
        
        return document
        
    except Exception as e:
        db.rollback()
        raise HTTPException(status_code=500, detail=f"Upload failed: {str(e)}")


@router.get("/documents", response_model=list[DocumentSchema])
async def get_documents(db: Session = Depends(get_db)):
    """Get all uploaded documents"""
    documents = db.query(Document).order_by(Document.created_at.desc()).all()
    return documents


@router.get("/documents/{document_id}", response_model=DocumentSchema)
async def get_document(document_id: int, db: Session = Depends(get_db)):
    """Get a specific document by ID"""
    document = db.query(Document).filter(Document.id == document_id).first()
    if not document:
        raise HTTPException(status_code=404, detail="Document not found")
    return document