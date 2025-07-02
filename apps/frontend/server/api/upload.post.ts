import { ListBucketsCommand, PutObjectCommand } from "@aws-sdk/client-s3";
import { s3 } from "~/lib/s3";

export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig();
  const formData = await readMultipartFormData(event);

  if (!formData || formData.length === 0) {
    throw createError({
      statusCode: 400,
      message: "No file uploaded",
    });
  }

  const file = formData[0];
  const key = `${Date.now()}-${file.filename}`;

  try {
    await s3.send(
      new PutObjectCommand({
        Bucket: "open-reader",
        Key: key,
        Body: file.data,
        ContentType: file.type || "application/octet-stream",
      })
    );

    return {
      success: true,
      key,
    };
  } catch (error) {
    console.error("Upload failed:", error);
    if (error instanceof Error) {
      console.error("Upload failed:", error.message);
    }
    throw createError({
      statusCode: 500,
      message: "Failed to upload file",
    });
  }
});
