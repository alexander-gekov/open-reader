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

// TogetherProvider implements TTSProvider for Together AI
type TogetherProvider struct {
	apiKey string
}

// ReplicateProvider implements TTSProvider for Replicate
type ReplicateProvider struct {
	apiKey string
}

type ReplicateResponse struct {
	ID      string `json:"id"`
	Output  string `json:"output"`
	Status  string `json:"status"`
	Error   string `json:"error"`
	Version string `json:"version"`
}

// CartesiaTTSProvider implements TTSProvider using Cartesia's API
type CartesiaTTSProvider struct {
	folder string
	apiKey string
	rateLimiter *time.Ticker
	processing bool
	mutex sync.Mutex
}

// NewCartesiaTTSProvider creates a new CartesiaTTSProvider instance
func NewCartesiaTTSProvider(folder string, apiKey string) *CartesiaTTSProvider {
	return &CartesiaTTSProvider{
		folder: folder,
		apiKey: apiKey,
		rateLimiter: time.NewTicker(500 * time.Millisecond), // Rate limit to 2 requests per second
	}
}

// NewTTSProvider creates a new TTS provider based on the provider name
func NewTTSProvider(provider string, apiKey string) TTSProvider {
	switch provider {
	case "elevenlabs":
		return &ElevenLabsProvider{apiKey: apiKey}
	case "together":
		return &TogetherProvider{apiKey: apiKey}
	case "replicate":
		return &ReplicateProvider{apiKey: apiKey}
	case "fallback":
		return &HTGoTTSProvider{folder: "uploads/audio"}
	case "cartesia":
		return NewCartesiaTTSProvider("uploads/audio", apiKey)
	default:
		// If no API key is provided, use the fallback provider
		if apiKey == "" {
			return &HTGoTTSProvider{folder: "uploads/audio"}
		}
		return &ElevenLabsProvider{apiKey: apiKey}
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

// GenerateAudio generates audio using Together AI API
func (p *TogetherProvider) GenerateAudio(text string, options map[string]string) ([]byte, error) {
	model := options["model"]
	if model == "" {
		model = "Cartesia/Sonic"
	}
	voice := options["voice"]
	if voice == "" {
		voice = "default"
	}

	reqBody := map[string]interface{}{
		"text":  text,
		"voice": voice,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	url := fmt.Sprintf("https://api.together.xyz/inference/%s", model)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("together api error: %s", string(body))
	}

	return io.ReadAll(resp.Body)
}

// GenerateAudio generates audio using Replicate API
func (p *ReplicateProvider) GenerateAudio(text string, options map[string]string) ([]byte, error) {
	model := options["model"]
	if model == "" {
		model = "jaaari/kokoro-82m"
	}
	voice := options["voice"]
	if voice == "" {
		voice = "af_bella"
	}

	reqBody := map[string]interface{}{
		"version": "f559560eb822dc509045f3921a1921234918b91739db4bf3daab2169b71c7a13",
		"input": map[string]interface{}{
			"text":  text,
			"voice": voice,
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	url := "https://api.replicate.com/v1/predictions"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Token "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("replicate api error: %s", string(body))
	}

	var result ReplicateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	// Poll until the prediction is complete
	for result.Status == "processing" {
		time.Sleep(1 * time.Second)

		req, err = http.NewRequest("GET", fmt.Sprintf("https://api.replicate.com/v1/predictions/%s", result.ID), nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create status request: %v", err)
		}

		req.Header.Set("Authorization", "Token "+p.apiKey)
		resp, err = client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to check status: %v", err)
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("failed to decode status response: %v", err)
		}
		resp.Body.Close()
	}

	if result.Error != "" {
		return nil, fmt.Errorf("replicate error: %s", result.Error)
	}

	if result.Output == "" {
		return nil, fmt.Errorf("no output URL in response")
	}

	// Download the audio file from the output URL
	req, err = http.NewRequest("GET", result.Output, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create download request: %v", err)
	}

	resp, err = client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download audio: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download audio, status: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// HTGoTTSProvider implements TTSProvider using Google TTS
type HTGoTTSProvider struct {
	folder string
}

// NewHTGoTTSProvider creates a new HTGoTTSProvider instance
func NewHTGoTTSProvider(folder string) *HTGoTTSProvider {
	return &HTGoTTSProvider{
		folder: folder,
	}
}

func (p *HTGoTTSProvider) GenerateAudio(text string, options map[string]string) ([]byte, error) {
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

	// Get the TTS URL from Google
	url := GetGoogleTTSURL(text, "en")

	// Download the audio file
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download audio: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download audio, status: %d", resp.StatusCode)
	}

	// Save the audio file
	audioData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read audio data: %v", err)
	}

	if err := os.WriteFile(audioPath, audioData, 0644); err != nil {
		return nil, fmt.Errorf("failed to save audio file: %v", err)
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