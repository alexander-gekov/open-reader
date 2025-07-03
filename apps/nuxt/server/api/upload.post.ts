import { getAuth } from "@clerk/nuxt/server";
import { prisma } from "~/lib/prisma";
import { PDFDocument } from "pdf-lib";
import { s3Client, getBucketName } from "~/lib/s3";
import { PutObjectCommand } from "@aws-sdk/client-s3";
import { getGoBackendUrl } from "~/lib/utils";

interface UploadResponse {
  success: boolean;
  message: string;
  chunks: string[];
  audio_id: string;
}

export default defineEventHandler(async (event) => {
  const { userId } = await getAuth(event);

  if (!userId) {
    throw createError({
      statusCode: 401,
      message: "Unauthorized",
    });
  }

  const formData = await readMultipartFormData(event);

  if (!formData || formData.length === 0) {
    throw createError({
      statusCode: 400,
      message: "No file uploaded",
    });
  }

  const file = formData[0];

  if (!file.type || !file.type.includes("pdf")) {
    throw createError({
      statusCode: 400,
      message: "Invalid file type. Please upload a PDF file.",
    });
  }

  try {
    // Load PDF document to get total pages
    const pdfDoc = await PDFDocument.load(file.data);
    const totalPages = pdfDoc.getPageCount();

    // Generate unique keys for S3
    const timestamp = Date.now();
    const pdfKey = `pdfs/${userId}/${timestamp}-${file.filename}`;
    const bucketName = getBucketName();

    // Upload PDF to S3
    await s3Client.send(
      new PutObjectCommand({
        Bucket: bucketName,
        Key: pdfKey,
        Body: file.data,
        ContentType: file.type || "application/pdf",
      })
    );

    // Store PDF metadata in the database
    const pdfRecord = await prisma.pDF.create({
      data: {
        userId,
        title: file.filename || "Untitled",
        url: `https://${bucketName}.s3.amazonaws.com/${pdfKey}`,
        totalPages,
        isArchived: false,
      },
    });

    // Process for TTS
    const goBackendFormData = new FormData();
    const blob = new Blob([file.data], {
      type: file.type || "application/pdf",
    });
    goBackendFormData.append("file", blob, file.filename);

    const headers = event.node.req.headers;
    const ttsHeaders: Record<string, string> = {
      "X-TTS-Provider": (headers["x-tts-provider"] as string) || "",
      "X-TTS-API-Key": (headers["x-tts-api-key"] as string) || "",
      "X-TTS-Model": (headers["x-tts-model"] as string) || "",
      "X-TTS-Voice": (headers["x-tts-voice"] as string) || "",
    };

    // Send to Go backend for processing
    const goBackendUrl = getGoBackendUrl();
    const response = await $fetch<UploadResponse>(`${goBackendUrl}/upload`, {
      method: "POST",
      body: goBackendFormData,
      headers: ttsHeaders,
    });

    return {
      success: true,
      message: "File uploaded and processed successfully",
      chunks: response.chunks,
      audioId: response.audio_id,
      totalChunks: response.chunks.length,
    };
  } catch (error: any) {
    console.error("Error processing file:", error);
    throw createError({
      statusCode: error.statusCode || 500,
      message: error.message || "Failed to process file",
    });
  }
});
