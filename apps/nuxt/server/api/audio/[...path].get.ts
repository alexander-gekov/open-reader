export default defineEventHandler(async (event) => {
  const path = getRouterParam(event, "path");

  try {
    const response = await fetch(`http://localhost:8080/static/audio/${path}`);

    if (!response.ok) {
      throw new Error(`Go backend error: ${response.status}`);
    }

    // Get the original headers
    const contentType = response.headers.get("Content-Type") || "audio/mpeg";
    const contentLength = response.headers.get("Content-Length");

    // Set response headers
    setHeader(event, "Content-Type", contentType);
    if (contentLength) {
      setHeader(event, "Content-Length", parseInt(contentLength));
    }
    setHeader(event, "Cache-Control", "no-cache");

    // Return the raw response body as a stream
    return response.body;
  } catch (error) {
    console.error("Error fetching audio file:", error);
    throw createError({
      statusCode: 500,
      message:
        error instanceof Error ? error.message : "Failed to fetch audio file",
    });
  }
});
