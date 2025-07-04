package main

import (
	"bytes"
	"context"
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
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/ledongthuc/pdf"
	"github.com/lucsky/cuid"
	"github.com/sentencizer/sentencizer"
)

var db *pgx.Conn

func initDB() error {
	dbUrl := os.Getenv("DATABASE_URL")
	if dbUrl == "" {
		return fmt.Errorf("DATABASE_URL is not set")
	}
	var err error
	db, err = pgx.Connect(context.Background(), dbUrl)
	return err
}

func loadEnv() {
	// Get the directory where the executable is running
	dir, err := os.Getwd()
	if err != nil {
		log.Printf("Warning: Could not get working directory: %v", err)
		return
	}

	log.Printf("Current working directory: %s", dir)

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
		
		// Log environment variables being set (but mask sensitive values)
		if strings.Contains(strings.ToLower(key), "key") || strings.Contains(strings.ToLower(key), "secret") {
			log.Printf("Set %s=***masked***", key)
		} else {
			log.Printf("Set %s=%s", key, value)
		}
	}

	// Log final AWS-related environment variables
	log.Printf("Final AWS configuration:")
	log.Printf("AWS_REGION=%s", os.Getenv("AWS_REGION"))
	log.Printf("AWS_S3_BUCKET=%s", os.Getenv("AWS_S3_BUCKET"))
	log.Printf("AWS_ACCESS_KEY=%s", maskString(os.Getenv("AWS_ACCESS_KEY")))
	log.Printf("AWS_SECRET_ACCESS_KEY=%s", maskString(os.Getenv("AWS_SECRET_ACCESS_KEY")))
}

// Helper function to mask sensitive strings
func maskString(s string) string {
	if len(s) == 0 {
		return ""
	}
	if len(s) <= 4 {
		return "****"
	}
	return s[:4] + "****"
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
	s3Client    *s3.S3       // AWS S3 client
	bucketName  string       // AWS S3 bucket name
	chunkIDs    []string     // Store chunk DB IDs from Nuxt (unused now)
	pdfId       string       // Store the current pdfId
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

	// Get AWS configuration
	awsRegion := os.Getenv("AWS_REGION")
	if awsRegion == "" {
		awsRegion = "us-east-1" // Default region
	}

	bucketName := os.Getenv("AWS_S3_BUCKET")
	if bucketName == "" {
		log.Fatal("AWS_S3_BUCKET environment variable is required")
	}

	log.Printf("AWS Configuration - Region: %s, Bucket: %s", awsRegion, bucketName)

	// Initialize AWS session
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(awsRegion),
		Credentials: credentials.NewStaticCredentials(
			os.Getenv("AWS_ACCESS_KEY"),
			os.Getenv("AWS_SECRET_ACCESS_KEY"),
			"",
		),
	}))

	// Initialize S3 client
	s3Client := s3.New(sess)

	// Test S3 connection and bucket access
	_, err := s3Client.HeadBucket(&s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		log.Fatalf("Failed to access S3 bucket %s: %v", bucketName, err)
	}

	log.Printf("Successfully connected to S3 bucket: %s", bucketName)

	// Initialize the processor
	processor = &ChunkProcessor{
		audioFiles:  make(map[int]string),
		processing:  make(map[int]bool),
		client:     &http.Client{Timeout: 30 * time.Second},
		apiKey:     apiKey,
		audioCache: make(map[string][]byte),
		s3Client:   s3Client,
		bucketName: bucketName, // Make sure bucketName is set
	}

	// Log processor initialization
	log.Printf("Initialized processor with bucket: %s", processor.bucketName)

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

	if err := initDB(); err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

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
	r.POST("/save-chunks", func(c *gin.Context) {
		var req struct {
			PdfId  string `json:"pdfId"`
			Chunks []struct {
				Text  string `json:"text"`
				Index int    `json:"index"`
			} `json:"chunks"`
		}
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Forward to Nuxt backend (assume http://localhost:3000/api/chunks)
		nuxtUrl := os.Getenv("NUXT_CHUNKS_URL")
		if nuxtUrl == "" {
			nuxtUrl = "http://localhost:3000/api/chunks" // fallback
		}
		payload, _ := json.Marshal(req)
		resp, err := http.Post(nuxtUrl, "application/json", bytes.NewReader(payload))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		c.Data(resp.StatusCode, "application/json", body)
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
		// // First, handle cases where words are incorrectly joined
		// // Look for patterns of lowercase followed by uppercase and add a space
		// content = regexp.MustCompile(`([a-z])([A-Z])`).ReplaceAllString(content, "$1 $2")
		
		// // Look for patterns of letter followed by number or vice versa and add a space
		// content = regexp.MustCompile(`([a-zA-Z])(\d)`).ReplaceAllString(content, "$1 $2")
		// content = regexp.MustCompile(`(\d)([a-zA-Z])`).ReplaceAllString(content, "$1 $2")
		
		// // Replace multiple spaces with a single space
		// content = regexp.MustCompile(`\s+`).ReplaceAllString(content, " ")
		
		// // Add space after period if missing
		// content = regexp.MustCompile(`\.(\S)`).ReplaceAllString(content, ". $1")
		
		// // Add space after comma if missing
		// content = regexp.MustCompile(`,(\S)`).ReplaceAllString(content, ", $1")
		
		// // Add space after colon if missing
		// content = regexp.MustCompile(`\:(\S)`).ReplaceAllString(content, ": $1")
		
		// // Add space after semicolon if missing
		// content = regexp.MustCompile(`\;(\S)`).ReplaceAllString(content, "; $1")
		
		// // Fix spaces around parentheses
		// content = regexp.MustCompile(`\s*\(\s*`).ReplaceAllString(content, " (")
		// content = regexp.MustCompile(`\s*\)\s*`).ReplaceAllString(content, ") ")
		
		// // Fix spaces around brackets
		// content = regexp.MustCompile(`\s*\[\s*`).ReplaceAllString(content, " [")
		// content = regexp.MustCompile(`\s*\]\s*`).ReplaceAllString(content, "] ")
		
		// // Fix spaces around special characters
		// content = regexp.MustCompile(`([a-zA-Z])([.,!?;:])`).ReplaceAllString(content, "$1$2 ")
		
		// // Fix spaces around quotes
		// content = regexp.MustCompile(`"(\S)`).ReplaceAllString(content, `" $1`)
		// content = regexp.MustCompile(`(\S)"`).ReplaceAllString(content, `$1 "`)
		
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

	// Read pdfId from form (sent as a string)
	pdfId := c.Request.FormValue("pdfId")
	if pdfId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "pdfId is required"})
		return
	}

	processor.pdfId = pdfId

	audioID := processor.ProcessChunks(chunks, cleanFilename, settings)

	for idx, text := range chunks {
		_, err := db.Exec(context.Background(),
			`INSERT INTO pdf_chunks (id, pdf_id, index, text, audio_url, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			cuid.New(), pdfId, idx, text, nil, time.Now(), time.Now(),
		)
		if err != nil {
			log.Printf("Failed to insert chunk: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save chunks to DB"})
			return
		}
	}

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

	// First check if the chunk index is valid
	if chunkIndex < 0 || chunkIndex >= len(processor.chunks) {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Chunk index out of bounds",
		})
		return
	}

	// Check if we have an audio file for this chunk
	audioURL, exists := processor.audioFiles[chunkIndex]
	if exists {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ready",
			"url":       audioURL,
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

	// If we get here, the chunk exists but hasn't started processing yet
	c.JSON(http.StatusOK, gin.H{
		"status": "pending",
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

	// Upload to S3
	testFileName := fmt.Sprintf("test_audio_%d.mp3", time.Now().Unix())
	audioURL, err := processor.uploadToS3(audioData, testFileName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Failed to upload test audio: %v", err)})
		return
	}

	// Store in cache
	processor.mutex.Lock()
	processor.audioCache[testText] = audioData
	processor.mutex.Unlock()

	// Redirect to the S3 URL
	c.Redirect(http.StatusTemporaryRedirect, audioURL)
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

	// Mark this chunk as being processed
	cp.processing[index] = true
	text := cp.chunks[index]
	settings := cp.settings
	pdfId := cp.pdfId
	cp.mutex.Unlock()

	options := map[string]string{
		"model": settings.Model,
		"voice": settings.Voice,
		"filename": cp.filename,
		"chunk": fmt.Sprintf("%d", index),
	}

	audioData, err := generateAudio(text, settings, options)
	if err != nil {
		cp.mutex.Lock()
		cp.lastError = err.Error()
		delete(cp.processing, index)
		cp.mutex.Unlock()
		log.Printf("Error generating audio for chunk %d: %v", index, err)
		return
	}

	// Upload to S3
	expectedFileName := fmt.Sprintf("%s_chunk_%d.mp3", cp.filename, index)
	audioURL, err := cp.uploadToS3(audioData, expectedFileName)
	if err != nil {
		cp.mutex.Lock()
		cp.lastError = fmt.Sprintf("failed to upload audio to S3: %v", err)
		delete(cp.processing, index)
		cp.mutex.Unlock()
		log.Printf("Error uploading audio for chunk %d: %v", index, err)
		return
	}

	cp.mutex.Lock()
	cp.audioFiles[index] = audioURL // Store the full S3 URL
	delete(cp.processing, index)
	cp.mutex.Unlock()
	log.Printf("Successfully generated and uploaded audio for chunk %d", index)

	// Update the audioUrl in the database for this chunk
	if db != nil {
		_, err := db.Exec(context.Background(),
			`UPDATE pdf_chunks SET audio_url = $1, updated_at = $2 WHERE pdf_id = $3 AND index = $4`,
			audioURL, time.Now(), pdfId, index,
		)
		if err != nil {
			log.Printf("Failed to update audio_url for chunk %d: %v", index, err)
		}
	}
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

	// Validate chunk index
	if currentChunk < 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Chunk index cannot be negative",
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

	processor.mutex.Unlock()

	// Start processing in a goroutine
	go func() {
		// First check if we need to generate the current chunk
		if currentChunk >= 0 && currentChunk < len(processor.chunks) {
			if _, exists := processor.audioFiles[currentChunk]; !exists && !processor.processing[currentChunk] {
				processor.generateTTS(currentChunk)
			}
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

func (cp *ChunkProcessor) uploadToS3(audioData []byte, filename string) (string, error) {
	// Validate inputs
	if cp.bucketName == "" {
		log.Printf("Error: S3 bucket name is empty")
		return "", fmt.Errorf("S3 bucket name is not configured")
	}

	if cp.s3Client == nil {
		log.Printf("Error: S3 client is not initialized")
		return "", fmt.Errorf("S3 client is not initialized")
	}

	if len(audioData) == 0 {
		log.Printf("Error: No audio data to upload")
		return "", fmt.Errorf("no audio data to upload")
	}

	log.Printf("Uploading %d bytes to S3 bucket '%s' with key 'audio/%s'", len(audioData), cp.bucketName, filename)

	input := &s3.PutObjectInput{
		Bucket:      aws.String(cp.bucketName),
		Key:         aws.String(fmt.Sprintf("audio/%s", filename)),
		Body:        bytes.NewReader(audioData),
		ContentType: aws.String("audio/mpeg"),
	}

	// Log the actual values being used in the PutObject call
	log.Printf("S3 PutObject Input - Bucket: %s, Key: audio/%s", *input.Bucket, filename)

	_, err := cp.s3Client.PutObject(input)
	if err != nil {
		log.Printf("Failed to upload to S3: %v", err)
		return "", fmt.Errorf("failed to upload to S3: %v", err)
	}

	log.Printf("Successfully uploaded audio file to S3: %s", filename)

	// Return the S3 URL
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/audio/%s", 
		cp.bucketName,
		*cp.s3Client.Config.Region,
		filename), nil
}

// Helper function to get bucket names
func getBucketNames(buckets []*s3.Bucket) []string {
	names := make([]string, len(buckets))
	for i, bucket := range buckets {
		if bucket.Name != nil {
			names[i] = *bucket.Name
		}
	}
	return names
}

// Add a new method to ChunkProcessor to accept chunkIDs
func (cp *ChunkProcessor) ProcessChunksWithIDs(chunks []string, filename string, settings TTSSettings, chunkIDs []string) string {
	cp.mutex.Lock()
	cp.chunks = chunks
	cp.currentIdx = 0
	cp.audioFiles = make(map[int]string)
	cp.processing = make(map[int]bool)
	cp.lastError = ""
	cp.stopProcess = make(chan bool, 1)
	cp.filename = filename
	cp.settings = settings
	cp.chunkIDs = chunkIDs
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