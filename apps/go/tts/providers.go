package tts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/polly"
	htgotts "github.com/hegedustibor/htgo-tts"
	"github.com/hegedustibor/htgo-tts/voices"
)

// GetGoogleTTSURL generates a URL for Google TTS service
func GetGoogleTTSURL(text, lang string) string {
	baseURL := "https://translate.google.com/translate_tts"
	params := url.Values{}
	params.Add("ie", "UTF-8")
	params.Add("tl", lang)
	params.Add("client", "tw-ob")
	params.Add("q", text)
	return baseURL + "?" + params.Encode()
}

// TTSProvider defines the interface for text-to-speech providers
type TTSProvider interface {
	GenerateAudio(text string, options map[string]string) ([]byte, error)
}

// ElevenLabsProvider implements TTSProvider for ElevenLabs
type ElevenLabsProvider struct {
	apiKey string
}

// PollyProvider implements TTSProvider using Amazon Polly
type PollyProvider struct {
	client *polly.Polly
}

// CartesiaTTSProvider implements TTSProvider using Cartesia's API
type CartesiaTTSProvider struct {
	folder string
	apiKey string
	rateLimiter *time.Ticker
	processing bool
	mutex sync.Mutex
}

// HTGoTTSProvider implements TTSProvider using htgo-tts
type HTGoTTSProvider struct {
	folder string
}

// NewCartesiaTTSProvider creates a new CartesiaTTSProvider instance
func NewCartesiaTTSProvider(folder string, apiKey string) *CartesiaTTSProvider {
	return &CartesiaTTSProvider{
		folder: folder,
		apiKey: apiKey,
		rateLimiter: time.NewTicker(500 * time.Millisecond), // Rate limit to 2 requests per second
	}
}

// NewHTGoTTSProvider creates a new HTGoTTSProvider instance
func NewHTGoTTSProvider(folder string) *HTGoTTSProvider {
	return &HTGoTTSProvider{
		folder: folder,
	}
}

// NewTTSProvider creates a new TTS provider based on the provider name
func NewTTSProvider(provider string, apiKey string) TTSProvider {
	// Create AWS session with custom credentials provider
	creds := credentials.NewStaticCredentials(
		os.Getenv("AWS_ACCESS_KEY"),
		os.Getenv("AWS_SECRET_ACCESS_KEY"),
		"", // Token can be empty for regular API keys
	)
	// Use eu-central-1 as default region, can still be overridden by AWS_REGION env var
	sess := session.Must(session.NewSession(&aws.Config{
		Region:      aws.String("eu-central-1"),
		Credentials: creds,
	}))

	switch provider {
	case "elevenlabs":
		return &ElevenLabsProvider{apiKey: apiKey}
	default:
		// Use Polly as the fallback provider
		return &PollyProvider{client: polly.New(sess)}
	}
}

// GenerateAudio generates audio using ElevenLabs API
func (p *ElevenLabsProvider) GenerateAudio(text string, options map[string]string) ([]byte, error) {
	model := options["model"]
	if model == "" {
		model = "eleven_flash_v2_5"
	}
	voice := options["voice"]
	if voice == "" {
		voice = "cgSgspJ2msm6clMCkdW9"
	}

	reqBody := map[string]interface{}{
		"text":     text,
		"model_id": model,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	url := fmt.Sprintf("https://api.elevenlabs.io/v1/text-to-speech/%s/stream?output_format=mp3_44100_128", voice)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("xi-api-key", p.apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("elevenlabs api error: %s", string(body))
	}

	return io.ReadAll(resp.Body)
}

func (p *PollyProvider) GenerateAudio(text string, options map[string]string) ([]byte, error) {
	// Configure Polly input with Joanna voice
	input := &polly.SynthesizeSpeechInput{
		OutputFormat: aws.String("mp3"),
		Text:         aws.String(text),
		VoiceId:     aws.String("Joanna"), // Female, US English
		Engine:      aws.String("neural"),  // Use neural engine for better quality
	}

	// Override voice if specified in options
	if voice, ok := options["voice"]; ok && voice != "" {
		input.VoiceId = aws.String(voice)
	}

	// Generate speech
	output, err := p.client.SynthesizeSpeech(input)
	if err != nil {
		return nil, fmt.Errorf("failed to synthesize speech: %v", err)
	}
	defer output.AudioStream.Close()

	// Read the entire audio stream
	audioData, err := io.ReadAll(output.AudioStream)
	if err != nil {
		return nil, fmt.Errorf("failed to read audio stream: %v", err)
	}

	// Save to file if filename is provided
	if filename, ok := options["filename"]; ok && filename != "" {
		chunkStr := options["chunk"]
		audioFileName := fmt.Sprintf("%s_chunk_%s.mp3", filename, chunkStr)
		audioPath := path.Join("uploads", "audio", audioFileName)

		err = os.WriteFile(audioPath, audioData, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to save audio file: %v", err)
		}
	}

	return audioData, nil
}

func (p *CartesiaTTSProvider) GenerateAudio(text string, options map[string]string) ([]byte, error) {
	p.mutex.Lock()
	if p.processing {
		p.mutex.Unlock()
		return nil, fmt.Errorf("another request is being processed")
	}
	p.processing = true
	p.mutex.Unlock()

	// Wait for rate limiter
	<-p.rateLimiter.C

	defer func() {
		p.mutex.Lock()
		p.processing = false
		p.mutex.Unlock()
	}()

	// Ensure the audio folder exists
	if err := os.MkdirAll(p.folder, 0755); err != nil {
		return nil, fmt.Errorf("failed to create audio folder: %v", err)
	}

	// Get the filename and chunk number from options
	filename := options["filename"]
	if filename == "" {
		return nil, fmt.Errorf("filename is required in options")
	}
	chunkStr := options["chunk"]
	if chunkStr == "" {
		return nil, fmt.Errorf("chunk number is required in options")
	}

	// Create the audio filename with PDF name and chunk number
	audioFilename := fmt.Sprintf("%s_chunk_%s.mp3", filename, chunkStr)
	audioPath := path.Join(p.folder, audioFilename)

	// Check if file already exists
	if _, err := os.Stat(audioPath); err == nil {
		// File exists, read and return it
		return os.ReadFile(audioPath)
	}

	// Prepare the request payload
	payload := map[string]interface{}{
		"model_id": options["model"],
		"transcript": text,
		"voice": map[string]string{
			"mode": "id",
			"id":   options["voice"],
		},
		"output_format": map[string]interface{}{
			"container":   "mp3",
			"bit_rate":   128000,
			"sample_rate": 44100,
		},
		"language": "en",
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	// Create the request
	req, err := http.NewRequest("POST", "https://api.cartesia.ai/tts/bytes", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Add headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.apiKey))
	req.Header.Set("Cartesia-Version", "2025-04-16")

	// Make the request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (HTTP %d): %s", resp.StatusCode, string(body))
	}

	// Read the response
	audioData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	// Save the audio file
	if err := os.WriteFile(audioPath, audioData, 0644); err != nil {
		return nil, fmt.Errorf("failed to save audio file: %v", err)
	}

	return audioData, nil
}

func (p *HTGoTTSProvider) GenerateAudio(text string, options map[string]string) ([]byte, error) {
	// Get the filename and chunk number from options
	filename := options["filename"]
	if filename == "" {
		return nil, fmt.Errorf("filename is required in options")
	}
	chunkStr := options["chunk"]
	if chunkStr == "" {
		return nil, fmt.Errorf("chunk is required in options")
	}

	// Create a unique filename for this chunk
	outputFile := fmt.Sprintf("%s_chunk_%s.mp3", filename, chunkStr)
	outputPath := path.Join(p.folder, outputFile)

	// Initialize htgo-tts
	speech := htgotts.Speech{
		Folder:   p.folder,
		Language: voices.English,
	}

	// Generate the audio file
	if err := speech.Speak(text); err != nil {
		return nil, fmt.Errorf("failed to generate speech: %v", err)
	}

	// Read the generated file
	audioData, err := os.ReadFile(outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read generated audio file: %v", err)
	}

	return audioData, nil
} 