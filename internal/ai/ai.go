package ai

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/joho/godotenv"
	"google.golang.org/genai"
)

//go:embed prompt.md
var SystemPrompt string

const DefaultModel = "gemini-2.5-flash"

// ClipboardReader interface for testability.
type ClipboardReader interface {
	ReadAll() (string, error)
}

// OSClipboardReader implements ClipboardReader using the OS clipboard.
type OSClipboardReader struct{}

// ReadAll reads raw text from the OS clipboard.
func (c *OSClipboardReader) ReadAll() (string, error) {
	text, err := clipboard.ReadAll()
	if err != nil {
		return "", fmt.Errorf("failed to read from clipboard: %w", err)
	}
	if text == "" {
		return "", fmt.Errorf("clipboard is empty")
	}
	return text, nil
}

// Generator wraps the AI client and handles generating content.
type Generator struct {
	client    *genai.Client
	clipboard ClipboardReader
}

// NewGenerator initializes the environment and returns a new AI Generator.
func NewGenerator(ctx context.Context, cb ClipboardReader) (*Generator, error) {
	// Attempt to load .env file, but don't fail if it doesn't exist,
	// as the environment variable might already be set.
	_ = godotenv.Load()

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY environment variable is not set")
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	return &Generator{
		client:    client,
		clipboard: cb,
	}, nil
}

// constructPrompt builds the final prompt to be sent to Gemini.
func constructPrompt(studentName, studentLevel, lessonTime, lessonType, transcript string) string {
	return fmt.Sprintf(
		"Student Name: %s\nTarget Level: %s\nLesson Time: %s minutes\nLesson Type: %s\n\nTranscript:\n%s",
		studentName, studentLevel, lessonTime, lessonType, transcript,
	)
}

// GenerateWorksheet reads the clipboard transcript, calls the Gemini API, and returns the markdown response.
func (g *Generator) GenerateWorksheet(ctx context.Context, studentName, studentLevel, lessonTime, lessonType string) (string, error) {
	transcript, err := g.clipboard.ReadAll()
	if err != nil {
		return "", fmt.Errorf("clipboard error: %w", err)
	}

	prompt := constructPrompt(studentName, studentLevel, lessonTime, lessonType, transcript)

	result, err := g.client.Models.GenerateContent(ctx, DefaultModel, genai.Text(prompt), &genai.GenerateContentConfig{
		SystemInstruction: genai.NewContentFromText(SystemPrompt, genai.RoleUser),
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate content from Gemini API: %w", err)
	}

	if result == nil || len(result.Candidates) == 0 {
		return "", fmt.Errorf("received empty response from Gemini API")
	}

	text := result.Text()
	if text == "" {
		return "", fmt.Errorf("received empty text from Gemini API")
	}

	return text, nil
}

// SaveWorksheet saves the markdown content to the worksheets directory and returns the file path.
func (g *Generator) SaveWorksheet(content, studentName, baseDir string) (string, error) {
	dir := filepath.Join(baseDir, "worksheets")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create worksheets directory: %w", err)
	}

	dateStr := time.Now().Format("2006-01-02_1504")
	safeName := strings.ReplaceAll(studentName, " ", "_")
	filename := filepath.Join(dir, fmt.Sprintf("%s_%s_lesson.md", dateStr, safeName))

	if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write worksheet file: %w", err)
	}

	return filename, nil
}

// GenerateAndSaveWorksheet coordinates reading from the clipboard, generating the worksheet, and saving it.
func (g *Generator) GenerateAndSaveWorksheet(ctx context.Context, studentName, studentLevel, lessonTime, lessonType, baseDir string) (string, error) {
	content, err := g.GenerateWorksheet(ctx, studentName, studentLevel, lessonTime, lessonType)
	if err != nil {
		return "", err
	}

	path, err := g.SaveWorksheet(content, studentName, baseDir)
	if err != nil {
		return "", err
	}

	return path, nil
}
