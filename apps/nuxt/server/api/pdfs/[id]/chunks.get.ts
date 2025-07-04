import { getAuth } from "@clerk/nuxt/server";
import { prisma } from "~/lib/prisma";

export default defineEventHandler(async (event) => {
  const { userId } = await getAuth(event);
  if (!userId) {
    throw createError({ statusCode: 401, message: "Unauthorized" });
  }

  const pdfId = event.context.params?.id;
  if (!pdfId) {
    throw createError({ statusCode: 400, message: "Missing PDF id" });
  }

  // Check PDF ownership
  const pdf = await prisma.pdf.findUnique({
    where: { id: pdfId },
  });
  if (!pdf || pdf.userId !== userId) {
    throw createError({ statusCode: 404, message: "PDF not found" });
  }

  // Fetch chunks
  const chunks = await prisma.pdfChunk.findMany({
    where: { pdfId },
    orderBy: { index: "asc" },
    select: { id: true, text: true, audioUrl: true, index: true },
  });

  return { chunks };
});
