package main

import (
	"context"
	"fmt"
	"os"

	"teaching-assistant-app/internal/ai"
	"teaching-assistant-app/internal/auth"
	"teaching-assistant-app/internal/calendar"
	"teaching-assistant-app/internal/db"
	"teaching-assistant-app/internal/drive"
	"teaching-assistant-app/internal/notify"
	"teaching-assistant-app/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
	calAPI "google.golang.org/api/calendar/v3"
	driveAPI "google.golang.org/api/drive/v3"
)

func main() {
	fmt.Println("Initializing Teaching Assistant App...")

	// 1. Initialize Database
	store, err := db.NewStore("test_app.db")
	if err != nil {
		fmt.Printf("Error initializing db: %v\n", err)
		os.Exit(1)
	}
	defer store.Close()

	// Note: We used to seed dummy data here, but removed it for production
	// so you start with a clean local database.

	ctx := context.Background()

	// 2. Initialize AI Generator
	generator, err := ai.NewGenerator(ctx)
	if err != nil {
		fmt.Printf("Warning: AI generator not fully initialized (is GEMINI_API_KEY set?): %v\n", err)
	}

	// 3. Initialize Shared Google Auth (Calendar + Drive)
	httpClient, err := auth.GetHTTPClient(ctx,
		"config/credentials.json",
		"config/token.json",
		calAPI.CalendarReadonlyScope,
		driveAPI.DriveScope,
	)
	if err != nil {
		fmt.Printf("Warning: Google API authorization required. Please run: go run cmd/auth/main.go\n")
	}

	// 4. Initialize Google Drive Client
	var driveClient *drive.Client
	var calClient *calendar.Client

	if httpClient != nil {
		driveClient, err = drive.NewClient(ctx, httpClient)
		if err != nil {
			fmt.Printf("Warning: Drive client not fully initialized: %v\n", err)
		}

		calClient, err = calendar.NewClientWithHTTP(ctx, httpClient)
		if err != nil {
			fmt.Printf("Warning: Calendar client not fully initialized: %v\n", err)
		}
	}

	// 5. Initialize WhatsApp Daemon
	waClient, err := notify.InitWhatsApp("config/whatsapp_store.db")
	if err != nil {
		fmt.Printf("Warning: WhatsApp client failed to initialize: %v\n", err)
	} else {
		defer waClient.Disconnect()

		err = waClient.Authenticate()
		if err != nil {
			fmt.Printf("Warning: WhatsApp authentication failed. Run 'go run cmd/wa_test/main.go' to scan QR: %v\n", err)
		} else if calClient != nil {
			// Start the scheduler if WA is authenticated and Calendar is available
			scheduler := notify.NewScheduler(waClient, store, calClient)
			err = scheduler.Start()
			if err != nil {
				fmt.Printf("Warning: Failed to start WhatsApp scheduler: %v\n", err)
			} else {
				defer scheduler.Stop()
			}
		}
	}

	// 6. Start TUI
	baseDir, _ := os.Getwd()
	m := tui.NewModel(store, generator, driveClient, baseDir)

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running TUI program: %v\n", err)
		os.Exit(1)
	}
}
