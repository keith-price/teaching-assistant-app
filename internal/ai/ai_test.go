package ai

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

type MockClipboardReader struct {
	Text string
	Err  error
}

func (m *MockClipboardReader) ReadAll() (string, error) {
	if m.Err != nil {
		return "", m.Err
	}
	if m.Text == "" {
		return "", fmt.Errorf("clipboard is empty")
	}
	return m.Text, nil
}

func TestConstructPrompt(t *testing.T) {
	name := "Alice"
	level := "B1"
	lessonTime := "45"
	lessonType := "Conversation"
	transcript := "Hello, how are you?"

	prompt := constructPrompt(name, level, lessonTime, lessonType, transcript)
	if !strings.Contains(prompt, "Alice") {
		t.Errorf("Expected prompt to contain name, got: %s", prompt)
	}
	if !strings.Contains(prompt, "B1") {
		t.Errorf("Expected prompt to contain level, got: %s", prompt)
	}
	if !strings.Contains(prompt, "45 minutes") {
		t.Errorf("Expected prompt to contain lesson time, got: %s", prompt)
	}
	if !strings.Contains(prompt, "Conversation") {
		t.Errorf("Expected prompt to contain lesson type, got: %s", prompt)
	}
	if !strings.Contains(prompt, transcript) {
		t.Errorf("Expected prompt to contain transcript, got: %s", prompt)
	}
}

func TestClipboardReaderErrors(t *testing.T) {
	// Test empty clipboard
	mock := &MockClipboardReader{Text: ""}
	generator := &Generator{clipboard: mock}

	_, err := generator.GenerateWorksheet(context.Background(), "Alice", "B1", "60", "General")
	if err == nil {
		t.Error("Expected error for empty clipboard, got nil")
	} else if !strings.Contains(err.Error(), "clipboard is empty") {
		t.Errorf("Expected 'clipboard is empty' error, got: %v", err)
	}

	// Test read error
	mock = &MockClipboardReader{Err: fmt.Errorf("read failed")}
	generator.clipboard = mock

	_, err = generator.GenerateWorksheet(context.Background(), "Alice", "B1", "60", "General")
	if err == nil {
		t.Error("Expected error for clipboard read failure, got nil")
	} else if !strings.Contains(err.Error(), "read failed") {
		t.Errorf("Expected 'read failed' error, got: %v", err)
	}
}

func TestSaveWorksheet(t *testing.T) {
	mock := &MockClipboardReader{Text: "Sample transcript"}
	generator := &Generator{clipboard: mock}

	content := "# Worksheet"
	name := "Bob Smith"
	tmpDir := t.TempDir()

	path, err := generator.SaveWorksheet(content, name, tmpDir)
	if err != nil {
		t.Fatalf("Failed to save worksheet: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("Expected file to exist at %s", path)
	}

	// Verify filename format
	expectedDate := time.Now().Format("2006-01-02_1504")
	expectedName := "Bob_Smith"
	if !strings.Contains(path, expectedDate) || !strings.Contains(path, expectedName) {
		t.Errorf("Filename does not match expected format: %s", path)
	}
}
