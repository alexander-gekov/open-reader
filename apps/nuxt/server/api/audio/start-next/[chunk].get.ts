export default defineEventHandler(async (event) => {
  const chunk = getRouterParam(event, "chunk");

  try {
    const response = await fetch(`http://localhost:8080/start-next/${chunk}`);

    if (!response.ok) {
      throw new Error(`Go backend error: ${response.status}`);
    }

    return response.json();
  } catch (error) {
    console.error("Error starting next chunk:", error);
    throw createError({
      statusCode: 500,
      message:
        error instanceof Error ? error.message : "Failed to start next chunk",
    });
  }
});
