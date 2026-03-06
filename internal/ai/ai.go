package ai

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"google.golang.org/genai"
)

//go:embed prompt.md
var SystemPrompt string

const DefaultModel = "gemini-2.5-flash"

// Generator wraps the AI client and handles generating content.
type Generator struct {
	client *genai.Client
}

// NewGenerator initializes the environment and returns a new AI Generator.
func NewGenerator(ctx context.Context) (*Generator, error) {
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
		client: client,
	}, nil
}

// constructPrompt builds the final prompt to be sent to Gemini.
func constructPrompt(level, lessonTime, lessonType, sourceText, lessonTitle string) string {
	return fmt.Sprintf(
		"Target Level: %s\nLesson Time: %s minutes\nLesson Type: %s\nLesson Title: %s\n\nSource Material:\n%s",
		level, lessonTime, lessonType, lessonTitle, sourceText,
	)
}

// splitResponse splits a Gemini response into the student worksheet and teacher's key
// using the prompt-defined delimiters. It tries multiple delimiter variants to handle
// inconsistent LLM formatting.
func splitResponse(fullResponse string) (worksheet string, teacherKey string, err error) {
	// Strip wrapping markdown code fences if present
	cleaned := fullResponse
	cleaned = strings.TrimSpace(cleaned)
	if strings.HasPrefix(cleaned, "```") {
		// Remove opening fence (e.g., ```markdown\n)
		firstNewline := strings.Index(cleaned, "\n")
		if firstNewline != -1 {
			cleaned = cleaned[firstNewline+1:]
		}
		// Remove closing fence
		if strings.HasSuffix(strings.TrimSpace(cleaned), "```") {
			cleaned = strings.TrimSpace(cleaned)
			cleaned = cleaned[:len(cleaned)-3]
			cleaned = strings.TrimSpace(cleaned)
		}
	}
	fullResponse = cleaned

	// Try multiple delimiter variants — Gemini sometimes wraps them in markdown formatting
	endStudentDelimiters := []string{
		"[END STUDENT WORKSHEET]",
		"**[END STUDENT WORKSHEET]**",
		"`[END STUDENT WORKSHEET]`",
	}

	endStudentIdx := -1
	endStudentLen := 0
	for _, d := range endStudentDelimiters {
		idx := strings.Index(fullResponse, d)
		if idx != -1 {
			endStudentIdx = idx
			endStudentLen = len(d)
			break
		}
	}

	// If no end-student delimiter found, try splitting on the teacher key start instead
	if endStudentIdx == -1 {
		beginTeacherIdx := findDelimiter(fullResponse, []string{
			"[BEGIN TEACHER KEY]",
			"**[BEGIN TEACHER KEY]**",
			"`[BEGIN TEACHER KEY]`",
		})
		if beginTeacherIdx != -1 {
			worksheet = strings.TrimSpace(fullResponse[:beginTeacherIdx])
			// Also strip a leading [BEGIN STUDENT WORKSHEET] if present
			for _, prefix := range []string{"[BEGIN STUDENT WORKSHEET]", "**[BEGIN STUDENT WORKSHEET]**", "`[BEGIN STUDENT WORKSHEET]`"} {
				worksheet = strings.TrimPrefix(worksheet, prefix)
				worksheet = strings.TrimSpace(worksheet)
			}
			teacherKey = extractTeacherKey(fullResponse[beginTeacherIdx:])
			return worksheet, teacherKey, nil
		}
		// No delimiters at all — return the whole response as the worksheet
		worksheet = strings.TrimSpace(fullResponse)
		for _, prefix := range []string{"[BEGIN STUDENT WORKSHEET]", "**[BEGIN STUDENT WORKSHEET]**", "`[BEGIN STUDENT WORKSHEET]`"} {
			worksheet = strings.TrimPrefix(worksheet, prefix)
			worksheet = strings.TrimSpace(worksheet)
		}
		return worksheet, "", nil
	}

	worksheet = strings.TrimSpace(fullResponse[:endStudentIdx])
	// Also strip a leading [BEGIN STUDENT WORKSHEET] if present
	for _, prefix := range []string{"[BEGIN STUDENT WORKSHEET]", "**[BEGIN STUDENT WORKSHEET]**", "`[BEGIN STUDENT WORKSHEET]`"} {
		worksheet = strings.TrimPrefix(worksheet, prefix)
		worksheet = strings.TrimSpace(worksheet)
	}

	remaining := fullResponse[endStudentIdx+endStudentLen:]
	teacherKey = extractTeacherKey(remaining)

	return worksheet, teacherKey, nil
}

// findDelimiter searches for any of the given delimiter variants and returns the index of the first match.
func findDelimiter(s string, delimiters []string) int {
	for _, d := range delimiters {
		idx := strings.Index(s, d)
		if idx != -1 {
			return idx
		}
	}
	return -1
}

// extractTeacherKey extracts the teacher key content from text that starts at or after
// a [BEGIN TEACHER KEY] delimiter.
func extractTeacherKey(s string) string {
	beginDelimiters := []string{"[BEGIN TEACHER KEY]", "**[BEGIN TEACHER KEY]**", "`[BEGIN TEACHER KEY]`"}
	endDelimiters := []string{"[END TEACHER KEY]", "**[END TEACHER KEY]**", "`[END TEACHER KEY]`"}

	beginIdx := -1
	beginLen := 0
	for _, d := range beginDelimiters {
		idx := strings.Index(s, d)
		if idx != -1 {
			beginIdx = idx
			beginLen = len(d)
			break
		}
	}

	if beginIdx == -1 {
		return ""
	}

	content := s[beginIdx+beginLen:]

	for _, d := range endDelimiters {
		idx := strings.Index(content, d)
		if idx != -1 {
			return strings.TrimSpace(content[:idx])
		}
	}

	return strings.TrimSpace(content)
}

// GenerateWorksheet calls the Gemini API, and returns the markdown response split into worksheet and teacher key.
func (g *Generator) GenerateWorksheet(ctx context.Context, level, lessonTime, lessonType, sourceText, lessonTitle string) (worksheet string, teacherKey string, err error) {
	prompt := constructPrompt(level, lessonTime, lessonType, sourceText, lessonTitle)

	result, err := g.client.Models.GenerateContent(ctx, DefaultModel, genai.Text(prompt), &genai.GenerateContentConfig{
		SystemInstruction: genai.NewContentFromText(SystemPrompt, genai.RoleUser),
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to generate content from Gemini API: %w", err)
	}

	if result == nil || len(result.Candidates) == 0 {
		return "", "", fmt.Errorf("received empty response from Gemini API")
	}

	text := result.Text()
	if text == "" {
		return "", "", fmt.Errorf("received empty text from Gemini API")
	}

	ws, tk, _ := splitResponse(text)
	// splitResponse never returns a fatal error — it always returns usable content
	return ws, tk, nil
}

// SaveDocuments saves the worksheet and teacher key to the worksheets directory and returns the file paths.
func SaveDocuments(worksheetContent, teacherKeyContent, level, lessonType, lessonTitle, baseDir string) (worksheetPath, teacherKeyPath string, err error) {
	dir := filepath.Join(baseDir, "worksheets")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", "", fmt.Errorf("failed to create worksheets directory: %w", err)
	}

	dateStr := time.Now().Format("2006-01-02_1504")

	safeTitle := strings.ReplaceAll(lessonTitle, " ", "_")
	safeTitle = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			return r
		}
		return -1
	}, safeTitle)

	if safeTitle == "" {
		safeTitle = "lesson"
	}

	wsFilename := filepath.Join(dir, fmt.Sprintf("%s_%s_%s_%s_worksheet.md", level, lessonType, safeTitle, dateStr))
	tkFilename := filepath.Join(dir, fmt.Sprintf("%s_%s_%s_%s_teacher_key.md", level, lessonType, safeTitle, dateStr))

	if err := os.WriteFile(wsFilename, []byte(worksheetContent), 0644); err != nil {
		return "", "", fmt.Errorf("failed to write worksheet file: %w", err)
	}

	if teacherKeyContent != "" {
		if err := os.WriteFile(tkFilename, []byte(teacherKeyContent), 0644); err != nil {
			return "", "", fmt.Errorf("failed to write teacher key file: %w", err)
		}
	}

	return wsFilename, tkFilename, nil
}
