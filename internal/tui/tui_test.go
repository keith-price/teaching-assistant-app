package tui

import (
	"testing"
	"teaching-assistant-app/internal/db"
	"teaching-assistant-app/internal/drive"
	tea "github.com/charmbracelet/bubbletea"
)

func TestModelInitAndNavigation(t *testing.T) {
	m := NewModel(nil, nil, nil, "")

	m.todayLessons = []db.LessonWithStudent{
		{Lesson: db.Lesson{ID: 1}, Student: db.Student{Name: "Alice"}},
		{Lesson: db.Lesson{ID: 2}, Student: db.Student{Name: "Bob"}},
	}
	m.tomorrowLessons = []db.LessonWithStudent{
		{Lesson: db.Lesson{ID: 3}, Student: db.Student{Name: "Charlie"}},
	}

	if m.activePane != 0 || m.cursor != 0 {
		t.Errorf("Expected active pane 0 and cursor 0, got %d and %d", m.activePane, m.cursor)
	}

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown, Alt: false})
	m = newModel.(Model)
	if m.cursor != 1 {
		t.Errorf("Expected cursor 1, got %d", m.cursor)
	}

	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight, Alt: false})
	m = newModel.(Model)
	if m.activePane != 1 {
		t.Errorf("Expected active pane 1, got %d", m.activePane)
	}
	if m.cursor != 0 {
		t.Errorf("Expected cursor to bound to 0, got %d", m.cursor)
	}
}

func TestToggleVocabUpdate(t *testing.T) {
	m := NewModel(nil, nil, nil, "")
	m.todayLessons = []db.LessonWithStudent{
		{Lesson: db.Lesson{ID: 1, VocabSent: false}, Student: db.Student{Name: "Alice"}},
	}

	newModel, _ := m.Update(vocabToggledMsg{lessonID: 1})
	m = newModel.(Model)

	if !m.todayLessons[0].Lesson.VocabSent {
		t.Errorf("Expected vocab to be toggled to true")
	}
}

func TestGKeyOpensForm(t *testing.T) {
	m := NewModel(nil, nil, nil, "")
	// Empty lists
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")})
	m = newModel.(Model)

	if !m.showForm {
		t.Errorf("Expected showForm to be true")
	}
}

func TestSelectorCycling(t *testing.T) {
	m := NewModel(nil, nil, nil, "")
	m.showForm = true
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
	m := NewModel(nil, nil, nil, "")
	m.showPreview = true

	// Accept preview
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
	m = newModel.(Model)
	if m.showPreview {
		t.Errorf("Expected showPreview to be false after pressing y")
	}
	// "y" actually triggers listDriveFoldersCmd and eventually shows folder picker, 
	// but strictly in the model `showPreview` becomes false

	m.showPreview = true
	// Reject preview
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	m = newModel.(Model)
	if m.showPreview {
		t.Errorf("Expected showPreview to be false after pressing n")
	}
}

func TestFolderPickerNavigation(t *testing.T) {
	m := NewModel(nil, nil, nil, "")
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
	m := NewModel(nil, nil, nil, "")
	m.showFolderPicker = true

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	m = newModel.(Model)

	if !m.showCreateFolder {
		t.Errorf("Expected showCreateFolder to be true after pressing 'n'")
	}
}
