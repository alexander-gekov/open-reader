from fastapi import APIRouter, HTTPException, Depends, BackgroundTasks
from sqlalchemy.orm import Session
from typing import List

from app.database import get_db
from app.models.document import Document, TextChunk
from app.models.schemas import ProcessPDFResponse, TextChunk as TextChunkSchema
from app.services.pdf_service import pdf_service
from app.services.s3_service import s3_service

router = APIRouter()


@router.post("/pdf/process/{document_id}", response_model=ProcessPDFResponse)
async def process_pdf(
    document_id: int,
    background_tasks: BackgroundTasks,
    db: Session = Depends(get_db)
):
    """Process a PDF document and extract text chunks"""
    
    document = db.query(Document).filter(Document.id == document_id).first()
    if not document:
        raise HTTPException(status_code=404, detail="Document not found")
    
    if document.status not in ["uploaded", "error"]:
        raise HTTPException(status_code=400, detail="Document is already being processed or completed")
    
    document.status = "processing"
    db.commit()
    
    background_tasks.add_task(process_pdf_background, document_id, db)
    
    return ProcessPDFResponse(
        document_id=document_id,
        total_chunks=0,
        message="PDF processing started"
    )


async def process_pdf_background(document_id: int, db: Session):
    """Background task to process PDF"""
    try:
        document = db.query(Document).filter(Document.id == document_id).first()
        if not document:
            return
        
        pdf_content = s3_service.download_file(document.file_key)
        if not pdf_content:
            document.status = "error"
            db.commit()
            return
        
        chunks = pdf_service.process_pdf(pdf_content)
        
        existing_chunks = db.query(TextChunk).filter(TextChunk.document_id == document_id).all()
        for chunk in existing_chunks:
            db.delete(chunk)
        
        for index, chunk_text in enumerate(chunks):
            text_chunk = TextChunk(
                document_id=document_id,
                chunk_index=index,
                content=chunk_text,
                has_audio=False
            )
            db.add(text_chunk)
        
        document.status = "completed"
        db.commit()
        
    except Exception as e:
        print(f"Error processing PDF {document_id}: {e}")
        document = db.query(Document).filter(Document.id == document_id).first()
        if document:
            document.status = "error"
            db.commit()


@router.get("/pdf/{document_id}/chunks", response_model=List[TextChunkSchema])
async def get_document_chunks(document_id: int, db: Session = Depends(get_db)):
    """Get all text chunks for a document"""
    
    document = db.query(Document).filter(Document.id == document_id).first()
    if not document:
        raise HTTPException(status_code=404, detail="Document not found")
    
    chunks = db.query(TextChunk).filter(
        TextChunk.document_id == document_id
    ).order_by(TextChunk.chunk_index).all()
    
    return chunks


@router.get("/pdf/{document_id}/status")
async def get_processing_status(document_id: int, db: Session = Depends(get_db)):
    """Get the processing status of a document"""
    
    document = db.query(Document).filter(Document.id == document_id).first()
    if not document:
        raise HTTPException(status_code=404, detail="Document not found")
    
    chunk_count = db.query(TextChunk).filter(TextChunk.document_id == document_id).count()
    
    return {
        "document_id": document_id,
        "status": document.status,
        "total_chunks": chunk_count,
        "filename": document.filename
    }