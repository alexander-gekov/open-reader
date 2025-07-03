import { getAuth } from "@clerk/nuxt/server";
import { prisma } from "~/lib/prisma";
import { s3Client } from "~/lib/s3";
import { PutObjectCommand } from "@aws-sdk/client-s3";

export default defineEventHandler(async (event) => {
  const { userId } = await getAuth(event);
  const config = useRuntimeConfig();

  if (!userId) {
    throw createError({
      statusCode: 401,
      message: "Unauthorized",
    });
  }

  const id = event.context.params?.id;
  if (!id) {
    throw createError({
      statusCode: 400,
      message: "PDF ID is required",
    });
  }

  // Check if PDF exists and belongs to user
  const pdf = await prisma.pDF.findUnique({
    where: { id },
  });

  if (!pdf) {
    throw createError({
      statusCode: 404,
      message: "PDF not found",
    });
  }

  if (pdf.userId !== userId) {
    throw createError({
      statusCode: 403,
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

  if (!file.type || !file.type.startsWith("image/")) {
    throw createError({
      statusCode: 400,
      message: "Invalid file type. Please upload an image file.",
    });
  }

  try {
    // Upload cover image to S3
    const coverKey = `covers/${userId}/${Date.now()}-${file.filename}`;
    const bucketName = process.env.AWS_BUCKET_NAME;

    await s3Client.send(
      new PutObjectCommand({
        Bucket: bucketName,
        Key: coverKey,
        Body: file.data,
        ContentType: file.type,
        ACL: "public-read",
      })
    );

    // Update PDF record with new cover URL
    const updatedPdf = await prisma.pDF.update({
      where: { id },
      data: {
        coverUrl: `https://${bucketName}.s3.amazonaws.com/${coverKey}`,
      },
    });

    return {
      success: true,
      message: "Cover image uploaded successfully",
      pdf: updatedPdf,
    };
  } catch (error: any) {
    console.error("Error uploading cover image:", error);
    throw createError({
      statusCode: 500,
      message: error.message || "Failed to upload cover image",
    });
  }
});
