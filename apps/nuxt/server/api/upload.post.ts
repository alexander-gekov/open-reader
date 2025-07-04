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
    // Remove pdf-lib usage and text extraction

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
    const pdfRecord = await prisma.pdf.create({
      data: {
        userId,
        title: file.filename || "Untitled",
        url: `https://${bucketName}.s3.amazonaws.com/${pdfKey}`,
        totalPages: 0, // Will update after Go returns chunk info if needed
        isArchived: false,
      },
    });

    // Prepare form data for Go backend
    const goBackendFormData = new FormData();
    const blob = new Blob([file.data], {
      type: file.type || "application/pdf",
    });
    goBackendFormData.append("file", blob, file.filename);

    // Add chunkIds as a JSON string (empty for now, or you can pass [] if needed)
    goBackendFormData.append("chunkIds", JSON.stringify([]));

    const headers = event.node.req.headers;
    const ttsHeaders: Record<string, string> = {
      "X-TTS-Provider": (headers["x-tts-provider"] as string) || "",
      "X-TTS-API-Key": (headers["x-tts-api-key"] as string) || "",
      "X-TTS-Model": (headers["x-tts-model"] as string) || "",
      "X-TTS-Voice": (headers["x-tts-voice"] as string) || "",
    };

    // Send to Go backend for processing
    const goBackendUrl = getGoBackendUrl();
    const response = await $fetch<{
      message: string;
      chunks: string[];
      audio_id: string;
    }>(`${goBackendUrl}/upload`, {
      method: "POST",
      body: goBackendFormData,
      headers: ttsHeaders,
    });

    // Save real chunks to DB
    const chunkRecords = await prisma.pdfChunk.createMany({
      data: response.chunks.map((text, idx) => ({
        pdfId: pdfRecord.id,
        index: idx,
        text,
        audioUrl: null,
      })),
    });

    // Fetch the created chunks to return to the frontend
    const savedChunks = await prisma.pdfChunk.findMany({
      where: { pdfId: pdfRecord.id },
      orderBy: { index: "asc" },
      select: { id: true, text: true, audioUrl: true, index: true },
    });

    return {
      success: true,
      message: "File uploaded and processed successfully",
      pdfId: pdfRecord.id,
      chunks: savedChunks,
      audioId: response.audio_id,
      totalChunks: savedChunks.length,
    };
  } catch (error: any) {
    console.error("Error processing file:", error);
    throw createError({
      statusCode: error.statusCode || 500,
      message: error.message || "Failed to process file",
    });
  }
});
