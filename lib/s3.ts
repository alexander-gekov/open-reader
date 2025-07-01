import { S3Client } from "@aws-sdk/client-s3";

const config = useRuntimeConfig();

export const s3 = new S3Client({
  region: "auto",
  endpoint: config.r2Endpoint,
  credentials: {
    accessKeyId: config.r2AccessKeyId,
    secretAccessKey: config.r2SecretAccessKey,
  },
});
