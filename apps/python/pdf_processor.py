import io
import re
import fitz  # PyMuPDF
from PyPDF2 import PdfReader
from typing import List
import nltk
from nltk.tokenize import sent_tokenize
import logging

logger = logging.getLogger(__name__)

class PDFProcessor:
    def __init__(self):
        try:
            nltk.data.find('tokenizers/punkt')
        except LookupError:
            nltk.download('punkt', quiet=True)
    
    def extract_text_from_pdf(self, pdf_content: bytes) -> str:
        """
        Extract text from PDF using PyMuPDF first, fallback to PyPDF2.
        """
        try:
            # Try PyMuPDF first (better OCR and text extraction)
            doc = fitz.open(stream=pdf_content, filetype="pdf")
            text = ""
            
            for page_num in range(len(doc)):
                page = doc[page_num]
                page_text = page.get_text()
                if page_text.strip():
                    text += page_text + "\n"
            
            doc.close()
            
            if text.strip():
                return self._clean_text(text)
            
        except Exception as e:
            logger.warning(f"PyMuPDF failed, trying PyPDF2: {str(e)}")
        
        try:
            # Fallback to PyPDF2
            reader = PdfReader(io.BytesIO(pdf_content))
            text = ""
            
            for page in reader.pages:
                page_text = page.extract_text()
                if page_text:
                    text += page_text + "\n"
            
            if text.strip():
                return self._clean_text(text)
            
        except Exception as e:
            logger.error(f"PyPDF2 also failed: {str(e)}")
            raise ValueError("Could not extract text from PDF")
        
        raise ValueError("No text found in PDF")
    
    def get_page_count(self, pdf_content: bytes) -> int:
        """
        Get the number of pages in the PDF.
        """
        try:
            doc = fitz.open(stream=pdf_content, filetype="pdf")
            page_count = len(doc)
            doc.close()
            return page_count
        except Exception:
            try:
                reader = PdfReader(io.BytesIO(pdf_content))
                return len(reader.pages)
            except Exception as e:
                logger.error(f"Could not get page count: {str(e)}")
                return 1
    
    def _clean_text(self, text: str) -> str:
        """
        Clean and normalize text extracted from PDF.
        """
        # Remove excessive whitespace
        text = re.sub(r'\s+', ' ', text)
        
        # Fix common PDF extraction issues
        text = re.sub(r'([a-z])([A-Z])', r'\1 \2', text)
        text = re.sub(r'([a-zA-Z])(\d)', r'\1 \2', text)
        text = re.sub(r'(\d)([a-zA-Z])', r'\1 \2', text)
        
        # Fix punctuation spacing
        text = re.sub(r'\.(\S)', r'. \1', text)
        text = re.sub(r',(\S)', r', \1', text)
        text = re.sub(r':(\S)', r': \1', text)
        text = re.sub(r';(\S)', r'; \1', text)
        
        # Fix quotes
        text = re.sub(r'"(\S)', r'" \1', text)
        text = re.sub(r'(\S)"', r'\1 "', text)
        
        # Remove extra spaces and normalize
        text = re.sub(r'\s+', ' ', text).strip()
        
        return text
    
    def chunk_text(self, text: str, max_words_per_chunk: int = 50) -> List[str]:
        """
        Split text into chunks using NLTK sentence tokenization.
        """
        if not text.strip():
            return []
        
        # Split into sentences using NLTK
        sentences = sent_tokenize(text)
        
        chunks = []
        current_chunk = ""
        current_word_count = 0
        
        for sentence in sentences:
            sentence = sentence.strip()
            if not sentence:
                continue
            
            words = sentence.split()
            sentence_word_count = len(words)
            
            # If a single sentence is too long, split it
            if sentence_word_count > max_words_per_chunk:
                # Save current chunk if it exists
                if current_chunk:
                    chunks.append(current_chunk.strip())
                    current_chunk = ""
                    current_word_count = 0
                
                # Split the long sentence
                for i in range(0, len(words), max_words_per_chunk):
                    chunk_words = words[i:i + max_words_per_chunk]
                    chunk_text = " ".join(chunk_words)
                    
                    # Add ellipsis if this is not the end of the sentence
                    if i + max_words_per_chunk < len(words):
                        chunk_text += "..."
                    
                    chunks.append(chunk_text)
                
                continue
            
            # Check if adding this sentence would exceed the word limit
            if current_word_count + sentence_word_count > max_words_per_chunk and current_chunk:
                chunks.append(current_chunk.strip())
                current_chunk = sentence
                current_word_count = sentence_word_count
            else:
                if current_chunk:
                    current_chunk += " " + sentence
                else:
                    current_chunk = sentence
                current_word_count += sentence_word_count
        
        # Add the last chunk if it exists
        if current_chunk:
            chunks.append(current_chunk.strip())
        
        return chunks