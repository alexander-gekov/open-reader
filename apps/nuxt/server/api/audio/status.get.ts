export default defineEventHandler(async (event) => {
  try {
    const response = await fetch("http://localhost:8080/status");

    if (!response.ok) {
      throw new Error(`Go backend error: ${response.status}`);
    }

    const result = await response.json();
    return result;
  } catch (error) {
    console.error("Error fetching status:", error);
    throw createError({
      statusCode: 500,
      message:
        error instanceof Error ? error.message : "Failed to fetch status",
    });
  }
});
