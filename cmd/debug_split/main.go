package main

import (
	"context"
	"fmt"
	"teaching-assistant-app/internal/ai"
)

func main() {
	ctx := context.Background()
	generator, err := ai.NewGenerator(ctx)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	ws, tk, err := generator.GenerateWorksheet(ctx, "B2", "50", "Reading", "This is a test article about climate change.", "Climate Change")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Worksheet len:", len(ws))
	fmt.Println("Teacher Key len:", len(tk))
}
