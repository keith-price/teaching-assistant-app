package ai

import (
	"os"
	"strings"
	"testing"
)

func TestConstructPrompt(t *testing.T) {
	level := "B1"
	lessonTime := "45"
	lessonType := "Conversation"
	lessonTitle := "Climate Change"
	transcript := "Hello, how are you?"

	prompt := constructPrompt(level, lessonTime, lessonType, transcript, lessonTitle)
	if !strings.Contains(prompt, "B1") {
		t.Errorf("Expected prompt to contain level, got: %s", prompt)
	}
	if !strings.Contains(prompt, "45 minutes") {
		t.Errorf("Expected prompt to contain lesson time, got: %s", prompt)
	}
	if !strings.Contains(prompt, "Conversation") {
		t.Errorf("Expected prompt to contain lesson type, got: %s", prompt)
	}
	if !strings.Contains(prompt, "Climate Change") {
		t.Errorf("Expected prompt to contain lesson title, got: %s", prompt)
	}
	if !strings.Contains(prompt, transcript) {
		t.Errorf("Expected prompt to contain transcript, got: %s", prompt)
	}
}

func TestSplitResponse(t *testing.T) {
	// Test valid
	fullResp := `This is the worksheet
[END STUDENT WORKSHEET]
[BEGIN TEACHER KEY]
This is the teacher key
[END TEACHER KEY]`

	ws, tk, err := splitResponse(fullResp)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if ws != "This is the worksheet" {
		t.Errorf("Unexpected worksheet content: %s", ws)
	}
	if tk != "This is the teacher key" {
		t.Errorf("Unexpected teacher key content: %s", tk)
	}

	// Test missing delimiters — should gracefully fallback, not error
	fullRespMissing := `Just some text without delimiters`
	wsMissing, tkMissing, errMissing := splitResponse(fullRespMissing)
	if errMissing != nil {
		t.Errorf("Expected no error for missing delimiters (graceful fallback), got: %v", errMissing)
	}
	if wsMissing != fullRespMissing {
		t.Errorf("Expected fallback worksheet to be full response, got: %s", wsMissing)
	}
	if tkMissing != "" {
		t.Errorf("Expected fallback teacher key to be empty, got: %s", tkMissing)
	}

	// Test with markdown code fence wrapping
	fencedResp := "```markdown\n[BEGIN STUDENT WORKSHEET]\nWorksheet content\n[END STUDENT WORKSHEET]\n[BEGIN TEACHER KEY]\nKey content\n[END TEACHER KEY]\n```"
	wsFenced, tkFenced, errFenced := splitResponse(fencedResp)
	if errFenced != nil {
		t.Errorf("Unexpected error for fenced response: %v", errFenced)
	}
	if !strings.Contains(wsFenced, "Worksheet content") {
		t.Errorf("Expected worksheet content in fenced response, got: %s", wsFenced)
	}
	if strings.Contains(wsFenced, "```") {
		t.Errorf("Expected fences to be stripped from worksheet in fenced response, got: %s", wsFenced)
	}
	if !strings.Contains(tkFenced, "Key content") {
		t.Errorf("Expected teacher key in fenced response, got: %s", tkFenced)
	}
	if strings.Contains(tkFenced, "```") {
		t.Errorf("Expected fences to be stripped from teacher key in fenced response, got: %s", tkFenced)
	}

	// Test with bold-wrapped delimiters
	boldResp := "**[BEGIN STUDENT WORKSHEET]**\nBold worksheet\n**[END STUDENT WORKSHEET]**\n**[BEGIN TEACHER KEY]**\nBold key\n**[END TEACHER KEY]**"
	wsBold, tkBold, errBold := splitResponse(boldResp)
	if errBold != nil {
		t.Errorf("Unexpected error for bold response: %v", errBold)
	}
	if !strings.Contains(wsBold, "Bold worksheet") {
		t.Errorf("Expected worksheet content in bold response, got: %s", wsBold)
	}
	if !strings.Contains(tkBold, "Bold key") {
		t.Errorf("Expected teacher key in bold response, got: %s", tkBold)
	}

	// Test fallback: [BEGIN TEACHER KEY] present but [END STUDENT WORKSHEET] missing
	noEndResp := "Worksheet content\n[BEGIN TEACHER KEY]\nTeacher content\n[END TEACHER KEY]"
	wsNoEnd, tkNoEnd, errNoEnd := splitResponse(noEndResp)
	if errNoEnd != nil {
		t.Errorf("Unexpected error: %v", errNoEnd)
	}
	if !strings.Contains(wsNoEnd, "Worksheet content") {
		t.Errorf("Expected worksheet content, got: %s", wsNoEnd)
	}
	if !strings.Contains(tkNoEnd, "Teacher content") {
		t.Errorf("Expected teacher key content, got: %s", tkNoEnd)
	}
}

func TestSaveDocuments(t *testing.T) {
	wsContent := "# Worksheet"
	tkContent := "# Teacher Key"
	level := "B2"
	lessonType := "Reading"
	title := "Climate Change"
	tmpDir := t.TempDir()

	wsPath, tkPath, err := SaveDocuments(wsContent, tkContent, level, lessonType, title, tmpDir)
	if err != nil {
		t.Fatalf("Failed to save documents: %v", err)
	}

	// Verify files exist
	if _, err := os.Stat(wsPath); os.IsNotExist(err) {
		t.Errorf("Expected worksheet file to exist at %s", wsPath)
	}
	if _, err := os.Stat(tkPath); os.IsNotExist(err) {
		t.Errorf("Expected teacher key file to exist at %s", tkPath)
	}

	// Verify filename format
	if !strings.Contains(wsPath, "B2_Reading_Climate_Change_") || !strings.HasSuffix(wsPath, "_worksheet.md") {
		t.Errorf("Worksheet filename does not match expected format: %s", wsPath)
	}
	if !strings.Contains(tkPath, "B2_Reading_Climate_Change_") || !strings.HasSuffix(tkPath, "_teacher_key.md") {
		t.Errorf("Teacher key filename does not match expected format: %s", tkPath)
	}
}
