package db_test

import (
	"testing"
	"time"

	"teaching-assistant-app/internal/db"
)

func setupTestStore(t *testing.T) *db.Store {
	t.Helper()
	store, err := db.NewStore(":memory:")
	if err != nil {
		t.Fatalf("failed to init test db: %v", err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Errorf("failed to close test db: %v", err)
		}
	})
	return store
}

func TestStudentCRUD(t *testing.T) {
	store := setupTestStore(t)

	// Create
	studentID, err := store.CreateStudent("John Doe", "Intermediate", "john@example.com")
	if err != nil {
		t.Fatalf("failed to create student: %v", err)
	}

	// Get
	student, err := store.GetStudent(studentID)
	if err != nil {
		t.Fatalf("failed to get student: %v", err)
	}
	if student == nil || student.Name != "John Doe" {
		t.Errorf("expected student 'John Doe', got %v", student)
	}

	// Update
	err = store.UpdateStudent(studentID, "John Doe Updated", "Advanced", "john.doe@example.com")
	if err != nil {
		t.Fatalf("failed to update student: %v", err)
	}

	// Get again
	student, err = store.GetStudent(studentID)
	if err != nil {
		t.Fatalf("failed to get student: %v", err)
	}
	if student == nil || student.Name != "John Doe Updated" {
		t.Errorf("expected student 'John Doe Updated', got %v", student)
	}
	
	// Error on duplicate name (since it's UNIQUE)
	_, err = store.CreateStudent("John Doe Updated", "Beginner", "test@test.com")
	if err == nil {
		t.Error("expected error creating student with duplicate name, got nil")
	}

	// Delete
	err = store.DeleteStudent(studentID)
	if err != nil {
		t.Fatalf("failed to delete student: %v", err)
	}

	// Verify Delete
	student, err = store.GetStudent(studentID)
	if err != nil {
		t.Fatalf("failed to get deleted student (expected nil, got error): %v", err)
	}
	if student != nil {
		t.Errorf("expected deleted student to be nil, got %v", student)
	}
}

func TestLessonCRUD(t *testing.T) {
	store := setupTestStore(t)

	studentID, _ := store.CreateStudent("Jane Doe", "Beginner", "jane@example.com")

	start := time.Now().Add(-2 * time.Hour)
	end := start.Add(1 * time.Hour)

	lessonID, err := store.CreateLesson(studentID, start, end, "Intro")
	if err != nil {
		t.Fatalf("failed to create lesson: %v", err)
	}

	lesson, err := store.GetLesson(lessonID)
	if err != nil {
		t.Fatalf("failed to get lesson: %v", err)
	}
	if lesson == nil || !lesson.VocabSent {
		// It's false initially
		if lesson == nil {
			t.Fatalf("lesson was nil")
		}
	}

	err = store.ToggleVocabSent(lessonID)
	if err != nil {
		t.Fatalf("failed to toggle vocab sent: %v", err)
	}

	lesson, err = store.GetLesson(lessonID)
	if err != nil {
		t.Fatalf("failed to get lesson: %v", err)
	}
	if !lesson.VocabSent {
		t.Errorf("expected vocab_sent to be true, got %v", lesson.VocabSent)
	}

	err = store.DeleteLesson(lessonID)
	if err != nil {
		t.Fatalf("failed to delete lesson: %v", err)
	}
}

func TestMaterialCRUD(t *testing.T) {
	store := setupTestStore(t)

	studentID, _ := store.CreateStudent("Bob", "Advanced", "bob@example.com")
	lessonID, _ := store.CreateLesson(studentID, time.Now(), time.Now().Add(1*time.Hour), "Review")

	materialID, err := store.CreateMaterial(lessonID, "/files/doc.pdf")
	if err != nil {
		t.Fatalf("failed to create material: %v", err)
	}

	materials, err := store.GetMaterialsByLesson(lessonID)
	if err != nil {
		t.Fatalf("failed to get materials: %v", err)
	}
	if len(materials) != 1 {
		t.Errorf("expected 1 material, got %d", len(materials))
	}
	if materials[0].ID != materialID {
		t.Errorf("expected material ID %d, got %d", materialID, materials[0].ID)
	}

	err = store.DeleteMaterial(materialID)
	if err != nil {
		t.Fatalf("failed to delete material: %v", err)
	}
}