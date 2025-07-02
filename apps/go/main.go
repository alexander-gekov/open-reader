package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/alexandergekov/open-reader/apps/go/tts"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/ledongthuc/pdf"
	"github.com/sentencizer/sentencizer"
)

func loadEnv() {
	// Get the directory where the executable is running
	dir, err := os.Getwd()
	if err != nil {
		log.Printf("Warning: Could not get working directory: %v", err)
		return
	}

	// Try to find env file in current directory or apps/go
	envPaths := []string{
		path.Join(dir, "env"),
		path.Join(dir, "apps", "go", "env"),
	}

	var envFile string
	for _, path := range envPaths {
		if _, err := os.Stat(path); err == nil {
			envFile = path
			break
		}
	}

	if envFile == "" {
		log.Printf("Warning: Could not find env file in paths: %v", envPaths)
		return
	}

	content, err := os.ReadFile(envFile)
	if err != nil {
		log.Printf("Warning: Error loading env file %s: %v", envFile, err)
		return
	}

	log.Printf("Loading environment from: %s", envFile)

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		
		// Remove any quotes from the value
		value = strings.Trim(value, `"'`)
		
		os.Setenv(key, value)
		if key == "ELEVENLABS_API_KEY" {
			log.Printf("Successfully loaded ElevenLabs API key: %s", value[:10] + "...")
		}
	}
}

type TTSRequest struct {
	Text    string `json:"text"`
	ModelID string `json:"model_id"`
}

type TTSResponse struct {
	AudioURL string `json:"audio_url"`
	AudioData []byte `json:"audio_data"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type ChunkProcessor struct {
	mutex       sync.RWMutex
	chunks      []string
	currentIdx  int
	audioFiles  map[int]string // Map of chunk index to audio file path
	processing  map[int]bool
	client      *http.Client
	apiKey      string
	lastError   string
	audioCache  map[string][]byte
	stopProcess chan bool      // Channel to stop processing
	filename    string        // Store the current file's name
	settings    TTSSettings   // Store TTS settings
}

type UploadResponse struct {
	Message string   `json:"message"`
	Chunks  []string `json:"chunks"`
	AudioID string   `json:"audio_id"`
}

type TTSSettings struct {
	Provider string `json:"provider"`
	APIKey   string `json:"apiKey"`
	Model    string `json:"model"`
	Voice    string `json:"voice"`
}

func main() {
	loadEnv()

	// Get API key but don't require it
	apiKey := os.Getenv("ELEVENLABS_API_KEY")

	// Create audio directory if it doesn't exist
	if err := os.MkdirAll("./uploads/audio", 0755); err != nil {
		log.Fatal("Failed to create audio directory:", err)
	}

	// Initialize the processor
	processor = &ChunkProcessor{
		audioFiles: make(map[int]string),
		processing: make(map[int]bool),
		client:     &http.Client{Timeout: 30 * time.Second},
		apiKey:     apiKey, // This is now optional
		audioCache: make(map[string][]byte),
	}

	// Start a goroutine to periodically clean old cache entries
	go func() {
		for {
			time.Sleep(24 * time.Hour) // Clean cache every 24 hours
			log.Printf("Cleaning audio cache...")
			processor.mutex.Lock()
			processor.audioCache = make(map[string][]byte)
			processor.mutex.Unlock()
		}
	}()

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Serve static files from the uploads directory under a different path
	r.Static("/static/audio", "./uploads/audio")

	r.POST("/upload", uploadHandler)
	r.GET("/audio/status/:chunk", getAudioStatusHandler)
	r.GET("/status", statusHandler)
	r.GET("/health", healthHandler)
	r.GET("/test-audio", testAudioHandler)
	r.GET("/start-next/:chunk", startNextChunkHandler)
	r.GET("/settings", func(c *gin.Context) {
		// Get settings from request header
		settings := TTSSettings{
			Provider: c.GetHeader("X-TTS-Provider"),
			APIKey:   c.GetHeader("X-TTS-API-Key"),
			Model:    c.GetHeader("X-TTS-Model"),
			Voice:    c.GetHeader("X-TTS-Voice"),
		}
		c.JSON(http.StatusOK, settings)
	})
	r.POST("/generate-audio", func(c *gin.Context) {
		var req struct {
			Text     string      `json:"text"`
			Settings TTSSettings `json:"settings"`
			Filename string      `json:"filename"`
			Chunk    int         `json:"chunk"`
		}

		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if req.Filename == "" {
			req.Filename = fmt.Sprintf("generated_%d", time.Now().Unix())
		}

		options := map[string]string{
			"model": req.Settings.Model,
			"voice": req.Settings.Voice,
			"filename": req.Filename,
			"chunk": fmt.Sprintf("%d", req.Chunk),
		}
		audioData, err := generateAudio(req.Text, req.Settings, options)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.Header("Content-Type", "audio/mpeg")
		c.Header("Content-Length", fmt.Sprintf("%d", len(audioData)))
		c.Header("Cache-Control", "no-cache")
		c.Data(http.StatusOK, "audio/mpeg", audioData)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(r.Run(":" + port))
}

var processor *ChunkProcessor

func extractTextFromPDF(filepath string) (string, error) {
	f, r, err := pdf.Open(filepath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var text strings.Builder
	for i := 1; i <= r.NumPage(); i++ {
		p := r.Page(i)
		if p.V.IsNull() {
			continue
		}

		content, err := p.GetPlainText(nil)
		if err != nil {
			continue
		}
		
		// Clean up the text
		// Replace multiple spaces with a single space
		content = regexp.MustCompile(`\s+`).ReplaceAllString(content, " ")
		// Add space after period if missing
		content = regexp.MustCompile(`\.(\S)`).ReplaceAllString(content, ". $1")
		// Add space after comma if missing
		content = regexp.MustCompile(`,(\S)`).ReplaceAllString(content, ", $1")
		// Add space after colon if missing
		content = regexp.MustCompile(`\:(\S)`).ReplaceAllString(content, ": $1")
		// Add space after semicolon if missing
		content = regexp.MustCompile(`\;(\S)`).ReplaceAllString(content, "; $1")
		// Fix spaces around parentheses
		content = regexp.MustCompile(`\s*\(\s*`).ReplaceAllString(content, " (")
		content = regexp.MustCompile(`\s*\)\s*`).ReplaceAllString(content, ") ")
		// Fix spaces around brackets
		content = regexp.MustCompile(`\s*\[\s*`).ReplaceAllString(content, " [")
		content = regexp.MustCompile(`\s*\]\s*`).ReplaceAllString(content, "] ")
		
		text.WriteString(content)
		text.WriteString("\n") // Add newline between pages
	}

	return strings.TrimSpace(text.String()), nil
}

func chunkText(text string) []string {
	// Clean up the text first
	text = strings.TrimSpace(text)
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")

	// Initialize the segmenter with English language
	segmenter := sentencizer.NewSegmenter("en")

	// Split text into sentences using Sentencizer
	sentences := segmenter.Segment(text)

	var allChunks []string
	var currentChunk strings.Builder
	wordCount := 0
	const maxWordsPerChunk = 50 // Keep the same word limit for TTS optimization

	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if sentence == "" {
			continue
		}

		words := strings.Fields(sentence)
		
		// If a single sentence is longer than maxWordsPerChunk, split it
		if len(words) > maxWordsPerChunk {
			// First, add any existing chunk
			if currentChunk.Len() > 0 {
				chunk := strings.TrimSpace(currentChunk.String())
				if chunk != "" {
					allChunks = append(allChunks, chunk)
				}
				currentChunk.Reset()
				wordCount = 0
			}

			// Then split the long sentence into chunks
			for i := 0; i < len(words); i += maxWordsPerChunk {
				end := i + maxWordsPerChunk
				if end > len(words) {
					end = len(words)
				}
				subChunk := strings.Join(words[i:end], " ")
				// Only add ellipsis if this is not the end of the sentence
				if end < len(words) {
					subChunk += "..."
				}
				allChunks = append(allChunks, subChunk)
			}
			continue
		}

		// Start a new chunk if adding this sentence would exceed the word limit
		if wordCount + len(words) > maxWordsPerChunk {
			chunk := strings.TrimSpace(currentChunk.String())
			if chunk != "" {
				allChunks = append(allChunks, chunk)
			}
			currentChunk.Reset()
			wordCount = 0
		}

		// Add the sentence to the current chunk
		if wordCount > 0 {
			currentChunk.WriteString(" ")
		}
		currentChunk.WriteString(sentence)
		wordCount += len(words)
	}

	// Add any remaining text as a chunk
	if currentChunk.Len() > 0 {
		chunk := strings.TrimSpace(currentChunk.String())
		if chunk != "" {
			allChunks = append(allChunks, chunk)
		}
	}

	return allChunks
}

func uploadHandler(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}
	defer file.Close()

	if !strings.HasSuffix(strings.ToLower(header.Filename), ".pdf") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only PDF files are allowed"})
		return
	}

	uploadsDir := "./uploads"
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create uploads directory"})
		return
	}

	// Create a clean filename without extension and special characters
	baseFilename := strings.TrimSuffix(header.Filename, ".pdf")
	cleanFilename := regexp.MustCompile(`[^a-zA-Z0-9]+`).ReplaceAllString(baseFilename, "_")
	filepath := path.Join(uploadsDir, cleanFilename+".pdf")

	if _, err := os.Stat(filepath); err == nil {
		log.Printf("File %s already exists, reusing it", filepath)
	} else {
		out, err := os.Create(filepath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
			return
		}
		defer out.Close()

		_, err = io.Copy(out, file)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
			return
		}
	}

	text, err := extractTextFromPDF(filepath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to extract text from PDF"})
		return
	}

	chunks := chunkText(text)
	if len(chunks) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No text found in PDF"})
		return
	}

	// Get TTS settings from headers
	settings := TTSSettings{
		Provider: c.GetHeader("X-TTS-Provider"),
		APIKey:   c.GetHeader("X-TTS-API-Key"),
		Model:    c.GetHeader("X-TTS-Model"),
		Voice:    c.GetHeader("X-TTS-Voice"),
	}

	// Validate required settings
	if settings.Provider == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "TTS provider is required",
		})
		return
	}

	// Only require API key for non-fallback providers
	if settings.Provider != "fallback" && settings.APIKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "API key is required for non-fallback providers",
		})
		return
	}

	audioID := processor.ProcessChunks(chunks, cleanFilename, settings)

	c.JSON(http.StatusOK, UploadResponse{
		Message: "PDF processed successfully",
		Chunks:  chunks,
		AudioID: audioID,
	})

	// Schedule file cleanup after 24 hours
	go func() {
		time.Sleep(24 * time.Hour)
		os.Remove(filepath)
	}()
}

func getAudioStatusHandler(c *gin.Context) {
	chunkStr := c.Param("chunk")
	if chunkStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "No chunk index provided",
		})
		return
	}

	chunkIndex := 0
	if _, err := fmt.Sscanf(chunkStr, "%d", &chunkIndex); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Invalid chunk index",
		})
		return
	}

	processor.mutex.RLock()
	defer processor.mutex.RUnlock()

	// Check if we have an audio file for this chunk
	audioPath, exists := processor.audioFiles[chunkIndex]
	if exists {
		// Return the full URL path
		fullURL := fmt.Sprintf("http://localhost:8080/static/audio/%s", path.Base(audioPath))
		c.JSON(http.StatusOK, gin.H{
			"status":    "ready",
			"url":       fullURL,
			"hasNext":   chunkIndex + 1 < len(processor.chunks),
			"nextReady": processor.audioFiles[chunkIndex + 1] != "",
		})
		return
	}

	// Check if the chunk is being processed
	if processor.processing[chunkIndex] {
		c.JSON(http.StatusOK, gin.H{
			"status": "processing",
		})
		return
	}

	c.JSON(http.StatusNotFound, gin.H{
		"status": "error",
		"error":  "Audio not found",
	})
}

func statusHandler(c *gin.Context) {
	processor.mutex.RLock()
	hasMore := processor.currentIdx < len(processor.chunks)
	currentIdx := processor.currentIdx
	totalChunks := len(processor.chunks)
	lastError := processor.lastError
	processor.mutex.RUnlock()

	c.JSON(http.StatusOK, gin.H{
		"hasMore":     hasMore,
		"currentIdx":  currentIdx,
		"totalChunks": totalChunks,
		"error":       lastError,
	})
}

func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}

func testAudioHandler(c *gin.Context) {
	testText := "Hello! This is a test of the text-to-speech system. How does it sound?"

	// Get settings from headers
	settings := TTSSettings{
		Provider: c.GetHeader("X-TTS-Provider"),
		APIKey:   c.GetHeader("X-TTS-API-Key"),
		Model:    c.GetHeader("X-TTS-Model"),
		Voice:    c.GetHeader("X-TTS-Voice"),
	}

	// Check cache first
	processor.mutex.RLock()
	if cachedAudio, exists := processor.audioCache[testText]; exists {
		log.Printf("Using cached test audio")
		c.Header("Content-Type", "audio/mpeg")
		c.Header("Content-Length", fmt.Sprintf("%d", len(cachedAudio)))
		c.Header("Cache-Control", "no-cache")
		c.Data(http.StatusOK, "audio/mpeg", cachedAudio)
		processor.mutex.RUnlock()
		return
	}
	processor.mutex.RUnlock()

	options := map[string]string{
		"model": settings.Model,
		"voice": settings.Voice,
		"filename": "test_audio",
		"chunk": "1", // Start from chunk 1 for testing
	}
	audioData, err := generateAudio(testText, settings, options)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Test audio failed: %v", err)})
		return
	}

	// Store in cache
	processor.mutex.Lock()
	processor.audioCache[testText] = audioData
	processor.mutex.Unlock()

	c.Header("Content-Type", "audio/mpeg")
	c.Header("Content-Length", fmt.Sprintf("%d", len(audioData)))
	c.Header("Cache-Control", "no-cache")
	c.Data(http.StatusOK, "audio/mpeg", audioData)
}

func (cp *ChunkProcessor) ProcessChunks(chunks []string, filename string, settings TTSSettings) string {
	cp.mutex.Lock()
	cp.chunks = chunks
	cp.currentIdx = 0
	cp.audioFiles = make(map[int]string)
	cp.processing = make(map[int]bool)
	cp.lastError = ""
	cp.stopProcess = make(chan bool, 1)
	cp.filename = filename
	cp.settings = settings
	cp.mutex.Unlock()

	audioID := cp.filename

	// Start processing the first pair of chunks
	go func() {
		cp.generateTTS(0)
		if len(cp.chunks) > 1 {
			cp.generateTTS(1)
		}
	}()

	return audioID
}

func (cp *ChunkProcessor) generateTTS(index int) {
	cp.mutex.Lock()
	if index >= len(cp.chunks) {
		cp.mutex.Unlock()
		return
	}

	// Check if audio file already exists
	expectedFileName := fmt.Sprintf("%s_chunk_%d.mp3", cp.filename, index)
	expectedFilePath := path.Join("uploads", "audio", expectedFileName)
	
	if _, err := os.Stat(expectedFilePath); err == nil {
		// File exists, reuse it
		cp.audioFiles[index] = expectedFilePath
		delete(cp.processing, index)
		cp.mutex.Unlock()
		return
	}

	// Mark this chunk as being processed
	cp.processing[index] = true
	cp.mutex.Unlock()

	text := cp.chunks[index]
	options := map[string]string{
		"model": cp.settings.Model,
		"voice": cp.settings.Voice,
		"filename": cp.filename,
		"chunk": fmt.Sprintf("%d", index),
	}

	audioData, err := generateAudio(text, cp.settings, options)
	if err != nil {
		cp.mutex.Lock()
		cp.lastError = err.Error()
		delete(cp.processing, index)
		cp.mutex.Unlock()
		return
	}

	// Save the audio file
	audioPath := path.Join("uploads", "audio", expectedFileName)
	if err := os.WriteFile(audioPath, audioData, 0644); err != nil {
		cp.mutex.Lock()
		cp.lastError = fmt.Sprintf("failed to save audio file: %v", err)
		delete(cp.processing, index)
		cp.mutex.Unlock()
		return
	}

	cp.mutex.Lock()
	cp.audioFiles[index] = audioPath
	delete(cp.processing, index)
	cp.mutex.Unlock()
}

func (cp *ChunkProcessor) callElevenLabsTTS(text string) ([]byte, error) {
	reqBody := TTSRequest{
		Text:    text,
		ModelID: "eleven_flash_v2_5",
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	url := fmt.Sprintf("https://api.elevenlabs.io/v1/text-to-speech/%s/stream?output_format=mp3_44100_128", "cgSgspJ2msm6clMCkdW9")
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	if cp.apiKey == "" {
		return nil, fmt.Errorf("ElevenLabs API key is not set")
	}

	req.Header.Set("xi-api-key", cp.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := cp.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("TTS API error (HTTP %d): %s", resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}

// New endpoint to trigger processing of next chunk
func startNextChunkHandler(c *gin.Context) {
	chunkStr := c.Param("chunk")
	if chunkStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No chunk index provided",
		})
		return
	}

	currentChunk := 0
	if _, err := fmt.Sscanf(chunkStr, "%d", &currentChunk); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid chunk index",
		})
		return
	}

	processor.mutex.Lock()
	nextChunk := currentChunk + 1
	if nextChunk >= len(processor.chunks) {
		processor.mutex.Unlock()
		c.JSON(http.StatusOK, gin.H{
			"message": "No more chunks to process",
		})
		return
	}

	// Check if next chunk is already processed or being processed
	if _, exists := processor.audioFiles[nextChunk]; exists {
		processor.mutex.Unlock()
		c.JSON(http.StatusOK, gin.H{
			"message": "Next chunk already processed",
		})
		return
	}
	if processor.processing[nextChunk] {
		processor.mutex.Unlock()
		c.JSON(http.StatusOK, gin.H{
			"message": "Next chunk already being processed",
		})
		return
	}

	// Start processing current chunk if needed and next chunk
	processor.mutex.Unlock()
	go func() {
		// First check if we need to generate the current chunk
		if _, exists := processor.audioFiles[currentChunk]; !exists && !processor.processing[currentChunk] {
			processor.generateTTS(currentChunk)
		}
		// Then generate the next chunk
		processor.generateTTS(nextChunk)
		// Process the chunk after next if it exists
		if nextChunk+1 < len(processor.chunks) {
			processor.generateTTS(nextChunk + 1)
		}
	}()

	c.JSON(http.StatusOK, gin.H{
		"message": "Started processing chunks",
	})
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func generateAudio(text string, settings TTSSettings, options map[string]string) ([]byte, error) {
	provider := tts.NewTTSProvider(settings.Provider, settings.APIKey)
	return provider.GenerateAudio(text, options)
} 