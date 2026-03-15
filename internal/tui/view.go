package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	cursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	itemStyle   = lipgloss.NewStyle().PaddingLeft(2)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginTop(1).
			MarginBottom(1)
            
	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1, 2).
			Width(60)
)

func (m Model) View() string {
	if m.showFolderPicker {
		return m.folderPickerView()
	}
	if m.showPreview {
		return m.previewView()
	}
	
	return m.formView()
}

func (m Model) formView() string {
	var b strings.Builder
	
	b.WriteString(lipgloss.NewStyle().Bold(true).Render("📝 Create Worksheet") + "\n\n")

	// Field 0: Lesson Title
	titleCursor := " "
	if m.activeFormField == 0 {
		titleCursor = ">"
	}
	b.WriteString(fmt.Sprintf("%s %s\n\n", titleCursor, m.lessonTitleInput.View()))

	// Field 1: Level
	lvlCursor := " "
	if m.activeFormField == 1 {
		lvlCursor = ">"
	}
	b.WriteString(fmt.Sprintf("%s Target Level: ◀ %s ▶\n\n", lvlCursor, levelOptions[m.levelIndex]))

	// Field 2: Lesson Type
	typeCursor := " "
	if m.activeFormField == 2 {
		typeCursor = ">"
	}
	b.WriteString(fmt.Sprintf("%s Lesson Type:  ◀ %s ▶\n\n", typeCursor, typeOptions[m.typeIndex]))

	// Field 3: Duration
	durCursor := " "
	if m.activeFormField == 3 {
		durCursor = ">"
	}
	b.WriteString(fmt.Sprintf("%s Duration:     ◀ %s mins ▶\n\n", durCursor, durationOptions[m.durationIndex]))

	// Field 4: Source Text
	txtCursor := " "
	if m.activeFormField == 4 {
		txtCursor = ">"
	}
	b.WriteString(fmt.Sprintf("%s Source Text:\n%s\n", txtCursor, m.sourceTextArea.View()))

	b.WriteString("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("Tab/Shift+Tab: switch fields • ←/→: cycle options • Ctrl+S: submit • Ctrl+C: quit"))

	statusText := m.statusMsg
	if m.uploading || m.generating {
		statusText = m.spinner.View() + " " + m.statusMsg
	}
	if statusText != "" {
		b.WriteString("\n\n" + statusStyle.Render(statusText))
	}

	return boxStyle.Render(b.String())
}

func (m Model) previewView() string {
	var b strings.Builder
	b.WriteString(lipgloss.NewStyle().Bold(true).Render("--- Student Worksheet Preview ---") + "\n\n")
	
	lines := strings.Split(m.worksheetContent, "\n")
	limit := 20
	if len(lines) < limit {
		limit = len(lines)
	}
	
	b.WriteString(strings.Join(lines[:limit], "\n"))
	if len(lines) > limit {
		b.WriteString("\n... (truncated)\n")
	}
	
	if m.teacherKeyContent != "" {
		b.WriteString("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("43")).Render("[Teacher Key Also Generated]"))
	}
	
	b.WriteString("\n\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("y: accept & upload to Drive • n: discard • Esc: cancel"))
	
	statusText := m.statusMsg
	if m.uploading || m.generating {
		statusText = m.spinner.View() + " " + m.statusMsg
	}
	if statusText != "" {
		b.WriteString("\n\n" + statusStyle.Render(statusText))
	}

	return boxStyle.Render(b.String())
}

func (m Model) folderPickerView() string {
	var b strings.Builder
	b.WriteString(lipgloss.NewStyle().Bold(true).Render("📁 Select Drive Folder") + "\n\n")
	
	// Build breadcrumb string
	path := "📁 root"
	for _, bc := range m.folderPath {
		path += " > " + bc.name
	}
	b.WriteString(path + "\n\n")

	if m.showCreateFolder {
		b.WriteString("New folder name:\n")
		b.WriteString(m.createFolderInput.View() + "\n")
		b.WriteString("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("Enter: create • Esc: cancel"))
	} else {
		if len(m.folders) == 0 {
			b.WriteString("No subfolders. Press 'n' to create one.\n")
		} else {
			for i, f := range m.folders {
				cursor := " "
				style := itemStyle
				if i == m.folderCursor {
					cursor = ">"
					style = cursorStyle
				}
				b.WriteString(style.Render(fmt.Sprintf("%s %s", cursor, f.Name)) + "\n")
			}
		}
		
		b.WriteString("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("j/k: navigate • Enter: open folder • n: new folder • Space: upload here • Backspace: go up • Esc: cancel"))
	}
	
	statusText := m.statusMsg
	if m.uploading || m.generating {
		statusText = m.spinner.View() + " " + m.statusMsg
	}
	if statusText != "" {
		b.WriteString("\n\n" + statusStyle.Render(statusText))
	}

	return boxStyle.Render(b.String())
}
