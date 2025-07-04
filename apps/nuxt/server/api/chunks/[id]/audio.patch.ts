import { getAuth } from "@clerk/nuxt/server";
import { prisma } from "~/lib/prisma";

export default defineEventHandler(async (event) => {
  const { userId } = await getAuth(event);
  if (!userId) {
    throw createError({ statusCode: 401, message: "Unauthorized" });
  }

  const chunkId = event.context.params?.id;
  if (!chunkId) {
    throw createError({ statusCode: 400, message: "Missing chunk id" });
  }

  const body = await readBody(event);
  if (!body.audioUrl) {
    throw createError({ statusCode: 400, message: "Missing audioUrl in body" });
  }

  // Find the chunk and its parent PDF
  const chunk = await prisma.pdfChunk.findUnique({
    where: { id: chunkId },
    include: { pdf: true },
  });
  if (!chunk || !chunk.pdf || chunk.pdf.userId !== userId) {
    throw createError({
      statusCode: 404,
      message: "Chunk not found or unauthorized",
    });
  }

  // Update the audioUrl
  const updatedChunk = await prisma.pdfChunk.update({
    where: { id: chunkId },
    data: { audioUrl: body.audioUrl },
    select: { id: true, text: true, audioUrl: true, index: true },
  });

  return { chunk: updatedChunk };
});
