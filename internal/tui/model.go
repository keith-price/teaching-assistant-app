package tui

import (
	"teaching-assistant-app/internal/ai"
	"teaching-assistant-app/internal/drive"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	levelOptions    = []string{"A1", "A2", "B1", "B2", "C1", "C2"}
	typeOptions     = []string{"Reading", "Listening"}
	durationOptions = []string{"50", "70", "90", "110"}
)

type folderBreadcrumb struct {
	id   string
	name string
}

type Model struct {
	generator   *ai.Generator
	driveClient *drive.Client

	width  int
	height int

	statusMsg string
	err       error

	// Form state
	lessonTitleInput textinput.Model
	sourceTextArea   textarea.Model
	levelIndex       int
	typeIndex        int
	durationIndex    int
	activeFormField  int // 0: Title, 1: Level, 2: Type, 3: Duration, 4: Text

	// Preview state
	showPreview       bool
	worksheetContent  string
	teacherKeyContent string

	// Folder picker state
	showFolderPicker  bool
	folders           []drive.Folder
	folderCursor      int
	folderParentID    string
	folderPath        []folderBreadcrumb // Stack for back-navigation
	showCreateFolder  bool               // Inline folder name input active
	createFolderInput textinput.Model    // Text input for new folder name

	// Progress states
	spinner    spinner.Model
	uploading  bool
	generating bool

	baseDir string
}

func NewModel(generator *ai.Generator, driveClient *drive.Client, baseDir string) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	m := Model{
		generator:      generator,
		driveClient:    driveClient,
		baseDir:        baseDir,
		folderParentID: "root",
		spinner:        s,
	}

	m.initForm()
	return m
}

func (m *Model) initForm() {
	ti := textinput.New()
	ti.Placeholder = "Lesson Title (e.g., Climate Change)"
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 50
	m.lessonTitleInput = ti

	ta := textarea.New()
	ta.Placeholder = "Paste source text here..."
	ta.SetWidth(50)
	ta.SetHeight(10)
	m.sourceTextArea = ta

	ci := textinput.New()
	ci.Placeholder = "New folder name..."
	ci.CharLimit = 50
	ci.Width = 30
	m.createFolderInput = ci

	m.levelIndex = 3    // B2
	m.typeIndex = 0     // Reading
	m.durationIndex = 0 // 50
	m.activeFormField = 0
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		textarea.Blink,
		m.spinner.Tick,
	)
}
