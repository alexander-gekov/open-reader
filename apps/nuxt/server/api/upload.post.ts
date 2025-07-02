export default defineEventHandler(async (event) => {
  const formData = await readMultipartFormData(event);

  if (!formData || formData.length === 0) {
    throw createError({
      statusCode: 400,
      message: "No file uploaded",
    });
  }

  const file = formData[0];

  if (!file.filename?.toLowerCase().endsWith(".pdf")) {
    throw createError({
      statusCode: 400,
      message: "Only PDF files are allowed",
    });
  }

  try {
    const goBackendFormData = new FormData();
    const blob = new Blob([file.data], {
      type: file.type || "application/pdf",
    });
    goBackendFormData.append("file", blob, file.filename);

    // Forward TTS settings headers
    const headers: Record<string, string> = {};
    const provider = getHeader(event, "X-TTS-Provider");
    const apiKey = getHeader(event, "X-TTS-API-Key");
    const model = getHeader(event, "X-TTS-Model");
    const voice = getHeader(event, "X-TTS-Voice");

    if (provider) headers["X-TTS-Provider"] = provider;
    if (apiKey) headers["X-TTS-API-Key"] = apiKey;
    if (model) headers["X-TTS-Model"] = model;
    if (voice) headers["X-TTS-Voice"] = voice;

    const response = await fetch("http://localhost:8080/upload", {
      method: "POST",
      body: goBackendFormData,
      headers,
    });

    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(`Go backend error: ${response.status} - ${errorText}`);
    }

    const result = await response.json();

    return {
      success: true,
      message: result.message,
      chunks: result.chunks,
      audioId: result.audio_id,
      totalChunks: result.chunks.length,
    };
  } catch (error) {
    console.error("Upload to Go backend failed:", error);
    throw createError({
      statusCode: 500,
      message: error instanceof Error ? error.message : "Failed to process PDF",
    });
  }
});
