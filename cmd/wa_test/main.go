package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"teaching-assistant-app/internal/notify"
)

func main() {
	fmt.Println("Initializing WhatsApp client...")
	waClient, err := notify.InitWhatsApp("config/whatsapp_store.db")
	if err != nil {
		fmt.Printf("Failed to initialize WhatsApp: %v\n", err)
		os.Exit(1)
	}
	defer waClient.Disconnect()

	fmt.Println("Attempting to authenticate...")
	if err := waClient.Authenticate(); err != nil {
		fmt.Printf("Failed to authenticate: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Authentication successful (or already logged in). Press Ctrl+C to exit.")

	// Wait for a termination signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	fmt.Println("\nShutting down...")
}
