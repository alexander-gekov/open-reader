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

  if (event.method === "GET") {
    // List PDFs
    const pdfs = await prisma.pdf.findMany({
      where: {
        userId,
        isArchived: false,
      },
      orderBy: {
        updatedAt: "desc",
      },
    });

    return pdfs;
  }

  throw createError({
    statusCode: 405,
    message: "Method not allowed",
  });
});
