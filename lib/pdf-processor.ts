import * as pdfParse from 'pdf-parse'

export interface ProcessedPDF {
  text: string
  totalPages: number
  chunks: TextChunk[]
}

export interface TextChunk {
  index: number
  text: string
  wordCount: number
  startPage?: number
  endPage?: number
}

export interface ChunkingOptions {
  maxWordsPerChunk: number
  overlapWords: number
}

const DEFAULT_CHUNKING_OPTIONS: ChunkingOptions = {
  maxWordsPerChunk: 200,
  overlapWords: 20
}

export class PDFProcessor {
  static async parsePDF(buffer: Buffer): Promise<{ text: string; totalPages: number }> {
    try {
      const data = await pdfParse(buffer)
      return {
        text: data.text,
        totalPages: data.numpages
      }
    } catch (error) {
      throw new Error(`Failed to parse PDF: ${error instanceof Error ? error.message : 'Unknown error'}`)
    }
  }

  static splitIntoSentences(text: string): string[] {
    // Split by sentence endings, but preserve the punctuation
    const sentences = text
      .split(/(?<=[.!?])\s+/)
      .filter(sentence => sentence.trim().length > 0)
      .map(sentence => sentence.trim())

    return sentences
  }

  static chunkSentences(sentences: string[], options: ChunkingOptions = DEFAULT_CHUNKING_OPTIONS): TextChunk[] {
    const chunks: TextChunk[] = []
    let currentChunk: string[] = []
    let currentWordCount = 0
    let chunkIndex = 0

    for (const sentence of sentences) {
      const words = sentence.split(/\s+/).filter(word => word.length > 0)
      const sentenceWordCount = words.length

      // If adding this sentence would exceed the limit, create a new chunk
      if (currentWordCount + sentenceWordCount > options.maxWordsPerChunk && currentChunk.length > 0) {
        // Create chunk
        chunks.push({
          index: chunkIndex++,
          text: currentChunk.join(' '),
          wordCount: currentWordCount
        })

        // Start new chunk with overlap
        const overlapSentences = this.getOverlapSentences(currentChunk, options.overlapWords)
        currentChunk = [...overlapSentences, sentence]
        currentWordCount = this.countWords(currentChunk.join(' '))
      } else {
        // Add sentence to current chunk
        currentChunk.push(sentence)
        currentWordCount += sentenceWordCount
      }
    }

    // Add the last chunk if it has content
    if (currentChunk.length > 0) {
      chunks.push({
        index: chunkIndex,
        text: currentChunk.join(' '),
        wordCount: currentWordCount
      })
    }

    return chunks
  }

  private static getOverlapSentences(sentences: string[], overlapWords: number): string[] {
    let wordCount = 0
    const overlapSentences: string[] = []

    // Start from the end and work backwards
    for (let i = sentences.length - 1; i >= 0; i--) {
      const sentenceWordCount = this.countWords(sentences[i])
      if (wordCount + sentenceWordCount <= overlapWords) {
        overlapSentences.unshift(sentences[i])
        wordCount += sentenceWordCount
      } else {
        break
      }
    }

    return overlapSentences
  }

  private static countWords(text: string): number {
    return text.split(/\s+/).filter(word => word.length > 0).length
  }

  static async processPDF(buffer: Buffer, options?: Partial<ChunkingOptions>): Promise<ProcessedPDF> {
    const chunkingOptions = { ...DEFAULT_CHUNKING_OPTIONS, ...options }
    
    // Parse PDF
    const { text, totalPages } = await this.parsePDF(buffer)
    
    // Split into sentences
    const sentences = this.splitIntoSentences(text)
    
    // Create chunks
    const chunks = this.chunkSentences(sentences, chunkingOptions)
    
    return {
      text,
      totalPages,
      chunks
    }
  }
}