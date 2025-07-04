import { prisma } from "~/lib/prisma";

export default defineEventHandler(async (event) => {
  const params = event.context.params as { id: string };
  const { id } = params;
  const { audioUrl } = await readBody(event);

  await prisma.pdfChunk.update({
    where: { id },
    data: { audioUrl },
  });

  return { success: true };
});
