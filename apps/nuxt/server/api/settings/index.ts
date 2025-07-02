import { prisma } from "~/lib/prisma";

export default defineEventHandler(async (event) => {
  const { userId } = event.context.auth();

  if (!userId) {
    throw createError({
      statusCode: 401,
      message: "Unauthorized",
    });
  }

  // Get TTS settings from database
  const ttsSettings = await prisma.tTSSettings.findUnique({
    where: { userId },
  });

  // Handle GET request
  if (event.method === "GET") {
    return ttsSettings;
  }

  // Handle POST request
  if (event.method === "POST") {
    const body = await readBody(event);

    // Validate required fields
    if (!body.provider) {
      throw createError({
        statusCode: 400,
        message: "Provider is required",
      });
    }

    // Only require API key for non-fallback providers
    if (body.provider !== "fallback" && !body.apiKey) {
      throw createError({
        statusCode: 400,
        message: "API key is required for this provider",
      });
    }

    // Update or create TTS settings
    const updatedSettings = await prisma.tTSSettings.upsert({
      where: { userId },
      update: {
        provider: body.provider,
        apiKey: body.provider === "fallback" ? "" : body.apiKey,
        model: body.model || null,
        voice: body.voice || null,
      },
      create: {
        userId,
        provider: body.provider,
        apiKey: body.provider === "fallback" ? "" : body.apiKey,
        model: body.model || null,
        voice: body.voice || null,
      },
    });

    return updatedSettings;
  }

  throw createError({
    statusCode: 405,
    message: "Method not allowed",
  });
});
