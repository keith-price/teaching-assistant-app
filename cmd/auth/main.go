package main

import (
	"context"
	"fmt"
	"log"

	"teaching-assistant-app/internal/auth"

	driveAPI "google.golang.org/api/drive/v3"
)

func main() {
	ctx := context.Background()

	fmt.Println("Starting Google API authorization...")

	err := auth.AuthorizeInteractively(ctx,
		"config/credentials.json",
		"config/token.json",
		driveAPI.DriveScope,
	)
	if err != nil {
		log.Fatalf("❌ Authorization failed: %v", err)
	}

	fmt.Println("✅ Authorization complete. You can now run the app.")
}
