package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	paneStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(1, 2).
			Width(50)

	activePaneStyle = paneStyle.Copy().
			BorderForeground(lipgloss.Color("62"))

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
	if m.showForm {
		return m.formView()
	}

	maxVisible := m.height - 15
	if maxVisible < 5 {
		maxVisible = 5
	}

	leftPane := m.renderPane("📅 Today", m.todayLessons, m.activePane == 0, m.viewportToday, maxVisible)
	rightPane := m.renderPane("📅 Tomorrow", m.tomorrowLessons, m.activePane == 1, m.viewportTomorrow, maxVisible)

	panes := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)

	statusText := m.statusMsg
	if m.uploading || m.generating {
		statusText = m.spinner.View() + " " + m.statusMsg
	}
	status := statusStyle.Render(statusText)
	
	help := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("↑/k: up • ↓/j: down • ←/h: left • →/l: right • tab: switch • v: toggle vocab • g: create worksheet • q: quit")

	return lipgloss.JoinVertical(lipgloss.Left, panes, status, help)
}

func (m Model) renderPane(title string, lessons []LessonView, isActive bool, viewportStart int, maxVisible int) string {
	var b strings.Builder

	b.WriteString(lipgloss.NewStyle().Bold(true).Render(title) + "\n\n")

	if len(lessons) == 0 {
		b.WriteString(itemStyle.Render("No lessons scheduled."))
	} else {
		end := viewportStart + maxVisible
		if end > len(lessons) {
			end = len(lessons)
		}
		
		for i := viewportStart; i < end; i++ {
			lv := lessons[i]
			cursor := " "
			style := itemStyle
			if isActive && i == m.cursor {
				cursor = ">"
				style = cursorStyle
			}

			vocabStatus := "⬜"
			if lv.Lesson.VocabSent {
				vocabStatus = "✅"
			}

			timeStr := lv.Lesson.StartTime.Format("15:04")

			row := fmt.Sprintf("%s %s | %s | Lvl: %s | Vocab: %s",
				cursor, timeStr, lv.Student.Name, lv.Student.Level, vocabStatus)

			b.WriteString(style.Render(row) + "\n")
		}
		
		if len(lessons) > maxVisible {
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(fmt.Sprintf("\n  Showing %d-%d of %d", viewportStart+1, end, len(lessons))))
		}
	}

	style := paneStyle
	if isActive {
		style = activePaneStyle
	}

	return style.Render(b.String())
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

	b.WriteString("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("Tab/Shift+Tab: switch fields • ←/→: cycle options • Ctrl+S: submit • Esc: cancel"))

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
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("Enter: create • Esc: cancel"))
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
	
	return boxStyle.Render(b.String())
}
