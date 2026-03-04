package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"teaching-assistant-app/internal/db"
)

func main() {
	dbName := "test_app.db"
	defer os.Remove(dbName) // Clean up after test

	err := db.InitDB(dbName)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.CloseDB()
	fmt.Println("Database initialized successfully.")

	// Test Student CRUD
	studentID, err := db.CreateStudent("John Doe", "Intermediate", "john@example.com")
	if err != nil {
		log.Fatalf("Failed to create student: %v", err)
	}
	fmt.Printf("Created student with ID: %d\n", studentID)

	student, err := db.GetStudent(studentID)
	if err != nil {
		log.Fatalf("Failed to get student: %v", err)
	}
	fmt.Printf("Retrieved student: %+v\n", student)

	err = db.UpdateStudent(studentID, "John Doe", "Advanced", "john.doe@example.com")
	if err != nil {
		log.Fatalf("Failed to update student: %v", err)
	}
	fmt.Println("Updated student successfully.")

	// Test Lesson CRUD
	startTime := time.Now()
	endTime := startTime.Add(1 * time.Hour)
	lessonID, err := db.CreateLesson(studentID, startTime, endTime, "Focus on grammar")
	if err != nil {
		log.Fatalf("Failed to create lesson: %v", err)
	}
	fmt.Printf("Created lesson with ID: %d\n", lessonID)

	err = db.ToggleVocabSent(lessonID, true)
	if err != nil {
		log.Fatalf("Failed to toggle vocab sent: %v", err)
	}
	fmt.Println("Toggled vocab sent status.")

	// Test Material CRUD
	materialID, err := db.CreateMaterial(lessonID, "/worksheets/lesson_1_vocab.md")
	if err != nil {
		log.Fatalf("Failed to create material: %v", err)
	}
	fmt.Printf("Created material with ID: %d\n", materialID)

	materials, err := db.GetMaterialsByLesson(lessonID)
	if err != nil {
		log.Fatalf("Failed to get materials: %v", err)
	}
	fmt.Printf("Retrieved %d materials for lesson.\n", len(materials))

	fmt.Println("All database tests passed successfully!")
}
