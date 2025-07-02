export default defineEventHandler(async (event) => {
  try {
    const response = await fetch("http://localhost:8080/audio/next");

    if (response.status === 200) {
      const audioData = await response.arrayBuffer();

      setHeader(event, "Content-Type", "audio/mpeg");
      setHeader(event, "Content-Disposition", "attachment; filename=audio.mp3");

      return new Uint8Array(audioData);
    } else if (response.status === 202) {
      const result = await response.json();
      return {
        waiting: true,
        message: result.message,
        retry: result.retry,
      };
    } else if (response.status === 204) {
      return {
        completed: true,
        message: "All chunks processed",
      };
    }

    throw new Error("Unexpected response from TTS service");
  } catch (error) {
    console.error("Error fetching audio:", error);
    throw createError({
      statusCode: 500,
      message: "Failed to fetch audio",
    });
  }
});
