export default defineEventHandler(async (event) => {
  const path = getRouterParam(event, "path");

  try {
    const response = await fetch(`http://localhost:8080/static/audio/${path}`);

    if (!response.ok) {
      throw new Error(`Go backend error: ${response.status}`);
    }

    // Get the audio data as an array buffer
    const audioData = await response.arrayBuffer();

    // Set the correct headers
    event.node.res.setHeader("Content-Type", "audio/mpeg");
    event.node.res.setHeader("Content-Length", audioData.byteLength);
    event.node.res.setHeader("Cache-Control", "no-cache");

    // Send the audio data
    return Buffer.from(audioData);
  } catch (error) {
    console.error("Error fetching audio file:", error);
    throw createError({
      statusCode: 500,
      message:
        error instanceof Error ? error.message : "Failed to fetch audio file",
    });
  }
});
