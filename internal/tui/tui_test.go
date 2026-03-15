package tui

import (
	"testing"
	"teaching-assistant-app/internal/drive"
	tea "github.com/charmbracelet/bubbletea"
)

func TestSelectorCycling(t *testing.T) {
	m := NewModel(nil, nil, "")
	m.activeFormField = 1 // Level Index
	
	initialIndex := m.levelIndex
	// Press left
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyLeft})
	m = newModel.(Model)
	expectedIndex := initialIndex - 1
	if expectedIndex < 0 {
		expectedIndex = len(levelOptions) - 1
	}
	if m.levelIndex != expectedIndex {
		t.Errorf("Expected levelIndex to be %d, got %d", expectedIndex, m.levelIndex)
	}

	// Press right
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight})
	m = newModel.(Model)
	if m.levelIndex != initialIndex {
		t.Errorf("Expected levelIndex to return to %d, got %d", initialIndex, m.levelIndex)
	}
}

func TestPreviewStateTransitions(t *testing.T) {
	m := NewModel(nil, nil, "")
	m.showPreview = true

	// Accept preview
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
	m = newModel.(Model)
	if m.showPreview {
		t.Errorf("Expected showPreview to be false after pressing y")
	}

	m.showPreview = true
	// Reject preview
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	m = newModel.(Model)
	if m.showPreview {
		t.Errorf("Expected showPreview to be false after pressing n")
	}
}

func TestFolderPickerNavigation(t *testing.T) {
	m := NewModel(nil, nil, "")
	m.showFolderPicker = true
	m.folders = []drive.Folder{{ID: "1", Name: "A"}, {ID: "2", Name: "B"}}
	m.folderCursor = 0

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = newModel.(Model)
	if m.folderCursor != 1 {
		t.Errorf("Expected folderCursor to be 1, got %d", m.folderCursor)
	}

	// Bound check
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = newModel.(Model)
	if m.folderCursor != 1 {
		t.Errorf("Expected folderCursor to bound at 1, got %d", m.folderCursor)
	}

	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = newModel.(Model)
	if m.folderCursor != 0 {
		t.Errorf("Expected folderCursor to be 0, got %d", m.folderCursor)
	}
}

func TestFolderPickerCreateFolder(t *testing.T) {
	m := NewModel(nil, nil, "")
	m.showFolderPicker = true

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	m = newModel.(Model)

	if !m.showCreateFolder {
		t.Errorf("Expected showCreateFolder to be true after pressing 'n'")
	}
}
