import { prisma } from "~/lib/prisma";

export default defineEventHandler(async (event) => {
  const chunk = getRouterParam(event, "chunk");
  // Assume pdfId is passed as a query param for now
  const pdfId = getQuery(event).pdfId as string | undefined;

  try {
    const response = await fetch(`http://localhost:8080/audio/status/${chunk}`);

    if (!response.ok) {
      throw new Error(`Go backend error: ${response.status}`);
    }

    const data = await response.json();

    // If audio is ready and url is present, patch the DB
    if (data.status === "ready" && data.url && pdfId) {
      await prisma.pdfChunk.updateMany({
        where: { pdfId, index: Number(chunk) },
        data: { audioUrl: data.url },
      });
    }

    return data;
  } catch (error) {
    console.error("Error fetching audio status:", error);
    throw createError({
      statusCode: 500,
      message:
        error instanceof Error ? error.message : "Failed to fetch audio status",
    });
  }
});
