package tui

import (
	"fmt"
	"strings"

	"teaching-assistant-app/internal/db"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case tea.KeyMsg:
		if m.showFolderPicker {
			return m.handleFolderPickerKeys(msg)
		}
		if m.showPreview {
			return m.handlePreviewKeys(msg)
		}
		if m.showForm {
			return m.handleFormKeys(msg)
		}
		return m.handleMainKeys(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case lessonsFetchedMsg:
		m.todayLessons = msg.today
		m.tomorrowLessons = msg.tomorrow
		m.boundCursor()

	case vocabToggledMsg:
		for i, lv := range m.todayLessons {
			if lv.Lesson.ID == msg.lessonID {
				m.todayLessons[i].Lesson.VocabSent = !lv.Lesson.VocabSent
				m.statusMsg = fmt.Sprintf("✅ Vocab toggled for %s", lv.Student.Name)
			}
		}
		for i, lv := range m.tomorrowLessons {
			if lv.Lesson.ID == msg.lessonID {
				m.tomorrowLessons[i].Lesson.VocabSent = !lv.Lesson.VocabSent
				m.statusMsg = fmt.Sprintf("✅ Vocab toggled for %s", lv.Student.Name)
			}
		}

	case worksheetPreviewMsg:
		m.generating = false
		m.worksheetContent = msg.worksheet
		m.teacherKeyContent = msg.teacherKey
		m.showPreview = true

		m.statusMsg = "Preview ready. Press 'y' to accept, 'n' to discard."
		// Save locally as fallback too
		cmds = append(cmds, saveLocallyCmd(msg.worksheet, msg.teacherKey, msg.level, msg.lessonType, msg.title, m.baseDir))

	case foldersListedMsg:
		m.folders = msg.folders
		m.folderCursor = 0
		m.showFolderPicker = true

	case uploadCompleteMsg:
		m.uploading = false
		m.showFolderPicker = false
		m.statusMsg = fmt.Sprintf("✅ Uploaded to Drive: %s", msg.folderName)

	case folderCreatedMsg:
		m.statusMsg = fmt.Sprintf("✅ Created folder: %s", msg.folder.Name)
		// Refresh the folder list for the current parent
		cmds = append(cmds, listDriveFoldersCmd(m.driveClient, msg.parentID))

	case errMsg:
		m.generating = false
		m.uploading = false
		m.err = msg.err
		m.statusMsg = fmt.Sprintf("❌ Error: %v", msg.err)
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) getActiveLesson() *db.LessonWithStudent {
	var list []db.LessonWithStudent
	if m.activePane == 0 {
		list = m.todayLessons
	} else {
		list = m.tomorrowLessons
	}
	if m.cursor >= 0 && m.cursor < len(list) {
		return &list[m.cursor]
	}
	return nil
}

func (m Model) handleMainKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}

	case "down", "j":
		var max int
		if m.activePane == 0 {
			max = len(m.todayLessons) - 1
		} else {
			max = len(m.tomorrowLessons) - 1
		}
		if m.cursor < max {
			m.cursor++
		}

	case "left", "h":
		if m.activePane == 1 {
			m.activePane = 0
			m.boundCursor()
		}

	case "right", "l":
		if m.activePane == 0 {
			m.activePane = 1
			m.boundCursor()
		}

	case "tab":
		m.activePane = 1 - m.activePane
		m.boundCursor()

	case "v":
		lv := m.getActiveLesson()
		if lv != nil {
			return m, toggleVocabCmd(m.store, lv.Lesson.ID)
		}

	case "g":
		m.showForm = true
		m.initForm()
	}
	return m, nil
}

func (m Model) handleFormKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg.String() {
	case "esc", "ctrl+c":
		m.showForm = false
		return m, nil

	case "tab", "shift+tab":
		if msg.String() == "shift+tab" {
			m.activeFormField--
			if m.activeFormField < 0 {
				m.activeFormField = 4
			}
		} else {
			m.activeFormField++
			if m.activeFormField > 4 {
				m.activeFormField = 0
			}
		}

		if m.activeFormField == 0 {
			cmd = m.lessonTitleInput.Focus()
			m.sourceTextArea.Blur()
		} else if m.activeFormField == 4 {
			cmd = m.sourceTextArea.Focus()
			m.lessonTitleInput.Blur()
		} else {
			m.lessonTitleInput.Blur()
			m.sourceTextArea.Blur()
		}
		return m, cmd

	case "left", "right":
		if m.activeFormField == 1 {
			if msg.String() == "left" {
				m.levelIndex--
			} else {
				m.levelIndex++
			}
			if m.levelIndex < 0 {
				m.levelIndex = len(levelOptions) - 1
			} else if m.levelIndex >= len(levelOptions) {
				m.levelIndex = 0
			}
		} else if m.activeFormField == 2 {
			if msg.String() == "left" {
				m.typeIndex--
			} else {
				m.typeIndex++
			}
			if m.typeIndex < 0 {
				m.typeIndex = len(typeOptions) - 1
			} else if m.typeIndex >= len(typeOptions) {
				m.typeIndex = 0
			}
		} else if m.activeFormField == 3 {
			if msg.String() == "left" {
				m.durationIndex--
			} else {
				m.durationIndex++
			}
			if m.durationIndex < 0 {
				m.durationIndex = len(durationOptions) - 1
			} else if m.durationIndex >= len(durationOptions) {
				m.durationIndex = 0
			}
		}

	case "enter":
		if m.activeFormField == 4 {
			break
		}
		// Advance to next field
		m.activeFormField++
		if m.activeFormField == 4 {
			cmd = m.sourceTextArea.Focus()
			m.lessonTitleInput.Blur()
			return m, cmd
		}

	case "ctrl+s":
		title := strings.TrimSpace(m.lessonTitleInput.Value())
		text := strings.TrimSpace(m.sourceTextArea.Value())
		if title == "" || text == "" {
			m.statusMsg = "❌ Title and Source Text are required."
			return m, nil
		}

		level := levelOptions[m.levelIndex]
		duration := durationOptions[m.durationIndex]
		lessonType := typeOptions[m.typeIndex]

		m.showForm = false
		m.generating = true
		m.statusMsg = "Generating worksheet..."
		return m, generateWorksheetCmd(m.generator, level, duration, lessonType, text, title)
	}

	if m.activeFormField == 0 {
		m.lessonTitleInput, cmd = m.lessonTitleInput.Update(msg)
	} else if m.activeFormField == 4 {
		m.sourceTextArea, cmd = m.sourceTextArea.Update(msg)
	}

	return m, cmd
}

func (m Model) handlePreviewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y":
		m.showPreview = false
		m.folderParentID = "root"
		m.folderPath = nil
		m.statusMsg = "Loading Drive folders..."
		return m, listDriveFoldersCmd(m.driveClient, "root")
	case "n", "esc", "ctrl+c":
		m.showPreview = false
		m.statusMsg = "Discarded worksheet."
		return m, nil
	}
	return m, nil
}

func (m Model) handleFolderPickerKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.showCreateFolder {
		var cmd tea.Cmd
		switch msg.String() {
		case "enter":
			name := strings.TrimSpace(m.createFolderInput.Value())
			if name != "" {
				m.showCreateFolder = false
				m.statusMsg = "Creating folder..."
				return m, createFolderCmd(m.driveClient, m.folderParentID, name)
			}
		case "esc", "ctrl+c":
			m.showCreateFolder = false
			return m, nil
		}
		m.createFolderInput, cmd = m.createFolderInput.Update(msg)
		return m, cmd
	}

	switch msg.String() {
	case "esc", "ctrl+c":
		m.showFolderPicker = false
		m.statusMsg = "Upload cancelled."
		return m, nil
	case "j", "down":
		if m.folderCursor < len(m.folders)-1 {
			m.folderCursor++
		}
	case "k", "up":
		if m.folderCursor > 0 {
			m.folderCursor--
		}
	case "enter":
		if len(m.folders) > 0 {
			selectedFolder := m.folders[m.folderCursor]
			m.folderPath = append(m.folderPath, folderBreadcrumb{id: m.folderParentID, name: selectedFolder.Name})
			m.folderParentID = selectedFolder.ID
			m.statusMsg = "Loading folder..."
			return m, listDriveFoldersCmd(m.driveClient, selectedFolder.ID)
		}
	case "backspace":
		if len(m.folderPath) > 0 {
			last := m.folderPath[len(m.folderPath)-1]
			m.folderPath = m.folderPath[:len(m.folderPath)-1]
			m.folderParentID = last.id
			m.statusMsg = "Loading folder..."
			return m, listDriveFoldersCmd(m.driveClient, m.folderParentID)
		} else {
			m.statusMsg = "Loading Drive folders..."
			return m, listDriveFoldersCmd(m.driveClient, "root")
		}
	case "n":
		m.showCreateFolder = true
		m.createFolderInput.Reset()
		return m, m.createFolderInput.Focus()
	case " ", "s":
		level := levelOptions[m.levelIndex]
		lessonType := typeOptions[m.typeIndex]
		title := m.lessonTitleInput.Value()

		var targetID, targetName string
		if len(m.folders) > 0 {
			// Upload into the selected subfolder
			folder := m.folders[m.folderCursor]
			targetID = folder.ID
			targetName = folder.Name
		} else {
			// Upload into the current directory (no subfolders to select)
			targetID = m.folderParentID
			targetName = "current folder"
		}

		m.uploading = true
		m.showFolderPicker = false
		m.statusMsg = fmt.Sprintf("Creating subfolder and uploading to %s...", targetName)
		return m, createSubfolderAndUploadCmd(m.driveClient, targetID, m.worksheetContent, m.teacherKeyContent, level, lessonType, title, m.baseDir)
	}
	return m, nil
}

func (m *Model) boundCursor() {
	var max int
	if m.activePane == 0 {
		max = len(m.todayLessons) - 1
	} else {
		max = len(m.tomorrowLessons) - 1
	}
	if max < 0 {
		max = 0
	}
	if m.cursor > max {
		m.cursor = max
	}

	maxVisible := m.height - 15
	if maxVisible < 5 {
		maxVisible = 5
	}

	if m.activePane == 0 {
		if m.cursor < m.viewportToday {
			m.viewportToday = m.cursor
		} else if m.cursor >= m.viewportToday+maxVisible {
			m.viewportToday = m.cursor - maxVisible + 1
		}
	} else {
		if m.cursor < m.viewportTomorrow {
			m.viewportTomorrow = m.cursor
		} else if m.cursor >= m.viewportTomorrow+maxVisible {
			m.viewportTomorrow = m.cursor - maxVisible + 1
		}
	}
}
