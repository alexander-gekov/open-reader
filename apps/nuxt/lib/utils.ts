import { type ClassValue, clsx } from "clsx";
import { twMerge } from "tailwind-merge";
import { redis } from "~/lib/redis";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function chunkText(text: string, maxWordsPerChunk = 50): string[] {
  const segmenter = new Intl.Segmenter("en", { granularity: "sentence" });
  const sentences = [...segmenter.segment(text)].map((s) => s.segment.trim());

  const chunks: string[] = [];
  let current = "";

  for (const sentence of sentences) {
    if (!sentence) continue;

    const potentialChunk = current ? `${current} ${sentence}` : sentence;
    const wordCount = potentialChunk.split(/\s+/).length;

    if (wordCount > maxWordsPerChunk && current) {
      chunks.push(current.trim());
      current = sentence;
    } else {
      current = potentialChunk;
    }
  }

  if (current) {
    chunks.push(current.trim());
  }

  return chunks;
}

export async function storeTextChunks(docId: string, chunks: string[]) {
  const pipeline = redis.pipeline();

  chunks.forEach((chunk, index) => {
    const key = `text:${docId}:${index}`;
    pipeline.set(key, chunk, { ex: 86400 }); // 24 hour expiry
  });

  await pipeline.exec();

  return chunks.length;
}
