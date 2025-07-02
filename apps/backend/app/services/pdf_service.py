import PyPDF2
import re
import nltk
from typing import List
from io import BytesIO

try:
    nltk.data.find('tokenizers/punkt')
except LookupError:
    nltk.download('punkt')


class PDFService:
    def __init__(self):
        pass
    
    def extract_text_from_pdf(self, pdf_content: bytes) -> str:
        """Extract text from PDF bytes"""
        try:
            pdf_reader = PyPDF2.PdfReader(BytesIO(pdf_content))
            text = ""
            
            for page in pdf_reader.pages:
                text += page.extract_text() + "\n"
            
            return text.strip()
        except Exception as e:
            print(f"Error extracting text from PDF: {e}")
            return ""
    
    def clean_text(self, text: str) -> str:
        """Clean and normalize text"""
        text = re.sub(r'\s+', ' ', text)
        text = re.sub(r'\n\s*\n', '\n', text)
        text = text.replace('\n', ' ')
        text = re.sub(r'[^\w\s.,!?;:()-]', '', text)
        
        return text.strip()
    
    def chunk_text_by_sentences(self, text: str) -> List[str]:
        """Split text into sentence-based chunks"""
        cleaned_text = self.clean_text(text)
        
        sentences = nltk.sent_tokenize(cleaned_text)
        
        chunks = []
        current_chunk = ""
        max_chunk_length = 500
        
        for sentence in sentences:
            sentence = sentence.strip()
            if not sentence:
                continue
            
            if len(current_chunk) + len(sentence) + 1 <= max_chunk_length:
                if current_chunk:
                    current_chunk += " " + sentence
                else:
                    current_chunk = sentence
            else:
                if current_chunk:
                    chunks.append(current_chunk)
                current_chunk = sentence
        
        if current_chunk:
            chunks.append(current_chunk)
        
        return [chunk for chunk in chunks if len(chunk.strip()) > 10]
    
    def process_pdf(self, pdf_content: bytes) -> List[str]:
        """Process PDF and return text chunks"""
        text = self.extract_text_from_pdf(pdf_content)
        if not text:
            return []
        
        chunks = self.chunk_text_by_sentences(text)
        return chunks


pdf_service = PDFService()