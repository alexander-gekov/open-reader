import { Redis } from "@upstash/redis";

const config = useRuntimeConfig();

export const redis = new Redis({
  url: config.upstashRedisRestUrl,
  token: config.upstashRedisRestToken,
});
