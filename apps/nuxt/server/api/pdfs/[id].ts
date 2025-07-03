import { getAuth } from "@clerk/nuxt/server";
import { prisma } from "~/lib/prisma";

export default defineEventHandler(async (event) => {
  const { userId } = await getAuth(event);

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

  // Check if the PDF exists and belongs to the user
  const pdf = await prisma.PDF.findFirst({
    where: {
      id,
      userId,
    },
  });

  if (!pdf) {
    throw createError({
      statusCode: 404,
      message: "PDF not found",
    });
  }

  if (event.method === "GET") {
    return pdf;
  }

  if (event.method === "PATCH") {
    const body = await readBody(event);

    const updatedPdf = await prisma.PDF.update({
      where: { id },
      data: {
        title: body.title,
        description: body.description,
        lastReadPage: body.lastReadPage,
        lastReadAt: body.lastReadPage !== undefined ? new Date() : undefined,
        isArchived: body.isArchived,
        metadata: body.metadata,
      },
    });

    return updatedPdf;
  }

  if (event.method === "DELETE") {
    // Instead of actually deleting, we archive the PDF
    const archivedPdf = await prisma.PDF.update({
      where: { id },
      data: {
        isArchived: true,
      },
    });

    return archivedPdf;
  }

  throw createError({
    statusCode: 405,
    message: "Method not allowed",
  });
});
