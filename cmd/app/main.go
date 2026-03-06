package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"teaching-assistant-app/internal/ai"
	"teaching-assistant-app/internal/auth"
	"teaching-assistant-app/internal/db"
	"teaching-assistant-app/internal/drive"
	"teaching-assistant-app/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
	"google.golang.org/api/calendar/v3"
	driveAPI "google.golang.org/api/drive/v3"
)

func main() {
	// Initialize Database
	store, err := db.NewStore("test_app.db")
	if err != nil {
		fmt.Printf("Error initializing db: %v\n", err)
		os.Exit(1)
	}
	defer store.Close()

	// Seed some dummy data for testing the UI
	seedDummyData(store)

	ctx := context.Background()

	// Initialize AI Generator
	generator, err := ai.NewGenerator(ctx)
	if err != nil {
		fmt.Printf("Warning: AI generator not fully initialized (is GEMINI_API_KEY set?): %v\n", err)
		// We continue anyway so the UI can be previewed
	}

	// Initialize shared OAuth HTTP client
	httpClient, err := auth.GetHTTPClient(ctx,
		"config/credentials.json",
		"config/token.json",
		calendar.CalendarReadonlyScope,
		driveAPI.DriveScope,
	)
	if err != nil {
		fmt.Printf("Warning: Google API authorization required. Please run: go run cmd/auth/main.go\n")
	}

	// Initialize Google Drive client
	var driveClient *drive.Client
	if httpClient != nil {
		driveClient, err = drive.NewClient(ctx, httpClient)
		if err != nil {
			fmt.Printf("Warning: Drive client not fully initialized: %v\n", err)
		}
	}

	baseDir, _ := os.Getwd()

	// Initialize TUI Model
	m := tui.NewModel(store, generator, driveClient, baseDir)
	
	// Start Bubble Tea program
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}

func seedDummyData(store *db.Store) {
	// Check if already seeded to avoid duplicates
	students, _ := store.GetAllStudents()
	if len(students) > 0 {
		return
	}

	// Create Students
	s1, _ := store.CreateStudent("John Doe", "B1 Intermediate", "john@example.com")
	s2, _ := store.CreateStudent("Jane Smith", "A2 Pre-Intermediate", "jane@example.com")
	s3, _ := store.CreateStudent("Alice Johnson", "C1 Advanced", "alice@example.com")

	loc := time.FixedZone("WIB", 7*3600)
	now := time.Now().In(loc)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	tomorrow := today.AddDate(0, 0, 1)

	// Create Lessons for Today
	store.CreateLesson(s1, today.Add(10*time.Hour), today.Add(11*time.Hour), "Focus on past tense")
	store.CreateLesson(s2, today.Add(14*time.Hour), today.Add(15*time.Hour), "Vocabulary expansion")

	// Create Lessons for Tomorrow
	store.CreateLesson(s3, tomorrow.Add(9*time.Hour), tomorrow.Add(10*time.Hour), "Business English")
	store.CreateLesson(s1, tomorrow.Add(16*time.Hour), tomorrow.Add(17*time.Hour), "Reading comprehension")
}
