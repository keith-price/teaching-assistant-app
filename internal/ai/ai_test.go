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
