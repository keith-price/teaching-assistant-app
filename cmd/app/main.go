package main

import (
	"context"
	"fmt"
	"os"

	"teaching-assistant-app/internal/ai"
	"teaching-assistant-app/internal/auth"
	"teaching-assistant-app/internal/drive"
	"teaching-assistant-app/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
	driveAPI "google.golang.org/api/drive/v3"
)

func main() {
	fmt.Println("Initializing Teaching Assistant App...")

	ctx := context.Background()

	// 1. Initialize AI Generator
	generator, err := ai.NewGenerator(ctx)
	if err != nil {
		fmt.Printf("Warning: AI generator not fully initialized (is GEMINI_API_KEY set?): %v\n", err)
	}

	// 2. Initialize Shared Google Auth (Drive only)
	httpClient, err := auth.GetHTTPClient(ctx,
		"config/credentials.json",
		"config/token.json",
		driveAPI.DriveScope,
	)
	if err != nil {
		fmt.Printf("Warning: Google API authorization required. Please run: go run cmd/auth/main.go\n")
	}

	// 3. Initialize Google Drive Client
	var driveClient *drive.Client

	if httpClient != nil {
		driveClient, err = drive.NewClient(ctx, httpClient)
		if err != nil {
			fmt.Printf("Warning: Drive client not fully initialized: %v\n", err)
		}
	}

	// 4. Start TUI
	baseDir, _ := os.Getwd()
	m := tui.NewModel(generator, driveClient, baseDir)

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running TUI program: %v\n", err)
		os.Exit(1)
	}
}
