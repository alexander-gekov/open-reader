import { Redis } from 'ioredis'

let redis: Redis | null = null

export function getRedisClient(): Redis {
  if (!redis) {
    const redisUrl = process.env.REDIS_URL || 'redis://localhost:6379'
    redis = new Redis(redisUrl)
  }
  return redis
}

export interface ChunkCacheData {
  audio_url?: string
  status: 'processing' | 'completed' | 'failed'
  created_at: string
  updated_at?: string
}

export class ChunkCache {
  private redis: Redis

  constructor() {
    this.redis = getRedisClient()
  }

  private getKey(docId: string, chunkIndex: number): string {
    return `chunk:${docId}:${chunkIndex}`
  }

  async get(docId: string, chunkIndex: number): Promise<ChunkCacheData | null> {
    const key = this.getKey(docId, chunkIndex)
    const data = await this.redis.get(key)
    return data ? JSON.parse(data) : null
  }

  async set(docId: string, chunkIndex: number, data: ChunkCacheData): Promise<void> {
    const key = this.getKey(docId, chunkIndex)
    await this.redis.set(key, JSON.stringify(data), 'EX', 86400) // 24 hours TTL
  }

  async setProcessing(docId: string, chunkIndex: number): Promise<void> {
    await this.set(docId, chunkIndex, {
      status: 'processing',
      created_at: new Date().toISOString()
    })
  }

  async setCompleted(docId: string, chunkIndex: number, audioUrl: string): Promise<void> {
    await this.set(docId, chunkIndex, {
      audio_url: audioUrl,
      status: 'completed',
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString()
    })
  }

  async setFailed(docId: string, chunkIndex: number): Promise<void> {
    await this.set(docId, chunkIndex, {
      status: 'failed',
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString()
    })
  }

  async delete(docId: string, chunkIndex: number): Promise<void> {
    const key = this.getKey(docId, chunkIndex)
    await this.redis.del(key)
  }
}