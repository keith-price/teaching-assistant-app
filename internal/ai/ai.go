package ai

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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

// Precompile regexes for robust matching
var (
	endStudentRegex   = regexp.MustCompile(`(?i)(?:#|\*|_|-|\s)*\[END STUDENT WORKSHEET\](?:#|\*|_|-|\s)*`)
	beginTeacherRegex = regexp.MustCompile(`(?i)(?:#|\*|_|-|\s)*\[BEGIN TEACHER KEY\](?:#|\*|_|-|\s)*`)
	endTeacherRegex   = regexp.MustCompile(`(?i)(?:#|\*|_|-|\s)*\[END TEACHER KEY\](?:#|\*|_|-|\s)*`)
	beginStudentRegex = regexp.MustCompile(`(?i)(?:#|\*|_|-|\s)*\[BEGIN STUDENT WORKSHEET\](?:#|\*|_|-|\s)*`)
)

// splitResponse splits a Gemini response into the student worksheet and teacher's key
// using the prompt-defined delimiters. It uses regex to handle inconsistent LLM formatting.
func splitResponse(fullResponse string) (worksheet string, teacherKey string, err error) {
	// Strip wrapping markdown code fences if present
	cleaned := strings.TrimSpace(fullResponse)
	if strings.HasPrefix(cleaned, "```") {
		firstNewline := strings.Index(cleaned, "\n")
		if firstNewline != -1 {
			cleaned = cleaned[firstNewline+1:]
		}
		if strings.HasSuffix(strings.TrimSpace(cleaned), "```") {
			cleaned = strings.TrimSpace(cleaned)
			cleaned = cleaned[:len(cleaned)-3]
			cleaned = strings.TrimSpace(cleaned)
		}
	}
	fullResponse = cleaned

	endStudentMatch := endStudentRegex.FindStringIndex(fullResponse)

	if endStudentMatch == nil {
		// If no [END STUDENT WORKSHEET] delimiter found, try splitting on [BEGIN TEACHER KEY] instead
		beginTeacherMatch := beginTeacherRegex.FindStringIndex(fullResponse)
		if beginTeacherMatch != nil {
			worksheet = strings.TrimSpace(fullResponse[:beginTeacherMatch[0]])
			worksheet = beginStudentRegex.ReplaceAllString(worksheet, "")

			teacherKey = extractTeacherKey(fullResponse[beginTeacherMatch[0]:])
			return strings.TrimSpace(worksheet), teacherKey, nil
		}

		// No delimiters at all — return the whole response as the worksheet
		worksheet = beginStudentRegex.ReplaceAllString(fullResponse, "")
		return strings.TrimSpace(worksheet), "", nil
	}

	worksheet = strings.TrimSpace(fullResponse[:endStudentMatch[0]])
	worksheet = beginStudentRegex.ReplaceAllString(worksheet, "")

	remaining := fullResponse[endStudentMatch[1]:]
	teacherKey = extractTeacherKey(remaining)

	return strings.TrimSpace(worksheet), teacherKey, nil
}

// extractTeacherKey extracts the teacher key content from text that starts at or after
// a [BEGIN TEACHER KEY] delimiter.
func extractTeacherKey(s string) string {
	beginMatch := beginTeacherRegex.FindStringIndex(s)

	if beginMatch == nil {
		return ""
	}

	content := s[beginMatch[1]:]

	endMatch := endTeacherRegex.FindStringIndex(content)
	if endMatch != nil {
		return strings.TrimSpace(content[:endMatch[0]])
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

	// DEBUG: dump raw response to disk so we can see why it's not splitting
	os.WriteFile("debug_split.txt", []byte(text), 0644)

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
